# 01 — Fondasi proyek (config, waktu, envelope, logging, error, CLI)

> **Fase 0 · Grup: Fondasi · Modul infrastruktur (bukan salah satu dari 19 modul bisnis)**
> Prasyarat: — (tidak ada). Prasyarat untuk: **semua** fase dan modul berikutnya.
> Sumber otoritatif: `slamteam_db.dbml` (header konvensi baris 18–25), rancangan Bab 4.5 (zona waktu), Bab 6.6 (daftar periksa unggah), Bab 10.1–10.4 (migrasi, seed, super admin), Bab 12 (roadmap), serta `workflow/_shared/conventions-api.md`. Bila prosa dan `.dbml` berselisih, **`.dbml` menang**.

---

## 1. Ringkasan & tujuan modul

Modul ini adalah **fondasi lintas-modul** pada Fase 0. Ia **tidak membuat tabel baru** dan hanya mengekspos satu endpoint operasional (`GET /health`). Tujuannya menetapkan sekali — untuk dipakai 19 modul berikutnya — hal-hal yang, jika tidak dikunci di awal, akan melahirkan 19 variasi berbeda:

1. **Config** — konfigurasi ber-tipe dari environment (+ `.env`), dengan nilai bawaan yang aman.
2. **Waktu** — konvensi `timestamptz` (UTC di layer service) + `import _ "time/tzdata"` agar zona IANA (`Asia/Jakarta`, `Asia/Makassar`, `Asia/Jayapura`) resolve di binary minimal (scratch/alpine). Tanpa import ini, `time.LoadLocation` gagal senyap di produksi (rancangan Bab 4.5).
3. **Response envelope** — satu bentuk JSON `{success, message, data?, errors?}` untuk setiap endpoint (sudah ada di scaffold; **dilengkapi** `Forbidden` 403, `NotFound` 404, dan mapper `FromError`).
4. **Logging & error** — Zap terstruktur (sudah ada) + sentinel error bersama + satu titik log 500 (`response.Internal`).
5. **CORS** — dikonfigurasi agar mengizinkan preflight multipart (butuh oleh file layer & absensi), bukan `cors.Default()` yang terlalu longgar/kurang tepat.
6. **Graceful shutdown** — SIGINT/SIGTERM (sudah ada).
7. **Konvensi model bersama** — PK `bigint identity`, soft delete `is_deleted + deleted_at + deleted_by` (BUKAN sihir `gorm.DeletedAt`), identitas publik lewat kolom `uuid`/`token` terpisah, `*_file_id` untuk berkas. Disediakan sebagai `domain.Audit` embed + scope `is_deleted = false`.
8. **CLI `slamctl`** — binary operasional terpisah (`cmd/slamctl`), khususnya `create-superadmin` (BUKAN endpoint HTTP), plus placeholder `migrate` / `seed`.

**Yang SUDAH disediakan scaffold vs yang HARUS ditambahkan** (kontras inilah inti dokumen modul ini):

| Aspek | Scaffold saat ini | Aksi modul ini |
|---|---|---|
| `config.Load()` ber-tipe | ADA (`internal/config/config.go`, DB Postgres default `slamteam_db`) | **Lengkapi**: `App.Timezone`, `App.Version`, `CORS.Origins`, `File.MaxUploadMB`, `JWT.RefreshTTL` (dipakai modul berikut) |
| Zap logger | ADA (`pkg/logger`) | pakai apa adanya |
| Response envelope `OK/Created/BadRequest/Unprocess/Unauthorized/Internal` | ADA (`internal/shared/response`) | **Tambah** `Forbidden`, `NotFound`, `FromError` |
| Sentinel error bersama | TIDAK ADA | **Buat** `internal/shared/apperr` |
| `import _ "time/tzdata"` | **TIDAK ADA** | **Tambah** di `cmd/api/main.go` dan `cmd/slamctl/main.go` |
| CORS | `cors.Default()` (kurang tepat untuk multipart/kredensial) | **Ganti** dengan CORS berbasis config |
| Graceful shutdown | ADA (`cmd/api/main.go`) | pakai apa adanya |
| `domain.Audit` embed + soft-delete scope | TIDAK ADA | **Buat** `internal/shared/model` |
| `GET /health` | ADA (`status`, `redis`) | **Perkaya** (`time` UTC, `version`, `env`) |
| `cmd/slamctl` | **TIDAK ADA** | **Buat** binary + `create-superadmin` |
| Auth module contoh (`users(id,name,email,password)`) | ADA — **tabel mainan** | **JANGAN dipakai**; diganti pada modul auth Fase 0 berikutnya sesuai skema asli (`users` punya `anggota_id, username, email, role_id, …`) |

---

## 2. Tabel & kolom

**Tidak ada tabel baru pada modul ini.** Modul ini menetapkan konvensi kolom yang berlaku untuk SELURUH tabel (`slamteam_db.dbml` baris 18–25):

- **PK**: `bigint identity` → Go `ID int64 gorm:"primaryKey"`.
- **Waktu**: setiap kolom waktu adalah `timestamptz` (UTC). Tidak ada `datetime`, tidak ada penyimpanan waktu lokal, tidak ada offset `+07:00` (offset adalah hasil, bukan sumber — rancangan Bab 4.5). Default `now()` ditangani DB. Zona waktu disimpan sebagai nama IANA pada kolom `timezone` di tabel yang relevan (`users.timezone` default `Asia/Jakarta`, `jadwal.timezone`, dll).
- **Soft delete**: `is_deleted bool default false` + `deleted_at timestamptz` + `deleted_by bigint`. DELETE = `UPDATE ... SET is_deleted=true, deleted_at=now(), deleted_by=<aktor>`. Query baca selalu `WHERE is_deleted = false`.
- **Audit**: `created_at/created_by/modified_at/modified_by` pada tabel data.
- **Identitas publik**: jangan pernah membuka `id` berurutan di URL publik — pakai kolom `uuid`/`token` terpisah (`gen_random_uuid()`).
- **Berkas**: kolom gambar/dokumen adalah `*_file_id bigint` FK ke `mst_file.id` (nullable → `*int64`); jangan simpan path/URL pada baris pemilik.

Tabel sistem yang **dibuat di Fase 0 tetapi bukan milik modul fondasi ini** (didokumentasikan agar urutan migrasi jelas — lihat modul terpisah): `mst_pengaturan` (Bab 8.5, 20 baris seed), `log_aktivitas` (audit generik), `mst_file` + `mst_file_varian` (file layer), `sesi_login` + `users` (auth). Urutan pembuatan skema Bab 10.2 dan urutan pengisian data Bab 10.3 dikunci di sini; runner-nya (`slamctl migrate` / `slamctl seed`) lahir di modul ini.

Referensi kolom untuk `create-superadmin` (agar CLI menulis kolom yang benar; tabel dibuat oleh migrasi Fase 1/2):

- `mst_role` (`.dbml` 205–230): `id, kode, nama, level (SA=0), is_sistem, is_super (bypass semua izin), is_aktif, …`.
- `anggota` (`.dbml` 487–535): NOT NULL `instansi_id, nama_lengkap, jenis_anggota, tanggal_lahir`; `no_induk` boleh kosong sampai KTA pertama terbit.
- `users` (`.dbml` 424–467): NOT NULL `anggota_id, username, email, role_id`; `password` bcrypt; `timezone` default `Asia/Jakarta`; unik `username`/`email` (partial `WHERE is_deleted=false`).

---

## 3. Endpoint

| Method | Path | Auth | Permission | Payload | Response (`data`) |
|---|---|---|---|---|---|
| GET | `/health` | Publik (tanpa JWT) | — | — | `{ "status":"ok", "time":"<RFC3339 UTC>", "version":"<app.version>", "env":"development", "redis":true\|false }` |

`GET /health` **tidak** dibungkus envelope standar (ia adalah probe liveness/readiness untuk load balancer/uptime monitor, bukan endpoint bisnis) — tetap kembalikan JSON datar seperti scaffold, diperkaya dengan `time`, `version`, `env`. Semua endpoint bisnis lain memakai envelope `{success,message,data?,errors?}`.

**Bukan endpoint** (operasional, lewat CLI `slamctl`, tidak pernah HTTP — rancangan Bab 10.4): `create-superadmin`, `migrate up|down`, `seed`.

---

## 4. Hak akses

Modul ini **tidak memiliki baris** pada matriks `role_permission`. `GET /health` publik. Middleware RBAC (`RequirePermission`, `PermGuard`) baru lahir di **Fase 1 (hak_akses)**; modul fondasi hanya menyiapkan `middleware.JWTAuth` (sudah ada) dan struktur envelope `Forbidden` yang akan dipakai guard tersebut. `create-superadmin` sengaja berada di luar sistem izin karena ia membuat pemegang izin pertama (masalah ayam-dan-telur, Bab 10.4).

---

## 5. Kebutuhan BACKEND (Go)

Semua di bawah `slam-team-api/`. Ikuti `_shared/conventions-api.md`.

### 5.1 Waktu / tzdata (WAJIB, satu baris tapi krusial)
- Tambah `import _ "time/tzdata"` di `cmd/api/main.go` dan `cmd/slamctl/main.go`. Tanpa ini deploy ke image minimal gagal `LoadLocation` tanpa pesan jelas.
- Sediakan helper kecil `pkg/utils` (atau `internal/shared/timeutil`) `InZone(t time.Time, iana string) time.Time` yang membungkus `time.LoadLocation` + fallback UTC bila zona tak dikenal. Layer service selalu UTC; konversi hanya di presentasi.

### 5.2 Config (`internal/config/config.go`) — perluas struct yang sudah ada
Tambahkan tanpa merombak: `App.Timezone` (default `Asia/Jakarta`), `App.Version` (default `dev`, boleh dari `-ldflags`), `CORS.Origins []string` (dari `CORS_ORIGINS`, koma-separated, default `http://localhost:4200`), `CORS.AllowCredentials bool`, `File.MaxUploadMB int` (default 15, dipakai file layer), `JWT.RefreshTTL time.Duration` (default 720h, dipakai modul auth). Pertahankan helper `getenv`/`getenvInt`; tambah `getenvBool`, `getenvList`.

### 5.3 Response envelope (`internal/shared/response/response.go`) — lengkapi, jangan fork
Tambah ke package yang sama:
```go
func Forbidden(c *gin.Context, msg string) // 403
func NotFound(c *gin.Context, msg string)  // 404
func FromError(c *gin.Context, err error)  // map sentinel → status, else Internal
```
`FromError` men-switch sentinel `apperr` → status; default `Internal(c, err)` (satu-satunya titik log 500). Envelope tetap `{success,message,data?,errors?}`.

### 5.4 Sentinel error (`internal/shared/apperr/apperr.go`) — baru
```go
var (
    ErrNotFound   = errors.New("data tidak ditemukan")
    ErrForbidden  = errors.New("tidak diizinkan")
    ErrConflict   = errors.New("data bentrok")
    ErrValidation = errors.New("validasi gagal")
    ErrUnauthorized = errors.New("tidak terautentikasi")
)
```
Repository menerjemahkan `gorm.ErrRecordNotFound` → `ErrNotFound`; error DB lain menggelembung apa adanya (handler → `Internal`, yang log). Jangan pernah kembalikan teks error GORM/PG mentah ke klien (bocor skema).

### 5.5 Model bersama (`internal/shared/model/audit.go`) — baru
`Audit` embed (created/modified/deleted) + scope soft-delete `NotDeleted(db *gorm.DB) *gorm.DB` = `db.Where("is_deleted = false")`. Semua domain data meng-embed `Audit`; semua list/read memanggil scope ini. (Definisi persis: lihat `_shared/conventions-api.md` §6.)

### 5.6 CORS (`internal/middleware/setup.go`) — ganti `cors.Default()`
Konfigurasi `cors.New` dari `cfg.CORS`: `AllowOrigins` = `cfg.CORS.Origins`; `AllowMethods` = GET/POST/PUT/PATCH/DELETE/OPTIONS; `AllowHeaders` = Origin, Content-Type, Authorization, Accept; `AllowCredentials` = `cfg.CORS.AllowCredentials`. Penting: **preflight OPTIONS untuk multipart** harus lolos (daftar periksa unggah Bab 6.6) — jangan whitelist `Content-Type` yang menutup `multipart/form-data`; biarkan browser menetapkan boundary. Pertahankan `gin.Recovery()`.

### 5.7 Health (`internal/router/router.go`)
Perkaya handler `GET /health` yang sudah ada: `{"status":"ok","time": time.Now().UTC().Format(time.RFC3339), "version": cfg.App.Version, "env": cfg.App.Env, "redis": rdb != nil}`. Tetap publik (di luar `/api/v1`).

### 5.8 CLI `slamctl` (`cmd/slamctl/main.go`) — baru
Binary kedua (selain `cmd/api`). Sub-command sederhana (pakai `os.Args` / `flag`, tidak perlu framework CLI berat):
- `slamctl create-superadmin` — flag `--nama`, `--username`, `--email`, `--password` (atau prompt interaktif). Alur:
  1. `config.Load()` → `database.New(cfg.DB)`.
  2. **Guard**: hitung user yang tertaut ke `mst_role` dengan `is_super=true` dan `is_deleted=false`. Bila > 0 → **REFUSE** (`log.Fatal("super admin sudah ada; perintah dibatalkan")`). Ini mencegah backdoor pasca-launch (Bab 10.4).
  3. Pastikan peran Super Admin ada (`mst_role` `is_super=true`, `level=0`, `kode='super_admin'`); jika seed belum jalan, instruksikan `slamctl seed` dulu.
  4. Butuh `instansi_id` valid (NOT NULL di `anggota`) — ambil instansi pertama atau flag `--instansi-id`; jika tak ada, instruksikan seed `mst_instansi` dulu.
  5. Dalam satu transaksi: INSERT `anggota` (nama_lengkap, instansi_id, jenis_anggota=`dewasa`, tanggal_lahir dari flag/placeholder yang wajib), lalu INSERT `users` (anggota_id, username, email, `password`=bcrypt via `pkg/utils.HashPassword`, role_id=Super Admin). Commit.
  6. Tulis `log_aktivitas` (`modul='sistem', aksi='buat', ringkasan='Membuat super admin pertama'`) bila tabel sudah ada.
- `slamctl migrate up|down [n]` dan `slamctl seed [name]` — placeholder yang memanggil runner migrasi (mis. golang-migrate) & seeder idempoten. Detail migrasi/seed penuh dikerjakan di modul sistem terkait; di modul fondasi cukup rangka perintah + pesan bila belum tersedia.

**Edge cases CLI**: DB tak terjangkau → fatal jelas; tabel `mst_role`/`users` belum ada (migrasi belum jalan) → pesan "jalankan `slamctl migrate up` dulu"; `username`/`email` sudah dipakai → tolak; password kosong/lemah → tolak. **`create-superadmin` TIDAK BOLEH menjadi endpoint HTTP** (Bab 10.4) — hanya CLI.

**Tidak ada `AutoMigrate`.** Skema = file SQL bernomor (Bab 10.1). Seeder idempoten (`INSERT ... ON CONFLICT DO NOTHING/UPDATE`).

**Background jobs**: tidak ada pada modul ini.

---

## 6. Kebutuhan FRONTEND (Angular)

Fondasi FE = kerangka lintas-fitur di `slam-team-app/` (halaman fitur lahir di modul berikutnya). **Wajib**: mulai dengan Taste Skill (announce **"Using design-taste-frontend"**), jalankan pre-flight/audit, petakan ke Bootstrap 5 (sudah terpasang), dan TEGAKKAN token di `workflow/_shared/design-tokens.md` (dark default: near-black bg, dark-gray surface, SLAM red accent, teks putih, heading UPPERCASE berjarak lebar). **Catatan blog-fe**: sumber `blog-fe` sedang kosong — jika kelak dipulihkan, tiru tata letak/menunya untuk layar terkait; jika tidak, ikuti design tokens + taste-skill.

Yang dibangun / dipastikan di fondasi FE:

- **Model envelope** (`core/models/api-response.ts`): `interface ApiResponse<T> { success: boolean; message: string; data?: T; errors?: Record<string,string> | unknown; }` dan `interface Paginated<T> { items: T[]; page: number; per_page: number; total: number; last_page: number; }` — persis mengikuti envelope backend.
- **Core ApiService** (`core/services/api.service.ts`): pembungkus tipis `HttpClient` ke `${environment.apiURL}`, membuka `.data` dari `ApiResponse<T>`. **Untuk upload**: kirim `FormData` **tanpa** menetapkan header `Content-Type` manual (browser menaruh boundary — daftar periksa unggah Bab 6.6).
- **Interceptors** (sudah discaffold, pastikan benar): `authInterceptor` menambah `Authorization: Bearer <slam_token>`; `errorInterceptor` menangkap 401 → logout + redirect `/auth/login`, dan 403 → pesan "tidak diizinkan" (dipakai RBAC Fase 1).
- **Environment**: `environment.local.ts` `apiURL: 'http://localhost:8080/api/v1'` (sudah ada); pastikan `serve:local` memakainya.
- **i18n bootstrap**: ngx-translate IND/ENG dari `public/i18n/{IND,ENG}.json`; sediakan kunci fondasi (lihat §7). Default IND.
- **Timezone display util** (`shared/util/tz.ts` atau pipe): tampilkan waktu menurut zona jadwal (mis. `07:00 WIB`), **bukan** zona peramban (rancangan Bab 4.5). Fondasi menyiapkan helper; modul jadwal/absensi memakainya.
- **Design tokens applied**: buat `src/styles` yang mengeset variabel `--slam-*` dan override variabel Bootstrap sesuai `design-tokens.md` (dark default). Ini memastikan 19 modul memakai satu gaya.
- **Health/connectivity check** (opsional, ringan): pada bootstrap, panggil `GET /health` untuk indikator "API tersambung" di footer/topbar dev. Ponytail: cukup satu call, bukan poller.

Tidak ada form bisnis di fondasi FE. Guard `authGuard` + `*hasPermission` directive + sidebar dinamis lahir di Fase 1.

---

## 7. ALUR form → API → DATABASE (kontrak koherensi)

Modul fondasi hampir tak punya form UI. Dua alur konkret yang dikunci:

### 7.1 `GET /health` (probe)
| Aksi | Endpoint | Tulis DB | Baca DB |
|---|---|---|---|
| Load balancer / FE cek koneksi | `GET /health` | — (tidak ada) | — (opsional ping DB/redis; scaffold cukup cek `rdb != nil`) |

Respons: `{status, time (UTC RFC3339), version, env, redis}`. Tidak ber-envelope.

### 7.2 `slamctl create-superadmin` (CLI → DB) — satu-satunya alur tulis modul ini
| Argumen CLI | Kolom DB yang ditulis | Tabel |
|---|---|---|
| `--nama` | `nama_lengkap` | `anggota` |
| `--instansi-id` (atau instansi pertama) | `instansi_id` (NOT NULL) | `anggota` |
| (tetap) `dewasa` | `jenis_anggota` | `anggota` |
| `--tanggal-lahir` (wajib) | `tanggal_lahir` (NOT NULL) | `anggota` |
| — (kaitan) | `users.anggota_id` = id anggota baru | `users` |
| `--username` | `username` (unik) | `users` |
| `--email` | `email` (unik) | `users` |
| `--password` → bcrypt | `password` | `users` |
| — (peran Super Admin, `is_super=true`) | `role_id` | `users` |
| default | `timezone='Asia/Jakarta'`, `is_aktif=true` | `users` |
| sistem | baris audit `modul='sistem', aksi='buat'` | `log_aktivitas` (bila ada) |

Urutan aksi → tulisan DB: **(1)** guard hitung user is_super → bila ada, **abort tanpa menulis apa pun**; **(2)** BEGIN tx; **(3)** INSERT `anggota`; **(4)** INSERT `users` (role Super Admin, password ter-hash); **(5)** INSERT `log_aktivitas`; **(6)** COMMIT. Bila langkah mana pun gagal → ROLLBACK (tidak ada anggota/user setengah jadi).

---

## 8. Dependencies / prasyarat & Acceptance criteria

**Prasyarat**: tidak ada (ini modul pertama). **Menyediakan untuk semua modul**: envelope lengkap, sentinel error, `domain.Audit` + scope soft-delete, tzdata, CORS multipart-ready, config ter-tipe, runner `slamctl`. Catatan urutan: `create-superadmin` bisa **berhasil** hanya setelah migrasi RBAC + identitas (Fase 1/2) dan seed peran/instansi jalan; rangka perintahnya tetap dikirim di Fase 0 (Bab 12), dan ia memberi pesan jelas bila prasyarat tabel/seed belum ada.

**Acceptance criteria (checklist):**

- [ ] `cmd/api/main.go` dan `cmd/slamctl/main.go` meng-import `_ "time/tzdata"`; `time.LoadLocation("Asia/Jakarta")` sukses pada binary yang dibuild.
- [ ] `make build` / `go build ./...` hijau; `go vet ./...` bersih.
- [ ] `GET /health` mengembalikan `status=ok`, `time` UTC RFC3339, `version`, `env`, `redis`.
- [ ] `internal/shared/response` punya `Forbidden` (403), `NotFound` (404), `FromError`; envelope tetap `{success,message,data?,errors?}`.
- [ ] `internal/shared/apperr` menyediakan sentinel; `FromError` memetakannya (mis. `ErrNotFound`→404, `ErrForbidden`→403, `ErrConflict`→409, `ErrValidation`→422, else 500).
- [ ] `internal/shared/model` punya `Audit` embed + scope `NotDeleted`.
- [ ] CORS berbasis `cfg.CORS`, mengizinkan preflight `OPTIONS` untuk `multipart/form-data` dari origin FE.
- [ ] `config.Load()` mengisi `App.Timezone`, `App.Version`, `CORS.*`, `File.MaxUploadMB`, `JWT.RefreshTTL` dari env dengan default aman.
- [ ] `slamctl create-superadmin` **menolak** berjalan bila sudah ada user is_super; sukses membuat 1 `anggota` + 1 `users` (role Super Admin, password bcrypt) dalam satu transaksi bila belum ada; **tidak** tersedia sebagai endpoint HTTP.
- [ ] `slamctl migrate` / `seed` minimal ada sebagai perintah (boleh delegasi ke runner) dan memberi pesan jelas bila belum diimplementasi.
- [ ] Modul auth mainan (`users(id,name,email,password)`) ditandai untuk diganti; tidak dipakai modul lain.
- [ ] (FE) `ApiResponse<T>`/`Paginated<T>` cocok dengan envelope; `ApiService` membuka `.data`; upload memakai `FormData` tanpa `Content-Type` manual.
- [ ] (FE) Token `--slam-*` + override Bootstrap terpasang (dark default); i18n IND/ENG bootstrap jalan; taste-skill pre-flight dijalankan.
