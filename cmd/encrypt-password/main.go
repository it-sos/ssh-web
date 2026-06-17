package main

import (
	"fmt"
	"os"

	"github.com/ssh-web/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <password>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Reads encryption_key from config.yaml and encrypts the given password.\n")
		os.Exit(1)
	}

	password := os.Args[1]

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if cfg.EncryptionKey == "" {
		fmt.Fprintln(os.Stderr, "Error: encryption_key is not set in config.yaml")
		os.Exit(1)
	}

	encrypted, err := config.Encrypt(cfg.EncryptionKey, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encrypting password: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(encrypted)
}
