# 06 — Master Data Anggota

> Modul ke-1 dari 19 (`kode = anggota`) · Grup **Master Data** · **Fase 2** (Anggota + master data).
> Sumber otoritatif: `slamteam_db.dbml` (tabel `anggota`) + rancangan Bab 2, 3.5, 6, 7. Bila prosa dan `.dbml` berbeda, **`.dbml` menang**.

---

## 1. Ringkasan & tujuan modul

Modul **Anggota** adalah master data inti keanggotaan SLAM Team: profil lengkap tiap airsofter (identitas, foto, instansi/sekolah, jenis & status keanggotaan). Ia menjadi hulu dari hampir semua modul operasional — `kta` (NRA melekat pada kartu milik anggota), `jadwal_peserta` (penugasan memakai `anggota_id`), `absensi`, `unit`, `prestasi`, `medsos`, dan `users` (setiap akun user menunjuk satu anggota).

Tiga hal yang membuat modul ini bukan CRUD biasa:

1. **Tiga jenis foto dengan sifat berbeda.** `foto_profil` (foto bebas untuk tampilan aplikasi), `foto_formal` (pas foto latar merah — KHUSUS dicetak di KTA), dan `file_identitas` (**PRIVAT** — KTP/kartu pelajar, hanya lewat endpoint ber-auth).
2. **`no_induk` (NRA) BUKAN input form.** NRA melekat pada **kartu** (`kta.no_kta`), bukan pada anggota. Kolom `anggota.no_induk` hanya **salinan** NRA kartu yang sedang aktif, ditulis oleh modul KTA (Fase 6). Di modul ini `no_induk` **read-only** dan kosong sampai KTA pertama terbit.
3. **`tanggal_lahir` memasok NRA.** Digit 5–8 NRA berasal dari bulan & tahun lahir (Bab 7.1). Karena itu `tanggal_lahir` **wajib** di modul ini walau NRA baru dihitung di Fase 6.

**Bukan tugas modul ini:** membuat akun `users` (tidak setiap anggota punya akun — dibuat terpisah di modul user-management), menerbitkan NRA/KTA (Fase 6), atau registrasi mandiri (tidak ada sama sekali — baris `anggota` hanya dibuat lewat panel admin oleh pemegang `anggota.create`).

---

## 2. Tabel & kolom

Sumber: **`Table anggota`** di `slamteam_db.dbml` (baris 487–535). Satu tabel; seluruh referensi gambar berupa `*_file_id` ke `mst_file.id`.

| Kolom | Tipe | Aturan / catatan |
|---|---|---|
| `id` | bigint PK identity | Kunci internal; **tidak** dipakai di URL publik. |
| `instansi_id` | bigint **NOT NULL** | FK → `mst_instansi.id` (`delete: restrict`). Wajib. |
| `wilayah_id` | bigint (nullable) | FK → `mst_wilayah.id`. Opsional. **Bukan** sumber kode wilayah NRA (itu dari `pengaturan umum.kode_wilayah_paten`). |
| `no_induk` | varchar(20) **UNIQUE, nullable** | NRA yang **sedang aktif** — salinan dari baris `kta` berstatus `aktif`. **Kosong** sampai KTA pertama terbit. **Read-only di modul ini** (ditulis modul KTA). Index unik partial (abaikan `is_deleted`). |
| `nama_lengkap` | varchar(150) **NOT NULL** | Dicetak di KTA. Uji nilai terpanjang (paling sering meluap di kartu). |
| `nama_panggilan` | varchar(50) | Opsional. |
| `foto_profil_file_id` | bigint (nullable) | FK → `mst_file.id`. Foto bebas untuk profil aplikasi. Publik-lunak. |
| `foto_formal_file_id` | bigint (nullable) | FK → `mst_file.id`. **Pas foto formal latar merah — khusus dicetak di KTA.** Dibutuhkan sebelum cetak KTA (Fase 6), tidak wajib saat buat anggota. |
| `jenis_anggota` | enum `jenis_anggota` **NOT NULL** | `siswa` / `dewasa` / `siswa_ke_dewasa`. Yang terakhir WAJIB punya 2 KTA (pelajar arsip + dewasa aktif) — konsekuensi di Fase 6. |
| `jenis_kelamin` | int (nullable) | Kode int (DBML tidak mengunci enum). Konvensi tampilan: `1 = Laki-laki`, `2 = Perempuan`. |
| `jenis_identitas` | int (nullable) | Kode int jenis dokumen identitas. Konvensi: `1 = KTP`, `2 = Kartu Pelajar`, `3 = Kartu Keluarga`, `4 = SIM`, `5 = Paspor`. |
| `no_identitas` | varchar(50) | Nomor dokumen identitas. |
| `file_identitas_file_id` | bigint (nullable) | FK → `mst_file.id`. **PRIVAT** — scan KTP/kartu pelajar. WAJIB lewat `GET /files/{uuid}/{varian}` ber-auth; jangan pernah disajikan statis. |
| `pekerjaan` | varchar(100) | Opsional. |
| `alamat` | text | **Dicetak di KTA.** |
| `kode_pos` | varchar(10) | Opsional. |
| `tempat_lahir` | varchar(100) | **Dicetak di KTA.** TIDAK dipakai untuk kode wilayah NRA. |
| `tanggal_lahir` | date **NOT NULL** | **Memasok digit 5–8 NRA** (bulan+tahun). Wajib. |
| `tanggal_bergabung` | date (nullable) | Opsional. |
| `status_anggota` | enum `status_anggota` default `aktif` | `aktif` / `non_aktif`. **Diset MANUAL.** Masa berlaku KTA habis **TIDAK** mengubahnya. |
| `is_deleted` / `deleted_at` / `deleted_by` | bool / timestamptz / bigint | **Soft delete.** DELETE = set flag, bukan hard delete. |
| `created_at` / `created_by` / `modified_at` / `modified_by` | timestamptz / bigint | Audit. `created_at` default `now()`. |

**Enum terkait (dari `.dbml`):**
- `jenis_anggota { siswa, dewasa, siswa_ke_dewasa }`
- `status_anggota { aktif, non_aktif }`

**Indexes (dari `.dbml`):** `no_induk [unique]`, `instansi_id`, `jenis_anggota`, `status_anggota`. → filter & pencarian modul ini bersandar tepat pada indeks ini.

**Relasi FK (dari `.dbml`):**
- `anggota.instansi_id > mst_instansi.id` (restrict)
- `anggota.wilayah_id > mst_wilayah.id`
- `anggota.foto_profil_file_id > mst_file.id`
- `anggota.foto_formal_file_id > mst_file.id`
- `anggota.file_identitas_file_id > mst_file.id`
- Dirujuk balik oleh: `users.anggota_id` (restrict), `kta.anggota_id` (restrict), `jadwal_peserta.anggota_id` (restrict), `absensi.anggota_id`, `unit.anggota_id`, `prestasi.anggota_id`, `mst_medsos.anggota_id`, `absensi_izin.anggota_id`, `absensi_rekap.anggota_id`.

---

## 3. Endpoint

Standar REST di bawah `/api/v1`, seluruhnya di-guard `middleware.JWTAuth` + `RequirePermission("anggota.<aksi>")`. Envelope: `{ success, message, data?, errors? }`.

| Method | Path | Permission | Auth | Ringkas |
|---|---|---|---|---|
| GET | `/anggota` | `anggota.read` | JWT | List + filter + paginate + search |
| GET | `/anggota/{id}` | `anggota.read` | JWT | Detail satu anggota |
| POST | `/anggota` | `anggota.create` | JWT | Buat anggota baru |
| PUT | `/anggota/{id}` | `anggota.update` | JWT | Ubah anggota |
| DELETE | `/anggota/{id}` | `anggota.delete` | JWT | **Soft delete** |

### 3.1 `GET /anggota` — list

Query params:
- `page` (default 1, min 1), `per_page` (default 20, min 1, max 100)
- `q` — pencarian pada `no_induk` **atau** `nama_lengkap` (ILIKE)
- `instansi_id` (bigint) — filter `instansi_id`
- `jenis` — filter `jenis_anggota` (`siswa|dewasa|siswa_ke_dewasa`)
- `status` — filter `status_anggota` (`aktif|non_aktif`)
- `sort` — whitelist: `nama_lengkap`, `-nama_lengkap`, `created_at`, `-created_at`, `tanggal_bergabung`, `-tanggal_bergabung` (default `-created_at`)

Response `data`:
```json
{
  "items": [
    {
      "id": 12,
      "no_induk": "35731002021",
      "nama_lengkap": "Achmad Maulana Azzam",
      "nama_panggilan": "Azzam",
      "jenis_anggota": "dewasa",
      "status_anggota": "aktif",
      "instansi_id": 3,
      "instansi_nama": "SMK Negeri 1 Malang",
      "foto_profil": { "uuid": "9f1c...", "is_publik": false },
      "created_at": "2026-09-01T02:11:00Z"
    }
  ],
  "page": 1, "per_page": 20, "total": 137, "last_page": 7
}
```

### 3.2 `GET /anggota/{id}` — detail

Response `data` = objek anggota penuh (semua kolom non-audit + `*_file` diresolusi jadi `{uuid,is_publik}` + `instansi_nama`, `wilayah_nama`). `no_induk` disertakan (read-only).

### 3.3 `POST /anggota` — create

Request body (JSON — file diunggah lebih dulu ke `POST /files`, lalu kirim `*_file_id`):
```json
{
  "instansi_id": 3,
  "wilayah_id": 1,
  "nama_lengkap": "Achmad Maulana Azzam",
  "nama_panggilan": "Azzam",
  "jenis_anggota": "dewasa",
  "jenis_kelamin": 1,
  "jenis_identitas": 2,
  "no_identitas": "0035731002",
  "pekerjaan": "Pelajar",
  "alamat": "Jl. Contoh No. 1, Malang",
  "kode_pos": "65125",
  "tempat_lahir": "Malang",
  "tanggal_lahir": "2002-10-16",
  "tanggal_bergabung": "2026-01-10",
  "status_anggota": "aktif",
  "foto_profil_file_id": 91,
  "foto_formal_file_id": 92,
  "file_identitas_file_id": 93
}
```
**`no_induk` TIDAK diterima** (server mengabaikannya bila dikirim). Response `201` = objek anggota yang baru dibuat (dengan `id`, `no_induk = null`).

### 3.4 `PUT /anggota/{id}` — update

Body sama dengan create (semua field kecuali `no_induk`). `no_induk` tetap tak bisa disunting lewat sini. Response `200` = objek terbaru.

### 3.5 `DELETE /anggota/{id}` — soft delete

Set `is_deleted=true, deleted_at=now(), deleted_by=<aktor>`. Response `200` `{ success:true, message:"..." }`.

---

## 4. Hak akses

Baris `anggota` dari matriks (rancangan Bab 3.5 / `permission-matrix.md`). Seluruh cakupan = **`semua`** (modul ini tidak dibatasi per-instansi/pemilik).

| Role | Level | anggota |
|---|---:|---|
| Super Admin | 0 | **CRUD** (via `is_super` bypass) |
| Admin | 10 | **CRUD** |
| Moderator | 20 | **CRUD** |
| User | 30 | **R** (read only) |
| Guest | 99 | **—** (tidak ada akses) |

Permission yang di-seed (Fase 1, tabel `mst_permission`): `anggota.create`, `anggota.read`, `anggota.update`, `anggota.delete`. Cakupan default `semua` di `role_permission`. `is_super` (Super Admin) melewati seluruh pengecekan tanpa membaca `role_permission`.

> Menyembunyikan tombol di UI **bukan** keamanan: setiap izin di atas WAJIB dijaga handler Go (`RequirePermission`). Endpoint publik `anggota` **tidak ada** — data anggota tidak boleh bocor ke Guest.

---

## 5. Kebutuhan BACKEND (Go)

Modul di `internal/modules/core/anggota/` mengikuti pola `domain/dto/repository/service/handler/main.anggota.go`.

### 5.1 domain
- `Anggota` struct GORM → tabel `anggota`. Kolom sesuai §2. Embed `Audit` (created/modified/soft-delete) dari konvensi shared.
- Value types string untuk enum: `JenisAnggota` (`siswa|dewasa|siswa_ke_dewasa`) dan `StatusAnggota` (`aktif|non_aktif`). `TableName() => "anggota"`.
- Field `*_file_id` = `*int64` (nullable).

### 5.2 dto
- `CreateAnggotaReq` / `UpdateAnggotaReq` (identik; boleh 1 struct) dengan `binding`:
  - `instansi_id` `required`
  - `nama_lengkap` `required,max=150`
  - `jenis_anggota` `required,oneof=siswa dewasa siswa_ke_dewasa`
  - `tanggal_lahir` `required` (date `2006-01-02`)
  - `status_anggota` `omitempty,oneof=aktif non_aktif` (default `aktif`)
  - `jenis_kelamin` `omitempty,oneof=1 2`
  - `nama_panggilan` `omitempty,max=50`, `no_identitas` `omitempty,max=50`, `pekerjaan` `omitempty,max=100`, `kode_pos` `omitempty,max=10`, `tempat_lahir` `omitempty,max=100`
  - `*_file_id`, `wilayah_id`, `jenis_identitas`, `tanggal_bergabung`, `alamat` → `omitempty`
  - **`no_induk` TIDAK ada di DTO request.**
- `AnggotaListItem`, `AnggotaResponse` (detail), `ListQuery` (page/per_page/q/instansi_id/jenis/status/sort), `Paginated[T]`.

### 5.3 repository
- `List(ctx, filter) ([]Anggota, total, err)` — `Where("is_deleted = false")` + filter opsional (`instansi_id`, `jenis_anggota`, `status_anggota`), `q` → `ILIKE` pada `no_induk` OR `nama_lengkap`. `Count` lalu `Find` berpaginasi pada query yang sama. Sort dari whitelist (jangan interpolasi mentah).
- `FindByID(ctx, id)` → `Where("is_deleted = false")`; `gorm.ErrRecordNotFound` → `ErrNotFound`.
- `Create`, `Update`, `SoftDelete(id, actor)`.
- Preload/join ringan untuk `instansi_nama`, `wilayah_nama` (JOIN `mst_instansi`/`mst_wilayah`), atau resolve di service.

### 5.4 service
Aturan bisnis:
1. **`no_induk` tidak pernah ditulis dari sini.** Create → `no_induk = NULL`. Update → jangan sentuh `no_induk`. (Diisi oleh modul KTA saat kartu aktif terbit — Fase 6.)
2. **Validasi FK `instansi_id`** ada & tidak terhapus (`mst_instansi`), else `ErrValidation`/`ErrNotFound`. `wilayah_id` bila diisi, cek ada.
3. **`tanggal_lahir` wajib** (memasok NRA nanti) — sudah di binding, tegaskan.
4. **`status_anggota` manual**; default `aktif` bila kosong. Jangan ada logika otomatis yang mengubahnya berdasarkan KTA.
5. **File refs**: bila `*_file_id` dikirim, opsional verifikasi baris `mst_file` ada (kategori diharapkan `foto_profil`/`foto_formal`/dokumen identitas). Tidak menulis file sendiri — hanya menyimpan id.
6. **Soft delete**: cegah/ peringatkan bila anggota masih menaut baris lain (FK `restrict` pada `users`, `kta`, `jadwal_peserta`, `absensi`). Karena kita soft-delete (bukan hard delete), FK tidak trigger; tetap kembalikan `ErrConflict` bila ada `kta` aktif / akun `users` aktif agar tidak menghapus anggota yang masih dipakai. (Minimal: cek `users` aktif & `kta` status `aktif`.)
7. Set audit (`created_by`/`modified_by` dari `Claims.UserID`).
8. Tulis `log_aktivitas` untuk create/update/delete (`modul="anggota"`, aksi `buat/ubah/hapus`, `reff_type="anggota"`, `reff_id`, `nilai_lama`/`nilai_baru`, IP, UA).

Sentinel errors dipetakan handler via `response.FromError`.

### 5.5 handler
Bind → `validator.Explain` pada error → panggil service dengan `middleware.Claims(c)` → envelope (`OK`/`Created`). Tidak menyentuh `*gorm.DB`.

### 5.6 migrations + seeders
- **Migrasi**: enum `jenis_anggota`, `status_anggota` (di `0001_enums`), tabel `anggota` (numbered SQL `.up`/`.down`, cocok DBML: kolom, FK restrict ke `mst_instansi`, FK ke `mst_wilayah`/`mst_file`, index `instansi_id`/`jenis_anggota`/`status_anggota`, **partial unique** `no_induk WHERE is_deleted=false`). No `AutoMigrate`.
- **Prasyarat migrasi**: `mst_instansi`, `mst_wilayah`, `mst_file`/`mst_file_varian` sudah ada (Fase 0/2).
- **Seeder permission**: pastikan `anggota.{create,read,update,delete}` ada di `mst_permission` + `role_permission` (di-seed Fase 1). Idempotent (`ON CONFLICT DO NOTHING`).
- **Seeder data anggota**: opsional beberapa contoh untuk dev; idempotent keyed pada natural key (mis. `no_identitas`). Tidak wajib.

### 5.7 edge cases & jobs
- Anggota tanpa `foto_formal` tidak bisa dicetak KTA (ditangani Fase 6, bukan di sini).
- `jenis_anggota = siswa_ke_dewasa`: modul ini hanya menyimpan enum; kewajiban 2 KTA ditegakkan modul KTA.
- Pencarian `q` harus menyertakan `no_induk` yang mungkin `NULL` (anggota belum ber-KTA) — gunakan `OR nama_lengkap ILIKE`.
- **Tidak ada background job** di modul ini.

---

## 6. Kebutuhan FRONTEND (Angular)

> **Metode desain WAJIB:** panggil taste-skill — umumkan **"Using design-taste-frontend"**, jalankan pre-flight/audit, map ke **Bootstrap 5** (sudah terpasang), dan **tegakkan token** di `workflow/_shared/design-tokens.md` (near-black bg, dark-gray surface, SLAM red accent, teks putih, heading UPPERCASE renggang). *Bila source `blog-fe` dipulihkan, cermin layout/menu-nya untuk layar ini; jika tidak, ikuti design-tokens.*

Folder: `src/app/pages/anggota/`.

### 6.1 Halaman/komponen
- `anggota-list.ts/.html/.scss` — tabel signal-based (§9 conventions-app), search box (`q`), filter dropdown `instansi_id`/`jenis`/`status`, pagination. Tombol **Tambah** di-gate `*hasPermission="'anggota.create'"`. Kolom: NRA (`no_induk`, tampil "—" bila kosong), Nama, Jenis, Status, Instansi, aksi (Detail/Edit/Hapus gated).
- `anggota-form.ts/.html/.scss` — Reactive Form untuk create & edit (route `new` dan `:id/edit`). Upload 3 foto lewat `FileService`.
- `anggota-detail.ts/.html/.scss` — tampilan read-only profil + foto (foto_formal & file_identitas privat → fetch blob).

### 6.2 Form (fields + validasi mirror DTO)
| Field | Kontrol | Validasi klien |
|---|---|---|
| `instansi_id` | select (dari `/instansi`) | `required` |
| `wilayah_id` | select (dari `/wilayah` atau default) | opsional |
| `nama_lengkap` | text | `required`, maxLength 150 |
| `nama_panggilan` | text | maxLength 50 |
| `jenis_anggota` | select `siswa/dewasa/siswa_ke_dewasa` | `required` (oneof) |
| `jenis_kelamin` | select `1 L / 2 P` | opsional |
| `jenis_identitas` | select (KTP/Kartu Pelajar/…) | opsional |
| `no_identitas` | text | maxLength 50 |
| `pekerjaan` | text | maxLength 100 |
| `alamat` | textarea | opsional |
| `kode_pos` | text | maxLength 10 |
| `tempat_lahir` | text | maxLength 100 |
| `tanggal_lahir` | `<input type="date">` | `required` |
| `tanggal_bergabung` | `<input type="date">` | opsional |
| `status_anggota` | select `aktif/non_aktif` | default `aktif` |
| `foto_profil` | file (image) → upload → `foto_profil_file_id` | opsional |
| `foto_formal` | file (image) → upload → `foto_formal_file_id` | opsional; note "latar merah, untuk KTA" |
| `file_identitas` | file → upload → `file_identitas_file_id` | opsional; note "PRIVAT" |

**`no_induk` ditampilkan read-only** (chip/label "NRA: 357…" atau "Belum ada KTA"), **tidak pernah** jadi kontrol form. Pada `422/400`, map `res.errors` (field→pesan) ke kontrol.

### 6.3 API service (`anggota.service.ts`)
Thin wrapper atas `ApiService` (unwrap envelope):
```ts
list(q)            -> api.get<Page<Anggota>>('/anggota', q)
detail(id)         -> api.get<Anggota>(`/anggota/${id}`)
create(body)       -> api.post<Anggota>('/anggota', body)
update(id, body)   -> api.put<Anggota>(`/anggota/${id}`, body)
remove(id)         -> api.delete<void>(`/anggota/${id}`)
```
Model `Anggota`/`AnggotaForm` di `core/models/`.

### 6.4 Guards + permission-gating
- Route: `canActivate: [authGuard, permissionGuard]`, `data: { permission: 'anggota.read' }`.
- Tombol Tambah/Edit/Hapus dibungkus `*hasPermission="'anggota.create|update|delete'"`. Ingat: gating UI = UX saja; server tetap otoritatif.

### 6.5 File handling
- Upload: `FormData` **tanpa** set `Content-Type` (browser set boundary) via `FileService.upload(field, file, {kategori})`. Simpan `uuid`→resolve id, atau backend menerima id. (Alur: unggah file dulu → dapat `uuid`/`file_id` → kirim `*_file_id` di body anggota.)
- Tampil foto privat (`foto_formal`, `file_identitas`) → `FileService.imageUrl(uuid,'medium')` (blob + Bearer), **revoke** object URL saat destroy. `foto_profil` boleh `low/medium`.

### 6.6 i18n
Namespace `ANGGOTA.*` + `COMMON.*` di `public/i18n/IND.json` & `ENG.json` (sinkron). Contoh keys: `ANGGOTA.TITLE`, `ANGGOTA.NAMA`, `ANGGOTA.NRA`, `ANGGOTA.JENIS`, `ANGGOTA.STATUS`, `ANGGOTA.FORM.FOTO_FORMAL`, `ANGGOTA.FORM.FILE_IDENTITAS_PRIVAT`, `ANGGOTA.FILTER.INSTANSI`, `COMMON.ADD/EDIT/DELETE/SAVE/CANCEL/SEARCH/DETAIL/EMPTY`, `VALIDATION.REQUIRED`.

---

## 7. ALUR form → API → DATABASE (kontrak koherensi)

### 7.1 Field form → payload key → kolom DB
| Field form (Angular) | Payload key (JSON) | Kolom DB (`anggota`) | Catatan |
|---|---|---|---|
| Instansi | `instansi_id` | `instansi_id` | FK, wajib |
| Wilayah | `wilayah_id` | `wilayah_id` | FK opsional |
| Nama lengkap | `nama_lengkap` | `nama_lengkap` | wajib |
| Nama panggilan | `nama_panggilan` | `nama_panggilan` | |
| Jenis anggota | `jenis_anggota` | `jenis_anggota` | enum |
| Jenis kelamin | `jenis_kelamin` | `jenis_kelamin` | int |
| Jenis identitas | `jenis_identitas` | `jenis_identitas` | int |
| No identitas | `no_identitas` | `no_identitas` | |
| Pekerjaan | `pekerjaan` | `pekerjaan` | |
| Alamat | `alamat` | `alamat` | dicetak di KTA |
| Kode pos | `kode_pos` | `kode_pos` | |
| Tempat lahir | `tempat_lahir` | `tempat_lahir` | dicetak KTA; **bukan** kode wilayah NRA |
| Tanggal lahir | `tanggal_lahir` | `tanggal_lahir` | **wajib**; memasok digit 5–8 NRA |
| Tanggal bergabung | `tanggal_bergabung` | `tanggal_bergabung` | |
| Status | `status_anggota` | `status_anggota` | manual; default `aktif` |
| Foto profil (upload) | `foto_profil_file_id` | `foto_profil_file_id` | id hasil `POST /files` |
| Foto formal (upload) | `foto_formal_file_id` | `foto_formal_file_id` | latar merah, untuk KTA |
| File identitas (upload) | `file_identitas_file_id` | `file_identitas_file_id` | **PRIVAT** |
| — (read-only chip) | *(tidak dikirim)* | `no_induk` | ditulis modul KTA, bukan di sini |
| — (auto) | — | `created_by`/`modified_by` | dari JWT Claims |
| — (auto) | — | `is_deleted`/`deleted_*` | soft delete |

### 7.2 Aksi user → endpoint → tulisan DB
| Aksi user | Endpoint | Tulisan DB |
|---|---|---|
| Unggah foto (profil/formal/identitas) | `POST /files` (multipart) | 1 baris `mst_file` + 3 `mst_file_varian`; kembalikan `uuid`/`file_id` |
| Simpan anggota baru | `POST /anggota` | INSERT `anggota` (`no_induk=NULL`, audit terisi) + 1 baris `log_aktivitas` (buat) |
| Ubah anggota | `PUT /anggota/{id}` | UPDATE `anggota` (kecuali `no_induk`) + `modified_*` + `log_aktivitas` (ubah) |
| Hapus anggota | `DELETE /anggota/{id}` | UPDATE `anggota` set `is_deleted=true, deleted_at, deleted_by` + `log_aktivitas` (hapus) |
| Lihat daftar/detail | `GET /anggota`, `GET /anggota/{id}` | SELECT (hanya `is_deleted=false`) — tidak menulis |
| Lihat foto privat | `GET /files/{uuid}/{varian}` | SELECT + stream (cek izin) — tidak menulis |

> **NRA/`no_induk` diisi di Fase 6:** saat `POST /kta` menerbitkan kartu aktif, service KTA meng-UPDATE `anggota.no_induk = kta.no_kta`. Modul anggota hanya membacanya.

---

## 8. Dependencies & Acceptance criteria

### 8.1 Prasyarat
- **Fase 0** selesai: skeleton Go+Angular, `mst_pengaturan`, **file layer** (`POST /files`, `GET /files/{uuid}/{varian}`, 3 varian), auth (login/JWT), `log_aktivitas`.
- **Fase 1** selesai: RBAC (5 tabel, seed matriks, `RequirePermission`, guard/directive Angular, sidebar dinamis). Permission `anggota.*` ter-seed.
- Tabel master **`mst_instansi`** & **`mst_wilayah`** ada (migrasi Fase 2 / seed). `mst_file`/`mst_file_varian` ada.
- Modul KTA (Fase 6) adalah **konsumen** `no_induk`, bukan prasyarat modul ini.

### 8.2 Acceptance criteria (checklist)
- [ ] Migrasi `anggota` cocok `.dbml` (kolom, enum, FK restrict `mst_instansi`, index, partial-unique `no_induk WHERE is_deleted=false`); `slamctl migrate up/down` jalan.
- [ ] `go build ./...` & `go vet ./...` bersih; modul terdaftar di `router.go`.
- [ ] `GET /anggota` mengembalikan envelope terpaginasi; filter `instansi_id`/`jenis`/`status` & search `q` (no_induk/nama) bekerja; hanya `is_deleted=false`.
- [ ] `POST /anggota` membuat baris (`no_induk` **NULL**, mengabaikan `no_induk` bila dikirim); validasi `instansi_id`/`nama_lengkap`/`jenis_anggota`/`tanggal_lahir` ditegakkan; `422` memetakan field.
- [ ] `PUT /anggota/{id}` tidak pernah mengubah `no_induk`.
- [ ] `DELETE /anggota/{id}` = soft delete; menolak bila anggota masih punya akun `users` aktif / KTA `aktif` (`409/conflict`).
- [ ] **Permission ditegakkan server-side**: User (role 30) ditolak `403` pada create/update/delete; Guest/anonim ditolak; SA bypass. Bukan sekadar tombol tersembunyi.
- [ ] `log_aktivitas` tertulis untuk create/update/delete.
- [ ] **APP**: taste-skill dipanggil; token diterapkan; list+form+detail responsif; foto privat via blob endpoint (bukan `<img src>` token-less); `*hasPermission` menyembunyikan aksi yang tak diizinkan; i18n IND/ENG sinkron.
- [ ] Alur end-to-end: submit form → `POST /anggota` → baris DB baru terbaca di list; upload foto → `*_file_id` tersimpan → foto tampil di detail.
