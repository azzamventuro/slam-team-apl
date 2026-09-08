# Modul 20 — KTA / NRA / QR (`kta`)

> Sumber otoritatif: `slamteam_db.dbml` (skema + Note tabel `kta`) > rancangan Bab 7.1–7.10 > `_shared/api-endpoints.md` §11 > `_shared/conventions-*.md` > aplikasi lama (tidak pernah). Jika prosa dan `.dbml` berbeda, `.dbml` menang.

---

## 1. Ringkasan & tujuan modul

**KTA** (Kartu Tanda Anggota) adalah kartu anggota fisik/digital yang **memiliki NRA**. Modul ke-19 (`kta`), grup **Operasional**, **Fase 6** (hanya butuh Fase 0 + Anggota). Paling padat aturan domain:

1. **NRA melekat pada KARTU, bukan orang.** `kta.no_kta` unik global; `anggota.no_induk` hanya salinan NRA aktif. Anggota `siswa_ke_dewasa` punya **dua** kartu (pelajar arsip + dewasa aktif) dengan **dua NRA berbeda**.
2. **Penomoran global via SEQUENCE.** `no_urut_kartu` berjalan lintas seluruh kartu (PostgreSQL SEQUENCE, boleh disunting manual dengan **setval guard**). NRA melar 3→4 digit setelah 999 (karena itu `varchar(20)`).
3. **QR publik hanya membuka data terbatas.** `token_publik` (bukan id) → halaman publik menampilkan hanya foto, nama, NRA, status, instansi, masa berlaku. Kartu arsip → keterangan "digantikan" tanpa membocorkan token pengganti.

---

## 2. Tabel & kolom

### `kta` (DBML baris 537)

| Kolom | Tipe | Aturan / catatan |
|-------|------|------------------|
| `id` | `bigint` [pk] | |
| `anggota_id` | `bigint` [not null] | FK → `anggota.id` (**restrict**). |
| `jenis_kta` | `jenis_kta` [not null] | pelajar / dewasa. |
| `no_kta` | `varchar(20)` [unique, not null] | **NRA** tercetak. Unik global. |
| `no_urut_kartu` | `int` [not null] | Nomor urut global (dari SEQUENCE). |
| `token_publik` | `varchar(32)` [unique, not null] | Kode acak URL QR. |
| `qr_url` | `text` | `https://slamteam.id/kta/{token_publik}`. |
| `versi_cetak` | `int` [default 1] | Naik saat cetak ulang; NRA tidak berubah. |
| `status` | `status_kta` [default `aktif`] | aktif/arsip/dicabut/kadaluarsa. |
| `berlaku_dari` | `date` [not null] | Default = tanggal terbit. |
| `berlaku_sampai` | `date` [not null] | Auto: pelajar +12 bln, dewasa +24 bln (dari `mst_pengaturan`). |
| `file_kta_id` | `bigint` | PNG kartu → `mst_file` (**PRIVAT**). |
| `alasan_cetak` | `varchar(150)` | baru/naik_jenis/hilang/rusak/perpanjangan/data_berubah. |
| `dicetak_pada`/`dicetak_oleh` | | |
| `batch_cetak_id` | `varchar(50)` | Menandai cetak massal. |
| `dicabut_pada`/`dicabut_oleh`/`alasan_dicabut` | | Pencabutan. |
| soft-delete + audit | | |

**Index:** `token_publik` unique, `no_kta` unique, `UNIQUE(anggota_id, jenis_kta, versi_cetak)`, `(anggota_id, status)`, `batch_cetak_id`, `berlaku_sampai`.

### `mst_wilayah` (DBML baris 469) — penyuplai 4 digit pertama NRA
`kode` (BPS 4 digit, **paten 3573** Kota Malang, dari `mst_pengaturan.umum.kode_wilayah_paten`), `nama`, `provinsi`, `tingkat`, `is_aktif`.

**Relasi:** `kta.anggota_id → anggota.id` [restrict]; `kta.file_kta_id → mst_file.id`; `dicetak_oleh`/`dicabut_oleh → users.id`; `anggota.wilayah_id → mst_wilayah.id`.

### Format NRA (Note tabel `kta`)
`3573 | 10 | 02 | 021 → 35731002021` (11 digit): digit 1-4 wilayah (paten 3573) · digit 5-6 **bulan lahir** · digit 7-8 **dua digit tahun lahir** (`anggota.tanggal_lahir`) · digit 9+ **nomor urut kartu**. Wilayah **bukan** dari tempat lahir (contoh: lahir Blitar tetap 3573).

---

## 3. Endpoint (api-endpoints §11)

| Method | Path | Permission | Auth | Body |
|--------|------|-----------|------|------|
| `POST` | `/kta` | `kta.create` | JWT | `{anggota_id, jenis_kta, alasan_cetak, berlaku_dari?, berlaku_sampai?}` |
| `POST` | `/kta/bulk` | `kta.create` | JWT | `{anggota_ids[], jenis_kta, alasan_cetak}` → batch |
| `GET` | `/kta/anggota/{anggotaId}` | `kta.read` | JWT | semua kartu (aktif+arsip) |
| `POST` | `/kta/{id}/cetak-ulang` | `kta.print` | JWT | bump `versi_cetak`, token baru, **no_kta tetap** |
| `PATCH` | `/kta/{id}/cabut` | `kta.cabut` | JWT | `{alasan_dicabut}` |
| `GET` | `/kta/batch/{batchId}` | `kta.read` | JWT | kartu + PNG satu batch |
| `GET` | `/public/kta/{token}` | — | **public** | halaman verifikasi (rate-limited) |

`POST /kta`: server alokasi `no_urut_kartu` dari SEQUENCE, susun `no_kta` (NRA), `token_publik` (acak 32), `qr_url`, generate PNG CR80/ID-1 300dpi (trim+bleed) → `file_kta_id`. `berlaku_sampai` auto dari `mst_pengaturan`. `POST /kta/bulk` → `{batch_cetak_id, diterbitkan, gagal, kta[]}`, tiap kartu NRA berurutan sendiri.

`GET /public/kta/{token}` (PUBLIK): hanya `{foto_url, nama, nra, status, instansi, berlaku_sampai, digantikan}`. Arsip → `{digantikan:true, pesan:"Kartu ini telah digantikan…"}` **tanpa** token pengganti. **Tidak pernah** alamat/no identitas/DOB lengkap/kontak.

---

## 4. Hak akses (rancangan Bab 3.5)

| Izin | SA | Admin | Moderator | User | Guest |
|------|----|-------|-----------|------|-------|
| `kta.create` | ✔ | ✔ | — | — | — |
| `kta.read` | ✔ | ✔ | — | — | — |
| `kta.print` (cetak-ulang) | ✔ | ✔ | — | — | — |
| `kta.cabut` | ✔ | ✔ | — | — | — |
| `GET /public/kta/{token}` | — | — | — | — | ✔ (public) |

---

## 5. Kebutuhan BACKEND (Go)

Modul `internal/modules/core/kta/`.

### domain / dto
- `Kta` map `kta` (enum string types). `CreateKtaReq{ anggota_id required; jenis_kta oneof=pelajar dewasa; alasan_cetak; berlaku_dari?; berlaku_sampai? }`, `BulkKtaReq`, `CabutReq{alasan_dicabut required}`, `KtaResp`, `PublicKtaResp`.

### repository
- `NextNoUrut()` (SEQUENCE `nextval`; **setval guard** saat nilai manual melampaui), `Create`, `FindByID`, `ListByAnggota`, `ListByBatch`, `BumpVersiCetak`, `Cabut`, `FindByToken` (untuk publik), `ArsipkanPelajar(anggotaID)`.

### service (inti)
- **Terbitkan (POST /kta):** ambil `anggota` (butuh `tanggal_lahir`), susun NRA: `wilayah_paten + bulan(2) + tahun(2) + no_urut`. Alokasi `no_urut_kartu` dari SEQUENCE. `token_publik` acak. `berlaku_sampai` = `berlaku_dari` + (pelajar 12 / dewasa 24) bulan dari `mst_pengaturan`. Generate PNG (lihat cetak) → file-service → `file_kta_id`. Bila `jenis_anggota=siswa_ke_dewasa` menerbitkan **dewasa**: set kartu pelajar lama → `status=arsip`, dan `anggota.no_induk` = NRA dewasa. Log `log_aktivitas` (aksi cetak).
- **Bulk:** loop `anggota_ids` dalam satu `batch_cetak_id`; tiap kartu NRA berurutan; kembalikan ringkasan.
- **cetak-ulang:** `versi_cetak+1`, `token_publik` baru, **`no_kta` TETAP**, regenerate PNG.
- **cabut:** `status=dicabut` + `alasan_dicabut` + `dicabut_oleh/pada`; token publik berhenti valid.
- **publik:** `FindByToken`; arsip/dicabut/kadaluarsa → keterangan sesuai; kembalikan hanya field whitelist.

### PNG generate (Bab 7.6–7.8)
- Ukuran ID-1/CR80 85.6×54mm @ 300dpi + bleed 2mm; QR 24mm; tata letak depan/belakang dari rancangan. Pustaka Go (mis. `fogleman/gg` + `golang.org/x/image`, QR `skip2/go-qrcode` atau setara) — **finalisasi library saat build** (catat di 99-report). Teks aturan di belakang dirangkai dari `mst_pengaturan` yang sama dengan penghitung masa berlaku.

### handler / main / router
- `RequirePermission` sesuai tabel; `GET /public/kta/{token}` **tanpa** JWT + **rate limit**. Daftar di router.

### migrations + seeders
- `mst_wilayah` (seed `3573`) + `kta` **persis `.dbml`** (enum `jenis_kta`/`status_kta`; unik `no_kta`/`token_publik`; `UNIQUE(anggota_id,jenis_kta,versi_cetak)`). Buat **SEQUENCE** untuk `no_urut_kartu`. **Setelah** `anggota`, `mst_file`.

### edge cases
- Nomor urut manual > posisi sequence → `setval` majukan (hindari tabrakan UNIQUE nanti). NRA melar ke 12 digit setelah 999. `siswa_ke_dewasa` wajib dua kartu. Cetak ulang tidak mengubah `no_kta`. Publik token arsip/dicabut → keterangan, bukan data. `anggota.tanggal_lahir` kosong → tolak (NRA butuh bulan/tahun).

---

## 6. Kebutuhan FRONTEND (Angular)

**WAJIB taste-skill.** Kartu selalu **palet kartu gelap** (near-black + merah) apa pun tema app (design-tokens).

### Komponen (`pages/kta/`)
- `kta-issue.*` — form terbitkan: pilih anggota + `jenis_kta` + `alasan_cetak`; opsi ubah `berlaku_*`. **Preview kartu** on-screen (frame terang, kartu gelap) menampilkan foto formal, nama, NRA, QR.
- `kta-anggota.*` — daftar kartu satu anggota (aktif/arsip) + aksi **cetak ulang** / **cabut**.
- `kta-bulk.*` — cetak massal (pilih banyak anggota → `batch_cetak_id`), unduh PNG batch (8 kartu/lembar dari pengaturan).
- `pages/public/kta-verify.*` — **halaman publik** (di luar authGuard, **tema terang**) tujuan QR: tampilkan hanya field whitelist; arsip → notis "digantikan"; **jangan** fetch berkas privat.

### Service, i18n
- `kta.service.ts` + `public-kta.service.ts`. i18n `KTA` (`ISSUE`, `REPRINT`, `REVOKE`, `NRA`, `VALID_UNTIL`, `PUBLIC.*`).

### Gating
- Terbitkan/cetak-ulang/cabut `*hasPermission="'kta.create'|'kta.print'|'kta.cabut'"`. Halaman publik tanpa gate.

---

## 7. ALUR form → API → DATABASE

| Aksi | Endpoint | Tulisan DB |
|------|----------|------------|
| Terbitkan KTA | `POST /kta` | `nextval` sequence; INSERT `kta` (no_kta, token, berlaku); generate PNG → INSERT `mst_file`; UPDATE `anggota.no_induk`; (dewasa) UPDATE kartu pelajar → arsip; `log_aktivitas` |
| Bulk | `POST /kta/bulk` | INSERT banyak `kta` + PNG dalam satu `batch_cetak_id` |
| Cetak ulang | `POST /kta/{id}/cetak-ulang` | UPDATE `versi_cetak`+token; regenerate PNG (no_kta tetap) |
| Cabut | `PATCH /kta/{id}/cabut` | UPDATE status=dicabut + alasan |
| Verifikasi publik | `GET /public/kta/{token}` | SELECT by token (field whitelist) |

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- **Fase 0** file layer (PNG privat) + `mst_pengaturan` (masa berlaku, dpi, wilayah paten). **Anggota (06)** (foto formal + tanggal lahir). RBAC.

### Acceptance criteria
- [ ] Migrasi cocok `.dbml` (enum, unik no_kta/token, `UNIQUE(anggota_id,jenis_kta,versi_cetak)`) + SEQUENCE no_urut_kartu.
- [ ] NRA disusun benar (wilayah paten + bulan/tahun lahir + urut); melar ke 12 digit setelah 999.
- [ ] `siswa_ke_dewasa` menghasilkan dua kartu (pelajar arsip + dewasa aktif) dengan NRA berbeda; `no_induk` → NRA dewasa.
- [ ] setval guard mencegah tabrakan saat nomor manual melampaui sequence.
- [ ] Cetak ulang menaikkan `versi_cetak` + token baru tetapi **no_kta tetap**.
- [ ] `GET /public/kta/{token}` hanya field whitelist; arsip → notis digantikan tanpa bocor token.
- [ ] `kta.*` diblok untuk Mod/User (403); publik tanpa auth + rate-limited.
- [ ] PNG CR80 300dpi (trim+bleed) tergenerate; `log_aktivitas` tertulis; taste-skill diterapkan (kartu palet gelap).
