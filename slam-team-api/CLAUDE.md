# CLAUDE.md — slam-team-api

Guidance for AI agents (and humans) working in this repository.

## Overview

Go REST API built with **Go 1.24**, **Gin**, **GORM (PostgreSQL)**, **JWT (HS256)**,
**Zap**, and **go-playground/validator**.

## Development commands

```bash
cp .env.example .env     # configure DB_* and JWT_SECRET
make run                 # go run ./cmd/api      (or: make dev — hot reload via Air)
make build               # → bin/slam-team-api
make test                # go test ./...
make vet                 # go vet ./...
make tidy                # go mod tidy
```

Air hot reload needs: `go install github.com/air-verse/air@latest`.

## Bootstrap sequence — `cmd/api/main.go`

1. `config.Load()` → env (+ optional `.env`) into typed structs
2. `logger.Initialize(env)` → Zap (JSON in prod, colour console in dev)
3. `validator.Register()` → custom rules on Gin's validator
4. `jwt.New(secret, issuer, ttl)` → HS256 issuer/verifier
5. `redis.Initialize(cfg)` → **optional**, returns nil if unavailable
6. `database.New(cfg)` → GORM PostgreSQL pool (**fatal** on failure)
7. `router.Setup(...)` → mounts `/api/v1`, registers modules
8. HTTP server with graceful shutdown (SIGINT/SIGTERM)

## Layout

```
cmd/api/main.go                entry point + bootstrap
internal/
  config/                      typed env config
  database/                    GORM PostgreSQL open + pool + ping
  middleware/                  Global (CORS/recovery), JWTAuth, context helpers
  router/                      engine + /api/v1 + module registration
  shared/response/             JSON envelope helpers
  shared/redis/                optional Redis client
  modules/core/<feature>/      feature modules (see pattern below)
pkg/
  jwt/                         HS256 issue/verify
  logger/                      Zap wrapper
  validator/                   custom validation rules
  utils/                       bcrypt password helpers
```

## Module pattern — `internal/modules/core/<name>/`

```
<name>/
├── domain/          GORM entities, value objects
├── dto/             request/response structs with binding tags
├── repository/      GORM data access
├── service/         business logic, composes repositories
├── handler/         Gin handlers (bind → call service → response)
└── main.<name>.go   Initialize(deps) *Module + (m *Module) SetupRoutes(rg)
```

Request flow: **router → middleware → handler → service → repository → GORM**.

`auth` is the reference implementation (login + me).

## Conventions

- **Responses** — always use `internal/shared/response`: `OK`, `Created`,
  `BadRequest`, `Unprocess`, `Unauthorized`, `Internal`. Envelope shape:
  `{ "success": bool, "message": string, "data"?: any, "errors"?: any }`.
  `Internal` logs the real error and returns a generic 500 (no leak).
- **Logging** — `pkg/logger` (Zap): `logger.Info("msg", zap.String("k", v))`.
- **Validation** — Gin `binding` tags on DTOs; register custom rules in
  `pkg/validator`.
- **Auth** — protect routes with `middleware.JWTAuth(jwtMgr)`; read the user
  with `middleware.Claims(c)`.
- **Passwords** — `pkg/utils.HashPassword` / `CheckPassword` (bcrypt).
- **Dependency injection** — modules receive `*gorm.DB` (and other deps) at
  `Initialize()`; there is no per-request DB resolution (single DB).

## Adding an endpoint

1. Pick or create a module under `internal/modules/core/`.
2. Add `domain` → `repository` → `service` → `dto` → `handler`.
3. Wire them in `main.<module>.go` `Initialize()` and `SetupRoutes()`.
4. Register the module in `internal/router/router.go`.
5. Attach `middleware.JWTAuth(jwtMgr)` to protected routes.

## Gotchas

- **No `AutoMigrate`.** Create the `users` table yourself; `password` is a
  bcrypt hash (see README for generating one).
- **PostgreSQL is required to boot** — `main.go` fatals if the DB is
  unreachable. Use `docker-compose up` or set `DB_*` (default port 5432).
- **Redis is optional** — empty `REDIS_ADDR` disables it (non-fatal).
- **`JWT_SECRET` must be stable** across restarts, or issued tokens stop
  verifying.
- **Endpoints live under `/api/v1`** (e.g. `POST /api/v1/auth/login`).
