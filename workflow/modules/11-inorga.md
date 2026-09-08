# Modul 11 — Master Data Inorga (Kepengurusan)

> nn=11 · key=`inorga` · Fase 2 (Anggota + master data) · Grup **Master Data** ·
> Modul ke-8 dari kelompok Master Data dalam daftar 19 modul.
>
> Sumber otoritatif dibaca untuk dokumen ini: `slamteam_db.dbml` (tabel
> `mst_inorga`, relasi, konvensi header + Bab 8.4), `workflow/.rancangan.txt`
> (Bab 3.5 matriks hak akses, Bab 8 konvensi, Bab 9 endpoint, Bab 12 fase),
> `_shared/conventions-api.md`, `_shared/conventions-app.md`,
> `_shared/design-tokens.md`, `_shared/api-endpoints.md`,
> `_shared/permission-matrix.md`. **Bila prosa dan `.dbml` berbeda, `.dbml`
> menang.**

---

## 1. Ringkasan & tujuan modul

**Inorga** = *informasi organisasi* / **kepengurusan** SLAM Team: satu baris per
**periode kepengurusan** (struktur organisasi untuk rentang waktu tertentu,
mis. "Kepengurusan 2024–2026"). Tiap periode menyimpan **kode**, **nama**, dua
gambar (**logo** + **banner**), rentang tanggal (**tanggal_mulai** /
**tanggal_selesai**), sebuah **berkas SK** (Surat Keputusan pengangkatan
pengurus), dan **konten** teks bebas (susunan pengurus / jabatan / deskripsi).

- **Modul ke-8 kelompok Master Data**, dibangun di **Fase 2** dengan pola CRUD
  master-data berulang yang sama seperti `instansi`, `unit`, `prestasi`.
- **Prasyarat keras:** File layer Fase 0 (`mst_file` + `mst_file_varian` +
  `POST /files`), RBAC Fase 1 (middleware `RequirePermission`, seed `mst_modul`
  + `mst_permission` + `role_permission`), `log_aktivitas` Fase 0.
- **Relasi keluar:** `jadwal.inorga_id → mst_inorga.id` — sebuah jadwal boleh
  menautkan dirinya ke periode kepengurusan (dipakai di Fase 3, bukan di sini).
  Karena itu **DELETE = soft delete** (baris tidak boleh hilang selagi jadwal
  historis merujuknya).
- **Tidak ada** endpoint publik khusus untuk inorga di peta endpoint Bab 9
  (kepengurusan bisa muncul di landing lewat `/public/profil` di Fase 8, tapi
  bukan tanggung jawab modul ini). Semua route inorga di fase ini
  **ber-auth (bearer) + RequirePermission**.

Tujuan: pengelola (SA/Admin) dapat menambah/mengubah/menonaktifkan periode
kepengurusan beserta logo, banner, SK, dan susunan pengurus; peran lain
(Moderator/User) hanya membaca.

---

## 2. Tabel & kolom — `mst_inorga`

Definisi persis dari `slamteam_db.dbml` (`Table mst_inorga`):

| Kolom | Tipe (DBML) | Aturan / catatan |
|---|---|---|
| `id` | `bigint` `[pk, increment]` | PK identity. **Tidak** dipakai di URL publik (tak ada URL publik di sini). |
| `kode` | `varchar(50)` | Kode periode (mis. `PENGURUS-2024`). **Tidak unik** menurut `.dbml` (tak ada unique index di Bab 8.4) — jangan paksakan unik. Opsional. |
| `nama` | `varchar(150)` | Nama kepengurusan. **Wajib** di DTO (business rule); dihardening `NOT NULL` di migrasi (lihat §5). |
| `logo_file_id` | `bigint` | FK → `mst_file.id`, **ON DELETE SET NULL** (Bab 8.4: relasi `*_file_id` = SET NULL). Nullable → `*int64`. |
| `banner_file_id` | `bigint` | FK → `mst_file.id`, **ON DELETE SET NULL**. Nullable → `*int64`. |
| `tanggal_mulai` | `date` | Tanggal (BUKAN timestamptz — tanpa zona). Awal periode. |
| `tanggal_selesai` | `date` | Akhir periode. **Nullable** = kepengurusan masih berjalan. |
| `file_sk_file_id` | `bigint` | FK → `mst_file.id`, **ON DELETE SET NULL**. Berkas SK (dokumen). Nullable → `*int64`. |
| `konten` | `text` | Susunan pengurus / deskripsi. Boleh HTML/markdown/plain; sanitasi saat render. Opsional. |
| `is_deleted` | `boolean` `[default: false]` | Soft delete. Semua query baca/tulis menyaring `is_deleted = false`. |
| `deleted_at` | `timestamptz` | Diisi saat soft delete. |
| `deleted_by` | `bigint` | User penghapus. |
| `created_at` | `timestamptz` `[default: now()]` | UTC. |
| `created_by` | `bigint` | User pembuat. |
| `modified_at` | `timestamptz` | UTC, diisi saat update. |
| `modified_by` | `bigint` | User pengubah. |

**Relasi (`.dbml`):**
- `mst_inorga.logo_file_id     > mst_file.id` (SET NULL)
- `mst_inorga.banner_file_id   > mst_file.id` (SET NULL)
- `mst_inorga.file_sk_file_id  > mst_file.id` (SET NULL)
- `jadwal.inorga_id            > mst_inorga.id` (dipakai modul jadwal, Fase 3)

**Kategori berkas (`mst_file.kategori`)** yang dipakai saat unggah: logo →
`logo`, banner → `banner`, SK → `dokumen`. `reff_type = 'mst_inorga'`,
`reff_id = <id inorga>` (di-set oleh file layer / backfill).

**Tak ada** kolom `is_aktif` di tabel — status "aktif/berjalan" **diturunkan dari
tanggal** (`tanggal_mulai <= today AND (tanggal_selesai IS NULL OR
tanggal_selesai >= today)`), jangan menambah kolom baru.

**Enum:** tidak ada enum khusus pada tabel ini.

---

## 3. Endpoint — standard REST `/inorga` (rancangan Bab 9 + `_shared/api-endpoints.md` §13)

Semua di bawah `/api/v1`, `Auth: bearer` (`middleware.JWTAuth`), tiap route
dijaga `RequirePermission("inorga.<aksi>")`. **Tidak ada route publik.**

| Method | Path | Permission | Auth | Fungsi |
|---|---|---|---|---|
| GET | `/inorga` | `inorga.read` | bearer | List + filter + paginate. |
| GET | `/inorga/{id}` | `inorga.read` | bearer | Satu periode (detail + uuid berkas). |
| POST | `/inorga` | `inorga.create` | bearer | Buat periode baru. |
| PUT | `/inorga/{id}` | `inorga.update` | bearer | Ubah periode. |
| DELETE | `/inorga/{id}` | `inorga.delete` | bearer | **Soft delete**. |

`cakupan` untuk semua aksi inorga = **`semua`** (data inorga tak berdimensi
kepemilikan per-anggota/instansi).

### 3.1 Query list — `GET /inorga`

`?page=1&per_page=20&q=<cari nama/kode>&sort=-created_at&aktif=<true|false>`

- `q` → ILIKE pada `nama` dan `kode`.
- `aktif=true` → hanya periode berjalan (rumus tanggal di §2). `aktif=false` →
  hanya yang sudah lewat / belum mulai. Kosong → semua.
- `sort` whitelist: `created_at`, `tanggal_mulai`, `nama` (prefiks `-` = DESC).
  Kolom di luar whitelist ditolak (jangan interpolasi ke SQL).

**Response** (`data` = halaman):
```json
{
  "success": true,
  "message": "OK",
  "data": {
    "items": [
      {
        "id": 3,
        "kode": "PENGURUS-2024",
        "nama": "Kepengurusan 2024-2026",
        "tanggal_mulai": "2024-01-01",
        "tanggal_selesai": "2026-01-01",
        "aktif": true,
        "logo": { "id": 12, "uuid": "6f...", "nama_asli": "logo.png" },
        "banner": { "id": 13, "uuid": "7a...", "nama_asli": "banner.jpg" },
        "file_sk": { "id": 14, "uuid": "8b...", "nama_asli": "sk.pdf" },
        "created_at": "2024-01-02T03:04:05Z"
      }
    ],
    "page": 1, "per_page": 20, "total": 1, "last_page": 1
  }
}
```

### 3.2 `GET /inorga/{id}`

Sama seperti item list **plus** `konten` (teks penuh). 404 (`response.NotFound`)
bila tidak ada / sudah soft-deleted.

### 3.3 `POST /inorga` — payload

```json
{
  "kode": "PENGURUS-2024",
  "nama": "Kepengurusan 2024-2026",
  "tanggal_mulai": "2024-01-01",
  "tanggal_selesai": "2026-01-01",
  "logo_file_id": 12,
  "banner_file_id": 13,
  "file_sk_file_id": 14,
  "konten": "<h3>Ketua ...</h3>"
}
```
Response `201` = detail baris baru (bentuk §3.2). Berkas **sudah** diunggah
lebih dulu via `POST /files`; di sini hanya id numerik yang dikirim.

### 3.4 `PUT /inorga/{id}` — payload sama seperti POST (semua field replaceable). Response `200` = detail terbaru.

### 3.5 `DELETE /inorga/{id}` — soft delete. Response `200 { success:true }`. Tidak menghapus berkas terkait (berkas dikelola modul file, `file.delete`).

---

## 4. Hak akses (matriks) — rancangan Bab 3.5

Baris `inorga` pada matriks **mentah** rancangan Bab 3.5 (`.rancangan.txt` baris
237–242) dan `_shared/api-endpoints.md` (baris 374):

| Modul (`kode`) | Grup | Super Admin | Admin | Moderator | User | Guest |
|---|---|---|---|---|---|---|
| `inorga` | Master Data | CRUD | CRUD | **R** | **R** | **R** |

`cakupan` semua baris = `semua`.

Permission yang di-seed (Fase 1) — `mst_permission.kode`:
`inorga.read`, `inorga.create`, `inorga.update`, `inorga.delete`.

`role_permission` (SEED, runtime-editable Super Admin):
- Super Admin (`is_super`) → **bypass** semua cek (baris seed hanya deskriptif).
- Admin → read + create + update + delete (`semua`).
- Moderator → read (`semua`).
- User → read (`semua`).
- Guest → read (`semua`).

> **⚠️ Ketidaksesuaian sumber (diselesaikan):** `_shared/permission-matrix.md`
> Tabel 1 mencantumkan `inorga` Guest = `-`, sedangkan **rancangan Bab 3.5
> mentah** dan `_shared/api-endpoints.md` mencantumkan Guest = **R**. Aturan
> proyek: matriks otoritatif = **rancangan Bab 3.5** → **Guest = R**. Namun
> peta endpoint Bab 9 **tidak** mendefinisikan route publik `/public/inorga`,
> sehingga izin `inorga.read` untuk Guest **saat ini bersifat seed pasif** (tak
> ada route tanpa-auth yang memeriksanya di Fase 2). Bila di Fase 8 kepengurusan
> diekspos di `/public/profil`, izin ini menjadi hidup. Catat perbedaan ini dan
> selaraskan `permission-matrix.md`.

`RequirePermission` di backend adalah otoritasnya — **menyembunyikan tombol di UI
bukan keamanan**.

---

## 5. Kebutuhan BACKEND (Go) — `internal/modules/core/inorga/`

Pola modul standar (lihat `_shared/conventions-api.md` §1). Alur:
**router → JWTAuth → RequirePermission → handler → service → repository → GORM**.

### 5.1 `domain/inorga.go`
- Struct `Inorga` map ke tabel `mst_inorga`; `TableName() string { return "mst_inorga" }`.
- Kolom: `ID int64`, `Kode *string`, `Nama string`, `LogoFileID *int64`,
  `BannerFileID *int64`, `TanggalMulai *time.Time` (date), `TanggalSelesai
  *time.Time` (date), `FileSKFileID *int64`, `Konten *string`, plus embed
  `Audit` (created/modified/soft-delete, sesuai conventions §6).
- **Belongs-to** ke proyeksi baca-saja `FileRef` (id, uuid, nama_asli;
  `TableName() = "mst_file"`) untuk `LogoFile`, `BannerFile`, `SKFile` — di-
  `Preload` saat detail & list agar response berisi `uuid` untuk render gambar.
  Jangan impor domain modul file (hindari coupling); cukup proyeksi ringan.

### 5.2 `dto/inorga.go`
- `ListQuery { Page, PerPage, Q, Sort string; Aktif *bool }` (tag `form`).
- `CreateInorgaReq`:
  - `Kode *string        json:"kode"            binding:"omitempty,max=50"`
  - `Nama string         json:"nama"            binding:"required,max=150"`
  - `TanggalMulai *string json:"tanggal_mulai"  binding:"omitempty,datetime=2006-01-02"`
  - `TanggalSelesai *string json:"tanggal_selesai" binding:"omitempty,datetime=2006-01-02"`
  - `LogoFileID *int64   json:"logo_file_id"    binding:"omitempty"`
  - `BannerFileID *int64 json:"banner_file_id"  binding:"omitempty"`
  - `FileSKFileID *int64 json:"file_sk_file_id" binding:"omitempty"`
  - `Konten *string      json:"konten"          binding:"omitempty"`
- `UpdateInorgaReq` = sama (semua field replaceable).
- `InorgaResponse` / `InorgaListItem` + `FileRefDTO { ID int64; UUID string; NamaAsli string }` + `Aktif bool` (computed). Bungkus list dgn `Paginated[InorgaListItem]`.

### 5.3 `repository/inorga.go`
- `List(ctx, ListQuery) ([]Inorga, int64, error)` — scope `is_deleted=false`;
  `q` → `WHERE (nama ILIKE ? OR kode ILIKE ?)`; filter `aktif` dgn rumus tanggal;
  `Count` lalu `Preload("LogoFile").Preload("BannerFile").Preload("SKFile")` +
  paged `Find`; sort whitelist.
- `FindByID(ctx, id) (*Inorga, error)` — scope `is_deleted=false`, preload 3
  berkas; `gorm.ErrRecordNotFound` → `ErrNotFound`.
- `Create(ctx, *Inorga) error`, `Update(ctx, *Inorga) error`.
- `SoftDelete(ctx, id, actor int64) error` — set `is_deleted=true, deleted_at=now(), deleted_by=actor` (jangan hard delete).
- `FileExists(ctx, ids ...int64) (bool, error)` — cek `mst_file` (is_deleted=false) untuk validasi id berkas.

### 5.4 `service/inorga.go`
Aturan bisnis & validasi (server otoritatif):
- Parse `tanggal_mulai`/`tanggal_selesai` (`2006-01-02`). Bila keduanya ada →
  **`tanggal_selesai >= tanggal_mulai`**, else `ErrValidation`.
- Untuk tiap `*_file_id` non-nil → cek exist via `FileExists` → `ErrValidation`
  ("berkas tidak ditemukan") agar tak bocor FK error mentah 500.
- Map DTO → entity (jangan bind langsung ke domain). Set `created_by`/`modified_by`
  dari `middleware.Claims(c).UserID`.
- `Aktif` dihitung dari tanggal + `time.Now().UTC()`.
- Tulis **`log_aktivitas`** untuk create/update/delete: `modul='inorga'`,
  `aksi` ∈ `buat|ubah|hapus`, `reff_type='mst_inorga'`, `reff_id=id`,
  `ringkasan` (mis. "Membuat kepengurusan 2024-2026"), `nilai_lama`/`nilai_baru`
  (jsonb) untuk update/hapus, `ip_address`/`user_agent` dari request.
- Sentinel errors: `ErrNotFound`, `ErrValidation` (map via `response.FromError`).
- **Tak ada** cakupan scoping (semua `semua`); tak ada background job.

### 5.5 `handler/inorga.go`
- `List`, `Detail`, `Create`, `Update`, `Delete`. Bind → panggil service →
  envelope (`response.OK/Created/...`). `ShouldBindJSON` untuk body; `ShouldBind`
  (query) untuk `ListQuery`. Validasi gagal → `response.Unprocess(c, msg,
  validator.Explain(err))`.

### 5.6 `main.inorga.go`
```go
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard) *Module
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
    g := rg.Group("/inorga", middleware.JWTAuth(m.jwtMgr))
    g.GET("",        m.perm.Require("inorga.read"),   m.h.List)
    g.GET("/:id",    m.perm.Require("inorga.read"),   m.h.Detail)
    g.POST("",       m.perm.Require("inorga.create"), m.h.Create)
    g.PUT("/:id",    m.perm.Require("inorga.update"), m.h.Update)
    g.DELETE("/:id", m.perm.Require("inorga.delete"), m.h.Delete)
}
```
Daftarkan di `internal/router/router.go`:
`inorga.Initialize(db, jwtMgr, permGuard).SetupRoutes(apiV1)`.

### 5.7 Migrasi & seeder
- **Migrasi** `NNNN_mst_inorga.{up,down}.sql` (nomor bebas berikutnya di
  `migrations/`, blok Fase 2, mis. `0015`). `up` = `CREATE TABLE mst_inorga`
  persis kolom `.dbml`, FK 3 berkas `REFERENCES mst_file(id) ON DELETE SET
  NULL`, `nama NOT NULL` (hardening), `CHECK (tanggal_selesai IS NULL OR
  tanggal_mulai IS NULL OR tanggal_selesai >= tanggal_mulai)` (analog Bab 8.4
  "berlaku_sampai > berlaku_dari"), index parsial
  `(tanggal_mulai, tanggal_selesai) WHERE is_deleted=false`. `down` = `DROP TABLE`.
  **Bukan `AutoMigrate`.**
- **Seeder RBAC** untuk inorga (`mst_modul` kode `inorga`, 4 `mst_permission`,
  `role_permission` band §4) **sudah** di-seed Fase 1 secara terpusat &
  idempotent. Modul ini **tidak** menduplikasi; hanya memverifikasi keberadaannya.
  Bila dijalankan standalone sebelum seed Fase 1, sediakan seeder idempotent
  (`INSERT ... ON CONFLICT (kode) DO NOTHING`) yang mencerminkan band di §4
  (ingat **Guest = R**).

### 5.8 Edge cases
- id berkas menunjuk baris terhapus/absen → `ErrValidation` (bukan 500).
- Update dgn `*_file_id` diganti → id lama tetap di `mst_file` (dibersihkan lewat
  `file.delete`, bukan di sini).
- DELETE saat ada `jadwal.inorga_id` merujuk → tetap soft delete (aman; FK
  RESTRICT hanya relevan untuk hard delete yang tak dilakukan).
- `tanggal_selesai` null = periode berjalan → `aktif` bisa true tanpa akhir.

---

## 6. Kebutuhan FRONTEND (Angular) — `pages/inorga/`

Ikuti `_shared/conventions-app.md` (standalone, signals, zoneless, `@if/@for`,
lazy route, `ApiService` + envelope, `*hasPermission`, `permissionGuard`).

### 6.1 Berkas
```
pages/inorga/
  inorga-list.ts / .html / .scss     # tabel + cari + paginate (pola §9 conventions-app)
  inorga-form.ts / .html             # create + edit (reactive form)
  inorga-detail.ts / .html           # tampilan baca (banner, logo, periode, konten, unduh SK)
  inorga.service.ts                  # wrapper ApiService (paths /inorga)
  inorga.model.ts                    # interface Inorga, InorgaForm, FileRef
core/models/…                        # reuse Page<T>, ApiResponse<T>
```

### 6.2 `inorga.service.ts`
```ts
list(q: InorgaQuery)  { return this.api.get<Page<Inorga>>('/inorga', q); }
detail(id: number)    { return this.api.get<Inorga>(`/inorga/${id}`); }
create(b: InorgaForm) { return this.api.post<Inorga>('/inorga', b); }
update(id, b)         { return this.api.put<Inorga>(`/inorga/${id}`, b); }
remove(id: number)    { return this.api.delete<void>(`/inorga/${id}`); }
```
Untuk gambar/berkas pakai `FileService` bersama (`upload(field,file,{kategori})`,
`imageUrl(uuid,varian)`) — jangan implementasi HTTP file sendiri.

### 6.3 Form (mirror DTO)
Typed Reactive Form (`fb.nonNullable`):
| Field form | Validator klien | → payload key | → kolom DB |
|---|---|---|---|
| `kode` | `maxLength(50)` | `kode` | `kode` |
| `nama` | `required, maxLength(150)` | `nama` | `nama` |
| `tanggal_mulai` | `<input type="date">` (opsional) | `tanggal_mulai` | `tanggal_mulai` |
| `tanggal_selesai` | opsional; cross-check `>= tanggal_mulai` | `tanggal_selesai` | `tanggal_selesai` |
| `konten` | textarea, opsional | `konten` | `konten` |
| logo (file) | image, opsional | `logo_file_id` (dari upload) | `logo_file_id` |
| banner (file) | image, opsional | `banner_file_id` | `banner_file_id` |
| SK (file) | pdf/doc/image, opsional | `file_sk_file_id` | `file_sk_file_id` |

- **Native** `<input type="date">` (rung 4 — tanpa lib date-picker).
- File: `FileService.upload(...)` per berkas → simpan `data.id` numerik ke kontrol
  hidden `*_file_id`; preview via `FileService.imageUrl(uuid,'medium')`, **revoke**
  objectURL saat destroy. `FormData` tanpa `Content-Type` manual (browser set
  boundary). `kategori`: logo→`logo`, banner→`banner`, SK→`dokumen`.
- Pada `422`/`400`: map `res.errors` (field→pesan) ke kontrol
  (`setErrors({server: msg})`).

### 6.4 List/detail/tabel
- Tabel kolom: Kode · Nama · Periode (`tanggal_mulai`–`tanggal_selesai`) · badge
  **Aktif** (hijau `--slam-success`) / **Arsip** (muted) dari flag `aktif` (jangan
  hanya warna — sertakan label). Cari (`q`) + paginate signal (`toSignal` +
  `debounceTime` + `switchMap`, pola §9).
- Detail: banner di atas, logo, periode, `konten` (bila HTML → `safe` pipe
  ter-sanitasi), tautan unduh SK via blob `FileService`.

### 6.5 Guard & permission-gating
- Route (lazy) di `app.routes.ts` di bawah VerticalLayout:
  `{ path: 'inorga', canActivate:[authGuard, permissionGuard],
     data:{ permission:'inorga.read' }, loadComponent: … }` + child `new`/`:id`/`:id/edit`.
- Tombol **Tambah** → `*hasPermission="'inorga.create'"`; **Edit** →
  `'inorga.update'`; **Hapus** → `'inorga.delete'`. Semua tombol punya pasangan
  cek handler Go (§5.6).

### 6.6 i18n (`public/i18n/{IND,ENG}.json`, namespace `INORGA.*`)
`INORGA.TITLE`, `INORGA.ADD`, `INORGA.FORM.KODE`, `INORGA.FORM.NAMA`,
`INORGA.FORM.TANGGAL_MULAI`, `INORGA.FORM.TANGGAL_SELESAI`,
`INORGA.FORM.KONTEN`, `INORGA.FORM.LOGO`, `INORGA.FORM.BANNER`,
`INORGA.FORM.SK`, `INORGA.STATUS.AKTIF`, `INORGA.STATUS.ARSIP`,
`INORGA.PERIODE`, plus reuse `COMMON.*` / `VALIDATION.*`. IND default; jaga
kedua file sinkron.

### 6.7 Desain (WAJIB)
- Awali dengan **"Using design-taste-frontend"** → pre-flight/audit → map ke
  **Bootstrap 5** → enforce token di `_shared/design-tokens.md` (dark default,
  SLAM red accent, heading UPPERCASE wide-tracking, focus ring merah, min tap
  44px). Tanpa hex ad-hoc, tanpa font baru.
- **Catatan blog-fe:** *"Jika sumber `blog-fe` dipulihkan, tiru layout/menunya
  untuk layar ini; jika tidak, ikuti design tokens."* (`blog-fe` sekarang kosong.)

---

## 7. ALUR form → API → DATABASE (kontrak koherensi)

### 7.1 Pemetaan field (create/update)
| Field form (Angular) | Request payload key (JSON) | Kolom DB (`mst_inorga`) | Tipe |
|---|---|---|---|
| `kode` | `kode` | `kode` | varchar(50), null-ok |
| `nama` | `nama` | `nama` | varchar(150), wajib |
| `tanggal_mulai` (`<input date>`) | `tanggal_mulai` (`"YYYY-MM-DD"`) | `tanggal_mulai` | date, null-ok |
| `tanggal_selesai` | `tanggal_selesai` (`"YYYY-MM-DD"`) | `tanggal_selesai` | date, null-ok |
| upload logo → `data.id` | `logo_file_id` | `logo_file_id` | bigint FK, null-ok |
| upload banner → `data.id` | `banner_file_id` | `banner_file_id` | bigint FK, null-ok |
| upload SK → `data.id` | `file_sk_file_id` | `file_sk_file_id` | bigint FK, null-ok |
| `konten` (textarea) | `konten` | `konten` | text, null-ok |
| — (server) | — | `created_by/modified_by` | dari JWT claims |
| — (server) | — | `created_at/modified_at` | `now()` UTC |

### 7.2 Aksi → endpoint → tulisan DB
| Aksi user | Endpoint | Tulisan DB |
|---|---|---|
| Unggah logo/banner/SK | `POST /files` (multipart, `kategori=logo/banner/dokumen`, `reff_type=mst_inorga`) | INSERT `mst_file` (+`mst_file_varian` untuk gambar) → balikan `{id, uuid, variants}` |
| Simpan periode baru | `POST /inorga` (JSON body §3.3) | INSERT `mst_inorga` (file_id dari langkah di atas) + INSERT `log_aktivitas` (`aksi=buat`) |
| Ubah periode | `PUT /inorga/{id}` | UPDATE `mst_inorga` (`modified_at/by`) + INSERT `log_aktivitas` (`aksi=ubah`, nilai_lama/baru) |
| Hapus periode | `DELETE /inorga/{id}` | UPDATE `mst_inorga` SET `is_deleted=true, deleted_at, deleted_by` + INSERT `log_aktivitas` (`aksi=hapus`) |
| Lihat daftar | `GET /inorga?...` | SELECT (scope `is_deleted=false`) + Preload `mst_file` (uuid untuk gambar) |
| Render gambar | `GET /files/{uuid}/{varian}` | (baca) stream berkas privat ber-auth |

**Kontrak:** nama payload key = nama kolom DB (kecuali file_id yang berasal dari
langkah upload terpisah). Tak ada penggantian nama di tengah jalan.

---

## 8. Dependencies / prasyarat & Acceptance criteria

### 8.1 Prasyarat
- **Fase 0:** file layer (`mst_file`, `mst_file_varian`, `POST /files`,
  `GET /files/{uuid}/{varian}`, `DELETE /files/{uuid}`), `log_aktivitas`,
  `mst_pengaturan` (ukuran varian), konvensi timestamptz + `_ "time/tzdata"`.
- **Fase 1:** RBAC (`PermGuard.Require`, seed `mst_modul` `inorga`,
  `mst_permission` `inorga.{read,create,update,delete}`, `role_permission` band
  §4), Angular `permissionGuard` + `*hasPermission` + sidebar dinamis.
- Envelope `response` (+ `Forbidden`/`NotFound`), `validator.Explain`, `ApiService`.

### 8.2 Acceptance criteria (checkable)
- [ ] Migrasi `mst_inorga` naik/turun bersih; kolom & FK persis `.dbml`
  (3× `*_file_id` SET NULL), `nama NOT NULL`, CHECK periode, index parsial.
- [ ] `go build ./...` & `go vet ./...` bersih; modul terdaftar di `router.go`.
- [ ] 5 route `/inorga` merespons; tiap route menolak tanpa izin (**403** untuk
  Moderator/User pada create/update/delete; **200** untuk read semua peran login).
- [ ] Super Admin bypass; Guest R hanya seed (tak ada route publik) — didokumentasikan.
- [ ] `POST /inorga` menyimpan baris; `logo/banner/file_sk_file_id` tersimpan &
  `GET /inorga/{id}` mengembalikan `uuid` berkas; gambar tampil via
  `GET /files/{uuid}/{varian}`.
- [ ] `tanggal_selesai < tanggal_mulai` → **422**; `*_file_id` tak ada → **422**
  (bukan 500).
- [ ] `PUT` mengubah baris + `modified_at/by`; `DELETE` set `is_deleted=true`
  (baris tetap ada); keduanya menulis `log_aktivitas`.
- [ ] List: `q`, `aktif`, paginate, sort whitelist bekerja; hanya
  `is_deleted=false`.
- [ ] Frontend: form submit → baris DB dibuat/diubah; tombol Tambah/Edit/Hapus
  tersembunyi **dan** terblokir tanpa izin; responsif; token desain diterapkan;
  i18n IND/ENG sinkron.
