# PROMPT — Build API module `izin` (absensi_izin) (Fase 5)

> **Repo-accurate (revised 2026-09-13).** RBAC (with cakupan), `jadwal_sesi`, `anggota`, file-management, the notifikasi service (module 17), audit built; `izin.*` seeded; migration `0017`. `absensi` NOT built yet — do not touch it.

Working directory: `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

## 0. Read first
1. `slamteam_db.dbml` — **`absensi_izin`** (line 929). DBML WINS.
2. `workflow/modules/19-izin.md`; `_shared/api-endpoints.md` §9; `_shared/conventions-api.md` §8 (cakupan).
3. Reference: `internal/modules/core/unit/` for the **cakupan-reading pattern** and `internal/modules/core/notifikasi/` for the reusable notifikasi service. `Claims` carries `AnggotaID`.

## 1. Scope + endpoints (real RBAC — already seeded)
| Method | Path | Permission |
|--------|------|-----------|
| POST | `/izin` | `izin.create` (all logged-in) |
| GET | `/izin` | `izin.read` (cakupan: `semua` Admin/Mod, `milik_sendiri` User) |
| PATCH | `/izin/:id/approve` | `izin.approve` (Admin/Mod) |
| PATCH | `/izin/:id/tolak` | `izin.approve` |

## 2. Files under `internal/modules/core/izin/`
- **domain** `Izin` (`absensi_izin`, embed `model.Audit`; enums `jenis_izin`/`status_izin`).
- **dto**: `CreateIzinReq{ sesi_id required; jenis oneof=izin sakit dinas pulang_cepat; alasan required; waktu_pulang_diminta omitempty; lampiran_file_id omitempty,gt=0 }`; `TolakReq{ catatan_peninjau required }`; `IzinResp`; `ListIzinQuery` (filter `status`, `sesi_id`).
- **repository**: `Create`, `FindByID`, `List(scope, filters)` (`is_deleted=false`), `Approve`, `Tolak`.
- **service**:
  - **Create:** sesi exists; `waktu_pulang_diminta` **required only when `jenis=pulang_cepat`** (reject if set for other jenis); validate `lampiran_file_id` if given; for `milik_sendiri` scope force `anggota_id = claims.AnggotaID`; `auditor.Write`.
  - **List:** read the resolved cakupan for `izin.read` (mirror `unit`): `milik_sendiri` → `WHERE anggota_id = claims.AnggotaID`; `semua` → no owner filter.
  - **Approve:** `status=disetujui`, `diproses_oleh/pada`; fan-out notifikasi `izin_disetujui`. **SEAM:** the absensi linkage (create/update an `absensi` row with `izin_id` + `status_kehadiran`) is DEFERRED to `18-absensi` — leave a `// TODO(18-absensi)` marker; do NOT reference an `absensi` table (it doesn't exist).
  - **Tolak:** `status=ditolak` + `catatan_peninjau` (required); notifikasi `izin_ditolak`.
- **handler/main.izin.go/router**: `Initialize(db, jwtMgr, permGuard, auditor, <notifikasi service>)`; `RequirePermission("izin.<aksi>")` (`approve`/`tolak` → `izin.approve`); register in `router.go`.

## 3. Migration `0017_absensi_izin.up.sql` (+ `.down.sql`)
`absensi_izin` EXACTLY per `.dbml`: enum `jenis_izin`/`status_izin` (from `0001_enums`); FKs (targets exist → immediate) `sesi_id → jadwal_sesi(id)` **RESTRICT**, `jadwal_id → jadwal(id)`, `anggota_id → anggota(id)` **RESTRICT**, `lampiran_file_id → mst_file(id)`, `diproses_oleh → users(id)`; audit `*_by → users(id)`; soft-delete triplet; indexes `(sesi_id, anggota_id)` + `(status, created_at)`. `.down` drops the table. (`absensi.izin_id → absensi_izin` is added later by the absensi migration.)

## 4. Conventions
Envelope via `response`; sentinels via `apperr` + `response.FromError`; no AutoMigrate; soft-delete; cakupan enforced in service (read AND create); `auditor.Write` on create/approve/tolak; notifikasi via the module-17 service.

## 5. Verification checklist
- [ ] `go build ./...` + `go vet ./...` pass; `slamctl migrate up`/`down` clean (`0017`).
- [ ] `waktu_pulang_diminta` valid only for `pulang_cepat`; missing `alasan` → 422.
- [ ] `izin.create` for all logged-in; User `izin.read` sees only their own; Admin/Mod see all.
- [ ] Approve → `disetujui` + notifikasi `izin_disetujui`; Tolak requires `catatan_peninjau` + notifikasi `izin_ditolak`; `izin.approve` blocked for User (403).
- [ ] Absensi linkage left as a marked seam (no reference to a non-existent `absensi`).
- [ ] Lampiran validated via `mst_file`; `auditor.Write` rows written; no raw errors leak.
