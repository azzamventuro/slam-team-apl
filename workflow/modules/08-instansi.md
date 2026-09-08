# Modul 08 — Master Data Instansi / Sekolah (`instansi`)

> Sumber otoritatif: `slamteam_db.dbml` (skema) > rancangan Bab 9/3.5/4.4 (rute, hak akses, penugasan) > `_shared/conventions-*.md` > aplikasi lama (tidak pernah). Jika prosa dan `.dbml` berbeda, `.dbml` menang.

---

## 1. Ringkasan & tujuan modul

**Instansi** adalah master data institusi / sekolah tempat setiap anggota terdaftar. Ini adalah modul ke-**5** dari 19 (`instansi`), grup **Master Data**, dikerjakan pada **Fase 2 (Anggota + master data)**.

Dua alasan modul ini penting melebihi CRUD biasa:

1. **Instansi menjadi lingkup (scope) anggota.** Kolom `anggota.instansi_id` → `mst_instansi.id` (relasi **restrict**). Sebuah instansi tidak boleh dihapus selama masih ada anggota yang menunjuknya.
2. **Instansi menggerakkan dimensi cakupan RBAC `instansi_sendiri`** (rancangan Bab 3.2). Moderator/User yang izinnya bercakupan `instansi_sendiri` hanya melihat data yang instansinya sama dengan instansi anggota si aktor. Instansi juga menjadi **penyaring penambahan massal peserta jadwal** ("latihan per sekolah", rancangan Bab 4.4) — bulk-assign by `mst_instansi`.

Instansi sendiri adalah **data referensi global**: setiap peran yang punya `instansi.read` melihat SELURUH instansi (cakupan efektif `semua`) — daftar sekolah bukan data pribadi. Cakupan `instansi_sendiri` diterapkan oleh modul LAIN (anggota, absensi, jadwal), bukan oleh modul instansi terhadap dirinya sendiri.

Prasyarat: **file-management (Fase 0)** untuk dua logo, dan **RBAC (Fase 1)** untuk `RequirePermission`.

---

## 2. Tabel & kolom

### `mst_instansi` (rancangan Kelompok 7; DBML baris 983)

| Kolom | Tipe | Aturan / catatan |
|-------|------|------------------|
| `id` | `bigint` [pk, increment] | PK internal. TIDAK pernah dipakai di URL publik. |
| `kode` | `varchar(50)` [unique] | Kode instansi, unik. Partial-unique `WHERE is_deleted = false` (kode boleh dipakai ulang setelah baris lama di-soft-delete). |
| `nama` | `varchar(150)` | Nama resmi instansi/sekolah. Wajib. |
| `nama_club` | `varchar(150)` | Nama klub/subunit di instansi tersebut (nullable). |
| `alamat` | `text` | Alamat lengkap (nullable). |
| `no_telepon` | `varchar(30)` | Nomor kontak (nullable). |
| `logo_utama_file_id` | `bigint` | FK → `mst_file.id` (nullable). Logo utama. |
| `logo_tambahan_file_id` | `bigint` | FK → `mst_file.id` (nullable). Logo alternatif/sekunder. |
| `tanggal_bergabung` | `date` | Tanggal instansi bergabung (nullable). `date`, bukan `timestamptz`. |
| `status` | `int` | Status instansi. **1 = aktif, 0 = nonaktif** (default 1). Divalidasi `oneof 0 1`. |
| `is_deleted` | `boolean` [default false] | Soft delete. |
| `deleted_at` | `timestamptz` | Diisi saat soft delete. |
| `deleted_by` | `bigint` | User id pelaku hapus. |
| `created_at` | `timestamptz` [default now()] | UTC. |
| `created_by` | `bigint` | User id pembuat. |
| `modified_at` | `timestamptz` | UTC, diisi saat update. |
| `modified_by` | `bigint` | User id pengubah. |

**Relasi (DBML baris 1180, 1230–1231):**
- `anggota.instansi_id` → `mst_instansi.id` **[delete: restrict]** — instansi dipakai anggota.
- `mst_instansi.logo_utama_file_id` → `mst_file.id`
- `mst_instansi.logo_tambahan_file_id` → `mst_file.id`

**Aturan lintas-modul (CLAUDE global):**
- Soft delete: `is_deleted + deleted_at + deleted_by` (BUKAN `gorm.DeletedAt` magic). Semua query list/read memfilter `is_deleted = false`.
- Waktu: `timestamptz` UTC. `tanggal_bergabung` adalah `date` murni (tanpa jam/zona).
- File: kolom `*_file_id` menyimpan **id numerik** `mst_file`, bukan path/URL. Berkas privat hanya disajikan lewat `GET /files/{uuid}/{varian}`.
- Index yang disarankan: partial-unique `kode WHERE is_deleted=false`; index `status` (untuk filter aktif).

---

## 3. Endpoint

Semua di bawah `/api/v1`, JWT wajib, dijaga `RequirePermission("instansi.<aksi>")`. Pola REST master-data standar (rancangan Bab 9).

| Method | Path | Permission | Auth | Body | Response `data` |
|--------|------|-----------|------|------|-----------------|
| `GET` | `/instansi` | `instansi.read` | JWT | — (query: `page, per_page, q, sort, status`) | `Paginated<InstansiResp>` |
| `GET` | `/instansi/{id}` | `instansi.read` | JWT | — | `InstansiResp` |
| `POST` | `/instansi` | `instansi.create` | JWT | `CreateInstansiReq` | `InstansiResp` (201) |
| `PUT` | `/instansi/{id}` | `instansi.update` | JWT | `UpdateInstansiReq` | `InstansiResp` |
| `DELETE` | `/instansi/{id}` | `instansi.delete` | JWT | — | `null` (soft delete) |

Logo diunggah lebih dulu melalui modul file **`POST /files`** (kategori `logo`, `reff_type=instansi`); respon berisi `id` + `uuid`. Form instansi lalu mengirim **`logo_utama_file_id` / `logo_tambahan_file_id`** (id numerik).

### Bentuk payload

**`CreateInstansiReq`**
```json
{
  "kode": "SMAN3-MLG",
  "nama": "SMA Negeri 3 Malang",
  "nama_club": "SLAM Chapter SMAN 3",
  "alamat": "Jl. Sultan Agung Utara No.7, Malang",
  "no_telepon": "0341-362426",
  "logo_utama_file_id": 42,
  "logo_tambahan_file_id": null,
  "tanggal_bergabung": "2024-01-15",
  "status": 1
}
```

**`UpdateInstansiReq`** — sama, semua field boleh dikirim ulang (full update). `kode` unik dicek mengabaikan baris ini sendiri.

**`InstansiResp`** (envelope `data`)
```json
{
  "id": 7,
  "kode": "SMAN3-MLG",
  "nama": "SMA Negeri 3 Malang",
  "nama_club": "SLAM Chapter SMAN 3",
  "alamat": "Jl. Sultan Agung Utara No.7, Malang",
  "no_telepon": "0341-362426",
  "logo_utama_file_id": 42,
  "logo_utama_uuid": "9b1f...-uuid",
  "logo_tambahan_file_id": null,
  "logo_tambahan_uuid": null,
  "tanggal_bergabung": "2024-01-15",
  "status": 1,
  "jumlah_anggota": 23,
  "created_at": "2024-01-15T02:00:00Z",
  "modified_at": null
}
```
`logo_*_uuid` di-join dari `mst_file` agar frontend dapat menarik gambar lewat `GET /files/{uuid}/medium`. `jumlah_anggota` = COUNT anggota aktif ber-instansi ini (memberi tahu apakah instansi boleh dihapus). Semua dalam envelope `{success,message,data,errors}`.

---

## 4. Hak akses (rancangan Bab 3.5)

| Modul | Super Admin | Admin | Moderator | User | Guest |
|-------|-------------|-------|-----------|------|-------|
| `instansi` | CRUD | CRUD | R | R | — |

Permission codes yang harus ada di `mst_permission` (di-seed Fase 1): `instansi.create`, `instansi.read`, `instansi.update`, `instansi.delete`.

**Cakupan:** instansi adalah data referensi global → `instansi.read` bercakupan **`semua`** untuk semua peran yang memilikinya (SA/Admin/Mod/User). Tidak ada penyempitan `instansi_sendiri` PADA modul ini. Super Admin `is_super` bypass semua cek. Guest tidak punya akses (—).

Menyembunyikan tombol di UI bukan keamanan: setiap aksi C/U/D punya pasangan `RequirePermission` di handler Go.

---

## 5. Kebutuhan BACKEND (Go)

Modul `internal/modules/core/instansi/` mengikuti pola `domain/dto/repository/service/handler + main.instansi.go`.

### domain
- `Instansi` struct memetakan `mst_instansi` (`TableName() = "mst_instansi"`). Embed `Audit` bersama (created/modified/deleted). File id nullable → `*int64`. `Status int`. `TanggalBergabung *time.Time` (kolom `date`; GORM `type:date`).

### dto
- `CreateInstansiReq`, `UpdateInstansiReq`, `InstansiResp`, `ListInstansiQuery` (embed `ListQuery` + `Status *int`).
- Binding: `kode required,max=50`; `nama required,max=150`; `nama_club max=150`; `no_telepon max=30`; `status oneof=0 1` (default 1); `logo_utama_file_id omitempty,gt=0`; `tanggal_bergabung` format `2006-01-02` (bind sebagai string lalu parse, atau `time.Time` dengan `time_format:"2006-01-02"`).

### repository
- `Create`, `Update`, `SoftDelete`, `FindByID`, `List(query) (items, total)`.
- Semua query filter `is_deleted = false`.
- `List`: `Count` lalu `Find` ter-paginate pada query yang sama; whitelist kolom `sort` (`kode, nama, status, created_at`); filter `q` LIKE pada `kode`/`nama`/`nama_club`; filter opsional `status`.
- `FindByID` & `List` **LEFT JOIN** `mst_file` dua kali untuk mengambil `logo_utama_uuid`/`logo_tambahan_uuid`, dan sub-select `COUNT(anggota)` `WHERE instansi_id=mst_instansi.id AND is_deleted=false` untuk `jumlah_anggota`.
- `ExistsKode(kode, exceptID)` untuk uji unik.
- `CountAnggotaAktif(instansiID)` untuk guard hapus.

### service
Aturan bisnis:
- **Unik kode:** `Create`/`Update` menolak bila `ExistsKode` true (`ErrConflict` → 409). Uji mengabaikan baris sendiri saat update.
- **Default status:** bila tidak dikirim → `1` (aktif).
- **Audit:** set `created_by`/`modified_by`/`deleted_by` dari `middleware.Claims(c).UserID`; `created_at` default DB, `modified_at = now()` saat update.
- **Guard hapus (penting):** `Delete` menolak bila `CountAnggotaAktif > 0` → `ErrConflict` "instansi masih dipakai oleh N anggota" (relasi anggota.instansi_id restrict; walau soft-delete tak melanggar FK DB, membiarkannya akan meninggalkan lingkup RBAC yatim). Soft delete: set `is_deleted=true, deleted_at=now(), deleted_by`.
- **Validasi file id:** bila `logo_*_file_id` dikirim, pastikan baris `mst_file` ada, `kategori='logo'`, belum di-soft-delete (opsional tapi disarankan; minimal cek keberadaan agar tidak menyimpan FK menggantung).
- **Audit log:** tulis `log_aktivitas` untuk create/update/delete (aktor, modul `instansi`, aksi, `reff_type=instansi`, `reff_id`, `nilai_lama`/`nilai_baru`, ip, user_agent).
- Map ke `InstansiResp` (termasuk uuid logo & jumlah_anggota).

### handler
- Bind → panggil service → `response.Created`/`OK`/`FromError`. Baca aktor via `middleware.Claims(c)`. RBAC dilakukan middleware; service tidak mengulang cek permission tetapi TETAP menerapkan aturan bisnis.

### main.instansi.go + router
```go
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/instansi", middleware.JWTAuth(m.jwtMgr))
	g.GET("",        m.perm.Require("instansi.read"),   m.h.List)
	g.GET("/:id",    m.perm.Require("instansi.read"),   m.h.Detail)
	g.POST("",       m.perm.Require("instansi.create"), m.h.Create)
	g.PUT("/:id",    m.perm.Require("instansi.update"), m.h.Update)
	g.DELETE("/:id", m.perm.Require("instansi.delete"), m.h.Delete)
}
```
Daftarkan di `internal/router/router.go`: `instansi.Initialize(db, jwtMgr, permGuard).SetupRoutes(apiV1)`.

### migrations + seeders
- **Migrasi** `00XX_mst_instansi.up.sql` / `.down.sql` — dibuat **setelah RBAC, SEBELUM anggota** (karena `anggota.instansi_id` FK **restrict** menunjuk tabel ini). Kolom persis `.dbml`. Sertakan partial-unique index `CREATE UNIQUE INDEX ux_mst_instansi_kode ON mst_instansi(kode) WHERE is_deleted=false;` dan index `status`. FK `logo_*_file_id → mst_file(id)`, `deleted_by/created_by/modified_by → users(id)` (nullable). `.down.sql` drop table + index.
- **Seeder idempotent** (`slamctl seed instansi`): satu instansi awal klub SLAM (rancangan Bab 12: "Data klub dan instansi awal"), mis. `kode='SLAM'`, `nama='Scouting Legion Airsofter Malang'`, `status=1`, `INSERT ... ON CONFLICT (kode) DO NOTHING`. Ini instansi default untuk anggota inti yang bukan dari sekolah tertentu.

### edge cases
- `kode` duplikat (termasuk bentrok dengan baris ter-soft-delete → partial index mengizinkan reuse).
- Hapus instansi yang masih dipakai anggota → 409, bukan 500.
- `logo_*_file_id` menunjuk file tak ada / bukan kategori logo.
- `page` melebihi `last_page` → kembalikan items kosong, bukan error.
- Update parsial vs full: gunakan full-replace PUT (kirim ulang semua field); field kosong string menimpa; `null` untuk logo = melepas logo.
- **Tidak ada background job.**

---

## 6. Kebutuhan FRONTEND (Angular)

**WAJIB dahulukan taste-skill.** Umumkan **"Using design-taste-frontend"**, jalankan pre-flight/audit-nya, petakan ke **Bootstrap 5**, dan tegakkan token `workflow/_shared/design-tokens.md` (near-black bg, dark surface, SLAM red accent, judul UPPERCASE wide-tracking). Jika sumber `blog-fe` dipulihkan, tiru layout/menu screen itu; jika tidak, ikuti design tokens.

### Halaman & komponen (`pages/instansi/`)
- `instansi-list.ts/.html/.scss` — tabel list (kolom: Kode, Nama, Nama Club, No. Telepon, Status badge, Jumlah Anggota, aksi). Filter: search `q` + dropdown `status`. Pagination signal-based (`toSignal` + `switchMap`, pola conventions-app §9).
- `instansi-form.ts/.html` — form create/edit (satu komponen, mode dibedakan by route `:id`).
- (opsional) `instansi-detail.ts` — read-only view; boleh digabung ke form read-only.

### Service (`pages/instansi/instansi.service.ts`, `providedIn:'root'`)
Wrapper tipis di atas `ApiService`:
```ts
list(q: InstansiQuery)  { return this.api.get<Page<Instansi>>('/instansi', q); }
detail(id: number)      { return this.api.get<Instansi>(`/instansi/${id}`); }
create(b: InstansiForm) { return this.api.post<Instansi>('/instansi', b); }
update(id: number, b: InstansiForm) { return this.api.put<Instansi>(`/instansi/${id}`, b); }
remove(id: number)      { return this.api.delete<void>(`/instansi/${id}`); }
```
Model `Instansi` mencerminkan `InstansiResp` (termasuk `logo_utama_uuid`, `jumlah_anggota`).

### Form (Reactive, `FormBuilder.nonNullable`) — cermin DTO
| Kontrol | Validator klien (cermin DTO Go) |
|---------|--------------------------------|
| `kode` | `required, maxLength(50)` |
| `nama` | `required, maxLength(150)` |
| `nama_club` | `maxLength(150)` |
| `alamat` | — (textarea) |
| `no_telepon` | `maxLength(30)`, pola telepon opsional |
| `logo_utama_file_id` | number \| null (diisi dari hasil upload) |
| `logo_tambahan_file_id` | number \| null |
| `tanggal_bergabung` | `<input type="date">` native → string `YYYY-MM-DD` \| null |
| `status` | select `oneof 0 1`, default 1 |

Pada `422/400`, petakan `res.errors` (field→pesan) ke kontrol yang cocok (`setErrors({server: msg})`). Server tetap otoritatif.

### File handling (logo)
- Dua input file (logo utama & tambahan). Alur: `FileService.upload('file', file, {kategori:'logo', reff_type:'instansi'})` → dapat `{id, uuid}` → simpan `id` ke kontrol `logo_*_file_id`, tampilkan preview via `FileService.imageUrl(uuid,'low')` (blob, token via interceptor; **revoke** objectURL saat destroy). **JANGAN set `Content-Type`** — `FormData` biarkan browser set boundary.
- Di list/detail, render logo dari `logo_utama_uuid` via `/files/{uuid}/medium` (blob). Fallback placeholder bila null.

### Guards & permission-gating
- Route: `{ path:'instansi', canActivate:[authGuard, permissionGuard], data:{permission:'instansi.read'}, loadComponent: ... }`.
- Tombol **Tambah** `*hasPermission="'instansi.create'"`; **Edit** `*hasPermission="'instansi.update'"`; **Hapus** `*hasPermission="'instansi.delete'"`. Mod/User hanya lihat (read) → tombol C/U/D tersembunyi. Ingat: sembunyi tombol = UX; backend tetap menegakkan.

### i18n (`public/i18n/{IND,ENG}.json`, namespace `INSTANSI`)
`INSTANSI.TITLE`, `INSTANSI.ADD`, `INSTANSI.FORM.KODE`, `.NAMA`, `.NAMA_CLUB`, `.ALAMAT`, `.NO_TELEPON`, `.LOGO_UTAMA`, `.LOGO_TAMBAHAN`, `.TANGGAL_BERGABUNG`, `.STATUS`, `INSTANSI.STATUS.AKTIF`, `INSTANSI.STATUS.NONAKTIF`, `INSTANSI.JUMLAH_ANGGOTA`, `INSTANSI.DELETE_CONFIRM`, `INSTANSI.DELETE_IN_USE` + reuse `COMMON.*`/`VALIDATION.*`. Jaga IND & ENG sinkron.

---

## 7. ALUR form → API → DATABASE (kontrak koherensi)

### Pemetaan field form → payload → kolom DB
| Field form (Angular) | Key payload (JSON) | Kolom `mst_instansi` |
|----------------------|--------------------|----------------------|
| Kode | `kode` | `kode` |
| Nama | `nama` | `nama` |
| Nama Club | `nama_club` | `nama_club` |
| Alamat | `alamat` | `alamat` |
| No. Telepon | `no_telepon` | `no_telepon` |
| Logo Utama (upload → id) | `logo_utama_file_id` | `logo_utama_file_id` |
| Logo Tambahan (upload → id) | `logo_tambahan_file_id` | `logo_tambahan_file_id` |
| Tanggal Bergabung | `tanggal_bergabung` | `tanggal_bergabung` |
| Status | `status` | `status` |
| — (dari JWT Claims) | — | `created_by` / `modified_by` / `deleted_by` |
| — (DB) | — | `created_at` / `modified_at` / `deleted_at` / `is_deleted` / `id` |

### Aksi pengguna → endpoint → tulisan DB
| Aksi | Endpoint | Tulisan DB |
|------|----------|------------|
| Unggah logo (sebelum simpan) | `POST /files` | INSERT `mst_file` (kategori=logo, reff_type=instansi) + 3 `mst_file_varian`; balikan `{id,uuid}` |
| Simpan instansi baru | `POST /instansi` | INSERT `mst_instansi` (kolom di atas, `created_by`, `created_at=now()`, `is_deleted=false`); INSERT `log_aktivitas` |
| Simpan perubahan | `PUT /instansi/{id}` | UPDATE `mst_instansi` (field + `modified_by`, `modified_at=now()`); INSERT `log_aktivitas` |
| Hapus instansi | `DELETE /instansi/{id}` | Jika ada anggota aktif → 409, TIDAK menulis. Jika tidak → UPDATE `is_deleted=true, deleted_at=now(), deleted_by`; INSERT `log_aktivitas` |
| Lihat daftar | `GET /instansi` | SELECT `WHERE is_deleted=false` + filter/paginate + join uuid logo + count anggota |
| Lihat detail | `GET /instansi/{id}` | SELECT satu baris (join uuid logo + count anggota) |

Nama key payload = nama kolom DB, 1:1, tanpa rename. Tidak meniru bentuk API aplikasi lama.

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- **Fase 0** — file layer (`POST /files`, `GET /files/{uuid}/{varian}`, 3 varian), `mst_pengaturan`, auth/JWT, `slamctl`.
- **Fase 1** — RBAC (`mst_permission` berisi `instansi.*`, `role_permission` ter-seed matriks, `PermGuard.Require`, guard/direktif Angular).
- Tabel `mst_instansi` termigrasi **sebelum** `anggota` (FK restrict).

### Acceptance criteria (checklist)
- [ ] Migrasi `mst_instansi` cocok persis dengan `.dbml` (kolom, tipe, default); partial-unique `kode WHERE is_deleted=false`; FK logo & audit.
- [ ] Seeder instansi klub SLAM idempoten (jalan 2× tidak duplikat).
- [ ] `GET/POST/PUT/DELETE /instansi` merespons dengan envelope `{success,message,data,errors}`.
- [ ] `instansi.create/update/delete` diblok 403 untuk Moderator & User; `instansi.read` diizinkan untuk SA/Admin/Mod/User; Guest 401/403.
- [ ] Kode duplikat ditolak 409; hapus instansi yang masih dipakai anggota ditolak 409 (bukan 500).
- [ ] POST membuat baris DB (`created_by` terisi); PUT memperbarui (`modified_at` terisi); DELETE men-soft-delete (`is_deleted=true`), tidak hard delete.
- [ ] List mendukung `page/per_page/q/sort/status`, mengembalikan `total` + `last_page`; `sort` di-whitelist (tidak ada SQL injection).
- [ ] Logo terunggah via `POST /files`, `logo_*_file_id` tersimpan; detail/list mengembalikan `logo_*_uuid`; frontend menampilkan gambar lewat `/files/{uuid}/medium` (blob, ber-token).
- [ ] Frontend: tombol C/U/D tersembunyi tanpa izin DAN diblok backend; form memvalidasi cermin DTO; `422` dipetakan inline.
- [ ] `log_aktivitas` tertulis untuk create/update/delete.
- [ ] taste-skill diumumkan & token design diterapkan (dark default, red accent, judul UPPERCASE); responsif; focus ring merah tidak dihapus.
