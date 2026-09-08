# 05 · Hak Akses Dinamis (RBAC)

> Modul: `hak_akses` · Key: `hak-akses` · Fase 1 · Grup: Sistem
> Tabel: `mst_modul`, `mst_permission`, `mst_role`, `role_permission`, `user_role`
> Prasyarat: Fase 0 (fondasi + auth-session)
> Sumber otoritatif: `slamteam_db.dbml` (Kelompok 1) > rancangan Bab 3 & Bab 9 > `_shared/*` > file ini. Bila prosa dan `.dbml` berbeda, **`.dbml` menang**.

---

## 1. Ringkasan & tujuan

Ini adalah **jantung sistem**: matriks hak akses tidak hidup di dokumen atau di kode, melainkan sebagai baris di basis data yang dapat diubah Super Admin lewat antarmuka (rancangan Bab 3.1). Modul ini menyediakan tiga hal yang dipakai SELURUH modul lain:

1. **5 tabel RBAC** (`mst_modul`, `mst_permission`, `mst_role`, `role_permission`, `user_role`) + seed 19 modul, permission per modul, 5 peran, dan matriks Bab 3.5.
2. **Middleware Go `PermGuard.Require("modul.aksi")`** yang dipasang pada setiap route terlindungi di seluruh modul (bukan hanya modul ini). Cache izin per-peran (`perm:role:{id}`), `is_super` bypass, `perm_version` bump saat matriks berubah, dan penegakan `cakupan`.
3. **Halaman Hak Akses (Angular)** + `permissionGuard` (canActivate) + directive `*hasPermission` + **sidebar dinamis** yang dibangkitkan dari modul + izin efektif. Layar utamanya adalah **grid peran × modul** — salah satu dari tiga layar tersulit tanpa preseden (rancangan Bab 11.3), dirancang lebih dahulu.

Modul ini sendiri adalah modul ke-14 dari 19 (`hak_akses`), grup **Sistem**, dan izinnya (`hak_akses.*`) hanya milik **Super Admin**. Tetapi kode yang dihasilkannya adalah infrastruktur yang menegakkan izin SEMUA modul.

**Prinsip yang tidak boleh dilanggar (Bab 3.4):** menyembunyikan tombol di UI **bukan** keamanan. Setiap izin yang ditampilkan di antarmuka WAJIB punya pasangan `PermGuard.Require(...)` pada handler Go. Frontend hanya cermin UX; backend adalah otoritas.

---

## 2. Tabel & kolom (dari `.dbml` Kelompok 1)

Semua timestamp `timestamptz` (UTC). PK `bigint identity`. `mst_role` memakai soft delete (`is_deleted`+`deleted_at`+`deleted_by`); empat tabel lain adalah referensi/pivot tanpa soft delete.

### `mst_modul` — 19 modul (menu + area fungsi)
| Kolom | Tipe | Aturan |
|---|---|---|
| `id` | bigint PK | |
| `kode` | varchar(50) | **UNIQUE, NOT NULL** — mis. `anggota`, `hak_akses` |
| `nama` | varchar(100) | NOT NULL |
| `grup` | varchar(50) | `Master Data` / `Konten` / `Sistem` / `Operasional` |
| `route` | varchar(150) | path frontend, mis. `/anggota` |
| `icon` | varchar(50) | nama ikon (bukan berkas) |
| `urutan` | int | default 0 — urutan menu |
| `tampil_di_menu` | boolean | default true |
| `is_aktif` | boolean | default true |
| `created_at`/`created_by`/`modified_at`/`modified_by` | audit | |

Seed: 19 baris (Bab 2): `anggota, admin, moderator, user, instansi, unit, prestasi, inorga, kegiatan, artikel, medsos, profile_club, dokumen, hak_akses, lokasi, jadwal, absensi, izin, kta`. **Tidak ada modul `registrasi`** (tidak ada registrasi mandiri).

### `mst_permission` — pasangan modul × aksi
| Kolom | Tipe | Aturan |
|---|---|---|
| `id` | bigint PK | |
| `modul_id` | bigint | NOT NULL → `mst_modul.id` |
| `aksi` | enum `aksi_permission` | `create, read, update, delete, approve, assign, export, override, print, cabut, batal_sesi` |
| `kode` | varchar(100) | **UNIQUE, NOT NULL** — format `modul.aksi`, mis. `anggota.create`, `absensi.override` |
| `nama` | varchar(150) | |
| `keterangan` | text | |
| `is_berbahaya` | boolean | default false — tandai izin berdampak besar (delete, override, cabut, hak_akses.*) |
| Index | `(modul_id, aksi) UNIQUE` | |

### `mst_role` — peran (soft delete)
| Kolom | Tipe | Aturan |
|---|---|---|
| `id` | bigint PK | |
| `kode` | varchar(50) | **UNIQUE, NOT NULL** |
| `nama` | varchar(100) | NOT NULL |
| `level` | int | NOT NULL — **SA=0, Admin=10, Moderator=20, User=30, Guest=99** (kecil = tinggi) |
| `keterangan` | text | |
| `is_sistem` | boolean | default false — 5 peran bawaan bernilai true (tak bisa dihapus) |
| `is_super` | boolean | default false — **bypass semua pengecekan izin** |
| `is_aktif` | boolean | default true |
| `is_deleted`/`deleted_at`/`deleted_by` | soft delete | |
| `created_*`/`modified_*` | audit | |

Aturan yang ditegakkan backend (Note tabel + Bab 3.3): (1) peran `is_super` tak bisa dihapus/dinonaktifkan/diturunkan; (2) super admin terakhir tak boleh dihapus; (3) anti-eskalasi; (4) hanya super admin mengelola peran.

### `role_permission` — **matriks** (jantung Bab 3.5)
| Kolom | Tipe | Aturan |
|---|---|---|
| `id` | bigint PK | |
| `role_id` | bigint | NOT NULL → `mst_role.id` (`ON DELETE CASCADE`) |
| `permission_id` | bigint | NOT NULL → `mst_permission.id` |
| `cakupan` | enum `cakupan_permission` | default `semua` · `semua` / `instansi_sendiri` / `milik_sendiri` |
| `created_at`/`created_by` | audit | |
| Index | `(role_id, permission_id) UNIQUE`, `role_id` | |

Kolom `cakupan` membedakan "lihat absensi semua orang" dari "lihat absensi sendiri" — tanpa kolom ini aturan itu tersembunyi di handler dan tak bisa diubah Super Admin (Bab 3.2).

### `user_role` — pemetaan user → peran
| Kolom | Tipe | Aturan |
|---|---|---|
| `id` | bigint PK | |
| `user_id` | bigint | NOT NULL → `users.id` (`ON DELETE CASCADE`) |
| `role_id` | bigint | NOT NULL → `mst_role.id` |
| `is_utama` | boolean | default true — peran utama (dipakai untuk JWT) |
| `berlaku_sampai` | timestamptz | opsional (peran sementara) |
| `created_at`/`created_by` | audit | |
| Index | `(user_id, role_id) UNIQUE` | |

Satu peran per user untuk sekarang; kode membaca lewat tabel ini sejak awal agar multi-peran nanti tanpa migrasi. `users.role_id` tetap menyimpan peran utama (untuk JWT & FK `ON DELETE RESTRICT`); `user_role` adalah sumber yang dapat berkembang.

**Enum yang dipakai** (dibuat sekali di migrasi enum Fase 0, jangan buat ulang): `aksi_permission`, `cakupan_permission`.

---

## 3. Endpoint (rancangan Bab 9 / `_shared/api-endpoints.md` §2)

Semua di bawah `/api/v1`, `Auth: bearer` (JWTAuth), digerbang `PermGuard.Require("hak_akses.<aksi>")`. Peran `is_super` **bypass**. Envelope `{success,message,data,errors}`.

| Method | Path | Permission | Fungsi |
|---|---|---|---|
| GET | `/modul` | `hak_akses.read` | 19 modul (untuk grid RBAC). Terurut `grup,urutan`. |
| GET | `/permissions` | `hak_akses.read` | Semua baris `mst_permission` (modul × aksi), dikelompokkan per modul. |
| GET | `/roles` | `hak_akses.read` | Daftar peran + `level` + `is_super` + jumlah user. |
| POST | `/roles` | `hak_akses.create` | Buat peran. |
| PUT | `/roles/{id}` | `hak_akses.update` | Ubah nama/level/keterangan/is_aktif. Anti-eskalasi: tak bisa mengubah peran `level <= level pemanggil`. |
| DELETE | `/roles/{id}` | `hak_akses.delete` | Soft delete peran. **Tolak** jika `is_super` / `is_sistem` / super admin terakhir / masih dipakai user. |
| GET | `/roles/{id}/permissions` | `hak_akses.read` | Matriks izin peran itu (per permission + `cakupan`). |
| PUT | `/roles/{id}/permissions` | `hak_akses.update` | **Ganti** matriks peran. Anti-eskalasi: tak bisa memberi izin yang pemanggil tak punya. **Bump `perm_version`** + invalidasi cache. |

### Payload

`POST /roles` request:
```json
{ "kode": "pelatih", "nama": "Pelatih", "level": 25, "keterangan": "Pelatih lapangan", "is_aktif": true }
```
`PUT /roles/{id}` request: sama tanpa `kode` (kode peran tak diubah setelah dibuat), boleh sebagian.

`GET /roles` response `data`:
```json
[ { "id": 2, "kode": "admin", "nama": "Admin", "level": 10, "is_super": false,
    "is_sistem": true, "is_aktif": true, "jumlah_user": 3 } ]
```

`GET /modul` response `data` (untuk grid):
```json
[ { "id": 1, "kode": "anggota", "nama": "Anggota", "grup": "Master Data",
    "icon": "users", "urutan": 1,
    "permissions": [ { "id": 3, "aksi": "create", "kode": "anggota.create", "is_berbahaya": false }, … ] } ]
```

`GET /roles/{id}/permissions` response `data`:
```json
{ "role": { "id": 2, "nama": "Admin", "level": 10, "is_super": false },
  "permissions": [ { "permission_id": 14, "kode": "absensi.read", "cakupan": "semua" },
                   { "permission_id": 15, "kode": "file.delete", "cakupan": "milik_sendiri" } ] }
```

`PUT /roles/{id}/permissions` request (mengganti seluruh matriks peran):
```json
{ "permissions": [
  { "permission_id": 14, "cakupan": "semua" },
  { "permission_id": 15, "cakupan": "milik_sendiri" }
] }
```
Server menghapus baris `role_permission` peran ini yang tak ada di payload, meng-upsert sisanya, lalu bump `perm_version` global + evict `perm:role:{id}`.

### Integrasi dengan `/me/permissions` (milik modul auth, Fase 0)
`GET /me/permissions` (Auth bearer, tanpa gerbang izin) TIDAK dibangun di modul ini tetapi **memanggil resolver efektif** yang modul ini sediakan (`PermGuard.Effective(roleID)`). Ia mengembalikan izin efektif + daftar modul menu untuk sidebar:
```json
{ "is_super": false, "perm_version": 7,
  "permissions": ["anggota.read","jadwal.create", …],
  "cakupan": { "absensi.read": "milik_sendiri" },
  "modul": [ { "kode":"anggota","nama":"Anggota","grup":"Master Data","icon":"users","route":"/anggota","urutan":1 } ] }
```
`modul` = modul `tampil_di_menu=true` yang peran itu punya minimal `.read`. **Keputusan koherensi:** `GET /modul` digerbang `hak_akses.read` (SA-only, untuk grid), sehingga sidebar untuk semua peran dibangun dari `modul` di `/me/permissions`, BUKAN dari `/modul` langsung. Kalau tidak, sidebar akan kosong untuk non-SA.

---

## 4. Hak akses modul ini (matriks Bab 3.5)

| Izin (`modul.aksi`) | Super Admin | Admin | Moderator | User | Guest |
|---|---|---|---|---|---|
| `hak_akses.create` / `.read` / `.update` / `.delete` | **CRUD** | — | — | — | — |

`hak_akses` adalah **Super Admin only**. Karena SA `is_super=true` melakukan bypass, praktisnya SA melewati semua gerbang; peran lain tak punya baris `role_permission` untuk `hak_akses.*` sama sekali. `cakupan` untuk `hak_akses.*` selalu `semua` (tak relevan — hanya SA yang menyentuhnya).

---

## 5. Kebutuhan BACKEND (Go)

Modul di `internal/modules/core/hakakses/` (paket `hakakses` — tanpa tanda hubung). Middleware bersama di `internal/middleware/perm.go` (dipakai semua modul, bukan hanya ini). Ikuti `_shared/conventions-api.md`.

### 5.1 Prasyarat lintas-modul: perluas JWT Claims (koordinasi dengan Fase 0 auth)
`pkg/jwt.Claims` scaffold hanya membawa `UserID/Email/Name`. RBAC butuh tambahan (bila Fase 0 belum menambahkannya, modul ini yang menambahkan):
```go
type Claims struct {
    UserID      uint   `json:"user_id"`
    AnggotaID   uint   `json:"anggota_id"`
    InstansiID  uint   `json:"instansi_id"` // untuk cakupan instansi_sendiri
    RoleID      uint   `json:"role_id"`
    IsSuper     bool   `json:"is_super"`
    PermVersion int    `json:"perm_version"`
    jwt.RegisteredClaims
}
```
`Manager.Issue(...)` diperluas mengisi field ini saat login (baca peran utama dari `user_role.is_utama=true` / `users.role_id`, `is_super` dari `mst_role`, `perm_version` dari counter global).

### 5.2 `internal/middleware/perm.go` — `PermGuard` (infrastruktur inti)
```go
type PermGuard struct { db *gorm.DB; rdb *redis.Client } // rdb boleh nil → cache in-proc
func NewPermGuard(db *gorm.DB, rdb *redis.Client) *PermGuard

// Require memasang pengecekan izin pada satu route.
func (g *PermGuard) Require(perm string) gin.HandlerFunc

// Effective mengembalikan map[kode_permission]cakupan efektif untuk sebuah peran (cached).
func (g *PermGuard) Effective(roleID uint) (map[string]string, error)

// Invalidate menghapus cache satu peran (dipanggil saat matriks/peran berubah).
func (g *PermGuard) Invalidate(roleID uint)
```
`Require(perm)` logic: baca `Claims(c)`; jika `nil` → 401; jika `IsSuper` → `c.Next()` (bypass); ambil `Effective(RoleID)` (cache `perm:role:{id}`, Redis bila ada, else `sync.Map` in-proc); jika `perm` tak ada di map → `response.Forbidden(c, ...)` + `c.Abort()`; jika ada, simpan cakupan ke context (`c.Set("cakupan:"+perm, cak)`) agar service bisa mempersempit query, lalu `c.Next()`.

### 5.3 Tambahan `response`
Tambah `Forbidden` (403) dan `NotFound` (404) ke `internal/shared/response/response.go` (jangan fork paket). Envelope tetap `{success,message,data?,errors?}`.

### 5.4 domain (`domain/`)
Struct GORM 1:1 dengan tabel + `TableName()` eksplisit: `Modul`→`mst_modul`, `Permission`→`mst_permission`, `Role`→`mst_role`, `RolePermission`→`role_permission`, `UserRole`→`user_role`. Tipe enum Go: `AksiPermission string` (const create…batal_sesi), `CakupanPermission string` (const semua/instansi_sendiri/milik_sendiri). Embed `Audit` (lihat conventions §6). `Role` pakai kolom soft delete eksplisit.

### 5.5 dto (`dto/`)
`CreateRoleReq{Kode,Nama,Level,Keterangan,IsAktif}`, `UpdateRoleReq` (pointer field untuk partial), `SetPermissionsReq{Permissions []PermItem}` dengan `PermItem{PermissionID uint binding:"required"; Cakupan string binding:"omitempty,oneof=semua instansi_sendiri milik_sendiri"}`. Response: `RoleResp`, `ModulResp{…, Permissions []PermissionResp}`, `RolePermissionsResp`. Binding: `kode` `required,max=50,alphanum|contains=_` (lowercase), `nama` `required,max=100`, `level` `required,min=1,max=98` (tolak 0 = reserved SA dan 99 hanya untuk Guest bawaan; peran baru tak boleh super).

### 5.6 repository (`repository/`)
`ListModulWithPermissions()`, `ListPermissions()`, `ListRoles()` (+ hitung `jumlah_user` via join `user_role`), `GetRole(id)`, `CreateRole`, `UpdateRole`, `SoftDeleteRole`, `GetRolePermissions(roleID)`, `ReplaceRolePermissions(roleID, items)` (dalam **transaksi**: hapus yang tak ada + upsert), `CountSuperAdminRoles()`, `CountUsersByRole(roleID)`. Semua query peran filter `is_deleted=false`. Terjemahkan `gorm.ErrRecordNotFound` → `ErrNotFound`.

### 5.7 service (`service/`) — aturan bisnis (OTORITAS)
- **Anti-eskalasi level (Bab 3.3):** `UpdateRole` & `SetPermissions` menolak bila `target.level <= pemanggil.level` (kecuali pemanggil `is_super`). Peran baru (`CreateRole`) `level` harus `> pemanggil.level`.
- **Anti-eskalasi izin:** `SetPermissions` menolak bila payload memuat `permission_id` yang **tidak dimiliki** pemanggil (bandingkan dengan `Effective(pemanggil.RoleID)`; SA bypass). Tak seorang pun bisa memberi izin yang ia sendiri tak punya.
- **Lindungi super/sistem:** `UpdateRole` tak boleh mengubah `is_super`/`level` peran `is_super`, tak menonaktifkan `is_sistem`. `DeleteRole` menolak bila peran `is_super`, `is_sistem`, **super admin terakhir** (`CountSuperAdminRoles()<=1` untuk peran super), atau masih dipakai (`CountUsersByRole>0`).
- **Bump versi & cache:** setiap mutasi (`Create/Update/Delete/SetPermissions`) → increment counter global `perm_version` (di `mst_pengaturan` kunci `rbac.perm_version`, atau sequence khusus) + `perm.Invalidate(roleID)`. `ponytail: counter global, bump me-reload SEMUA token; kolom perm_version per-peran perlu migrasi mst_role — tambah bila reload selektif jadi masalah.`
- **Audit:** tiap mutasi tulis `log_aktivitas` (`modul="hak_akses"`, `aksi`, `reff_type="mst_role"`, `reff_id`, `nilai_lama`/`nilai_baru` jsonb, `aktor_user_id`, `ip_address`, `user_agent`).

### 5.8 handler (`handler/`)
Bind DTO → `response.Unprocess` on error → panggil service dengan `middleware.Claims(c)` → `response.FromError(c,err)` (map sentinel → 400/403/404/409, else `Internal`). Baca `id` path via `c.Param("id")`.

### 5.9 `main.hakakses.go` + registrasi router
```go
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard) *Module
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
    g := rg.Group("", middleware.JWTAuth(m.jwtMgr))
    g.GET("/modul",       m.perm.Require("hak_akses.read"),   m.h.ListModul)
    g.GET("/permissions", m.perm.Require("hak_akses.read"),   m.h.ListPermissions)
    g.GET("/roles",       m.perm.Require("hak_akses.read"),   m.h.ListRoles)
    g.POST("/roles",      m.perm.Require("hak_akses.create"), m.h.CreateRole)
    g.PUT("/roles/:id",   m.perm.Require("hak_akses.update"), m.h.UpdateRole)
    g.DELETE("/roles/:id",m.perm.Require("hak_akses.delete"), m.h.DeleteRole)
    g.GET("/roles/:id/permissions", m.perm.Require("hak_akses.read"),   m.h.GetRolePermissions)
    g.PUT("/roles/:id/permissions", m.perm.Require("hak_akses.update"), m.h.SetRolePermissions)
}
```
Di `router.go`: bangun `permGuard := middleware.NewPermGuard(db, rdb)` SEKALI, teruskan ke semua modul; `hakakses.Initialize(db, jwtMgr, permGuard).SetupRoutes(apiV1)`.

### 5.10 Migrasi + Seeder
- **Migrasi** `migrations/0004_rbac.up.sql` (+ `.down.sql`): buat 5 tabel + indeks + FK (`role_permission.role_id … ON DELETE CASCADE`, `user_role.user_id … ON DELETE CASCADE`, `users.role_id … ON DELETE RESTRICT`). Enum sudah dibuat di `0001_enums`.
- **Seeder** `cmd/slamctl` (`slamctl seed rbac`), **idempotent** (`ON CONFLICT (kode) DO NOTHING/UPDATE`):
  1. **19 modul** menu (`tampil_di_menu=true`) sesuai Bab 2, dengan `grup`+`urutan`+`icon`.
  2. **Modul non-menu `file`** (`tampil_di_menu=false`) — agar `file.create`/`file.delete` di matriks Bab 3.5 punya `modul_id` valid. (Idempotent: bila Fase 0 file-layer sudah menyisipkannya, dilewati.) `ponytail: rekonsiliasi FK — Bab 3.5 memberi izin file.* tapi "file" bukan salah satu 19 modul menu; satu baris modul non-menu menutup celah tanpa mengubah skema.`
  3. **`mst_permission`** per modul: modul master/konten → `create,read,update,delete`; `jadwal` +`assign,batal_sesi`; `absensi` → `create,read,override,delete,export`; `izin` → `create,read,approve`; `kta` → `create,read,print,cabut`; `file` → `create,delete`; `hak_akses` → `create,read,update,delete`. Tandai `is_berbahaya=true` untuk `*.delete`, `absensi.override`, `kta.cabut`, `hak_akses.*`.
  4. **5 peran** (`is_sistem=true`): SA(0,`is_super=true`), Admin(10), Moderator(20), User(30), Guest(99).
  5. **`role_permission`** = matriks Bab 3.5 untuk **Admin/Moderator/User/Guest saja** (SA `is_super` bypass → tak diberi baris, mencegah SA kehilangan akses saat modul baru ditambah). `cakupan='semua'` kecuali: `absensi.read` User→`milik_sendiri`; `izin.read` User→`milik_sendiri`; `file.delete` Mod & User→`milik_sendiri`.
  6. Set `mst_pengaturan` `rbac.perm_version = 1` bila belum ada.

**Edge cases:** peran tanpa user boleh dihapus; peran dengan user → 409 (harus dipindah dulu); mengganti izin dengan array kosong = mencabut semua izin peran (valid, tapi Guest/User bawaan tetap boleh); `cakupan` hanya bermakna pada aksi berdimensi kepemilikan (read/update/delete data orang lain vs sendiri) — untuk aksi lain simpan `semua`.

**Background job:** tidak ada.

---

## 6. Kebutuhan FRONTEND (Angular 21)

Ikuti `_shared/conventions-app.md`. **Wajib** memanggil taste-skill (lihat §10 conventions-app + `_shared/design-tokens.md`).

### 6.1 Core (dipakai semua modul — bangun di sini bila belum ada)
- `core/services/permission.service.ts` — `load()` dari `/me/permissions`, signal `_perms:Set<string>`, `_super`, `_cakupan:Map`, `_modul`, `ready`. `can(code)= _super() || _perms().has(code)`. `scope(code)= _cakupan().get(code) ?? 'semua'`.
- `core/guards/permission.guard.ts` — `permissionGuard` membaca `route.data['permission']`, izinkan bila `perms.can(need)`, else redirect `/forbidden`.
- `shared/directives/has-permission.directive.ts` — `*hasPermission="'kode'"` (structural, `effect()` reaktif zoneless).
- `layouts/sidebar/sidebar.ts` — menu di-`computed` dari `perms.modul()` (dari `/me/permissions`), dikelompokkan per `grup`, item aktif = teks merah + bg merah-soft + bar kiri merah 3px.

### 6.2 Halaman modul ini — `pages/hak-akses/`
Route: `{ path:'hak-akses', canActivate:[authGuard, permissionGuard], data:{ permission:'hak_akses.read' }, loadComponent: … }`.

**(A) Daftar peran — `roles-list.ts`.** Tabel peran (`GET /roles`): kode, nama, level, badge `is_super`/`is_sistem`, jumlah user. Aksi digerbang `*hasPermission`: `Tambah`(`hak_akses.create`), `Edit`(`hak_akses.update`), `Hapus`(`hak_akses.delete`). Tombol Edit/Hapus **disable** bila `role.level <= currentUser.level` atau `is_super`/`is_sistem` (cermin anti-eskalasi; backend tetap menegakkan).

**(B) Form peran — `role-form.ts`.** Reactive form `nonNullable`: `kode`(required,max50,pattern lowercase `^[a-z0-9_]+$`, hanya saat create), `nama`(required,max100), `level`(required,min pemanggil.level+1,max98), `keterangan`, `is_aktif`. `POST /roles` / `PUT /roles/{id}`. Map `res.errors` ke kontrol.

**(C) Grid izin peran × modul — `permission-grid.ts` (LAYAR TERSULIT #1, rancang lebih dahulu).**
- Ambil `GET /modul` (baris = modul, dikelompokkan `grup`) dan `GET /roles/{id}/permissions` (matriks peran terpilih).
- Layout: pemilih peran di atas; **grid** dengan baris = modul (dikelompokkan header grup), kolom = aksi (`create/read/update/delete` + kolom khusus `assign/batal_sesi/override/export/approve/print/cabut` yang hanya muncul bila modul punya aksi itu). Sel = checkbox izin.
- Untuk izin berdimensi kepemilikan (`absensi.read`, `izin.read`, `file.delete`, dan read/update/delete data), sel punya **dropdown `cakupan`** (`semua`/`instansi_sendiri`/`milik_sendiri`) yang aktif saat dicentang.
- **Anti-eskalasi UI:** checkbox untuk izin yang pemanggil **tak punya** → disabled + tooltip "Anda tidak memiliki izin ini". Bila peran terpilih `level <= currentUser.level` atau `is_super` → seluruh grid read-only.
- Simpan (`hak_akses.update`) → `PUT /roles/{id}/permissions` dengan `{permissions:[{permission_id,cakupan}]}` untuk sel tercentang. Setelah sukses tampilkan toast + reload (perm_version berubah).
- Density tinggi: sticky header + kolom modul sticky, zebra baris, chip status berwarna token, header UPPERCASE wide-tracking. Responsif: di layar sempit jadi accordion per modul (kartu berisi daftar aksi + toggle), bukan tabel yang melebar.

### 6.3 i18n (`public/i18n/{IND,ENG}.json`)
Namespace `HAK_AKSES.*`: `TITLE`, `ROLES`, `PERMISSIONS`, `GRID_TITLE`, `ROLE`, `LEVEL`, `MODUL`, `AKSI`, `CAKUPAN`, `CAKUPAN.SEMUA/INSTANSI_SENDIRI/MILIK_SENDIRI`, `FORM.KODE/NAMA/LEVEL/KETERANGAN/AKTIF`, `IS_SUPER`, `IS_SISTEM`, `JUMLAH_USER`, `NO_PERMISSION_TOOLTIP`, `LAST_SUPER_ADMIN`, `PROTECTED_ROLE`. Plus `COMMON.SAVE/CANCEL/ADD/EDIT/DELETE/DETAIL/EMPTY`, `VALIDATION.*`. IND default; kedua file sinkron.

### 6.4 File handling
Tidak ada unggah berkas di modul ini.

### 6.5 Catatan desain (wajib di prompt)
"Using design-taste-frontend" → pre-flight/audit → map ke Bootstrap 5 → enforce `_shared/design-tokens.md` (dark default, SLAM red aksen, heading UPPERCASE wide-tracking, focus ring merah, min tap 44px, `prefers-reduced-motion`, status tak hanya lewat warna). **Blog-fe restore note:** "Jika sumber `blog-fe` dipulihkan, tiru tata letak/menunya untuk layar ini; jika tidak, ikuti design tokens." (`blog-fe` saat ini kosong.)

---

## 7. ALUR form → API → DATABASE (kontrak koherensi)

### 7.1 Buat peran
| Field form | Payload key | Kolom DB (`mst_role`) |
|---|---|---|
| Kode | `kode` | `kode` (UNIQUE) |
| Nama | `nama` | `nama` |
| Level | `level` | `level` |
| Keterangan | `keterangan` | `keterangan` |
| Aktif | `is_aktif` | `is_aktif` |
| — | — | `is_sistem=false`, `is_super=false`, `created_by=<aktor>` |

Aksi **Simpan (create)** → `POST /roles` → INSERT `mst_role` + INSERT `log_aktivitas` + increment `rbac.perm_version`.

### 7.2 Ubah / hapus peran
- **Simpan (edit)** → `PUT /roles/{id}` → UPDATE `mst_role` (`modified_at/by`) setelah cek anti-eskalasi level. → `log_aktivitas`, bump `perm_version`, `Invalidate(id)`.
- **Hapus** → `DELETE /roles/{id}` → set `is_deleted=true,deleted_at,deleted_by` bila lolos guard (bukan super/sistem/super-terakhir/tak dipakai). → `log_aktivitas`, bump, invalidate.

### 7.3 Simpan grid izin
| Sel grid (checkbox+cakupan) | Payload item | Kolom DB (`role_permission`) |
|---|---|---|
| centang izin `X` | `{permission_id: X.id, cakupan}` | INSERT/UPSERT baris `(role_id, permission_id, cakupan)` |
| hilangkan centang izin `Y` | (tidak ada di array) | DELETE baris `(role_id=?, permission_id=Y)` |
| dropdown cakupan | `cakupan` | `cakupan` (`semua`/`instansi_sendiri`/`milik_sendiri`) |

Aksi **Simpan grid** → `PUT /roles/{id}/permissions` (transaksi: hapus selisih + upsert) → increment `rbac.perm_version` → `perm.Invalidate(id)` → `log_aktivitas` (`nilai_lama`=matriks lama, `nilai_baru`=matriks baru). Token lama peran itu memuat ulang izin pada request berikut (cache miss) atau saat refresh (perm_version berubah).

### 7.4 Penegakan runtime (setiap request modul apa pun)
`JWTAuth` → `Claims` (role_id, is_super, perm_version) → `PermGuard.Require("modul.aksi")`: `is_super`→lanjut; else `Effective(role_id)` (cache `perm:role:{id}`) → cek `modul.aksi` ada; simpan `cakupan` ke context → service mempersempit query (`semua`/`instansi_sendiri`→`WHERE instansi_id=?`/`milik_sendiri`→`WHERE anggota_id=?|created_by=?`).

---

## 8. Dependencies & Acceptance criteria

**Prasyarat:** Fase 0 selesai — skeleton Go+Angular, `mst_pengaturan` (untuk `rbac.perm_version`), `log_aktivitas`, migrasi enum (`aksi_permission`,`cakupan_permission`), auth login/refresh (JWT + `sesi_login`), tabel `users`/`anggota` nyata (bukan placeholder). Redis opsional (fallback in-proc).

**Acceptance criteria (checklist):**
- [ ] Migrasi `0004_rbac` membuat 5 tabel + indeks + FK sesuai `.dbml`; `.down.sql` merollback bersih.
- [ ] `slamctl seed rbac` idempotent: menjalankan dua kali tak menduplikasi (19 modul + `file`, permission, 5 peran, matriks Bab 3.5, `rbac.perm_version=1`).
- [ ] SA `is_super` tak punya baris `role_permission`; tetap dapat mengakses semua route (bypass terbukti).
- [ ] `PermGuard.Require` menolak (403) request tanpa izin; mengizinkan yang punya; menaruh `cakupan` ke context.
- [ ] `GET /modul`, `/permissions`, `/roles`, `/roles/{id}/permissions` mengembalikan bentuk envelope + data seperti §3; hanya SA (bypass) yang lolos gerbang `hak_akses.read`.
- [ ] `POST/PUT/DELETE /roles` menegakkan: anti-eskalasi level, tak bisa ubah `is_super`/`is_sistem`, tak bisa hapus super admin terakhir / peran yang dipakai user.
- [ ] `PUT /roles/{id}/permissions` menolak pemberian izin yang pemanggil tak punya; mengganti matriks; bump `perm_version`; evict cache `perm:role:{id}`; tulis `log_aktivitas`.
- [ ] Perubahan matriks tercermin pada request peran itu berikutnya (tanpa restart).
- [ ] Angular: `permissionGuard` memblokir route tanpa izin; `*hasPermission` menyembunyikan tombol; sidebar dibangkitkan dari `/me/permissions.modul` per grup.
- [ ] Grid peran × modul: centang→dropdown cakupan aktif; izin yang tak dimiliki pemanggil disabled; simpan → `PUT` → baris `role_permission` berubah di DB.
- [ ] Semua string UI lewat i18n (IND+ENG sinkron); design tokens diterapkan (dark, SLAM red, focus ring merah); responsif (grid → accordion di layar sempit).
