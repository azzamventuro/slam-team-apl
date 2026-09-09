# 99 — Alignment Report (rencana ↔ realita)

> Catatan hidup: rekam **setiap kali realita menyimpang dari rencana** — perbaikan skema,
> koreksi prompt, keputusan yang diambil di tengah build. Diperbarui setiap langkah build
> yang menyimpang. Sumber otoritatif tetap: `slamteam_db.dbml` (skema) > `.rancangan.txt`
> (kebutuhan) > `_shared/conventions-*.md` > file ini. Jika file ini bertentangan dengan
> `.dbml`, **`.dbml` menang** — dan catat konfliknya di §3.

---

## 1. Status build saat ini (per 2026-09-09)

Kode nyata di repo, dibaca langsung dari `slam-team-api/` dan `slam-team-app/`:

| Area | Status kode | Catatan |
|------|-------------|---------|
| **API skeleton** (Fase 0) | ✅ ada | `cmd/api/main.go`, `internal/{config,database,middleware,router}`, `shared/response`, `shared/redis`, `pkg/{jwt,logger,validator,utils}`. |
| **Auth** (Fase 0) | ⚠️ sebagian | `internal/modules/core/auth` ada (login + `/me`) tapi **masih memakai tabel `users` "mainan"** — belum skema `.dbml` penuh (`anggota_id`, `role_id`, dst) dan **belum ada `sesi_login`/refresh**. Diselesaikan oleh `04-auth-session`. |
| **response envelope** | ✅ ada | `Forbidden`/`NotFound`/`Conflict`/`FromError` + `apperr.FieldsError` ditambahkan di `01-foundation`, bukan menunggu Fase 1. |
| **File layer** (Fase 0) | ✅ ada | `02-file-management` (API) selesai 2026-09-09: migrasi `0003_file_layer`, modul `internal/modules/core/file`, `imaging` + 3 varian, serve ber-izin. Komponen Angular (`FileUpload`/`SecureImage`) menyusul di prompt APP. |
| **`mst_pengaturan` + `log_aktivitas`** (Fase 0) | ✅ ada | `03-pengaturan-log` selesai 2026-09-09 (API + APP). |
| **slamctl / migrations** | ✅ ada | `cmd/slamctl` (migrate/seed/create-superadmin), `migrations/` 0001–0003, `_ "time/tzdata"` — selesai `01-foundation`. |
| **RBAC dinamis** (Fase 1) | ❌ belum | 5 tabel, `PermGuard`, seed matriks belum ada. `router.go` mendaftarkan `auth`, `pengaturan`, `file`. |
| **APP skeleton** (Fase 0) | ✅ ada | Angular 21 zoneless, `core/{guards,interceptors,services,models}`, `layouts/{vertical,topbar,sidebar,footer}`, `pages/{auth/login,dashboard}`. |
| **APP RBAC** (`PermissionService`, `*hasPermission`, sidebar dinamis) | ❌ belum | Dibangun di `05-hak-akses` (APP). |
| **Semua modul Fase 2–9** | ❌ belum | Baru pada tahap perencanaan (`workflow/`). |

**Kesimpulan:** Fase 0 tinggal `04-auth-session` (dan sisi APP `02`). Skeleton (01),
pengaturan/log (03) dan lapisan berkas (02 API) sudah berdiri; auth masih memakai
tabel `users` mainan. Fase 1–9 masih rencana.

---

## 2. Status kelengkapan `workflow/` (playbook)

Sesi ini menuntaskan playbook agar 25 work-item punya set lengkap **doc + prompt api + prompt app**.

| Bagian | Sebelum sesi ini | Ditambahkan sesi ini |
|--------|------------------|----------------------|
| `modules/*.md` (rencana) | 01,02,03,05,06,07,08,09,10,11,12,13,14 | **04, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25** |
| `prompts/*-api.md` | 08, 12 | 01,02,03,04,05,06,07,09,10,11,13,14,15,16,17,18,19,20,21,22,23,24 |
| `prompts/*-app.md` | 08 | 01,02,03,04,05,06,07,09,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25 |
| `99-alignment-report.md` | — | file ini |

> `25-pwa` sengaja **tidak** punya prompt `-api` (kerja platform front-end, tanpa endpoint baru).
> `12-medsos` sudah punya doc + prompt api sebelumnya; sesi ini hanya menambah prompt app.

Verifikasi: prompt-prompt baru dibuat oleh agen generator lalu **diaudit adversarial**
terhadap `.dbml` + `_shared/api-endpoints.md` + `_shared/permission-matrix.md`. Temuan audit
(jika ada) dan perbaikannya dicatat di §4.

---

## 3. Keputusan & konflik skema (rencana ↔ `.dbml`)

Rekam di sini setiap tempat prosa/rancangan berbeda dari `.dbml`, dan bagaimana diselesaikan
(default: `.dbml` menang).

- **`artikel.status`** — kolom status pada `artikel` bernama **`status_kegiatan` (varchar)**,
  sama seperti `kegiatan`, BUKAN `status_artikel`. Ikuti nama `.dbml` apa adanya.
- **`users` vs `user`** — tabel bernama **`users`** (bukan `user`) karena `USER` reserved di Postgres.
- **NRA milik kartu** — NRA (`kta.no_kta`) milik **kartu**, bukan orang; `anggota.no_induk`
  hanya salinan NRA aktif. `siswa_ke_dewasa` ⇒ 2 kartu, 2 NRA berbeda.
- **`absensi_rekap` opsional** — awalnya hitung rekap langsung dari `absensi`; materialisasi
  tabel `absensi_rekap` hanya bila laporan bulanan melambat.
- **`mst_file` / `mst_file_varian` — cocok 100% dengan `.dbml`.** Migrasi
  `0003_file_layer` diverifikasi terhadap skema hasil ekspor dbdiagram yang sudah
  ada di database lokal: kolom, tipe, nullability, default, indeks
  (`ix_file_reff`, `ix_file_hash`, `ix_file_kategori`, `uq_file_varian`) dan FK
  cascade `fk_filevarian_file` **identik byte-per-byte** (diff kosong). Enum
  `varian_file` / `status_proses_file` dibuat di `0001_enums`, jadi `0003` hanya
  membuat dua tabel.
- **FK audit `mst_file.created_by/modified_by/deleted_by → users.id` ditunda.**
  Tabel `users` ada di migrasi yang lebih akhir (auth/identity); sebuah FK tidak
  boleh mendahului targetnya. Keputusan sama dengan `0002_pengaturan_log`
  (`mst_pengaturan.modified_by`, `log_aktivitas.aktor_user_id`) — migrasi auth
  yang menambahkan ketiga constraint itu.
- **`file` adalah namespace izin, bukan modul menu.** `file` BUKAN salah satu dari
  19 baris `mst_modul`, tetapi seeder RBAC di `05-hak-akses` **tetap wajib**
  membuat `file.create` dan `file.delete` (SA/Admin `semua`, Mod/User
  `milik_sendiri`). Tanpa itu, gerbang `perm.Require("file.create")` yang
  disambung nanti tidak punya baris untuk dilihat.
- _(tambah entri baru di bawah saat build menemukan penyimpangan)_

---

## 4. Koreksi prompt / plan (log audit)

Diisi dari hasil audit dan temuan saat eksekusi build.

- **2026-09-08 — Metode generasi playbook.** 12 doc modul (04, 15–25) + 46 prompt
  (api+app) ditulis **langsung (inline)** dengan grounding dari `slamteam_db.dbml`
  yang dibaca penuh. (Percobaan awal via workflow 30-agen paralel gagal karena
  **session usage limit** — 0 file dihasilkan; diganti metode inline yang lebih
  hemat dan terkendali.)
- **Self-audit (grep) — bersih:** tidak ada nama tabel salah (mis. tak ada
  `mst_anggota`/`mst_kta`); `status_artikel` hanya muncul sebagai peringatan
  "jangan ganti nama" (kolom sebenarnya `artikel.status_kegiatan varchar`);
  seluruh 25 prompt app memanggil `design-taste-frontend`; seluruh 24 prompt api
  merujuk `slamteam_db.dbml`. Prompt api yang tidak memakai kata literal
  `RequirePermission` (01 fondasi, 04 auth, 24 publik) memang **tanpa gerbang izin
  by design**; sisanya mencantumkan kode `modul.aksi` per-endpoint + merujuk
  `conventions-api.md §8`.
- `25-pwa` sengaja tanpa prompt `-api` (kerja platform front-end).

### 2026-09-09 — `02-file-management` (API) dibangun

- **Pustaka gambar: `github.com/disintegration/imaging` v1.6.2** (murni Go, tanpa
  cgo → deploy tetap satu binary), plus `github.com/rwcarlsen/goexif` untuk
  mengisi `mst_file.metadata_exif` secara best-effort. `github.com/google/uuid`
  dipakai karena `uuid` harus diketahui SEBELUM insert (ia menjadi nama folder di
  disk), jadi tidak bisa mengandalkan `gen_random_uuid()` di Postgres — default
  di kolom tetap ada sebagai jaring pengaman.
- **WebP ditunda (sadar).** `imaging` tidak bisa meng-encode WebP tanpa cgo, jadi
  varian `medium`/`low` di-encode **JPEG** (q80/q70) — itulah "cadangan JPEG"
  Bab 6.2. Ditandai `// ponytail:` di `service/image.go`.
- **Sumber ber-transparansi tidak dipaksa JPEG.** Sumber PNG/GIF menghasilkan
  varian `medium`/`low` **PNG**, bukan JPEG: memipihkan alpha ke JPEG memberi
  kotak hitam di belakang setiap logo transparan. `mst_file_varian.mime_type`
  merekam apa yang benar-benar ditulis, sesuai catatan tabel `.dbml`.
- **Tipe gambar yang tidak bisa di-decode ditolak 422** (WebP, HEIC, SVG).
  Diterima: JPEG, PNG, GIF, BMP, TIFF. Lebih jujur daripada menyimpan berkas yang
  tidak bisa dibuat varian-nya.
- **`variants[]` berisi OBJEK, bukan string.** `_shared/api-endpoints.md` §10
  mencontohkan `["original","medium","low"]`; doc modul `02` §3 mencontohkan objek
  (`varian/url/lebar_px/tinggi_px/mime_type`). Dipakai bentuk **objek** (doc modul
  menang — ia spesifikasi khusus modul ini) karena komponen `SecureImage` butuh
  dimensi + mime tanpa request kedua. `url_publik` dan `kualitas` ikut disertakan.
- **Ditambah `GET /files/{uuid}`** (detail + `status_proses`), yang doc modul §3
  sebut sebagai polling opsional. Tanpa itu klien tidak punya cara melihat varian
  asinkron sudah selesai. Aturan bacanya sama persis dengan endpoint stream.
- **`GET /files/{uuid}/{varian}` memakai auth OPSIONAL** (`middleware.JWTOptional`
  baru), bukan `JWTAuth` wajib. Satu route melayani dua hal: logo publik harus
  bisa dimuat `<img src>` tanpa token, sementara selfie absensi di bentuk URL yang
  sama harus 401/403. Hanya handler yang tahu mana yang mana, jadi keputusan
  ditegakkan di sana — memenuhi kriteria terima doc §8 ("berkas publik bisa
  diakses", "privat menolak tanpa JWT").
- **Gerbang izin Fase 0 (seam untuk `05-hak-akses`).** `PermGuard` belum ada:
  `POST /files` dan `DELETE /files/{uuid}` dijaga `JWTAuth` saja (`file.create` =
  semua role berakun, jadi JWTAuth sudah efektif menjadi gerbangnya), sedangkan
  **kepemilikan** (`milik_sendiri`) ditegakkan di service lewat `created_by`.
  Baca berkas privat memakai pendekatan Fase 0: super admin / Admin menjangkau
  semua, selain itu hanya berkas sendiri. Titik sambungnya ditandai `SEAM:` di
  `service.authorizeRead` dan di `main.file.go`; saat Fase 1 selesai cukup
  membungkus route dengan `perm.Require("file.create"/"file.delete")` dan
  menerjemahkan `reff_type` → izin modul pemilik. **Tidak** dibuat PermGuard
  tandingan.
- **Nama berkas di disk selalu dari tipe hasil deteksi**, bukan dari nama unggahan
  (`{varian}.{ext}`). `mst_file.ekstensi` tetap menyimpan ekstensi yang dikirim
  pengguna untuk ditampilkan; tidak ada jalan bagi klien memilih akhiran berkas
  yang kita tulis.
- **Config baru `STORAGE_ROOT`** (default `./storage`, sudah di `.gitignore`).
  Root yang tidak bisa ditulis membuat proses **gagal saat boot**, bukan saat
  unggahan pertama — karena itu `router.Setup` kini mengembalikan `error` dan
  `cmd/api/main.go` `Fatal` bila modul gagal dibangun. `client_max_body_size`
  nginx didokumentasikan di `.env.example` bersama `FILE_MAX_UPLOAD_MB`.
- **Catatan database lokal.** `slamteam_db` di mesin dev berisi skema LENGKAP hasil
  ekspor dbdiagram yang dimuat manual, dan `schema_migrations`-nya tersangkut
  `version=1, dirty=true` — jadi `slamctl migrate up` tidak bisa dijalankan di sana
  apa adanya. Verifikasi migrasi (up → down 1 → up → down all) dan seluruh uji
  runtime dijalankan di database bersih `slamteam_migtest`. Merapikan
  `slamteam_db` (drop + migrate dari nol, atau `force 3`) adalah keputusan
  operator, bukan bagian modul ini.

---

## 5. Pertanyaan terbuka / keputusan yang tertunda

- **Library peta** (`15-lokasi`) — Leaflet + OSM (tanpa API key) vs input lat/long numerik.
  Prompt menawarkan keduanya; finalisasi saat build UI.
- **Library render KTA PNG** (`20-kta-nra-qr`) — pustaka Go untuk generate CR80 300dpi
  (mis. `fogleman/gg` / `golang.org/x/image`) + QR; finalisasi saat build.
- **Library export** (`21-laporan-rekap`) — Excel (`excelize`) + PDF (mis. `gofpdf` /
  `maroto`); finalisasi saat build.
- **Editor konten artikel** (`23-artikel`) — textarea vs rich-text ringan; hindari dependency berat.

---

## 6. Cara pakai file ini

Setiap kali sebuah langkah build **menyimpang** dari rencana (skema diperbaiki, prompt salah,
keputusan library, workaround), tambahkan satu baris di §3/§4/§5 dengan tanggal absolut.
Jangan hapus entri lama — file ini adalah jejak audit rencana ↔ realita.
