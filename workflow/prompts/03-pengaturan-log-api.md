# PROMPT — Build API modules `pengaturan` + `log_aktivitas` (Fase 0)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

You are building the **typed settings** (`mst_pengaturan`) and the **audit log** (`log_aktivitas`) into slam-team-api. Fase 0.

## 0. Read first
1. `D:/xampp/htdocs/slam-team-apl/slamteam_db.dbml` — **`mst_pengaturan`** (line 270, includes the ~20-row seed list in its Note) and **`log_aktivitas`** (line 370). DBML WINS.
2. `D:/xampp/htdocs/slam-team-apl/workflow/modules/03-pengaturan-log.md`.
3. `_shared/conventions-api.md` §12 (logging & audit), `_shared/api-endpoints.md` §3.

## 1. Scope + endpoints
`mst_pengaturan` (kunci unique `grup.nama`, grup, label, nilai text, tipe_nilai enum, nilai_bawaan, opsi jsonb, satuan, urutan, is_publik, is_terkunci, modified_by). `log_aktivitas` (aktor_user_id nullable, modul, aksi, reff_type/reff_id, ringkasan, nilai_lama/nilai_baru jsonb, ip_address inet, user_agent; **no soft-delete**).

| Method | Path | Permission |
|--------|------|-----------|
| GET | `/pengaturan` | `pengaturan.read` (grouped by `grup`) |
| GET | `/pengaturan/:grup` | `pengaturan.read` |
| PUT | `/pengaturan` | `pengaturan.update` — bulk upsert `{items:[{kunci,nilai}]}` |
| GET | `/public/pengaturan` | — (public, only `is_publik=true`) |

## 2. Files under `internal/modules/core/pengaturan/` + `internal/shared/audit/`
- **pengaturan**: domain `Pengaturan` (`mst_pengaturan`); repository `ListByGrup`, `List`, `FindByKunci`, `BulkUpsert`; service validates each `nilai` against `tipe_nilai` (`string/integer/boolean/json/date`), **rejects `is_terkunci` rows unless caller is_super**, writes `log_aktivitas` on change; handler; `RequirePermission("pengaturan.read|update")`. Public handler returns `is_publik=true` rows only (never thresholds/secrets).
- **shared/audit** — the reusable writer every module calls:
```go
func Write(ctx, db, entry audit.Entry) // aktor_user_id, modul, aksi, reff_type, reff_id, ringkasan, nilai_lama, nilai_baru, ip, user_agent
```
  Implement this now so modules 04+ can log create/update/delete.

## 3. Migration + seeder
- `mst_pengaturan` + `log_aktivitas` EXACTLY per `.dbml` (enum `tipe_nilai_pengaturan`; jsonb; inet; indexes `kunci`, `(grup,urutan)`, `(modul,created_at)`, `(reff_type,reff_id)`, `(aktor_user_id,created_at)`). Order after `0001_enums`.
- **Idempotent seeder** (`slamctl seed pengaturan`) inserting the ~20 seed rows from the DBML Note (`umum.nama_klub`, `nra.*`, `kta.*`, `absensi.*`, `file.*`, …) with `ON CONFLICT (kunci) DO NOTHING`.

## 4. Conventions
Envelope; no AutoMigrate; audit table has NO soft-delete; `Internal` is the only 500 log point; never leak raw errors.

## 5. Verification checklist
- [ ] `go build`/`go vet` pass; migration up/down clean; matches `.dbml`.
- [ ] Seeder run twice → the seed rows exist exactly once (idempotent).
- [ ] `GET /pengaturan` groups by `grup`; `PUT /pengaturan` bulk-upserts and validates by `tipe_nilai`; `is_terkunci` rows rejected for non-super.
- [ ] `GET /public/pengaturan` returns ONLY `is_publik=true` (no thresholds/secrets).
- [ ] `audit.Write` inserts a `log_aktivitas` row (nilai_lama/nilai_baru jsonb, ip, user_agent) and is callable by other modules.
