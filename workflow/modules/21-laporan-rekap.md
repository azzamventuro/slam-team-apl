# Modul 21 — Laporan / Rekap Absensi (`laporan` / `absensi_rekap`)

> Sumber otoritatif: `slamteam_db.dbml` (skema) > rancangan Bab 5.4 (status kehadiran) / Bab 12.1 (Fase 7) > `_shared/api-endpoints.md` §8 > `_shared/conventions-*.md`. Jika prosa dan `.dbml` berbeda, `.dbml` menang.

---

## 1. Ringkasan & tujuan modul

**Laporan/Rekap** menghasilkan rekapitulasi kehadiran per anggota/jadwal/periode dan **ekspor Excel/PDF**. **Fase 7**, prasyarat **Absensi (18)**.

Keputusan penting: `absensi_rekap` **OPSIONAL** — awalnya **hitung langsung dari `absensi`** (GROUP BY). Materialisasi tabel `absensi_rekap` hanya bila laporan bulanan melambat.

---

## 2. Tabel & kolom

### `absensi_rekap` (DBML baris 956, OPSIONAL)

| Kolom | Tipe | Aturan / catatan |
|-------|------|------------------|
| `id` | `bigint` [pk] | |
| `anggota_id` | `bigint` [not null] | FK → `anggota.id` (**restrict**). |
| `periode` | `date` [not null] | Awal bulan periode. |
| `jml_wajib` | `int` [default 0] | |
| `jml_hadir` | `int` [default 0] | |
| `jml_telat` | `int` [default 0] | |
| `jml_pulang_cepat` | `int` [default 0] | |
| `jml_izin` | `int` [default 0] | |
| `jml_sakit` | `int` [default 0] | |
| `jml_alfa` | `int` [default 0] | |
| `total_menit_telat` | `int` [default 0] | |
| `persen_kehadiran` | `decimal(5,2)` | |
| `dihitung_pada` | `timestamptz` | |

**Index:** `UNIQUE(anggota_id, periode)`.

Sumber utama: agregasi tabel **`absensi`** (`status_kehadiran`, `menit_telat`) di-join `jadwal_peserta` (`jml_wajib`).

---

## 3. Endpoint (api-endpoints §8)

| Method | Path | Permission | Auth | Body / Query |
|--------|------|-----------|------|--------------|
| `GET` | `/absensi/rekap` | `absensi.read` | JWT | `?periode=YYYY-MM&jadwal_id=&anggota_id=` (cakupan semua/milik_sendiri) |
| `POST` | `/absensi/export` | `absensi.export` | JWT | `{format: xlsx\|pdf, periode, jadwal_id?, anggota_id?}` → file biner/URL |

`GET /absensi/rekap` data: array per anggota/periode dengan kolom hitung + `persen_kehadiran`.

---

## 4. Hak akses (rancangan Bab 3.5)

| Izin | SA | Admin | Moderator | User |
|------|----|-------|-----------|------|
| `absensi.read` (rekap) | `semua` | `semua` | `semua` | `milik_sendiri` |
| `absensi.export` | ✔ | ✔ | ✔ | — |

User melihat rekap **miliknya** saja (cakupan). Ekspor bukan untuk User.

---

## 5. Kebutuhan BACKEND (Go)

Tambahkan ke modul `absensi` (bukan tabel baru wajib): handler rekap + export.

### dto
- `RekapQuery{periode, jadwal_id?, anggota_id?}`, `RekapRow{anggota_id, nama, jml_*, persen_kehadiran}`, `ExportReq{format oneof=xlsx pdf, periode, jadwal_id?, anggota_id?}`.

### repository
- `Rekap(query, cakupan)` — GROUP BY `anggota_id` dari `absensi` (filter periode via `waktu_server_utc`/sesi, `is_deleted=false`); hitung per status; join `jadwal_peserta` untuk `jml_wajib`; `persen = hadir/wajib*100`. Cakupan `milik_sendiri` → `anggota_id=claims.AnggotaID`.
- (Opsional) `UpsertRekap(periode)` untuk materialisasi ke `absensi_rekap` (`ON CONFLICT (anggota_id,periode)`), dijalankan job.

### service
- Hitung rekap; untuk **export**, render ke Excel (mis. `excelize`) atau PDF (mis. `gofpdf`/`maroto`) — **finalisasi library saat build** (catat 99-report). Kembalikan file biner (Content-Disposition) atau URL file tergenerate. Log `log_aktivitas` (aksi export).

### handler / main / router
- `GET /absensi/rekap` → `absensi.read`; `POST /absensi/export` → `absensi.export`.

### migrations
- Bila materialisasi dipakai: `absensi_rekap` **persis `.dbml`** (`UNIQUE(anggota_id,periode)`). Bila belum, tidak perlu migrasi tabel (hitung langsung).

### edge cases
- Periode tanpa data → array kosong (bukan error). Pembagian `persen` saat `jml_wajib=0` → 0. Format tak dikenal → 422. User minta export → 403.

---

## 6. Kebutuhan FRONTEND (Angular)

**WAJIB taste-skill.** Grafik pakai warna status design-tokens (bukan warna acak). Ikuti panduan dataviz bila membuat chart.

### Komponen (`pages/laporan/`)
- `rekap.*` — tabel rekap per anggota/periode (kolom hitung + `persen_kehadiran` bar), filter periode + jadwal. Ringkasan kartu (total hadir/telat/alfa).
- Grafik kehadiran sederhana (bar/stacked) — hadir=success, telat=warning, alfa=danger, izin=info.
- Tombol **Export Excel** / **Export PDF** (`*hasPermission="'absensi.export'"`), unduh hasil.

### Service, i18n
- `laporan.service.ts` (rekap, export → blob). i18n `LAPORAN` (`REKAP`, `PERIODE`, `EXPORT_XLSX`, `EXPORT_PDF`, kolom status).

### Gating
- `data:{permission:'absensi.read'}`; export gated `absensi.export`.

---

## 7. ALUR aksi → API → DATABASE

| Aksi | Endpoint | Baca/Tulis DB |
|------|----------|---------------|
| Lihat rekap | `GET /absensi/rekap` | SELECT/GROUP BY dari `absensi` (+ `jadwal_peserta`) |
| Export | `POST /absensi/export` | SELECT lalu render file; `log_aktivitas` (export) |
| (Job) materialisasi | (cron) | UPSERT `absensi_rekap` |

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- **Absensi (18)** (sumber data) + **Penugasan (17)** (`jml_wajib`).

### Acceptance criteria
- [ ] Rekap dihitung benar dari `absensi` (per status + `persen_kehadiran`); cakupan User = miliknya.
- [ ] Export `xlsx` dan `pdf` menghasilkan file yang benar; format lain → 422.
- [ ] `absensi.export` diblok untuk User (403).
- [ ] (Bila dipakai) `absensi_rekap` cocok `.dbml` + `UNIQUE(anggota_id,periode)`.
- [ ] Grafik memakai warna status token; `log_aktivitas` export tertulis; taste-skill diterapkan.
