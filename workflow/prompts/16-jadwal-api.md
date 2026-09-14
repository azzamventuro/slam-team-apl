# PROMPT — Build API `jadwal` + `jadwal_sesi` (recurrence + generate-sesi) (Fase 3) ★

> **Repo-accurate (revised 2026-09-13).** RBAC, `mst_lokasi`, `mst_inorga`, audit, shared helpers all built; `jadwal.*`+`assign`+`batal_sesi` seeded; migration `0015`. `kegiatan` NOT built → defer that FK.

Working directory: `D:/xampp/htdocs/slam-team-apl/slam-team-api`. Domain-heavy — read carefully.

## 0. Read first (authoritative)
1. `slamteam_db.dbml` — **`jadwal`** (line 739) and **`jadwal_sesi`** (line 785). DBML WINS.
2. `workflow/modules/16-jadwal.md`; `_shared/api-endpoints.md` §6 (generate-sesi verbatim); `_shared/conventions-api.md` §9 (timezone — `_ "time/tzdata"` already imported; use `time.LoadLocation`).
3. Reference: `internal/modules/core/lokasi/` (same `Initialize(db, jwtMgr, perm, auditor)` shape) + `internal/shared/{model,pagination,apperr,response,audit}`.

## 1. Scope + endpoints (real RBAC — all perms already seeded)
| Method | Path | Permission |
|--------|------|-----------|
| GET/GET:id/POST/PUT/DELETE | `/jadwal[/:id]` | `jadwal.read/create/update/delete` |
| POST | `/jadwal/:id/generate-sesi` | `jadwal.update` |
| GET | `/jadwal/:id/sesi` | `jadwal.read` |
| PATCH | `/sesi/:id/batalkan` | `jadwal.batal_sesi` |
Band: SA/Admin/Mod CRUD, User read; `jadwal.batal_sesi` = SA/Admin/Mod (verify seed, don't re-seed).

## 2. Files under `internal/modules/core/jadwal/`
- **domain** `Jadwal` (all columns; `HariUlang` = `pq.Int64Array`/`[]int64` for `int[]`; `JamMulai/JamSelesai` `string` `type:time`; enums `pola_ulang`/`mode_absen` as string types; `status varchar`; embed `model.Audit`), `Sesi` (`jadwal_sesi`; enum `status_sesi`).
- **dto** `CreateJadwalReq`/`UpdateJadwalReq` (`kode required`; `nama required`; validate `jam_selesai>jam_mulai` in service; `pola_ulang oneof=tidak_berulang harian mingguan bulanan kustom`; `hari_ulang dive,min=0,max=6`; `mode_absen oneof=masuk_saja masuk_pulang`; `status oneof=draft terbit dibatalkan selesai`), `GenerateSesiReq` (per §6), `JadwalResp`, `SesiResp`, `ListJadwalQuery` (filter `status`, `lokasi_id`, date range).
- **repository** CRUD (`is_deleted=false`), `ExistsKode`; sesi `CreateBatch` (`INSERT … ON CONFLICT (jadwal_id, tanggal_lokal) DO NOTHING` → idempotent), `ListByJadwal`, `Cancel(id, alasan)`.
- **service**:
  - Create/Update: validate `jam_selesai>jam_mulai`, `time.LoadLocation(timezone)`; snapshot `latitude/longitude/radius_meter` from the referenced `mst_lokasi` when empty; unique `kode` → `apperr.ErrConflict`; `auditor.Write`.
  - **generate-sesi (THE algorithm):** iterate dates per `pola_ulang`/`hari_ulang`/`interval_ulang` up to `tanggal_akhir_ulang`/`sampai_tanggal`. Per date: combine `tanggal_lokal`+`mulai_lokal` in the schedule `loc` → `mulai_utc` (`.UTC()`); same for selesai; `absen_buka_utc = mulai_utc - buka menit`, `absen_tutup_utc = mulai_utc + tutup menit`. Batch insert **idempotent** on `(jadwal_id, tanggal_lokal)`; return `{dibuat, dilewati, sesi[]}`. **Single-date schedule → exactly ONE sesi.**
  - batalkan: `status=dibatalkan` + `alasan_batal` (required); `auditor.Write`.
- **handler/main.jadwal.go/router**: `Initialize(db, jwtMgr, permGuard, auditor)`; `RequirePermission` per the table; register in `router.go`.

## 3. Migration `0015_jadwal.up.sql` (+ `.down.sql`)
`jadwal` + `jadwal_sesi` columns EXACTLY per `.dbml`: enum types `pola_ulang`/`mode_absen`/`status_sesi` (from `0001_enums` — reference, don't recreate); `int[]` for `hari_ulang`; CHECK `jam_selesai>jam_mulai`; `UNIQUE(jadwal_id, tanggal_lokal)` on `jadwal_sesi`. FK targets that exist → immediate: `lokasi_id → mst_lokasi(id)` **RESTRICT**, `inorga_id → mst_inorga(id)`, `jadwal_sesi.jadwal_id → jadwal(id)` **RESTRICT**, audit `*_by → users(id)`. **DEFER `kegiatan_id` FK** — `kegiatan` is Fase 8; create the `kegiatan_id bigint` column now, leave a marked seam (`// TODO(22-kegiatan): ADD CONSTRAINT fk_jadwal_kegiatan`). `.down` drops both.

## 4. Close the `15-lokasi` seam (now that `jadwal` exists)
In `internal/modules/core/lokasi/`: implement the deferred usage check — raw count `SELECT count(*) FROM jadwal WHERE lokasi_id=? AND is_deleted=false` in the lokasi repo, enable the **409 delete-guard** in the lokasi service, populate `LokasiResp.jumlah_jadwal`. Remove the `// TODO(16-jadwal)` markers.

## 5. Conventions
Envelope via `response`; sentinels via `apperr` + `response.FromError`; no AutoMigrate; explicit soft-delete; whitelist sort; `per_page` cap 100; `auditor.Write`.

## 6. Verification checklist
- [ ] `go build ./...` + `go vet ./...` pass; `slamctl migrate up`/`down` clean (`0015`); matches `.dbml` (enum, `int[]`, CHECK, unique).
- [ ] generate-sesi correct per pola; local→UTC uses the **schedule** timezone; **idempotent**; single-date → one sesi.
- [ ] `PATCH /sesi/:id/batalkan` requires `alasan_batal`; `jadwal.batal_sesi` blocked for User; User can read but not create/assign/cancel.
- [ ] `lokasi_id`/`inorga_id`/audit FKs immediate; `kegiatan_id` FK deferred with a seam.
- [ ] Deleting a `mst_lokasi` used by an active jadwal now returns **409**; `jumlah_jadwal` reflects real count.
- [ ] `auditor.Write` rows written; no raw errors leak.
