# Modul 15 — Master Data Lokasi (`lokasi`)

> Sumber otoritatif: `slamteam_db.dbml` (skema) > rancangan Bab 5.5 (validasi lokasi) / Bab 4.5 (zona waktu) > `_shared/conventions-*.md` > aplikasi lama (tidak pernah). Jika prosa dan `.dbml` berbeda, `.dbml` menang.

---

## 1. Ringkasan & tujuan modul

**Lokasi** adalah master data tempat latihan/kegiatan dengan **geofence** (titik + radius) dan **zona waktu IANA**. Modul ke-15 (`lokasi`), grup **Operasional**, dikerjakan **Fase 3**.

Dua alasan penting melebihi CRUD biasa:

1. **Menyuplai geofence & timezone ke jadwal.** `jadwal.lokasi_id → mst_lokasi.id` (**restrict**). Saat jadwal dibuat, `latitude/longitude/radius_meter/timezone` di-**snapshot** dari lokasi; absensi memvalidasi jarak (Haversine) terhadap radius yang berlaku (rancangan Bab 5.5).
2. **Zona waktu adalah nama IANA, bukan offset.** `timezone` (mis. `Asia/Jakarta`) dipakai untuk mengonversi jam sesi lokal → UTC dan menampilkan 07:00 WIB (Bab 4.5).

Prasyarat: **file-management (Fase 0)** untuk `foto_file_id`, **RBAC (Fase 1)** untuk `RequirePermission`.

---

## 2. Tabel & kolom

### `mst_lokasi` (DBML baris 717)

| Kolom | Tipe | Aturan / catatan |
|-------|------|------------------|
| `id` | `bigint` [pk] | |
| `kode` | `varchar(50)` [unique, not null] | Partial-unique `WHERE is_deleted=false`. |
| `nama` | `varchar(150)` [not null] | |
| `jenis_lokasi` | `varchar(50)` | mis. lapangan / indoor / sekretariat. |
| `alamat` | `text` | |
| `latitude` | `decimal(10,7)` [not null] | **CHECK BETWEEN -90 AND 90**. |
| `longitude` | `decimal(10,7)` [not null] | **CHECK BETWEEN -180 AND 180**. |
| `radius_meter` | `int` [default 100, not null] | **CHECK > 0**. Dinamis per lokasi. |
| `timezone` | `varchar(50)` [default `Asia/Jakarta`, not null] | Nama IANA. |
| `foto_file_id` | `bigint` | FK → `mst_file.id` (nullable). |
| `keterangan` | `text` | |
| `is_aktif` | `boolean` [default true] | |
| soft-delete + audit | | `is_deleted/deleted_at/deleted_by` + `created_*`/`modified_*`. |

**Relasi:** `mst_lokasi.foto_file_id → mst_file.id`; ← `jadwal.lokasi_id` [**restrict**].

**Aturan lintas-modul:** validasi rentang koordinat & `radius_meter>0` di DB (CHECK) DAN service. `timezone` divalidasi terhadap `time.LoadLocation`. Hapus ditolak bila masih dipakai jadwal.

---

## 3. Endpoint

| Method | Path | Permission | Auth | Body | Response `data` |
|--------|------|-----------|------|------|-----------------|
| `GET` | `/lokasi` | `lokasi.read` | JWT | query `page,per_page,q,sort,is_aktif` | `Paginated<LokasiResp>` |
| `GET` | `/lokasi/{id}` | `lokasi.read` | JWT | — | `LokasiResp` |
| `POST` | `/lokasi` | `lokasi.create` | JWT | `CreateLokasiReq` | `LokasiResp` (201) |
| `PUT` | `/lokasi/{id}` | `lokasi.update` | JWT | `UpdateLokasiReq` | `LokasiResp` |
| `DELETE` | `/lokasi/{id}` | `lokasi.delete` | JWT | — | `null` (soft delete) |

**`CreateLokasiReq`**
```json
{
  "kode": "LAP-BRAWIJAYA",
  "nama": "Lapangan Brawijaya",
  "jenis_lokasi": "lapangan",
  "alamat": "Jl. Brawijaya, Malang",
  "latitude": -7.9666,
  "longitude": 112.6326,
  "radius_meter": 120,
  "timezone": "Asia/Jakarta",
  "foto_file_id": 51,
  "is_aktif": true
}
```
`LokasiResp` = kolom di atas + `foto_uuid` (join `mst_file`) + `jumlah_jadwal` (COUNT jadwal aktif) untuk memberi tahu apakah lokasi boleh dihapus.

---

## 4. Hak akses (rancangan Bab 3.5)

| Modul | Super Admin | Admin | Moderator | User | Guest |
|-------|-------------|-------|-----------|------|-------|
| `lokasi` | CRUD | CRUD | R | R | — |

Permission: `lokasi.create/read/update/delete`. Cakupan **`semua`** (data referensi). Guest tanpa akses.

---

## 5. Kebutuhan BACKEND (Go)

Modul `internal/modules/core/lokasi/`.

### domain
- `Lokasi` map ke `mst_lokasi`; `Latitude/Longitude float64` (decimal), `RadiusMeter int`, `FotoFileID *int64`, embed `Audit`.

### dto
- `CreateLokasiReq`/`UpdateLokasiReq`: `kode required,max=50`; `nama required,max=150`; `latitude required,latitude`; `longitude required,longitude`; `radius_meter required,min=1,max=100000`; `timezone required` (validasi IANA di service); `foto_file_id omitempty,gt=0`; `is_aktif` default true.
- `LokasiResp` (+`foto_uuid`, `jumlah_jadwal`), `ListLokasiQuery` (embed `ListQuery` + `IsAktif *bool`).

### repository
- `Create/Update/SoftDelete/FindByID/List`; semua filter `is_deleted=false`. `List`: `q` ILIKE `kode|nama|alamat`; whitelist sort `kode,nama,created_at`; filter `is_aktif`. `FindByID`/`List` LEFT JOIN `mst_file` (foto_uuid) + subselect `COUNT(jadwal WHERE lokasi_id=… AND is_deleted=false)`. `ExistsKode(kode, exceptID)`. `CountJadwalAktif(lokasiID)`.

### service
- Unik `kode` (409). `time.LoadLocation(timezone)` gagal → `ErrValidation`. Guard hapus: `CountJadwalAktif>0` → `ErrConflict` ("lokasi masih dipakai N jadwal"). Audit `log_aktivitas` create/update/delete. Set `created_by`/`modified_by` dari Claims.

### handler / main / router
- Pola standar; `SetupRoutes` dengan `RequirePermission("lokasi.<aksi>")`. Daftar di router: `lokasi.Initialize(db, jwtMgr, permGuard).SetupRoutes(apiV1)`.

### migrations + seeders
- `mst_lokasi` **persis `.dbml`** termasuk **CHECK** koordinat & radius. Partial-unique `kode`. FK `foto_file_id → mst_file`, audit → `users`. **Sebelum** `jadwal`. Seeder opsional: satu lokasi sekretariat SLAM (idempoten `ON CONFLICT (kode)`).

### edge cases
- Koordinat di luar rentang → 422 (CHECK + binding). `radius_meter<=0` → 422. Timezone tak valid → 422. Hapus lokasi terpakai → 409.

---

## 6. Kebutuhan FRONTEND (Angular)

**WAJIB taste-skill.** Dark, SLAM red, judul UPPERCASE.

### Halaman & komponen (`pages/lokasi/`)
- `lokasi-list.*` — tabel (Kode, Nama, Jenis, Radius, Timezone, Status, Jumlah Jadwal, aksi); filter `q` + `is_aktif`.
- `lokasi-form.*` — form create/edit dengan **map picker**: klik peta untuk set `latitude/longitude`, lingkaran radius yang bisa diseret menampilkan `radius_meter`. Gunakan **Leaflet + tile OpenStreetMap (tanpa API key)** ATAU fallback input numerik lat/long bila peta tidak diizinkan — jangan tambah UI kit berat. `timezone` = `<select>` daftar IANA. Foto via `FileService`.

### Service, form, i18n
- `lokasi.service.ts` tipis di atas `ApiService`. Reactive form cermin DTO (`radius_meter min=1`, koordinat rentang, timezone required). i18n namespace `LOKASI` (`TITLE`, `FORM.KODE/NAMA/JENIS/ALAMAT/RADIUS/TIMEZONE/FOTO`, `STATUS.*`, `DELETE_IN_USE`).

### Guards & gating
- `data:{permission:'lokasi.read'}`; tombol C/U/D `*hasPermission`.

---

## 7. ALUR form → API → DATABASE

| Field form | Payload | Kolom DB |
|------------|---------|----------|
| Peta (klik) | `latitude`,`longitude` | `latitude`,`longitude` |
| Radius (slider) | `radius_meter` | `radius_meter` |
| Timezone | `timezone` | `timezone` |
| Foto (upload→id) | `foto_file_id` | `foto_file_id` |

| Aksi | Endpoint | Tulisan DB |
|------|----------|------------|
| Simpan lokasi | `POST /lokasi` | INSERT `mst_lokasi` + `log_aktivitas` |
| Ubah | `PUT /lokasi/{id}` | UPDATE + `log_aktivitas` |
| Hapus | `DELETE /lokasi/{id}` | Jika dipakai jadwal → 409; else soft delete + log |

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- **Fase 0** — file layer, `mst_pengaturan` (radius bawaan), JWT. **Fase 1** — RBAC `lokasi.*`.
- Termigrasi **sebelum** `jadwal`.

### Acceptance criteria
- [ ] Migrasi cocok `.dbml` termasuk CHECK koordinat & `radius_meter>0`.
- [ ] CRUD merespons envelope; `kode` duplikat → 409.
- [ ] Koordinat/timezone tak valid → 422; hapus lokasi terpakai jadwal → 409.
- [ ] Map picker menyimpan lat/long + radius; timezone IANA; foto via `/files`.
- [ ] C/U/D disembunyikan untuk Mod/User DAN diblok backend 403.
- [ ] `log_aktivitas` tertulis; taste-skill diterapkan.
