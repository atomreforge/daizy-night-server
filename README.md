# Daizy Night

[English](README.md) | [中文](i18n/README.zh-CN.md)

**Daizy Night** is the account and authentication backend service of the ATOM Reforge community.

The project name _Daizy Night_ is a homophone of _The Hazy Night_.

> Note: the server is at an early stage (`v0.7.2`); the API surface is still evolving.

## Features

- **Registercode-based registration** — every sign-up requires a *registercode*: an Ed25519-signed, expiry-checked token that is claimed atomically on first use. Every claim is recorded for audit.
- **Legacy password login** — passwords are hashed with **argon2id**; plaintext passwords are never stored.
- **Dual JWT** — a login issues an *access token* and a *refresh token*, both signed with Ed25519 (EdDSA). Refresh tokens are persisted server-side (digest only) and support revocation.
- **Token rotation** — `POST /api/v1/refresh-access-token` exchanges a refresh token for an entirely new token pair. The presented refresh token is revoked upon a successful refresh; replaying it yields `401`. Each device holds an independent refresh token, therefore a refresh on one device does not affect the sessions of other devices.
- **Single-device signout** — `POST /api/v1/user/signout` revokes the refresh token of the current session. The operation is idempotent.
- **Calendar** — a per-user weekly schedule with full-replacement semantics, available through `GET`, `PUT` and `DELETE` on `/api/v1/user/{username}/calendar`.
- **Public read-only domain** — any authenticated user may read the *sanitized* profile (without `email`, `github_id`, `github_login`) and the calendar of any user under `/api/v1/public/user/{username}/...`. Ownership checks apply to self-access routes only.
- **Health check** — `GET /api/v1/public/health/db` reports database connectivity.
- **Authentication & authorization** — `echo-jwt` middleware performs authentication; role control distinguishes `user` and `admin`. Self-access routes under `/api/v1/user/{username}/...` verify that the path segment matches the authenticated identity (`403` otherwise); the acting `uid` is always taken from the JWT claims.
- **Rate limiting** — a per-client-IP, in-memory token bucket; `enabled`, `rate`, `burst` and `expiresIn` are configurable. Exceeding the limit yields `429`.
- **Request body limit** — request bodies are capped at 1 MiB; larger requests are rejected with `413`.
- **Notification system** — notification read endpoints registered under `/api/v1/public/notif/get/{single,batch,update}` (authenticated) and `/api/v1/admin/notif/post` (admin, placeholder). `nid` is always server-assigned: a strictly positive, monotonically increasing integer (`1, 2, 3, ...`) that is never zero and never reused — the nid of a removed notification stays retired. Notification create/update/delete/withdraw admin operations are defined but not registered yet.
- **GitHub OAuth (reserved)** — OAuth register and login ways are defined in the constant layer but not implemented; requests currently return `FeatureUnsupported`.

## Tech Stack

- [Go](https://go.dev) 1.26
- [Echo](https://echo.labstack.com) v5 — HTTP framework
- [GORM](https://gorm.io) + SQLite — ORM and storage; the `mattn/go-sqlite3` driver requires CGO
- [confx](https://github.com/atomreforge/confx) (viper-based) — configuration loading
- [argon2id](https://github.com/alexedwards/argon2id) — password hashing
- [golang-jwt v5](https://github.com/golang-jwt/jwt) + [echo-jwt v5](https://github.com/labstack/echo-jwt) — JWT signing and verification
- `crypto/ed25519` — key pairs for JWTs and registercodes

## Repositories

- **Server** (this repository): <https://github.com/atomreforge/daizy-night-server>
- **Android App**: <https://github.com/atomreforge/nilset>
- **CLI test client**: <https://github.com/Syerain/dnappcli>

## Quick Start

### 1. Prerequisites

- Go 1.26 or later
- A C toolchain (`gcc` or `clang`); the SQLite driver `mattn/go-sqlite3` requires CGO

### 2. Get the code

```bash
git clone https://github.com/atomreforge/daizy-night-server.git
cd daizy-night-server
```

### 3. Configure

```bash
cp config.example.yaml config.yaml
```

The `security` section is **required**; the server refuses to start if any key is missing. All keys are hex-encoded Ed25519 keys:

| Key | Purpose |
| --- | --- |
| `registercodeEnckey` / `registercodeDeckey` | Sign / verify registercodes |
| `passwordEnckey` / `passwordDeckey` | Reserved password-layer keys |
| `jwtAccessTokenEnckey` / `jwtAccessTokenDeckey` | Sign / verify access tokens |
| `jwtRefreshTokenEnckey` / `jwtRefreshTokenDeckey` | Sign / verify refresh tokens |

An Ed25519 key pair may be generated as follows:

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

> **Test profile** — if a `config.test.yaml` exists next to the working directory or the executable, it takes precedence over `config.yaml`. This is intended for local development.

### 4. Run

```bash
go mod tidy
go run ./cmd/main.go
```

By default the server listens on `127.0.0.1:4703`. The SQLite database file is created and migrated automatically on first start.

### 5. Tests

```bash
go test ./...
```

### 6. Helper scripts

`run`, `setup` and `build` are shell scripts written for Windows / Git Bash environments:

| Script | Purpose |
| --- | --- |
| `run` | Starts the server via `go run` |
| `setup` | Frees port 4703, removes `./data.db`, then starts the server |
| `build` | Cross-compilation script; see the notes inside regarding CGO limitations |

## Configuration Reference

| Section | Field | Default | Required | Description |
| --- | --- | --- | --- | --- |
| `main` | `isDebugMode` | `true` | no | Debug logging and behavior |
| `http` | `port` | `4703` | no | Listening port |
| `http` | `address` | `127.0.0.1` | no | Bind address |
| `http.rateLimit` | `enabled` | `true` | no | Enables rate limiting |
| `http.rateLimit` | `rate` | `2` | no | Tokens refilled per second |
| `http.rateLimit` | `burst` | `7` | no | Bucket capacity |
| `http.rateLimit` | `expiresIn` | `3m` | no | Entry expiry |
| `database` | `isDebugMode` | `false` | no | Verbose GORM logging |
| `database` | `DSN` | `./data.db` | **yes** | SQLite file path |
| `log` | `isColored` | `false` | no | Colored log output |
| `security` | `*Enckey` / `*Deckey` | — | **yes** | Hex Ed25519 key pairs (see above) |
| `security` | `jwtAccessTokenExpireTime` | `15m` | **yes** | Access-token lifetime |
| `security` | `jwtRefreshTokenExpireTime` | `168h` | **yes** | Refresh-token lifetime |
| `security` | `jwtRevokedTokensRetainTime` | `72h` | **yes** | Retention period of revoked-token records |

Supported duration units: `h`, `m`, `s`, `ms`, `us`, `ns`.

## API

All routes are prefixed with `/api/v1`; request and response bodies are JSON (`Content-Type: application/json`). The complete specification, including field constraints and error semantics, is maintained in [docs/api.v1.md](docs/api.v1.md).

### Authentication model

- Public endpoints require no authentication.
- All other endpoints require the header `Authorization: Bearer <access_token>`.
- Self-access routes under `/api/v1/user/{username}/...` enforce an ownership check: the `{username}` path segment must equal the authenticated identity from the JWT claims, otherwise `403` is returned. The acting `uid` is always taken from the JWT claims and is never accepted from a request body.
- Routes under `/api/v1/public/user/{username}/...` constitute a read-only public domain: any authenticated user may read any user's data without an ownership check. Responses are sanitized where applicable.
- Routes under `/api/v1/admin` require the `admin` role.

### Endpoints

| Method | Path | Access | Description |
| --- | --- | --- | --- |
| `POST` | `/api/v1/register` | public | Creates an account; a valid registercode is required |
| `POST` | `/api/v1/login` | public | Exchanges credentials for a token pair |
| `POST` | `/api/v1/refresh-access-token` | public | Exchanges a refresh token for a new token pair (rotation) |
| `POST` | `/api/v1/user/signout` | authenticated | Revokes the current session's refresh token; idempotent |
| `GET` | `/api/v1/user/{username}/info` | authenticated (self) | Full profile of the authenticated user |
| `GET` | `/api/v1/user/{username}/calendar` | authenticated (self) | The user's calendar with all records |
| `PUT` | `/api/v1/user/{username}/calendar` | authenticated (self) | Full replacement; creates the calendar on first use; idempotent |
| `DELETE` | `/api/v1/user/{username}/calendar` | authenticated (self) | Deletes the calendar with all records; idempotent |
| `GET` | `/api/v1/public/user/{username}/info` | authenticated | Sanitized profile of any user |
| `GET` | `/api/v1/public/user/{username}/calendar` | authenticated | Calendar of any user, `roaming` fields stripped |
| `GET` | `/api/v1/public/health/db` | authenticated | Database connectivity check |
| `POST` | `/api/v1/admin/sudo` | `admin` | Placeholder administrative operation |
| `GET` | `/api/v1/public/notif/get/single` | authenticated | A single notification by `nid`, with the repo's latest nid |
| `GET` | `/api/v1/public/notif/get/batch` | authenticated | Notifications in the closed `[nid_begin, nid_end]` range, ascending |
| `GET` | `/api/v1/public/notif/get/update` | authenticated | Delta strictly after the client's `client_latest_nid` |
| `POST` | `/api/v1/admin/notif/post` | `admin` | Marks a notification as posted (placeholder; currently `FeatureUnsupported`) |

Notification admin operations `notif/update`, `notif/delete` and `notif/withdraw` are defined but not registered yet. `nid` is always server-assigned: strictly positive, monotonically increasing (`1, 2, 3, ...`), never zero, never reused after removal.

### Error responses

Every error response has the form `{"message": string}`. Business errors carry a readable description; non-business failures fall back to `internal server error`. Rate-limit exceedance yields `429`; request bodies exceeding 1 MiB are rejected with `413`.

### Examples

Registration (`POST /api/v1/register`):

```json
{
  "registerway": "legacy",
  "username": "alice",
  "nickname": "Alice",
  "password": "secret",
  "registercode": "<payload_hex>.<sig_hex>"
}
```

Login (`POST /api/v1/login`):

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

## Project Structure

```
cmd/                     entry point & server assembly
docs/                    API specification (api.v1.md)
internal/
  abstract/interface/    provider contracts (crypto, db, repo, service)
  api/v1/                request / response DTOs
  config/                config struct & loader (confx)
  consts/                roles, register/login ways, domains, levels, log expressions
  crypto/                Ed25519, JWT & registercode providers
  dbware/                GORM connection & repositories
  errs/                  centralized error system
  handler/               HTTP handlers (thin layer)
  middleware/            injector, rate limiter, JWT auth, role control, domain marking
  model/                 GORM models & shared DTOs
  router/                Echo route wiring
  service/               business logic
  utils/                 hash, hex, logger, rand, trace
etc/                     internal notes
i18n/                    localized README files
```

Layering: `router` → `handler` → `service` → `dbware` → `model`, with `middleware`, `crypto`, `errs` and `utils` as cross-cutting packages. Unit tests are located next to the packages they cover (`*_test.go`).

## Changelog

- **v0.7.2** — calendar model & PUT logic fixes
- **v0.7.1** — maintenance; notification endpoints temporarily withdrawn from the router
- **v0.7.0** — notification & event subsystems (initial); severity levels; API and router refactor; public read-only domain
- **v0.6.3** — automatic creation of `calendar_items`; AutoMigrate fix for has-many children
- **v0.6.2** — database health-check service; request-body limit; token-pruning and error-code fixes
- **v0.6.1** — API documentation; directory structure update; configuration field transfer
- **v0.6.0** — router reorganization; calendar CRUD
- **v0.5.6** — refresh-access-token endpoint (rotation, multi-device support); single-device signout; session placeholder field; standalone API documentation
- **v0.5.5** — registercode claim fixes; UID generation; error-message rework; default rate-limit values; launch scripts
- **v0.5.4** — random `UserID` identity; atomic registercode claim; adaptive chores
- **v0.5.3** — registercode logic fixes
- **v0.5.2** — config dependency-injection hotfix
- **v0.5.0** — exception system; trace system
- **v0.4.1** — auth middleware & `echo-jwt`
- **v0.4.0** — migration to Echo v5; error-handling rework
- **v0.3.x** — login flow
- **v0.1.0** — initial structure
