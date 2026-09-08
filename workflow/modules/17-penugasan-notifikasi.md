# Modul 17 — Penugasan Peserta + Notifikasi (`jadwal_peserta`, `notifikasi`)

> Sumber otoritatif: `slamteam_db.dbml` (skema) > rancangan Bab 4.4 (penugasan) / Bab 8.6 (notifikasi) > `_shared/api-endpoints.md` §7 & §4 > `_shared/conventions-*.md` > aplikasi lama (tidak pernah). Jika prosa dan `.dbml` berbeda, `.dbml` menang.

---

## 1. Ringkasan & tujuan modul

Modul ini menugaskan anggota ke jadwal/sesi (**`jadwal_peserta`**) dan mengirim **notifikasi in-app** (**`notifikasi`**) saat penugasan terjadi. **Fase 4**, prasyarat langsung **Jadwal (16)**.

Dua aturan penting:

1. **Penugasan memakai `anggota_id`, BUKAN `user_id`.** Anggota tanpa akun tetap bisa didaftarkan sebagai peserta, tetapi `wajib_absen` otomatis **`false`** untuk mereka — job penutup sesi hanya menandai `alfa` baris `wajib_absen=true`. Aturan ini terlihat di data, bukan tersembunyi di query.
2. **Notifikasi adalah fan-out — satu baris per penerima.** Hanya anggota **berakun** yang menerima. Menugaskan 50 orang → 50 baris `notifikasi`; status baca melekat per orang.

---

## 2. Tabel & kolom

### `jadwal_peserta` (DBML baris 813)

| Kolom | Tipe | Aturan / catatan |
|-------|------|------------------|
| `id` | `bigint` [pk] | |
| `jadwal_id` | `bigint` [not null] | FK → `jadwal.id` (**restrict**). |
| `sesi_id` | `bigint` | Kosong = **berlaku semua sesi** jadwal ini. FK → `jadwal_sesi.id`. |
| `anggota_id` | `bigint` [not null] | FK → `anggota.id` (**restrict**). |
| `peran_peserta` | `varchar(50)` | peserta/pelatih/panitia/pengawas. |
| `wajib_absen` | `boolean` [default true] | **Otomatis false** bila anggota tak berakun. |
| `status_tugas` | `status_tugas` [default `ditugaskan`] | ditugaskan/diterima/ditolak/izin. |
| `ditugaskan_oleh` | `bigint` [not null] | FK → `users.id`. |
| `ditugaskan_pada` | `timestamptz` [default now()] | |
| `direspon_pada` | `timestamptz` | Diisi saat terima/tolak. |
| `keterangan` | `text` | |
| soft-delete | | `is_deleted/deleted_at/deleted_by`. |

Index: `(anggota_id, jadwal_id)`, `(jadwal_id, sesi_id)`.

### `notifikasi` (DBML baris 329)

| Kolom | Tipe | Aturan / catatan |
|-------|------|------------------|
| `id` | `bigint` [pk] | |
| `uuid` | `uuid` [unique, default gen] | Identitas publik. |
| `user_id` | `bigint` [not null] | Penerima. FK → `users.id` (**cascade**). |
| `tipe` | `varchar(50)` [not null] | jadwal_ditugaskan / sesi_dibatalkan / izin_disetujui / … |
| `judul` | `varchar(200)` [not null] | |
| `isi` | `text` | |
| `ikon`/`warna` | `varchar` | info/sukses/peringatan/bahaya. |
| `route` | `varchar(255)` | Tujuan saat diklik. |
| `reff_type`/`reff_id` | polymorphic | Sumber. |
| `prioritas` | `prioritas_notifikasi` [default `normal`] | rendah/normal/tinggi. |
| `is_dibaca` | `boolean` [default false] | |
| `dibaca_pada` | `timestamptz` | |
| `is_diarsipkan` | `boolean` [default false] | |
| `kedaluwarsa_pada` | `timestamptz` | Dibersihkan job. |
| `created_at`/`created_by` | | Pemicu; kosong untuk sistem. |

Index: `(user_id, is_dibaca, created_at)`, `(reff_type, reff_id)`, `kedaluwarsa_pada`, `uuid`.

**Relasi:** `jadwal_peserta.jadwal_id → jadwal.id` [restrict], `.sesi_id → jadwal_sesi.id`, `.anggota_id → anggota.id` [restrict], `.ditugaskan_oleh → users.id`; `notifikasi.user_id → users.id` [cascade].

---

## 3. Endpoint

### Penugasan (api-endpoints §7)

| Method | Path | Permission | Auth | Body |
|--------|------|-----------|------|------|
| `POST` | `/jadwal/{id}/peserta` | `jadwal.assign` | JWT | `{anggota_id, sesi_id, peran_peserta, wajib_absen, keterangan}` |
| `POST` | `/jadwal/{id}/peserta/bulk` | `jadwal.assign` | JWT | `{anggota_ids[], peran_peserta, wajib_absen}` → `{ditugaskan, dilewati}` |
| `GET` | `/jadwal/{id}/peserta` | `jadwal.read` | JWT | — |
| `DELETE` | `/jadwal/{id}/peserta/{anggotaId}` | `jadwal.assign` | JWT | — (soft delete) |
| `PATCH` | `/peserta/{id}/respon` | — (bearer) | JWT | `{status_tugas: diterima\|ditolak, keterangan}` |

`sesi_id:null` = semua sesi. Strategi bulk (rancangan Bab 4.4): daftar `anggota_ids` eksplisit; dukung juga filter **by instansi** (ambil semua anggota satu instansi) dan by jenis/status anggota.

### Notifikasi (api-endpoints §4)

| Method | Path | Permission | Auth |
|--------|------|-----------|------|
| `GET` | `/notifikasi` | — (bearer) | JWT |
| `GET` | `/notifikasi/jumlah-belum-dibaca` | — | JWT → `{jumlah}` |
| `PATCH` | `/notifikasi/{id}/baca` | — | JWT |
| `PATCH` | `/notifikasi/baca-semua` | — | JWT |

---

## 4. Hak akses

| Aksi | SA | Admin | Moderator | User |
|------|----|-------|-----------|------|
| `jadwal.assign` (assign/bulk/remove) | ✔ | ✔ | ✔ | — |
| respon penugasan sendiri | ✔ | ✔ | ✔ | ✔ (bearer) |
| notifikasi sendiri | ✔ | ✔ | ✔ | ✔ (bearer) |

Notifikasi & respon adalah **own-scope**: handler membatasi ke `claims.UserID` / peserta yang bersangkutan.

---

## 5. Kebutuhan BACKEND (Go)

Modul `internal/modules/core/penugasan/` (peserta) + `internal/modules/core/notifikasi/` (atau satu modul `jadwal` yang diperluas). Sediakan **service notifikasi** yang bisa dipanggil modul lain (izin, absensi, kta) untuk fan-out.

### domain / dto
- `JadwalPeserta`, `Notifikasi`. DTO: `AssignReq`, `BulkAssignReq`, `ResponReq{status_tugas oneof=diterima ditolak}`, `NotifikasiResp`.

### repository
- Peserta: `Create`, `BulkCreate` (skip duplikat by `(anggota_id, jadwal_id)`), `List`, `SoftDelete`, `Respon`, `AnggotaBerakun(anggotaID)` (untuk set `wajib_absen`). Notifikasi: `CreateMany` (fan-out), `ListByUser`, `CountUnread`, `MarkRead`, `MarkAllRead`.

### service
- **Assign:** validasi anggota ada; **set `wajib_absen=false` bila anggota tak berakun**; simpan; **fan-out notifikasi** `tipe=jadwal_ditugaskan` ke `user_id` anggota (jika berakun) dengan `route` ke jadwal/sesi. Update counter `jadwal_sesi.jml_ditugaskan`.
- **Bulk:** iterasi `anggota_ids` (atau ekspansi by-instansi); kembalikan `{ditugaskan, dilewati}`.
- **Respon:** hanya peserta bersangkutan; set `status_tugas` + `direspon_pada`; notifikasi opsional ke penugas.
- **Notifikasi:** list/unread/mark-read hanya untuk `claims.UserID`. Log `log_aktivitas` untuk assign/remove.

### handler / main / router
- Rute peserta di grup jadwal (`RequirePermission("jadwal.assign")`), `respon` bearer-only. Rute notifikasi bearer-only. Daftar di router.

### migrations
- `jadwal_peserta` + `notifikasi` **persis `.dbml`** (enum `status_tugas`/`prioritas_notifikasi`; index `(user_id,is_dibaca,created_at)`; FK cascade `notifikasi.user_id`). **Setelah** `jadwal_sesi`, `anggota`, `users`.

### edge cases
- Assign anggota tak berakun → `wajib_absen=false` paksa. Duplikat assign → dilewati. Respon oleh non-peserta → 403/404. Notifikasi milik user lain → 404. Anggota tanpa akun → tidak ada notifikasi.

---

## 6. Kebutuhan FRONTEND (Angular)

**WAJIB taste-skill.**

### Komponen
- `pages/jadwal/peserta-panel.*` — daftar peserta jadwal + **penugasan tunggal** (picker anggota) dan **bulk** (multi-select / filter by instansi), kolom peran + `wajib_absen` (readonly bila anggota tak berakun) + `status_tugas` badge.
- `pages/jadwal/respon.*` (untuk peserta) — tombol **Terima/Tolak** penugasan.
- `layouts/topbar/notif-bell.*` — lonceng dengan badge `jumlah-belum-dibaca` (poll/refresh), dropdown daftar notifikasi, mark-read & mark-all-read, klik → navigate `route`.

### Service, i18n
- `penugasan.service.ts`, `notifikasi.service.ts`. i18n `PESERTA` + `NOTIF` namespaces.

### Gating
- Penugasan `*hasPermission="'jadwal.assign'"`; notifikasi & respon bearer (tanpa gate izin).

---

## 7. ALUR form → API → DATABASE

| Aksi | Endpoint | Tulisan DB |
|------|----------|------------|
| Tugaskan 1 | `POST /jadwal/{id}/peserta` | INSERT `jadwal_peserta` (+`wajib_absen` sesuai akun); INSERT `notifikasi` (bila berakun); UPDATE `jml_ditugaskan`; `log_aktivitas` |
| Tugaskan bulk | `POST /jadwal/{id}/peserta/bulk` | INSERT banyak `jadwal_peserta` + `notifikasi` |
| Terima/Tolak | `PATCH /peserta/{id}/respon` | UPDATE `status_tugas` + `direspon_pada` |
| Baca notif | `PATCH /notifikasi/{id}/baca` | UPDATE `is_dibaca=true, dibaca_pada` |

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- **Jadwal (16)** + **Anggota (06)** + **User/auth (04)**. Notifikasi butuh `users`.

### Acceptance criteria
- [ ] Migrasi cocok `.dbml`; `notifikasi.user_id` cascade.
- [ ] Assign anggota tak berakun → `wajib_absen=false` otomatis; tidak ada notifikasi untuknya.
- [ ] Bulk mengembalikan `{ditugaskan, dilewati}`; duplikat dilewati.
- [ ] Penugasan memicu fan-out `notifikasi` (satu baris per penerima berakun) dan menaikkan `jml_ditugaskan`.
- [ ] `PATCH /peserta/{id}/respon` hanya oleh peserta; notifikasi/list hanya milik `claims.UserID`.
- [ ] `jadwal.assign` diblok untuk User (403); respon/notifikasi bearer.
- [ ] Lonceng notifikasi menampilkan badge + list; taste-skill diterapkan.
