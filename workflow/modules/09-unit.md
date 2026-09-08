# Modul 09 — Master Data Unit (airsoft gun) + Approval

> Sumber otoritatif: `slamteam_db.dbml` (tabel `unit`) > rancangan Bab 3.5 / 9 / 12 >
> `_shared/*` > file ini > app lama (tidak pernah). Bila prosa dan `.dbml` berbeda, **`.dbml` menang**.
> Nama route, key payload, dan nama kolom DB ditulis **verbatim** sesuai `.dbml` + peta endpoint.

---

## 1. Ringkasan & tujuan modul

**Unit** adalah registri senjata airsoft milik anggota. Setiap baris `unit`
mendata satu pucuk (kode inventaris, model, dimensi/berat/FPS, warna, foto sampul)
dan **melekat pada satu anggota** (`unit.anggota_id`). Modul ini juga memuat
**alur persetujuan** ringan: tiga kolom `disetujui` / `disetujui_oleh` / `disetujui_pada`
menandai apakah data unit sudah diverifikasi moderator/admin.

- **Posisi di 19 modul:** modul ke-**09**, `kode = unit`, grup **Master Data**.
- **Fase:** **Fase 2** (Anggota + master data). Prasyarat: **Fase 1** (RBAC middleware/guard) sudah ada,
  serta modul **anggota** dan **file layer Fase 0**.
- **Pola:** master-data CRUD standar (pola berulang, rancangan Bab 12 baris "Anggota dan master data").
  Approval **tidak** memiliki endpoint terpisah — ia diset lewat `PUT /unit/{id}` (lihat §3, §5).
- **Izin:** `unit` = CRUD untuk **Super Admin / Admin / Moderator / User** (semua role login boleh kelola unit;
  rancangan Bab 3.5: "unit memberi CRUD ke semua role login").

Kegunaan nyata: pendataan senjata per anggota untuk keperluan safety/legalitas klub,
tampilan pada profil anggota, dan verifikasi (approval) oleh pengurus.

---

## 2. Tabel & kolom (`unit`)

Sumber: `slamteam_db.dbml` `Table unit` (grup `7 · Konten & Master`). Kolom **persis**:

| Kolom | Tipe | Aturan / catatan |
|---|---|---|
| `id` | `bigint` | PK identity (`increment`). Dipakai di route admin `/unit/{id}`. |
| `anggota_id` | `bigint` | **NOT NULL**. FK → `anggota.id`. Pemilik unit. |
| `kode` | `varchar(50)` | Kode inventaris unit (opsional, bebas). |
| `model` | `varchar(100)` | Nama/model senjata (mis. "M4A1", "AK-74U"). |
| `panjang` | `decimal(8,2)` | Panjang total. |
| `panjang_inbar` | `decimal(8,2)` | Panjang inner barrel. |
| `lebar` | `decimal(8,2)` | Lebar. |
| `berat` | `decimal(8,2)` | Berat unit. |
| `berat_bb` | `decimal(8,3)` | Berat BB (presisi 3 desimal). |
| `fps` | `decimal(8,2)` | Feet per second (muzzle velocity). |
| `deskripsi_warna` | `varchar(150)` | Deskripsi warna/livery. |
| `foto_sampul_file_id` | `bigint` | FK → `mst_file.id` (nullable). Foto sampul unit. **Jangan simpan path/URL** — resolve lewat file layer. |
| `disetujui` | `boolean` | `default false`. Status approval. |
| `disetujui_oleh` | `bigint` | FK → `users.id`. Diisi **server-side** saat transisi `false → true`. |
| `disetujui_pada` | `timestamptz` | Diisi **server-side** (UTC) saat transisi `false → true`. |
| `is_deleted` | `boolean` | `default false`. Soft delete. |
| `deleted_at` | `timestamptz` | Diisi saat soft delete. |
| `deleted_by` | `bigint` | FK → `users.id`, aktor penghapus. |
| `created_at` | `timestamptz` | `default now()` (UTC). |
| `created_by` | `bigint` | FK → `users.id`, aktor pembuat. |
| `modified_at` | `timestamptz` | Diisi saat update. |
| `modified_by` | `bigint` | FK → `users.id`, aktor pengubah. |

**Relasi (`.dbml`):**
- `Ref: unit.anggota_id > anggota.id`
- `Ref: unit.foto_sampul_file_id > mst_file.id`
- `Ref: unit.disetujui_oleh > users.id`

**Aturan lintas-modul yang berlaku:**
- **Soft delete** eksplisit (`is_deleted + deleted_at + deleted_by`), **bukan** `gorm.DeletedAt`. Semua query list/read wajib `WHERE is_deleted = false`.
- **Time** = `timestamptz` UTC. `created_at` default `now()` di DB.
- **File** disimpan sebagai `foto_sampul_file_id bigint` → `mst_file.id`; publik/URL diresolusi lewat file layer, bukan disimpan mentah.
- **Index yang perlu dibuat migrasi** (tidak ada di `.dbml` secara eksplisit, tapi wajib untuk performa & scoping): index pada `anggota_id` dan pada `is_deleted`. Tidak ada unique di tabel ini.

---

## 3. Endpoint (REST CRUD standar, semua di bawah `/api/v1`)

Auth = **Bearer JWT**. Setiap route dijaga `RequirePermission("unit.<aksi>")` di middleware Go.
`cakupan` (`semua` / `milik_sendiri`) diresolusi di handler/service (lihat §4, §5).

| Method | Path | Permission | Auth | Fungsi |
|---|---|---|---|---|
| GET | `/unit` | `unit.read` | Bearer | List + filter + paginate. |
| GET | `/unit/{id}` | `unit.read` | Bearer | Detail satu unit. |
| POST | `/unit` | `unit.create` | Bearer | Buat unit. |
| PUT | `/unit/{id}` | `unit.update` | Bearer | Ubah unit **dan** set/lepas approval (`disetujui`). |
| DELETE | `/unit/{id}` | `unit.delete` | Bearer | Soft delete. |

> Tidak ada `/unit/{id}/approve` — peta endpoint (Bab 9) menetapkan **"standard CRUD /unit"**. Approval naik lewat field `disetujui` pada `PUT`.
> `ponytail:` approval lewat field di PUT (tanpa endpoint/izin `unit.approve` terpisah); upgrade path — bila approval perlu izin sendiri, tambah permission `unit.approve` + gate transisi `disetujui` di service, tanpa mengubah endpoint.

### 3.1 Payload — `POST /unit` (request JSON)

```json
{
  "anggota_id": 12,
  "kode": "SLAM-U-001",
  "model": "M4A1 CQB",
  "panjang": 690.00,
  "panjang_inbar": 260.00,
  "lebar": 60.00,
  "berat": 2750.00,
  "berat_bb": 0.250,
  "fps": 330.00,
  "deskripsi_warna": "Black / tan handguard",
  "foto_sampul_uuid": "8f2c1d94-...-uuid-dari-POST-/files"
}
```

- `disetujui` **tidak diterima** pada create (selalu `false`) — anti self-approve saat buat.
- `foto_sampul_uuid` = `mst_file.uuid` hasil `POST /files`; service me-resolve ke `foto_sampul_file_id` (bigint). Nullable (boleh tanpa foto).

### 3.2 Payload — `PUT /unit/{id}` (request JSON)

Sama seperti create + boleh menyertakan `disetujui`:

```json
{
  "kode": "SLAM-U-001",
  "model": "M4A1 CQB v2",
  "panjang": 700.00,
  "panjang_inbar": 300.00,
  "lebar": 60.00,
  "berat": 2800.00,
  "berat_bb": 0.280,
  "fps": 345.00,
  "deskripsi_warna": "Black",
  "foto_sampul_uuid": "8f2c1d94-...",
  "disetujui": true
}
```

`anggota_id` **tidak boleh dipindah** lewat update (unit tidak berpindah pemilik) — abaikan bila dikirim.

### 3.3 Response shape (`data`)

Satu unit (dipakai detail, create, update):

```json
{
  "id": 34,
  "anggota_id": 12,
  "anggota_nama": "Budi Santoso",
  "anggota_no_induk": "35730102021",
  "kode": "SLAM-U-001",
  "model": "M4A1 CQB",
  "panjang": 690.00,
  "panjang_inbar": 260.00,
  "lebar": 60.00,
  "berat": 2750.00,
  "berat_bb": 0.250,
  "fps": 330.00,
  "deskripsi_warna": "Black / tan handguard",
  "foto_sampul": { "uuid": "8f2c1d94-...", "url": "/api/v1/files/8f2c1d94-.../medium" },
  "disetujui": false,
  "disetujui_oleh": null,
  "disetujui_oleh_nama": null,
  "disetujui_pada": null,
  "created_at": "2026-09-08T04:12:00Z",
  "modified_at": null
}
```

List (`GET /unit`) mengembalikan `data` berbentuk halaman:

```json
{ "success": true, "message": "ok",
  "data": { "items": [ { ...unit... } ], "page": 1, "per_page": 20, "total": 42, "last_page": 3 } }
```

**Query list:** `?page=1&per_page=20&q=<kode|model|deskripsi_warna>&anggota_id=<id>&disetujui=<true|false>&sort=-created_at`.

Envelope selalu `{ success, message, data?, errors? }`.

---

## 4. Hak akses (baris matriks modul ini)

Sumber: rancangan Bab 3.5 / `_shared/permission-matrix.md`.

| Modul (`kode`) | Grup | Super Admin | Admin | Moderator | User | Guest |
|---|---|---|---|---|---|---|
| `unit` | Master | CRUD | CRUD | CRUD | **CRUD** | — |

- **Cakupan (seed):** tidak disebut khusus ⇒ default **`semua`** untuk keempat role login. Guest = tidak ada akses.
- Permission codes yang di-seed pada Fase 1: `unit.read`, `unit.create`, `unit.update`, `unit.delete` (row `role_permission` untuk SA/Admin/Mod/User dengan `cakupan = 'semua'`).
- **Super Admin** (`is_super`) BYPASS semua cek.
- **Dimensi kepemilikan tersedia** lewat `unit.anggota_id`. Karena cakupan bisa diedit runtime oleh Super Admin (halaman Hak Akses), service **wajib** mengimplementasikan resolusi cakupan generik agar admin bisa menyempitkan (mis.) `User → milik_sendiri` **tanpa ubah kode**:
  - `semua` → tanpa filter tambahan.
  - `milik_sendiri` → `WHERE anggota_id = <anggota_id aktor>` (dari klaim JWT). Berlaku untuk read **dan** write/delete.
- Menyembunyikan tombol di UI **bukan** keamanan — setiap aksi UI punya cek `RequirePermission` + filter cakupan yang cocok di Go.

---

## 5. Kebutuhan BACKEND (Go) — `internal/modules/core/unit/`

Ikuti pola modul `_shared/conventions-api.md`: `domain → dto → repository → service → handler → main.unit.go`. Handler tidak menyentuh `*gorm.DB`; repository tidak menyentuh `*gin.Context`; semua aturan bisnis di service.

### 5.1 `domain/unit.go`
- Struct `Unit` map ke tabel `unit` (`TableName() = "unit"` bila perlu). Embed `Audit` bersama (`created_at/by`, `modified_at/by`, `is_deleted`, `deleted_at/by`) sesuai `conventions-api.md §6`.
- Kolom desimal opsional → `*float64` (`panjang`, `panjang_inbar`, `lebar`, `berat`, `berat_bb`, `fps`).
- `AnggotaID int64` (not null), `FotoSampulFileID *int64`, `Disetujui bool` (default false), `DisetujuiOleh *int64`, `DisetujuiPada *time.Time`.
- Field navigasi ringan (opsional, untuk preload nama): relasi ke `anggota` (nama_lengkap, no_induk) dan `users` (approver) — atau join di repository.

### 5.2 `dto/unit.go`
- `CreateUnitReq`: `anggota_id` (`required`), `kode` (`omitempty,max=50`), `model` (`omitempty,max=100`), `panjang/panjang_inbar/lebar/berat/berat_bb/fps` (`omitempty,gte=0`), `deskripsi_warna` (`omitempty,max=150`), `foto_sampul_uuid` (`omitempty,uuid4`). **Tidak ada** field `disetujui`.
- `UpdateUnitReq`: sama tanpa `anggota_id`, **tambah** `disetujui *bool` (`omitempty`).
- `ListUnitQuery`: `page (default 1, min 1)`, `per_page (default 20, min 1, max 100)`, `q`, `anggota_id`, `disetujui *bool`, `sort`.
- `UnitResp` + `Paginated[UnitResp]` sesuai §3.3.
- Binding tags mengikuti kolom `.dbml`; validasi bisnis (uniqueness kepemilikan, resolve uuid, transisi approval) di service, bukan tag.

### 5.3 `repository/unit.go`
- `Create`, `Update`, `SoftDelete`, `FindByID`, `List(query, scope)`.
- Semua query filter `is_deleted = false`.
- List: `Count` lalu `Find` berpaginasi pada query yang **sama** (termasuk filter cakupan). Whitelist kolom `sort`: `created_at`, `model`, `kode`, `fps`, `disetujui` (prefix `-` = DESC). Jangan interpolasi `sort` ke SQL.
- `q` → `ILIKE` pada `kode` / `model` / `deskripsi_warna`.
- Join/preload ringan untuk `anggota_nama`, `anggota_no_induk`, `disetujui_oleh_nama`.
- Terjemahkan `gorm.ErrRecordNotFound → ErrNotFound`.

### 5.4 `service/unit.go` — aturan bisnis
1. **Resolusi cakupan** (baca `c.Get("cakupan:unit.<aksi>")` yang diset middleware): `semua` (tanpa filter) vs `milik_sendiri` (`anggota_id = claims.AnggotaID`). Berlaku untuk list, detail, update, delete.
2. **Create:**
   - Bila cakupan `milik_sendiri`, **paksa** `anggota_id = claims.AnggotaID` (abaikan/ tolak nilai lain → `ErrForbidden` bila berbeda). Ini cek trust-boundary, **jangan** dilewati.
   - Validasi `anggota_id` ada & tidak terhapus (repo anggota / `EXISTS`). Bila tidak → `ErrValidation`/`ErrNotFound`.
   - `disetujui` selalu `false`; `disetujui_oleh/pada` `NULL`.
   - Resolve `foto_sampul_uuid` → `mst_file.id` via file layer (`FileByUUID`); simpan `foto_sampul_file_id`. UUID tak dikenal → `ErrValidation`.
   - Set `created_by = claims.UserID`, `created_at = now() UTC` (atau default DB).
3. **Update:**
   - Ambil row (dengan filter cakupan). Tidak ketemu → `ErrNotFound` (juga menutup kasus "milik orang lain" bagi `milik_sendiri`).
   - Abaikan `anggota_id` bila dikirim (tidak pindah pemilik).
   - **Transisi approval** (`disetujui`):
     - `false → true`: set `disetujui = true`, `disetujui_oleh = claims.UserID`, `disetujui_pada = now() UTC`.
     - `true → false`: set `disetujui = false`, `disetujui_oleh = NULL`, `disetujui_pada = NULL`.
     - tidak berubah: jangan sentuh `disetujui_oleh/pada` (idempoten — jangan menimpa approver awal).
   - Resolve `foto_sampul_uuid` bila dikirim; `null` eksplisit → kosongkan `foto_sampul_file_id` (opsional, sesuai kebutuhan UI).
   - `modified_by = claims.UserID`, `modified_at = now() UTC`.
4. **Delete:** soft delete — `is_deleted = true`, `deleted_at = now() UTC`, `deleted_by = claims.UserID`. Jangan hard delete. **Tidak** menghapus baris `mst_file` foto (file layer punya siklus hidup sendiri).
5. **Audit:** tulis `log_aktivitas` untuk create/update/delete dan **khususnya** perubahan `disetujui` (aktor, `modul='unit'`, `aksi`, `reff_type='unit'`, `reff_id=id`, `nilai_lama/nilai_baru` jsonb, ip, user_agent). Sistem/job → `aktor_user_id` null (tidak ada job di modul ini).
6. **Sentinel errors** dipetakan handler via `response.FromError`.

### 5.5 `handler/unit.go`
- `List`, `Get`, `Create`, `Update`, `Delete`. Bind (`ShouldBindJSON` / `ShouldBind` query) → `validator.Explain` pada error → `response.Unprocess`. Panggil service dengan `middleware.Claims(c)`. Sukses → `response.OK` / `response.Created`.

### 5.6 `main.unit.go` + router
```go
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
    g := rg.Group("/unit", middleware.JWTAuth(m.jwtMgr))
    g.GET("",       m.perm.Require("unit.read"),   m.h.List)
    g.GET("/:id",   m.perm.Require("unit.read"),   m.h.Get)
    g.POST("",      m.perm.Require("unit.create"), m.h.Create)
    g.PUT("/:id",   m.perm.Require("unit.update"), m.h.Update)
    g.DELETE("/:id",m.perm.Require("unit.delete"), m.h.Delete)
}
```
Register di `internal/router/router.go`: `unit.Initialize(db, jwtMgr, permGuard).SetupRoutes(apiV1)`.

### 5.7 Migrations & seeders
- Migrasi bernomor (bukan `AutoMigrate`) sesuai `.dbml`: `CREATE TABLE unit (...)` dengan tipe persis, `default false` untuk `disetujui`/`is_deleted`, `default now()` untuk `created_at`, FK ke `anggota`, `mst_file`, `users`. Tambah index `anggota_id` dan `is_deleted`. Sertakan `.down.sql` (`DROP TABLE unit`).
- **Permission seed** (idempoten, di seeder Fase 1 / modul): pastikan baris `mst_permission` untuk `unit.read/create/update/delete` ada (di bawah `mst_modul.kode='unit'`) dan `role_permission` untuk SA/Admin/Mod/User dengan `cakupan='semua'`. `INSERT ... ON CONFLICT DO NOTHING`.
- Tidak ada seed data unit contoh (data operasional).

### 5.8 Edge cases
- `anggota_id` menunjuk anggota terhapus → tolak.
- `milik_sendiri` mencoba baca/ubah/hapus unit anggota lain → `ErrNotFound` (jangan bocorkan keberadaannya).
- Update `disetujui=true` berulang → idempoten, approver awal dipertahankan.
- `foto_sampul_uuid` menunjuk file milik reff lain → boleh, file layer generik; tapi validasi UUID valid & ada.
- Angka desimal negatif → ditolak (`gte=0`).

---

## 6. Kebutuhan FRONTEND (Angular) — `pages/unit/`

Ikuti `_shared/conventions-app.md`: standalone components, signals, zoneless, `@if/@for`, Reactive Forms typed, `ApiService` + envelope, `*hasPermission`, i18n.

### 6.1 Halaman/komponen
- `pages/unit/unit-list.ts|html|scss` — tabel + filter (`q`, `anggota_id`, `disetujui`), paginasi, tombol Tambah (`*hasPermission="'unit.create'"`), aksi Detail/Ubah/Hapus per baris (gated). Pola list kanonik `conventions-app §9` (`toSignal` + `switchMap` + `debounceTime`).
- `pages/unit/unit-form.ts|html|scss` — form create/edit (Reactive Forms). Untuk create-by-user, field `anggota_id` bisa dikunci ke anggota sendiri (mengikuti cakupan yang diberikan backend); untuk Admin/Mod pilih anggota (autocomplete/select dari `/anggota`).
- `pages/unit/unit-detail.ts|html|scss` (opsional; bisa gabung dengan form read-only) — tampilkan spesifikasi + foto sampul + status approval + tombol Setujui/Batalkan setuju (`*hasPermission="'unit.update'"`).
- `pages/unit/unit.service.ts` — wrapper `ApiService` (list/detail/create/update/remove).
- `pages/unit/unit.model.ts` — interface `Unit`, `UnitForm`, `UnitQuery` (mirror DTO).
- Route lazy di `app.routes.ts`: `{ path: 'unit', loadComponent: ... , canActivate: [authGuard, permissionGuard], data: { permission: 'unit.read' } }` + subroute `new` / `:id`.

### 6.2 Form fields (mirror DTO Go)
| Field form | Validator (mirror binding) | Kontrol UI |
|---|---|---|
| `anggota_id` | `required` (create) | select/autocomplete anggota; dikunci bila cakupan `milik_sendiri` |
| `kode` | `maxLength(50)` | text |
| `model` | `maxLength(100)` | text |
| `panjang` | `min(0)` | number |
| `panjang_inbar` | `min(0)` | number |
| `lebar` | `min(0)` | number |
| `berat` | `min(0)` | number |
| `berat_bb` | `min(0)`, step 0.001 | number |
| `fps` | `min(0)` | number |
| `deskripsi_warna` | `maxLength(150)` | text |
| `foto_sampul_uuid` | `omitempty` | file upload → `FileService.upload` → simpan uuid |
| `disetujui` | (edit only) | toggle/switch, hanya render `*hasPermission="'unit.update'"` |

Pada `422/400`, map `res.errors` (field→pesan) ke kontrol via `setErrors({ server: msg })`.

### 6.3 File handling (foto sampul)
- **Upload:** `FileService.upload('foto', file)` → dapat `{ uuid }`; kirim `foto_sampul_uuid` di payload unit. **Jangan** set `Content-Type` manual (browser set boundary).
- **Tampil:** foto privat → ambil blob via `GET /files/{uuid}/{varian}` (`medium` untuk list/detail, `low` untuk thumbnail); bind object URL; **revoke** saat destroy. Jangan `<img src>` token-less.

### 6.4 Permission-gating
- Guard route: `data.permission = 'unit.read'`.
- Tombol/aksi: `*hasPermission="'unit.create'|'unit.update'|'unit.delete'"`. Ingat: gating UI hanya UX; backend tetap otoritatif.

### 6.5 i18n (`public/i18n/IND.json` & `ENG.json`, ter-sinkron)
Namespace `UNIT.*`:
`UNIT.TITLE`, `UNIT.ADD`, `UNIT.LIST.KODE`, `UNIT.LIST.MODEL`, `UNIT.LIST.ANGGOTA`, `UNIT.LIST.FPS`, `UNIT.LIST.STATUS`,
`UNIT.FORM.ANGGOTA`, `UNIT.FORM.KODE`, `UNIT.FORM.MODEL`, `UNIT.FORM.PANJANG`, `UNIT.FORM.PANJANG_INBAR`, `UNIT.FORM.LEBAR`, `UNIT.FORM.BERAT`, `UNIT.FORM.BERAT_BB`, `UNIT.FORM.FPS`, `UNIT.FORM.DESKRIPSI_WARNA`, `UNIT.FORM.FOTO_SAMPUL`,
`UNIT.APPROVAL.STATUS`, `UNIT.APPROVAL.DISETUJUI`, `UNIT.APPROVAL.BELUM`, `UNIT.APPROVAL.SETUJUI`, `UNIT.APPROVAL.BATAL`, `UNIT.APPROVAL.OLEH`, `UNIT.APPROVAL.PADA`,
plus `COMMON.*` (SAVE/CANCEL/DELETE/DETAIL/EDIT/ADD/SEARCH/EMPTY) dan `VALIDATION.*`.

### 6.6 Desain (WAJIB)
- Ikuti **taste-skill** + **design-tokens** (lihat prompt APP §desain). Status approval sebagai `badge` (`text-bg-success` = disetujui, `text-bg-warning`/`text-bg-secondary` = belum) — jangan sampaikan status hanya lewat warna, pasangkan label/ikon.
- **Blog-fe note:** "Jika sumber `blog-fe` dipulihkan, tiru layout/menu-nya untuk layar ini; jika tidak, ikuti design tokens."

---

## 7. ALUR form → API → DATABASE (kontrak koherensi)

### 7.1 Peta field (form → payload → kolom DB)
| Field form (Angular) | Key payload (JSON) | Kolom DB (`unit`) | Catatan |
|---|---|---|---|
| Anggota (select) | `anggota_id` | `anggota_id` | create saja; milik_sendiri dipaksa = anggota aktor |
| Kode | `kode` | `kode` | |
| Model | `model` | `model` | |
| Panjang | `panjang` | `panjang` | decimal(8,2) |
| Panjang inner barrel | `panjang_inbar` | `panjang_inbar` | decimal(8,2) |
| Lebar | `lebar` | `lebar` | decimal(8,2) |
| Berat | `berat` | `berat` | decimal(8,2) |
| Berat BB | `berat_bb` | `berat_bb` | decimal(8,3) |
| FPS | `fps` | `fps` | decimal(8,2) |
| Deskripsi warna | `deskripsi_warna` | `deskripsi_warna` | |
| Foto sampul (upload) | `foto_sampul_uuid` | `foto_sampul_file_id` | service resolve uuid → id (`mst_file`) |
| Toggle setujui (edit) | `disetujui` | `disetujui` (+ `disetujui_oleh`,`disetujui_pada`) | oleh/pada diisi server pada transisi |
| — (server) | — | `created_by/at`, `modified_by/at`, `is_deleted`, `deleted_by/at` | dari klaim + waktu server |

### 7.2 Aksi user → endpoint → tulisan DB
| Aksi user | Endpoint | Tulisan DB |
|---|---|---|
| Buka daftar unit | `GET /unit?...` | (read) `WHERE is_deleted=false` [+ `anggota_id=` bila `milik_sendiri`] |
| Lihat detail | `GET /unit/{id}` | (read) 1 baris (dengan filter cakupan) |
| Simpan unit baru | `POST /unit` | INSERT baris `unit`: kolom form + `foto_sampul_file_id` (resolve) + `disetujui=false` + `created_by/at` |
| Ubah unit | `PUT /unit/{id}` | UPDATE kolom form + `modified_by/at`; `anggota_id` diabaikan |
| Setujui unit | `PUT /unit/{id}` `{disetujui:true}` | UPDATE `disetujui=true, disetujui_oleh=<aktor>, disetujui_pada=now()` + `modified_by/at` + `log_aktivitas` |
| Batalkan setuju | `PUT /unit/{id}` `{disetujui:false}` | UPDATE `disetujui=false, disetujui_oleh=NULL, disetujui_pada=NULL` + audit |
| Hapus unit | `DELETE /unit/{id}` | UPDATE `is_deleted=true, deleted_at=now(), deleted_by=<aktor>` (soft) |

Setiap key payload = nama kolom `.dbml` verbatim (kecuali `foto_sampul_uuid` yang sengaja UUID publik-aman lalu diresolusi ke `foto_sampul_file_id`).

---

## 8. Dependencies / prasyarat & Acceptance criteria

### 8.1 Prasyarat
- **Fase 0:** file layer (`mst_file` + `POST /files` + `GET /files/{uuid}/{varian}`) siap; tersedia cara resolve `uuid → mst_file.id` (`FileByUUID`). `log_aktivitas` & response envelope siap.
- **Fase 1:** RBAC middleware `PermGuard.Require`, seed `mst_modul.kode='unit'` + permission `unit.*`, guard/directive Angular.
- **Modul anggota (Fase 2):** tabel `anggota` + endpoint `/anggota` (untuk memilih pemilik & menampilkan nama/no_induk).

### 8.2 Acceptance criteria (checkable)
- [ ] Migrasi `unit` cocok **persis** dengan `.dbml` (tipe, `default`, FK, soft-delete, index `anggota_id`/`is_deleted`); `.down.sql` menghapus tabel.
- [ ] `GET/POST/PUT/DELETE /unit` terdaftar di router di bawah `/api/v1` dan dijaga `unit.read/create/update/delete`.
- [ ] Tanpa Bearer → 401; punya token tapi tanpa permission → 403; Super Admin bypass.
- [ ] `POST /unit` menyimpan baris dengan `disetujui=false`, `created_by`=aktor; `foto_sampul_uuid` diresolusi ke `foto_sampul_file_id` yang benar.
- [ ] Cakupan: dengan `unit.read=milik_sendiri`, User hanya melihat/ubah/hapus unit `anggota_id`-nya sendiri; unit orang lain → 404.
- [ ] `PUT /unit/{id}` dengan `disetujui:true` mengisi `disetujui_oleh`+`disetujui_pada` (UTC), transisi ke `false` mengosongkannya; idempoten saat tak berubah.
- [ ] `DELETE /unit/{id}` = soft delete; baris tak lagi muncul di list; `mst_file` foto tidak ikut terhapus.
- [ ] Validasi: `anggota_id` required (create), desimal `gte=0`, panjang string sesuai batas kolom; error → 422 `{errors:{field:[msg]}}`.
- [ ] FE: form submit → baris DB tercipta/terupdate; tombol Tambah/Ubah/Hapus/Setujui tersembunyi tanpa permission **dan** ditolak backend bila dipaksa.
- [ ] FE: foto sampul tampil via `/files/{uuid}/medium` (blob), object URL di-revoke saat destroy.
- [ ] FE: i18n IND/ENG sinkron, tanpa string keras; design tokens + taste-skill diterapkan; responsif.
- [ ] `log_aktivitas` tercatat untuk create/update/delete dan perubahan approval.
