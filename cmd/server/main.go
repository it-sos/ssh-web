package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ssh-web/internal/auth"
	"github.com/ssh-web/internal/config"
	"github.com/ssh-web/internal/ssh"
	"github.com/ssh-web/internal/ws"
	"github.com/ssh-web/web"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	sessionStore := auth.NewSessionStore()
	loginLimiter := auth.NewRateLimiter(5, 15*time.Minute)
	totpLimiter := auth.NewRateLimiter(5, 15*time.Minute)

	var password string
	if cfg.DefaultHost.PasswordEncrypted != "" {
		password, err = config.Decrypt(cfg.EncryptionKey, cfg.DefaultHost.PasswordEncrypted)
		if err != nil {
			slog.Error("Failed to decrypt password", "error", err)
			os.Exit(1)
		}
	}

	sshCfg, err := ssh.NewClientConfig(ssh.Config{
		Host:           cfg.DefaultHost.Host,
		Port:           cfg.DefaultHost.Port,
		Username:       cfg.DefaultHost.Username,
		AuthMethod:     cfg.DefaultHost.AuthMethod,
		Password:       password,
		PrivateKeyPath: cfg.DefaultHost.PrivateKeyPath,
		HostKeyCheck:   cfg.DefaultHost.HostKeyCheck,
	})
	if err != nil {
		slog.Error("Failed to create SSH config", "error", err)
		os.Exit(1)
	}

	if cfg.Auth.TOTPSecret != "" {
		slog.Info("TOTP Secret (scan with authenticator app)", "secret", cfg.Auth.TOTPSecret)
		slog.Info("Default password", "password", "changeme")
	}

	mux := http.NewServeMux()

	staticFS, _ := fs.Sub(web.StaticFiles, ".")
	mux.Handle("/css/", http.FileServer(http.FS(staticFS)))
	mux.Handle("/js/", http.FileServer(http.FS(staticFS)))
	mux.Handle("/vendor/", http.FileServer(http.FS(staticFS)))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFileFS(w, r, web.StaticFiles, "index.html")
	})
	mux.HandleFunc("/totp", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, web.StaticFiles, "totp.html")
	})
	mux.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, web.StaticFiles, "terminal.html")
	})

	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
			return
		}

		if !loginLimiter.Allow(req.Username) {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "Too many attempts. Try again later."})
			return
		}

		if req.Username != cfg.Auth.Username || !auth.VerifyPassword(req.Password, cfg.Auth.PasswordHash) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid username or password"})
			return
		}

		slog.Info("Login successful", "user", req.Username)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/totp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		token := sessionStore.GetSessionToken(r)
		userID, ok := sessionStore.ValidateSession(token)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
			return
		}

		if !totpLimiter.Allow(userID) {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "Too many attempts. Try again later."})
			return
		}

		if !auth.VerifyTOTP(cfg.Auth.TOTPSecret, req.Code) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid verification code"})
			return
		}

		slog.Info("TOTP verification successful", "user", userID)

		sessionStore.DeleteSession(token)
		newToken := sessionStore.CreateSession(userID)
		secure := cfg.Server.TLSCert != ""
		sessionStore.SetCookie(w, r, newToken, secure)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	wsHandler := ws.NewHandler(sessionStore, sshCfg)
	mux.HandleFunc("/ws", wsHandler.ServeHTTP)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		slog.Info("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Server shutdown error", "error", err)
		}
	}()

	slog.Info("Starting server", "addr", addr)
	if cfg.Server.TLSCert != "" && cfg.Server.TLSKey != "" {
		if err := server.ListenAndServeTLS(cfg.Server.TLSCert, cfg.Server.TLSKey); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	} else {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}
}
