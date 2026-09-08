# PROMPT — Build API `jadwal_peserta` (assignment) + `notifikasi` (Fase 4)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build participant **assignment** (`jadwal_peserta`) and in-app **notifikasi** (fan-out). Provide a reusable notification service other modules call.

## 0. Read first
1. `slamteam_db.dbml` — **`jadwal_peserta`** (line 813), **`notifikasi`** (line 329); relations (`jadwal_peserta.*` restrict; `notifikasi.user_id → users.id` **cascade**). DBML WINS.
2. `workflow/modules/17-penugasan-notifikasi.md`.
3. `_shared/api-endpoints.md` §7 (peserta) + §4 (notifikasi) verbatim.

## 1. Scope + endpoints
Assignment (`jadwal.assign` = SA/Admin/Mod):
- POST `/jadwal/:id/peserta` `{anggota_id, sesi_id, peran_peserta, wajib_absen, keterangan}`
- POST `/jadwal/:id/peserta/bulk` `{anggota_ids[], peran_peserta, wajib_absen}` → `{ditugaskan, dilewati}`
- GET `/jadwal/:id/peserta` (`jadwal.read`)
- DELETE `/jadwal/:id/peserta/:anggotaId` (soft delete)
- PATCH `/peserta/:id/respon` `{status_tugas: diterima|ditolak, keterangan}` — **bearer** (assignee only)

Notifikasi (bearer, own-scope): GET `/notifikasi`, GET `/notifikasi/jumlah-belum-dibaca` → `{jumlah}`, PATCH `/notifikasi/:id/baca`, PATCH `/notifikasi/baca-semua`.

## 2. Files
- `internal/modules/core/penugasan/`: domain `JadwalPeserta`; dto (`AssignReq`, `BulkAssignReq`, `ResponReq{status_tugas oneof=diterima ditolak}`); repository (`Create`, `BulkCreate` skip `(anggota_id,jadwal_id)` dupes, `List`, `SoftDelete`, `Respon`, `AnggotaBerakun(anggotaID)`); service — **set `wajib_absen=false` when the anggota has no user account**; on assign **fan-out a notifikasi** to the anggota's user (if any) with `tipe=jadwal_ditugaskan` + `route`; bump `jadwal_sesi.jml_ditugaskan`; bulk supports an anggota_ids list (document by-instansi expansion); respon restricted to the assignee; `audit.Write` on assign/remove.
- `internal/modules/core/notifikasi/` (+ a callable `notifikasi.Service` for other modules): domain `Notifikasi`; repository (`CreateMany`, `ListByUser`, `CountUnread`, `MarkRead`, `MarkAllRead`); service own-scoped by `claims.UserID`; handlers bearer-only.

## 3. Migration
`jadwal_peserta` + `notifikasi` EXACTLY per `.dbml` (enum `status_tugas`/`prioritas_notifikasi`; `notifikasi.uuid` default gen; index `(user_id,is_dibaca,created_at)`; `notifikasi.user_id` cascade; `jadwal_peserta` soft-delete). After `jadwal_sesi`, `anggota`, `users`. Confirm `jadwal.assign` seeded.

## 4. Verification checklist
- [ ] build/vet pass; migration up/down clean; matches `.dbml`.
- [ ] Assign an account-less anggota → `wajib_absen=false` forced; no notifikasi for them.
- [ ] Bulk returns `{ditugaskan, dilewati}`; duplicates skipped; `jml_ditugaskan` bumped.
- [ ] Assign triggers a fan-out notifikasi (one row per account-holding recipient).
- [ ] `PATCH /peserta/:id/respon` only by the assignee; notifikasi list/read scoped to `claims.UserID`.
- [ ] `jadwal.assign` blocked for User (403); `audit.Write` on assign/remove.
