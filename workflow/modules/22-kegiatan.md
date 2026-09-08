# Modul 22 — Master Data Kegiatan (`kegiatan`)

> Sumber otoritatif: `slamteam_db.dbml` (skema) > rancangan Bab 8.1 / Bab 12.1 (Fase 8) > `_shared/api-endpoints.md` §13 & §12 > `_shared/conventions-*.md`. Jika prosa dan `.dbml` berbeda, `.dbml` menang.

---

## 1. Ringkasan & tujuan modul

**Kegiatan** adalah aktivitas/acara klub untuk **landing page publik** dan referensi jadwal. Modul ke-9 dari 19 (`kegiatan`), grup **Konten**, **Fase 8**. Dua permukaan: **admin CRUD** (`/kegiatan`) dan **read publik** (`/public/kegiatan`). Direferensikan `jadwal.kegiatan_id`.

---

## 2. Tabel & kolom

### `kegiatan` (DBML baris 1003)

| Kolom | Tipe | Aturan / catatan |
|-------|------|------------------|
| `id` | `bigint` [pk] | |
| `user_id` | `bigint` [not null] | Penulis. FK → `users.id`. Diisi dari Claims, bukan form. |
| `kode` | `varchar(50)` | |
| `judul` | `varchar(200)` | |
| `gambar_highlight_file_id` | `bigint` | FK → `mst_file.id`. |
| `konten` | `text` | |
| `tanggal_mulai` | `date` | |
| `tanggal_selesai` | `date` | |
| `status_kegiatan` | `int` | Konvensi: 0 draft / 1 terbit (int). |
| soft-delete + audit | | |

**Relasi:** `kegiatan.user_id → users.id`; `kegiatan.gambar_highlight_file_id → mst_file.id`; ← `jadwal.kegiatan_id`.

---

## 3. Endpoint

| Method | Path | Permission | Auth | Body |
|--------|------|-----------|------|------|
| `GET` | `/kegiatan` | `kegiatan.read` | JWT | list+filter+paginate |
| `GET` | `/kegiatan/{id}` | `kegiatan.read` | JWT | — |
| `POST` | `/kegiatan` | `kegiatan.create` | JWT | `CreateKegiatanReq` |
| `PUT` | `/kegiatan/{id}` | `kegiatan.update` | JWT | `UpdateKegiatanReq` |
| `DELETE` | `/kegiatan/{id}` | `kegiatan.delete` | JWT | — (soft delete) |
| `GET` | `/public/kegiatan` | — | public | hanya `status_kegiatan` terbit (whitelist) |

`CreateKegiatanReq`: `judul required,max=200`, `kode max=50`, `konten`, `tanggal_mulai/selesai`, `status_kegiatan oneof=0 1`, `gambar_highlight_file_id omitempty,gt=0`. `user_id` dari Claims. Resp + `gambar_uuid`.

---

## 4. Hak akses (rancangan Bab 3.5)

| Modul | SA | Admin | Moderator | User | Guest |
|-------|----|-------|-----------|------|-------|
| `kegiatan` | CRUD | CRUD | CRUD | R | R (public) |

Permission `kegiatan.create/read/update/delete`, cakupan `semua`. Guest baca via `/public/kegiatan`.

---

## 5. Kebutuhan BACKEND (Go)

Modul `internal/modules/core/kegiatan/`.

### domain / dto / repository / service / handler
- `Kegiatan` map `kegiatan` (`StatusKegiatan int`, `GambarHighlightFileID *int64`). CRUD standar (join `gambar_uuid`, whitelist sort `judul,tanggal_mulai,created_at`, `q` ILIKE `judul`). `user_id=claims.UserID` saat create. Public read: query `status_kegiatan=1`, kembalikan whitelist. Log create/update/delete.

### main / router / migrations
- `SetupRoutes` `RequirePermission("kegiatan.<aksi>")` + rute publik tanpa JWT. `kegiatan` **persis `.dbml`** (FK `user_id`, `gambar_highlight_file_id`). **Setelah** `users`, `mst_file`.

### edge cases
- `status_kegiatan` selain 0/1 → 422. Public tidak menampilkan draft. Gambar id tak ada → 422.

---

## 6. Kebutuhan FRONTEND (Angular)

**WAJIB taste-skill.**

### Komponen (`pages/kegiatan/`)
- `kegiatan-list.*` — tabel (Judul, Tanggal, Status badge, aksi); filter `q`+status.
- `kegiatan-form.*` — form (judul, konten textarea, `gambar_highlight` upload, rentang tanggal, status select). Bagian publik dirender modul **Landing (24)**.

### Service, i18n, gating
- `kegiatan.service.ts`. i18n `KEGIATAN`. `data:{permission:'kegiatan.read'}`; C/U/D `*hasPermission`.

---

## 7. ALUR form → API → DATABASE

| Field | Payload | Kolom |
|-------|---------|-------|
| Judul/Konten | `judul`,`konten` | idem |
| Gambar (upload→id) | `gambar_highlight_file_id` | idem |
| — (Claims) | — | `user_id` |

| Aksi | Endpoint | Tulisan DB |
|------|----------|------------|
| Simpan | `POST /kegiatan` | INSERT `kegiatan` (user_id=Claims) + log |
| Publik | `GET /public/kegiatan` | SELECT `status_kegiatan=1` (whitelist) |

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- **User/auth (04)** (`user_id`), file layer, RBAC.

### Acceptance criteria
- [ ] Migrasi cocok `.dbml`; `user_id` FK.
- [ ] CRUD envelope; `user_id` dari Claims (bukan form).
- [ ] `/public/kegiatan` hanya terbit, tanpa data privat.
- [ ] C/U/D diblok untuk User (403), read publik untuk Guest.
- [ ] `log_aktivitas`; taste-skill diterapkan.
