# PROMPT — Build API module `lokasi` (`mst_lokasi`, geofence) (Fase 3)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build the **`lokasi`** CRUD module (`mst_lokasi` — attendance locations with geofence + IANA timezone).

## 0. Read first
1. `slamteam_db.dbml` — **`mst_lokasi`** (line 717): `kode` (unique), `nama`, `jenis_lokasi`, `alamat`, `latitude/longitude decimal(10,7)` (**CHECK ±90 / ±180**), `radius_meter int` (default 100, **CHECK > 0**), `timezone` (IANA, default Asia/Jakarta), `foto_file_id`, `keterangan`, `is_aktif`, soft-delete + audit. DBML WINS.
2. `workflow/modules/15-lokasi.md`; `_shared/conventions-api.md` §9 (timezone); `_shared/api-endpoints.md` §5.

## 1. Scope + endpoints
CRUD `/lokasi`, `RequirePermission("lokasi.<aksi>")`. Band: SA/Admin CRUD, Mod/User R, Guest none. Referenced by `jadwal.lokasi_id` [**restrict**] → guard delete.

## 2. Files under `internal/modules/core/lokasi/`
- domain `Lokasi` (`mst_lokasi`, `TableName`; `Latitude/Longitude float64`); dto (`CreateLokasiReq`: `kode required,max=50`; `nama required,max=150`; `latitude required,latitude`; `longitude required,longitude`; `radius_meter required,min=1,max=100000`; `timezone required`; `foto_file_id omitempty,gt=0`); repository (CRUD `is_deleted=false`, join `foto_uuid` + subselect `COUNT(jadwal WHERE lokasi_id=… AND is_deleted=false)` as `jumlah_jadwal`, `q` ILIKE `kode|nama|alamat`, sort `kode,nama,created_at`, filter `is_aktif`; `ExistsKode`, `CountJadwalAktif`); service (unique kode → 409; `time.LoadLocation(timezone)` invalid → `ErrValidation`; **guard delete** `CountJadwalAktif>0` → `ErrConflict`; `audit.Write`); handler; main+router.

## 3. Migration
`mst_lokasi` EXACTLY per `.dbml` including **CHECK** constraints (coords range, `radius_meter>0`), partial-unique `kode`. FK `foto_file_id → mst_file`. Order **before** `jadwal`. Optional idempotent seeder (secretariat lokasi). Confirm `lokasi.*` seeded.

## 4. Verification checklist
- [ ] build/vet pass; migration up/down clean incl. CHECK constraints.
- [ ] CRUD round-trips; duplicate `kode` → 409; out-of-range coords / `radius<=0` / bad timezone → 422.
- [ ] Deleting a lokasi still used by a jadwal → **409** (not 500).
- [ ] Mod/User read-only (writes → 403); `audit.Write`; foto via `/files`.
