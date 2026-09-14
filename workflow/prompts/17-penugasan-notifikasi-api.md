# PROMPT — Build API `jadwal_peserta` (assignment) + `notifikasi` (Fase 4)

> **Repo-accurate (revised 2026-09-13).** RBAC, `jadwal`/`jadwal_sesi`, `anggota`, `users`, audit built; `jadwal.assign` seeded; `Claims.AnggotaID` available; migration `0016`.

Working directory: `D:/xampp/htdocs/slam-team-apl/slam-team-api`. Provide a reusable notification service later modules will call (izin, absensi, kta).

## 0. Read first
1. `slamteam_db.dbml` — **`jadwal_peserta`** (line 813), **`notifikasi`** (line 329). DBML WINS.
2. `workflow/modules/17-penugasan-notifikasi.md`; `_shared/api-endpoints.md` §7 + §4 (verbatim).
3. Reference: `internal/modules/core/jadwal/` (same `Initialize(db, jwtMgr, perm, auditor)` shape) + `internal/shared/{model,apperr,response,audit}`. `Claims` carries `AnggotaID` (from `04-auth`).

## 1. Scope + endpoints
Assignment (`jadwal.assign` — seeded, Admin/Mod):
- POST `/jadwal/:id/peserta` `{anggota_id, sesi_id, peran_peserta, wajib_absen, keterangan}`
- POST `/jadwal/:id/peserta/bulk` `{anggota_ids[], peran_peserta, wajib_absen}` → `{ditugaskan, dilewati}`
- GET `/jadwal/:id/peserta` (`jadwal.read`)
- DELETE `/jadwal/:id/peserta/:anggotaId` (soft delete, `jadwal.assign`)
- PATCH `/peserta/:id/respon` `{status_tugas: diterima|ditolak, keterangan}` — **bearer only** (assignee: `claims.AnggotaID` == row.anggota_id else 403/404)

Notifikasi (bearer, own-scope by `claims.UserID`): GET `/notifikasi`, GET `/notifikasi/jumlah-belum-dibaca` → `{jumlah}`, PATCH `/notifikasi/:id/baca`, PATCH `/notifikasi/baca-semua`.

## 2. Files
- `internal/modules/core/penugasan/`: domain `JadwalPeserta` (soft-delete triplet); dto (`AssignReq`, `BulkAssignReq`, `ResponReq{status_tugas oneof=diterima ditolak}`); repository (`Create`, `BulkCreate` skip dupes by `(anggota_id, jadwal_id)`, `List`, `SoftDelete`, `Respon`, `AnggotaBerakun(anggotaID)`); service:
  - **Assign:** validate anggota; **force `wajib_absen=false` when `AnggotaBerakun` is false**; persist; **fan-out notifikasi** to the anggota's user (only if account-holding) with `tipe=jadwal_ditugaskan` + `route`; bump `jadwal_sesi.jml_ditugaskan`; `auditor.Write`.
  - **Bulk:** iterate `anggota_ids`; return `{ditugaskan, dilewati}`.
  - **Respon:** assignee only (`claims.AnggotaID`); set `status_tugas` + `direspon_pada`.
- `internal/modules/core/notifikasi/`: domain `Notifikasi`; repository (`CreateMany`, `ListByUser`, `CountUnread`, `MarkRead`, `MarkAllRead`); a **reusable exported `notifikasi.Service`** (later modules fan-out through it); handlers bearer-only, always scoped to `claims.UserID`.
- `main.*.go` + router: assignment routes under the jadwal path group (`RequirePermission("jadwal.assign")`; `respon` bearer-only); notifikasi routes bearer-only. Register both.

## 3. Migration `0016_jadwal_peserta_notifikasi` (+ `.down.sql`)
`jadwal_peserta` + `notifikasi` EXACTLY per `.dbml`: enum `status_tugas`/`prioritas_notifikasi` (from `0001_enums`); `notifikasi.uuid` default `gen_random_uuid()`; index `(user_id,is_dibaca,created_at)` + `(reff_type,reff_id)` + `kedaluwarsa_pada`; `notifikasi.user_id → users(id)` **CASCADE**, `notifikasi.created_by → users(id)`; `jadwal_peserta` FKs `jadwal_id → jadwal(id)` restrict, `sesi_id → jadwal_sesi(id)`, `anggota_id → anggota(id)` restrict, `ditugaskan_oleh → users(id)`; `jadwal_peserta` soft-delete triplet. All FK targets exist → immediate. `.down` drops both.

## 4. Conventions
Envelope via `response`; sentinels via `apperr` + `response.FromError`; no AutoMigrate; soft-delete on jadwal_peserta; `auditor.Write` on assign/remove; notifikasi never leaks across users.

## 5. Verification checklist
- [ ] `go build ./...` + `go vet ./...` pass; `slamctl migrate up`/`down` clean (`0016`).
- [ ] Assign an account-less anggota → `wajib_absen` forced `false`; no notifikasi for them.
- [ ] Assigning an account-holder creates a `notifikasi` row (fan-out) + bumps `jml_ditugaskan`.
- [ ] Bulk returns `{ditugaskan, dilewati}`; duplicate `(anggota_id, jadwal_id)` skipped.
- [ ] `PATCH /peserta/:id/respon` only by the assignee; `jadwal.assign` blocked for User (403).
- [ ] `/notifikasi*` scoped to `claims.UserID`; no cross-user leak; unread + mark-read/all work.
- [ ] `auditor.Write` on assign/remove; notifikasi service exported/reusable.
