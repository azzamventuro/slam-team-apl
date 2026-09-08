# 12 · Master Data Medsos (`medsos`)

> Fase 2 · Grup **Konten** · Tabel `mst_medsos` (link medsos per anggota) · CRUD `/medsos`
> Sumber otoritatif: `slamteam_db.dbml` (skema) > rancangan Bab 9 (rute) > `_shared/*` > app lama (jangan).

---

## 1. Ringkasan & tujuan modul

Modul **Medsos** (nomor 12 dari 19 modul, `kode = medsos`, grup **Konten**, dikerjakan di **Fase 2 — Anggota & Master Data**) mengelola **tautan media sosial milik seorang anggota** (Instagram, TikTok, YouTube, Facebook, WhatsApp, dst.). Setiap baris menautkan satu akun medsos ke satu `anggota`. Ini adalah modul CRUD "pola berulang" (sama seperti prestasi/unit/inorga) tanpa logika khusus berat.

Poin pembeda modul ini:

- **`icon` adalah NAMA ikon dari icon-set** (mis. `"instagram"`, `"tiktok"`), **BUKAN berkas unggahan**. Tidak ada `*_file_id`, tidak menyentuh file-layer sama sekali. (`slamteam_db.dbml` baris 1127, catatan kolom.)
- Data selalu **milik seorang anggota** (`anggota_id` NOT NULL) → modul ini bergantung pada modul **anggota** (Fase 2).
- Hak akses: **SA/Admin/Moderator/User CRUD, Guest R** — User boleh mengelola medsos **miliknya sendiri** (cakupan `milik_sendiri`); Guest hanya baca (disurface lewat profil publik, bukan rute terpisah).

Tujuan: menyimpan dan menampilkan tautan medsos anggota untuk profil (internal admin panel dan, sebagai bahan baca, profil publik).

---

## 2. Tabel & kolom (`mst_medsos`)

Definisi persis dari `slamteam_db.dbml` (baris 1122–1137). Tidak ada enum PG khusus untuk tabel ini; `tipe` adalah `int` biasa, sisanya `varchar`.

| Kolom | Tipe (DBML) | Aturan / catatan |
|---|---|---|
| `id` | `bigint` `[pk, increment]` | PK identity. Dipakai di rute admin `/medsos/{id}`. |
| `anggota_id` | `bigint` `[not null]` | **Wajib.** FK → `anggota.id` (`Ref: mst_medsos.anggota_id > anggota.id`, baris 1238). Pemilik baris. |
| `kode` | `varchar(50)` | Kode/slug opsional (mis. `ig-main`). Boleh null. |
| `tipe` | `int` | Kode kategori numerik opsional (klasifikasi/urutan tampil). Tanpa enum di DBML → perlakukan sebagai int bebas ≥ 0, default `0`. Jangan mengarang semantik keras. |
| `icon` | `varchar(50)` | **Nama ikon dari icon-set** (mis. `"instagram"`). BUKAN berkas. Catatan DBML: *'Nama ikon dari icon set ("instagram") — BUKAN berkas unggahan'*. |
| `jenis_medsos` | `varchar(50)` | Nama platform (mis. `instagram`, `facebook`, `tiktok`, `youtube`, `whatsapp`). Wajib diisi (validasi service, bukan constraint DB). |
| `konten_medsos` | `varchar(255)` | Isi tautan: URL profil / handle / nomor. Wajib diisi. |
| `is_deleted` | `boolean` `[default: false]` | Soft delete. Semua query list/read menyaring `is_deleted = false`. |
| `deleted_at` | `timestamptz` | Diisi saat soft delete. |
| `deleted_by` | `bigint` | User yang menghapus. |
| `created_at` | `timestamptz` `[default: now()]` | UTC. Default DB. |
| `created_by` | `bigint` | Diisi dari `Claims` aktor. |
| `modified_at` | `timestamptz` | Diisi saat update. |
| `modified_by` | `bigint` | Diisi saat update. |

**Indexes / constraints yang perlu dibuat di migrasi** (DBML tidak menuliskan index eksplisit, jadi ikuti konvensi `_shared/conventions-api.md` §6–7):

- `FOREIGN KEY (anggota_id) REFERENCES anggota(id)` (tanpa cascade khusus di DBML → default `RESTRICT`/`NO ACTION`).
- Index `idx_mst_medsos_anggota_id ON mst_medsos(anggota_id) WHERE is_deleted = false` — list per anggota adalah query paling sering.
- **Bukan** `gorm.DeletedAt` — pakai kolom soft-delete eksplisit sesuai DBML.

---

## 3. Endpoint (rancangan Bab 9 — master-data CRUD standar, semua di bawah `/api/v1`)

Semua rute di-guard `middleware.JWTAuth` + `perm.Require("medsos.<aksi>")`. Envelope `{success,message,data?,errors?}`.

| Method | Path | Permission | Auth | Deskripsi |
|---|---|---|---|---|
| GET | `/medsos` | `medsos.read` | Bearer | List + filter + paginate. Query: `page, per_page, q, sort, anggota_id, jenis_medsos`. |
| GET | `/medsos/{id}` | `medsos.read` | Bearer | Detail satu baris. |
| POST | `/medsos` | `medsos.create` | Bearer | Buat baris baru. |
| PUT | `/medsos/{id}` | `medsos.update` | Bearer | Ubah baris. |
| DELETE | `/medsos/{id}` | `medsos.delete` | Bearer | **Soft delete**. |

**Payload — POST `/medsos`** (`CreateMedsosReq`):

```json
{
  "anggota_id": 42,
  "jenis_medsos": "instagram",
  "konten_medsos": "https://instagram.com/slam.team",
  "icon": "instagram",
  "kode": "ig-main",
  "tipe": 0
}
```

**Payload — PUT `/medsos/{id}`** (`UpdateMedsosReq`) — sama tanpa `anggota_id` (pemilik tidak dipindah):

```json
{
  "jenis_medsos": "tiktok",
  "konten_medsos": "https://tiktok.com/@slam.team",
  "icon": "tiktok",
  "kode": "tt-main",
  "tipe": 0
}
```

**Response — item (`MedsosResponse`, di dalam `data`):**

```json
{
  "id": 101,
  "anggota_id": 42,
  "kode": "ig-main",
  "tipe": 0,
  "icon": "instagram",
  "jenis_medsos": "instagram",
  "konten_medsos": "https://instagram.com/slam.team",
  "created_at": "2026-09-08T03:00:00Z",
  "created_by": 7,
  "modified_at": null,
  "modified_by": null
}
```

**Response — list** (`data` = page):

```json
{
  "success": true,
  "message": "OK",
  "data": { "items": [ /* MedsosResponse[] */ ], "page": 1, "per_page": 20, "total": 3, "last_page": 1 }
}
```

DELETE mengembalikan `data: null` dengan `message` sukses.

---

## 4. Hak akses (rancangan Bab 3.5 — `_shared/permission-matrix.md` baris 58)

| Modul | Grup | Super Admin (0) | Admin (10) | Moderator (20) | User (30) | Guest (99) |
|---|---|---|---|---|---|---|
| `medsos` | Konten | CRUD | CRUD | CRUD | CRUD | R |

Permission codes yang di-seed: `medsos.create`, `medsos.read`, `medsos.update`, `medsos.delete`.

**Cakupan (dimensi `role_permission.cakupan`):**

- **Super Admin / Admin / Moderator → `semua`** untuk keempat aksi (kelola medsos anggota mana pun).
- **User → `milik_sendiri`** untuk keempat aksi. "Milik sendiri" = baris yang `anggota_id`-nya sama dengan `anggota_id` milik akun user (`users.anggota_id` dari Claims). User tidak boleh membuat/mengubah/menghapus medsos anggota lain.
- **Guest → `read`** saja. Tidak ada rute publik `/medsos` khusus di peta endpoint; baca Guest disurface sebagai bagian dari agregasi profil publik (`GET /public/profil`) bila profil menampilkan medsos. Modul ini sendiri hanya mengimplementasi CRUD ber-auth di atas.
- `is_super` bypass semua cek (konvensi RBAC global).

Backend WAJIB menerapkan cakupan di query (baca DAN tulis/hapus), bukan hanya menyembunyikan tombol — lihat `_shared/conventions-api.md` §8.

---

## 5. Kebutuhan BACKEND (Go) — `internal/modules/core/medsos/`

Ikuti module pattern (`domain/dto/repository/service/handler` + `main.medsos.go`).

### domain/ — `medsos.go`
- Struct `Medsos` map ke tabel `mst_medsos`; `func (Medsos) TableName() string { return "mst_medsos" }`.
- Field: `ID int64`, `AnggotaID int64`, `Kode *string`, `Tipe int`, `Icon *string`, `JenisMedsos string`, `KontenMedsos string`, embed `Audit` (created/modified/deleted + `IsDeleted`) sesuai contoh `_shared/conventions-api.md` §6.
- Kolom nullable (`kode`, `icon`, `modified_*`, `deleted_*`) → pointer.

### dto/ — `medsos.go`
- `CreateMedsosReq`: `AnggotaID int64 binding:"required"`, `JenisMedsos string binding:"required,max=50"`, `KontenMedsos string binding:"required,max=255"`, `Icon string binding:"omitempty,max=50"`, `Kode string binding:"omitempty,max=50"`, `Tipe int binding:"omitempty,min=0"`.
- `UpdateMedsosReq`: sama tanpa `AnggotaID` (immutable).
- `ListMedsosQuery`: embed `ListQuery` (`page, per_page, q, sort`) + `AnggotaID int64 form:"anggota_id"` + `JenisMedsos string form:"jenis_medsos"`.
- `MedsosResponse`: seluruh kolom non-audit-internal (lihat §3). Sertakan `created_at/created_by/modified_at/modified_by`; sembunyikan `is_deleted/deleted_*` (json `-`).
- Mapper `ToResponse(m domain.Medsos) MedsosResponse`.

### repository/ — `medsos_repository.go`
- `NewMedsosRepository(db *gorm.DB)`.
- `List(ctx, q, scope)` → `Count` lalu `Find` pada query ter-scope yang sama; selalu `Where("is_deleted = false")`; filter `anggota_id` & `jenis_medsos` bila diisi; `q` → `ILIKE` pada `jenis_medsos`/`konten_medsos`/`kode`; whitelist kolom sort (`created_at`, `jenis_medsos`, `tipe`); default `-created_at`. Cap `per_page` 100.
- `FindByID(ctx, id, scope)` → satu baris, `is_deleted=false`; `gorm.ErrRecordNotFound` → `ErrNotFound`.
- `Create/Update/SoftDelete`. `SoftDelete` set `is_deleted=true, deleted_at=now(), deleted_by=<actor>` (bukan hard delete).
- `AnggotaExists(ctx, anggotaID)` → validasi FK sebelum insert (anggota ada & `is_deleted=false`).
- `scope` = filter cakupan (mis. `func(*gorm.DB) *gorm.DB`) yang di-inject service.

### service/ — `medsos_service.go`
Semua aturan bisnis + scoping cakupan:
- **Create**: validasi `AnggotaID` ada (`AnggotaExists`, else `ErrValidation`/`ErrNotFound`). Jika cakupan aktor `milik_sendiri`, paksa `AnggotaID == claims.AnggotaID` (else `ErrForbidden`) — user tak boleh menautkan ke anggota lain. Trim `konten_medsos`; default `tipe=0`; jika `icon` kosong, boleh isi otomatis dari `jenis_medsos` (opsional, aman). Set `created_by` dari Claims.
- **Update**: `FindByID` ter-scope; jika cakupan `milik_sendiri` dan `anggota_id` baris ≠ `claims.AnggotaID` → `ErrForbidden`. `anggota_id` tidak diubah. Set `modified_by/modified_at`.
- **Delete**: `FindByID` ter-scope; soft delete; set `deleted_by`.
- **List/Detail**: terapkan scope cakupan → `milik_sendiri` menambahkan `Where("anggota_id = ?", claims.AnggotaID)`.
- Resolusi cakupan: baca `c.Get("cakupan:medsos.<aksi>")` yang di-set `PermGuard`; `is_super`/`semua` → tanpa filter.
- **Audit**: tulis `log_aktivitas` untuk create/update/delete (aktor, modul `medsos`, aksi, `reff_type=mst_medsos`, `reff_id`, ringkasan, `nilai_lama`/`nilai_baru`).
- Sentinel errors dipetakan handler via `response.FromError`.

### handler/ — `medsos_handler.go`
- `List/Detail/Create/Update/Delete`. Bind (`ShouldBindQuery`/`ShouldBindJSON`) → `validator.Explain` pada error → panggil service dengan `middleware.Claims(c)` → envelope (`OK`/`Created`). Handler tidak menyentuh `*gorm.DB`.

### main.medsos.go
```go
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
    g := rg.Group("/medsos", middleware.JWTAuth(m.jwtMgr))
    g.GET("",        m.perm.Require("medsos.read"),   m.h.List)
    g.GET("/:id",    m.perm.Require("medsos.read"),   m.h.Detail)
    g.POST("",       m.perm.Require("medsos.create"), m.h.Create)
    g.PUT("/:id",    m.perm.Require("medsos.update"), m.h.Update)
    g.DELETE("/:id", m.perm.Require("medsos.delete"), m.h.Delete)
}
```
Registrasi di `internal/router/router.go`: `medsos.Initialize(db, jwtMgr, permGuard).SetupRoutes(apiV1)`.

### Migrasi & seeder
- `migrations/00XX_mst_medsos.up.sql` (+ `.down.sql`) — buat tabel `mst_medsos` persis kolom DBML, FK `anggota_id → anggota(id)`, index parsial `idx_mst_medsos_anggota_id ... WHERE is_deleted=false`. Nomor migrasi = berikutnya yang tersedia (setelah anggota).
- **Seeder permission**: pastikan 4 baris `mst_permission` (`medsos.create/read/update/delete`) untuk `modul_id` = modul `medsos`, dan baris `role_permission` sesuai matriks §4 dengan cakupan (SA/Admin/Mod=`semua`, User=`milik_sendiri`). Idempoten (`ON CONFLICT DO NOTHING/UPDATE`). Biasanya ini bagian dari seeder RBAC global Fase 1 — modul ini hanya memastikan barisnya ada.

### Edge cases
- `anggota_id` tidak ada / anggota terhapus → `422`/`404`, jangan insert.
- User (`milik_sendiri`) POST dengan `anggota_id` orang lain → `403`.
- `konten_medsos` > 255 / `jenis_medsos` kosong → `422` via binding.
- Update/Delete id yang bukan milik user ter-scope → `404` (bukan `403`, agar tidak membocorkan keberadaan).
- Tidak ada background job. Tidak ada file. Tidak ada NRA/QR.

---

## 6. Kebutuhan FRONTEND (Angular) — `pages/medsos/`

> **WAJIB:** setiap prompt/skrin APP mengumumkan **"Using design-taste-frontend"**, jalankan pre-flight/audit, map ke **Bootstrap 5** (sudah terpasang), dan **ENFORCE** token di `workflow/_shared/design-tokens.md` (near-black bg, dark-gray surface, aksen SLAM red, teks putih, heading UPPERCASE tebal ber-letter-spacing). Note: *"Jika sumber `blog-fe` dipulihkan, mirror layout/menu-nya untuk skrin ini; jika tidak, ikuti design tokens."*

### Halaman/komponen (standalone, signals, zoneless)
- `pages/medsos/medsos-list.ts|html|scss` — tabel signal-based (pola canonical `_shared/conventions-app.md` §9). Kolom: Anggota, Platform (`jenis_medsos`) + ikon, Konten (link klik), aksi. Filter: `q` (search), `anggota_id` (bila dibuka dari detail anggota, prefill & lock), `jenis_medsos`. Pagination via signal `page`.
- `pages/medsos/medsos-form.ts|html|scss` — form tambah/ubah (Reactive Forms `nonNullable`).
- `pages/medsos/medsos.service.ts` — thin wrapper `ApiService` (feature-scoped, `providedIn:'root'`).
- `core/models/medsos.model.ts` — interface `Medsos` + `MedsosForm` + `MedsosQuery`.
- Rute lazy di `app.routes.ts`: `{ path: 'medsos', loadComponent: ..., canActivate: [authGuard, permissionGuard], data: { permission: 'medsos.read' } }` (+ subrute `new`, `:id/edit`).

### Form (fields — mirror DTO API, validator sejajar `binding`)
| Field UI | Control | Validator klien (mirror Go) |
|---|---|---|
| Anggota | select (dari `/anggota`) — di `create`; disembunyikan/dikunci saat konteks user `milik_sendiri` (auto ke anggota sendiri) | `required` (create) |
| Platform (`jenis_medsos`) | select preset (instagram/tiktok/youtube/facebook/whatsapp/x/lainnya) atau text | `required, maxLength(50)` |
| Ikon (`icon`) | select nama ikon dari icon-set (bukan upload) — default mengikuti platform | `maxLength(50)` |
| Konten (`konten_medsos`) | text (URL/handle) | `required, maxLength(255)` |
| Kode (`kode`) | text opsional | `maxLength(50)` |
| Tipe (`tipe`) | number opsional | `min(0)` |

Pada `422/400`, map `res.errors` (field→pesan) ke control (`setErrors({ server: msg })`) — server otoritatif.

### API service (envelope via `ApiService`)
```ts
@Injectable({ providedIn: 'root' })
export class MedsosService {
  private api = inject(ApiService);
  list(q: MedsosQuery)             { return this.api.get<Page<Medsos>>('/medsos', q); }
  detail(id: number)               { return this.api.get<Medsos>(`/medsos/${id}`); }
  create(b: MedsosForm)            { return this.api.post<Medsos>('/medsos', b); }
  update(id: number, b: MedsosForm){ return this.api.put<Medsos>(`/medsos/${id}`, b); }
  remove(id: number)               { return this.api.delete<void>(`/medsos/${id}`); }
}
```

### Permission-gating (UX mirror; backend tetap otoritatif)
- Tombol **Tambah** `*hasPermission="'medsos.create'"`; **Ubah** `medsos.update`; **Hapus** `medsos.delete`; **Detail/lihat** `medsos.read`.
- Route guard `permissionGuard` + `data.permission = 'medsos.read'`.
- Sidebar otomatis dari `/modul` bila user punya `medsos.read` (grup **Konten**).

### File handling
- **Tidak ada.** `icon` hanya nama string dari icon-set (render `<i>`/`<span>` ikon), tidak menyentuh `/files`. Jangan buat uploader.

### i18n (`public/i18n/IND.json` & `ENG.json`, namespace `MEDSOS.*`)
`MEDSOS.TITLE`, `MEDSOS.ADD`, `MEDSOS.FORM.ANGGOTA`, `MEDSOS.FORM.JENIS`, `MEDSOS.FORM.ICON`, `MEDSOS.FORM.KONTEN`, `MEDSOS.FORM.KODE`, `MEDSOS.FORM.TIPE`, `MEDSOS.COL.PLATFORM`, `MEDSOS.COL.KONTEN`, plus `COMMON.*`/`VALIDATION.*`. IND default; kedua file sinkron.

---

## 7. ALUR form → API → DATABASE (kontrak koherensi)

**Pemetaan field (form → payload key → kolom DB `mst_medsos`):**

| Field form (Angular) | Payload key (JSON) | Kolom DB | Catatan |
|---|---|---|---|
| Anggota (select) | `anggota_id` | `anggota_id` | Wajib (create); immutable saat update; dipaksa = anggota sendiri jika cakupan `milik_sendiri`. |
| Platform (select/text) | `jenis_medsos` | `jenis_medsos` | Wajib, ≤ 50. |
| Ikon (select nama) | `icon` | `icon` | Nama icon-set, ≤ 50. BUKAN berkas. |
| Konten (text URL/handle) | `konten_medsos` | `konten_medsos` | Wajib, ≤ 255. |
| Kode (text) | `kode` | `kode` | Opsional, ≤ 50. |
| Tipe (number) | `tipe` | `tipe` | Opsional int ≥ 0, default 0. |
| — (server) | — | `created_at/created_by` | Diisi server dari `now()`/Claims saat create. |
| — (server) | — | `modified_at/modified_by` | Diisi server saat update. |
| — (server) | — | `is_deleted/deleted_at/deleted_by` | Diisi server saat soft delete. |

**Pemetaan aksi (user action → endpoint → tulisan DB):**

| Aksi user | Endpoint | Tulisan DB |
|---|---|---|
| Simpan medsos baru | `POST /medsos` | `INSERT mst_medsos` (semua field form + `created_by`, `created_at=now()`, `is_deleted=false`). + `log_aktivitas` (aksi `create`). |
| Simpan perubahan | `PUT /medsos/{id}` | `UPDATE mst_medsos SET jenis_medsos, konten_medsos, icon, kode, tipe, modified_by, modified_at WHERE id=? AND is_deleted=false` (+scope cakupan). + `log_aktivitas` (aksi `ubah`, `nilai_lama`/`nilai_baru`). |
| Hapus medsos | `DELETE /medsos/{id}` | `UPDATE mst_medsos SET is_deleted=true, deleted_at=now(), deleted_by=? WHERE id=?` (+scope). + `log_aktivitas` (aksi `hapus`). Tidak ada hard delete. |
| Buka list | `GET /medsos?...` | `SELECT ... WHERE is_deleted=false [+ anggota_id/jenis filter] [+ cakupan]` + `COUNT`. |
| Buka detail | `GET /medsos/{id}` | `SELECT ... WHERE id=? AND is_deleted=false [+ cakupan]`. |

---

## 8. Dependencies / prasyarat & Acceptance criteria

**Prasyarat:**
- **Fase 0** — skeleton Go+Angular, `response` envelope, `middleware.JWTAuth`, `log_aktivitas`, konvensi timestamptz.
- **Fase 1** — RBAC (`PermGuard.Require`, seed permission `medsos.*` + `role_permission` cakupan, sidebar dinamis, `permissionGuard`/`*hasPermission`).
- **Modul anggota (Fase 2)** — tabel `anggota` + endpoint `/anggota` (untuk FK & dropdown pemilih anggota).

**Acceptance criteria (checkable):**
- [ ] Migrasi `mst_medsos` cocok persis dengan DBML (kolom, tipe, FK `anggota_id`, soft-delete, index parsial). Tanpa `AutoMigrate`; ada `.up`/`.down`.
- [ ] `POST /medsos` menyimpan baris dengan `created_by` aktor; `anggota_id` invalid → 422/404.
- [ ] `GET /medsos` mendukung `page/per_page/q/sort/anggota_id/jenis_medsos`, hanya baris `is_deleted=false`, envelope page benar (`items/page/per_page/total/last_page`).
- [ ] `PUT /medsos/{id}` mengubah field & `modified_by/at`; `anggota_id` tidak berubah.
- [ ] `DELETE /medsos/{id}` = soft delete (baris tetap ada, `is_deleted=true`), tak muncul lagi di list.
- [ ] Permission enforced di **backend**: role tanpa `medsos.*` → 403; **User cakupan `milik_sendiri`** hanya bisa CRUD medsos anggota-nya sendiri (tulis & baca), medsos anggota lain → 404/403.
- [ ] `is_super` bypass; Guest tak punya rute tulis.
- [ ] `log_aktivitas` tercatat untuk create/update/delete.
- [ ] Frontend: list+form jalan, validasi mirror DTO, tombol ter-gate `*hasPermission`, `icon` dirender sebagai ikon (bukan uploader), i18n IND/ENG sinkron, token desain diterapkan (taste-skill diinvoke).
- [ ] Form submit → baris DB benar-benar dibuat/diubah/di-soft-delete (verifikasi di `slamteam_db`).
