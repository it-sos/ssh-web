# Web SSH 终端 - 设计文档

## 概述

构建一个基于 Web 的 SSH 终端应用，用户通过浏览器访问，经过用户名/密码 + TOTP 二次认证后，自动连接到管理员配置的默认主机，获得完整的 bash 终端体验。用户可在终端内自行 SSH 到其他主机。

## 需求总结

- **认证**：用户名/密码 + TOTP 二次认证
- **终端**：登录后自动连接管理员配置的默认主机
- **连接**：用户可在终端内自行 SSH 到其他主机
- **用户**：单用户模式
- **技术栈**：Go + xterm.js
- **部署**：单二进制文件，开箱即用

## 架构设计

```
┌─────────────────────────────────────────────────────┐
│                    浏览器                            │
│  ┌───────────────────────────────────────────────┐  │
│  │  登录页 → TOTP 验证页 → 终端页 (xterm.js)     │  │
│  └───────────────────────────────────────────────┘  │
│                         │ WebSocket                 │
├─────────────────────────┼───────────────────────────┤
│                    Go 后端                            │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────┐  │
│  │ HTTP 路由   │  │ WebSocket    │  │ SSH 客户端 │  │
│  │ (登录/配置) │  │ 终端代理     │  │ 连接池     │  │
│  └─────────────┘  └──────────────┘  └────────────┘  │
│  ┌─────────────┐  ┌──────────────┐                   │
│  │ TOTP 验证   │  │ 配置管理     │                   │
│  │ (pquerna/otp)│ │ (YAML/JSON)  │                   │
│  └─────────────┘  └──────────────┘                   │
├─────────────────────────────────────────────────────┤
│                    目标主机                          │
│  ┌───────────────────────────────────────────────┐  │
│  │  默认主机 (SSH)                                │  │
│  └───────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### 核心组件

1. **认证层**：Session 管理 + TOTP 验证，使用 `pquerna/otp` 库
2. **终端代理**：WebSocket ↔ SSH 双向数据流，使用 `golang.org/x/crypto/ssh`
3. **配置管理**：YAML 配置文件存储默认主机、用户名/密码哈希、TOTP Secret
4. **前端嵌入**：`go:embed` 将静态资源打包到二进制

### 数据流

用户登录 → 密码验证 → TOTP 验证 → 建立 Session → 打开终端页 → WebSocket 连接 → 后端创建 SSH 连接到默认主机 → 终端数据双向转发

## 认证与安全

### 认证流程

1. 用户访问 → 显示登录页
2. 输入用户名 + 密码 → 后端验证（bcrypt 哈希比对）
3. 验证通过 → 显示 TOTP 验证码页
4. 输入 6 位动态码 → 后端验证（TOTP 时间窗口比对）
5. 验证通过 → 创建 Session（HttpOnly Cookie）
6. 跳转到终端页 → WebSocket 携带 Session 认证

### 安全设计

| 项目 | 方案 |
|------|------|
| 密码存储 | bcrypt 哈希，不存明文 |
| TOTP Secret | 首次启动时生成，显示二维码供用户绑定 |
| Session | HttpOnly + Secure Cookie，30 分钟过期 |
| WebSocket 认证 | 通过 Cookie 验证 Session，拒绝未认证连接 |
| 暴力破解防护 | 登录失败 5 次锁定 15 分钟 |
| HTTPS | 支持配置 TLS 证书（可选） |

### 配置文件示例 (`config.yaml`)

```yaml
server:
  port: 8080
  tls_cert: ""
  tls_key: ""

auth:
  username: "admin"
  password_hash: "$2a$10$..."  # 首次启动自动生成
  totp_secret: ""                 # 首次启动自动生成

default_host:
  host: "192.168.1.100"
  port: 22
  username: "root"
  password: ""        # SSH 密码认证
  private_key: ""     # 或 SSH 私钥路径
```

## 终端通信与 SSH 代理

### WebSocket 协议

前端与后端通过 JSON 消息通信：

```json
{
  "type": "data" | "resize" | "close" | "error",
  "payload": string | object
}
```

| 类型 | 方向 | 说明 |
|------|------|------|
| `data` | 前端→后端 | 用户输入的字符 |
| `data` | 后端→前端 | SSH 输出内容 |
| `resize` | 前端→后端 | 终端尺寸变化 `{cols, rows}` |
| `close` | 任意 | 关闭连接 |
| `error` | 后端→前端 | 错误信息 |

### SSH 连接管理

1. 用户打开终端页 → 后端读取配置创建 SSH 客户端
2. 连接成功 → 打开 Session，分配伪终端 (PTY)
3. 终端 ↔ SSH 双向数据转发
4. 连接断开 → 自动重连（最多 3 次，间隔递增）
5. 用户关闭页面 → 清理 SSH Session

### 终端配置

- 默认尺寸：80 cols × 24 rows
- 字体：Monaco / Consolas / 等宽字体栈
- 主题：黑色背景（可切换）
- 支持：UTF-8、256 色、鼠标事件

## 项目结构

```
ssh-web/
├── cmd/
│   └── server/main.go          # 入口
├── internal/
│   ├── auth/
│   │   ├── auth.go             # 密码验证
│   │   └── totp.go             # TOTP 生成/验证
│   ├── config/
│   │   └── config.go           # 配置加载
│   ├── ssh/
│   │   └── client.go           # SSH 客户端封装
│   └── ws/
│       └── handler.go          # WebSocket 处理
├── web/
│   ├── index.html              # 登录页
│   ├── totp.html               # TOTP 验证页
│   ├── terminal.html           # 终端页
│   ├── css/style.css
│   └── js/
│       ├── login.js
│       ├── totp.js
│       └── terminal.js
├── config.yaml                 # 配置文件（自动生成）
├── go.mod
└── Makefile
```

## 错误处理

| 场景 | 处理方式 |
|------|----------|
| SSH 连接失败 | 终端显示红色错误信息，提供"重连"按钮 |
| WebSocket 断开 | 自动重连 3 次，失败后显示提示 |
| TOTP 验证失败 | 显示错误，允许重试 |
| 配置缺失 | 首次启动自动生成默认配置 |
| 密码错误 | 统一提示"用户名或密码错误"，不区分具体原因 |

## 依赖

### Go 依赖

- `github.com/gorilla/websocket` - WebSocket 支持
- `github.com/pquerna/otp` - TOTP 生成/验证
- `golang.org/x/crypto` - SSH 客户端 + bcrypt
- `gopkg.in/yaml.v3` - YAML 配置解析

### 前端依赖

- `xterm.js` - 终端模拟器（通过 CDN 或内嵌）
- `qrcode.js` - TOTP 二维码显示（首次设置时）
