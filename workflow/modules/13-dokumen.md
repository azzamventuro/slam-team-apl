# 13 — Master Data Dokumen (`dokumen`)

> Modul CRUD berulang. Lampiran dokumen generik (polimorfik) yang bisa menempel
> ke entitas mana pun. Fase 2. Grup **Sistem** (bersama `hak_akses`).
> Sumber otoritatif: `slamteam_db.dbml` (tabel `mst_dokumen`) menang atas prosa;
> rute & permission mengikuti endpoint map + rancangan Bab 9 / Bab 3.5.

---

## 1. Ringkasan & tujuan modul

`dokumen` adalah salah satu dari **19 modul** aplikasi SLAM Team dan salah satu
modul **CRUD master-data berpola berulang** (`GET/POST/PUT/DELETE /dokumen`).

- **Nomor modul:** 13 · **kode `mst_modul`:** `dokumen`
- **Fase:** 2 (Anggota + master data) — prasyarat File Layer (Fase 0) & RBAC (Fase 1).
- **Grup sidebar:** `Sistem` (kolom `grup` di `mst_modul`; sekelompok dengan `hak_akses`).
  Meskipun judulnya "Master Data Dokumen", di matriks hak akses & ERD ia hidup di grup **Sistem**.

**Tujuan.** Menyimpan *lampiran dokumen* (SK, sertifikat, surat, berkas internal,
dsb.) yang **tidak dimiliki satu tabel tunggal** melainkan bisa merujuk ke entitas
apa pun lewat pasangan **polimorfik** `reff_type` + `reff_id` (mis. dokumen milik
`inorga`, `instansi`, `anggota`, `unit`, `prestasi`, `kegiatan`, `artikel`). Berkas
fisiknya **tidak** disimpan di baris ini — kolom `file_id` menunjuk ke `mst_file`
(kategori `dokumen`), sehingga seluruh aturan file layer (3 varian, penyajian
ber-auth, soft delete) berlaku otomatis. Baris `mst_dokumen` hanyalah *metadata +
tautan* ke berkas dan ke entitas pemiliknya.

Modul ini **tidak** punya logika bisnis rumit: tidak ada background job, tidak ada
penomoran khusus, tidak ada perhitungan. Ia menumpang File Layer (Fase 0) dan
mengikuti pola CRUD standar. Jangan menambah abstraksi baru — cukup ulang pola
CRUD yang sudah dipakai `instansi`/`inorga`.

---

## 2. Tabel & kolom

### Tabel utama: `mst_dokumen` (DBML baris 1139–1156)

| Kolom | Tipe (DBML) | Aturan / catatan |
|---|---|---|
| `id` | `bigint` [pk, increment] | PK identity. Dipakai di rute admin `/dokumen/{id}`. |
| `kode` | `varchar(50)` | Kode/label internal opsional (mis. `SK-2024-01`). Tidak unik di DBML. |
| `file_id` | `bigint` | **FK → `mst_file.id`** (`Ref: mst_dokumen.file_id > mst_file.id`). Nullable di DBML → `*int64`. Diisi dari hasil upload File Layer (lihat §7). |
| `tipe` | `varchar(50)` | Label tipe berkas bebas (mis. `sk`, `sertifikat`, `surat`, `lampiran`). |
| `format` | `varchar(20)` | Format berkas (mis. `pdf`, `docx`, `jpg`, `png`). Biasanya diturunkan dari ekstensi `mst_file.ekstensi`. |
| `reff_id` | `bigint` | ID entitas pemilik (polimorfik). Bersama `reff_type` menentukan "dokumen ini milik siapa". |
| `reff_type` | `varchar(50)` | Tipe entitas pemilik (polimorfik): `anggota` / `instansi` / `inorga` / `unit` / `prestasi` / `kegiatan` / `artikel` (whitelist di service). |
| `jenis` | `int` | Kode kategori dokumen (integer). Bebas; disarankan konvensi tim (mis. 1=SK, 2=sertifikat, 3=surat, 9=lainnya). Tidak ada tabel lookup di DBML. |
| `keterangan` | `varchar(255)` | Deskripsi singkat / judul dokumen yang tampil di daftar. |
| `is_deleted` | `boolean` [default false] | **Soft delete.** Semua query baca/daftar wajib `WHERE is_deleted = false`. |
| `deleted_at` | `timestamptz` | Diisi saat soft delete. |
| `deleted_by` | `bigint` | User penghapus. |
| `created_at` | `timestamptz` [default `now()`] | UTC. |
| `created_by` | `bigint` | User pembuat (dari klaim JWT). |
| `modified_at` | `timestamptz` | UTC, diisi saat update. |
| `modified_by` | `bigint` | User pengubah. |

**Catatan penting**
- **Tidak ada index/constraint eksplisit** di DBML untuk `mst_dokumen` selain PK.
  Direkomendasikan menambah index komposit `(reff_type, reff_id)` saat migrasi
  (pola sama dengan `mst_file`) karena query paling umum adalah "semua dokumen milik
  satu entitas". (Migrasi boleh menambah index yang tidak melanggar DBML.)
- **Soft delete triplet** (`is_deleted`/`deleted_at`/`deleted_by`) + audit
  (`created_*`/`modified_*`) — pakai embed `Audit` dari konvensi API §6. Bukan
  `gorm.DeletedAt`.
- **Berkas privat.** File dokumen di-upload dengan `mst_file.kategori = 'dokumen'`
  dan umumnya `is_publik = false` → **hanya** bisa dibaca lewat handler ber-auth
  `GET /files/{uuid}/{varian}`. Dokumen non-gambar hanya punya varian `original`.

### Tabel dependensi (baca-saja dari modul ini): `mst_file` (DBML baris 639–671)
Kolom relevan: `id`, `uuid` (identitas publik), `nama_asli`, `nama_slug`, `ekstensi`,
`mime_type`, `ukuran_byte`, `kategori` (`dokumen`), `reff_type`/`reff_id`,
`is_publik`, `status_proses`. Modul dokumen **tidak menulis** ke `mst_file` langsung
— ia memanggil File Service (Fase 0) dan menyimpan `file_id` yang dikembalikan.

---

## 3. Endpoint

Pola CRUD master-data standar (endpoint map / rancangan Bab 9). Semua di bawah
`/api/v1`, di-guard `middleware.JWTAuth` + `RequirePermission("dokumen.<aksi>")`.
Envelope selalu `{ success, message, data?, errors? }`.

| Method | Path | Permission | Auth | Keterangan |
|---|---|---|---|---|
| GET | `/dokumen` | `dokumen.read` | JWT | Daftar + filter + paginate. |
| GET | `/dokumen/{id}` | `dokumen.read` | JWT | Detail satu dokumen. |
| POST | `/dokumen` | `dokumen.create` | JWT | Buat metadata dokumen (menautkan `file_uuid` + `reff_type`/`reff_id`). |
| PUT | `/dokumen/{id}` | `dokumen.update` | JWT | Ubah metadata / ganti file. |
| DELETE | `/dokumen/{id}` | `dokumen.delete` | JWT | **Soft delete** (`is_deleted=true`). |

### 3.1 GET `/dokumen` — query & respons

Query params (mengikuti konvensi list API §5 + filter modul ini):

```
?page=1&per_page=20&q=&sort=-created_at
 &reff_type=inorga&reff_id=5      (filter pemilik — biasanya dipakai dari halaman entitas induk)
 &jenis=1&tipe=sk&format=pdf      (filter opsional)
```

- `q` mencari di `keterangan` dan `kode` (ILIKE).
- `sort` di-whitelist: `created_at`, `modified_at`, `jenis`, `tipe` (prefix `-` = DESC).
- Respons `data` = `Paginated[DokumenResponse]`:

```json
{
  "success": true,
  "message": "OK",
  "data": {
    "items": [
      {
        "id": 12,
        "kode": "SK-2024-01",
        "tipe": "sk",
        "format": "pdf",
        "jenis": 1,
        "keterangan": "SK Kepengurusan 2024",
        "reff_type": "inorga",
        "reff_id": 5,
        "file": {
          "uuid": "b1f2...-uuid",
          "nama_asli": "sk-kepengurusan.pdf",
          "ekstensi": "pdf",
          "mime_type": "application/pdf",
          "ukuran_byte": 348122,
          "is_publik": false
        },
        "created_at": "2026-02-01T03:00:00Z",
        "created_by": 3
      }
    ],
    "page": 1, "per_page": 20, "total": 1, "last_page": 1
  }
}
```

> `file` di-*embed* dari `mst_file` (join by `file_id`) sehingga klien punya `uuid`
> untuk memanggil `GET /files/{uuid}/original` tanpa lookup terpisah. Jangan
> mengekspos `mst_file.id` numerik.

### 3.2 POST `/dokumen` — payload

```json
{
  "reff_type": "inorga",
  "reff_id": 5,
  "file_uuid": "b1f2...-uuid",   // uuid dari hasil POST /files (kategori=dokumen)
  "kode": "SK-2024-01",          // opsional
  "tipe": "sk",                  // opsional
  "format": "pdf",              // opsional; jika kosong diturunkan dari ekstensi file
  "jenis": 1,                    // opsional
  "keterangan": "SK Kepengurusan 2024"
}
```

Respons `201 Created` → `data` = `DokumenResponse` (bentuk sama seperti item di §3.1).

### 3.3 PUT `/dokumen/{id}` — payload
Sama seperti POST tetapi semua field metadata boleh diubah. `file_uuid` **opsional**:
jika diisi → ganti berkas (resolusi ulang `file_id`); jika kosong → pertahankan
`file_id` lama. `reff_type`/`reff_id` boleh dipindah (jarang, tapi diizinkan).

### 3.4 DELETE `/dokumen/{id}`
Soft delete baris `mst_dokumen`. **Tidak** otomatis menghapus `mst_file` (berkas
bisa dipakai/diarsipkan terpisah); penghapusan berkas fisik lewat `DELETE /files/{uuid}`
sesuai permission `file.delete`. Respons `200 OK`.

---

## 4. Hak akses (matriks Bab 3.5)

Baris `dokumen` dari matriks (semua cakupan `semua` — dokumen tidak berdimensi
kepemilikan per-instansi/per-anggota di level modul ini):

| Modul (`kode`) | Grup | Super Admin | Admin | Moderator | User | Guest |
|---|---|---|---|---|---|---|
| `dokumen` | Sistem | CRUD | CRUD | CRUD | **R** | **—** |

Permission `modul.aksi` yang di-seed ke `mst_permission` + `role_permission`:

| Permission | SA | Admin | Mod | User | Guest | cakupan |
|---|:--:|:--:|:--:|:--:|:--:|---|
| `dokumen.read` | ✔ | ✔ | ✔ | ✔ | — | `semua` |
| `dokumen.create` | ✔ | ✔ | ✔ | — | — | `semua` |
| `dokumen.update` | ✔ | ✔ | ✔ | — | — | `semua` |
| `dokumen.delete` | ✔ | ✔ | ✔ | — | — | `semua` |

- Super Admin `is_super` **BYPASS** semua cek (baris SA hanya deskriptif).
- User hanya boleh **membaca** (list/detail + tarik berkas via `/files`).
- Guest **tanpa akses** — tidak ada rute publik untuk dokumen.
- Upload berkasnya sendiri lewat `POST /files` butuh `file.create` (dimiliki semua
  role login). Jadi Moderator (punya `dokumen.create` + `file.create`) bisa
  meng-upload lalu menautkan; User (hanya `dokumen.read`) tidak bisa membuat.

---

## 5. Kebutuhan BACKEND (Go)

Folder: `internal/modules/core/dokumen/{domain,dto,repository,service,handler}` +
`main.dokumen.go`. Ikuti `conventions-api.md`.

### 5.1 domain — `domain/dokumen.go`
```go
type Dokumen struct {
    ID         int64   `gorm:"primaryKey"`
    Kode       *string `gorm:"column:kode"`
    FileID     *int64  `gorm:"column:file_id"`
    Tipe       *string `gorm:"column:tipe"`
    Format     *string `gorm:"column:format"`
    ReffID     int64   `gorm:"column:reff_id"`
    ReffType   string  `gorm:"column:reff_type"`
    Jenis      *int    `gorm:"column:jenis"`
    Keterangan *string `gorm:"column:keterangan"`
    Audit              // embed dari shared (created/modified/deleted triplet)
    File *filedomain.File `gorm:"foreignKey:FileID"` // preload untuk embed uuid
}
func (Dokumen) TableName() string { return "mst_dokumen" }
```
- `reff_type` sebaiknya bertipe Go `string` dengan konstanta whitelist.
- Preload `File` (atau join manual) agar respons memuat `uuid` file. Jangan
  ekspos `mst_file.id`.

### 5.2 dto — `dto/dokumen.go`
- `CreateDokumenReq`: `ReffType (required, oneof=<whitelist>)`, `ReffID (required, gt=0)`,
  `FileUUID (required, uuid4)`, `Kode (max=50)`, `Tipe (max=50)`, `Format (max=20)`,
  `Jenis (omitempty)`, `Keterangan (max=255)`.
- `UpdateDokumenReq`: sama, tapi `FileUUID (omitempty,uuid4)` (opsional untuk ganti berkas).
- `ListQuery`: embed `ListQuery` standar (`page/per_page/q/sort`) + `ReffType`,
  `ReffID`, `Jenis`, `Tipe`, `Format` (semua `form:"...", omitempty`).
- `DokumenResponse` + nested `FileRef` (`uuid`, `nama_asli`, `ekstensi`,
  `mime_type`, `ukuran_byte`, `is_publik`). Mapper `ToResponse(d Dokumen)`.

### 5.3 repository — `repository/dokumen_repo.go`
- `Create`, `Update`, `FindByID` (preload File, `is_deleted=false`),
  `SoftDelete(id, by)`, `List(q) (items, total)`.
- `List`: bangun query dengan filter `reff_type`/`reff_id`/`jenis`/`tipe`/`format`,
  `q ILIKE keterangan/kode`, whitelist sort, `Count` lalu paged `Find` dengan
  `Preload("File")`. Semua query tambahkan scope `is_deleted = false`.
- Terjemahkan `gorm.ErrRecordNotFound` → `ErrNotFound`.

### 5.4 service — `service/dokumen_service.go`
Aturan bisnis (semua di service, bukan binding tag):
1. **Resolve file:** `file_uuid` → cari `mst_file` (via File Service/repo) yang
   `is_deleted=false`. Tidak ketemu → `ErrValidation` ("file tidak ditemukan").
   Ambil `mst_file.id` → simpan ke `Dokumen.FileID`. Idealnya validasi
   `kategori='dokumen'` (tolak jika seseorang menautkan foto profil sebagai dokumen).
2. **Auto-format:** jika `format` kosong, isi dari `mst_file.ekstensi`.
3. **Whitelist `reff_type`:** hanya nilai yang diizinkan (`anggota`, `instansi`,
   `inorga`, `unit`, `prestasi`, `kegiatan`, `artikel`). Nilai lain → `ErrValidation`.
   *(ponytail: verifikasi keberadaan baris entitas induk lintas-modul DILEWATI dulu —
   whitelist + `reff_id>0` cukup; tambah cek eksistensi bila laporan integritas butuh.)*
4. **Audit:** set `created_by`/`modified_by` dari `Claims(ctx)`; `created_at`/
   `modified_at` biar DB/GORM.
5. **Delete:** soft delete saja; jangan sentuh `mst_file`.
6. **log_aktivitas:** tulis baris untuk create/update/delete (`modul="dokumen"`,
   `aksi`, `reff_type/reff_id`, `ringkasan`).

Cakupan: seluruh permission dokumen bercakupan `semua` → tidak perlu penyempitan
query per instansi/pemilik. (Super admin tetap bypass.)

### 5.5 handler — `handler/dokumen_handler.go`
Bind → panggil service → envelope. `List`, `Detail`, `Create`, `Update`, `Delete`.
`ShouldBindJSON` untuk body, `ShouldBindQuery` untuk list. Error → `response.FromError`.

### 5.6 main + router
`main.dokumen.go`: `Initialize(db, jwtMgr, perm, fileSvc)` → wire repo/service/handler;
`SetupRoutes(rg)` mendaftarkan 5 rute dengan guard permission. Daftarkan di
`internal/router/router.go`:
```go
dokumen.Initialize(db, jwtMgr, permGuard, fileSvc).SetupRoutes(apiV1)
```

### 5.7 Migrasi & seeder
- **Migrasi** `00xx_mst_dokumen.up.sql` / `.down.sql` (numbered, sesuai DBML):
  buat tabel `mst_dokumen` persis kolom §2, FK `file_id → mst_file(id)`, index
  `(reff_type, reff_id)`. `down` = `DROP TABLE mst_dokumen`.
- **Seeder RBAC (Fase 1, idempotent):** pastikan modul `dokumen` ada di `mst_modul`
  (grup `Sistem`), 4 permission `dokumen.read/create/update/delete` di
  `mst_permission`, dan baris `role_permission` sesuai §4 (`ON CONFLICT DO NOTHING`).
- Tidak ada seeder data contoh yang wajib.

### 5.8 Edge cases
- `file_uuid` valid tapi berkas sudah soft-deleted → tolak.
- `file_uuid` menunjuk file `kategori != 'dokumen'` → tolak (integritas kategori).
- Update tanpa `file_uuid` → jangan menimpa `file_id` menjadi null.
- `reff_type` di luar whitelist / `reff_id <= 0` → 422.
- Tidak ada background job.

---

## 6. Kebutuhan FRONTEND (Angular)

Folder: `src/app/pages/dokumen/`. Standalone, signals, zoneless, Bootstrap 5,
ngx-translate. Ikuti `conventions-app.md`.

> **Design method (WAJIB):** invoke **design-taste-frontend** (umumkan
> "Using design-taste-frontend"), jalankan pre-flight/audit, map ke Bootstrap 5,
> dan **ENFORCE** token di `workflow/_shared/design-tokens.md` (near-black bg,
> dark-gray surfaces, SLAM red accent, white text, heading UPPERCASE tebal
> berspasi lebar). **Blog-fe restore note:** *jika sumber `blog-fe` dipulihkan,
> tiru layout/menu-nya untuk layar ini; jika tidak, ikuti design tokens.*

### 6.1 Halaman/komponen
- `dokumen-list.ts/.html/.scss` — tabel daftar (kolom: keterangan, tipe, jenis,
  format, pemilik `reff_type`/`reff_id`, ukuran, tanggal, aksi). Filter reaktif
  (`q`, `reff_type`, `reff_id`, `jenis`) + pagination (pola list kanonik §9 app).
  Tombol "Tambah" `*hasPermission="'dokumen.create'"`.
- `dokumen-form.ts/.html` — form buat/ubah (reactive, typed). Termasuk **file
  upload** (lihat §6.3).
- (Opsional) `dokumen-detail` atau modal preview — bisa cukup baris tabel +
  tombol "Unduh/Buka".
- **Reusable:** modul lain (inorga, instansi, dst.) dapat menampilkan sub-daftar
  dokumen dengan memanggil `GET /dokumen?reff_type=inorga&reff_id=5`. Pertimbangkan
  komponen `dokumen-panel` **hanya jika** konsumen kedua benar-benar ada (YAGNI).

### 6.2 Service — `pages/dokumen/dokumen.service.ts`
```ts
list(q: DokumenQuery)  { return this.api.get<Page<Dokumen>>('/dokumen', q); }
detail(id: number)     { return this.api.get<Dokumen>(`/dokumen/${id}`); }
create(b: DokumenForm) { return this.api.post<Dokumen>('/dokumen', b); }
update(id, b)          { return this.api.put<Dokumen>(`/dokumen/${id}`, b); }
remove(id)             { return this.api.delete<void>(`/dokumen/${id}`); }
```
Upload berkas lewat `FileService.upload('file', file, { kategori:'dokumen',
reff_type, reff_id })` (§8 app) → dapat `{ uuid }` → kirim sebagai `file_uuid`.

### 6.3 Form fields (mirror DTO)
| Field | Kontrol | Validator (mirror Go) |
|---|---|---|
| `reff_type` | select (whitelist) | `required` |
| `reff_id` | number (biasanya prefilled dari konteks entitas) | `required`, `min(1)` |
| berkas → `file_uuid` | `<input type="file">` → upload → uuid | `required` saat create |
| `kode` | text | `maxLength(50)` |
| `tipe` | text/select | `maxLength(50)` |
| `format` | text (readonly, auto dari ekstensi) | `maxLength(20)` |
| `jenis` | number/select | opsional |
| `keterangan` | text | `maxLength(255)` |

`FormData` upload **tanpa** set `Content-Type` (browser set boundary). Pada 422,
map `res.errors` ke kontrol. Server tetap otoritatif.

### 6.4 Guard & permission-gating
- Route: `{ path:'dokumen', loadComponent:..., canActivate:[authGuard,permissionGuard],
  data:{ permission:'dokumen.read' } }`.
- Tombol Tambah/Ubah/Hapus dibungkus `*hasPermission="'dokumen.create|update|delete'"`.
  (Ingat: menyembunyikan tombol bukan keamanan — handler Go tetap mengecek.)

### 6.5 File handling (baca berkas)
Berkas privat (`is_publik=false`) → tarik via `FileService.imageUrl(uuid,'original')`
(blob, interceptor menambah Bearer). Untuk PDF/dokumen non-gambar, sama: fetch blob
`GET /files/{uuid}/original` → buat object URL → buka di tab baru / tautan unduh.
**Revoke** object URL saat destroy.

### 6.6 i18n keys (IND + ENG sinkron)
`DOKUMEN.TITLE`, `DOKUMEN.LIST.*`, `DOKUMEN.FORM.REFF_TYPE`, `DOKUMEN.FORM.REFF_ID`,
`DOKUMEN.FORM.FILE`, `DOKUMEN.FORM.KODE`, `DOKUMEN.FORM.TIPE`, `DOKUMEN.FORM.FORMAT`,
`DOKUMEN.FORM.JENIS`, `DOKUMEN.FORM.KETERANGAN`, `DOKUMEN.MSG.CREATED`,
`DOKUMEN.MSG.DELETED`, plus `COMMON.*` / `VALIDATION.*`.

---

## 7. ALUR form → API → DATABASE (kontrak koherensi)

### 7.1 Pemetaan field form → payload → kolom DB
| Field form (Angular) | Key payload (JSON) | Kolom DB (`mst_dokumen`) | Catatan |
|---|---|---|---|
| Pemilik: jenis entitas | `reff_type` | `reff_type` | whitelist string |
| Pemilik: id entitas | `reff_id` | `reff_id` | bigint |
| Unggah berkas → uuid | `file_uuid` | `file_id` | **uuid di-resolve ke `mst_file.id` di service**; DB simpan id numerik |
| Kode | `kode` | `kode` | opsional |
| Tipe | `tipe` | `tipe` | opsional |
| Format | `format` | `format` | kosong → auto dari `mst_file.ekstensi` |
| Jenis | `jenis` | `jenis` | int opsional |
| Keterangan | `keterangan` | `keterangan` | ≤255 |
| (klaim JWT) | — | `created_by`/`modified_by` | dari `Claims(ctx)` |
| (server) | — | `created_at`/`modified_at`/`is_deleted` | default DB / soft delete |

### 7.2 Aksi user → endpoint → tulisan DB
| Aksi user | Endpoint | Tulisan DB |
|---|---|---|
| Pilih berkas & unggah | `POST /files` (kategori=dokumen, reff_type, reff_id) | INSERT `mst_file` (+`mst_file_varian` original) → respons `{uuid}` |
| Simpan dokumen baru | `POST /dokumen` (`file_uuid`, reff, meta) | resolve uuid→`file_id`; INSERT `mst_dokumen` (metadata + `file_id` + audit); INSERT `log_aktivitas` |
| Ubah dokumen | `PUT /dokumen/{id}` | UPDATE `mst_dokumen` (meta; `file_id` diganti hanya bila `file_uuid` dikirim); `modified_*`; `log_aktivitas` |
| Hapus dokumen | `DELETE /dokumen/{id}` | UPDATE `mst_dokumen` set `is_deleted=true, deleted_at, deleted_by`; `log_aktivitas`. `mst_file` **tidak** disentuh |
| Lihat daftar / filter | `GET /dokumen?...` | SELECT (join `mst_file` untuk uuid), `WHERE is_deleted=false` + filter |
| Buka/unduh berkas | `GET /files/{uuid}/original` | SELECT `mst_file` + cek permission/auth; stream dari disk |

Kolom `reff_type`+`reff_id`+`file_uuid`→`file_id` adalah **kontrak inti** yang tidak
boleh menyimpang antara form, payload, dan DB.

---

## 8. Dependencies / prasyarat & Acceptance criteria

### Prasyarat
- **Fase 0 — File Layer** selesai (`POST /files`, `mst_file`/`mst_file_varian`,
  `GET /files/{uuid}/{varian}`, kategori `dokumen`). **Hard dependency.**
- **Fase 1 — RBAC** aktif (`PermGuard.Require`, seed `mst_modul`/`mst_permission`/
  `role_permission`, `perm_version`).
- Envelope `response`, `Audit` embed, `ListQuery`/`Paginated` sudah ada dari modul
  master-data sebelumnya.
- Frontend: `ApiService`, `AuthService`, `PermissionService`, `HasPermissionDirective`,
  `FileService`, interceptors sudah ada.

### Acceptance criteria (checkable)
- [ ] Migrasi membuat `mst_dokumen` persis kolom DBML + FK `file_id→mst_file` + index `(reff_type,reff_id)`; `down` menjatuhkan tabel.
- [ ] Seed RBAC memuat modul `dokumen` (grup Sistem) + 4 permission + baris matriks (SA/Admin/Mod CRUD, User read).
- [ ] `POST /files` (kategori=dokumen) lalu `POST /dokumen` dengan `file_uuid` membuat 1 baris `mst_dokumen` dengan `file_id` terisi benar.
- [ ] `GET /dokumen?reff_type=inorga&reff_id=5` mengembalikan hanya dokumen milik entitas itu, ter-paginate, dengan `file.uuid` ter-embed (tanpa `mst_file.id`).
- [ ] `PUT` tanpa `file_uuid` mempertahankan `file_id`; dengan `file_uuid` menggantinya.
- [ ] `DELETE` menyetel `is_deleted=true` (soft), berkas `mst_file` tetap ada.
- [ ] User (role User) mendapat 403 pada create/update/delete tetapi 200 pada read; Guest 401/403.
- [ ] `reff_type` di luar whitelist / `file_uuid` tidak ada / file bukan kategori `dokumen` → 422 dengan pesan field.
- [ ] Frontend: form submit → baris DB tercipta/terubah; tombol create/update/delete tersembunyi + terblokir tanpa permission; layar responsif; design tokens diterapkan; i18n IND/ENG sinkron.
- [ ] Semua waktu `timestamptz` UTC; tidak ada hard delete; `mst_file.id` numerik tidak pernah bocor ke klien.
