# Modul 18 — Absensi (kamera + GPS) (`absensi`)

> Sumber otoritatif: `slamteam_db.dbml` (skema) > rancangan Bab 5.1–5.6 > `_shared/api-endpoints.md` §8 > `_shared/conventions-*.md` > aplikasi lama (tidak pernah). Jika prosa dan `.dbml` berbeda, `.dbml` menang.

---

## 1. Ringkasan & tujuan modul

**Absensi** adalah rekaman check-in/check-out yang terikat ke **`jadwal_sesi`**. Modul ke-17 dari 19 (`absensi`), grup **Operasional**, **Fase 5**. Ini **layar tersulit** (rancangan Bab 11.3): selfie kamera depan + GPS pada **secure context (HTTPS)**.

Tiga prinsip yang membedakannya dari modul lain:

1. **Waktu server otoritatif.** `waktu_server_utc` adalah acuan resmi; `waktu_perangkat` hanya bukti (+`selisih_jam_detik`). **Jangan percaya jam telepon.**
2. **Validasi lokasi Haversine.** `jarak_meter` dihitung dari koordinat vs `radius_berlaku` (snapshot radius sesi/lokasi) → set `dalam_radius`. Di luar radius butuh konfirmasi eksplisit.
3. **Hanya anggota berakun yang absen.** `user_id NOT NULL`. Anggota tanpa akun tidak pernah ditandai `alfa` (dijaga `jadwal_peserta.wajib_absen`).

Dua jalur pembuatan baris: **selfie** (normal — koordinat/foto/dalam_radius terisi) dan **manual** (override moderator — koordinat/foto kosong, `is_override=true`, `alasan_override` wajib).

---

## 2. Tabel & kolom

### `absensi` (DBML baris 853)

| Kolom | Tipe | Aturan / catatan |
|-------|------|------------------|
| `id` | `bigint` [pk] | |
| `uuid` | `uuid` [unique, default gen] | Identitas publik. |
| `sesi_id` | `bigint` [not null] | FK → `jadwal_sesi.id` (**restrict**). |
| `jadwal_id` | `bigint` [not null] | FK → `jadwal.id` (**restrict**). |
| `anggota_id` | `bigint` [not null] | Subjek. FK → `anggota.id` (**restrict**). |
| `user_id` | `bigint` [not null] | Akun anggota. FK → `users.id` (**restrict**). |
| `tipe` | `tipe_absensi` [not null] | masuk / pulang. |
| `waktu_server_utc` | `timestamptz` [not null] | **ACUAN RESMI**. |
| `waktu_perangkat` | `timestamptz` | Bukti. |
| `timezone_jadwal` | `varchar(50)` [not null] | |
| `timezone_perangkat` | `varchar(50)` | |
| `selisih_jam_detik` | `int` | server − perangkat. |
| `latitude`/`longitude` | `decimal(10,7)` | Kosong hanya untuk override manual. |
| `akurasi_meter` | `decimal(8,2)` | Akurasi GPS. |
| `jarak_meter` | `decimal(10,2)` | Haversine ke titik. |
| `radius_berlaku` | `int` | Radius saat absen (snapshot). |
| `dalam_radius` | `boolean` | Kosong hanya untuk manual. |
| `konfirmasi_luar_radius` | `boolean` [default false] | true bila user tekan OK peringatan. |
| `alamat_perkiraan` | `text` | |
| `foto_file_id` | `bigint` | Selfie, **PRIVAT**. Kosong untuk manual. |
| `status_kehadiran` | `status_kehadiran` [not null] | hadir/terlambat/pulang_cepat/hadir_luar_radius/izin/sakit/dinas/alfa. |
| `menit_telat` | `int` [default 0] | |
| `menit_pulang_cepat` | `int` [default 0] | |
| `izin_id` | `bigint` | FK → `absensi_izin.id`. |
| `metode` | `metode_absensi` [default `selfie`] | selfie/qr/manual. |
| `info_perangkat` | `jsonb` | |
| `ip_address` | `inet` | |
| `flag_mencurigakan` | `jsonb` | jam_perangkat_beda / akurasi_rendah / exif_tidak_cocok / perangkat_dipakai_banyak_akun. |
| `is_override` | `boolean` [default false] | |
| `override_oleh` | `bigint` | FK → `users.id`. |
| `override_pada` | `timestamptz` | |
| `alasan_override` | `text` | **Wajib** untuk manual. |
| `catatan` | `text` | |
| soft-delete + audit | | |

**Index:** `UNIQUE(sesi_id, anggota_id, tipe)` (satu masuk + satu pulang per sesi), `(anggota_id, waktu_server_utc)`, `(jadwal_id, status_kehadiran)`, `sesi_id`, `uuid`.

**Relasi:** empat FK **restrict** (`sesi_id`, `jadwal_id`, `anggota_id`, `user_id`); `foto_file_id → mst_file.id`; `override_oleh → users.id`; `izin_id → absensi_izin.id`.

---

## 3. Endpoint (api-endpoints §8)

| Method | Path | Permission | Auth | Body |
|--------|------|-----------|------|------|
| `GET` | `/absensi/sesi-aktif` | `absensi.create` | JWT | — → sesi yang jendelanya terbuka bagi caller |
| `POST` | `/absensi/check-in` | `absensi.create` | JWT | **multipart** (selfie+GPS) |
| `POST` | `/absensi/check-out` | `absensi.create` | JWT | **multipart** |
| `GET` | `/absensi` | `absensi.read` | JWT | cakupan semua/milik_sendiri |
| `GET` | `/absensi/rekap` | `absensi.read` | JWT | (modul 21) |
| `PATCH` | `/absensi/{id}/override` | `absensi.override` | JWT | `{status_kehadiran, alasan_override}` |
| `POST` | `/absensi/export` | `absensi.export` | JWT | (modul 21) |
| `DELETE` | `/absensi/{id}` | `absensi.delete` | JWT | **SA only** (soft delete) |

### `check-in`/`check-out` — multipart/form-data (Bab 6.6)
Angular kirim `FormData` **tanpa** set `Content-Type`. Field: `sesi_id`, `latitude`, `longitude`, `akurasi_meter`, `waktu_perangkat`, `timezone_perangkat`, `konfirmasi_luar_radius`, `selfie(file)`. Go: `ParseMultipartForm` **sebelum** `FormFile`; cek `io.Copy`/`os.Create`.

Server: `waktu_server_utc = now()` (acuan); hitung `selisih_jam_detik`; Haversine `jarak_meter` vs `radius_berlaku` → `dalam_radius`; turunkan `status_kehadiran` + `menit_telat`/`menit_pulang_cepat` dari jendela sesi; simpan `metode=selfie`, `foto_file_id`, `info_perangkat`, `ip_address`, `flag_mencurigakan`. `UNIQUE(sesi_id,anggota_id,tipe)`.

Respons `data`: `{id, uuid, tipe, status_kehadiran, menit_telat, waktu_server_utc, jarak_meter, dalam_radius, radius_berlaku, foto_url, flag_mencurigakan}`.

### `override`
`{status_kehadiran, alasan_override}` → `metode=manual`, `is_override=true`, `override_oleh/pada`; **tanpa** koordinat/foto; `alasan_override` **wajib**.

---

## 4. Hak akses (rancangan Bab 3.5)

| Izin | SA | Admin | Moderator | User |
|------|----|-------|-----------|------|
| `absensi.create` | ✔ | ✔ | ✔ | ✔ |
| `absensi.read` | ✔ `semua` | ✔ `semua` | ✔ `semua` | ✔ `milik_sendiri` |
| `absensi.override` | ✔ | ✔ | — | — |
| `absensi.delete` | ✔ | — | — | — |
| `absensi.export` | ✔ | ✔ | ✔ | — |

Cakupan ditegakkan service/repository: User `absensi.read` hanya `anggota_id=claims.AnggotaID`.

---

## 5. Kebutuhan BACKEND (Go)

Modul `internal/modules/core/absensi/`.

### domain / dto
- `Absensi` (map `absensi`; enum sebagai string types; `Flag`/`InfoPerangkat` `datatypes.JSON`). `CheckinForm` (multipart — bind via `ShouldBind` field-per-field), `OverrideReq{status_kehadiran, alasan_override required}`, `AbsensiResp`.

### repository
- `Create`, `FindByID`, `List` (cakupan), `ExistsAbsen(sesi_id, anggota_id, tipe)`, `SesiAktifUntuk(userID)`, `Override`, `SoftDelete`. Semua `is_deleted=false`.

### service (inti aturan)
- **check-in/out:** pastikan caller **berakun** & **ditugaskan** (peserta) pada sesi; sesi ada & **jendela terbuka** (`now BETWEEN absen_buka_utc AND absen_tutup_utc` — pakai waktu server). Cegah duplikat (`ExistsAbsen`). Upload selfie via file-service → `foto_file_id`. Hitung Haversine, `dalam_radius`; jika di luar radius & `izinkan_luar_radius=false` → tolak; jika true & `konfirmasi_luar_radius=false` → minta konfirmasi (422/khusus). Turunkan `status_kehadiran`: masuk → `hadir`/`terlambat` (telat > toleransi)/`hadir_luar_radius`; pulang → `pulang_cepat` bila lebih awal. Isi `flag_mencurigakan` (selisih jam besar, akurasi rendah, dst). Simpan; update `jadwal_sesi.jml_hadir`. Log.
- **override:** hanya `absensi.override`; set status manual + `alasan_override` wajib; koordinat/foto kosong; log.
- **session-closer (job terjadwal):** untuk sesi lewat `absen_tutup_utc`, set sesi `selesai`, dan tandai `alfa` untuk `jadwal_peserta.wajib_absen=true` yang tak punya baris masuk. Jalankan lewat `slamctl`/cron; `aktor_user_id` null.

### handler / main / router
- Rute multipart (`ShouldBind` + FormFile). `RequirePermission` sesuai tabel. `DELETE` → `absensi.delete` (SA only). Daftar di router.

### migrations
- `absensi` **persis `.dbml`** (enum `tipe_absensi`/`status_kehadiran`/`metode_absensi`; `inet`; `jsonb`; `UNIQUE(sesi_id,anggota_id,tipe)`; empat FK restrict). **Setelah** `jadwal_sesi`, `anggota`, `users`, `absensi_izin` (`izin_id`).

### edge cases
- Absen di luar jendela → tolak (server time). Duplikat masuk → 409. Anggota tak berakun/tak ditugaskan → 403. GPS di luar radius tanpa konfirmasi → minta konfirmasi. `DELETE` oleh non-SA → 403. Foto gagal upload → tolak transaksi.

---

## 6. Kebutuhan FRONTEND (Angular)

**WAJIB taste-skill.** Ini layar absen — **bangun lebih dulu**. Warna status per design-tokens (hadir=success, telat/pulang_cepat/luar_radius=warning, izin/sakit/dinas=info, alfa=danger).

### Layar absen (`pages/absensi/absen.*`)
- Ambil sesi aktif (`GET /absensi/sesi-aktif`). **Secure context (HTTPS) wajib** — `getUserMedia`/`geolocation` gagal di HTTP polos.
- Preview **kamera depan**, tombol ambil selfie; baca **GPS** (akurasi), tampilkan jarak ke lokasi.
- **Dialog konfirmasi luar radius** (set `konfirmasi_luar_radius`). **Panduan izin ditolak** (kamera/GPS diblok): instruksi mengaktifkan.
- Submit **multipart** (selfie + field) tanpa set `Content-Type`. Kartu hasil menampilkan `waktu_server_utc`, status, jarak.

### Layar admin (`pages/absensi/absensi-list.*`)
- Daftar absensi (cakupan), aksi **override** (moderator/admin) dengan `alasan_override` wajib.

### Service, i18n
- `absensi.service.ts` (sesiAktif, checkIn, checkOut, list, override). i18n `ABSENSI` (status, telat, luar radius, izin kamera).

### Gating
- Absen `*hasPermission="'absensi.create'"`; override `*hasPermission="'absensi.override'"`; delete `*hasPermission="'absensi.delete'"`.

---

## 7. ALUR form → API → DATABASE

| Aksi | Endpoint | Tulisan DB |
|------|----------|------------|
| Upload selfie | `POST /files` | INSERT `mst_file` (kategori=absensi, privat) + 3 varian |
| Check-in | `POST /absensi/check-in` | INSERT `absensi` (waktu_server_utc, jarak, status); UPDATE `jml_hadir`; `log_aktivitas` |
| Override | `PATCH /absensi/{id}/override` | UPDATE (metode=manual, is_override, alasan) + log |
| Job penutup | (cron) | UPDATE sesi=selesai; INSERT `absensi` alfa untuk wajib_absen tak hadir |

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- **Fase 0** file layer (selfie privat) + `mst_pengaturan` (radius/toleransi). **Jadwal (16)** + **Penugasan (17)** (peserta & wajib_absen). **Izin (19)** untuk `izin_id`. RBAC.

### Acceptance criteria
- [ ] Migrasi cocok `.dbml` (enum, jsonb, `UNIQUE(sesi_id,anggota_id,tipe)`).
- [ ] `waktu_server_utc` dipakai sebagai acuan; jam perangkat hanya disimpan sebagai bukti.
- [ ] Haversine + `dalam_radius` benar; luar radius butuh `konfirmasi_luar_radius`.
- [ ] Duplikat check-in/out → 409; absen di luar jendela → ditolak; anggota tak ditugaskan → 403.
- [ ] `status_kehadiran`/`menit_telat`/`menit_pulang_cepat` diturunkan dari jendela sesi.
- [ ] Override butuh `alasan_override`; `absensi.delete` hanya SA.
- [ ] User `absensi.read` hanya melihat miliknya (cakupan).
- [ ] Layar absen di secure context; taste-skill diterapkan; warna status sesuai token.
