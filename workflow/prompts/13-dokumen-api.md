# PROMPT — Build API module `dokumen` (`mst_dokumen`, polymorphic) (Fase 2)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build the **`dokumen`** CRUD module (`mst_dokumen` — polymorphic private document attachments).

## 0. Read first
1. `slamteam_db.dbml` — **`mst_dokumen`** (line 1139): `kode`, `file_id` (FK mst_file — the document), `tipe varchar(50)`, `format varchar(20)`, `reff_id`, `reff_type varchar(50)` (polymorphic owner), `jenis int`, `keterangan varchar(255)`, soft-delete + audit. DBML WINS.
2. `workflow/modules/13-dokumen.md`; `_shared/conventions-api.md` §10 (file layer); `_shared/api-endpoints.md` §13.

## 1. Scope + endpoints
CRUD `/dokumen`, `RequirePermission("dokumen.<aksi>")`. Band: SA/Admin/Mod CRUD, User R, Guest none. Documents are **PRIVATE** files (non-image → `original` only), served via `GET /files/:uuid/:varian`.

## 2. Files under `internal/modules/core/dokumen/`
- domain `Dokumen` (`mst_dokumen`, `TableName`); dto (`CreateDokumenReq`: `file_id required,gt=0`; `tipe max=50`; `format max=20`; `reff_type max=50`; `jenis omitempty`; `keterangan max=255`); repository (CRUD `is_deleted=false`, join `file_uuid`+`nama_asli`, filter `reff_type`,`reff_id`,`tipe`,`jenis`, `q` ILIKE `kode|keterangan`); service (validate `file_id` exists; `audit.Write`); handler; main+router.

## 3. Migration
`mst_dokumen` EXACTLY per `.dbml` (FK `file_id → mst_file`, index `(reff_type, reff_id)`). After `mst_file`. Confirm `dokumen.*` seeded.

## 4. Verification checklist
- [ ] build/vet pass; migration up/down clean.
- [ ] CRUD round-trips; `file_id` validated; detail returns `file_uuid`+`nama_asli`.
- [ ] Polymorphic filter by `reff_type`/`reff_id` works; docs served only through the authenticated file endpoint.
- [ ] User read-only (writes → 403); `audit.Write` on changes.
