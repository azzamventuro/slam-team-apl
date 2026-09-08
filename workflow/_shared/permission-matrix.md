# SLAM Team — Permission Matrix (RBAC Seed)

Sumber otoritatif: rancangan **Bab 3.5**. Dokumen ini mereproduksi matriks izin dasar (SEED) yang di-*insert* ke tabel `role_permission` saat migrasi Fase 1. **Semua nilai di sini adalah SEED yang bisa diedit runtime oleh Super Admin** lewat halaman Hak Akses (`PUT /roles/{id}/permissions`). Bila prosa dan `.dbml` berbeda, `.dbml` menang.

## Peran (roles) & level

| Role | Level | Catatan |
|---|---:|---|
| Super Admin | 0 | `is_super = true` → **BYPASS semua pengecekan permission**. Baris matriks di bawah untuk SA hanya deskriptif. |
| Admin | 10 | |
| Moderator | 20 | |
| User | 30 | Anggota pemegang akun. |
| Guest | 99 | Tidak login / publik. |

Level lebih kecil = lebih tinggi. Anti-eskalasi: tidak bisa memberi izin yang tidak dimiliki; tidak bisa mengedit role dengan level ≤ level sendiri; Super Admin terakhir tidak bisa dihapus.

## Notasi

- **C R U D** = Create / Read / Update / Delete. Delete = *soft delete* (`is_deleted`).
- **`-`** = tidak ada akses sama sekali.
- Setiap izin dikodekan `modul.aksi` (mis. `anggota.create`, `absensi.override`) dan dicek backend via middleware `RequirePermission("modul.aksi")`. Menyembunyikan tombol di UI **bukan** keamanan — setiap izin UI wajib punya cek handler Go yang cocok.

## Dimensi `cakupan` (scope)

Kolom `cakupan` di `role_permission` membatasi **cakupan data** dari sebuah aksi, bukan boleh/tidaknya aksi:

| cakupan | Arti |
|---|---|
| `semua` | Berlaku ke seluruh baris data lintas instansi/pemilik. |
| `instansi_sendiri` | Hanya baris milik instansi yang sama dengan anggota si aktor (mis. moderator unit/instansinya). |
| `milik_sendiri` | Hanya baris milik si aktor sendiri (mis. absensi/izin/file dirinya). |

Bila sebuah baris matriks tidak menyebut cakupan, defaultnya `semua`. Cakupan **hanya** relevan pada aksi yang datanya berdimensi kepemilikan (baca/ubah/hapus data orang lain vs. data sendiri).

### Dua baris "sendiri" (contoh cakupan yang membedakan role)

1. **`absensi.read`** — SA / Admin / Moderator = `semua` (lihat absensi semua orang); **User = `milik_sendiri`** (hanya lihat absensinya sendiri).
2. **`file.delete`** — SA / Admin = `semua` (hapus file siapa pun); **Moderator / User = `milik_sendiri`** (hanya hapus file yang dia unggah).

---

## Tabel 1 — Modul Master Data & Konten

Semua CRUD di bawah bercakupan `semua` kecuali dinyatakan lain.

| Modul (`kode`) | Grup | Super Admin | Admin | Moderator | User | Guest |
|---|---|---|---|---|---|---|
| `anggota` | Master | CRUD | CRUD | CRUD | R | - |
| `admin` | Master | CRUD | R | R | R | - |
| `moderator` | Master | CRUD | CRUD | R | R | - |
| `user` | Master | CRUD | CRUD | CRU | R | - |
| `instansi` | Master | CRUD | CRUD | R | R | - |
| `unit` | Master | CRUD | CRUD | CRUD | CRUD | - |
| `prestasi` | Master | CRUD | CRUD | CRUD | R | R |
| `inorga` | Master | CRUD | CRUD | R | R | - |
| `kegiatan` | Konten | CRUD | CRUD | CRUD | R | R |
| `artikel` | Konten | CRUD | CRUD | CRUD | R | R |
| `medsos` | Konten | CRUD | CRUD | CRUD | CRUD | R |
| `profile_club` | Konten | CRUD | CRUD | R | R | R |
| `dokumen` | Sistem | CRUD | CRUD | CRUD | R | - |

Catatan:
- `admin` / `moderator` / `user` adalah **surface manajemen** di atas tabel `users` + `user_role` yang difilter berdasarkan role — bukan tabel terpisah.
- `unit` memberi CRUD ke semua role login (User pun boleh kelola unit).

---

## Tabel 2 — Modul Operasional & Sistem (dengan aksi khusus)

Baris aksi khusus memakai kode `modul.aksi`; kolom cakupan diberi bila berbeda antar-role.

| Izin (`modul.aksi`) | Super Admin | Admin | Moderator | User | Guest |
|---|---|---|---|---|---|
| **`lokasi`** (CRUD) | CRUD | CRUD | R | R | - |
| **`jadwal`** (CRUD) | CRUD | CRUD | CRUD | R | - |
| `jadwal.assign` | ✔ | ✔ | ✔ | - | - |
| `jadwal.batal_sesi` | ✔ | ✔ | ✔ | - | - |
| `absensi.create` | ✔ | ✔ | ✔ | ✔ | - |
| `absensi.read` | ✔ `semua` | ✔ `semua` | ✔ `semua` | ✔ `milik_sendiri` | - |
| `absensi.override` | ✔ | ✔ | - | - | - |
| `absensi.delete` | ✔ | - | - | - | - |
| `absensi.export` | ✔ | ✔ | ✔ | - | - |
| `izin.create` | ✔ | ✔ | ✔ | ✔ | - |
| `izin.approve` | ✔ | ✔ | ✔ | - | - |
| `kta.create` | ✔ | ✔ | - | - | - |
| `kta.print` | ✔ | ✔ | - | - | - |
| `kta.cabut` | ✔ | ✔ | - | - | - |
| `file.create` | ✔ | ✔ | ✔ | ✔ | - |
| `file.delete` | ✔ `semua` | ✔ `semua` | ✔ `milik_sendiri` | ✔ `milik_sendiri` | - |
| `hak_akses` (kelola role & izin) | ✔ | - | - | - | - |

Catatan:
- `absensi.delete` dan `hak_akses` **Super Admin only**.
- `file.create` tersedia untuk semua role yang login (semua modul mengunggah lewat file layer Fase 0).
- Untuk `jadwal`, aksi CRUD dasar dan aksi khusus (`assign`, `batal_sesi`) dipisah agar User bisa membaca jadwal tanpa bisa menugaskan/membatalkan sesi.

---

## Cara pakai saat seeding (Fase 1)

1. Seed `mst_modul` (19 modul) lalu `permission` (baris `modul.aksi`).
2. Seed 5 `role` di atas beserta `level` dan `is_super`.
3. Insert `role_permission` sesuai kedua tabel di atas, isi kolom `cakupan` (`semua` bila tak disebut).
4. Setiap perubahan runtime lewat halaman Hak Akses harus **bump `perm_version`** agar JWT/klien memuat ulang izin (cache per-request `perm:role:{id}`).
