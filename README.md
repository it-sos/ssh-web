# ssh-web

基于 Web 的 SSH 终端应用。用户通过浏览器访问，经过用户名/密码 + TOTP 二次认证后，自动连接到配置的 SSH 主机。

## 文档索引

### 配置参考

| 文档 | 说明 |
|------|------|
| [配置说明](docs/config.md) | `config.yaml` 全字段参考及示例 |
| [环境变量](docs/env.md) | `SSH_WEB_*` 环境变量覆盖参考 |
| [密码加密](docs/password-encryption.md) | AES-256-GCM 加密机制及使用方式 |

### 开发指南

| 文档 | 说明 |
|------|------|
| [Agent 须知](docs/agents.md) | Graphify 知识图谱使用规则 |
| [CLAUDE.md](CLAUDE.md) | 项目架构、构建、API 及测试 |

### 设计文档

| 文档 | 说明 |
|------|------|
| [Web SSH 终端设计](docs/superpowers/specs/2026-04-12-web-ssh-terminal-design.md) | 整体架构设计文档 |

## 快速开始

```bash
go build -o ssh-web ./cmd/server
./ssh-web
```

更多信息见 [CLAUDE.md](CLAUDE.md)。
