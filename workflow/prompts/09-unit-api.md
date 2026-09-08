# PROMPT — Build API module `unit` (member gun/unit registry) (Fase 2)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build the **`unit`** CRUD module. Follow the existing module pattern; `instansi`/`medsos` are references.

## 0. Read first
1. `slamteam_db.dbml` — **`unit`** (line 1039): `anggota_id` (not null FK), `kode`, `model`, `panjang/panjang_inbar/lebar/berat` `decimal(8,2)`, `berat_bb decimal(8,3)`, `fps decimal(8,2)`, `deskripsi_warna`, `foto_sampul_file_id`, `disetujui bool`, `disetujui_oleh`, `disetujui_pada`, soft-delete + audit. DBML WINS.
2. `workflow/modules/09-unit.md`; `_shared/conventions-api.md`; `_shared/api-endpoints.md` §13.

## 1. Scope + endpoints
CRUD `/unit` (`GET/GET:id/POST/PUT/DELETE`), `RequirePermission("unit.<aksi>")`. **All logged-in roles have CRUD** (permission-matrix) — apply cakupan `milik_sendiri` where the doc says User manages own units.

## 2. Files under `internal/modules/core/unit/`
- domain `Unit` (`unit`); dto (`CreateUnitReq`: `anggota_id required,gt=0`; `model max=100`; decimals `omitempty,gte=0`; `foto_sampul_file_id omitempty,gt=0`); repository (CRUD `is_deleted=false`, join foto uuid + `anggota_nama`, `q` ILIKE `kode|model`, sort `kode,model,created_at`, filter `anggota_id`, `disetujui`); service (validate anggota; `milik_sendiri` forces `anggota_id=claims.AnggotaID`; **approval** fields set by an approver action (`disetujui/_oleh/_pada`), not the owner; `audit.Write`); handler; main+router.
- Add an approve action if the doc specifies (e.g. `PUT /unit/:id` toggling `disetujui` gated to an approver permission/role).

## 3. Migration
`unit` EXACTLY per `.dbml` (FK `anggota_id → anggota.id`, `foto_sampul_file_id → mst_file.id`, `disetujui_oleh → users.id`). After `anggota`, `mst_file`. Confirm `unit.*` permissions seeded.

## 4. Verification checklist
- [ ] build/vet pass; migration up/down clean; matches `.dbml` (decimal scales).
- [ ] CRUD round-trips; decimals stored correctly; foto via `/files`.
- [ ] User can manage own units (`milik_sendiri`); cannot touch others'.
- [ ] Approval fields only settable by the approver; `audit.Write` on changes; no raw errors leaked.
