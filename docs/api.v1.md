# Daizy Night API v1

## 概览

- Base URL: `http://127.0.0.1:4703`，由 `config.yaml` 的 `http.address` / `http.port` 决定
- 所有路由前缀 `/api/v1`，请求与响应体均为 JSON（`Content-Type: application/json`）
- 认证方式：`Authorization: Bearer <access_token>`，JWT 签名算法 EdDSA（ed25519）
- Token 有效期：access token 15m，refresh token 168h（`config.yaml` 的 `security` 段可调）

## 通用行为

### 错误响应

所有错误响应统一为 `{"message": string}`：

| 场景 | 状态码 | message |
|---|---|---|
| 业务错误（校验失败、凭据错误、注册码无效、token 过期/验签失败等） | 400/401/404 | 可读描述，如 `failure in user login; details::...` |
| 参数解析失败、DB 错误等非业务错误 | 视错误而定 | 固定回退为 `internal server error` |
| 未捕获的未知错误 | 500 | `internal server error` |

### 限流

按客户端 IP 全局限流，超限返回 429 `{"message": "too many requests"}`。
默认：速率 2 req/s，突发 7，窗口 3m（`config.yaml` 的 `http.rateLimit`）。

### 请求体上限

所有端点的请求体上限为 1 MiB（同时依据 `Content-Length` 与实际读取长度判定）；超限返回 `413`，message 为固定回退文本 `internal server error`。

### 用户资源路由与属主校验

用户级端点（`info`、`calendar`）统一采用 `/api/v1/user/{username}/...` 路径形态：

- `{username}` 路径段仅用于**属主校验**：必须与 access token claims 中的 `username` 完全一致，否则返回 `403`
- 数据定位所用的 `uid` 永远取自 JWT claims；请求体不携带（也不接受）任何用户标识字段
- 通用错误码：`401` access token 缺失、无效或过期；`403` 属主校验失败（下文各端点列出的 `403` 均指此语义）
- 公共域例外：`GET /api/v1/public/user/{username}/calendar` 已认证后可读取**任意** `{username}` 的课表，不做属主校验（见对应端点）；公共写操作已废弃
- 公共域 `/api/v1/public/user/{username}/info` 为预留端点：披露语义（email、github 字段）未定稿，当前统一返回 `403`

## 端点

### POST /api/v1/register

注册新用户。公开端点，无需认证。

请求体：

| 字段 | 类型 | 约束 |
|---|---|---|
| `registerway` | string | `legacy`（唯一支持，`oauth-github` 未实现） |
| `username` | string | 1-15 字符，仅允许 `a-zA-Z0-9` |
| `nickname` | string | 1-15 字符 |
| `password` | string | 6-128 字符 |
| `registercode` | string | 注册码原文，格式 `<hex载荷>.<hex签名>`；签名段恒为 128 个十六进制字符（Ed25519），载荷段非空且 ≤ 2048 个十六进制字符 |

```json
{
  "registerway": "legacy",
  "username": "alice",
  "nickname": "Alice",
  "password": "secret123",
  "registercode": "ab12....cd.ef34....89"
}
```

响应：

- `200` `{"message": "ok"}`
- `400` 参数校验失败、注册码无效/过期、用户名已被占用，或注册码已被使用（`failure in ...`）

### POST /api/v1/login

登录并签发 token 对。公开端点。

请求体：

| 字段 | 类型 | 说明 |
|---|---|---|
| `loginway` | string | `legacy`（唯一支持，`oauth-github` 返回 400） |
| `username` | string | 用户名 |
| `password` | string | 密码 |
| `entrycode` | string | 预留字段，legacy 方式不使用 |

```json
{
  "loginway": "legacy",
  "username": "alice",
  "password": "secret123"
}
```

响应：

- `200`

```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ..."
}
```

- `400` 用户名或密码错误（`failure in user login; details::...`），或参数校验失败

### POST /api/v1/refresh-access-token

使用 refresh token 换发全新 token 对。公开端点，无需认证头。

请求体：

| 字段 | 类型 | 说明 |
|---|---|---|
| `refresh_token` | string | 登录或上次刷新获得的 refresh token |

```json
{
  "refresh_token": "eyJ..."
}
```

响应：

- `200`，返回全新 token 对

```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ..."
}
```

- `401` refresh token 无效：过期、伪造/验签失败、格式非法、未在库中找到，或已被吊销（含已使用 token 的重放）。统一业务错误格式 `failure in user login; details::...`：过期为 `token expired`，伪造/格式非法为 `invalid token`，未在库中找到/已吊销为 `incorrect user login params`
- `400` 请求体不是合法 JSON，或 `refresh_token` 字段缺失
- `500` 未预期的服务端错误（正常分支不应出现）

轮换语义：

- 每次成功刷新都会吊销被使用的这一枚 refresh token，并签发全新 token 对；客户端必须整体覆盖保存，旧 refresh token 立即失效
- 吊销精确到单枚 token：多设备各自持有独立的 refresh token，一台设备刷新不影响其他设备的登录态

### GET /api/v1/user/{username}/info

获取当前认证用户的信息。需要认证。路径中的 `{username}` 必须与认证身份（JWT claims）一致，否则 `403`。

请求头：`Authorization: Bearer <access_token>`

响应：

- `200`

```json
{
  "uid": 1527277,
  "username": "alice",
  "nickname": "Alice",
  "email": "",
  "register_time": "2026-08-31T12:00:00.000000Z",
  "role": "user",
  "github_id": null,
  "github_login": null
}
```

- `401` access token 缺失、无效或过期
- `403` 路径中的用户名与认证身份不一致
- `404` token 有效但用户记录不存在

### GET /api/v1/user/{username}/calendar

获取当前认证用户的课表（含全部课程条目，按 `weekday`、`start_min` 升序）。需要认证，属主校验同上。

请求头：`Authorization: Bearer <access_token>`

响应：

- `200`

```json
{
  "roaming": { "description": "", "annotation": "" },
  "uid": 1527277,
  "calendar_id": 7362514,
  "records": [
    { "calendar_id": 7362514, "weekday": 1, "start_min": 480, "end_min": 570, "title": "数学" }
  ]
}
```

`records` 按 `weekday`、`start_min` 升序返回；课表存在但无课程条目时返回 `[]`（不会是 `null`）。`roaming`（`description` / `annotation`）为预留的描述性字段，当前恒为空字符串。

- `401` access token 缺失、无效或过期
- `403` 路径中的用户名与认证身份不一致
- `404` 尚未创建课表

### PUT /api/v1/user/{username}/calendar

全量替换当前用户的课表；首次调用即创建，语义幂等。需要认证，属主校验同上。`uid` 永不出现在请求体中，始终取自 JWT claims。

请求体（`records` 可为空数组，表示清空课表；上限 200 条）：

```json
{
  "records": [
    { "weekday": 1, "start_min": 480, "end_min": 570, "title": "数学" }
  ]
}
```

| 字段 | 约束 |
| --- | --- |
| `records[].weekday` | 整数 `0`–`6`（`0` = 周日） |
| `records[].start_min` | 整数 `≥ 0`，距当天 00:00 的分钟数 |
| `records[].end_min` | 整数，满足 `start_min < end_min ≤ 1440` |
| `records[].title` | 非空，1–255 字符 |

> 注：`start_min` / `end_min` 在服务端为无符号整数，请求中出现负数字面量时会在 JSON 解析阶段直接拒绝（`400`，message 为固定回退文本 `internal server error`，而非业务错误文案）。

响应：

- `200` `{"message": "ok"}`
- `400` 参数校验失败
- `401` / `403` 同上

### DELETE /api/v1/user/{username}/calendar

删除当前用户的课表（连同全部课程条目）。需要认证，属主校验同上。幂等：课表不存在同样返回 `200`。

响应：

- `200` `{"message": "ok"}`
- `401` / `403` 同上

### GET /api/v1/public/user/{username}/calendar

公共只读域：任何已认证用户可查看任意用户的课表，不做属主校验。需要认证。

请求头：`Authorization: Bearer <access_token>`

响应：`200`

```json
{
  "uid": 1527277,
  "calendar_id": 7362514,
  "records": [
    { "calendar_id": 7362514, "weekday": 1, "start_min": 480, "end_min": 570, "title": "数学" }
  ]
}
```

`records` 排序规则同自我访问端点（`weekday`、`start_min` 升序）。

- `401` access token 缺失、无效或过期
- `404` 用户不存在，或该用户尚未创建课表

### GET /api/v1/public/health/db

数据库健康检查。公开端点，无需认证。

响应：

- `200` `{"message": "ok"}`（数据库可正常连接）
- `500` `{"message": "error with db."}`（数据库连接失败）

### GET /api/v1/public/notif/get

通知功能占位端点（init 阶段）。公开端点，无需认证。

响应：

- `200` `{"message": "ok"}`（占位实现，暂无实际内容）

### POST /api/v1/user/signout

退出登录：吊销当前会话的 refresh token。需要认证。

语义约定：

- `session` 字段为未来会话标识符预留：`session` 未指明时，后端默认吊销"本会话"；当前尚无会话标识符，"本会话"由请求体中的 `refresh_token` 指称，因此该字段必填
- 幂等：token 未知、无效（含伪造/过期）或已被吊销时同样返回 `200`，重复退出不会失败
- 吊销后该 refresh token 立即无法再刷新（重放刷新得到 `401`）；已签发的 access token 存活至其自身过期（≤15m），期间无法吊销（无黑名单机制）
- 单设备退出：不影响该用户其他设备的登录态

请求头：`Authorization: Bearer <access_token>`

请求体：

| 字段 | 类型 | 说明 |
|---|---|---|
| `refresh_token` | string | 当前会话的 refresh token（"本会话"的指称，必填） |
| `session` | string[] | 预留字段，当前必须省略或为空数组；非空返回 400 |

```json
{
  "refresh_token": "eyJ...",
  "session": []
}
```

响应：

- `200` `{"message": "ok"}`（含幂等重入）
- `400` 请求体不是合法 JSON、`refresh_token` 缺失，或 `session` 非空（按会话退出暂未支持）
- `401` access token 缺失、无效或过期

注意 access token 在 15min 过期后，无法通过鉴权，此时应当自行用 refresh token 取得新的 access token 并重试。

### POST /api/v1/admin/sudo

管理端点，占位实现。需要认证，且要求 `admin` 角色。

响应：

- `200` 空响应体
- `401` 认证失败
- `403` 已认证但角色非 `admin`

### POST /api/v1/admin/notif/post

通知功能占位端点（init 阶段）。需要认证，且要求 `admin` 角色。

响应：

- `200` `{"message": "ok"}`（占位实现，暂无实际内容）
- `401` 认证失败
- `403` 已认证但角色非 `admin`

## Token 语义补充

- access token 载荷：`uid`、`username`、`role`、`iat`、`exp`；角色取值 `user` / `admin`
- refresh token 载荷另含随机 `jti`；服务端只存哈希，不存原文
- 已吊销的 token 记录保留 72h，超期由服务端自动清理；自然过期且超过「有效期 + 保留窗口」的记录同样会被自动清理
- 客户端保存的 refresh token 一旦被使用即作废，重放会得到 401，此时应引导用户重新登录
