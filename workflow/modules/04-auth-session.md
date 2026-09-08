# Modul 04 — Auth + Sesi Login (`auth`)

> Sumber otoritatif: `slamteam_db.dbml` (skema) > rancangan Bab 10.4 (super admin) / Bab 3 (identitas) / Bab 8.2 (konvensi) > `_shared/conventions-*.md` > aplikasi lama (tidak pernah). Jika prosa dan `.dbml` berbeda, `.dbml` menang.

---

## 1. Ringkasan & tujuan modul

**Auth** adalah fondasi **Fase 0** yang membuat seluruh modul lain bisa dijaga. Ia menyediakan login (username ATAU email), refresh token, logout, serta dua endpoint identitas (`/me`, `/me/permissions`) yang menggerakkan sidebar dinamis dan `*hasPermission` di frontend.

Dua hal yang membuat modul ini lebih dari sekadar "form login":

1. **Mengganti tabel `users` "mainan".** Scaffold auth yang ada (`internal/modules/core/auth`) memakai tabel `users` sementara. Modul ini menggantinya dengan **skema `.dbml` penuh**: `users` (dengan `anggota_id`, `role_id`, `timezone`, penguncian akun) + tabel **`sesi_login`** untuk refresh-token dan logout paksa.
2. **Tidak ada registrasi mandiri.** Baris `users` HANYA dibuat lewat panel admin (`user.create`) atau, untuk super admin pertama, lewat CLI `slamctl create-superadmin`. **Tidak ada** endpoint HTTP register dan **tidak ada** endpoint HTTP untuk membuat super admin.

Prasyarat: **project skeleton (01)** — JWT manager, `response` envelope, middleware `JWTAuth`/`Claims` sudah ada. RBAC (Fase 1) belum wajib untuk login, tetapi `/me/permissions` membaca hasil seed RBAC begitu Fase 1 selesai.

---

## 2. Tabel & kolom

### `users` (DBML baris 424)

| Kolom | Tipe | Aturan / catatan |
|-------|------|------------------|
| `id` | `bigint` [pk, increment] | PK internal. |
| `anggota_id` | `bigint` [not null] | FK → `anggota.id` (restrict). Setiap user adalah anggota. |
| `username` | `varchar(50)` [not null] | Unik **partial** `WHERE is_deleted=false`. |
| `email` | `varchar(150)` [not null] | Unik **partial** `WHERE is_deleted=false`. |
| `password` | `varchar(255)` | Hash bcrypt/argon2 — **tidak pernah** dikembalikan ke klien. |
| `role_id` | `bigint` [not null] | FK → `mst_role.id` (restrict). |
| `timezone` | `varchar(50)` [default `Asia/Jakarta`] | IANA. |
| `is_aktif` | `boolean` [default true] | Nonaktif → login ditolak. |
| `login_terakhir` | `timestamptz` | Diisi saat login sukses. |
| `password_diubah` | `timestamptz` | Diisi saat ganti password. |
| `gagal_login` | `int` [default 0] | Counter gagal berturut. |
| `terkunci_sampai` | `timestamptz` | Penguncian sementara setelah N gagal. |
| `is_deleted`/`deleted_at`/`deleted_by` | soft delete | |
| `created_at`/`created_by`/`modified_at`/`modified_by` | audit | |

### `sesi_login` (DBML baris 400)

| Kolom | Tipe | Aturan / catatan |
|-------|------|------------------|
| `id` | `bigint` [pk] | |
| `user_id` | `bigint` [not null] | FK → `users.id` (**cascade**). |
| `refresh_token` | `varchar(255)` [unique, not null] | Opaque acak (bukan JWT). |
| `ip_address` | `inet` | Bukti. |
| `user_agent` | `text` | Bukti. |
| `info_perangkat` | `jsonb` | Ringkasan perangkat. |
| `berlaku_sampai` | `timestamptz` [not null] | Masa berlaku refresh. |
| `dicabut_pada` | `timestamptz` | Diisi saat logout / rotasi. |
| `created_at` | `timestamptz` [default now()] | |

**Relasi (DBML baris 1167–1176):** `users.anggota_id → anggota.id` [restrict], `users.role_id → mst_role.id` [restrict], `user_role.user_id → users.id` [cascade], `sesi_login.user_id → users.id` [cascade].

**Aturan lintas-modul:** login menerima **username ATAU email** dalam satu query (keduanya unik). Waktu `timestamptz` UTC. Password hanya hash. Soft-delete difilter di semua query `users`.

---

## 3. Endpoint

Semua di bawah `/api/v1`. Login & refresh **publik**; sisanya **bearer**.

| Method | Path | Permission | Auth | Body | Response `data` |
|--------|------|-----------|------|------|-----------------|
| `POST` | `/auth/login` | — | public | `LoginReq` | `AuthResp` (token + user) |
| `POST` | `/auth/refresh` | — | public | `{refresh_token}` | `AuthResp` (rotasi) |
| `POST` | `/auth/logout` | — | bearer | — | `null` |
| `GET` | `/me` | — | bearer | — | `MeResp` |
| `GET` | `/me/permissions` | — | bearer | — | `{is_super, permissions[]}` |

### Bentuk payload

**`LoginReq`**
```json
{ "identifier": "azzz", "password": "••••••••" }
```
`identifier` = `username` ATAU `email`.

**`AuthResp`** (envelope `data`)
```json
{
  "access_token": "<jwt>",
  "refresh_token": "<opaque>",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": 1, "username": "azzz", "email": "achmad@…",
    "anggota_id": 12, "nama": "Achmad", "foto_url": "/api/v1/files/<uuid>/low",
    "role": { "id": 2, "nama": "Admin", "level": 10, "is_super": false },
    "perm_version": 7
  }
}
```
Access token JWT HS256 memuat `user_id`, `anggota_id`, `role_id`, `perm_version`, `is_super`, `exp`. Saat login sukses satu baris `sesi_login` dibuat. Saat RBAC berubah, `perm_version` di-bump sehingga token lama memuat ulang izin.

**`/me/permissions`** → `{ "is_super": false, "permissions": ["anggota.read","jadwal.create", …] }`.

---

## 4. Hak akses

Tidak ada gerbang permission di modul ini — hanya `bearer` untuk rute pasca-login. Otorisasi per-modul ditegakkan modul lain lewat `RequirePermission`. `is_super` di klaim JWT dipakai middleware untuk bypass.

**Bootstrap super admin:** `slamctl create-superadmin` (CLI-only) membuat satu `anggota` + `users` terikat peran Super Admin, **menolak** bila sudah ada user `is_super`. Bukan endpoint HTTP.

---

## 5. Kebutuhan BACKEND (Go)

Perluas modul `internal/modules/core/auth/` yang sudah ada (referensi implementasi). **Jangan** bangun ulang dari nol; ganti bagian tabel-mainan.

### domain
- `User` map ke `users` (embed `Audit`; `Password` di-tag `json:"-"`). `SesiLogin` map ke `sesi_login`.

### dto
- `LoginReq{ Identifier string binding:"required"; Password string binding:"required,min=8" }`, `RefreshReq{ RefreshToken string binding:"required" }`, `AuthResp`, `MeResp`, `MePermissionsResp`.

### repository
- `FindByUsernameOrEmail(id)` — `WHERE (username=? OR email=?) AND is_deleted=false` (satu query). `TouchLogin`, `IncrementGagalLogin`/`ResetGagalLogin`, `LockUntil`. `CreateSesi`, `FindSesiByToken`, `RevokeSesi`, `RevokeAllSesi(userID)`.

### service
- **Login:** cari user; bila terkunci (`terkunci_sampai > now()`) → `ErrForbidden`. Cek `is_aktif`. `utils.CheckPassword`; gagal → `IncrementGagalLogin` (bila melewati ambang dari `mst_pengaturan` → set `terkunci_sampai`), balas `ErrUnauthorized` pesan generik ("kredensial salah") — **jangan** bocorkan mana yang salah. Sukses → reset gagal, `TouchLogin`, terbitkan access JWT + refresh opaque, `CreateSesi`. Tulis `log_aktivitas` (modul `auth`, aksi `login`).
- **Refresh:** `FindSesiByToken` (belum dicabut, belum kedaluwarsa) → terbitkan access baru + refresh baru, `RevokeSesi` lama (rotasi). Token invalid → `ErrUnauthorized`.
- **Logout:** `RevokeSesi` untuk sesi token aktif.
- **/me:** join `anggota` (nama, foto uuid) + `mst_role`. **/me/permissions:** kumpulkan `role_permission` efektif (atau via cache PermGuard Fase 1); `is_super` → set true.

### handler
- Bind → service → `response.OK/Created/FromError`. `Claims(c)` untuk rute bearer. Jangan pernah balas hash password.

### main.auth.go + router
- Sudah terdaftar: `auth.Initialize(db, jwtMgr).SetupRoutes(apiV1)`. Tambah rute refresh/logout/me/permissions.

### migrations + seeders
- Migrasi `users` & `sesi_login` **persis `.dbml`** (partial-unique `username`/`email` `WHERE is_deleted=false`, FK `anggota_id`/`role_id` restrict, `sesi_login.user_id` cascade). Urutan: **setelah** `mst_role` (Fase 1) & `anggota`. `slamctl create-superadmin` sebagai seeder-CLI (bukan HTTP), idempoten-menolak bila super admin ada.

### edge cases
- Username & email duplikat lintas baris ter-soft-delete diizinkan (partial index). Akun terkunci → 403 dengan sisa waktu. Refresh token dipakai ulang setelah rotasi → 401. Password < 8 → 422. Timezone tak valid → default `Asia/Jakarta`.

---

## 6. Kebutuhan FRONTEND (Angular)

**WAJIB taste-skill.** Umumkan **"Using design-taste-frontend"**, tegakkan token (dark, SLAM red, judul UPPERCASE). Halaman login **sudah ada** di `pages/auth/login` — **perluas**, jangan bangun ulang.

### Halaman & service
- `pages/auth/login/login.*` — form `identifier` + `password`, tombol submit dengan state `saving()`, pesan error generik pada 401. Setelah sukses: simpan `slam_token`/`slam_user`, panggil `PermissionService.load()` lalu navigate `/dashboard`.
- `core/services/auth.service.ts` (perluas) — `login(identifier,password)`, `refresh()`, `logout()`, signal `user`, `token()`. `PermissionService.load()` (`GET /me/permissions`) dipanggil setelah login & saat boot.
- Interceptor: `authInterceptor` menyisipkan Bearer; tambah alur refresh pada 401 (opsional: coba refresh sekali sebelum logout) di `errorInterceptor`.

### i18n (`AUTH` namespace)
`AUTH.LOGIN.TITLE`, `.IDENTIFIER`, `.PASSWORD`, `.SUBMIT`, `.ERROR_INVALID`, `.ERROR_LOCKED` + `COMMON.*`. IND & ENG sinkron.

---

## 7. ALUR form → API → DATABASE

| Field form | Payload | Kolom DB |
|------------|---------|----------|
| Identifier | `identifier` | dicocokkan ke `users.username`/`users.email` |
| Password | `password` | diverifikasi vs `users.password` (hash) |
| — (server) | — | `sesi_login.refresh_token`, `login_terakhir`, `gagal_login` |

| Aksi | Endpoint | Tulisan DB |
|------|----------|------------|
| Login sukses | `POST /auth/login` | UPDATE `users.login_terakhir`, reset `gagal_login`; INSERT `sesi_login`; INSERT `log_aktivitas` (login) |
| Login gagal | `POST /auth/login` | UPDATE `users.gagal_login`(+1); mungkin set `terkunci_sampai`; tanpa token |
| Refresh | `POST /auth/refresh` | INSERT `sesi_login` baru; UPDATE `dicabut_pada` sesi lama |
| Logout | `POST /auth/logout` | UPDATE `sesi_login.dicabut_pada` |

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- **Fase 0 (01)** — JWT manager, `response`, `JWTAuth`/`Claims`, `utils` password.
- **Anggota (06)** & **RBAC (05)** — `users.anggota_id`/`role_id` FK; `/me/permissions` butuh seed RBAC.

### Acceptance criteria
- [ ] Migrasi `users` + `sesi_login` cocok `.dbml` (partial-unique, FK).
- [ ] Login by **username** dan by **email** sama-sama berhasil (satu query).
- [ ] Kredensial salah → 401 pesan generik; N gagal → akun terkunci (`terkunci_sampai`).
- [ ] Access token memuat `user_id/anggota_id/role_id/perm_version/is_super`; login membuat baris `sesi_login`.
- [ ] Refresh merotasi token (lama dicabut); logout mencabut sesi aktif.
- [ ] `/me` mengembalikan profil + anggota + role; `/me/permissions` mengembalikan `is_super` + daftar izin.
- [ ] Password hash tidak pernah bocor ke klien; `slamctl create-superadmin` menolak bila super admin ada dan bukan endpoint HTTP.
- [ ] `log_aktivitas` mencatat login.
- [ ] Frontend login memanggil `PermissionService.load()` setelah sukses; taste-skill diterapkan.
