# PROMPT — Build API module `kegiatan` (admin + public) (Fase 8)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build the **`kegiatan`** CRUD module + a public read.

## 0. Read first
1. `slamteam_db.dbml` — **`kegiatan`** (line 1003): `user_id` (author, not null, FK users), `kode`, `judul`, `gambar_highlight_file_id`, `konten`, `tanggal_mulai/tanggal_selesai date`, `status_kegiatan int`, soft-delete + audit. DBML WINS.
2. `workflow/modules/22-kegiatan.md`; `_shared/api-endpoints.md` §13 + §12.

## 1. Scope + endpoints
CRUD `/kegiatan` (`RequirePermission("kegiatan.<aksi>")`) + `GET /public/kegiatan` (public, `status_kegiatan=1` only). Band: SA/Admin/Mod CRUD, User R, Guest R.

## 2. Files under `internal/modules/core/kegiatan/`
- domain `Kegiatan` (`StatusKegiatan int`, `GambarHighlightFileID *int64`); dto (`CreateKegiatanReq`: `judul required,max=200`; `kode max=50`; `status_kegiatan oneof=0 1`; `gambar_highlight_file_id omitempty,gt=0`; dates `omitempty,datetime=2006-01-02`); repository (CRUD `is_deleted=false`, join `gambar_uuid`, `q` ILIKE `judul`, sort `judul,tanggal_mulai,created_at`, public query `status_kegiatan=1`); service (`user_id=claims.UserID` on create; `audit.Write`); handler (+ public handler); main+router.

## 3. Migration
`kegiatan` EXACTLY per `.dbml` (FK `user_id → users`, `gambar_highlight_file_id → mst_file`). After `users`, `mst_file`. Confirm `kegiatan.*` seeded.

## 4. Verification checklist
- [ ] build/vet pass; migration up/down clean.
- [ ] CRUD round-trips; `user_id` from Claims (not form); `status_kegiatan oneof 0 1`.
- [ ] `/public/kegiatan` returns only `status_kegiatan=1`, no private data.
- [ ] User read-only (writes → 403); `audit.Write` on changes.
