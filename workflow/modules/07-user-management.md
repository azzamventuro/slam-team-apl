# 07 · Manajemen User (modul `admin` / `moderator` / `user`)

> **Modul key:** `user-management` · **nn:** 07 · **Fase:** 2 (Anggota + master data)
> **Grup mst_modul:** Master Data (kode `admin`, `moderator`, `user`)
> **Tabel inti:** `users`, `user_role`, `mst_role` (baca), `anggota` (baca)
> **Sumber otoritatif:** `slamteam_db.dbml` (menang bila berselisih dengan prosa) ·
> rancangan Bab 2/3.3/3.5/9 · `_shared/api-endpoints.md` §13 · `_shared/permission-matrix.md` ·
> `_shared/conventions-api.md` · `_shared/conventions-app.md`.

---

## 1. Ringkasan & tujuan modul

Manajemen User adalah **satu permukaan (surface) manajemen akun** di atas tabel
`users` + `user_role`, yang muncul sebagai **tiga modul** pada daftar 19 modul:
`admin`, `moderator`, `user` (semuanya kelompok **Master Data**). Ketiganya
memakai bentuk CRUD yang identik; yang membedakan hanya **band level peran** yang
dikelola dan **matriks izin** yang menggerbangnya. Tidak ada tabel terpisah untuk
admin/moderator/user — ketiganya adalah `users` yang **difilter menurut level
peran** akun sasaran.

Fungsi modul:

- Membuat akun untuk **anggota yang sudah ada** (tautan `users.anggota_id →
  anggota.id`) dan menetapkan peran lewat `users.role_id` + baris cermin di
  `user_role` (`is_utama = true`).
- Mengubah akun (username/email/timezone/aktif, reset password, ganti peran).
- Menonaktifkan / soft-delete akun.
- Menegakkan **anti-eskalasi** pada penetapan peran (rancangan Bab 3.3): tidak
  bisa menautkan peran yang setara/lebih tinggi dari peran pelaku, tidak bisa
  mengelola user yang peran aktifnya setara/lebih tinggi, dan **super admin tidak
  pernah dibuat lewat HTTP** (hanya `slamctl create-superadmin`).

**TIDAK ADA REGISTRASI MANDIRI** (rancangan Bab 1.3 / catatan tabel `users`).
Baris `users` **hanya** lahir di sini oleh pemegang izin `admin.create` /
`moderator.create` / `user.create`, atau dari CLI untuk super admin pertama.
Modul "Registrasi" pada rancangan awal **tidak dibuat**.

Kenapa dipisah tiga modul: agar matriks Bab 3.5 dapat memberi hak berbeda per
band. Contoh: seorang **Moderator** boleh membuat & mengubah akun **User**
(`user.create`, `user.update`) tetapi **tidak** boleh menghapusnya (`user.delete`
tidak diberikan ke Moderator) dan **tidak** boleh menyentuh akun **Admin** sama
sekali (`admin` hanya `R` untuk non-SA).

---

## 2. Tabel & kolom (dari `.dbml`)

### 2.1 `users` — tabel inti (tulis)

| Kolom | Tipe | Aturan penting |
|---|---|---|
| `id` | bigint identity | PK. |
| `anggota_id` | bigint | **NOT NULL**. FK → `anggota.id` (`ON DELETE RESTRICT`). Setiap user adalah anggota; sebaliknya tidak. |
| `username` | varchar(50) | **NOT NULL**. **UNIQUE partial `WHERE is_deleted = false`** (username bekas bisa dipakai lagi). |
| `email` | varchar(150) | **NOT NULL**. **UNIQUE partial `WHERE is_deleted = false`**. |
| `password` | varchar(255) | Hash **bcrypt** (`pkg/utils.HashPassword`). **Tidak pernah** dikirim ke klien. |
| `role_id` | bigint | **NOT NULL**. FK → `mst_role.id` (`ON DELETE RESTRICT`). **Sumber kebenaran peran aktif** — dibaca JWT saat login. |
| `timezone` | varchar(50) | default `Asia/Jakarta` (nama IANA). |
| `is_aktif` | boolean | default `true`. Nonaktif = tidak bisa login (dicek modul auth), bukan soft delete. |
| `login_terakhir` | timestamptz | Diisi modul auth, ditampilkan read-only di sini. |
| `password_diubah` | timestamptz | Diisi saat reset password. |
| `gagal_login` | int | default 0. Bisa direset (buka kunci) saat edit. |
| `terkunci_sampai` | timestamptz | Kunci sementara akibat gagal login; edit dapat mengosongkan (buka kunci). |
| `is_deleted` `deleted_at` `deleted_by` | soft delete | DELETE = set `is_deleted=true, deleted_at=now(), deleted_by=<aktor>`. **Bukan** hard delete. |
| `created_at` `created_by` `modified_at` `modified_by` | audit | Diisi otomatis. |

Index (dari `.dbml`): `username` unique partial, `email` unique partial,
`anggota_id`, `role_id`.

Catatan tabel (`.dbml`): namanya **`users`**, bukan `user` (kata bawaan SQL di
PostgreSQL). LOGIN memakai username **ATAU** email (keduanya UNIQUE). Tidak semua
anggota punya akun.

### 2.2 `user_role` — tabel-jembatan peran (tulis, cermin)

| Kolom | Tipe | Aturan |
|---|---|---|
| `id` | bigint identity | PK. |
| `user_id` | bigint | **NOT NULL**. FK → `users.id` (`ON DELETE CASCADE`). |
| `role_id` | bigint | **NOT NULL**. FK → `mst_role.id`. |
| `is_utama` | boolean | default `true`. Peran utama = cermin dari `users.role_id`. |
| `berlaku_sampai` | timestamptz | Opsional (peran sementara). Kosong = permanen. |
| `created_at` `created_by` | audit | |

Index: `(user_id, role_id)` **UNIQUE**. Catatan `.dbml`: *"Satu peran per user
untuk sekarang; kode membaca lewat tabel ini sejak awal agar multi-peran nanti
tanpa migrasi."* → Untuk sekarang **tepat satu baris `is_utama=true` per user**,
`role_id`-nya = `users.role_id`. Keduanya ditulis dalam satu transaksi dan
**wajib konsisten**.

### 2.3 `mst_role` — dibaca (band level & aturan)

Kolom yang dipakai: `id`, `kode`, `nama`, `level` (SA=0, Admin=10, Moderator=20,
User=30, Guest=99 — **angka lebih kecil = lebih tinggi**), `is_super` (bypass
semua izin; **tidak boleh** ditautkan lewat HTTP), `is_sistem`, `is_aktif`,
`is_deleted`. Dipakai untuk: menentukan **band** tiap surface, cek anti-eskalasi,
dan tampilan nama peran.

### 2.4 `anggota` — dibaca (tautan akun)

Kolom yang dipakai saat membuat/menampilkan akun: `id`, `nama_lengkap`,
`nama_panggilan`, `no_induk` (NRA aktif), `instansi_id`, `foto_profil_file_id`,
`status_anggota`. Aturan: hanya anggota **yang belum punya akun** (`is_deleted=false`)
yang boleh ditautkan.

---

## 3. Endpoint

Tiga base route, **satu implementasi handler** yang di-parametrisasi oleh
`surface` (`admin` | `moderator` | `user`). Semua `bearer` (JWTAuth). Cakupan
semua aksi = `semua` (tidak ada `milik_sendiri` untuk modul ini di Bab 3.5).

| Method | Path | Permission | Auth | Kegunaan |
|---|---|---|---|---|
| GET | `/admin` · `/moderator` · `/user` | `<surface>.read` | bearer | List akun band ybs (paginate + filter). |
| GET | `/{surface}/{id}` | `<surface>.read` | bearer | Detail satu akun. |
| POST | `/{surface}` | `<surface>.create` | bearer | Buat akun (tautan anggota + tetapkan peran). |
| PUT | `/{surface}/{id}` | `<surface>.update` | bearer | Ubah akun / reset password / ganti peran / buka kunci. |
| DELETE | `/{surface}/{id}` | `<surface>.delete` | bearer | Soft delete akun. |
| GET | `/{surface}/peran-tersedia` | `<surface>.create` | bearer | Peran yang **boleh ditautkan** aktor pada band ini (isi dropdown). |
| GET | `/{surface}/anggota-tersedia` | `<surface>.create` | bearer | Anggota **tanpa akun** untuk ditautkan (`?q=` pencarian). |

> Dua route `*-tersedia` adalah pembantu read internal modul (bukan route baru di
> peta) — ada karena form Create butuh daftar peran & anggota, sementara `GET
> /roles` digerbangi `hak_akses.read` (**SA-only**) sehingga Admin/Moderator tak
> bisa memakainya. Keduanya digerbangi `<surface>.create`, bukan `hak_akses`.

### 3.1 Band per surface (peran mana yang muncul & bisa dibuat)

Filter list & validasi peran memakai **level** peran sasaran (bukan `kode`, agar
peran dinamis buatan super admin tetap tertampung). Band bawaan:

| Surface | Band level peran sasaran | Peran seed yang masuk |
|---|---|---|
| `admin` | `1 ≤ level ≤ 19` | Admin (10) |
| `moderator` | `20 ≤ level ≤ 29` | Moderator (20) |
| `user` | `30 ≤ level ≤ 98` | User (30) |

Level `0` (Super Admin) dan `99` (Guest) **di luar semua surface**: SA hanya
dibuat via `slamctl` & dikelola di modul Hak Akses; Guest adalah peran semu tanpa
baris `users`. (Band adalah keputusan default yang bisa disetel — lihat catatan
`ponytail:` di service.)

### 3.2 Payload

**`POST /{surface}`** (JSON):
```json
{
  "anggota_id": 12,
  "username": "azzz",
  "email": "achmad@slamteam.id",
  "password": "••••••••",
  "role_id": 3,
  "timezone": "Asia/Jakarta",
  "is_aktif": true
}
```

**`PUT /{surface}/{id}`** (JSON — `password` opsional, hanya bila mengganti;
`role_id` opsional, hanya bila memindah peran; `buka_kunci` opsional):
```json
{
  "username": "azzz",
  "email": "achmad@slamteam.id",
  "password": null,
  "role_id": 3,
  "timezone": "Asia/Jakarta",
  "is_aktif": true,
  "buka_kunci": false
}
```

**Response `data`** (bentuk `UserResp` — **tanpa** `password`):
```json
{
  "id": 5,
  "anggota": { "id": 12, "nama_lengkap": "Achmad", "no_induk": "35731002021",
               "instansi_id": 1, "foto_url": "/api/v1/files/<uuid>/low" },
  "username": "azzz",
  "email": "achmad@slamteam.id",
  "role": { "id": 3, "nama": "Moderator", "level": 20, "is_super": false },
  "timezone": "Asia/Jakarta",
  "is_aktif": true,
  "login_terakhir": "2026-09-07T02:11:00Z",
  "terkunci_sampai": null,
  "created_at": "2026-09-01T00:00:00Z"
}
```

**List** `GET /{surface}?page=1&per_page=20&q=&sort=-created_at&is_aktif=&instansi_id=`
→ `data` array `UserResp`, `meta: { page, per_page, total, total_pages }`.

**`GET /{surface}/peran-tersedia`** → `data: [{ "id": 3, "nama": "Moderator",
"level": 20 }]` (hanya peran band ini dengan `level > level_aktor`, `is_super=false`,
`is_aktif=true`, `is_deleted=false`).

**`GET /{surface}/anggota-tersedia?q=ach`** → `data: [{ "id": 12,
"nama_lengkap": "Achmad", "no_induk": "35731002021" }]` (anggota `is_deleted=false`
yang belum punya baris `users` aktif).

---

## 4. Hak akses (matriks Bab 3.5)

| Modul | Super Admin | Admin | Moderator | User | Guest |
|---|---|---|---|---|---|
| `admin` | CRUD | R | R | R | — |
| `moderator` | CRUD | CRUD | R | R | — |
| `user` | CRUD | CRUD | CR**U** | R | — |

- **Cakupan:** seluruh aksi = `semua` (tidak ada dimensi `milik_sendiri` untuk
  modul ini). Tidak ada penyempitan query per pemilik; penyempitan satu-satunya
  adalah **band level** surface + `is_deleted=false`.
- **`user` Moderator = CRU** → izin `user.create`, `user.read`, `user.update`
  diberikan ke Moderator; `user.delete` **tidak**.
- Super Admin `is_super=true` **bypass** semua cek middleware (baris SA di atas
  hanya deskriptif), tetapi **tetap** tunduk pada aturan bisnis anti-eskalasi
  dan proteksi super admin terakhir (§5.3).
- Seed `role_permission` (band ini) dilakukan di **Fase 1 / modul Hak Akses**;
  modul ini hanya menegakkan cek `<surface>.<aksi>` di setiap route.

---

## 5. Kebutuhan BACKEND (Go)

Folder: `internal/modules/core/usermgmt/` — **package `usermgmt`** (Go tidak
mengizinkan tanda hubung pada nama package; key `user-management` → folder
`usermgmt`).

### 5.1 domain
- `User` (map `users`, `TableName() "users"`), embed `Audit` bersama
  (created/modified/soft-delete) sesuai `conventions-api.md` §6. Field
  `Password` diberi `json:"-"`. Kolom `login_terakhir`, `password_diubah`,
  `gagal_login`, `terkunci_sampai` sebagai `*time.Time`/`int`.
- `UserRole` (map `user_role`, `TableName() "user_role"`).
- `Role` (read-model map `mst_role`, `TableName() "mst_role"`): `id, kode, nama,
  level, is_super, is_aktif, is_deleted`.
- Preload/`Joins` `anggota` + `mst_role` untuk `UserResp`.

### 5.2 dto
- `CreateUserReq`, `UpdateUserReq`, `ListQuery` (embed `PageQuery` standar),
  `UserResp`, `RoleOpsi`, `AnggotaOpsi`. Binding tags mencerminkan §3.2:
  `username` `required,max=50`; `email` `required,email,max=150`; `password`
  `required,min=8` (Create) / `omitempty,min=8` (Update); `role_id`
  `required,gt=0`; `anggota_id` `required,gt=0`; `timezone` `omitempty,timezone`.
- **Jangan** bind langsung ke entity; map DTO → entity di service.

### 5.3 service — aturan bisnis (semua ada di sini, bukan di handler)
Buat/ubah/hapus akun dijalankan dalam **satu transaksi** (`users` + `user_role`
+ `log_aktivitas`).

**Validasi & keunikan**
- `username` & `email` unik antar baris `is_deleted=false` (cek + andalkan unique
  partial index; tabrakan → `ErrConflict`).
- `anggota_id` harus ada, `is_deleted=false`, **belum** punya akun aktif (satu
  akun per anggota — `ponytail: cek di service; tambah partial unique index
  users(anggota_id) WHERE is_deleted=false bila perlu paksa di DB`).
- `role_id` harus ada, `is_aktif=true`, `is_deleted=false`, **`level` di dalam
  band surface**.

**Anti-eskalasi (rancangan Bab 3.3) — ditegakkan di service, bukan hanya UI**
Ambil level aktor dari Claims (JWT: `role_id`→level; `is_super` → perlakukan
level `0`).
- **A (peran yang ditautkan):** `target_role.level > aktor.level`. Aktor tidak
  bisa menautkan peran setara/lebih tinggi dari dirinya. (Menautkan peran =
  memberi seluruh izinnya → ini wajah "tidak bisa memberi izin yang tidak
  dimiliki" untuk modul ini.)
- **B (SA CLI-only):** **tolak** membuat/menaikkan user ke peran `is_super=true`
  lewat HTTP apa pun, tanpa memandang aktor. Super admin hanya via
  `slamctl create-superadmin`.
- **C (proteksi SA terakhir & non-demote):** DELETE/nonaktif user yang perannya
  `is_super=true` ditolak bila itu **super admin aktif terakhir**. Mengganti peran
  user `is_super` (menurunkan) via modul ini juga ditolak (lihat B).
- **D (proteksi sasaran):** untuk UPDATE/DELETE, **peran user sasaran saat ini**
  harus `level > aktor.level`. Aktor tak boleh mengelola akun setara/lebih tinggi.

**Buat (Create)**
1. Validasi + anti-eskalasi (A, B, D-tidak-relevan).
2. `password_hash = utils.HashPassword(req.Password)`.
3. INSERT `users` (anggota_id, username, email, password=hash, role_id, timezone,
   is_aktif, created_by=aktor).
4. INSERT `user_role` (user_id, role_id, is_utama=true, created_by=aktor).
5. INSERT `log_aktivitas` (aktor, modul=`<surface>`, aksi=`buat`,
   reff_type=`users`, reff_id, ringkasan, nilai_baru).
6. Return `UserResp`.

**Ubah (Update)**
- Anti-eskalasi D (sasaran) + A (bila `role_id` berubah) + B.
- Field yang boleh diubah: `username`, `email`, `timezone`, `is_aktif`.
- `password` bila diisi → hash baru, `password_diubah=now()`, `gagal_login=0`.
- `buka_kunci=true` → `terkunci_sampai=NULL`, `gagal_login=0`.
- Bila `role_id` berubah → UPDATE `users.role_id` **dan** sinkron `user_role`
  (update baris `is_utama=true` ke role baru). **Opsional hardening:** cabut
  `sesi_login` user tsb agar token lama (masih membawa `role_id` lama) tidak lagi
  berlaku sampai login ulang — token aktif tidak otomatis ikut berubah karena
  `perm_version` hanya naik saat izin **peran** berubah, bukan saat penetapan
  peran user. Catat `log_aktivitas` (aksi=`ubah`, nilai_lama/nilai_baru).

**Hapus (Delete, soft)**
- Anti-eskalasi D + C (SA terakhir).
- Set `is_deleted=true, deleted_at=now(), deleted_by=aktor`; cabut semua
  `sesi_login` user. `log_aktivitas` aksi=`hapus`.

**List / Detail**
- Filter dasar: `users.is_deleted=false` **AND** `role.level` dalam band surface.
- `q` mencari `username` / `email` / `anggota.nama_lengkap` (ILIKE). Filter
  `is_aktif`, `instansi_id`. `sort` whitelist (`created_at`, `username`,
  `login_terakhir`). Join `anggota`, `mst_role`. Cakupan `semua` → tidak ada
  penyempitan kepemilikan.

**Peran/anggota tersedia** — seperti §3.

### 5.4 repository
GORM: `Create`, `Update` (map field), `SoftDelete`, `FindByID` (preload anggota
+ role), `List` (Count + Find pada query bertingkat sama, cap `per_page` 100),
`ExistsUsername/Email` (partial `is_deleted=false`), `AnggotaHasAccount(anggotaId)`,
`RolesInBand(min,max)`, `AnggotaWithoutAccount(q)`. Terjemahkan
`gorm.ErrRecordNotFound → ErrNotFound`. Semua tulis lewat `tx *gorm.DB` dari
service agar transaksional.

### 5.5 handler
Bind → panggil service dengan `middleware.Claims(c)` + `surface` (dari closure) →
`response.*`. `422` untuk validasi (`validator.Explain`), `response.FromError`
memetakan sentinel (`ErrNotFound`→404, `ErrForbidden`→403, `ErrConflict`→409/422,
lainnya→`Internal`). **Password tidak pernah** masuk response.

### 5.6 main.usermgmt.go — wiring & routes
Satu handler, tiga registrasi (closure `surface`):
```go
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
    for _, s := range []string{"admin", "moderator", "user"} {
        g := rg.Group("/"+s, middleware.JWTAuth(m.jwtMgr))
        g.GET("",  m.perm.Require(s+".read"),   m.h.List(s))
        g.GET("/peran-tersedia",   m.perm.Require(s+".create"), m.h.PeranTersedia(s))
        g.GET("/anggota-tersedia", m.perm.Require(s+".create"), m.h.AnggotaTersedia(s))
        g.GET("/:id",  m.perm.Require(s+".read"),   m.h.Detail(s))
        g.POST("",     m.perm.Require(s+".create"), m.h.Create(s))
        g.PUT("/:id",  m.perm.Require(s+".update"), m.h.Update(s))
        g.DELETE("/:id", m.perm.Require(s+".delete"), m.h.Delete(s))
    }
}
```
Register di `internal/router/router.go`:
`usermgmt.Initialize(db, jwtMgr, permGuard).SetupRoutes(apiV1)`.

### 5.7 migrations & seeders
- **Skema:** `users` & `user_role` sudah dibuat di migrasi Fase 0/1 (RBAC +
  identitas). **Modul ini tidak menambah tabel.** Pastikan migrasi berisi:
  unique **partial** `users(username) WHERE is_deleted=false`, `users(email)
  WHERE is_deleted=false`, unique `user_role(user_id, role_id)`, FK `anggota_id`
  RESTRICT & `role_id` RESTRICT. (Opsional: partial unique `users(anggota_id)
  WHERE is_deleted=false`.)
- **Seeder izin:** baris `mst_permission` `admin.*`/`moderator.*`/`user.*`
  (aksi create/read/update/delete) + `role_permission` sesuai §4 di-seed di
  **Fase 1**. Idempoten (`ON CONFLICT (kode) DO NOTHING`).
- **Tidak ada `AutoMigrate`.**

### 5.8 edge cases
- Membuat akun untuk anggota yang sudah punya akun → `ErrConflict`.
- Username/email tabrakan dengan baris **terhapus** → boleh (partial unique).
- Ganti peran ke luar band surface → `ErrForbidden` (mis. lewat `/user` mencoba
  set peran Admin).
- Aktor Moderator DELETE user → `403` (`user.delete` tak diberikan Mod).
- Menautkan/menaikkan ke peran `is_super` → `403` (aturan B).
- Menghapus SA terakhir → `403` (aturan C).
- Anggota/role tertaut tak bisa di-hard-delete (FK RESTRICT) — konsisten dgn
  soft-delete di sini.

### 5.9 background jobs
Tidak ada. (Kunci akibat gagal login & `login_terakhir` dikelola modul **auth**;
modul ini hanya membaca/menyetelnya via edit.)

---

## 6. Kebutuhan FRONTEND (Angular)

> **Design method (WAJIB):** mulai dengan mengumumkan **"Using
> design-taste-frontend"**, jalankan pre-flight/audit skill, petakan ke
> **Bootstrap 5**, dan **tegakkan token** `workflow/_shared/design-tokens.md`
> (dark default, SLAM red aksen, heading UPPERCASE wide-tracking, fokus ring
> merah, min tap 44px). **Catatan blog-fe:** *"Jika sumber `blog-fe` dipulihkan,
> tiru layout/menu-nya untuk layar ini; jika tidak, ikuti design tokens."*

### 6.1 Pages/components (satu set, di-parametrisasi `surface`)
`pages/user-management/`:
- `user-list.ts/.html/.scss` — tabel akun (list canonical, `toSignal` +
  `switchMap`, filter signal `q/page/perPage/isAktif/instansiId`). Kolom: foto
  (avatar `low` via blob), Nama Anggota + NRA (`no_induk`), Username, Email,
  Peran (badge), Status (`is_aktif` chip success/muted), Login terakhir, aksi.
- `user-form.ts/.html/.scss` — form Create/Edit (Reactive Forms typed).
- `user.service.ts` — wrapper `ApiService`, **surface-aware** path.

Surface ditentukan via `route.data.surface` (+ `route.data.permission`), sehingga
tiga route memakai komponen yang sama:
```ts
// app.routes.ts (di bawah VerticalLayout, canActivate: [authGuard, permissionGuard])
{ path: 'admin',     data: { surface: 'admin',     permission: 'admin.read' },     loadComponent: () => import('./pages/user-management/user-list') },
{ path: 'admin/new', data: { surface: 'admin',     permission: 'admin.create' },   loadComponent: () => import('./pages/user-management/user-form') },
{ path: 'admin/:id', data: { surface: 'admin',     permission: 'admin.update' },   loadComponent: () => import('./pages/user-management/user-form') },
// idem untuk 'moderator' & 'user'
```

### 6.2 Form fields (mirror DTO §3.2)
| Field | Kontrol | Validasi klien (mirror binding) |
|---|---|---|
| `anggota_id` | select cari (dari `GET /{surface}/anggota-tersedia?q=`) — **hanya Create** | `required`. Edit: read-only (tampil nama+NRA). |
| `username` | text | `required`, `maxLength(50)`. |
| `email` | email | `required`, `email`, `maxLength(150)`. |
| `password` | password | Create: `required`, `minLength(8)`. Edit: `minLength(8)` bila diisi (kosong = tetap). |
| `role_id` | select (dari `GET /{surface}/peran-tersedia`) | `required`. Opsi dengan `level ≤ level_aktor` **di-disable** (mirror anti-eskalasi). |
| `timezone` | select IANA (default `Asia/Jakarta`) | opsional. |
| `is_aktif` | switch | default `true`. |
| `buka_kunci` | switch (Edit, tampil bila `terkunci_sampai`) | opsional. |

Pada `422/409` peta `res.errors` (field→pesan) ke control (`setErrors({server}`)).

### 6.3 Service (envelope via `ApiService`)
```ts
@Injectable({ providedIn: 'root' })
export class UserService {
  private api = inject(ApiService);
  list(s: string, q: UserQuery)      { return this.api.get<Page<UserResp>>(`/${s}`, q); }
  detail(s: string, id: number)      { return this.api.get<UserResp>(`/${s}/${id}`); }
  create(s: string, b: CreateUserReq){ return this.api.post<UserResp>(`/${s}`, b); }
  update(s: string, id: number, b: UpdateUserReq){ return this.api.put<UserResp>(`/${s}/${id}`, b); }
  remove(s: string, id: number)      { return this.api.delete<void>(`/${s}/${id}`); }
  peran(s: string)                   { return this.api.get<RoleOpsi[]>(`/${s}/peran-tersedia`); }
  anggota(s: string, q: string)      { return this.api.get<AnggotaOpsi[]>(`/${s}/anggota-tersedia`, { q }); }
}
```

### 6.4 Permission-gating & anti-eskalasi mirror
- Route: `permissionGuard` + `data.permission`.
- Tombol: `*hasPermission="surface + '.create'"`, `.update`, `.delete`
  (bind ekspresi karena `surface` dinamis; gunakan computed string).
- Cermin anti-eskalasi (server tetap otoritatif): dari `AuthService.user()!.role.level`,
  **disable** opsi peran dengan `level ≤ myLevel`, sembunyikan Delete pada baris
  user yang `role.level ≤ myLevel`, dan pada surface `/admin` sembunyikan
  Create/Update/Delete untuk non-SA (mereka hanya `R`).

### 6.5 Foto & file
Avatar anggota privat → fetch blob lewat `FileService.imageUrl(uuid,'low')`
(interceptor menambah Bearer). Revoke object URL saat destroy. Tidak ada upload
di modul ini (foto dikelola modul `anggota`).

### 6.6 i18n keys (`public/i18n/{IND,ENG}.json`, sinkron)
`USER.TITLE.ADMIN` / `.MODERATOR` / `.USER`, `USER.COL.NAMA`, `USER.COL.USERNAME`,
`USER.COL.EMAIL`, `USER.COL.PERAN`, `USER.COL.STATUS`, `USER.COL.LOGIN_TERAKHIR`,
`USER.FORM.ANGGOTA`, `USER.FORM.USERNAME`, `USER.FORM.EMAIL`, `USER.FORM.PASSWORD`,
`USER.FORM.PERAN`, `USER.FORM.TIMEZONE`, `USER.FORM.AKTIF`, `USER.FORM.BUKA_KUNCI`,
`USER.MSG.CREATED`, `USER.MSG.UPDATED`, `USER.MSG.DELETED`,
`USER.HINT.NO_SELF_REGISTER`, `USER.ERR.ANGGOTA_SUDAH_PUNYA_AKUN`,
`USER.ERR.ESKALASI`, plus `COMMON.*`, `VALIDATION.*`.

---

## 7. ALUR form → API → DATABASE (kontrak koherensi)

### 7.1 Pemetaan field (Create)
| Field form (Angular) | Key payload (JSON) | Kolom DB | Tabel |
|---|---|---|---|
| Anggota (select cari) | `anggota_id` | `anggota_id` | `users` |
| Username | `username` | `username` | `users` |
| Email | `email` | `email` | `users` |
| Password | `password` | `password` (→ **bcrypt hash**) | `users` |
| Peran (select) | `role_id` | `role_id` **dan** `role_id` | `users` **&** `user_role` |
| Timezone | `timezone` | `timezone` | `users` |
| Aktif (switch) | `is_aktif` | `is_aktif` | `users` |
| — (server) | — | `is_utama=true` | `user_role` |
| — (server, Claims) | — | `created_by` | `users`, `user_role` |

### 7.2 Aksi → endpoint → tulis DB
| Aksi user (UI) | Endpoint | Tulis DB |
|---|---|---|
| Simpan akun baru | `POST /{surface}` | INSERT `users` (+hash) → INSERT `user_role` (is_utama=true) → INSERT `log_aktivitas`(buat) |
| Simpan perubahan | `PUT /{surface}/{id}` | UPDATE `users` (username/email/timezone/is_aktif; password+password_diubah bila diisi; buka_kunci→terkunci_sampai=NULL,gagal_login=0; role_id bila ganti) → UPDATE `user_role`(is_utama) bila ganti peran → (opsional) revoke `sesi_login` → INSERT `log_aktivitas`(ubah) |
| Nonaktifkan | `PUT /{surface}/{id}` (`is_aktif=false`) | UPDATE `users.is_aktif=false` → `log_aktivitas`(ubah) |
| Hapus akun | `DELETE /{surface}/{id}` | UPDATE `users` set is_deleted=true,deleted_at,deleted_by → revoke `sesi_login` → `log_aktivitas`(hapus) |
| Muat opsi peran | `GET /{surface}/peran-tersedia` | SELECT `mst_role` band, level>aktor, is_super=false |
| Muat opsi anggota | `GET /{surface}/anggota-tersedia?q=` | SELECT `anggota` tanpa baris `users` aktif |
| Buka daftar | `GET /{surface}` | SELECT `users` JOIN `anggota`,`mst_role` WHERE is_deleted=false AND role.level∈band |

**Invarian koherensi:** `users.role_id` == baris `user_role` yang `is_utama=true`
(ditulis dalam satu transaksi). JWT membaca `users.role_id`; RBAC membaca via
`user_role` — keduanya menunjuk peran yang sama.

---

## 8. Dependencies / prasyarat & Acceptance criteria

**Prasyarat**
- **Fase 0:** skeleton Go/Angular, `pkg/utils` bcrypt, `middleware.JWTAuth` +
  Claims (`role_id`, `is_super`, `anggota_id`), `response` (+ `Forbidden`,
  `NotFound`, `FromError`), `sesi_login`, `log_aktivitas`, `slamctl
  create-superadmin`.
- **Fase 1 (hak-akses):** `mst_modul`/`mst_permission`/`mst_role`/`role_permission`
  ter-seed; `middleware.PermGuard.Require`; izin `admin.*`/`moderator.*`/`user.*`
  ter-seed sesuai §4.
- **Fase 2 (anggota):** tabel & modul `anggota` ada (untuk tautan & pencarian).
- Migrasi `users` + `user_role` (partial unique, FK RESTRICT) sudah dijalankan.

**Acceptance criteria (checkable)**
- [ ] `GET /admin|/moderator|/user` mengembalikan hanya akun band yang sesuai
      (`is_deleted=false`), paginate + filter `q/is_aktif/instansi_id`.
- [ ] `POST /{surface}` membuat baris `users` (password ter-hash bcrypt) **dan**
      baris `user_role` `is_utama=true` dengan `role_id` sama, dalam satu transaksi.
- [ ] Response **tidak pernah** memuat `password`.
- [ ] `anggota_id` yang sudah punya akun → `409/422`, tidak membuat baris.
- [ ] Anti-eskalasi A: aktor Admin tak bisa menautkan/memindah ke peran
      level ≤ 10 → `403`.
- [ ] Aturan B: membuat/menaikkan ke peran `is_super=true` via HTTP → `403`.
- [ ] Aturan C: menghapus/menonaktifkan super admin **terakhir** → `403`.
- [ ] Aturan D: Admin mengelola akun Admin/SA lain (level ≤ dirinya) → `403`.
- [ ] `user.delete` Moderator → `403`; `user.create`/`user.update` Moderator → OK.
- [ ] `admin.*` create/update/delete untuk non-SA → `403` (hanya `R`).
- [ ] Ganti peran menyinkronkan `users.role_id` **dan** `user_role`(is_utama).
- [ ] DELETE = soft (`is_deleted=true`) + cabut `sesi_login`; username/email jadi
      bisa dipakai ulang.
- [ ] Setiap tulis mencatat `log_aktivitas` (aktor, modul=`<surface>`, aksi,
      nilai_lama/baru).
- [ ] FE: tombol Create/Update/Delete tersembunyi tanpa izin **dan** endpoint
      tetap menolak bila dipanggil langsung (bukan sekadar disembunyikan).
- [ ] FE: opsi peran `level ≤ level_aktor` di-disable; dark tokens & fokus ring
      merah diterapkan; responsif; IND/ENG sinkron.
