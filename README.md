# Daizy Night

[English](README.md) | [中文](/i18n/README.zh-CN.md)

**Daizy Night** is the account and authentication backend service for the ATOM Reforge community.

The project name _Daizy Night_ is a homophone of _The Hazy Night_.

> Note: the server is still in an early stage (`v0.5.4`) — the API surface keeps evolving.

## Features

- **Registercode-based registration** — every sign-up requires a *Registercode*: an Ed25519-signed, expiry-checked token that is claimed **atomically** on first use. Every claim is recorded for audit.
- **Legacy password login** — passwords are hashed with **argon2id** and never stored in plaintext.
- **Dual JWT** — login issues an *access token* and a *refresh token*, both signed with Ed25519 (EdDSA). Refresh tokens are persisted server-side and support revocation.
- **Auth & authorization** — `echo-jwt` middleware for authentication, plus role control (`user` / `admin`).
- **Rate limiting** — per-client-IP, in-memory token bucket; `rate`, `burst` and `expiresIn` are all configurable.
- **GitHub OAuth (reserved)** — OAuth register/login ways are defined in the constant layer, but not implemented yet (currently return `FeatureUnsupported`).

## Tech Stack

- [Go](https://go.dev) 1.26
- [Echo](https://echo.labstack.com) v5 — HTTP framework
- [GORM](https://gorm.io) + SQLite — ORM and storage
- [confx](https://github.com/atomreforge/confx) (viper-based) — configuration loading
- [argon2id](https://github.com/alexedwards/argon2id) — password hashing
- [golang-jwt v5](https://github.com/golang-jwt/jwt) + [echo-jwt v5](https://github.com/labstack/echo-jwt) — JWT signing and verification
- `crypto/ed25519` — key pairs for JWTs and registercodes

## Repositories

- **Server** (this repository): <https://github.com/atomreforge/daizy-night-server>
- **Scheduled client**: <https://github.com/atomreforge/daizy-night-app>
- **CLI test client**: <https://github.com/Syerain/dnappcli>

## Quick Start

### 1. Get the code

```bash
git clone https://github.com/atomreforge/daizy-night-server.git
cd daizy-night-server
```

### 2. Configure

```bash
cp config.example.yaml config.yaml
```

The whole `security` section is **required** — the server refuses to start if any key is missing. All values are hex-encoded Ed25519 keys:

| Key | Purpose |
| --- | --- |
| `registercodeEnckey` / `registercodeDeckey` | Sign / verify registercodes |
| `passwordEnckey` / `passwordDeckey` | Reserved password-layer keys |
| `jwtAccessTokenEnckey` / `jwtAccessTokenDeckey` | Sign / verify access tokens |
| `jwtRefreshTokenEnckey` / `jwtRefreshTokenDeckey` | Sign / verify refresh tokens |

Generate an Ed25519 key pair, for example:

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

> **Test profile** — if a `config.test.yaml` exists next to the working directory or the executable, it is preferred over `config.yaml`. Handy for local development.

### 3. Run

```bash
go mod tidy
go run ./cmd/main.go
```

By default the server listens on `127.0.0.1:4703`. The SQLite database is created and auto-migrated on the first start.

## Configuration Reference

| Section | Field | Default | Required | Description |
| --- | --- | --- | --- | --- |
| `main` | `isDebugMode` | `true` | no | Debug logging & behavior |
| `http` | `port` | `4703` | no | Listen port |
| `http` | `address` | `127.0.0.1` | no | Bind address |
| `http.rateLimit` | `enabled` | `true` | no | Toggle the rate limiter |
| `http.rateLimit` | `rate` | `10` | no | Tokens refilled per second |
| `http.rateLimit` | `burst` | `7` | no | Bucket capacity |
| `http.rateLimit` | `expiresIn` | `3m` | no | Entry expiry |
| `database` | `isDebugMode` | `false` | no | Verbose GORM logging |
| `database` | `DSN` | `./data.db` | **yes** | SQLite file path |
| `log` | `isColored` | `false` | no | Colored log output |
| `security` | `*Enckey` / `*Deckey` | — | **yes** | Hex Ed25519 key pairs (see above) |
| `security` | `jwtAccessTokenExpireTime` | `15m` | **yes** | Access-token lifetime |
| `security` | `jwtRefreshTokenExpireTime` | `168h` | **yes** | Refresh-token lifetime |
| `security` | `jwtRevokedTokensRetainTime` | `72h` | **yes** | How long revoked tokens are kept |

Supported duration units: `h`, `m`, `s`, `ms`, `us`, `ns`.

## API

All routes are prefixed with `/api/v1`.

### `POST /api/v1/register`

Public. Creates an account; a valid registercode is required.

```json
{
  "registerway": "legacy",
  "username": "alice",
  "nickname": "Alice",
  "password": "secret",
  "registercode": "<payload_hex>.<sig_hex>"
}
```

`200 OK`:

```json
{ "message": "ok" }
```

### `POST /api/v1/login`

Public. Exchanges credentials for a JWT pair.

```json
{
  "loginway": "legacy",
  "username": "alice",
  "password": "secret"
}
```

`200 OK`:

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<jwt>"
}
```

### `GET /api/v1/user/{username}/info`

Authenticated — requires `Authorization: Bearer <access_token>`. Returns the current user's info. The `{username}` path segment must match the authenticated identity (JWT claims), otherwise `403`.

`200 OK`:

```json
{
  "uid": 12345,
  "username": "alice",
  "nickname": "Alice",
  "email": "",
  "register_time": "2026-08-17T12:00:00Z",
  "role": "user",
  "github_id": null,
  "github_login": null
}
```

### `POST /api/v1/admin/sudo`

Authenticated with the `admin` role only. Placeholder admin operation.

## Project Structure

```
cmd/                     entry point & server assembly
internal/
  abstract/interface/    provider contracts (crypto, db, repo, service)
  api/v1/                request / response DTOs
  config/                config struct & loader (confx)
  consts/                roles, register/login ways, log expressions
  crypto/                Ed25519, JWT & registercode providers
  dbware/                GORM connection & repositories
  errs/                  centralized error system
  handler/               HTTP handlers (thin layer)
  middleware/            injector, rate limiter, JWT auth, role control
  model/                 GORM models & shared DTOs
  router/                Echo route wiring
  service/               business logic
  utils/                 hash, hex, logger, rand, trace
etc/                     internal notes & docs
```

Layering: `router` → `handler` → `service` → `dbware` → `model`, with `middleware`, `crypto`, `errs` and `utils` as cross-cutting packages.

## Changelog

- **v0.5.4** — random `UserID` identity; atomic registercode claim; adaptive chores
- **v0.5.3** — registercode logic fixes
- **v0.5.2** — config dependency-injection hotfix
- **v0.5.0** — exception system; trace system
- **v0.4.1** — auth middleware & `echo-jwt`
- **v0.4.0** — migration to Echo v5; error-handling rework
- **v0.3.x** — login flow
- **v0.1.0** — initial structure