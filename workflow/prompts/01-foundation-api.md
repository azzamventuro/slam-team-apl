# PROMPT — Foundation API skeleton, timestamptz & slamctl (Fase 0)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

You are completing the **Fase 0 foundation** of **slam-team-api** (Go 1.24 · Gin · GORM/PostgreSQL `slamteam_db` · JWT HS256 · Zap · go-playground/validator). **A scaffold ALREADY EXISTS — verify and FILL GAPS; do NOT rebuild existing files.**

## 0. Read first (authoritative, in this order)
1. `D:/xampp/htdocs/slam-team-apl/slamteam_db.dbml` — the schema you will migrate against (DBML WINS on any conflict). Note the cross-cutting conventions at the top (timestamptz UTC, soft-delete triplet, `mst_` prefix, file refs).
2. `D:/xampp/htdocs/slam-team-apl/workflow/modules/01-foundation.md` — this module's spec.
3. `D:/xampp/htdocs/slam-team-apl/workflow/_shared/conventions-api.md` — §1 layout, §2 envelope, §3 errors, §6 GORM, §7 migrations, §9 timezone, §11 slamctl. Follow exactly.
4. Skim the existing code: `cmd/api/main.go`, `internal/{config,database,middleware,router,shared/response}`, `pkg/{jwt,logger,validator,utils}`, `slam-team-api/CLAUDE.md`.

## 1. What already exists (confirm, don't recreate)
- `cmd/api/main.go` bootstrap; `internal/config`, `internal/database` (GORM pool), `internal/middleware` (`Global`, `JWTAuth`, `Claims`), `internal/router` (registers `auth` only), `internal/shared/response` (`OK/Created/BadRequest/Unprocess/Unauthorized/Internal`), `internal/shared/redis`, `pkg/jwt`, `pkg/logger`, `pkg/validator`, `pkg/utils` (bcrypt).

## 2. Gaps to fill
1. **Timezone**: ensure `import _ "time/tzdata"` is present (in `main.go` or a bootstrap file) so IANA zones resolve without OS tz files (conventions-api §9).
2. **Response sentinels + FromError**: add to `internal/shared/response` (do NOT fork it) — `Forbidden` (403) and `NotFound` (404) helpers, and a `FromError(c, err)` that switches on shared service sentinels (`ErrNotFound/ErrForbidden/ErrConflict/ErrValidation/ErrUnauthorized`) → status, else `Internal`. Define the sentinels in a shared package (e.g. `internal/shared/apperr`).
3. **Migration runner**: create `migrations/` (numbered `.up.sql`/`.down.sql`, monotonic, never edited after commit) and a runner invoked by `slamctl migrate up|down [n]` (e.g. golang-migrate). NO `AutoMigrate`. First migration `0001_enums.up.sql` creates every Postgres ENUM from the DBML (`aksi_permission`, `cakupan_permission`, `jenis_anggota`, `status_anggota`, `jenis_kta`, `status_kta`, `mode_absen`, `tipe_absensi`, `metode_absensi`, `status_kehadiran`, `jenis_izin`, `status_izin`, `status_sesi`, `status_tugas`, `pola_ulang`, `varian_file`, `status_proses_file`, `tipe_nilai_pengaturan`, `prioritas_notifikasi`).
4. **slamctl CLI** (`cmd/slamctl`, separate binary — NEVER HTTP): subcommands `migrate up|down [n]`, `seed [name]` (idempotent), `create-superadmin` (creates one `anggota` + `users` bound to the Super Admin role; **refuses if any is_super user exists**; interactive/env-driven, never an endpoint).
5. **Shared list helpers**: the `ListQuery` + `Paginated[T]` from conventions-api §5 (page/per_page cap 100, whitelisted sort), reusable by every module.
6. **Audit helper stub**: a `log_aktivitas` writer signature that later modules call (implemented fully in `03-pengaturan-log`).

## 3. Conventions to honor
- Envelope only via `internal/shared/response`. No AutoMigrate. Soft delete = explicit `is_deleted/deleted_at/deleted_by`. Times `timestamptz` UTC. `Makefile`/`slamctl` for migrate+seed. Modules under `internal/modules/core/<mod>/` with `domain/dto/repository/service/handler + main.<mod>.go`.

## 4. Verification checklist
- [ ] `go build ./...` + `go vet ./...` pass; `make run` boots (Postgres reachable).
- [ ] `_ "time/tzdata"` imported; `time.LoadLocation("Asia/Jakarta")` works.
- [ ] `response.Forbidden/NotFound/FromError` exist; sentinels defined once; no status codes scattered in services.
- [ ] `slamctl migrate up` applies `0001_enums` (+any present) and `migrate down` rolls back cleanly.
- [ ] `slamctl create-superadmin` creates the first super admin and **refuses** on a second run; it is NOT reachable over HTTP.
- [ ] `ListQuery`/`Paginated[T]` helpers available; `per_page` capped at 100.
- [ ] Existing files (`main.go`, `auth`, `response`) were extended, not rebuilt.
