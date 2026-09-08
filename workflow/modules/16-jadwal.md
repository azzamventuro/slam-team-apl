# Modul 16 — Jadwal + Generator Sesi (`jadwal`)

> Sumber otoritatif: `slamteam_db.dbml` (skema) > rancangan Bab 4.1–4.5 (empat tabel, atribut jadwal, sesi, penugasan, zona waktu) > `_shared/api-endpoints.md` §6 > `_shared/conventions-*.md` > aplikasi lama (tidak pernah). Jika prosa dan `.dbml` berbeda, `.dbml` menang.

---

## 1. Ringkasan & tujuan modul

**Jadwal** adalah definisi kegiatan/latihan dengan **aturan pengulangan** dan **aturan absensi**; **`jadwal_sesi`** adalah instance tanggal konkret yang dihasilkan dari pengulangan itu — **titik sandar absensi** (absensi mengikat ke `jadwal_sesi`, tidak pernah langsung ke `jadwal`). Modul ke-16 (`jadwal`), grup **Operasional**, **Fase 3**.

Ini salah satu dari **3 layar tersulit** (rancangan Bab 11.3): kalender jadwal yang **sadar zona waktu** (tampil 07:00 WIB memakai `jadwal.timezone`, bukan zona browser). Dua hal krusial:

1. **Generator sesi (`generate-sesi`) mengubah aturan berulang menjadi baris `jadwal_sesi`** dengan konversi jam lokal → UTC memakai `jadwal.timezone` (Go: `time.LoadLocation` + `_ "time/tzdata"`). Idempoten pada `UNIQUE(jadwal_id, tanggal_lokal)`. Jadwal tanggal tunggal tetap menghasilkan **satu** sesi.
2. **Snapshot geofence.** `jadwal.latitude/longitude/radius_meter` menyalin (dan boleh menimpa) nilai `mst_lokasi` saat jadwal dibuat, sehingga aturan absensi lama tetap konsisten walau lokasi berubah.

Prasyarat: **Lokasi (15)** untuk `lokasi_id` (restrict), **RBAC (Fase 1)**, dan konvensi timezone Fase 0.

---

## 2. Tabel & kolom

### `jadwal` (DBML baris 739)

| Kolom | Tipe | Aturan / catatan |
|-------|------|------------------|
| `id` | `bigint` [pk] | |
| `kode` | `varchar(50)` [unique, not null] | Partial-unique `WHERE is_deleted=false`. |
| `nama` | `varchar(150)` [not null] | |
| `jenis_jadwal` | `varchar(50)` | latihan / event / rapat. |
| `deskripsi` | `text` | |
| `lokasi_id` | `bigint` | FK → `mst_lokasi.id` (**restrict**). |
| `latitude`/`longitude` | `decimal(10,7)` | Snapshot lokasi. |
| `radius_meter` | `int` [default 100] | Menimpa radius lokasi. |
| `timezone` | `varchar(50)` [default `Asia/Jakarta`, not null] | IANA. |
| `tanggal_mulai` | `date` [not null] | |
| `tanggal_selesai` | `date` | |
| `jam_mulai` | `time` [not null] | |
| `jam_selesai` | `time` [not null] | **CHECK jam_selesai > jam_mulai**. |
| `is_berulang` | `boolean` [default false] | |
| `pola_ulang` | `pola_ulang` [default `tidak_berulang`] | enum: tidak_berulang/harian/mingguan/bulanan/kustom. |
| `hari_ulang` | `int[]` | Pola mingguan `{0..6}`, 0=Minggu. |
| `interval_ulang` | `int` [default 1] | tiap N hari/minggu/bulan. |
| `tanggal_akhir_ulang` | `date` | |
| `mode_absen` | `mode_absen` [default `masuk_saja`] | masuk_saja / masuk_pulang. |
| `wajib_absen` | `boolean` [default true] | |
| `butuh_selfie` | `boolean` [default true] | |
| `butuh_lokasi` | `boolean` [default true] | |
| `izinkan_luar_radius` | `boolean` [default true] | |
| `toleransi_telat_mnt` | `int` [default 15] | |
| `buka_absen_mnt` | `int` [default 30] | menit sebelum mulai. |
| `tutup_absen_mnt` | `int` [default 60] | menit setelah mulai. |
| `kuota` | `int` | |
| `kegiatan_id` | `bigint` | FK → `kegiatan.id`. |
| `inorga_id` | `bigint` | FK → `mst_inorga.id`. |
| `status` | `varchar(20)` [default `draft`] | **bukan enum**: draft/terbit/dibatalkan/selesai. |
| soft-delete + audit | | |

### `jadwal_sesi` (DBML baris 785)

| Kolom | Tipe | Aturan / catatan |
|-------|------|------------------|
| `id` | `bigint` [pk] | |
| `jadwal_id` | `bigint` [not null] | FK → `jadwal.id` (**restrict**). |
| `tanggal_lokal` | `date` [not null] | Tanggal di zona jadwal. |
| `mulai_utc`/`selesai_utc` | `timestamptz` [not null] | Hasil konversi jam lokal → UTC. |
| `timezone` | `varchar(50)` [not null] | Disalin dari jadwal. |
| `absen_buka_utc`/`absen_tutup_utc` | `timestamptz` [not null] | Jendela absen. |
| `status` | `status_sesi` [default `terjadwal`] | terjadwal/berlangsung/selesai/dibatalkan. |
| `alasan_batal` | `text` | |
| `jml_ditugaskan`/`jml_hadir` | `int` [default 0] | Counter. |
| `catatan` | `text` | |
| audit (created/modified) | | tanpa soft-delete. |

**Index:** `jadwal_sesi` `UNIQUE(jadwal_id, tanggal_lokal)`, `(tanggal_lokal, status)`, `(absen_buka_utc, absen_tutup_utc)`.

**Relasi:** `jadwal.lokasi_id → mst_lokasi.id` [restrict], `jadwal.kegiatan_id → kegiatan.id`, `jadwal.inorga_id → mst_inorga.id`; `jadwal_sesi.jadwal_id → jadwal.id` [restrict]; ← `jadwal_peserta.sesi_id`, `absensi.sesi_id`, `absensi_izin.sesi_id`.

---

## 3. Endpoint (api-endpoints §6)

| Method | Path | Permission | Auth | Body | Response |
|--------|------|-----------|------|------|----------|
| `GET` | `/jadwal` | `jadwal.read` | JWT | query filter/paginate | `Paginated<JadwalResp>` |
| `GET` | `/jadwal/{id}` | `jadwal.read` | JWT | — | `JadwalResp` |
| `POST` | `/jadwal` | `jadwal.create` | JWT | `CreateJadwalReq` | `JadwalResp` (201) |
| `PUT` | `/jadwal/{id}` | `jadwal.update` | JWT | `UpdateJadwalReq` | `JadwalResp` |
| `DELETE` | `/jadwal/{id}` | `jadwal.delete` | JWT | — | `null` (soft delete) |
| `POST` | `/jadwal/{id}/generate-sesi` | `jadwal.update` | JWT | `GenerateSesiReq` | `{dibuat,dilewati,sesi[]}` |
| `GET` | `/jadwal/{id}/sesi` | `jadwal.read` | JWT | — | `SesiResp[]` |
| `PATCH` | `/sesi/{id}/batalkan` | `jadwal.batal_sesi` | JWT | `{alasan_batal}` | `SesiResp` |

### `generate-sesi` (api-endpoints §6, verbatim)
```json
{
  "dari_tanggal": "2026-09-01",
  "sampai_tanggal": "2026-12-31",
  "timezone": "Asia/Jakarta",
  "mulai_lokal": "07:00",
  "selesai_lokal": "09:00",
  "absen_buka_menit_sebelum": 30,
  "absen_tutup_menit_setelah": 15
}
```
Server menghitung `tanggal_lokal` dari `pola_ulang`/`hari_ulang`/`interval_ulang`/`tanggal_akhir_ulang`, mengonversi `mulai_lokal`/`selesai_lokal` ke `mulai_utc`/`selesai_utc` **dengan `jadwal.timezone`**, menurunkan `absen_buka_utc`/`absen_tutup_utc`. Idempoten — `(jadwal_id, tanggal_lokal)` yang sudah ada dilewati (bukan duplikat). Respons `{dibuat, dilewati, sesi:[…]}`.

---

## 4. Hak akses (rancangan Bab 3.5)

| Modul | SA | Admin | Moderator | User | Guest |
|-------|----|-------|-----------|------|-------|
| `jadwal` (CRUD) | CRUD | CRUD | CRUD | R | — |
| `jadwal.assign` | ✔ | ✔ | ✔ | — | — |
| `jadwal.batal_sesi` | ✔ | ✔ | ✔ | — | — |

Permission: `jadwal.create/read/update/delete` + aksi khusus `jadwal.assign` (modul 17), `jadwal.batal_sesi`. User boleh **membaca** jadwal tapi tidak menugaskan/membatalkan.

---

## 5. Kebutuhan BACKEND (Go)

Modul `internal/modules/core/jadwal/` (mencakup entitas `jadwal` + `jadwal_sesi`).

### domain
- `Jadwal` (map `jadwal`; `HariUlang pq.Int64Array` / `[]int` untuk `int[]`; `JamMulai/JamSelesai` `string`/`time` `type:time`; enum sebagai string types). `Sesi` (map `jadwal_sesi`).

### dto
- `CreateJadwalReq`/`UpdateJadwalReq` cermin kolom (`kode required`; `jam_selesai gtfield`/cek di service; enum `pola_ulang`/`mode_absen` `oneof`; `hari_ulang` `dive,min=0,max=6`; `status oneof=draft terbit dibatalkan selesai`). `GenerateSesiReq` seperti §6. `JadwalResp`, `SesiResp`, `ListJadwalQuery` (filter `status`, `lokasi_id`, rentang tanggal).

### repository
- CRUD `jadwal` (`is_deleted=false`), `ExistsKode`. `Sesi`: `CreateBatch` (INSERT ... ON CONFLICT (jadwal_id, tanggal_lokal) DO NOTHING → idempoten), `ListByJadwal`, `Cancel(sesiID, alasan)`, `CountJadwalByLokasi`.

### service
- **Create/Update:** validasi `jam_selesai>jam_mulai`, `time.LoadLocation(timezone)`, snapshot lat/long/radius dari lokasi bila kosong. Unik `kode` (409). Audit log.
- **generate-sesi (inti):** iterasi tanggal sesuai pola (harian: tiap `interval_ulang` hari; mingguan: hari di `hari_ulang` tiap `interval_ulang` minggu; bulanan; kustom) hingga `tanggal_akhir_ulang`/`sampai_tanggal`. Untuk tiap tanggal: gabung `tanggal_lokal`+`mulai_lokal` di `loc` → `mulai_utc` (`.UTC()`), idem selesai; `absen_buka_utc = mulai_utc - buka menit`, `absen_tutup_utc = mulai_utc + tutup menit` (atau setelah selesai sesuai aturan). Batch insert idempoten; hitung `dibuat`/`dilewati`. **Jadwal tanggal tunggal → satu sesi.**
- **batalkan sesi:** set `status=dibatalkan` + `alasan_batal` (required); log.

### handler / main / router
- `SetupRoutes`: rute CRUD `RequirePermission("jadwal.<aksi>")`; `POST /:id/generate-sesi` → `jadwal.update`; `PATCH /sesi/:id/batalkan` → `jadwal.batal_sesi`. Daftar di router.

### migrations + seeders
- `jadwal` & `jadwal_sesi` **persis `.dbml`** (enum types `pola_ulang`/`mode_absen`/`status_sesi`; `int[]` untuk `hari_ulang`; CHECK `jam_selesai>jam_mulai`; `UNIQUE(jadwal_id,tanggal_lokal)`; FK restrict). **Setelah** `mst_lokasi`, `kegiatan`, `mst_inorga`. Pastikan permission `jadwal.*` + `jadwal.batal_sesi` ter-seed.

### edge cases
- `pola_ulang=tidak_berulang` → satu sesi (abaikan hari_ulang). Regenerasi tidak menduplikasi. Timezone tak valid → 422. Membatalkan sesi yang sudah lewat / sudah `selesai` → tolak. Hapus jadwal yang punya absensi → pertimbangkan guard (soft delete tetap konsisten).

---

## 6. Kebutuhan FRONTEND (Angular)

**WAJIB taste-skill.** Ini **layar kalender** — bangun lebih dulu. Dark, SLAM red, judul UPPERCASE.

### Halaman & komponen (`pages/jadwal/`)
- `jadwal-calendar.*` — tampilan **kalender** (bulan/minggu) menampilkan sesi; jam ditampilkan di **zona jadwal** (07:00 WIB), bukan browser. Gunakan komponen kalender ringan atau grid buatan sendiri sesuai token (jangan UI kit berat).
- `jadwal-form.*` — form dengan **sub-form aturan pengulangan**: `pola_ulang` mengendalikan kontrol kondisional (`hari_ulang` checkbox Sen–Min saat mingguan; `interval_ulang`; `tanggal_akhir_ulang`). Konfigurasi absensi (mode_absen, toleransi, buka/tutup menit, flag selfie/lokasi/luar-radius). Pemilih lokasi (dropdown `mst_lokasi`).
- `jadwal-sesi-list.*` — daftar sesi per jadwal + tombol **Generate Sesi** dan **Batalkan Sesi**.

### Service, i18n
- `jadwal.service.ts` (list/detail/create/update/remove/generateSesi/listSesi/batalkanSesi). i18n `JADWAL` (`TITLE`, `FORM.*`, `RECUR.*`, `SESI.*`, `GENERATE`, `CANCEL_SESI`).

### Gating
- `data:{permission:'jadwal.read'}`; tombol create/update/generate `*hasPermission="'jadwal.create'|'jadwal.update'"`; batalkan `*hasPermission="'jadwal.batal_sesi'"`.

---

## 7. ALUR form → API → DATABASE

| Field form | Payload | Kolom DB |
|------------|---------|----------|
| Pola ulang / hari / interval | `pola_ulang`,`hari_ulang`,`interval_ulang` | idem `jadwal` |
| Lokasi | `lokasi_id` | `lokasi_id` (+ snapshot lat/long/radius) |
| Aturan absen | `mode_absen`,`toleransi_telat_mnt`,`buka_absen_mnt`,`tutup_absen_mnt`,flag | idem |

| Aksi | Endpoint | Tulisan DB |
|------|----------|------------|
| Simpan jadwal | `POST /jadwal` | INSERT `jadwal` (+snapshot) + `log_aktivitas` |
| Generate sesi | `POST /jadwal/{id}/generate-sesi` | INSERT banyak `jadwal_sesi` (ON CONFLICT DO NOTHING) |
| Batalkan sesi | `PATCH /sesi/{id}/batalkan` | UPDATE `jadwal_sesi.status=dibatalkan` + log |
| Lihat kalender | `GET /jadwal` + `/jadwal/{id}/sesi` | SELECT (tampil di tz jadwal) |

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- **Lokasi (15)** (`lokasi_id`), **RBAC (Fase 1)**, konvensi timezone Fase 0 (`_ "time/tzdata"`). `kegiatan`/`mst_inorga` untuk FK opsional.

### Acceptance criteria
- [ ] Migrasi cocok `.dbml` (enum, `int[]`, CHECK jam, `UNIQUE(jadwal_id,tanggal_lokal)`).
- [ ] `generate-sesi` menghasilkan tanggal benar per pola; konversi lokal→UTC memakai `jadwal.timezone`; **idempoten** (regenerasi tidak menduplikasi).
- [ ] Jadwal tanggal tunggal → tepat satu sesi.
- [ ] `PATCH /sesi/{id}/batalkan` butuh `alasan_batal`; `jadwal.batal_sesi` diblok untuk User (403).
- [ ] User bisa `GET /jadwal` (read) tapi tidak create/assign/batal.
- [ ] Kalender menampilkan jam di zona jadwal (07:00 WIB), bukan browser tz.
- [ ] `log_aktivitas` tertulis; taste-skill diterapkan (layar kalender dibangun lebih dulu).
