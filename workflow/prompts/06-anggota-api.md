# PROMPT — Build API module `anggota` (Master Data Anggota) (Fase 2)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build the **`anggota`** master-data module (the core member record) into slam-team-api. Build INTO the existing layout.

## 0. Read first
1. `D:/xampp/htdocs/slam-team-apl/slamteam_db.dbml` — **`anggota`** (line 487) + relations (`anggota.instansi_id → mst_instansi.id` **restrict** not null, `wilayah_id → mst_wilayah.id`, three `*_file_id → mst_file.id`). DBML WINS.
2. `D:/xampp/htdocs/slam-team-apl/workflow/modules/06-anggota.md`.
3. `_shared/conventions-api.md` (module layout, envelope, soft-delete, file refs, RBAC) + `_shared/api-endpoints.md` §13. Use `instansi` as the reference CRUD module (`internal/modules/core/instansi`).

## 1. Scope + endpoints
Standard CRUD for `anggota`, RBAC-guarded, soft-delete, **three file references** (`foto_profil_file_id`, `foto_formal_file_id` [red-bg KTA photo], `file_identitas_file_id` [PRIVATE]).

| Method | Path | Permission |
|--------|------|-----------|
| GET | `/anggota` | `anggota.read` |
| GET | `/anggota/:id` | `anggota.read` |
| POST | `/anggota` | `anggota.create` |
| PUT | `/anggota/:id` | `anggota.update` |
| DELETE | `/anggota/:id` | `anggota.delete` |

## 2. Files under `internal/modules/core/anggota/`
- **domain** `Anggota` (`mst`? no — table is `anggota`): map columns EXACTLY: `instansi_id int64` (not null), `wilayah_id *int64`, `no_induk *string` (unique, nullable — a copy of active NRA, empty until first KTA), `nama_lengkap`, `nama_panggilan`, `foto_profil_file_id/foto_formal_file_id/file_identitas_file_id *int64`, `jenis_anggota` (enum `siswa/dewasa/siswa_ke_dewasa`), `jenis_kelamin int`, `jenis_identitas int`, `no_identitas`, `pekerjaan`, `alamat`, `kode_pos`, `tempat_lahir`, `tanggal_lahir date` (**not null** — supplies NRA month/year), `tanggal_bergabung date`, `status_anggota` (enum `aktif/non_aktif`, default aktif, **manual**), soft-delete + audit. `TableName() = "anggota"`.
- **dto**: `CreateAnggotaReq`/`UpdateAnggotaReq` with binding (`instansi_id required,gt=0`; `nama_lengkap required,max=150`; `jenis_anggota oneof=siswa dewasa siswa_ke_dewasa`; `tanggal_lahir required,datetime=2006-01-02`; `status_anggota oneof=aktif non_aktif`; file ids `omitempty,gt=0`). `AnggotaResp` (+ joined `*_uuid`, `instansi_nama`). `ListAnggotaQuery` (filter `instansi_id`, `jenis_anggota`, `status_anggota`, `q`).
- **repository**: CRUD (`is_deleted=false`); `List` joins `mst_instansi` (nama) + file uuids; `q` ILIKE `nama_lengkap|no_induk|nama_panggilan`; whitelist sort `nama_lengkap,no_induk,created_at`; filters. `ExistsNoInduk` (only when set). `InstansiExists`, `FileExists`.
- **service**: validate instansi exists; `no_induk` is NOT set here (it is a copy from KTA — leave null on create); parse `tanggal_lahir`; `status_anggota` manual (default aktif); `created_by` from Claims; `audit.Write` create/update/delete; cakupan `instansi_sendiri` for Moderator where the matrix/doc requires (filter by actor's instansi). Map to `AnggotaResp`.
- **handler/main/router**: `RequirePermission("anggota.<aksi>")`; register `anggota.Initialize(...).SetupRoutes(apiV1)`.

## 3. Migration + seeder
- `anggota` EXACTLY per `.dbml` (enum `jenis_anggota`/`status_anggota`; `no_induk` unique; FK restrict `instansi_id`; indexes `no_induk`,`instansi_id`,`jenis_anggota`,`status_anggota`). Order **after** `mst_instansi`, `mst_wilayah`, `mst_file`; **before** `users`/`kta`. Confirm `anggota.*` permissions seeded.

## 4. Conventions
Envelope; no AutoMigrate; soft-delete; `file_identitas` is PRIVATE (served via `/files/:uuid/:varian`); `tanggal_lahir` pure date; NRA is NOT computed here (that is module 20).

## 5. Verification checklist
- [ ] `go build`/`go vet` pass; migration up/down clean; matches `.dbml`.
- [ ] CRUD round-trips; `POST` sets `created_by`, leaves `no_induk` null; `tanggal_lahir` required.
- [ ] `jenis_anggota`/`status_anggota` validated against enum; invalid → 422.
- [ ] Three file ids stored; detail/list return `*_uuid`; `file_identitas` only reachable through the authenticated file endpoint.
- [ ] Moderator/User: create/update/delete blocked per matrix (User read-only → 403 on writes); cakupan applied where required.
- [ ] `audit.Write` on create/update/delete; no raw errors leaked.
