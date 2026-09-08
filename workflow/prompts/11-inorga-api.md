# PROMPT — Build API module `inorga` (`mst_inorga`) (Fase 2)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build the **`inorga`** CRUD module (`mst_inorga` — internal-org periods with an SK document).

## 0. Read first
1. `slamteam_db.dbml` — **`mst_inorga`** (line 1103): `kode`, `nama`, `logo_file_id`, `banner_file_id`, `tanggal_mulai/tanggal_selesai date`, `file_sk_file_id`, `konten`, soft-delete + audit. DBML WINS.
2. `workflow/modules/11-inorga.md`; `_shared/conventions-api.md`; `_shared/api-endpoints.md` §13.

## 1. Scope + endpoints
CRUD `/inorga`, `RequirePermission("inorga.<aksi>")`. Band: SA/Admin CRUD, Moderator R, User R, Guest R.

## 2. Files under `internal/modules/core/inorga/`
- domain `Inorga` (`mst_inorga`, `TableName`); dto (`CreateInorgaReq`: `nama max=150`; `kode max=50`; dates `omitempty,datetime=2006-01-02`; three file ids `omitempty,gt=0`); repository (CRUD `is_deleted=false`, join `logo_uuid`/`banner_uuid`/`file_sk_uuid`, `q` ILIKE `kode|nama`, sort `nama,tanggal_mulai,created_at`); service (`audit.Write`); handler; main+router.
- Referenced by `jadwal.inorga_id` — no delete guard required beyond soft-delete, but note the reference.

## 3. Migration
`mst_inorga` EXACTLY per `.dbml` (three `*_file_id → mst_file`). After `mst_file`. Confirm `inorga.*` seeded.

## 4. Verification checklist
- [ ] build/vet pass; migration up/down clean.
- [ ] CRUD round-trips; three file refs (logo/banner/SK) stored; detail returns uuids; `file_sk` treated as document.
- [ ] Moderator/User read-only (writes → 403); `audit.Write` on changes.
