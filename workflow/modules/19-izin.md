# Modul 19 — Izin & Pulang Cepat (`izin` / `absensi_izin`)

> Sumber otoritatif: `slamteam_db.dbml` (skema) > rancangan Bab 5.1/5.4 (izin sebagai status) > `_shared/api-endpoints.md` §9 > `_shared/conventions-*.md` > aplikasi lama (tidak pernah). Jika prosa dan `.dbml` berbeda, `.dbml` menang.

---

## 1. Ringkasan & tujuan modul

**Izin** adalah pengajuan izin/sakit/dinas/pulang_cepat yang terikat ke sebuah **sesi** (`absensi_izin`). Modul ke-18 (`izin`), grup **Operasional**, **Fase 5**. Izin yang **disetujui** menjadi `status_kehadiran` (`izin`/`sakit`/`dinas`) pada absensi sehingga anggota **tidak ditandai alfa**.

Poin penting: pengaju adalah anggota berakun (`izin.create` untuk semua role login); persetujuan oleh SA/Admin/Moderator (`izin.approve`). Lampiran (mis. surat dokter) adalah **berkas privat**.

---

## 2. Tabel & kolom

### `absensi_izin` (DBML baris 929)

| Kolom | Tipe | Aturan / catatan |
|-------|------|------------------|
| `id` | `bigint` [pk] | |
| `sesi_id` | `bigint` [not null] | FK → `jadwal_sesi.id` (**restrict**). |
| `jadwal_id` | `bigint` | FK → `jadwal.id`. |
| `anggota_id` | `bigint` [not null] | FK → `anggota.id` (**restrict**). |
| `jenis` | `jenis_izin` [not null] | izin / sakit / dinas / pulang_cepat. |
| `alasan` | `text` [not null] | **Wajib**. |
| `waktu_pulang_diminta` | `time` | **Hanya** untuk `jenis=pulang_cepat`. |
| `lampiran_file_id` | `bigint` | FK → `mst_file.id`. **PRIVAT** (surat dokter). |
| `status` | `status_izin` [default `menunggu`] | menunggu/disetujui/ditolak. |
| `diproses_oleh` | `bigint` | FK → `users.id`. |
| `diproses_pada` | `timestamptz` | |
| `catatan_peninjau` | `text` | Wajib saat tolak. |
| soft-delete + audit | | |

**Index:** `(sesi_id, anggota_id)`, `(status, created_at)`.

**Relasi:** `absensi_izin.sesi_id → jadwal_sesi.id` [restrict], `.anggota_id → anggota.id` [restrict], `.lampiran_file_id → mst_file.id`, `.diproses_oleh → users.id`; ← `absensi.izin_id`.

---

## 3. Endpoint (api-endpoints §9)

| Method | Path | Permission | Auth | Body |
|--------|------|-----------|------|------|
| `POST` | `/izin` | `izin.create` | JWT | `{sesi_id, jenis, alasan, waktu_pulang_diminta, lampiran_file_id}` |
| `GET` | `/izin` | `izin.read` | JWT | cakupan own vs all (filter `status`, `sesi_id`) |
| `PATCH` | `/izin/{id}/approve` | `izin.approve` | JWT | — |
| `PATCH` | `/izin/{id}/tolak` | `izin.approve` | JWT | `{catatan_peninjau}` |

`waktu_pulang_diminta` hanya untuk `jenis=pulang_cepat`; `lampiran_file_id` dari `POST /files` sebelumnya.

---

## 4. Hak akses (rancangan Bab 3.5)

| Izin | SA | Admin | Moderator | User |
|------|----|-------|-----------|------|
| `izin.create` | ✔ | ✔ | ✔ | ✔ |
| `izin.read` | `semua` | `semua` | `semua` | `milik_sendiri` |
| `izin.approve` | ✔ | ✔ | ✔ | — |

Cakupan `izin.read`/create ditegakkan service: User hanya `anggota_id=claims.AnggotaID`.

---

## 5. Kebutuhan BACKEND (Go)

Modul `internal/modules/core/izin/`.

### domain / dto
- `Izin` map `absensi_izin`. `CreateIzinReq{ sesi_id required; jenis oneof=izin sakit dinas pulang_cepat; alasan required; waktu_pulang_diminta omitempty; lampiran_file_id omitempty,gt=0 }`, `TolakReq{catatan_peninjau required}`, `IzinResp`.

### repository
- `Create`, `FindByID`, `List` (cakupan + filter status/sesi), `Approve`, `Tolak`. `is_deleted=false`.

### service
- **Create:** sesi ada; `waktu_pulang_diminta` wajib **hanya** bila `jenis=pulang_cepat` (tolak bila diisi untuk jenis lain); cakupan `milik_sendiri` paksa `anggota_id=claims.AnggotaID`. Validasi `lampiran_file_id` bila ada. Log.
- **Approve:** `status=disetujui`, `diproses_oleh/pada`. **Sinkron ke absensi:** buat/perbarui baris `absensi` dengan `status_kehadiran` sesuai `jenis` (izin/sakit/dinas) dan `izin_id` (agar anggota tak jadi alfa) — atau job penutup sesi membaca izin disetujui. Notifikasi `izin_disetujui`.
- **Tolak:** `status=ditolak` + `catatan_peninjau` (wajib). Notifikasi `izin_ditolak`.

### handler / main / router
- `RequirePermission("izin.<aksi>")` (`approve`/`tolak` → `izin.approve`). Daftar di router.

### migrations
- `absensi_izin` **persis `.dbml`** (enum `jenis_izin`/`status_izin`; FK restrict). **Setelah** `jadwal_sesi`, `anggota`. **Sebelum/berbarengan** `absensi` (relasi `absensi.izin_id`).

### edge cases
- `pulang_cepat` tanpa `waktu_pulang_diminta` → 422; jenis lain dengan waktu → 422. Tolak tanpa `catatan_peninjau` → 422. Approve izin yang sudah diproses → tolak. User melihat izin orang lain → 404 (cakupan).

---

## 6. Kebutuhan FRONTEND (Angular)

**WAJIB taste-skill.** Warna jenis izin = info (design-tokens).

### Komponen (`pages/izin/`)
- `izin-form.*` — ajukan izin: pilih **sesi**, `jenis` (select), `alasan` (textarea), `waktu_pulang_diminta` (muncul hanya saat `pulang_cepat`), unggah **lampiran** (opsional) via `FileService`.
- `izin-list.*` — daftar izin (User = miliknya; admin = semua) + filter status.
- `izin-approval.*` (untuk approver) — antrian `menunggu` dengan **Setujui/Tolak** (+`catatan_peninjau`).

### Service, i18n
- `izin.service.ts`. i18n `IZIN` (`FORM.JENIS/ALASAN/WAKTU_PULANG/LAMPIRAN`, `STATUS.*`, `APPROVE/REJECT`, `REJECT_NOTE`).

### Gating
- Ajukan `*hasPermission="'izin.create'"`; approve/tolak `*hasPermission="'izin.approve'"`.

---

## 7. ALUR form → API → DATABASE

| Aksi | Endpoint | Tulisan DB |
|------|----------|------------|
| Unggah lampiran | `POST /files` | INSERT `mst_file` (privat) |
| Ajukan izin | `POST /izin` | INSERT `absensi_izin` (status=menunggu) + `log_aktivitas` |
| Setujui | `PATCH /izin/{id}/approve` | UPDATE status=disetujui; sinkron `absensi` (izin_id); notifikasi |
| Tolak | `PATCH /izin/{id}/tolak` | UPDATE status=ditolak + catatan; notifikasi |

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- **Jadwal/Sesi (16)** + **Anggota (06)** + file layer. Terhubung ke **Absensi (18)** via `absensi.izin_id`.

### Acceptance criteria
- [ ] Migrasi cocok `.dbml`; enum `jenis_izin`/`status_izin`.
- [ ] `waktu_pulang_diminta` hanya valid untuk `pulang_cepat`.
- [ ] Approve → status disetujui + anggota tidak jadi alfa (sinkron absensi); Tolak butuh `catatan_peninjau`.
- [ ] `izin.approve` diblok untuk User (403); User hanya melihat izin sendiri (cakupan).
- [ ] Lampiran privat via `/files`; `log_aktivitas` + notifikasi tertulis.
- [ ] taste-skill diterapkan.
