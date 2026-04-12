package ssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type Config struct {
	Host           string
	Port           int
	Username       string
	AuthMethod     string
	Password       string
	PrivateKeyPath string
	HostKeyCheck   bool
}

type ClientConfig struct {
	*ssh.ClientConfig
	Host string
	Port int
}

func NewClientConfig(cfg Config) (*ClientConfig, error) {
	sshCfg := &ssh.ClientConfig{
		User: cfg.Username,
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			if !cfg.HostKeyCheck {
				return nil
			}
			return checkHostKey(hostname, remote, key)
		}),
	}

	switch cfg.AuthMethod {
	case "password":
		sshCfg.Auth = append(sshCfg.Auth, ssh.Password(cfg.Password))
	case "private_key":
		key, err := loadPrivateKey(cfg.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("load private key: %w", err)
		}
		sshCfg.Auth = append(sshCfg.Auth, key)
	default:
		return nil, fmt.Errorf("unknown auth method: %s", cfg.AuthMethod)
	}

	return &ClientConfig{
		ClientConfig: sshCfg,
		Host:         cfg.Host,
		Port:         cfg.Port,
	}, nil
}

func loadPrivateKey(path string) (ssh.AuthMethod, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key file: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("parse key: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}

func checkHostKey(hostname string, remote net.Addr, key ssh.PublicKey) error {
	home, _ := os.UserHomeDir()
	knownHostsPath := filepath.Join(home, ".ssh_web", "known_hosts")

	if err := os.MkdirAll(filepath.Dir(knownHostsPath), 0700); err != nil {
		return err
	}

	hostKeyCallback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return err
	}

	err = hostKeyCallback(hostname, remote, key)
	if err != nil {
		if knownHostError, ok := err.(*knownhosts.KeyError); ok && len(knownHostError.Want) == 0 {
			return addHostKey(knownHostsPath, hostname, remote, key)
		}
		return err
	}

	return nil
}

func addHostKey(path, hostname string, remote net.Addr, key ssh.PublicKey) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	line := knownhosts.Line([]string{hostname}, key)
	_, err = fmt.Fprintln(f, line)
	return err
}

func Connect(cfg *ClientConfig) (*ssh.Client, *ssh.Session, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client, err := ssh.Dial("tcp", addr, cfg.ClientConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}
