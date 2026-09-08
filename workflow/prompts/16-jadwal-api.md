# PROMPT — Build API module `jadwal` + `jadwal_sesi` (recurrence + generator) (Fase 3)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build the **`jadwal`** module (schedule definition + recurrence) and the **`jadwal_sesi`** generator. This is domain-heavy — read carefully.

## 0. Read first
1. `slamteam_db.dbml` — **`jadwal`** (line 739) and **`jadwal_sesi`** (line 785); relations (`jadwal.lokasi_id → mst_lokasi.id` **restrict**, `kegiatan_id`, `inorga_id`; `jadwal_sesi.jadwal_id → jadwal.id` restrict). DBML WINS.
2. `workflow/modules/16-jadwal.md` (full spec incl. the generate-sesi algorithm).
3. `_shared/api-endpoints.md` §6 (generate-sesi request/response verbatim) + `_shared/conventions-api.md` §9 (timezone: `_ "time/tzdata"`, `time.LoadLocation`).

## 1. Scope + endpoints
| Method | Path | Permission |
|--------|------|-----------|
| GET/GET:id/POST/PUT/DELETE | `/jadwal[/:id]` | `jadwal.read/create/update/delete` |
| POST | `/jadwal/:id/generate-sesi` | `jadwal.update` |
| GET | `/jadwal/:id/sesi` | `jadwal.read` |
| PATCH | `/sesi/:id/batalkan` | `jadwal.batal_sesi` |
Band: SA/Admin/Mod CRUD, User R; `jadwal.batal_sesi` = SA/Admin/Mod.

## 2. Files under `internal/modules/core/jadwal/`
- **domain** `Jadwal` (map all columns; `HariUlang` as `pq.Int64Array`/`[]int64` for `int[]`; `JamMulai/JamSelesai` `string` `type:time`; enums `pola_ulang`/`mode_absen` string types; `status varchar`), `Sesi` (`jadwal_sesi`; enum `status_sesi`).
- **dto** `CreateJadwalReq`/`UpdateJadwalReq` (`kode required`; `nama required`; validate `jam_selesai>jam_mulai` in service; `pola_ulang oneof=tidak_berulang harian mingguan bulanan kustom`; `hari_ulang dive,min=0,max=6`; `mode_absen oneof=masuk_saja masuk_pulang`; `status oneof=draft terbit dibatalkan selesai`), `GenerateSesiReq` (per §6: `dari_tanggal, sampai_tanggal, timezone, mulai_lokal, selesai_lokal, absen_buka_menit_sebelum, absen_tutup_menit_setelah`), `JadwalResp`, `SesiResp`, `ListJadwalQuery` (filter `status`,`lokasi_id`, date range).
- **repository** CRUD (`is_deleted=false`), `ExistsKode`; sesi `CreateBatch` (INSERT … ON CONFLICT (jadwal_id, tanggal_lokal) DO NOTHING), `ListByJadwal`, `Cancel(id, alasan)`.
- **service**:
  - Create/Update: validate `jam_selesai>jam_mulai`, `time.LoadLocation(timezone)`; snapshot lat/long/radius from lokasi when empty; unique kode (409); `audit.Write`.
  - **generate-sesi (THE algorithm):** iterate dates per `pola_ulang`/`hari_ulang`/`interval_ulang` up to `tanggal_akhir_ulang`/`sampai_tanggal`. For each date: combine `tanggal_lokal`+`mulai_lokal` in the schedule `loc` → `mulai_utc` (`.UTC()`); same for selesai; `absen_buka_utc = mulai_utc - buka menit`, `absen_tutup_utc = mulai_utc + tutup menit`. Batch insert **idempotent** on `(jadwal_id, tanggal_lokal)`; return `{dibuat, dilewati, sesi[]}`. **Single-date schedule → exactly one sesi.**
  - batalkan: `status=dibatalkan` + `alasan_batal` (required); `audit.Write`.
- **handler/main/router**: RequirePermission per table; register.

## 3. Migration
`jadwal` + `jadwal_sesi` EXACTLY per `.dbml`: enum types `pola_ulang`/`mode_absen`/`status_sesi`; `int[]` for `hari_ulang`; CHECK `jam_selesai>jam_mulai`; `UNIQUE(jadwal_id,tanggal_lokal)`; FK restrict. Order **after** `mst_lokasi` (+ `kegiatan`, `mst_inorga` for optional FKs). Confirm `jadwal.*` + `jadwal.batal_sesi` seeded.

## 4. Verification checklist
- [ ] build/vet pass; migration up/down clean; matches `.dbml` (enum, `int[]`, CHECK, unique).
- [ ] generate-sesi produces correct dates per pola; local→UTC uses the **schedule** timezone; **idempotent** (re-run adds no duplicates).
- [ ] Single-date schedule → exactly one sesi row.
- [ ] `PATCH /sesi/:id/batalkan` requires `alasan_batal`; `jadwal.batal_sesi` blocked for User (403).
- [ ] User can `GET /jadwal` but not create/assign/cancel; `audit.Write` on changes; no raw errors leaked.
