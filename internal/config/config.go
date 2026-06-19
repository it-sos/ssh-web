package config

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server        ServerConfig      `yaml:"server"`
	Auth          AuthConfig        `yaml:"auth"`
	EncryptionKey string            `yaml:"encryption_key"`
	DefaultHost   DefaultHostConfig `yaml:"default_host"`
}

type ServerConfig struct {
	Port     int    `yaml:"port"`
	TLSCert  string `yaml:"tls_cert"`
	TLSKey   string `yaml:"tls_key"`
	BasePath string `yaml:"base_path"`
}

type AuthConfig struct {
	Username     string `yaml:"username"`
	PasswordHash string `yaml:"password_hash"`
	TOTPSecret   string `yaml:"totp_secret"`
}

type DefaultHostConfig struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Username          string `yaml:"username"`
	AuthMethod        string `yaml:"auth_method"`
	PasswordEncrypted string `yaml:"password_encrypted"`
	PrivateKeyPath    string `yaml:"private_key_path"`
	HostKeyCheck      bool   `yaml:"host_key_check"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
		setDefaults(&cfg)
		return &cfg, nil
	}

	cfg = defaultConfig()
	if err := saveConfig(path, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func defaultConfig() Config {
	hash, _ := bcrypt.GenerateFromPassword([]byte("changeme"), bcrypt.DefaultCost)
	totpSecret := generateBase32Secret(20)
	encKey := generateRandomString(32)

	return Config{
		Server: ServerConfig{
			Port:     envOrInt("SSH_WEB_SERVER_PORT", 8080),
			TLSCert:  os.Getenv("SSH_WEB_SERVER_TLS_CERT"),
			TLSKey:   os.Getenv("SSH_WEB_SERVER_TLS_KEY"),
			BasePath: os.Getenv("SSH_WEB_SERVER_BASE_PATH"),
		},
		Auth: AuthConfig{
			Username:     envOrString("SSH_WEB_AUTH_USERNAME", "admin"),
			PasswordHash: envOrString("SSH_WEB_AUTH_PASSWORD_HASH", string(hash)),
			TOTPSecret:   envOrString("SSH_WEB_AUTH_TOTP_SECRET", totpSecret),
		},
		EncryptionKey: envOrString("SSH_WEB_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString([]byte(encKey))),
		DefaultHost: DefaultHostConfig{
			Host:              envOrString("SSH_WEB_DEFAULT_HOST_HOST", "127.0.0.1"),
			Port:              envOrInt("SSH_WEB_DEFAULT_HOST_PORT", 22),
			Username:          envOrString("SSH_WEB_DEFAULT_HOST_USERNAME", "root"),
			AuthMethod:        envOrString("SSH_WEB_DEFAULT_HOST_AUTH_METHOD", "password"),
			PasswordEncrypted: os.Getenv("SSH_WEB_DEFAULT_HOST_PASSWORD_ENCRYPTED"),
			PrivateKeyPath:    os.Getenv("SSH_WEB_DEFAULT_HOST_PRIVATE_KEY_PATH"),
			HostKeyCheck:      envOrBool("SSH_WEB_DEFAULT_HOST_HOST_KEY_CHECK", true),
		},
	}
}

func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = envOrInt("SSH_WEB_SERVER_PORT", 8080)
	}
	if cfg.Server.TLSCert == "" {
		cfg.Server.TLSCert = os.Getenv("SSH_WEB_SERVER_TLS_CERT")
	}
	if cfg.Server.TLSKey == "" {
		cfg.Server.TLSKey = os.Getenv("SSH_WEB_SERVER_TLS_KEY")
	}
	if cfg.Server.BasePath == "" {
		cfg.Server.BasePath = os.Getenv("SSH_WEB_SERVER_BASE_PATH")
	}
	if cfg.Auth.Username == "" {
		cfg.Auth.Username = envOrString("SSH_WEB_AUTH_USERNAME", "admin")
	}
	if cfg.Auth.PasswordHash == "" {
		v := os.Getenv("SSH_WEB_AUTH_PASSWORD_HASH")
		if v != "" {
			cfg.Auth.PasswordHash = v
		} else {
			hash, _ := bcrypt.GenerateFromPassword([]byte("changeme"), bcrypt.DefaultCost)
			cfg.Auth.PasswordHash = string(hash)
		}
	}
	if cfg.Auth.TOTPSecret == "" {
		v := os.Getenv("SSH_WEB_AUTH_TOTP_SECRET")
		if v != "" {
			cfg.Auth.TOTPSecret = v
		} else {
			cfg.Auth.TOTPSecret = generateBase32Secret(20)
		}
	}
	if cfg.EncryptionKey == "" {
		v := os.Getenv("SSH_WEB_ENCRYPTION_KEY")
		if v != "" {
			cfg.EncryptionKey = v
		} else {
			cfg.EncryptionKey = base64.StdEncoding.EncodeToString([]byte(generateRandomString(32)))
		}
	}
	if cfg.DefaultHost.Host == "" {
		cfg.DefaultHost.Host = envOrString("SSH_WEB_DEFAULT_HOST_HOST", "127.0.0.1")
	}
	if cfg.DefaultHost.Port == 0 {
		cfg.DefaultHost.Port = envOrInt("SSH_WEB_DEFAULT_HOST_PORT", 22)
	}
	if cfg.DefaultHost.Username == "" {
		cfg.DefaultHost.Username = envOrString("SSH_WEB_DEFAULT_HOST_USERNAME", "root")
	}
	if cfg.DefaultHost.AuthMethod == "" {
		cfg.DefaultHost.AuthMethod = envOrString("SSH_WEB_DEFAULT_HOST_AUTH_METHOD", "password")
	}
	if cfg.DefaultHost.PasswordEncrypted == "" {
		cfg.DefaultHost.PasswordEncrypted = os.Getenv("SSH_WEB_DEFAULT_HOST_PASSWORD_ENCRYPTED")
	}
	if cfg.DefaultHost.PrivateKeyPath == "" {
		cfg.DefaultHost.PrivateKeyPath = os.Getenv("SSH_WEB_DEFAULT_HOST_PRIVATE_KEY_PATH")
	}
	cfg.DefaultHost.HostKeyCheck = envOrBool("SSH_WEB_DEFAULT_HOST_HOST_KEY_CHECK", true)
}

func saveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func envOrString(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envOrInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func envOrBool(key string, defaultVal bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultVal
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n]
}

func generateBase32Secret(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	// Base32 encoding produces uppercase letters A-Z and digits 2-7,
	// which is the valid character set for TOTP secrets (RFC 4648).
	return strings.TrimRight(base32.StdEncoding.EncodeToString(b), "=")
}
