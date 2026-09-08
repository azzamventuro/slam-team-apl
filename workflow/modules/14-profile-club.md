# Modul 14 — Master Data Profile Club (`profile_club`)

> Sumber otoritatif: `slamteam_db.dbml` (skema) > rancangan Bab 9/3.5/12 (rute, hak akses, fase) > `_shared/*` > aplikasi lama (tidak pernah). Bila prosa dan `.dbml` berbeda, `.dbml` menang.

---

## 1. Ringkasan & tujuan modul

**Profile Club** menyimpan **identitas klub** SLAM Team (Scouting Legion Airsofter Malang): nama resmi, singkatan, banner, dua logo (simple + besar), alamat, dan keterangan. Data ini adalah sumber tunggal untuk **branding yang tampil di halaman publik / landing page** (header, footer, hero) dan sebagai referensi identitas di panel admin.

- **Nomor modul / kode:** `nn=14`, kode modul `profile_club` (salah satu dari 19 `mst_modul`).
- **Grup:** Konten (bersama `kegiatan`, `artikel`, `medsos`).
- **Fase:** Fase 2 (Anggota + master data), prasyarat Fase 0 (file layer + auth) dan Fase 1 (RBAC).
- **Sifat data:** praktis **singleton** — klub punya SATU identitas aktif. Endpoint tetap mengikuti pola CRUD standar, tetapi service menegakkan aturan "maksimal satu baris aktif" (lihat §5). Panel admin adalah **satu form edit**, bukan daftar panjang.
- **Dependensi:** `file-management` (Fase 0) — banner & kedua logo adalah `*_file_id` yang menunjuk ke `mst_file`. Modul ini TIDAK menulis file sendiri; ia memanggil file layer, menyimpan `file_id` yang dikembalikan, lalu me-resolve varian melalui file layer.

---

## 2. Tabel & kolom

Satu tabel: **`profile_club`** (tanpa prefix `mst_`; GORM `TableName()` HARUS eksplisit `"profile_club"` agar tidak ditebak menjadi `profile_clubs`). Klasifikasi: Master/Konten. Definisi persis dari `.dbml` (baris 1085–1101):

| Kolom | Tipe | Aturan / catatan |
|---|---|---|
| `id` | `bigint` PK identity | `gorm:"primaryKey"`, `int64`. Tidak dipakai di URL publik. |
| `nama` | `varchar(150)` | Nama resmi klub. Wajib. |
| `singkatan` | `varchar(50)` | Singkatan/akronim (mis. "SLAM"). Opsional tapi disarankan. |
| `banner_file_id` | `bigint` (nullable) | FK → `mst_file.id`. Gambar banner/hero landing. `*int64`. |
| `logo_simple_file_id` | `bigint` (nullable) | FK → `mst_file.id`. Logo ringkas (favicon/topbar). `*int64`. |
| `logo_besar_file_id` | `bigint` (nullable) | FK → `mst_file.id`. Logo penuh (hero/footer/KTA-adjacent). `*int64`. |
| `alamat` | `text` (nullable) | Alamat sekretariat/basecamp klub. |
| `keterangan` | `text` (nullable) | Deskripsi/tagline/riwayat singkat klub untuk landing. |
| `is_deleted` | `boolean` default `false` | Soft delete. Semua query list/read menambah `WHERE is_deleted = false`. |
| `deleted_at` | `timestamptz` (nullable) | Diisi saat soft delete. |
| `deleted_by` | `bigint` (nullable) | `users.id` aktor penghapus. |
| `created_at` | `timestamptz` default `now()` | UTC. |
| `created_by` | `bigint` (nullable) | `users.id` aktor pembuat. |
| `modified_at` | `timestamptz` (nullable) | UTC. |
| `modified_by` | `bigint` (nullable) | `users.id` aktor pengubah. |

**Relasi (dari `.dbml` baris 1232–1234):**
```
Ref: profile_club.banner_file_id       > mst_file.id
Ref: profile_club.logo_simple_file_id  > mst_file.id
Ref: profile_club.logo_besar_file_id   > mst_file.id
```

**Konvensi wajib (dari `conventions-api.md` §6):**
- `timestamptz` UTC di semua kolom waktu; import blank `_ "time/tzdata"`.
- Soft delete eksplisit (`is_deleted` + `deleted_at` + `deleted_by`), BUKAN `gorm.DeletedAt`. Embed struct `Audit`.
- File refs nullable → `*int64`. Jangan simpan path/URL di baris ini; resolve lewat file layer.
- Tiga gambar klub diunggah sebagai **publik** (`mst_file.is_publik = true`) agar landing page bisa menampilkannya tanpa auth.

---

## 3. Endpoint

Semua di bawah `/api/v1`. Admin CRUD = pola master-data standar (`api-endpoints.md` §13); publik = `api-endpoints.md` §12.

| # | Method | Path | Permission (`modul.aksi`) | Auth | Ringkas |
|---|---|---|---|---|---|
| 1 | GET | `/profile-club` | `profile_club.read` | bearer | List (praktis 0/1 baris) + filter/paginate. |
| 2 | GET | `/profile-club/{id}` | `profile_club.read` | bearer | Satu baris. |
| 3 | POST | `/profile-club` | `profile_club.create` | bearer | Buat identitas klub (ditolak bila sudah ada baris aktif). |
| 4 | PUT | `/profile-club/{id}` | `profile_club.update` | bearer | Ubah identitas klub. |
| 5 | DELETE | `/profile-club/{id}` | `profile_club.delete` | bearer | Soft delete. |
| 6 | GET | `/public/profil` | — | **publik** | Profil klub (+ prestasi publik). Tanpa auth. |

> Catatan rute: base path admin memakai tanda hubung `/profile-club` (bukan underscore); **kode permission tetap underscore** `profile_club.*` sesuai `mst_permission.kode`.

### Payload & bentuk respons

**POST `/profile-club`** / **PUT `/profile-club/{id}`** (JSON):
```json
{
  "nama": "Scouting Legion Airsofter Malang",
  "singkatan": "SLAM",
  "alamat": "Jl. Contoh No. 1, Malang",
  "keterangan": "Klub airsoft ...",
  "banner_file_id": 12,
  "logo_simple_file_id": 13,
  "logo_besar_file_id": 14
}
```
`*_file_id` opsional (boleh `null`). Upload gambar dilakukan lebih dulu ke `POST /files`; UI mengirim `file_id` hasilnya di sini.

**Respons admin (GET/POST/PUT), di dalam `data`):**
```json
{
  "id": 1,
  "nama": "Scouting Legion Airsofter Malang",
  "singkatan": "SLAM",
  "alamat": "Jl. Contoh No. 1, Malang",
  "keterangan": "Klub airsoft ...",
  "banner_file_id": 12,
  "logo_simple_file_id": 13,
  "logo_besar_file_id": 14,
  "banner_file_uuid": "b1f2...-uuid",
  "logo_simple_file_uuid": "c3d4...-uuid",
  "logo_besar_file_uuid": "e5f6...-uuid",
  "created_at": "2026-09-08T03:00:00Z",
  "modified_at": "2026-09-08T04:10:00Z"
}
```
> `*_file_uuid` diisi service dengan me-resolve `mst_file.uuid` dari `*_file_id` agar frontend bisa menarik gambar via `GET /files/{uuid}/{varian}` tanpa membocorkan id numerik. Field `is_deleted/deleted_*` tidak diserialisasi (`json:"-"`).

**GET `/profile-club`** — envelope `Paginated`:
```json
{ "success": true, "data": { "items": [ {..profil..} ], "page": 1, "per_page": 20, "total": 1, "last_page": 1 } }
```

**GET `/public/profil`** (publik, whitelist field aman):
```json
{ "success": true, "data": {
  "nama": "Scouting Legion Airsofter Malang",
  "singkatan": "SLAM",
  "alamat": "Jl. Contoh No. 1, Malang",
  "keterangan": "Klub airsoft ...",
  "banner_url": "/api/v1/files/b1f2...-uuid/medium",
  "logo_simple_url": "/api/v1/files/c3d4...-uuid/low",
  "logo_besar_url": "/api/v1/files/e5f6...-uuid/medium",
  "prestasi": [
    { "judul_kompetisi": "...", "peringkat": "...", "tingkat": "...", "tanggal_kompetisi": "2026-06-01" }
  ]
} }
```
> Publik TIDAK mengembalikan `id`, `created_by`, atau field audit. `prestasi` diambil dari tabel `prestasi` yang readable oleh Guest (rancangan Bab 3.5: `prestasi` Guest = R). Bila modul prestasi belum ada saat build, kembalikan `"prestasi": []` dan tambahkan join saat modul prestasi tersedia (lihat §5 catatan agregasi).

---

## 4. Hak akses

Baris matriks modul ini (rancangan Bab 3.5 / `permission-matrix.md` Tabel 1 / `api-endpoints.md` §13):

| Modul (`kode`) | Grup | Super Admin | Admin | Moderator | User | Guest |
|---|---|---|---|---|---|---|
| `profile_club` | Konten | CRUD | CRUD | R | R | R |

- **Cakupan:** `semua` untuk semua aksi. `profile_club` adalah data tunggal milik klub — **tidak** berdimensi kepemilikan (`instansi_sendiri`/`milik_sendiri` tidak berlaku).
- **Create/Update/Delete:** hanya Super Admin & Admin. Moderator, User, Guest hanya baca.
- **Guest = R:** identitas klub tampil di landing publik → dilayani lewat `GET /public/profil` (tanpa auth), bukan lewat rute admin ber-bearer.
- Super Admin (`is_super = true`) **bypass** semua pengecekan.
- Permission yang di-seed di Fase 1: `profile_club.read`, `profile_club.create`, `profile_club.update`, `profile_club.delete` (baris di `mst_permission`, kode `modul.aksi`).

---

## 5. Kebutuhan BACKEND (Go)

Ikuti pola modul `conventions-api.md` §1: `internal/modules/core/profile_club/{domain,dto,repository,service,handler} + main.profile_club.go`, didaftarkan di `internal/router/router.go`.

### domain/
- `ProfileClub` (GORM entity) memetakan seluruh kolom §2; `TableName() string { return "profile_club" }`.
- Embed struct audit bersama (`Audit`) sesuai `conventions-api.md` §6 (`is_deleted/deleted_at/deleted_by/created_*/modified_*`).
- `*_file_id` sebagai `*int64`.

### dto/
- `UpsertProfileClubReq` (dipakai POST & PUT): `Nama` (`binding:"required,max=150"`), `Singkatan` (`binding:"omitempty,max=50"`), `Alamat` (`binding:"omitempty"`), `Keterangan` (`binding:"omitempty"`), `BannerFileID *int64` (`binding:"omitempty"`), `LogoSimpleFileID *int64`, `LogoBesarFileID *int64`.
- `ProfileClubResp` — field admin termasuk `*_file_id` dan `*_file_uuid` (string, nullable) hasil resolve.
- `PublicProfilResp` — whitelist publik (§3): `nama, singkatan, alamat, keterangan, banner_url, logo_simple_url, logo_besar_url, prestasi[]`. Tanpa id/audit.
- `ListQuery` standar (`page, per_page, q, sort`) sesuai `conventions-api.md` §5.

### repository/
- `Create`, `GetByID`, `List` (count + paged find, semua scoped `is_deleted = false`), `Update`, `SoftDelete(id, actorID)`.
- `CountActive()` → jumlah baris `is_deleted = false` (untuk guard singleton).
- `GetActive()` → baris aktif pertama (dipakai `GET /public/profil`, urut `created_at ASC` atau `id ASC`).
- Resolve UUID file: sediakan helper yang mengambil `mst_file.uuid` untuk sekumpulan `file_id` (satu query `WHERE id IN (...)`), atau delegasikan ke file service. Repository menerjemahkan `gorm.ErrRecordNotFound` → `ErrNotFound`.

### service/
Aturan bisnis & validasi:
1. **Guard singleton:** `Create` menolak (`ErrConflict`, HTTP 409/422) bila `CountActive() > 0`. Pesan: "Profil klub sudah ada, gunakan ubah." UI diarahkan ke edit, bukan create ganda.
2. **Validasi file_id:** untuk tiap `*_file_id` non-nil, pastikan `mst_file` dengan id itu ada dan belum terhapus (via file service/repo). Bila tidak ada → `ErrValidation` pada field terkait. (Anti-orphan reference.)
3. **Publik = file publik:** saat menyimpan, tandai/verifikasi bahwa banner & logo bertipe `is_publik = true` di `mst_file` (identitas klub tampil di landing). Bila file layer menyediakan flag saat upload, UI mengunggahnya sebagai publik; service memvalidasi.
4. **Audit isian:** `Create` set `created_by = actor`; `Update` set `modified_at = now(), modified_by = actor`; `SoftDelete` set `is_deleted = true, deleted_at = now(), deleted_by = actor`.
5. **Resolve UUID:** setelah baca, isi `*_file_uuid` (admin) / `*_url` (publik) dari `mst_file.uuid`. URL publik dibentuk `"/api/v1/files/{uuid}/{varian}"` (banner→`medium`, logo_simple→`low`, logo_besar→`medium`).
6. **Log aktivitas:** tulis `log_aktivitas` untuk create/update/delete (modul `profile_club`, aksi, `reff_id`, `nilai_lama`/`nilai_baru` jsonb) sesuai `conventions-api.md` §12.
7. **Public profil aggregation:** `GetPublicProfil()` mengembalikan `GetActive()` + daftar prestasi Guest-readable. Bila modul prestasi belum tersedia saat build, kembalikan `prestasi: []` (jangan gagal). Saat prestasi ada, tambahkan read ringan (`SELECT judul_kompetisi, peringkat, tingkat, tanggal_kompetisi FROM prestasi WHERE is_deleted=false ORDER BY tanggal_kompetisi DESC LIMIT n`). Ini satu-satunya cross-module read; jangan menyalin logika prestasi.

Sentinel errors: pakai `response.FromError` + `ErrNotFound/ErrConflict/ErrValidation/ErrForbidden` (`conventions-api.md` §3).

### handler/
- Bind DTO (`ShouldBindJSON` untuk POST/PUT, `ShouldBind` query untuk list), panggil service dengan `middleware.Claims(c)`, balas via `response.OK/Created/...`.
- `PublicProfil` handler tanpa `JWTAuth`/`RequirePermission` (rute publik).
- Handler admin tidak menyentuh cakupan khusus (semua `semua`), tapi tetap di belakang `RequirePermission("profile_club.<aksi>")`.

### main.profile_club.go & router
```go
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
    g := rg.Group("/profile-club", middleware.JWTAuth(m.jwtMgr))
    g.GET("",        m.perm.Require("profile_club.read"),   m.h.List)
    g.GET("/:id",    m.perm.Require("profile_club.read"),   m.h.Detail)
    g.POST("",       m.perm.Require("profile_club.create"), m.h.Create)
    g.PUT("/:id",    m.perm.Require("profile_club.update"), m.h.Update)
    g.DELETE("/:id", m.perm.Require("profile_club.delete"), m.h.Delete)

    // publik — TANPA JWTAuth / RequirePermission
    rg.GET("/public/profil", m.h.PublicProfil)
}
```
Registrasi di `router.go`: `profile_club.Initialize(db, jwtMgr, permGuard).SetupRoutes(apiV1)`.

### Migrations & seeders
- **Migration** `profile_club` (numbered `.up.sql`/`.down.sql`, `conventions-api.md` §7) sesuai `.dbml`: kolom, default `is_deleted=false`, `created_at=now()`, FK ke `mst_file(id)` untuk tiga kolom file (ON DELETE tetap konsisten kebijakan proyek). Tidak ada `AutoMigrate`.
- **Seed permissions** (Fase 1, idempotent): `profile_club.read/create/update/delete` di `mst_permission` + baris `role_permission` sesuai matriks §4 (SA/Admin CRUD, Mod/User/Guest R), `cakupan = semua`.
- **Seed `mst_modul`:** baris `profile_club` (grup Konten) sudah termasuk dalam 19 seed (Fase 1). Modul ini tidak menambah seed modul baru.
- **Seed data awal (opsional):** satu baris `profile_club` placeholder (nama klub, singkatan "SLAM") boleh di-seed idempoten (`ON CONFLICT DO NOTHING` keyed by singkatan/nama) agar landing tidak kosong sebelum admin mengisi.

### Edge cases
- Create kedua kalinya → 409/422 (guard singleton), bukan diam-diam membuat dua identitas.
- `file_id` menunjuk file yang sudah terhapus/tidak ada → 422 field error.
- `GET /public/profil` saat belum ada baris → kembalikan objek dengan field kosong + `prestasi: []` (200), bukan 404, supaya landing tetap render.
- Soft delete baris satu-satunya → landing kembali kosong; `CountActive()` kembali 0 sehingga Create diizinkan lagi.
- Tidak ada background job untuk modul ini.

---

## 6. Kebutuhan FRONTEND (Angular)

> **Metode desain WAJIB (`conventions-app.md` §10):** setiap layar mengumumkan **"Using design-taste-frontend"**, menjalankan pre-flight/audit skill, memetakan ke **Bootstrap 5** (sudah terpasang), dan **menegakkan token** di `workflow/_shared/design-tokens.md` (near-black bg, dark-gray surface, aksen SLAM merah, teks putih, heading UPPERCASE wide-tracking). **Catatan blog-fe:** *"Jika sumber `blog-fe` dipulihkan, cermin layout/menu-nya untuk layar ini; jika tidak, ikuti design tokens."*

Angular 21 standalone + signals + zoneless (`conventions-app.md` §1). Folder: `pages/profile-club/`.

### Pages/components
- **`profile-club.ts` / `.html` / `.scss`** — SATU halaman edit-singleton (bukan tabel list). Karena data tunggal:
  - `OnInit`/effect: panggil `svc.getActive()` (mem-fetch baris pertama via `GET /profile-club?per_page=1`). Jika ada → mode **edit** (PUT), form terisi. Jika kosong → mode **create** (POST).
  - Preview banner + kedua logo (gambar privat → tarik blob via `FileService.imageUrl(uuid, varian)`; ingat `revokeObjectURL` on destroy, §8 conventions-app).
- (Opsional, YAGNI) tidak perlu halaman list terpisah; landing publik memakai komponen public terpisah bila fase 8 membutuhkannya.

### Form (Reactive, `FormBuilder.nonNullable`) — mirror DTO `UpsertProfileClubReq`
| Kontrol | Validator klien | DTO key |
|---|---|---|
| `nama` | `required`, `maxLength(150)` | `nama` |
| `singkatan` | `maxLength(50)` | `singkatan` |
| `alamat` | — (textarea) | `alamat` |
| `keterangan` | — (textarea) | `keterangan` |
| `banner_file_id` | — (di-set setelah upload) | `banner_file_id` |
| `logo_simple_file_id` | — | `logo_simple_file_id` |
| `logo_besar_file_id` | — | `logo_besar_file_id` |

- Tiga input file: pilih file → `FileService.upload(field, file, { is_publik: 'true', kategori: 'profile_club' })` → simpan `uuid`/`file_id` yang dikembalikan ke kontrol form + tampilkan preview. Kirim `FormData` **tanpa** set `Content-Type` (browser set boundary).
- Submit: `form.invalid` → `markAllAsTouched()`. Valid → `create()`/`update()`. Pada 422 map `res.errors` ke kontrol (`applyServerErrors`).

### API service (`pages/profile-club/profile-club.service.ts`, `providedIn:'root'`)
Wrapper tipis di atas `ApiService` (envelope di-unwrap otomatis):
```ts
getActive() { return this.api.get<Page<ProfileClub>>('/profile-club', { per_page: 1 }); }
detail(id: number)                 { return this.api.get<ProfileClub>(`/profile-club/${id}`); }
create(b: ProfileClubForm)         { return this.api.post<ProfileClub>('/profile-club', b); }
update(id: number, b: ProfileClubForm) { return this.api.put<ProfileClub>(`/profile-club/${id}`, b); }
remove(id: number)                 { return this.api.delete<void>(`/profile-club/${id}`); }
// publik (dipakai landing / fase 8) — tanpa auth:
publicProfil()                     { return this.api.get<PublicProfil>('/public/profil'); }
```

### Permission-gating (`conventions-app.md` §5)
- Route `pages/profile-club`: `canActivate: [authGuard, permissionGuard]`, `data: { permission: 'profile_club.read' }`.
- Tombol Simpan/Hapus dibungkus `*hasPermission="'profile_club.update'"` / `*hasPermission="'profile_club.create'"` / `*hasPermission="'profile_club.delete'"`. Moderator/User hanya melihat form read-only (disable kontrol bila tidak punya update/create).
- **Ingat:** menyembunyikan tombol bukan keamanan — handler Go tetap menegakkan. Frontend hanya cermin UX.

### i18n (`public/i18n/IND.json` + `ENG.json`, sinkron)
Namespace `PROFILE_CLUB`:
```
PROFILE_CLUB.TITLE, PROFILE_CLUB.SUBTITLE,
PROFILE_CLUB.FORM.NAMA, PROFILE_CLUB.FORM.SINGKATAN, PROFILE_CLUB.FORM.ALAMAT,
PROFILE_CLUB.FORM.KETERANGAN, PROFILE_CLUB.FORM.BANNER, PROFILE_CLUB.FORM.LOGO_SIMPLE,
PROFILE_CLUB.FORM.LOGO_BESAR, PROFILE_CLUB.UPLOAD_HINT,
COMMON.SAVE, COMMON.DELETE, COMMON.CANCEL, VALIDATION.REQUIRED, VALIDATION.MAXLENGTH
```
Tidak ada string UI hard-coded; identifier teknis (nama kolom/permission) tetap verbatim.

### File handling
- Gambar privat di panel admin → blob via `GET /files/{uuid}/{varian}` (interceptor menambah Bearer). Pakai `medium` untuk banner/logo besar preview, `low` untuk logo simple.
- Landing publik (fase 8) memakai `publicProfil()` → `*_url` sudah jadi (file publik), bisa langsung `<img [src]>`.

---

## 7. ALUR form → API → DATABASE (kontrak koherensi)

### 7.1 Pemetaan field form → payload → kolom DB
| Field form (Angular) | Request key (JSON) | Kolom `profile_club` | Catatan |
|---|---|---|---|
| `nama` | `nama` | `nama` | `varchar(150)`, wajib |
| `singkatan` | `singkatan` | `singkatan` | `varchar(50)` |
| `alamat` | `alamat` | `alamat` | `text` |
| `keterangan` | `keterangan` | `keterangan` | `text` |
| input file banner → upload → `banner_file_id` | `banner_file_id` | `banner_file_id` | `bigint` FK `mst_file.id` |
| input file logo simple → upload → `logo_simple_file_id` | `logo_simple_file_id` | `logo_simple_file_id` | `bigint` FK `mst_file.id` |
| input file logo besar → upload → `logo_besar_file_id` | `logo_besar_file_id` | `logo_besar_file_id` | `bigint` FK `mst_file.id` |
| — (server) | — | `created_by/modified_by/created_at/modified_at` | diisi service dari `Claims` + `now()` |

### 7.2 Aksi user → endpoint → tulisan DB
| Aksi user | Endpoint | Tulisan DB |
|---|---|---|
| Buka halaman (belum ada profil) | `GET /profile-club?per_page=1` | — (baca; kosong) → UI mode create |
| Unggah banner/logo | `POST /files` (FormData, `is_publik=true`) | INSERT `mst_file` + 3× `mst_file_varian` (original/medium/low) |
| Simpan (create) | `POST /profile-club` | INSERT `profile_club` (guard: hanya bila belum ada baris aktif) + `created_by`, `created_at`; INSERT `log_aktivitas` |
| Simpan (edit) | `PUT /profile-club/{id}` | UPDATE `profile_club` set kolom + `modified_by`, `modified_at`; INSERT `log_aktivitas` |
| Hapus | `DELETE /profile-club/{id}` | UPDATE `profile_club` set `is_deleted=true, deleted_at=now(), deleted_by`; INSERT `log_aktivitas` |
| Landing publik | `GET /public/profil` | — (baca `GetActive()` + prestasi Guest-readable) |

---

## 8. Dependencies/prasyarat & Acceptance criteria

### Prasyarat
- **Fase 0:** file layer (`POST /files`, `GET /files/{uuid}/{varian}`, `mst_file` + `mst_file_varian`, varian dari `mst_pengaturan`), auth/JWT, response envelope, `log_aktivitas`.
- **Fase 1:** RBAC (5 tabel), seed `mst_modul` (baris `profile_club`) + `mst_permission` (`profile_club.*`) + `role_permission` sesuai matriks, middleware `RequirePermission`, guard/`*hasPermission` di frontend.
- **Fase 2:** modul ini. Modul `prestasi` (juga Fase 2) diperlukan HANYA untuk melengkapi `GET /public/profil` (boleh menyusul; sementara `prestasi: []`).

### Acceptance criteria (checkable)
- [ ] Migration `profile_club` naik/turun bersih; kolom & FK persis `.dbml`; tidak ada `AutoMigrate`.
- [ ] Seed permission `profile_club.read/create/update/delete` + `role_permission` (SA/Admin CRUD, Mod/User/Guest R, cakupan `semua`) idempoten.
- [ ] `go build ./...` sukses; modul terdaftar di `router.go`.
- [ ] `POST /profile-club` membuat 1 baris; pemanggilan kedua saat sudah ada baris aktif → ditolak (409/422 guard singleton).
- [ ] `PUT /profile-club/{id}` mengubah baris; `modified_by/modified_at` terisi.
- [ ] `DELETE /profile-club/{id}` men-soft-delete (`is_deleted=true`), tidak hard delete.
- [ ] `RequirePermission` menegakkan: User/Moderator/Guest tidak bisa create/update/delete (403); Guest tidak bisa akses rute admin.
- [ ] `GET /public/profil` bekerja **tanpa** auth, mengembalikan hanya field whitelist (tanpa `id`/audit), dan tetap 200 (bukan 404) saat data kosong.
- [ ] `*_file_id` invalid ditolak 422; `*_file_uuid`/`*_url` ter-resolve benar saat data valid.
- [ ] Frontend: form men-submit → baris DB terbuat/terubah; upload gambar melewati `POST /files` (FormData tanpa `Content-Type`); preview gambar via blob endpoint.
- [ ] Frontend: tombol simpan/hapus tersembunyi tanpa permission DAN diblokir server; responsif; token desain (`design-tokens.md`) diterapkan; skill `design-taste-frontend` diumumkan & dijalankan.
- [ ] i18n IND/ENG sinkron; tidak ada string UI hard-coded.
