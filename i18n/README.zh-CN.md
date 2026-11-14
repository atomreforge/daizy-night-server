# Daizy Night

[English](../README.md) | [中文](README.zh-CN.md)

**Daizy Night** 是 ATOM Reforge 社区的账号与认证后端服务。

项目名 _Daizy Night_ 是 _The Hazy Night_ 的谐音。

> 说明：服务端处于早期阶段（`v0.7.2`），API 仍在演进。

## 功能特性

- **注册码注册** — 每次注册均需提供*注册码*（registercode）：由 Ed25519 签名、带过期校验的令牌，首次使用即被原子占用；每次占用均记录在案以供审计。
- **账号密码登录（legacy）** — 密码经 **argon2id** 哈希处理，不存储明文。
- **双 JWT** — 登录同时签发 *access token* 与 *refresh token*，均以 Ed25519（EdDSA）签名。refresh token 仅以摘要形式持久化于服务端，并支持吊销。
- **令牌轮换** — `POST /api/v1/refresh-access-token` 以 refresh token 换发全新的令牌对。刷新成功后，被使用的 refresh token 随即吊销，重放将得到 `401`。每台设备持有独立的 refresh token，一台设备刷新不影响其他设备的会话。
- **单设备登出** — `POST /api/v1/user/signout` 吊销当前会话的 refresh token，该操作幂等。
- **课表** — 面向单用户的周课表，采用整体替换语义，通过 `/api/v1/user/{username}/calendar` 上的 `GET`、`PUT`、`DELETE` 访问。
- **公共只读域** — 任何已认证用户均可读取任意用户的*脱敏*资料（不含 `email`、`github_id`、`github_login`）与课表，路径为 `/api/v1/public/user/{username}/...`。属主校验仅适用于自我访问路由。
- **健康检查** — `GET /api/v1/public/health/db` 报告数据库连通性。
- **认证与授权** — `echo-jwt` 中间件负责认证；角色控制区分 `user` 与 `admin`。`/api/v1/user/{username}/...` 下的自我访问路由校验路径段与认证身份一致（不一致返回 `403`）；实际生效的 `uid` 始终取自 JWT claims。
- **速率限制** — 按客户端 IP 维度的内存令牌桶；`enabled`、`rate`、`burst`、`expiresIn` 均可配置。超限返回 `429`。
- **请求体上限** — 请求体上限为 1 MiB，超限返回 `413`。
- **通知系统** — 通知读取端点已注册于 `/api/v1/public/notif/get/{single,batch,update}`（需认证）与 `/api/v1/admin/notif/post`（admin，占位）。`nid` 始终由服务端分配：严格正整数且单调递增（`1, 2, 3, ...`），永不为零、删除后亦不回收复用。通知的创建/更新/删除/撤回等管理操作已定义但尚未注册路由。
- **GitHub OAuth（预留）** — OAuth 注册与登录方式已在常量层定义但未实现，当前请求返回 `FeatureUnsupported`。

## 技术栈

- [Go](https://go.dev) 1.26
- [Echo](https://echo.labstack.com) v5 — HTTP 框架
- [GORM](https://gorm.io) + SQLite — ORM 与存储；`mattn/go-sqlite3` 驱动依赖 CGO
- [confx](https://github.com/atomreforge/confx)（基于 viper）— 配置加载
- [argon2id](https://github.com/alexedwards/argon2id) — 密码哈希
- [golang-jwt v5](https://github.com/golang-jwt/jwt) + [echo-jwt v5](https://github.com/labstack/echo-jwt) — JWT 签名与校验
- `crypto/ed25519` — JWT 与注册码使用的密钥对

## 相关仓库

- **服务端**（本仓库）：<https://github.com/atomreforge/daizy-night-server>
- **客户端**：<https://github.com/atomreforge/nilset>
- **测试用 CLI 客户端**：<https://github.com/Syerain/dnappcli>

## 快速开始

### 1. 前置条件

- Go 1.26 或更高版本
- C 工具链（`gcc` 或 `clang`）；SQLite 驱动 `mattn/go-sqlite3` 依赖 CGO

### 2. 获取代码

```bash
git clone https://github.com/atomreforge/daizy-night-server.git
cd daizy-night-server
```

### 3. 配置

```bash
cp config.example.yaml config.yaml
```

`security` 段整体为**必填**；任一密钥缺失将导致启动失败。所有密钥均为十六进制编码的 Ed25519 密钥：

| 键 | 用途 |
| --- | --- |
| `registercodeEnckey` / `registercodeDeckey` | 注册码签名 / 校验 |
| `passwordEnckey` / `passwordDeckey` | 密码层密钥（预留） |
| `jwtAccessTokenEnckey` / `jwtAccessTokenDeckey` | access token 签名 / 校验 |
| `jwtRefreshTokenEnckey` / `jwtRefreshTokenDeckey` | refresh token 签名 / 校验 |

Ed25519 密钥对可按如下方式生成：

```go
package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
)

func main() {
	pub, priv, _ := ed25519.GenerateKey(nil)
	fmt.Println("private (seed):", hex.EncodeToString(priv.Seed()))
	fmt.Println("private (full):", hex.EncodeToString(priv))
	fmt.Println("public:         ", hex.EncodeToString(pub))
}
```

> **测试配置** — 若工作目录或可执行文件旁存在 `config.test.yaml`，其优先级高于 `config.yaml`。该机制用于本地开发。

### 4. 运行

```bash
go mod tidy
go run ./cmd/main.go
```

默认监听 `127.0.0.1:4703`。SQLite 数据库文件在首次启动时自动创建并迁移。

### 5. 测试

```bash
go test ./...
```

### 6. 辅助脚本

`run`、`setup` 与 `build` 为面向 Windows / Git Bash 环境的 shell 脚本：

| 脚本 | 用途 |
| --- | --- |
| `run` | 以 `go run` 启动服务 |
| `setup` | 释放 4703 端口、删除 `./data.db` 后启动服务 |
| `build` | 交叉编译脚本；CGO 相关限制见脚本内注释 |

## 配置参考

| 配置段 | 字段 | 默认值 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| `main` | `isDebugMode` | `true` | 否 | 调试日志与行为 |
| `http` | `port` | `4703` | 否 | 监听端口 |
| `http` | `address` | `127.0.0.1` | 否 | 绑定地址 |
| `http.rateLimit` | `enabled` | `true` | 否 | 是否启用速率限制 |
| `http.rateLimit` | `rate` | `2` | 否 | 每秒补充的令牌数 |
| `http.rateLimit` | `burst` | `7` | 否 | 桶容量 |
| `http.rateLimit` | `expiresIn` | `3m` | 否 | 条目过期时间 |
| `database` | `isDebugMode` | `false` | 否 | GORM 详细日志 |
| `database` | `DSN` | `./data.db` | **是** | SQLite 文件路径 |
| `log` | `isColored` | `false` | 否 | 彩色日志输出 |
| `security` | `*Enckey` / `*Deckey` | — | **是** | 十六进制 Ed25519 密钥对（见上文） |
| `security` | `jwtAccessTokenExpireTime` | `15m` | **是** | access token 有效期 |
| `security` | `jwtRefreshTokenExpireTime` | `168h` | **是** | refresh token 有效期 |
| `security` | `jwtRevokedTokensRetainTime` | `72h` | **是** | 已吊销 token 记录的保留时长 |

支持的时间单位：`h`、`m`、`s`、`ms`、`us`、`ns`。

## API

所有路由以 `/api/v1` 为前缀；请求与响应体均为 JSON（`Content-Type: application/json`）。完整规范（含字段约束与错误语义）维护于 [docs/api.v1.md](docs/api.v1.md)。

### 认证模型

- 公开端点无需认证。
- 其余端点均要求请求头 `Authorization: Bearer <access_token>`。
- `/api/v1/user/{username}/...` 下的自我访问路由实施属主校验：路径段 `{username}` 必须与 JWT claims 中的认证身份一致，否则返回 `403`。实际生效的 `uid` 始终取自 JWT claims，不接受来自请求体的用户标识。
- `/api/v1/public/user/{username}/...` 构成只读公共域：任何已认证用户可读取任意用户的数据，不做属主校验；响应在适用处脱敏。
- `/api/v1/admin` 下的路由要求 `admin` 角色。

### 端点

| 方法 | 路径 | 访问要求 | 说明 |
| --- | --- | --- | --- |
| `POST` | `/api/v1/register` | 公开 | 创建账号；须提供有效注册码 |
| `POST` | `/api/v1/login` | 公开 | 以凭据换取令牌对 |
| `POST` | `/api/v1/refresh-access-token` | 公开 | 以 refresh token 换发新令牌对（轮换） |
| `POST` | `/api/v1/user/signout` | 需认证 | 吊销当前会话的 refresh token；幂等 |
| `GET` | `/api/v1/user/{username}/info` | 需认证（本人） | 认证用户的完整资料 |
| `GET` | `/api/v1/user/{username}/calendar` | 需认证（本人） | 当前用户课表及全部条目 |
| `PUT` | `/api/v1/user/{username}/calendar` | 需认证（本人） | 整体替换；首次调用即创建；幂等 |
| `DELETE` | `/api/v1/user/{username}/calendar` | 需认证（本人） | 删除课表及全部条目；幂等 |
| `GET` | `/api/v1/public/user/{username}/info` | 需认证 | 任意用户的脱敏资料 |
| `GET` | `/api/v1/public/user/{username}/calendar` | 需认证 | 任意用户课表（剥离 `roaming` 字段） |
| `GET` | `/api/v1/public/health/db` | 需认证 | 数据库连通性检查 |
| `POST` | `/api/v1/admin/sudo` | `admin` | 占位管理操作 |
| `GET` | `/api/v1/public/notif/get/single` | 需认证 | 按 `nid` 获取单条通知（附服务端最新 `nid`） |
| `GET` | `/api/v1/public/notif/get/batch` | 需认证 | 闭区间 `[nid_begin, nid_end]` 内的通知，按 `nid` 升序 |
| `GET` | `/api/v1/public/notif/get/update` | 需认证 | 严格大于 `client_latest_nid` 的增量通知 |
| `POST` | `/api/v1/admin/notif/post` | `admin` | 将通知标记为已发布（占位；当前返回 `FeatureUnsupported`） |

通知管理操作 `notif/update`、`notif/delete`、`notif/withdraw` 已定义但尚未注册。`nid` 始终由服务端分配：严格正整数且单调递增（`1, 2, 3, ...`），永不为零，删除后不回收复用。

### 错误响应

所有错误响应统一为 `{"message": string}`。业务错误携带可读描述；非业务错误回退为 `internal server error`。超过速率限制返回 `429`；请求体超过 1 MiB 返回 `413`。

### 示例

注册（`POST /api/v1/register`）：

```json
{
  "registerway": "legacy",
  "username": "alice",
  "nickname": "Alice",
  "password": "secret",
  "registercode": "<payload_hex>.<sig_hex>"
}
```

登录（`POST /api/v1/login`）：

```json
{
  "loginway": "legacy",
  "username": "alice",
  "password": "secret"
}
```

`200 OK`：

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<jwt>"
}
```

## 项目结构

```
cmd/                     入口与服务组装
docs/                    API 规范（api.v1.md）
internal/
  abstract/interface/    各 provider 契约（crypto、db、repo、service）
  api/v1/                请求 / 响应 DTO
  config/                配置结构体与加载器（confx）
  consts/                角色、注册/登录方式、域、等级、日志表达式
  crypto/                Ed25519、JWT 与注册码 provider
  dbware/                GORM 连接与仓储
  errs/                  统一错误体系
  handler/               HTTP handler（薄层）
  middleware/            injector、速率限制、JWT 认证、角色控制、域标记
  model/                 GORM 模型与共享 DTO
  router/                Echo 路由装配
  service/               业务逻辑
  utils/                 hash、hex、logger、rand、trace
etc/                     内部笔记
i18n/                    本地化 README
```

分层调用链：`router` → `handler` → `service` → `dbware` → `model`；`middleware`、`crypto`、`errs`、`utils` 为横切包。单元测试与所属包同目录存放（`*_test.go`）。

## 更新日志

- **v0.7.2** — 课表模型与 PUT 逻辑修复
- **v0.7.1** — 维护性修改；通知端点暂时撤出路由
- **v0.7.0** — 通知与事件子系统（初始）；严重级别；API 与路由重构；公共只读域
- **v0.6.3** — `calendar_items` 表自动创建；AutoMigrate 对 has-many 子表的修复
- **v0.6.2** — 数据库健康检查服务；请求体上限；令牌清理与错误码修正
- **v0.6.1** — API 文档；目录结构更新；配置字段调整
- **v0.6.0** — 路由重组；课表 CRUD
- **v0.5.6** — refresh-access-token 端点（轮换、多设备支持）；单设备登出；session 占位字段；独立 API 文档
- **v0.5.5** — 注册码占用修复；UID 生成；错误消息修订；默认限流值；启动脚本
- **v0.5.4** — 随机 `UserID` 身份；注册码原子占用；适应性收尾
- **v0.5.3** — 注册码逻辑修复
- **v0.5.2** — 配置依赖注入修复
- **v0.5.0** — 异常体系；trace 体系
- **v0.4.1** — 认证中间件与 `echo-jwt`
- **v0.4.0** — 迁移至 Echo v5；错误处理重构
- **v0.3.x** — 登录流程
- **v0.1.0** — 初始结构
