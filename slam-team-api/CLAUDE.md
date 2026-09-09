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
make build-cli           # → bin/slamctl
make test                # go test ./...
make vet                 # go vet ./...
make tidy                # go mod tidy

make migrate-up          # apply pending SQL migrations
make migrate-down N=1    # roll back n migrations (N=all for everything)
make migrate-version     # current schema version
make seed NAME=rbac      # idempotent reference-data seeders
```

Operational tasks live in the `slamctl` binary (`cmd/slamctl`), never behind an
HTTP endpoint:

```bash
go run ./cmd/slamctl migrate up|down [n|all]|version
go run ./cmd/slamctl seed [name]
go run ./cmd/slamctl create-superadmin --nama … --username … --email … --tanggal-lahir YYYY-MM-DD
# password comes from --password or $SLAMCTL_SUPERADMIN_PASSWORD; refuses to run
# once any super admin exists
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
cmd/slamctl/                   operational CLI (migrate / seed / create-superadmin)
migrations/                    numbered .up.sql/.down.sql, embedded via embed.FS
internal/
  config/                      typed env config
  database/                    GORM PostgreSQL open + pool + ping
  middleware/                  Global (CORS/recovery), JWTAuth, context helpers
  migrator/                    golang-migrate runner over the embedded SQL
  router/                      engine + /api/v1 + module registration
  shared/response/             JSON envelope helpers
  shared/apperr/               service sentinel errors
  shared/model/                Audit embed + soft-delete scope
  shared/pagination/           ListQuery + Paginated[T]
  shared/timeutil/             IANA zone helpers (UTC in services)
  shared/redis/                optional Redis client
  shared/audit/                log_aktivitas writer (every state change)
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

`auth` shows the layout (login + me), but its entity is a placeholder — see
Gotchas.

## Conventions

- **Responses** — always use `internal/shared/response`: `OK`, `Created`,
  `BadRequest`, `Unprocess`, `Unauthorized`, `Forbidden`, `NotFound`,
  `Conflict`, `Internal`. Envelope shape:
  `{ "success": bool, "message": string, "data"?: any, "errors"?: any }`.
  `Internal` logs the real error and returns a generic 500 (no leak).
- **Errors** — services return `internal/shared/apperr` sentinels; handlers call
  `response.FromError(c, err)`. Status codes never appear in service code.
- **Soft delete** — explicit `is_deleted/deleted_at/deleted_by` via
  `internal/shared/model.Audit`, not `gorm.DeletedAt`. Reads apply the
  `model.NotDeleted` scope.
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

- **No `AutoMigrate`.** The schema is the numbered SQL files in `migrations/`,
  applied with `slamctl migrate up`. A committed migration is never edited — add
  a new, higher-numbered `.up.sql`/`.down.sql` pair.
- **The scaffold `auth` module is a placeholder** (`users(id,name,email,password)`)
  and does not match `slamteam_db.dbml`. Don't build on its `domain.User`.
- **PostgreSQL is required to boot** — `main.go` fatals if the DB is
  unreachable. Use `docker-compose up` or set `DB_*` (default port 5432).
- **Redis is optional** — empty `REDIS_ADDR` disables it (non-fatal).
- **`STORAGE_ROOT` must be writable** — the file layer probes it at boot, so
  `router.Setup` (and therefore `cmd/api`) fails fast rather than accepting an
  upload whose bytes go nowhere. Disk layout:
  `{STORAGE_ROOT}/{kategori}/{tahun}/{bulan}/{uuid}/{varian}.{ext}`.
- **Never serve a private file from a static folder** — `is_publik = false`
  means `GET /api/v1/files/{uuid}/{varian}` only; the handler checks the caller
  before streaming. Modules store `*_file_id` (→ `mst_file.id`), never a path.
- **`JWT_SECRET` must be stable** across restarts, or issued tokens stop
  verifying.
- **Endpoints live under `/api/v1`** (e.g. `POST /api/v1/auth/login`).
