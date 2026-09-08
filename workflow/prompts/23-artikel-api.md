# PROMPT — Build API module `artikel` (blog + public) (Fase 8)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build the **`artikel`** CRUD module + public read.

## 0. Read first
1. `slamteam_db.dbml` — **`artikel`** (line 1022): `user_id` (author, not null), `kode`, `judul`, `gambar_highlight_file_id`, `konten`, **`status_kegiatan varchar(50)`** (NOTE: the status column is named `status_kegiatan` and is a **varchar** here — copy verbatim, do NOT rename to `status_artikel`), soft-delete + audit. DBML WINS.
2. `workflow/modules/23-artikel.md`; `_shared/api-endpoints.md` §13 + §12.

## 1. Scope + endpoints
CRUD `/artikel` (`RequirePermission("artikel.<aksi>")`) + `GET /public/artikel` (public: list `status_kegiatan='terbit'`, detail via `?slug=`). Band: SA/Admin/Mod CRUD, User R, Guest R.

## 2. Files under `internal/modules/core/artikel/`
- domain `Artikel` (`StatusKegiatan string`); dto (`CreateArtikelReq`: `judul required,max=200`; `kode max=50`; `status_kegiatan oneof=draft terbit`; `gambar_highlight_file_id omitempty,gt=0`); repository (CRUD `is_deleted=false`, join `gambar_uuid`, `q` ILIKE `judul`, sort `judul,created_at`, public `status_kegiatan='terbit'`, detail by `slug` derived from `kode`/`judul`); service (`user_id=claims.UserID`; `audit.Write`); handler (+ public); main+router.

## 3. Migration
`artikel` EXACTLY per `.dbml` — status column = **`status_kegiatan varchar(50)`**. FK `user_id`, `gambar_highlight_file_id`. After `users`, `mst_file`. Confirm `artikel.*` seeded.

## 4. Verification checklist
- [ ] build/vet pass; migration matches `.dbml` (status column is **`status_kegiatan varchar`**, not `status_artikel`).
- [ ] CRUD round-trips; `user_id` from Claims; `slug` from kode/judul.
- [ ] `/public/artikel` lists only `terbit`; detail via `?slug=`; draft hidden.
- [ ] User read-only (writes → 403); `audit.Write` on changes.
