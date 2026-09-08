# PROMPT — Build API module `absensi` (camera + GPS, server-time authority) (Fase 5)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build the **`absensi`** module. Server time is authoritative; device time is evidence only. Multipart selfie + GPS.

## 0. Read first
1. `slamteam_db.dbml` — **`absensi`** (line 853); relations (four FK **restrict**: `sesi_id`,`jadwal_id`,`anggota_id`,`user_id`; `foto_file_id`, `override_oleh`, `izin_id`). DBML WINS.
2. `workflow/modules/18-absensi.md` (full spec).
3. `_shared/api-endpoints.md` §8 (multipart fields + logic + response verbatim) + `_shared/conventions-api.md` §9,§10 (timezone, multipart/file layer, Bab 6.6 checklist).

## 1. Scope + endpoints
| Method | Path | Permission |
|--------|------|-----------|
| GET | `/absensi/sesi-aktif` | `absensi.create` |
| POST | `/absensi/check-in` | `absensi.create` (**multipart**) |
| POST | `/absensi/check-out` | `absensi.create` (**multipart**) |
| GET | `/absensi` | `absensi.read` (cakupan semua/milik_sendiri) |
| PATCH | `/absensi/:id/override` | `absensi.override` |
| DELETE | `/absensi/:id` | `absensi.delete` (**SA only**) |
(`/absensi/rekap`, `/absensi/export` live in module 21.)

## 2. Files under `internal/modules/core/absensi/`
- **domain** `Absensi` (map ALL columns; enums as string types; `flag_mencurigakan`/`info_perangkat` `datatypes.JSON`; `ip_address` inet).
- **dto** `CheckinForm` (multipart: `sesi_id, latitude, longitude, akurasi_meter, waktu_perangkat, timezone_perangkat, konfirmasi_luar_radius, selfie`), `OverrideReq{status_kehadiran, alasan_override required}`, `AbsensiResp`.
- **repository** `Create`, `FindByID`, `List(cakupan)`, `ExistsAbsen(sesi,anggota,tipe)`, `SesiAktifUntuk(userID)`, `Override`, `SoftDelete` — `is_deleted=false`.
- **service (rules):** caller must be account-holding + a **participant** of the sesi; sesi open (`now BETWEEN absen_buka_utc AND absen_tutup_utc`, **server time**); reject duplicate (`ExistsAbsen`). Upload selfie via file-service (Bab 6.6: `ParseMultipartForm` before `FormFile`) → `foto_file_id`. Haversine `jarak_meter` vs `radius_berlaku`; set `dalam_radius`; out-of-radius + `izinkan_luar_radius=false` → reject; out-of-radius + `konfirmasi_luar_radius=false` → ask confirmation. Derive `status_kehadiran` (masuk: hadir/terlambat[> toleransi]/hadir_luar_radius; pulang: pulang_cepat) + `menit_telat`/`menit_pulang_cepat`. Set `metode=selfie`, `waktu_server_utc`, `selisih_jam_detik`, `info_perangkat`, `ip_address`, `flag_mencurigakan`. Bump `jadwal_sesi.jml_hadir`. `audit.Write`.
- **override:** `absensi.override`; `metode=manual`, `is_override=true`, `alasan_override` required; no coords/photo.
- **session-closer job** (via `slamctl`/cron, `aktor_user_id` null): close sesi past `absen_tutup_utc`; mark `alfa` only for `jadwal_peserta.wajib_absen=true` with no check-in.
- **handler/main/router**: multipart binds; RequirePermission per table; `DELETE` = SA only.

## 3. Migration
`absensi` EXACTLY per `.dbml` (enum `tipe_absensi`/`status_kehadiran`/`metode_absensi`; jsonb; inet; `UNIQUE(sesi_id,anggota_id,tipe)`; four FK restrict; `uuid` default gen). After `jadwal_sesi`, `anggota`, `users`, `absensi_izin`. Confirm `absensi.*` seeded per matrix.

## 4. Verification checklist
- [ ] build/vet pass; migration up/down clean; matches `.dbml`.
- [ ] `waktu_server_utc` is authoritative; device time stored as evidence + `selisih_jam_detik`.
- [ ] Haversine + `dalam_radius` correct; out-of-radius requires `konfirmasi_luar_radius`.
- [ ] Duplicate check-in/out → 409; absen outside window → rejected; non-participant → 403.
- [ ] `status_kehadiran`/`menit_telat`/`menit_pulang_cepat` derived from the window.
- [ ] Override requires `alasan_override`; `absensi.delete` SA-only; User `absensi.read` = own only (cakupan).
- [ ] session-closer marks `alfa` only for `wajib_absen=true`; multipart follows Bab 6.6; `audit.Write`.
