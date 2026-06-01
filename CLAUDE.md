# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

This is a Go-based web application that provides SSH access through a browser interface.

### Key Components

| Component | Path | Purpose |
|-----------|------|---------|
| Server | `cmd/server/` | Main HTTP server with routes |
| Config | `internal/config/` | YAML config loading, password encryption/decryption |
| Auth | `internal/auth/` | Session management, password hashing (bcrypt), TOTP verification, rate limiting |
| SSH | `internal/ssh/` | SSH client configuration and connection |
| WebSocket | `internal/ws/` | WebSocket handler for terminal I/O streaming |
| Static | `web/` | Frontend HTML/JS files |

### Request Flow

1. **Login**: POST to `/api/login` with username/password → validates → creates session token → sets cookie
2. **TOTP**: POST to `/api/totp` with code → validates against TOTP secret → deletes old session → creates new session
3. **Terminal**: Browser connects to `/ws` → authenticates via cookie → establishes SSH connection → streams I/O bidirectionally

### Critical Files

- [`internal/auth/auth.go`](internal/auth/auth.go) - Session store with 30min expiry, rate limiter (5 req/15min)
- [`internal/auth/session.go`](internal/auth/session.go) - bcrypt password verification, TOTP (OTP) verification
- [`internal/ssh/client.go`](internal/ssh/client.go) - SSH client config with password/PrivateKey auth, host key checking
- [`internal/ws/handler.go`](internal/ws/handler.go) - Gorilla WebSocket server, bidirectional I/O streaming, resize handling
- [`cmd/server/main.go`](cmd/server/main.go) - HTTP routes, TLS support, graceful shutdown

## Build & Run

```bash
go build -o ssh-web ./cmd/server  # Build
./ssh-web                         # Run
# or use make
make build                        # Build
make run                          # Build and run
```

To test:
1. Edit [`config.yaml`](config.yaml) with your target SSH server credentials
2. Deploy the binary to a web server (or run locally)
3. Open web interface at http://server:8080

## Configuration

[`config.yaml`](config.yaml) contains:

- `server.port` - HTTP server port (default: 8080)
- `server.tls_cert` / `server.tls_key` - Optional TLS certificates
- `auth.username` - Login username (default: admin)
- `auth.password_hash` - bcrypt hash of login password
- `auth.totp_secret` - Base32 secret for TOTP 2FA
- `default_host.*` - Target SSH server settings

### Password Encryption

Passwords are encrypted with a key stored in `encryption_key`:

```go
// Encrypt
encrypted := Encrypt(key, password)

// Decrypt
password, _ := config.Decrypt(key, encrypted)
```

This prevents passwords from appearing in logs.

### Session Management

- Session tokens are base64-encoded random strings
- Default expiry: 30 minutes
- Rate limited: 5 login attempts per 15 minutes
- Max concurrent sessions: 10 (enforced via WebSocket handler)

## API Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/login` | POST | Login with username/password |
| `/api/totp` | POST | Verify TOTP code, get new session |
| `/ws` | WebSocket | Terminal I/O stream (after login) |
| `/`, `/totp`, `/terminal` | GET | Static web pages |

## Testing

```bash
go test ./... -v
```

- [`tests/integration_test.go`](tests/integration_test.go) - Session expiry test
- [`internal/config/*_test.go`](internal/config/*_test.go) - Config loading, encryption
- [`internal/auth/*_test.go`](internal/auth/*_test.go) - Auth tests including ratelimit
- [`internal/ssh/client_test.go`](internal/ssh/client_test.go) - SSH client tests

## Important Notes

- Static assets are served from `.web` directory
- WebSocket upgrade bypasses origin check (`CheckOrigin: func(r *http.Request) bool { return true }`)
- SSH host key checking defaults to `true` (creates `~/.ssh_web/known_hosts`)
- Go runtime will signal-shutdown gracefully on SIGINT/SIGTERM

## Graphify

Project has a [`graphify-out/`](graphify-out/) knowledge graph. Run after modifying code:

```bash
python3 -c "from graphify.watch import _rebuild_code; from pathlib import Path; _rebuild_code(Path('.'))"
```
