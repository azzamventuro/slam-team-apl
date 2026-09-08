# Modul 23 — Master Data Artikel (`artikel`)

> Sumber otoritatif: `slamteam_db.dbml` (skema) > rancangan Bab 8.1 / Bab 12.1 (Fase 8) > `_shared/api-endpoints.md` §13 & §12 > `_shared/conventions-*.md`. Jika prosa dan `.dbml` berbeda, `.dbml` menang.

---

## 1. Ringkasan & tujuan modul

**Artikel** adalah blog/berita klub. Modul ke-12 (`artikel`), grup **Konten**, **Fase 8**. Admin CRUD (`/artikel`) + read publik (`/public/artikel`, list + detail via `?slug=`).

> **Catatan skema penting:** kolom status pada `artikel` bernama **`status_kegiatan` (varchar)** — sama seperti `kegiatan`, TAPI di sini bertipe `varchar(50)` (bebas, konvensi `draft`/`terbit`). Ikuti nama `.dbml` apa adanya; jangan ganti jadi `status_artikel`.

---

## 2. Tabel & kolom

### `artikel` (DBML baris 1022)

| Kolom | Tipe | Aturan / catatan |
|-------|------|------------------|
| `id` | `bigint` [pk] | |
| `user_id` | `bigint` [not null] | Penulis. FK → `users.id`. Dari Claims. |
| `kode` | `varchar(50)` | Basis `slug`. |
| `judul` | `varchar(200)` | |
| `gambar_highlight_file_id` | `bigint` | FK → `mst_file.id`. |
| `konten` | `text` | |
| `status_kegiatan` | `varchar(50)` | **varchar** — `draft`/`terbit`. |
| soft-delete + audit | | |

**Relasi:** `artikel.user_id → users.id`; `artikel.gambar_highlight_file_id → mst_file.id`.

---

## 3. Endpoint

| Method | Path | Permission | Auth | Body |
|--------|------|-----------|------|------|
| `GET` | `/artikel` | `artikel.read` | JWT | list+filter+paginate |
| `GET` | `/artikel/{id}` | `artikel.read` | JWT | — |
| `POST` | `/artikel` | `artikel.create` | JWT | `CreateArtikelReq` |
| `PUT` | `/artikel/{id}` | `artikel.update` | JWT | `UpdateArtikelReq` |
| `DELETE` | `/artikel/{id}` | `artikel.delete` | JWT | — (soft delete) |
| `GET` | `/public/artikel` | — | public | list terbit + detail `?slug=` |

`CreateArtikelReq`: `judul required,max=200`, `kode max=50`, `konten`, `status_kegiatan oneof=draft terbit`, `gambar_highlight_file_id omitempty,gt=0`. `user_id` dari Claims. `slug` diturunkan dari `kode`/`judul`.

---

## 4. Hak akses (rancangan Bab 3.5)

| Modul | SA | Admin | Moderator | User | Guest |
|-------|----|-------|-----------|------|-------|
| `artikel` | CRUD | CRUD | CRUD | R | R (public) |

`artikel.create/read/update/delete`, cakupan `semua`.

---

## 5. Kebutuhan BACKEND (Go)

Modul `internal/modules/core/artikel/`.

- `Artikel` map `artikel` (`StatusKegiatan string`). CRUD standar (join `gambar_uuid`, `q` ILIKE `judul`, sort `judul,created_at`). `user_id=claims.UserID`. Public: `status_kegiatan='terbit'`; detail by `slug` (dari `kode`/`judul`). Log create/update/delete.
- Migrasi `artikel` **persis `.dbml`** (kolom status = `status_kegiatan varchar(50)`). **Setelah** `users`, `mst_file`.

### edge cases
- `status_kegiatan` selain draft/terbit → 422. Public detail slug tak ada → 404. Draft tak muncul publik.

---

## 6. Kebutuhan FRONTEND (Angular)

**WAJIB taste-skill.**

### Komponen (`pages/artikel/`)
- `artikel-list.*` — tabel (Judul, Status, aksi) + filter.
- `artikel-form.*` — form dengan **editor konten** (textarea atau rich-text ringan; **hindari dependency berat**), upload highlight, status select. Halaman publik blog + detail dirender modul **Landing (24)**.

### Service, i18n, gating
- `artikel.service.ts`. i18n `ARTIKEL`. `data:{permission:'artikel.read'}`; C/U/D `*hasPermission`.

---

## 7. ALUR form → API → DATABASE

| Aksi | Endpoint | Tulisan DB |
|------|----------|------------|
| Simpan | `POST /artikel` | INSERT `artikel` (user_id=Claims, status_kegiatan) + log |
| Publik | `GET /public/artikel` | SELECT `status_kegiatan='terbit'` (list/detail slug) |

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- **User/auth (04)**, file layer, RBAC.

### Acceptance criteria
- [ ] Migrasi cocok `.dbml` (kolom status = **`status_kegiatan varchar`**, bukan `status_artikel`).
- [ ] CRUD envelope; `user_id` dari Claims; `slug` dari kode/judul.
- [ ] `/public/artikel` hanya `terbit`; detail via `?slug=`.
- [ ] C/U/D diblok untuk User (403); read publik untuk Guest.
- [ ] `log_aktivitas`; taste-skill diterapkan.
