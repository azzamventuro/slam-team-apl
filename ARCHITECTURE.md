# Slam Team — Architecture & Standards

Standardized reference for both projects in this folder. Read this first; each
project also has its own `CLAUDE.md` with details.


| Project         | Stack                                                            | Doc                                                |
| --------------- | ---------------------------------------------------------------- | -------------------------------------------------- |
| `slam-team-api` | Go 1.24 · Gin · GORM/PostgreSQL · JWT · Zap                  | [slam-team-api/CLAUDE.md](slam-team-api/CLAUDE.md) |
| `slam-team-app` | Angular 21 · standalone/signals · Bootstrap 5 · ngx-translate | [slam-team-app/CLAUDE.md](slam-team-app/CLAUDE.md) |

## How the two connect

```
slam-team-app  ──HTTP──▶  slam-team-api
 (Angular :4200)          (Gin :8080, routes under /api/v1)
```

- `app` reads `environment.apiURL` → `http://localhost:8080/api/v1`.
- **Auth:** `POST /api/v1/auth/login` returns a JWT. The app stores it in
  `localStorage` and sends `Authorization: Bearer <jwt>` on every request
  (via `authInterceptor`). The API validates it in `middleware.JWTAuth`.

## Backend standard (Go) — feature module

Every feature is one folder under `internal/modules/core/<name>/`:

```
domain/ · dto/ · repository/ · service/ · handler/ · main.<name>.go
```

Request flow: **router → middleware → handler → service → repository → GORM**.
Cross-cutting code lives in `pkg/` (jwt, logger, validator, utils) and
`internal/shared/` (response envelope, redis). Modules get their dependencies
(`*gorm.DB`, JWT manager) at `Initialize()` and expose `SetupRoutes()`.

## Frontend standard (Angular) — feature page

Every feature is a standalone component under `pages/<name>/`, lazy-loaded via
`loadComponent`. Singletons (services, guards, interceptors, models) live in
`core/`; layout shells in `layouts/`; reusable UI in `shared/`. State is held in
**signals**; templates use `@if` / `@for`; text goes through ngx-translate keys.

## Adding a feature end-to-end (example: "projects")

**Backend**

1. Create `internal/modules/core/project/` with `domain / dto / repository / service / handler / main.project.go`.
2. Register it in `internal/router/router.go`:
   `project.Initialize(db, jwtMgr).SetupRoutes(apiV1)`.

**Frontend**

1. `core/services/project.service.ts` → `HttpClient` calls to
   `${environment.apiURL}/projects`.
2. `pages/project/project.ts` (+ `.html` / `.scss`); add a child route under the
   layout in `app.routes.ts`.
3. Add i18n keys to `public/i18n/IND.json` and `ENG.json`.

## Toolchain

Go 1.24+ (this repo built with 1.26) · Node 24 · Angular CLI 21 · PostgreSQL 16 ·
Redis (optional).

- **Go deps:** `cd slam-team-api && go mod tidy`
- **Angular deps:** `cd slam-team-app && npm install --legacy-peer-deps`
  (the `--legacy-peer-deps` flag works around an npm 11.3 peer-resolution crash).

## Run locally

```bash
# API (needs PostgreSQL — or use docker-compose)
cd slam-team-api && cp .env.example .env && make run

# APP
cd slam-team-app && npm run serve:local
```
