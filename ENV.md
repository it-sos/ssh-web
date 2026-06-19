# Environment Variable Configuration

ssh-web supports configuring all settings via environment variables. Environment variables are checked when generating the initial `config.yaml` or filling in missing values.

## Reference

### Server Settings

| Environment Variable | Config Field | Default | Description |
|---------------------|-------------|---------|-------------|
| `SSH_WEB_SERVER_PORT` | `server.port` | `8080` | HTTP server port |
| `SSH_WEB_SERVER_TLS_CERT` | `server.tls_cert` | (empty) | Path to TLS certificate file |
| `SSH_WEB_SERVER_TLS_KEY` | `server.tls_key` | (empty) | Path to TLS key file |
| `SSH_WEB_SERVER_BASE_PATH` | `server.base_path` | (empty) | URL path prefix (e.g. `/ssh-web`) |

### Auth Settings

| Environment Variable | Config Field | Default | Description |
|---------------------|-------------|---------|-------------|
| `SSH_WEB_AUTH_USERNAME` | `auth.username` | `admin` | Login username |
| `SSH_WEB_AUTH_PASSWORD_HASH` | `auth.password_hash` | (auto-generated) | bcrypt hash of the login password |
| `SSH_WEB_AUTH_TOTP_SECRET` | `auth.totp_secret` | (auto-generated) | Base32 TOTP secret for 2FA |

### Encryption Settings

| Environment Variable | Config Field | Default | Description |
|---------------------|-------------|---------|-------------|
| `SSH_WEB_ENCRYPTION_KEY` | `encryption_key` | (auto-generated) | AES-256 key for encrypting SSH passwords |

### Default Host Settings

| Environment Variable | Config Field | Default | Description |
|---------------------|-------------|---------|-------------|
| `SSH_WEB_DEFAULT_HOST_HOST` | `default_host.host` | `127.0.0.1` | Target SSH server hostname/IP |
| `SSH_WEB_DEFAULT_HOST_PORT` | `default_host.port` | `22` | Target SSH server port |
| `SSH_WEB_DEFAULT_HOST_USERNAME` | `default_host.username` | `root` | SSH login username |
| `SSH_WEB_DEFAULT_HOST_AUTH_METHOD` | `default_host.auth_method` | `password` | SSH auth method: `password` or `private_key` |
| `SSH_WEB_DEFAULT_HOST_PASSWORD_ENCRYPTED` | `default_host.password_encrypted` | (empty) | Pre-encrypted SSH password for target host |
| `SSH_WEB_DEFAULT_HOST_PRIVATE_KEY_PATH` | `default_host.private_key_path` | (empty) | Path to SSH private key file |
| `SSH_WEB_DEFAULT_HOST_HOST_KEY_CHECK` | `default_host.host_key_check` | `true` | Enable SSH host key verification |

## Usage

### Generate Config with Environment Variables

```bash
export SSH_WEB_SERVER_PORT=9090
export SSH_WEB_AUTH_USERNAME=myadmin
export SSH_WEB_DEFAULT_HOST_HOST=myserver.example.com
./ssh-web
```

When no `config.yaml` exists, these values are used instead of the built-in defaults.

### Override on Startup

Set values inline without modifying `config.yaml`:

```bash
SSH_WEB_SERVER_PORT=9090 SSH_WEB_DEFAULT_HOST_HOST=myserver.example.com ./ssh-web
```

### Encrypted SSH Password via Environment

Use the CLI tool to encrypt a password and pass it via environment:

```bash
PASSWORD_ENCRYPTED=$(./ssh-web encrypt-password mypassword)
SSH_WEB_DEFAULT_HOST_PASSWORD_ENCRYPTED=$PASSWORD_ENCRYPTED ./ssh-web
```

## Order of Precedence

1. `config.yaml` values (highest priority)
2. Environment variables
3. Built-in defaults and auto-generated secrets (lowest priority)

Auto-generated secrets (`password_hash`, `totp_secret`, `encryption_key`) are only generated when neither the env var nor the config file provides a value.
