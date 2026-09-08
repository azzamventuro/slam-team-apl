# Modul 10 — Master Data Prestasi (`prestasi`)

> Dokumen penjelasan modul. Sumber otoritatif: `slamteam_db.dbml` (skema) >
> rancangan Bab 9/10 (rute/migrasi) > `_shared/*.md` > app lama (tidak pernah).
> Rute, payload, dan nama kolom mengikuti `.dbml` + peta endpoint APA ADANYA.

---

## 1. Ringkasan & tujuan modul

**Prestasi** mendata capaian/kompetisi (achievements) milik seorang anggota:
judul kompetisi, peringkat, tingkat, tanggal, lokasi, plus dua berkas gambar
(flyer kompetisi + foto sampul). Ini adalah salah satu dari **19 modul**,
tepatnya modul **`prestasi`** pada grup **Master Data**.

- **Fase:** Fase 2 (Anggota + master data) — dibangun setelah Fase 0 (fondasi +
  file layer) dan Fase 1 (RBAC dinamis) siap.
- **Grup:** Master Data (`mst_modul.grup = 'Master Data'`).
- **Sifat khusus:** data ini **publik** — Guest boleh membaca (matrix Guest = R).
  Tampil di landing publik (Fase 8) dan disematkan pada `GET /public/profil`.
  Modul ini menyediakan CRUD standar `/prestasi` untuk panel admin, plus satu
  daftar publik `GET /public/prestasi` (tanpa auth, hanya field aman-publik).
- **Pola:** modul CRUD berulang (rancangan Bab 11.4 / 12.2 "modul CRUD dengan
  pola berulang") — sama bentuk dengan instansi/unit/inorga/medsos/dokumen.

---

## 2. Tabel & kolom

Tabel tunggal **`prestasi`** (blok `Table prestasi` di `slamteam_db.dbml`,
baris ~1064). Kolom persis:

| Kolom | Tipe (DBML) | Aturan / catatan |
|---|---|---|
| `id` | `bigint [pk, increment]` | PK bigint identity. Dipakai di rute admin `/{id}`; JANGAN dibocorkan ke URL publik. |
| `anggota_id` | `bigint [not null]` | FK → `anggota.id`. Pemilik prestasi. **Wajib.** |
| `kode` | `varchar(50)` | Kode bisnis opsional (mis. `PRS-0001`). Bebas, nullable. |
| `peringkat` | `varchar(50)` | Peringkat/juara (mis. `Juara 1`, `Runner-up`). Teks bebas (BUKAN enum di DBML). |
| `tingkat` | `varchar(50)` | Tingkat lomba (mis. `Nasional`, `Provinsi`, `Kabupaten`, `Internasional`). Teks bebas. |
| `judul_kompetisi` | `varchar(200)` | Nama kompetisi. Field identitas utama. |
| `flyer_file_id` | `bigint` | FK → `mst_file.id` (nullable → `*int64`). Gambar flyer kompetisi. |
| `tanggal_kompetisi` | `date` | Tanggal pelaksanaan kompetisi (nullable). |
| `alamat_kompetisi` | `text` | Alamat/lokasi kompetisi (nullable). |
| `foto_sampul_file_id` | `bigint` | FK → `mst_file.id` (nullable → `*int64`). Foto sampul/dokumentasi. |
| `keterangan` | `text` | Catatan bebas (nullable). |
| `is_deleted` | `boolean [default: false]` | **Soft delete** — jangan hard-delete. |
| `deleted_at` | `timestamptz` | Diisi saat soft delete. UTC. |
| `deleted_by` | `bigint` | User yang menghapus. |
| `created_at` | `timestamptz [default: now()]` | UTC. |
| `created_by` | `bigint` | User pembuat. |
| `modified_at` | `timestamptz` | UTC. |
| `modified_by` | `bigint` | User pengubah terakhir. |

**Relasi (dari `.dbml`):**
- `prestasi.anggota_id > anggota.id` (baris 1227).
- `prestasi.flyer_file_id > mst_file.id` (baris 1228).
- `prestasi.foto_sampul_file_id > mst_file.id` (baris 1229).

`anggota` disebut restrict oleh banyak tabel; di sini prestasi dihapus lunak
saat anggota dihapus lunak (tidak ada FK cascade — dikelola di service/tampilan).

**Index yang dibuat migrasi ini** (DBML tak mencantumkan index eksplisit untuk
`prestasi`, tambahkan yang wajar mengikuti konvensi Bab 10):
- Index `anggota_id` (filter "prestasi milik anggota X" & join).
- Index parsial soft-delete pada query list (`WHERE is_deleted = false`).
- Opsional: unique parsial `kode WHERE is_deleted = false` bila `kode` diisi
  (SKIP jika `kode` tetap bebas/boleh kosong — jangan paksakan unik pada NULL).

**Konvensi lintas modul yang berlaku:**
- Semua timestamp `timestamptz` UTC (`_shared/conventions-api.md` §6/§9).
- Soft delete eksplisit `is_deleted + deleted_at + deleted_by` (BUKAN
  `gorm.DeletedAt`).
- `*_file_id` adalah FK ke `mst_file.id` — jangan simpan path/URL di baris
  prestasi; resolusi lewat file layer (§10 conventions-api).

---

## 3. Endpoint

Semua di bawah `/api/v1`. CRUD admin dijaga `middleware.JWTAuth` +
`perm.Require("prestasi.<aksi>")`. Daftar publik tanpa auth.

| Method | Path | Permission (`modul.aksi`) | Auth | Fungsi |
|---|---|---|---|---|
| GET | `/prestasi` | `prestasi.read` | bearer | List + filter + paginate. |
| GET | `/prestasi/{id}` | `prestasi.read` | bearer | Satu record. |
| POST | `/prestasi` | `prestasi.create` | bearer | Buat. |
| PUT | `/prestasi/{id}` | `prestasi.update` | bearer | Ubah. |
| DELETE | `/prestasi/{id}` | `prestasi.delete` | bearer | Soft delete. |
| GET | `/public/prestasi` | — (publik) | none | Daftar publik (Guest R), field aman-publik saja. |

> Catatan koherensi: peta endpoint otoritatif (`_shared/api-endpoints.md` §12)
> menyebut prestasi publik disematkan pada `GET /public/profil`. Modul ini
> memiliki **service** yang menghasilkan daftar publik; modul `profile_club`
> memanggil service tsb untuk menyematkannya. `GET /public/prestasi` adalah
> permukaan landing Fase 8 ("public profil/artikel/kegiatan/prestasi") dan
> realisasi konkret sel matrix Guest = R. Keduanya memakai proyeksi field
> aman-publik yang sama.

### Payload & response

**List query** (`GET /prestasi`) — mengikuti bentuk standar
`_shared/conventions-api.md` §5:
```
?page=1&per_page=20&q=<judul>&sort=-tanggal_kompetisi
 &anggota_id=<id>&tingkat=<t>&peringkat=<p>
 &tanggal_dari=YYYY-MM-DD&tanggal_sampai=YYYY-MM-DD
```
- `q` mencari `judul_kompetisi ILIKE %q%`.
- `sort` di-whitelist: `tanggal_kompetisi`, `created_at`, `judul_kompetisi`,
  `peringkat`, `tingkat` (prefiks `-` = DESC). Jangan interpolasi ke SQL.

**Response list** (dalam `data`):
```json
{ "success": true, "message": "OK", "data": {
  "items": [ { PrestasiResponse } ],
  "page": 1, "per_page": 20, "total": 42, "last_page": 3
} }
```

**Create/Update body** (`application/json`):
```json
{
  "anggota_id": 12,
  "kode": "PRS-0001",
  "peringkat": "Juara 1",
  "tingkat": "Nasional",
  "judul_kompetisi": "SLAM Open Airsoft Championship 2026",
  "flyer_file_id": 88,
  "tanggal_kompetisi": "2026-05-17",
  "alamat_kompetisi": "GOR Ken Arok, Kota Malang",
  "foto_sampul_file_id": 89,
  "keterangan": "Kategori CQB beregu."
}
```
Field file (`flyer_file_id`, `foto_sampul_file_id`) diisi dengan `id` hasil
`POST /files` (upload dahulu, baru simpan id). Boleh null / dihilangkan.

**PrestasiResponse** (single & item list):
```json
{
  "id": 5,
  "anggota_id": 12,
  "anggota": { "id": 12, "nama_lengkap": "Achmad", "no_induk": "35731002021" },
  "kode": "PRS-0001",
  "peringkat": "Juara 1",
  "tingkat": "Nasional",
  "judul_kompetisi": "SLAM Open Airsoft Championship 2026",
  "flyer_file_id": 88,
  "flyer": { "file_id": 88, "uuid": "b1f2...", "is_publik": true },
  "tanggal_kompetisi": "2026-05-17",
  "alamat_kompetisi": "GOR Ken Arok, Kota Malang",
  "foto_sampul_file_id": 89,
  "foto_sampul": { "file_id": 89, "uuid": "c3d4...", "is_publik": true },
  "keterangan": "Kategori CQB beregu.",
  "created_at": "2026-05-18T02:00:00Z",
  "modified_at": null
}
```
- Nested `anggota`/`flyer`/`foto_sampul` di-preload agar frontend bisa render
  gambar via `GET /files/{uuid}/{varian}` tanpa round-trip tambahan.
- Field audit `deleted_*` TIDAK dikembalikan (json:"-").

**Public list** (`GET /public/prestasi?anggota_id=&tingkat=&page=&per_page=`):
proyeksi aman-publik saja — TIDAK ada `keterangan` internal sensitif bila
dianggap privat? `keterangan` bersifat deskriptif prestasi → aman. Yang
DIKELUARKAN: audit, id anggota internal boleh, tapi jangan bocorkan data
pribadi anggota (alamat/DOB/kontak). Bentuk item publik:
```json
{
  "judul_kompetisi": "...", "peringkat": "Juara 1", "tingkat": "Nasional",
  "tanggal_kompetisi": "2026-05-17", "alamat_kompetisi": "...",
  "anggota": { "nama_lengkap": "Achmad", "nama_panggilan": "Mad", "no_induk": "35731002021" },
  "flyer_uuid": "b1f2...", "foto_sampul_uuid": "c3d4..."
}
```
Gambar prestasi diunggah dengan `is_publik = true` (kategori `prestasi`) agar
halaman publik dapat menampilkannya (lihat §6 file handling).

---

## 4. Hak akses (dari matrix rancangan Bab 3.5)

| Modul | Super Admin (0) | Admin (10) | Moderator (20) | User (30) | Guest (99) |
|---|---|---|---|---|---|
| `prestasi` | **CRUD** | **CRUD** | **CRUD** | **R** | **R** |

- **Cakupan:** semua sel adalah **`semua`** (tidak ada penyempitan
  `instansi_sendiri`/`milik_sendiri` untuk prestasi — prestasi bersifat publik
  dan lintas anggota). Middleware tetap memasang `perm.Require`, cakupan resolve
  ke `semua` untuk read/write.
- **User = R:** hanya boleh GET (`prestasi.read`); tidak boleh create/update/
  delete. Tombol create/edit/hapus disembunyikan di UI DAN diblok handler.
- **Guest = R:** direalisasikan lewat `GET /public/prestasi` (tanpa auth) dan
  penyematan pada `/public/profil`. Guest TIDAK mengakses `/prestasi` (berjaga
  JWT). Ini penting: menyembunyikan tombol bukan keamanan — setiap izin UI punya
  cek handler Go yang cocok.
- Permission string yang harus ada di seed RBAC (Fase 1): `prestasi.read`,
  `prestasi.create`, `prestasi.update`, `prestasi.delete` (baris `mst_permission`
  + `role_permission`). Modul ini TIDAK re-seed RBAC; ia bergantung pada seed
  Fase 1.

---

## 5. Kebutuhan BACKEND (Go)

Lokasi: `internal/modules/core/prestasi/` mengikuti pola modul
(`_shared/conventions-api.md` §1).

### domain/
- `prestasi.go` — entity GORM `Prestasi` memetakan tabel `prestasi`:
  - `ID int64 gorm:"primaryKey"`, `AnggotaID int64`, `Kode *string`,
    `Peringkat *string`, `Tingkat *string`, `JudulKompetisi string`,
    `FlyerFileID *int64`, `TanggalKompetisi *datatypes.Date` (atau
    `*time.Time` kolom `date`), `AlamatKompetisi *string`,
    `FotoSampulFileID *int64`, `Keterangan *string`, + embed `Audit`
    (`created_at/by`, `modified_at/by`, `is_deleted/deleted_at/deleted_by`).
  - `func (Prestasi) TableName() string { return "prestasi" }`.
  - Relasi Preload: `Anggota` (belongs-to, pilih kolom aman), `FlyerFile` &
    `FotoSampulFile` (belongs-to `mst_file`, ambil `id, uuid, is_publik`).
- Reuse `Audit` bersama dan scope soft-delete dari paket shared bila sudah ada
  (Fase 0/2); JANGAN fork.

### dto/
- `CreatePrestasiReq` / `UpdatePrestasiReq` (binding tags mirror kolom):
  - `AnggotaID int64 binding:"required,gt=0"`
  - `Kode *string binding:"omitempty,max=50"`
  - `Peringkat *string binding:"omitempty,max=50"`
  - `Tingkat *string binding:"omitempty,max=50"`
  - `JudulKompetisi string binding:"required,max=200"`
  - `FlyerFileID *int64 binding:"omitempty,gt=0"`
  - `TanggalKompetisi *string binding:"omitempty,datetime=2006-01-02"`
  - `AlamatKompetisi *string binding:"omitempty"`
  - `FotoSampulFileID *int64 binding:"omitempty,gt=0"`
  - `Keterangan *string binding:"omitempty"`
- `ListPrestasiQuery` (form tags): `Page`, `PerPage`, `Q`, `Sort`, `AnggotaID`,
  `Tingkat`, `Peringkat`, `TanggalDari`, `TanggalSampai`.
- `PrestasiResponse`, `FileRef{FileID,UUID,IsPublik}`,
  `AnggotaRef{ID,NamaLengkap,NoInduk}`, `PublicPrestasiItem` (proyeksi publik).
- Mapper `ToResponse(e domain.Prestasi) PrestasiResponse`.

### repository/
- `NewPrestasiRepository(db *gorm.DB)`.
- Metode: `List(ctx, q, scope) ([]Prestasi, int64, error)` — bangun query dasar
  `WHERE is_deleted = false`, terapkan filter (`anggota_id`, `tingkat`,
  `peringkat`, `q` ILIKE judul, rentang `tanggal_kompetisi`), whitelist sort,
  `Count` lalu `Offset/Limit` `Find` dengan `Preload` Anggota/Flyer/FotoSampul.
- `FindByID(ctx, id)` — `WHERE id=? AND is_deleted=false`, Preload; map
  `gorm.ErrRecordNotFound → ErrNotFound`.
- `Create(ctx, *Prestasi)`, `Update(ctx, *Prestasi)`.
- `SoftDelete(ctx, id, actorID)` — set `is_deleted=true, deleted_at=now(),
  deleted_by=actor` (UPDATE, bukan DELETE).
- `PublicList(ctx, q) ([]Prestasi, int64, error)` — sama seperti List tapi hanya
  gabung anggota yang `is_deleted=false` dan proyeksi kolom aman.
- `anggotaExists(ctx, id)` helper (cek FK sebelum insert).

### service/
- `NewPrestasiService(repo, anggotaRepo/fileRepo opsional)`.
- Aturan bisnis & validasi (di service, bukan binding):
  1. **anggota_id wajib ada** dan `is_deleted=false` → jika tidak `ErrValidation`
     ("anggota tidak ditemukan").
  2. **File refs valid**: bila `flyer_file_id`/`foto_sampul_file_id` diisi,
     pastikan row `mst_file` ada, `is_deleted=false`, dan `kategori` sesuai
     gambar. (Cek ringan; jangan wajibkan keduanya.)
  3. **tanggal_kompetisi tidak boleh di masa depan** relatif ke tanggal server
     UTC (prestasi = capaian yang sudah terjadi) → `ErrValidation`. Guard
     sederhana; longgarkan bila klien butuh mencatat lomba mendatang.
  4. **RBAC cakupan** = `semua` → tidak ada penyempitan query. Middleware sudah
     menegakkan izin per-aksi.
  5. Set audit: `created_by`/`modified_by` = `Claims(c).UserID`.
- Update: muat existing (404 bila tidak ada), terapkan perubahan field yang
  dikirim, set `modified_at/by`.
- Delete: soft delete + tulis `log_aktivitas` (aksi `hapus`, modul `prestasi`,
  `reff_id`).
- Audit bisnis (`log_aktivitas`) untuk create/ubah/hapus (conventions §12).
- `PublicList` memakai proyeksi aman-publik (tanpa data pribadi anggota).

### handler/
- `List`, `Detail`, `Create`, `Update`, `Delete` (bearer + perm) + `PublicList`
  (publik). Pola: bind DTO → `response.Unprocess(validator.Explain(err))` bila
  gagal → panggil service → `response.FromError` untuk sentinel → envelope OK/
  Created. Tidak menyentuh `*gorm.DB`.

### main.prestasi.go
- `Initialize(db, jwtMgr, perm)` merangkai repo→service→handler.
- `SetupRoutes(rg)`:
  ```go
  g := rg.Group("/prestasi", middleware.JWTAuth(m.jwtMgr))
  g.GET("",        m.perm.Require("prestasi.read"),   m.h.List)
  g.GET("/:id",    m.perm.Require("prestasi.read"),   m.h.Detail)
  g.POST("",       m.perm.Require("prestasi.create"), m.h.Create)
  g.PUT("/:id",    m.perm.Require("prestasi.update"), m.h.Update)
  g.DELETE("/:id", m.perm.Require("prestasi.delete"), m.h.Delete)
  // publik (tanpa JWT) — daftarkan pada grup publik router
  pub.GET("/prestasi", m.h.PublicList)
  ```
- Registrasi di `internal/router/router.go`:
  `prestasi.Initialize(db, jwtMgr, permGuard).SetupRoutes(apiV1)` dan sematkan
  rute publik pada grup `/public`.

### Migrasi & seeder
- Migrasi bernomor (mengikuti Fase 2, mis. `00xx_prestasi.up.sql` /
  `.down.sql`) yang membuat tabel `prestasi` **persis** kolom DBML:
  - `CREATE TABLE prestasi (...)` dengan FK `anggota_id → anggota(id)`,
    `flyer_file_id → mst_file(id)`, `foto_sampul_file_id → mst_file(id)`.
  - `CREATE INDEX ON prestasi (anggota_id)`.
  - `CREATE INDEX ON prestasi (tanggal_kompetisi)` (untuk sort/rentang) — opsional.
  - `down.sql`: `DROP TABLE prestasi`.
- **Seeder:** tidak ada data referensi (isi dibuat pengguna). Baris `mst_modul`
  `prestasi` + `mst_permission` `prestasi.read|create|update|delete` +
  `role_permission` di-seed pada **Fase 1** (dependency), bukan di sini —
  modul ini hanya mengandalkan keberadaannya.

### Edge cases
- File ref menunjuk `mst_file` yang sudah dihapus → tolak (ErrValidation).
- Anggota dihapus lunak setelah prestasi dibuat → prestasi tetap ada; list admin
  dapat menandai anggota nonaktif; public list menyaring anggota `is_deleted`.
- `q`/`sort` tak valid → abaikan sort tak dikenal (fallback `-tanggal_kompetisi`),
  jangan error 500.
- Update tanpa mengubah file → biarkan `*_file_id` lama.

### Background jobs
- Tidak ada. (Modul CRUD murni.)

---

## 6. Kebutuhan FRONTEND (Angular)

Lokasi: `pages/prestasi/` (standalone, signals, zoneless) mengikuti
`_shared/conventions-app.md`. Bangun ke dalam scaffold; jangan buat layout baru.

**WAJIB desain:** awali dengan meng-invoke Taste Skill — umumkan
**"Using design-taste-frontend"**, jalankan pre-flight/audit, petakan ke
Bootstrap 5, dan **tegakkan** token di
`D:/xampp/htdocs/slam-team-apl/workflow/_shared/design-tokens.md` (dark default,
SLAM red aksen, heading UPPERCASE wide-tracking, fokus ring merah, min tap 44px).
Catatan: *"Jika sumber `blog-fe` dipulihkan, tiru layout/menu-nya untuk layar
ini; jika tidak, ikuti design tokens."*

### Halaman/komponen
- `pages/prestasi/prestasi-list.ts` (+ `.html`, `.scss`) — daftar CRUD kanonik
  (§9 conventions-app): filter signal (`q`, `anggota_id`, `tingkat`,
  `tanggal_dari/sampai`), pagination, `toSignal` params-driven + `switchMap`.
  Kolom tabel: Judul Kompetisi, Peringkat, Tingkat, Tanggal, Anggota (nama +
  NRA), aksi. Thumbnail foto_sampul (`low`) opsional.
- `pages/prestasi/prestasi-form.ts` — form create/edit (reactive, typed
  `nonNullable`), dipakai untuk tambah & ubah.
- `pages/prestasi/prestasi-detail.ts` — tampilan read-only (untuk role User = R)
  menampilkan flyer + foto sampul (`medium`) via FileService.
- `pages/prestasi/prestasi.service.ts` — wrapper tipis atas `ApiService`.

### Form (fields → mirror DTO)
| Field UI | Kontrol | Validator klien (mirror DTO) |
|---|---|---|
| `anggota_id` | select anggota (async search) | required, gt 0 |
| `judul_kompetisi` | text | required, maxLength 200 |
| `peringkat` | text / select bebas (`Juara 1`, ...) | maxLength 50 |
| `tingkat` | select (`Internasional/Nasional/Provinsi/Kabupaten/Kota/Klub` + lainnya) | maxLength 50 |
| `kode` | text | maxLength 50, opsional |
| `tanggal_kompetisi` | `<input type="date">` | opsional, ≤ hari ini |
| `alamat_kompetisi` | textarea | opsional |
| `keterangan` | textarea | opsional |
| `flyer_file_id` | file upload (gambar) | opsional |
| `foto_sampul_file_id` | file upload (gambar) | opsional |

`<input type="date">` native (rung 4 — hindari lib picker). Pada `422/400`,
map `res.errors` (field→pesan) ke kontrol yang cocok (`setErrors({server})`).

### API service methods (envelope via `ApiService`)
```ts
list(q)              -> api.get<Page<Prestasi>>('/prestasi', q)
detail(id)           -> api.get<Prestasi>(`/prestasi/${id}`)
create(body)         -> api.post<Prestasi>('/prestasi', body)
update(id, body)     -> api.put<Prestasi>(`/prestasi/${id}`, body)
remove(id)           -> api.delete<void>(`/prestasi/${id}`)
publicList(q)        -> api.get<Page<PublicPrestasi>>('/public/prestasi', q)
```
`ApiService` sudah membuka envelope `{success,message,data,errors}`.

### File handling
- Upload flyer & foto_sampul lewat `FileService.upload('file', file, {kategori:'prestasi', is_publik:'true'})` →
  dapat `{uuid, file_id}` → simpan `file_id` ke form → submit prestasi.
  Kirim `FormData` **tanpa** set `Content-Type` (browser set boundary;
  authInterceptor tetap menambah Bearer).
- Render gambar di panel admin: `FileService.imageUrl(uuid, 'medium'|'low')`
  (fetch blob + object URL; revoke saat destroy).
- Halaman publik memakai `flyer_uuid`/`foto_sampul_uuid` (file `is_publik=true`).

### Permission-gating (mirror backend; bukan keamanan)
- Route: `{ path: 'prestasi', canActivate: [authGuard, permissionGuard],
  data: { permission: 'prestasi.read' } }`.
- Tombol Tambah: `*hasPermission="'prestasi.create'"`; Edit:
  `'prestasi.update'`; Hapus: `'prestasi.delete'`. User (R) melihat daftar +
  detail saja.

### i18n keys (`public/i18n/{IND,ENG}.json`, namespace `PRESTASI`)
`PRESTASI.TITLE`, `PRESTASI.ADD`, `PRESTASI.EDIT`,
`PRESTASI.FIELD.ANGGOTA`, `.JUDUL`, `.PERINGKAT`, `.TINGKAT`, `.KODE`,
`.TANGGAL`, `.ALAMAT`, `.FLYER`, `.FOTO_SAMPUL`, `.KETERANGAN`,
`PRESTASI.EMPTY`, plus `COMMON.*` & `VALIDATION.*` yang sudah ada. IND & ENG
harus sinkron (kunci sama). Tidak ada string keras di template.

---

## 7. ALUR form → API → DATABASE (kontrak koherensi)

### Peta field (form → payload → kolom DB)
| Field form (Angular) | Key payload JSON | Kolom DB `prestasi` |
|---|---|---|
| Anggota (select) | `anggota_id` | `anggota_id` |
| Kode | `kode` | `kode` |
| Peringkat | `peringkat` | `peringkat` |
| Tingkat | `tingkat` | `tingkat` |
| Judul Kompetisi | `judul_kompetisi` | `judul_kompetisi` |
| Upload flyer → `file_id` | `flyer_file_id` | `flyer_file_id` (FK `mst_file.id`) |
| Tanggal Kompetisi | `tanggal_kompetisi` (`YYYY-MM-DD`) | `tanggal_kompetisi` (`date`) |
| Alamat Kompetisi | `alamat_kompetisi` | `alamat_kompetisi` |
| Upload foto sampul → `file_id` | `foto_sampul_file_id` | `foto_sampul_file_id` (FK `mst_file.id`) |
| Keterangan | `keterangan` | `keterangan` |
| (server) actor | — | `created_by` / `modified_by` |

### Aksi user → endpoint → tulisan DB
| Aksi | Endpoint | Tulisan DB |
|---|---|---|
| Unggah gambar | `POST /files` (multipart) | INSERT `mst_file` (+3 `mst_file_varian`), return `uuid`,`file_id` |
| Simpan (tambah) | `POST /prestasi` | INSERT `prestasi` (semua field + `created_at=now()`,`created_by`) + INSERT `log_aktivitas` |
| Simpan (edit) | `PUT /prestasi/{id}` | UPDATE `prestasi` field berubah + `modified_at=now()`,`modified_by` + `log_aktivitas` |
| Hapus | `DELETE /prestasi/{id}` | UPDATE `prestasi` SET `is_deleted=true,deleted_at=now(),deleted_by` + `log_aktivitas` |
| Lihat daftar | `GET /prestasi?...` | SELECT `WHERE is_deleted=false` + filter, Preload anggota/file |
| Lihat detail | `GET /prestasi/{id}` | SELECT satu row (bukan deleted), Preload |
| Publik | `GET /public/prestasi` | SELECT proyeksi aman-publik, anggota `is_deleted=false` |

Nama key payload = nama kolom DB persis (kecuali `tanggal_dari/sampai` yang
hanya filter query, bukan kolom). Tidak ada penggantian nama.

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- **Fase 0**: file layer (`mst_file` + `mst_file_varian`, `POST /files`,
  `GET /files/{uuid}/{varian}`), `log_aktivitas`, konvensi timestamptz.
- **Fase 1**: RBAC dinamis — `mst_modul` `prestasi`, `mst_permission`
  `prestasi.*`, `role_permission`, `PermGuard.Require`, guard/directive frontend.
- **Fase 2**: modul `anggota` (tabel + endpoint) — `anggota_id` merujuk ke sana;
  select anggota di form mengambil dari `GET /anggota`.
- `response.Forbidden`/`NotFound` sudah ditambahkan ke `shared/response`.

### Acceptance criteria (checkable)
- [ ] Tabel `prestasi` tercipta persis kolom `.dbml`; FK ke `anggota` & `mst_file`
      terpasang; index `anggota_id` ada; `down.sql` mengembalikan.
- [ ] `GET /prestasi` mengembalikan envelope terpaginasi; filter `q`,
      `anggota_id`, `tingkat`, rentang tanggal bekerja; sort di-whitelist.
- [ ] `POST /prestasi` menyisipkan baris; `anggota_id` tak valid → 422; tanggal
      masa depan → 422; file_id tak ada → 422.
- [ ] `PUT`/`DELETE` bekerja; DELETE soft (row tetap ada, `is_deleted=true`).
- [ ] Setiap rute admin menegakkan `prestasi.<aksi>`; User (R) diblok pada
      create/update/delete dengan 403; Guest tanpa token diblok di `/prestasi`.
- [ ] `GET /public/prestasi` bisa diakses tanpa auth; hanya field aman-publik;
      anggota terhapus tidak muncul.
- [ ] Gambar prestasi ter-resolve via `GET /files/{uuid}/{varian}`;
      flyer/foto_sampul tampil di detail.
- [ ] Frontend: form submit → row DB tercipta/terubah; validasi mirror DTO;
      tombol tersembunyi & terblok sesuai izin; responsif; token desain terpakai
      (dark, SLAM red, heading UPPERCASE, fokus ring merah).
- [ ] `log_aktivitas` tertulis untuk create/ubah/hapus.
- [ ] i18n IND/ENG sinkron; tidak ada string keras.
