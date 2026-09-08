# PROMPT — Build API laporan/rekap + export (Fase 7)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Add **rekap** + **export** to the `absensi` module. `absensi_rekap` is OPTIONAL — compute directly from `absensi` first.

## 0. Read first
1. `slamteam_db.dbml` — **`absensi`** (853) as the data source; **`absensi_rekap`** (956, OPTIONAL). DBML WINS.
2. `workflow/modules/21-laporan-rekap.md`; `_shared/api-endpoints.md` §8 (`/absensi/rekap`, `/absensi/export`).

## 1. Scope + endpoints
| Method | Path | Permission |
|--------|------|-----------|
| GET | `/absensi/rekap` | `absensi.read` (cakupan semua/milik_sendiri) |
| POST | `/absensi/export` | `absensi.export` |

## 2. Files (extend `internal/modules/core/absensi/`)
- dto: `RekapQuery{periode, jadwal_id?, anggota_id?}`, `RekapRow{anggota_id, nama, jml_wajib, jml_hadir, jml_telat, jml_pulang_cepat, jml_izin, jml_sakit, jml_alfa, total_menit_telat, persen_kehadiran}`, `ExportReq{format oneof=xlsx pdf, periode, jadwal_id?, anggota_id?}`.
- repository: `Rekap(query, cakupan)` — GROUP BY `anggota_id` from `absensi` (filter periode via session/`waktu_server_utc`, `is_deleted=false`), count per status, join `jadwal_peserta` for `jml_wajib`; `persen = hadir/wajib*100` (guard divide-by-zero). Cakupan `milik_sendiri` → `anggota_id=claims.AnggotaID`. Optional `UpsertRekap(periode)` → `absensi_rekap` (`ON CONFLICT (anggota_id,periode)`).
- service: compute rekap; **export** renders Excel (e.g. `excelize`) or PDF (e.g. `gofpdf`/`maroto`) — record the library in `workflow/99-alignment-report.md`; return binary (Content-Disposition) or a generated file URL; `audit.Write` (aksi export).
- handler: `GET /absensi/rekap` → `absensi.read`; `POST /absensi/export` → `absensi.export`.

## 3. Migration
Only if materializing: `absensi_rekap` EXACTLY per `.dbml` (`UNIQUE(anggota_id,periode)`). Otherwise none. Confirm `absensi.export` seeded.

## 4. Verification checklist
- [ ] build/vet pass.
- [ ] Rekap counts correct per status + `persen_kehadiran`; empty period → empty array (no error); divide-by-zero guarded.
- [ ] Export `xlsx` and `pdf` produce valid files; unknown format → 422.
- [ ] User `absensi.read` rekap = own only (cakupan); `absensi.export` blocked for User (403); `audit.Write` on export.
