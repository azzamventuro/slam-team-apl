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
| POST   | `/api/v1/auth/login` | —     | Email/password → JWT |
| GET    | `/api/v1/auth/me`    | Bearer | Current user's claims |

## Seed a user

Create the table (no `AutoMigrate` runs), then insert a user whose `password`
is a bcrypt hash:

```sql
CREATE TABLE users (
  id         BIGSERIAL PRIMARY KEY,
  name       VARCHAR(255) NOT NULL,
  email      VARCHAR(255) UNIQUE NOT NULL,
  password   VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ
);
```

Generate a bcrypt hash for the `password` column:

```go
import "slam-team-api/pkg/utils"
h, _ := utils.HashPassword("secret123") // insert h into users.password
```

## Make targets

`make dev | run | build | tidy | vet | test | clean`
