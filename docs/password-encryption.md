# Password Encryption

`password_encrypted` 用于在 `config.yaml` 中存储加密后的 SSH 密码，防止明文密码出现在日志或配置文件中。

## Encryption Key

首次启动时自动生成 32 字节随机值，base64 编码后写入 `config.yaml`：

```yaml
encryption_key: "NDhKaTBXXXVuUzFQNENhWmc2Vmc1VGhydzlnPT0="
```

## 加密算法

**AES-256-GCM**，流程：

1. `encryption_key` base64 解码 → 32 字节 AES-256 密钥
2. 生成 12 字节随机 nonce（每次加密结果不同）
3. AES-256-GCM 加密，输出 = `nonce(12B) + 密文` 拼接
4. 整体 base64 编码 → `password_encrypted` 值

## 使用方法

### 方式一：加密工具（推荐）

```bash
go build -o encrypt-password ./cmd/encrypt-password/
./encrypt-password "your-ssh-password"
```

输出即为 `password_encrypted` 的值，复制到 `config.yaml`：

```yaml
default_host:
  password_encrypted: "5a7B...d5k="
```

### 方式二：Go 代码

```go
encrypted, _ := config.Encrypt(cfg.EncryptionKey, "your-ssh-password")
```

### 方式三：解密验证

```go
password, _ := config.Decrypt(cfg.EncryptionKey, cfg.DefaultHost.PasswordEncrypted)
```
