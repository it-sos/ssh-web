package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

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
	Port    int    `yaml:"port"`
	TLSCert string `yaml:"tls_cert"`
	TLSKey  string `yaml:"tls_key"`
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
	totpSecret := generateRandomString(20)
	encKey := generateRandomString(32)

	return Config{
		Server: ServerConfig{
			Port: 8080,
		},
		Auth: AuthConfig{
			Username:     "admin",
			PasswordHash: string(hash),
			TOTPSecret:   totpSecret,
		},
		EncryptionKey: base64.StdEncoding.EncodeToString([]byte(encKey)),
		DefaultHost: DefaultHostConfig{
			Host:         "127.0.0.1",
			Port:         22,
			Username:     "root",
			AuthMethod:   "password",
			HostKeyCheck: true,
		},
	}
}

func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Auth.Username == "" {
		cfg.Auth.Username = "admin"
	}
	if cfg.Auth.PasswordHash == "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte("changeme"), bcrypt.DefaultCost)
		cfg.Auth.PasswordHash = string(hash)
	}
	if cfg.Auth.TOTPSecret == "" {
		cfg.Auth.TOTPSecret = generateRandomString(20)
	}
	if cfg.EncryptionKey == "" {
		cfg.EncryptionKey = base64.StdEncoding.EncodeToString([]byte(generateRandomString(32)))
	}
	if cfg.DefaultHost.Port == 0 {
		cfg.DefaultHost.Port = 22
	}
	if cfg.DefaultHost.AuthMethod == "" {
		cfg.DefaultHost.AuthMethod = "password"
	}
	cfg.DefaultHost.HostKeyCheck = true
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

func generateRandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n]
}
