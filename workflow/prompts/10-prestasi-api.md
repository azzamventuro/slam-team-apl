# PROMPT — Build API module `prestasi` (achievements, public-readable) (Fase 2)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build the **`prestasi`** CRUD module (member competition achievements, **public-readable incl. Guest**).

## 0. Read first
1. `slamteam_db.dbml` — **`prestasi`** (line 1064): `anggota_id` (not null FK), `kode`, `peringkat`, `tingkat`, `judul_kompetisi`, `flyer_file_id`, `tanggal_kompetisi date`, `alamat_kompetisi`, `foto_sampul_file_id`, `keterangan`, soft-delete + audit. DBML WINS.
2. `workflow/modules/10-prestasi.md`; `_shared/conventions-api.md`; `_shared/api-endpoints.md` §13 + §12 (public via `/public/profil`).

## 1. Scope + endpoints
CRUD `/prestasi`, `RequirePermission("prestasi.<aksi>")`. Band: SA/Admin/Mod CRUD, User R, **Guest R**. Public exposure via the landing module (`/public/profil`), not a private endpoint.

## 2. Files under `internal/modules/core/prestasi/`
- domain `Prestasi`; dto (`CreatePrestasiReq`: `anggota_id required,gt=0`; `judul_kompetisi max=200`; `tanggal_kompetisi omitempty,datetime=2006-01-02`; file ids `omitempty,gt=0`); repository (CRUD `is_deleted=false`, join `flyer_uuid`/`foto_uuid` + `anggota_nama`, `q` ILIKE `judul_kompetisi|tingkat|peringkat`, sort `tanggal_kompetisi,created_at`, filter `anggota_id`,`tingkat`); service (validate anggota; `audit.Write`); handler; main+router.

## 3. Migration
`prestasi` EXACTLY per `.dbml` (FK `anggota_id`, two `*_file_id → mst_file`). After `anggota`, `mst_file`. Confirm `prestasi.*` seeded (Guest R in matrix).

## 4. Verification checklist
- [ ] build/vet pass; migration up/down clean.
- [ ] CRUD round-trips; two file refs stored; detail returns uuids.
- [ ] User read-only (writes → 403); rows readable publicly via landing; `audit.Write` on changes.
