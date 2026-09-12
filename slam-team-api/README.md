# slam-team-api

Go REST API — Gin + GORM (PostgreSQL) + JWT + Zap. Layered module architecture. See [CLAUDE.md](CLAUDE.md) for the full structure and
conventions.

## Requirements

- Go 1.24+
- PostgreSQL (a `slamteam_db` database)
- Redis (optional)

## Quick start

```bash
cp .env.example .env        # edit DB_* and JWT_SECRET
go mod tidy
make run                    # or: go run ./cmd/api
```

Server listens on `:8080` (override with `APP_PORT`).

Hot reload: `make dev` (needs `go install github.com/air-verse/air@latest`).

## Endpoints


| Method | Path                 | Auth   | Description           |
| ------ | -------------------- | ------ | --------------------- |
| GET    | `/health`            | —     | Health + Redis status |
| POST   | `/api/v1/auth/login`   | —      | Username **or** email + password → access JWT + refresh token |
| POST   | `/api/v1/auth/refresh` | —      | Rotate the refresh token (old one revoked) |
| POST   | `/api/v1/auth/logout`  | Bearer | Revoke the session named by `refresh_token` |
| GET    | `/api/v1/auth/me`      | Bearer | Current user + role + linked anggota |
| GET    | `/api/v1/auth/me/permissions` | Bearer | Effective permission codes |

## Bootstrap the first user

There is no self-registration and no HTTP endpoint for creating a super admin.
Apply the schema, seed the reference data, then create the first account from
the CLI (it refuses to run once any super admin exists):

```bash
go run ./cmd/slamctl migrate up
go run ./cmd/slamctl seed
SLAMCTL_SUPERADMIN_PASSWORD='...' go run ./cmd/slamctl create-superadmin   --nama "Nama Lengkap" --username admin --email admin@example.com --tanggal-lahir 2000-01-01
```

Then `POST /api/v1/auth/login` with `{"identifier": "<username or email>", "password": "..."}`.
Five consecutive failures lock the account for 15 minutes (`users.terkunci_sampai`).

## Make targets

`make dev | run | build | tidy | vet | test | clean`
