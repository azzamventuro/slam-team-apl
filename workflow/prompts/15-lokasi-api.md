# PROMPT — Build API module `lokasi` (mst_lokasi, geofence) (Fase 3)

> **Repo-accurate (revised 2026-09-13).** RBAC, file-management, audit, and the shared helpers are all built; `lokasi.*` is already seeded; migration number is `0014`. Reuse — this is a standard CRUD module.

Working directory: `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

## 0. Read first
1. `slamteam_db.dbml` — **`mst_lokasi`** (line 717). DBML WINS.
2. `workflow/modules/15-lokasi.md`; `_shared/api-endpoints.md` §5.
3. **Use `internal/modules/core/instansi/` as the reference module** — copy its layout, its `Initialize(db, jwtMgr, perm *middleware.PermGuard, auditor *audit.Writer)` signature, its file-reference pattern, its list/pagination, and how it registers in `internal/router/router.go`. Also skim `internal/shared/{model,pagination,apperr,response,audit}`.

## 1. Scope + endpoints
CRUD `/lokasi`, guarded with the **real** `RequirePermission("lokasi.<aksi>")` (RBAC is live; `lokasi.*` is already seeded — do NOT re-seed, just verify). Band: SA/Admin CRUD, Moderator R, User R, Guest none.

| Method | Path | Permission |
|--------|------|-----------|
| GET | `/lokasi` | `lokasi.read` |
| GET | `/lokasi/:id` | `lokasi.read` |
| POST | `/lokasi` | `lokasi.create` |
| PUT | `/lokasi/:id` | `lokasi.update` |
| DELETE | `/lokasi/:id` | `lokasi.delete` (soft delete) |

## 2. Files under `internal/modules/core/lokasi/`
- **domain** `Lokasi` (`mst_lokasi`, `TableName`; embed `model.Audit`): `kode`, `nama`, `jenis_lokasi`, `alamat`, `latitude`/`longitude` (decimal(10,7) → float64), `radius_meter int`, `timezone`, `foto_file_id` (follow the existing `instansi`/`anggota` file-ref convention — match it, don't invent a third pattern), `keterangan`, `is_aktif`.
- **dto**: `CreateLokasiReq`/`UpdateLokasiReq` — `kode required,max=50`; `nama required,max=150`; `latitude required,latitude`; `longitude required,longitude`; `radius_meter required,min=1,max=100000`; `timezone required` (validate via `time.LoadLocation` in service); foto per convention. `LokasiResp` (+ resolved foto uuid, `jumlah_jadwal` — see seam below), `ListLokasiQuery` (embed the shared list query + `is_aktif *bool`).
- **repository**: `Create/Update/SoftDelete/FindByID/List` — all filter `is_deleted=false`. `List`: `q` ILIKE `kode|nama|alamat`; whitelist sort `kode,nama,created_at`; filter `is_aktif`; join foto. `ExistsKode(kode, exceptID)`.
- **service**: unique `kode` → `apperr.ErrConflict` (409, ignore self on update); invalid `timezone` (`time.LoadLocation` fails) → `apperr.ErrValidation`; set audit fields from `middleware.Claims`; `auditor.Write` on create/update/delete.
- **handler / main.lokasi.go / router**: `Initialize(db, jwtMgr, permGuard, auditor)`; `RequirePermission("lokasi.<aksi>")`; register `lokasi.Initialize(...).SetupRoutes(apiV1)` in `router.go`.

### SEAM — delete guard deferred
`mst_lokasi` is referenced by `jadwal.lokasi_id` (restrict), but the **`jadwal` table does not exist yet** (built in `16-jadwal`). So: do NOT query `jadwal`; implement `DELETE` as a plain soft-delete; set `jumlah_jadwal = 0` for now; leave a marked seam (`// TODO(16-jadwal): block delete when used by an active jadwal + populate jumlah_jadwal`). The 409-in-use guard is added when `16-jadwal` lands.

## 3. Migration + seeder
- `migrations/0014_mst_lokasi.up.sql` (+ `.down.sql`) — columns EXACTLY per `.dbml`, including **CHECK** (`latitude BETWEEN -90 AND 90`, `longitude BETWEEN -180 AND 180`, `radius_meter > 0`), partial-unique `kode WHERE is_deleted=false`, FK `foto_file_id → mst_file(id)`. **`users` exists** → add audit FKs `created_by/modified_by/deleted_by → users(id) ON DELETE SET NULL` immediately (no deferral). `.down` drops the table.
- Permissions: `lokasi.*` already seeded by `seed_rbac` — verify only, don't duplicate.

## 4. Conventions
Envelope via `response`; sentinels via `apperr` + `response.FromError`; no AutoMigrate; explicit soft-delete; whitelist sort; `per_page` cap 100; `auditor.Write` on create/update/delete.

## 5. Verification checklist
- [ ] `go build ./...` + `go vet ./...` pass; `slamctl migrate up`/`down` clean; matches `.dbml` incl. CHECK + partial-unique `kode`.
- [ ] CRUD round-trips; duplicate `kode` → 409; out-of-range coords / `radius<=0` / bad timezone → 422.
- [ ] `RequirePermission("lokasi.*")` enforced: Mod/User 403 on writes, 200 on read; SA bypasses.
- [ ] Foto via the same file-ref pattern as `instansi`; audit FKs to `users` immediate.
- [ ] Delete is soft; jadwal-usage guard left as a marked seam (no query to a non-existent `jadwal`).
- [ ] `auditor.Write` rows written; no raw errors leak.
