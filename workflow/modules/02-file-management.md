# 02 · Manajemen Berkas 3 Varian (`file-management`)

> Fase 0 · Fondasi · Grup: Sistem/Infrastruktur (lapisan bersama, bukan salah satu dari 19 modul menu)
> Tabel: `mst_file`, `mst_file_varian`
> Endpoint: `POST /files` · `GET /files/{uuid}/{varian}` · `DELETE /files/{uuid}`
> Permission: `file.create` = semua role · `file.delete` = SA/Admin (`semua`) + Mod/User (`milik_sendiri`)
> Prasyarat: Fondasi (skeleton Go+Angular, `mst_pengaturan`, `log_aktivitas`, konvensi timestamptz)

---

## 1. Ringkasan & tujuan modul

Lapisan unggah berkas **terpusat** yang dibangun **satu kali** di Fase 0, sebelum modul mana pun menyentuh file. Semua 18 modul lain (anggota, absensi, kta, artikel, kegiatan, instansi, unit, prestasi, inorga, medsos, dokumen, profile_club, lokasi, …) menyimpan gambar/dokumen **bukan** dengan kolom path sendiri, melainkan dengan rujukan bernama `*_file_id bigint` yang menunjuk `mst_file.id`. Dengan begitu logika unggah + pengubahan ukuran + auto-orient + varian ditulis sekali, bukan disalin ke setiap modul.

Ini **bukan** salah satu dari 19 modul yang tampil di sidebar (`mst_modul`). Ini infrastruktur Fase 0. Tetapi ia tetap punya baris `mst_permission` (`file.create`, `file.delete`) karena aksi unggah/hapus tetap dijaga RBAC seperti aksi lain (rancangan Bab 3.5: "create termasuk mengunggah berkas").

Tujuan konkret (rancangan Bab 6):
- Setiap gambar disimpan dalam **3 varian**: `original` (maks 4000 px, kualitas 90, format asal), `medium` (1200 px, kualitas 80), `low` (400 px, kualitas 70). Angka diambil dari `mst_pengaturan` grup `file`, **tidak** di-hardcode. Gambar **tidak pernah diperbesar**.
- Berkas **privat** (selfie absensi, berkas identitas, lampiran izin, dokumen internal) HANYA disajikan lewat handler Go ber-auth `GET /files/{uuid}/{varian}` yang memeriksa izin — tidak pernah langsung dari folder statis nginx.
- Berkas **publik** (logo, banner, gambar artikel/kegiatan, foto pada halaman publik KTA) boleh disajikan langsung dengan cache.
- Pustaka: `github.com/disintegration/imaging` (murni Go, tanpa cgo → deploy tetap satu binary).
- **Auto-orient EXIF wajib** (foto kamera telepon membawa EXIF Orientation; bila diabaikan seluruh selfie tampil miring 90°).
- **`hash_sha256`** untuk deduplikasi berkas identik + pemeriksaan keutuhan bukti absensi.
- **Varian dibangkitkan asinkron** (`status_proses`: `menunggu` → `selesai` / `gagal`); respons unggah tidak menunggu 3× resize.
- WAJIB lolos **daftar periksa Bab 6.6** sebelum Fase 0 dinyatakan selesai.

---

## 2. Tabel & kolom (sumber: `slamteam_db.dbml`, Kelompok 4 — File)

### `mst_file` — satu baris per berkas logis

| Kolom | Tipe | Aturan / catatan |
|---|---|---|
| `id` | bigint identity | PK |
| `uuid` | uuid | `unique, not null, default gen_random_uuid()` — **identitas publik** di URL, bukan `id` |
| `nama_asli` | varchar(255) | hanya untuk ditampilkan; TIDAK dipakai sebagai nama di disk |
| `nama_slug` | varchar(255) | dipakai di URL: `domain/{modul}/{reff_id}/{nama_slug}-{varian}.{ext}` |
| `ekstensi` | varchar(10) | ekstensi berkas asli |
| `mime_type` | varchar(100) | mime hasil deteksi isi (bukan ekstensi) |
| `ukuran_byte` | bigint | ukuran berkas asli |
| `hash_sha256` | char(64) | dedup + integritas; **indexed** |
| `lebar_px` | int | lebar gambar asli (null untuk non-gambar) |
| `tinggi_px` | int | tinggi gambar asli (null untuk non-gambar) |
| `kategori` | varchar(50) | **not null**; salah satu: `foto_profil / foto_formal / absensi / banner / logo / dokumen / flyer / kta`; **indexed** |
| `reff_type` | varchar(50) | nama tabel/entitas pemilik (polymorphic), mis. `anggota`, `absensi` |
| `reff_id` | bigint | id baris pemilik (boleh diisi belakangan) |
| `storage_driver` | varchar(20) | `default 'local'` |
| `path_dasar` | varchar(500) | pemetaan URL↔disk (folder dasar `{uuid}` di disk) |
| `is_publik` | boolean | `default false`; **false = WAJIB lewat endpoint ber-auth** |
| `status_proses` | `status_proses_file` | enum `menunggu / selesai / gagal`, `default 'menunggu'` |
| `metadata_exif` | jsonb | EXIF asli (bukti); dihapus dari varian medium/low |
| `is_deleted` / `deleted_at` / `deleted_by` | bool / timestamptz / bigint | **soft delete** |
| `created_at` / `created_by` / `modified_at` / `modified_by` | timestamptz / bigint | audit |

Indeks: `(reff_type, reff_id)`, `hash_sha256`, `kategori`, `uuid [unique]`.

Enum `status_proses_file`: `menunggu`, `selesai`, `gagal`.

**Pemisahan URL vs disk** (Note tabel):
- URL : `slamteam.id/{modul}/{reff_id}/{nama_slug}-{varian}.{ext}` (mudah dibaca)
- Disk: `storage/{kategori}/{tahun}/{bulan}/{uuid}/{varian}.{ext}` (terpartisi per tahun/bulan supaya satu folder tidak berisi puluhan ribu entri)

URL yang mudah ditebak **tidak masalah** selama berkas privat lewat handler Go ber-izin.

### `mst_file_varian` — 3 baris per gambar

| Kolom | Tipe | Aturan / catatan |
|---|---|---|
| `id` | bigint identity | PK |
| `file_id` | bigint | **not null**, FK → `mst_file.id` `[delete: cascade]` |
| `varian` | `varian_file` | enum `original / medium / low`, **not null** |
| `path` | varchar(500) | **not null**; path relatif di disk (relatif storage root) |
| `url_publik` | text | URL siap-pakai bila `is_publik` |
| `lebar_px` | int | dimensi varian |
| `tinggi_px` | int | dimensi varian |
| `ukuran_byte` | bigint | ukuran varian |
| `mime_type` | varchar(100) | mime yang benar-benar ditulis (mis. `image/jpeg`) |
| `kualitas` | int | 90 / 80 / 70 |
| `created_at` | timestamptz | `default now()` |

Indeks: `(file_id, varian) [unique]` — satu baris per (file, varian).

Enum `varian_file`: `original`, `medium`, `low`.

Spesifikasi varian (Note tabel + Bab 6.2). Nilai px dari `mst_pengaturan` grup `file`:
- `original`: sisi terpanjang maks `file.varian_original_maks_px` (=4000), kualitas 90, format asal (JPEG/PNG).
- `medium`: `file.varian_medium_px` (=1200), kualitas 80.
- `low`: `file.varian_low_px` (=400), kualitas 70.
- Gambar **tidak pernah diperbesar** (asli 800px → medium 800px, low tetap 400px).

### Pengaturan terkait (`mst_pengaturan` grup `file`, sudah di-seed di modul Pengaturan Fase 0)

```
file.varian_original_maks_px = 4000
file.varian_medium_px        = 1200
file.varian_low_px           = 400
```

Modul ini **membaca** ketiga nilai itu, tidak membuatnya.

### Relasi keluar (semua `*_file_id > mst_file.id`)
`anggota.foto_profil_file_id`, `anggota.foto_formal_file_id`, `anggota.file_identitas_file_id`, `kta.file_kta_id`, `mst_lokasi.foto_file_id`, `absensi.foto_file_id`, `absensi_izin.lampiran_file_id`, `kegiatan.gambar_highlight_file_id`, `artikel.gambar_highlight_file_id`, `unit.foto_sampul_file_id`, `prestasi.flyer_file_id`, `prestasi.foto_sampul_file_id`, `mst_instansi.logo_utama_file_id`, `mst_instansi.logo_tambahan_file_id`, `profile_club.{banner,logo_simple,logo_besar}_file_id`, `mst_inorga.{logo,banner,file_sk}_file_id`, `mst_dokumen.file_id`. Hanya `mst_file_varian.file_id` yang cascade delete; semua rujukan pemilik nullable dan tidak cascade.

---

## 3. Endpoint (rancangan Bab 9 — di bawah `/api/v1`)

| Method | Path | Permission | Auth | Fungsi |
|---|---|---|---|---|
| POST | `/files` | `file.create` (semua role berakun) | JWT | Unggah 1 berkas (multipart), buat `mst_file` + varian |
| GET | `/files/{uuid}/{varian}` | privat: auth + izin pemilik; publik: bebas | JWT untuk privat | Streaming byte 1 varian |
| DELETE | `/files/{uuid}` | `file.delete` (SA/Admin `semua`, Mod/User `milik_sendiri`) | JWT | Soft delete berkas |

### POST `/files` — request (multipart/form-data)

| Field | Tipe | Wajib | Catatan |
|---|---|---|---|
| `file` | file | ya | berkas biner; Angular **tidak** set `Content-Type` manual (browser set boundary) |
| `kategori` | string | ya | `oneof=foto_profil foto_formal absensi banner logo dokumen flyer kta` |
| `reff_type` | string | tidak | tabel pemilik (boleh menyusul; owning module set `*_file_id`) |
| `reff_id` | int | tidak | id pemilik |
| `is_publik` | bool | tidak | default dari kategori bila kosong (lihat §5) |

### POST `/files` — response `201` (envelope `data`)

```json
{
  "success": true,
  "message": "created",
  "data": {
    "uuid": "8f3c…",
    "nama_asli": "IMG_2201.jpg",
    "kategori": "foto_profil",
    "mime_type": "image/jpeg",
    "ukuran_byte": 1843201,
    "is_publik": false,
    "status_proses": "menunggu",
    "lebar_px": 3024,
    "tinggi_px": 4032,
    "variants": [
      { "varian": "original", "url": "/files/8f3c…/original", "lebar_px": 3024, "tinggi_px": 4032, "mime_type": "image/jpeg" }
    ]
  }
}
```

`status_proses = menunggu` saat balas; varian `medium`/`low` menyusul asinkron. Klien boleh polling `GET /files/{uuid}` (opsional) atau langsung minta `original` (selalu ada segera untuk gambar; untuk dokumen non-gambar hanya `original` dan `status_proses = selesai`).

### GET `/files/{uuid}/{varian}`

- `varian` ∈ `original|medium|low`. Respons: byte berkas dengan `Content-Type` sesuai `mime_type` varian, dukung `Range` + `Last-Modified` (pakai `c.File`).
- Bila `mst_file.is_publik = true`: layani langsung, `Cache-Control: public, max-age=…`.
- Bila privat: WAJIB JWT valid + pemeriksaan izin pemilik (lihat §5). `Cache-Control: private, no-store`.
- Varian belum siap (`status_proses = menunggu`) dan diminta `medium`/`low`: fallback ke `original` **atau** `409`/`404` — pilih fallback ke `original` (UX lebih baik).
- Tidak ada / soft-deleted → `404`.

### DELETE `/files/{uuid}` — response `200`

Soft delete: `is_deleted=true, deleted_at=now(), deleted_by=<aktor>`. Byte disk tidak dihapus segera (job pembersih terpisah nanti). Response `{ "success": true, "message": "OK" }`.

---

## 4. Hak akses (rancangan Bab 3.5)

| Permission (`modul.aksi`) | Super Admin (0) | Admin (10) | Moderator (20) | User (30) | Guest (99) | Cakupan |
|---|---|---|---|---|---|---|
| `file.create` | ✓ | ✓ | ✓ | ✓ | – | `semua` (semua role berakun boleh unggah) |
| `file.delete` | ✓ | ✓ | ✓ (milik sendiri) | ✓ (milik sendiri) | – | SA/Admin `semua`; Mod/User `milik_sendiri` (dibatasi `created_by = aktor`) |

`file.read` tidak berdiri sendiri sebagai baris matriks: akses baca berkas **privat** ditentukan oleh izin **pemilik** (mis. baca selfie absensi = izin `absensi.read` dengan cakupannya). Berkas **publik** bebas dibaca. `is_super` bypass semua.

> Catatan fase: middleware `PermGuard.Require(...)` baru ada di **Fase 1** (RBAC). Di Fase 0, ketiga route dijaga `middleware.JWTAuth` (create = semua yang berakun, jadi JWTAuth sudah efektif menjadi gerbang create). Kepemilikan `file.delete` (`milik_sendiri`) ditegakkan di **service** lewat `created_by`. Saat Fase 1 selesai, bungkus route dengan `perm.Require("file.create")` / `perm.Require("file.delete")` tanpa mengubah service.

---

## 5. Kebutuhan BACKEND (Go) — `internal/modules/core/file/`

### domain
- `File` (`TableName() = "mst_file"`) + embed `Audit` (created/modified/soft-delete) sesuai konvensi §6 conventions-api.
- `FileVarian` (`TableName() = "mst_file_varian"`).
- Value types Go untuk enum: `VarianFile` (`original|medium|low`), `StatusProses` (`menunggu|selesai|gagal`) sebagai `string` + konstanta.
- Konstanta kategori valid (slice) + himpunan kategori **publik-default** (`banner`, `logo`, `flyer`) vs **privat-default** (`foto_profil`? → publik? lihat aturan) — aturan `is_publik`:
  - Bila `is_publik` dikirim klien → pakai itu.
  - Bila kosong → derive dari kategori: `absensi`, `dokumen`, `foto_formal`, `foto_profil` (memuat wajah/identitas) → **privat**; `banner`, `logo`, `flyer`, `kta` (foto publik KTA) → **publik**. Owning module boleh menimpa saat set `*_file_id`.

### dto
- `UploadResponse` (bentuk §3), `VarianResponse`, `FileDetailResponse`.
- Query multipart di-bind dengan `ShouldBind` (bukan JSON): `kategori` (`binding:"required,oneof=foto_profil foto_formal absensi banner logo dokumen flyer kta"`), `reff_type`, `reff_id`, `is_publik *bool`.

### repository
- `Create(file *File) error`, `CreateVarian(v *FileVarian) error`, `FindByUUID(uuid) (*File, error)` (hanya `is_deleted=false`), `FindByHash(hash, kategori) (*File, error)` (dedup), `VariansOf(fileID) ([]FileVarian, error)`, `UpdateStatus(fileID, status)`, `SoftDelete(uuid, actorID)`.
- Terjemahkan `gorm.ErrRecordNotFound` → `ErrNotFound`.

### service — `FileService`
Tanggung jawab & aturan bisnis:
1. **Validasi ukuran & tipe dari ISI**, bukan ekstensi: baca 512 byte awal → `http.DetectContentType`. Kategori gambar (semua kecuali `dokumen`) wajib `image/*`; `dokumen` boleh `application/pdf` + gambar. Tolak lain → `ErrValidation`.
2. **Batas ukuran**: `MAX_UPLOAD_MB` dari config (default mis. 15 MB) → tolak lebih besar (413/422).
3. **Hash & dedup**: hitung `sha256` seluruh byte. Bila ada `mst_file` non-deleted dengan `hash_sha256` sama **dan** `created_by` sama **dan** `kategori` sama → kembalikan file yang sudah ada (dedup sederhana), jangan tulis ulang byte. (Ceiling: dedup hanya per-pengunggah; dedup lintas-pengguna berbagi byte disk ditunda.)
4. **Simpan original**: buat folder `storage/{kategori}/{tahun}/{bulan}/{uuid}/`, tulis `original.{ext}`. Untuk gambar: decode dengan `imaging.Decode(r, imaging.AutoOrientation(true))` (auto-orient EXIF wajib), lalu turunkan bila > `varian_original_maks_px` (`imaging.Fit`, tidak pernah upscale), encode format asal (JPEG q90 / PNG). Isi `lebar_px/tinggi_px`.
5. **Insert `mst_file`** (`status_proses = menunggu` untuk gambar; `selesai` untuk non-gambar) + baris `mst_file_varian` `original`.
6. **EXIF**: auto-orient wajib (langkah 4). `metadata_exif` best-effort — simpan map EXIF asli bila decoder mengekspos (opsional lib `github.com/rwcarlsen/goexif`); re-encode via imaging otomatis **membuang** EXIF dari medium/low (memenuhi "simpan dulu, hapus dari varian").
7. **Bangkitkan varian asinkron**: `go m.generateVariants(fileID, ...)` dengan `recover()`. Baca `medium_px`/`low_px` dari `mst_pengaturan`, `imaging.Fit` (no upscale), encode JPEG (q80/q70), tulis file, insert baris varian, set `status_proses = selesai`. Bila gagal → `status_proses = gagal` (kegagalan **terlihat**, bukan hilang). (Ceiling: goroutine per-unggah; ganti ke worker/queue bila throughput jadi masalah — `ponytail:` tandai di kode.)
8. **Delete**: soft delete + cek kepemilikan: non-SA/Admin hanya boleh hapus miliknya (`created_by == aktor`) → else `ErrForbidden`.
9. **`log_aktivitas`** untuk create (`aksi=buat`) & delete (`aksi=hapus`), `modul=file`, `reff_type/reff_id` = file uuid/id.

Edge case:
- Multipart tanpa field `file` → `422`.
- Gambar korup (imaging gagal decode) → `422`, tidak menulis baris.
- Disk tidak bisa ditulis (`os.MkdirAll`/`os.Create`/`io.Copy` error) → `Internal` (500), **cek semua error** (Bab 6.6: hindari "notifikasi berhasil tetapi berkas tidak ada").
- Gambar < target px → varian ikut ukuran asli (tidak upscale), tetap 3 baris.
- Non-gambar (`dokumen` PDF) → hanya `original`, no medium/low, `status_proses=selesai`.

**WebP (deferil sadar):** `disintegration/imaging` tidak meng-encode WebP (butuh cgo/lib sistem). Varian `medium`/`low` di-encode **JPEG** (itulah "cadangan JPEG" pada Bab 6.2). `mime_type` varian merekam apa yang benar-benar ditulis (`image/jpeg`). WebP ditambahkan bila kelak dipilih encoder pure-Go/cgo yang sepadan — struktur `mst_file_varian` tidak berubah. Tandai `// ponytail: JPEG variants (pure Go, no cgo); add WebP when encoder chosen`.

### migrations & seeders
- `migrations/000X_file_layer.up.sql` (+ `.down.sql`): `CREATE TYPE varian_file`, `CREATE TYPE status_proses_file` (bila belum di migrasi enum 0001), `CREATE TABLE mst_file`, `CREATE TABLE mst_file_varian` dengan indeks persis DBML (`(reff_type,reff_id)`, `hash_sha256`, `kategori`, `uuid unique`; `(file_id,varian) unique`, FK cascade). **Tanpa AutoMigrate.**
- Seeder: TIDAK ada seed data untuk file itu sendiri. Nilai `file.varian_*_px` di-seed oleh modul Pengaturan (Fase 0) — pastikan urutan seed itu jalan lebih dulu. Seeder permission (`file.create`, `file.delete`) di-seed oleh matriks RBAC (Fase 1); modul ini hanya bergantung padanya, tidak membuatnya.
- Config baru: `STORAGE_ROOT` (default `./storage`), `MAX_UPLOAD_MB` (default 15). Tambah ke `internal/config` + `.env.example`. Pastikan `STORAGE_ROOT` writable (Bab 6.6).

### background job
- Hanya goroutine generator varian (di atas). Tidak ada cron di modul ini.

---

## 6. Kebutuhan FRONTEND (Angular) — `slam-team-app`

Lapisan file di frontend adalah **service bersama**, bukan halaman menu penuh. (Tak ada item sidebar `file`.) Tetapi ia perlu 2 komponen reusable + 1 service yang dipakai semua form modul lain.

### Service — `shared/services/file.service.ts` (sudah dicontohkan di conventions-app §8; buat konkret)
- `upload(field, file, extra?)`: `FormData`, **tanpa** set `Content-Type` (browser set boundary). `authInterceptor` tetap menambah Bearer. Kembalikan `{ uuid, status_proses, ... }`.
- `imageUrl(uuid, varian='medium')`: `GET /files/{uuid}/{varian}` `responseType:'blob'` → `URL.createObjectURL`. **Revoke** object URL saat destroy (hindari kebocoran memori). Untuk berkas **publik** boleh pakai `url_publik` langsung di `<img src>`.
- `remove(uuid)`: `DELETE /files/{uuid}`.

### Komponen reusable — `shared/components/`
1. `FileUpload` (`file-upload.ts`): input file + drag/drop, preview lokal (`URL.createObjectURL` sebelum unggah), progress, panggil `FileService.upload`, emit `fileUuid` ke form induk (yang lalu set `*_file_id`). Props: `kategori` (wajib), `accept`, `reffType?`, `reffId?`, `isPublik?`. Validasi klien: tipe (`accept="image/*"` atau `.pdf`), ukuran maks (mirror `MAX_UPLOAD_MB`).
2. `SecureImage` (`secure-image.ts` atau pipe): terima `uuid` + `varian`, ambil blob via `FileService.imageUrl`, tampilkan; revoke URL lama saat `uuid`/`varian` berubah (`effect`) dan `ngOnDestroy`. Fallback placeholder saat memuat / gagal.

### Guard & permission-gating
- Tombol/aksi unggah muncul dengan `*hasPermission="'file.create'"` (semua role berakun punya ini, jadi praktis selalu tampil untuk yang login).
- Tombol hapus berkas: `*hasPermission="'file.delete'"`. **Ingat**: menyembunyikan tombol bukan keamanan — handler Go tetap cek kepemilikan.

### i18n (`public/i18n/{IND,ENG}.json`)
Namespace `FILE`:
```
FILE.UPLOAD, FILE.DRAG_HERE, FILE.CHOOSE, FILE.UPLOADING, FILE.PROCESSING,
FILE.REMOVE, FILE.TOO_LARGE, FILE.WRONG_TYPE, FILE.PREVIEW, FILE.FAILED
COMMON.SAVE, COMMON.CANCEL (sudah ada)
```
Jaga IND & ENG sinkron.

### Desain (WAJIB)
Setiap prompt APP mengawali dengan: **invoke design-taste-frontend** (umumkan "Using design-taste-frontend"), jalankan pre-flight/audit, map ke **Bootstrap 5**, dan **tegakkan token** di `workflow/_shared/design-tokens.md` (near-black bg, dark surfaces, SLAM red accent, teks putih, heading UPPERCASE wide-tracking). Dropzone = surface-2 dengan border putus-putus `--slam-border`, hover `--slam-primary`. Catatan restore: *"Bila source blog-fe dipulihkan, tiru layout/menu-nya untuk komponen ini; jika tidak, ikuti design tokens."*

---

## 7. ALUR form → API → DATABASE (kontrak koherensi)

### Pemetaan field (POST `/files`, multipart)

| Field form (Angular) | Key payload (multipart) | Kolom DB (`mst_file`) |
|---|---|---|
| berkas terpilih | `file` | `nama_asli`, `ekstensi`, `mime_type`, `ukuran_byte`, `hash_sha256`, `lebar_px`, `tinggi_px` (diturunkan server dari byte) |
| dropdown kategori | `kategori` | `kategori` |
| (dari konteks induk) | `reff_type` | `reff_type` |
| (dari konteks induk) | `reff_id` | `reff_id` |
| toggle / derive | `is_publik` | `is_publik` |
| — (server) | — | `uuid` (gen), `path_dasar`, `storage_driver='local'`, `status_proses='menunggu'`, `metadata_exif`, `created_by`, `created_at` |

Efek turunan: server juga menulis **1 baris** `mst_file_varian` (`varian='original'`, `path`, `lebar_px`, `tinggi_px`, `ukuran_byte`, `mime_type`, `kualitas=90`), lalu **asinkron** 2 baris lagi (`medium` q80, `low` q70) dan meng-update `mst_file.status_proses='selesai'`.

### Aksi user → endpoint → tulisan DB

| Aksi user | Endpoint | Tulisan DB |
|---|---|---|
| Pilih berkas di komponen unggah pada form modul apa pun | `POST /files` | INSERT `mst_file` (1) + `mst_file_varian` original (1); async +2 varian & UPDATE `status_proses`; INSERT `log_aktivitas` (buat) |
| Form induk disimpan (mis. simpan Anggota) | `POST/PUT /<modul>` (bukan modul ini) | Baris pemilik menyimpan `*_file_id = mst_file.id` hasil unggah tadi |
| Tampilkan gambar privat di list/detail | `GET /files/{uuid}/{varian}` | Tidak ada tulisan (baca) |
| Hapus berkas | `DELETE /files/{uuid}` | UPDATE `mst_file` set `is_deleted=true, deleted_at, deleted_by`; INSERT `log_aktivitas` (hapus). Byte disk dibiarkan (job pembersih terpisah) |

Kunci koherensi: **owning module tidak menulis path**. Ia hanya menyimpan `*_file_id`. Semua path/URL/varian hidup di `mst_file` + `mst_file_varian`.

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- Fase 0 skeleton Go+Angular sudah berdiri (router `/api/v1`, `response` envelope, `middleware.JWTAuth`, `pkg/logger`).
- `mst_pengaturan` sudah ada + di-seed dengan `file.varian_original_maks_px/medium_px/low_px`.
- `log_aktivitas` sudah ada.
- Enum `varian_file` & `status_proses_file` (di migrasi enum awal atau di migrasi file layer).
- `import _ "time/tzdata"` sudah ada (konvensi global).
- `PermGuard` (Fase 1) belum wajib — pakai `JWTAuth` dulu, bungkus `Require(...)` saat Fase 1 selesai.

### Acceptance criteria (checklist)
- [ ] `go build ./...` sukses; modul terdaftar di `internal/router/router.go`.
- [ ] Migrasi `file_layer` up/down jalan; tabel `mst_file` & `mst_file_varian` sesuai DBML (kolom, tipe, indeks, FK cascade `mst_file_varian.file_id`).
- [ ] `POST /files` (multipart, gambar JPEG) → `201`, baris `mst_file` (`status_proses` awal `menunggu`) + 1 varian `original` langsung; dalam ≤ beberapa detik `medium`+`low` muncul dan `status_proses='selesai'`.
- [ ] Foto ber-EXIF Orientation ≠ 1 tersimpan **tegak** (auto-orient) di semua varian.
- [ ] Nilai px varian mengikuti `mst_pengaturan` (ubah setting → varian baru ikut berubah); gambar kecil **tidak** diperbesar.
- [ ] `hash_sha256` terisi 64 hex; unggah byte identik oleh pengguna sama → dedup (tidak duplikat byte).
- [ ] `GET /files/{uuid}/original|medium|low` mengembalikan byte + `Content-Type` benar; mendukung `Range`.
- [ ] Berkas **privat** (`is_publik=false`) via `GET` menolak tanpa JWT (401) / tanpa izin pemilik (403); berkas **publik** bisa diakses.
- [ ] `DELETE /files/{uuid}` soft delete; User/Mod tidak bisa menghapus berkas milik orang lain (403); SA/Admin bisa semua.
- [ ] Tipe divalidasi dari isi (`http.DetectContentType`), bukan ekstensi; upload non-gambar ke kategori gambar ditolak `422`.
- [ ] Semua error `os.Create`/`io.Copy`/`MkdirAll` diperiksa (tidak ada "sukses tapi file hilang").
- [ ] `log_aktivitas` tercatat untuk create & delete.
- [ ] Daftar periksa Bab 6.6 lolos: Angular kirim FormData tanpa Content-Type manual; Go `ParseMultipartForm` sebelum `FormFile`; error I/O dicek; `client_max_body_size` didokumentasikan; storage writable; CORS preflight multipart OK.
- [ ] APP: komponen `FileUpload` mengunggah dan mengembalikan `uuid`; `SecureImage` menampilkan gambar privat lewat blob + revoke URL; tombol gated `*hasPermission`.
