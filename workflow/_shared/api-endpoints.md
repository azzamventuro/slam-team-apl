# SLAM Team — API Endpoint Catalogue (`/api/v1`)

> **Source of truth.** This file is the contract every module MD and every build prompt must match.
> It is derived from rancangan **Bab 9** (endpoint list), **Bab 3.5** (permission matrix), **Bab 6.6** (upload checklist), and `slamteam_db.dbml` (column names). If prose and `.dbml` disagree, **the `.dbml` wins**. Route names, payload shapes and column names follow this document + `.dbml`, **never** the old Laravel app.

## Conventions (apply to every row below)

- **Prefix:** all routes live under `/api/v1`.
- **Auth column:**
  - `public` = no token; served without `JWTAuth`. Only the `/public/*`, `/auth/login`, `/auth/refresh` routes.
  - `bearer` = requires `Authorization: Bearer <access_token>` (JWTAuth middleware). Claims carry `user_id`, `anggota_id`, `role_id`, `perm_version`, `is_super`.
- **Permission column:** the `modul.aksi` string checked by `RequirePermission("modul.aksi")`. `is_super` roles **bypass** all checks. `—` means the route has no permission gate beyond `bearer` (e.g. `/me`, notifikasi, own-scope reads resolved inside the handler). Permission `cakupan` (`semua` / `instansi_sendiri` / `milik_sendiri`) is enforced inside the handler/repository, not by the route string.
- **Standard response envelope** (`internal/shared/response`), every JSON response:
  ```json
  { "success": true, "message": "…", "data": {…} | [...] | null, "errors": null }
  ```
  On error: `success:false`, `data:null`, `errors` carries a field→messages map (validation) or a string.
- **List conventions** for CRUD `GET /<modul>`: query params `?page=1&per_page=20&q=<search>&sort=<col>&order=asc|desc` plus module-specific filters. `data` is the array; pagination in a `meta` sibling: `{ "page":1, "per_page":20, "total":137, "total_pages":7 }`.
- **Soft delete:** `DELETE` = soft delete (`is_deleted=true`, `deleted_at`, `deleted_by`). Rows with `is_deleted=true` are excluded from lists by default.
- **Time:** all `*_utc` timestamps are timestamptz in UTC on the wire (RFC 3339, e.g. `2026-09-08T00:00:00Z`). The client renders in the schedule's IANA timezone, never browser tz.
- **IDs in URLs:** internal `{id}` is the bigint PK (bearer-only surfaces). Public surfaces use the unguessable `{uuid}` / `{token}` column, never the sequential id.

---

## 1. Autentikasi

| Method | Path | Permission | Auth | Purpose |
|---|---|---|---|---|
| POST | `/auth/login` | — | public | Login by **username OR email** + password → access + refresh token. |
| POST | `/auth/refresh` | — | public | Exchange a valid refresh token for a new access token. |
| POST | `/auth/logout` | — | bearer | Revoke the current `sesi_login` row (server-side). |
| GET | `/me` | — | bearer | Current user profile + linked anggota summary. |
| GET | `/me/permissions` | — | bearer | Effective permission list for the current role (drives sidebar + `*hasPermission`). |

### `POST /auth/login`
Request:
```json
{ "identifier": "azzz", "password": "••••••••" }
```
`identifier` accepts either `users.username` or `users.email`. No self-registration exists; accounts are created only via the admin panel / `slamctl`.

Response `data`:
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
The access token (JWT HS256) embeds `user_id`, `anggota_id`, `role_id`, `perm_version`, `is_super`, `exp`. A `sesi_login` row is created (device/ip/agent as evidence). When RBAC changes, `perm_version` is bumped so stale tokens reload permissions.

### `POST /auth/refresh`
Request: `{ "refresh_token": "<opaque>" }` → same `data` shape as login (rotates the refresh token; old one revoked).

---

## 2. Hak Akses (RBAC) — permission `hak_akses.*`, SA only

| Method | Path | Permission | Auth | Purpose |
|---|---|---|---|---|
| GET | `/modul` | `hak_akses.read` | bearer | 19 modules from `mst_modul` (grouped) — also feeds sidebar. |
| GET | `/permissions` | `hak_akses.read` | bearer | All `mst_permission` rows (modul × aksi). |
| GET | `/roles` | `hak_akses.read` | bearer | Roles list with level + is_super. |
| POST | `/roles` | `hak_akses.create` | bearer | Create a role. |
| PUT | `/roles/{id}` | `hak_akses.update` | bearer | Rename/re-level a role. Anti-escalation: cannot edit a role whose `level <= yours`. |
| DELETE | `/roles/{id}` | `hak_akses.delete` | bearer | Delete a role. **The last super-admin role cannot be deleted.** |
| GET | `/roles/{id}/permissions` | `hak_akses.read` | bearer | The role's permission matrix (with `cakupan` per permission). |
| PUT | `/roles/{id}/permissions` | `hak_akses.update` | bearer | Replace the role's matrix. Anti-escalation: **cannot grant a permission you do not hold**; bumps `perm_version`. |

`PUT /roles/{id}/permissions` request:
```json
{ "permissions": [
  { "permission_id": 14, "cakupan": "semua" },
  { "permission_id": 15, "cakupan": "milik_sendiri" }
] }
```

---

## 3. Pengaturan (`mst_pengaturan`)

| Method | Path | Permission | Auth | Purpose |
|---|---|---|---|---|
| GET | `/pengaturan` | `pengaturan.read` | bearer | All settings (grouped by `grup`). |
| GET | `/pengaturan/{grup}` | `pengaturan.read` | bearer | Settings of one group (e.g. `kta`, `absensi`, `file`). |
| PUT | `/pengaturan` | `pengaturan.update` | bearer | Bulk upsert `{ "items": [{ "kunci": "kta.masa_pelajar_bulan", "nilai": "12" }] }`. |
| GET | `/public/pengaturan` | — | public | Whitelisted public settings only (club name, logo, contact) — never thresholds/secrets. |

---

## 4. Notifikasi

| Method | Path | Permission | Auth | Purpose |
|---|---|---|---|---|
| GET | `/notifikasi` | — | bearer | Current user's in-app notifications (paginated). |
| GET | `/notifikasi/jumlah-belum-dibaca` | — | bearer | Unread badge count → `data: { "jumlah": 3 }`. |
| PATCH | `/notifikasi/{id}/baca` | — | bearer | Mark one read. |
| PATCH | `/notifikasi/baca-semua` | — | bearer | Mark all read. |

Notifications only reach account-holding members; anggota without a user account receive none.

---

## 5. Lokasi (`mst_lokasi`)

| Method | Path | Permission | Auth | Purpose |
|---|---|---|---|---|
| GET | `/lokasi` | `lokasi.read` | bearer | List locations (Mod/User = read). |
| POST | `/lokasi` | `lokasi.create` | bearer | Create (map picker → lat/long + radius). SA/Admin. |
| PUT | `/lokasi/{id}` | `lokasi.update` | bearer | Update. SA/Admin. |
| DELETE | `/lokasi/{id}` | `lokasi.delete` | bearer | Soft delete. SA/Admin. |

---

## 6. Jadwal (`jadwal`, `jadwal_sesi`)

| Method | Path | Permission | Auth | Purpose |
|---|---|---|---|---|
| GET | `/jadwal` | `jadwal.read` | bearer | List schedules (User = read). Feeds calendar view. |
| POST | `/jadwal` | `jadwal.create` | bearer | Create a schedule (with recurrence rule). SA/Admin/Mod. |
| PUT | `/jadwal/{id}` | `jadwal.update` | bearer | Update. SA/Admin/Mod. |
| DELETE | `/jadwal/{id}` | `jadwal.delete` | bearer | Soft delete. SA/Admin/Mod. |
| POST | `/jadwal/{id}/generate-sesi` | `jadwal.update` | bearer | Materialise `jadwal_sesi` rows from the recurrence. |
| GET | `/jadwal/{id}/sesi` | `jadwal.read` | bearer | Sessions of one schedule. |
| PATCH | `/sesi/{id}/batalkan` | `jadwal.batal_sesi` | bearer | Cancel a session (`status=dibatalkan`, `alasan_batal` required). SA/Admin/Mod. |

### `POST /jadwal/{id}/generate-sesi`
Turns a schedule's recurrence into concrete `jadwal_sesi` rows. Single-date schedules still produce **one** session row (so the absensi flow is uniform).

Request:
```json
{
  "dari_tanggal": "2026-09-01",
  "sampai_tanggal": "2026-12-31",
  "timezone": "Asia/Jakarta",
  "mulai_lokal": "07:00",
  "selesai_lokal": "09:00",
  "absen_buka_menit_sebelum": 30,
  "absen_tutup_menit_setelah": 15
}
```
Server computes each `tanggal_lokal` from the recurrence, converts local start/end to `mulai_utc` / `selesai_utc` **using the schedule `timezone`** (not browser tz), and derives `absen_buka_utc` / `absen_tutup_utc`. `(jadwal_id, tanggal_lokal)` is unique — regenerating is idempotent (existing rows skipped, not duplicated).

Response `data`: `{ "dibuat": 17, "dilewati": 3, "sesi": [ { "id": 90, "tanggal_lokal": "2026-09-06", "mulai_utc": "2026-09-06T00:00:00Z", "selesai_utc": "2026-09-06T02:00:00Z", "status": "terjadwal" } ] }`

---

## 7. Penugasan (`jadwal_peserta`)

| Method | Path | Permission | Auth | Purpose |
|---|---|---|---|---|
| POST | `/jadwal/{id}/peserta` | `jadwal.assign` | bearer | Assign a single participant (by `anggota_id`). SA/Admin/Mod. |
| POST | `/jadwal/{id}/peserta/bulk` | `jadwal.assign` | bearer | Assign many at once. SA/Admin/Mod. |
| DELETE | `/jadwal/{id}/peserta/{anggotaId}` | `jadwal.assign` | bearer | Remove an assignment (soft delete). SA/Admin/Mod. |
| PATCH | `/peserta/{id}/respon` | — | bearer | Assignee accepts/rejects their own assignment. |

Assignment uses **`anggota_id`, not `user_id`** — anggota without an account can be listed as participants but `wajib_absen` is auto-set `false` for them (they never take attendance and are never marked alfa).

`POST /jadwal/{id}/peserta` request:
```json
{ "anggota_id": 12, "sesi_id": null, "peran_peserta": "peserta", "wajib_absen": true, "keterangan": null }
```
`sesi_id: null` = applies to all sessions of this schedule.

`POST /jadwal/{id}/peserta/bulk` request:
```json
{ "anggota_ids": [12,13,14], "peran_peserta": "peserta", "wajib_absen": true }
```
Response `data`: `{ "ditugaskan": 3, "dilewati": 0 }`.

`PATCH /peserta/{id}/respon` request: `{ "status_tugas": "diterima" | "ditolak", "keterangan": "…" }` → sets `direspon_pada`.

---

## 8. Absensi (`absensi`)

| Method | Path | Permission | Auth | Purpose |
|---|---|---|---|---|
| GET | `/absensi/sesi-aktif` | `absensi.create` | bearer | Sessions currently open for the caller to check in/out. |
| POST | `/absensi/check-in` | `absensi.create` | bearer | Selfie + GPS check-in. **multipart/form-data.** |
| POST | `/absensi/check-out` | `absensi.create` | bearer | Selfie + GPS check-out. **multipart/form-data.** |
| GET | `/absensi` | `absensi.read` | bearer | List attendance. `cakupan`: SA/Admin/Mod = `semua`, User = `milik_sendiri`. |
| GET | `/absensi/rekap` | `absensi.read` | bearer | Recap per anggota/jadwal/periode. |
| PATCH | `/absensi/{id}/override` | `absensi.override` | bearer | Moderator manual correction. SA/Admin/Mod. |
| POST | `/absensi/export` | `absensi.export` | bearer | Export recap (Excel/PDF). SA/Admin/Mod. |
| DELETE | `/absensi/{id}` | `absensi.delete` | bearer | **SA only.** Soft delete an attendance row. |

Only account-holding members take attendance. **Secure context (HTTPS) required** for `getUserMedia` + geolocation.

### `POST /absensi/check-in` (and `/check-out`) — multipart/form-data
Angular sends `FormData` **without** setting `Content-Type` manually (the browser sets the boundary). Go: `ParseMultipartForm` **before** `FormFile`; check `io.Copy` / `os.Create` errors; honour `client_max_body_size` (nginx) and CORS preflight for multipart.

Form fields:
| field | type | note |
|---|---|---|
| `sesi_id` | text (bigint) | target `jadwal_sesi.id` |
| `latitude` | text (decimal) | device GPS |
| `longitude` | text (decimal) | device GPS |
| `akurasi_meter` | text (decimal) | GPS accuracy radius |
| `waktu_perangkat` | text (RFC3339) | **evidence only, never trusted** |
| `timezone_perangkat` | text | IANA tz of device |
| `konfirmasi_luar_radius` | text (bool) | `true` if user pressed OK on the out-of-radius warning |
| `selfie` | **file** | the camera photo → stored PRIVATE, 3 variants |

Server logic: `waktu_server_utc` is the **authority** (device time stored only as evidence + `selisih_jam_detik`). Compute Haversine `jarak_meter` vs `radius_berlaku` (radius snapshot from the session/location at that moment) → set `dalam_radius`. Derive `status_kehadiran` (`hadir` / `terlambat` / `pulang_cepat` / `hadir_luar_radius` / …) and `menit_telat` / `menit_pulang_cepat` from the absen window. Persist `metode=selfie`, `foto_file_id`, `info_perangkat`, `ip_address`, and any `flag_mencurigakan` (jam_perangkat_beda / akurasi_rendah / exif_tidak_cocok / perangkat_dipakai_banyak_akun). `(sesi_id, anggota_id, tipe)` is unique — one check-in + one check-out per session.

Response `data`:
```json
{
  "id": 501, "uuid": "…", "tipe": "masuk",
  "status_kehadiran": "terlambat", "menit_telat": 8,
  "waktu_server_utc": "2026-09-06T00:08:00Z",
  "jarak_meter": 34.2, "dalam_radius": true, "radius_berlaku": 100,
  "foto_url": "/api/v1/files/<uuid>/medium",
  "flag_mencurigakan": null
}
```

### `PATCH /absensi/{id}/override`
Request: `{ "status_kehadiran": "hadir", "alasan_override": "…" }` → sets `metode=manual`, `is_override=true`, `override_oleh`, `override_pada`; `alasan_override` is **required**. No coordinates/photo for manual rows.

`POST /absensi/export` request: `{ "format": "xlsx" | "pdf", "periode": "2026-09", "jadwal_id": null, "anggota_id": null }` → returns the file (binary) or a generated file URL.

---

## 9. Izin (`absensi_izin`)

| Method | Path | Permission | Auth | Purpose |
|---|---|---|---|---|
| POST | `/izin` | `izin.create` | bearer | Submit an izin/sakit/pulang_cepat request. May attach a file (e.g. doctor's note, PRIVATE). |
| GET | `/izin` | `izin.read` | bearer | List izin requests (own vs all by `cakupan`). |
| PATCH | `/izin/{id}/approve` | `izin.approve` | bearer | Approve. SA/Admin/Mod. |
| PATCH | `/izin/{id}/tolak` | `izin.approve` | bearer | Reject (with `catatan_peninjau`). SA/Admin/Mod. |

`POST /izin` request: `{ "sesi_id": 90, "jenis": "sakit", "alasan": "…", "waktu_pulang_diminta": null, "lampiran_file_id": 77 }` (`waktu_pulang_diminta` only for `jenis=pulang_cepat`; `lampiran_file_id` from a prior `POST /files`).

---

## 10. Berkas / File (`mst_file`, `mst_file_varian`)

| Method | Path | Permission | Auth | Purpose |
|---|---|---|---|---|
| POST | `/files` | `file.create` | bearer | Upload a file. **multipart/form-data.** All roles. Every image → 3 variants. |
| GET | `/files/{uuid}/{varian}` | — | bearer | **Authenticated** serve of a private file variant. Permission checked per file/owner inside the handler. |
| DELETE | `/files/{uuid}` | `file.delete` | bearer | Delete. SA/Admin = all; Mod/User = `milik_sendiri`. |

`{varian}` ∈ `original` | `medium` | `low` (dimensions from `mst_pengaturan`). Files are **never** served from a public folder; the guessable path is the reason this route is authenticated (rancangan Bab 6.1). Public KTA foto still routes through the whitelisted public page, not here.

### `POST /files` — multipart/form-data
Follows the Bab 6.6 upload checklist (same rules as absensi check-in: FormData without manual Content-Type; `ParseMultipartForm` before `FormFile`; check `io.Copy`/`os.Create`; nginx body size; writable storage; multipart CORS preflight).

Form fields:
| field | type | note |
|---|---|---|
| `file` | **file** | the upload |
| `modul` | text | owning module code (e.g. `anggota`, `absensi`, `kta`) |
| `reff_id` | text (bigint) | owning row id (optional at upload, can be linked after) |

Response `data`:
```json
{
  "uuid": "…", "nama_asli": "foto.jpg", "ekstensi": "jpg", "mime_type": "image/jpeg",
  "ukuran_byte": 482113,
  "variants": ["original","medium","low"],
  "url": "/api/v1/files/<uuid>/medium"
}
```
Non-image files store only `original`.

---

## 11. KTA / NRA / QR (`kta`)

| Method | Path | Permission | Auth | Purpose |
|---|---|---|---|---|
| POST | `/kta` | `kta.create` | bearer | Issue a card for one anggota (allocates NRA from the global sequence). SA/Admin. |
| POST | `/kta/bulk` | `kta.create` | bearer | Issue + generate many cards in one batch. SA/Admin. |
| GET | `/kta/anggota/{anggotaId}` | `kta.read` | bearer | All cards of one anggota (active + arsip). |
| POST | `/kta/{id}/cetak-ulang` | `kta.print` | bearer | Reprint the **same** card: bumps `versi_cetak`, new `token_publik`, **keeps `no_kta`**. SA/Admin. |
| PATCH | `/kta/{id}/cabut` | `kta.cabut` | bearer | Revoke a card (`status`, `alasan_dicabut`, `dicabut_oleh/pada`). SA/Admin. |
| GET | `/kta/batch/{batchId}` | `kta.read` | bearer | Cards + PNGs of one bulk print batch. |
| GET | `/public/kta/{token}` | — | public | Public membership check page (QR target). See below. |

**NRA** (11 digits, expands to 12 after running-number 999): `wilayah(4, paten 3573)` + `birth-month(2)` + `birth-year(2)` + `running-number(3→4)`. Numbering is **global across all cards** via a Postgres SEQUENCE (manual edit allowed; guard bumps the sequence with `setval` when a manual value overtakes it). NRA belongs to the **card** (`kta.no_kta`); `anggota.no_induk` is a copy of the active NRA. `jenis_anggota=siswa_ke_dewasa` yields **two** cards (pelajar arsip + dewasa aktif) with **different** NRAs.

### `POST /kta` request
```json
{ "anggota_id": 12, "jenis_kta": "dewasa", "alasan_cetak": "baru",
  "berlaku_dari": null, "berlaku_sampai": null }
```
`berlaku_dari` defaults to issue date; `berlaku_sampai` auto = +12 months (pelajar) / +24 months (dewasa) from `mst_pengaturan` (both editable). Server allocates `no_urut_kartu` from the sequence, builds `no_kta` (NRA), `token_publik` (random 32 chars), `qr_url`, generates the CR80 PNG (300 dpi, trim + bleed variants) → `file_kta_id`.

Response `data`:
```json
{
  "id": 88, "anggota_id": 12, "jenis_kta": "dewasa",
  "no_kta": "35731002021", "no_urut_kartu": 21, "versi_cetak": 1, "status": "aktif",
  "berlaku_dari": "2026-09-08", "berlaku_sampai": "2028-09-08",
  "token_publik": "…", "qr_url": "https://slamteam.id/kta/…",
  "file_kta_url": "/api/v1/files/<uuid>/original"
}
```

### `POST /kta/bulk` request
```json
{ "anggota_ids": [12,13,14], "jenis_kta": "dewasa", "alasan_cetak": "baru" }
```
Response `data`: `{ "batch_cetak_id": "BATCH-20260908-01", "diterbitkan": 3, "gagal": 0, "kta": [ … ] }`. Each card gets its own sequential NRA.

### `GET /public/kta/{token}` — PUBLIC, rate-limited
Shows **only**: foto, nama, NRA, membership status, instansi, masa berlaku. **Never** address, id number, full DOB, or contact. Arsip cards return a "card has been replaced" notice **without** leaking the replacement token.
```json
{ "success": true, "data": {
  "foto_url": "https://slamteam.id/api/v1/public/kta/<token>/foto",
  "nama": "Achmad", "nra": "35731002021", "status": "aktif",
  "instansi": "SLAM Malang", "berlaku_sampai": "2028-09-08",
  "digantikan": false
} }
```
Arsip: `{ "digantikan": true, "pesan": "Kartu ini telah digantikan oleh kartu terbaru." }`.

---

## 12. Publik (landing / content)

| Method | Path | Permission | Auth | Purpose |
|---|---|---|---|---|
| GET | `/public/kta/{token}` | — | public | (see §11) |
| GET | `/public/profil` | — | public | Club profile (`profile_club`) + public prestasi. |
| GET | `/public/artikel` | — | public | Published articles (list + detail via `?slug=`). |
| GET | `/public/kegiatan` | — | public | Public activities. |
| GET | `/public/pengaturan` | — | public | (see §3) whitelisted settings. |

---

## 13. Master-data CRUD modules (standard REST)

All follow the same shape. Auth = `bearer`; permission = `<modul>.<aksi>` where `aksi` ∈ `read` (GET), `create` (POST), `update` (PUT), `delete` (DELETE). `cakupan` (`semua`/`instansi_sendiri`/`milik_sendiri`) is resolved inside the handler.

| Method | Path | Permission | Purpose |
|---|---|---|---|
| GET | `/<modul>` | `<modul>.read` | List + filter + paginate. |
| GET | `/<modul>/{id}` | `<modul>.read` | Single record. |
| POST | `/<modul>` | `<modul>.create` | Create. |
| PUT | `/<modul>/{id}` | `<modul>.update` | Update. |
| DELETE | `/<modul>/{id}` | `<modul>.delete` | Soft delete. |

Applies to: **`anggota`, `instansi`, `unit`, `prestasi`, `inorga`, `medsos`, `dokumen`, `profile_club`, `kegiatan`, `artikel`.**

### User-management surfaces (`users` + `user_role`, filtered by role)
`admin`, `moderator`, `user` are the same CRUD shape over `users` filtered by the target role band. Permission matrix (Bab 3.5) differs per surface:

| Modul | Route base | Super Admin | Admin | Moderator | User | Guest |
|---|---|---|---|---|---|---|
| `admin` | `/admin` | CRUD | R | R | R | — |
| `moderator` | `/moderator` | CRUD | CRUD | R | R | — |
| `user` | `/user` | CRUD | CRUD | CRU | R | — |

Per-module permission bands for the CRUD modules above (Bab 3.5), for quick reference:

| Modul | Super Admin | Admin | Moderator | User | Guest |
|---|---|---|---|---|---|
| `anggota` | CRUD | CRUD | CRUD | R | — |
| `instansi` | CRUD | CRUD | R | R | — |
| `unit` | CRUD | CRUD | CRUD | CRUD | — |
| `prestasi` | CRUD | CRUD | CRUD | R | R |
| `inorga` | CRUD | CRUD | R | R | R |
| `kegiatan` | CRUD | CRUD | CRUD | R | R |
| `artikel` | CRUD | CRUD | CRUD | R | R |
| `medsos` | CRUD | CRUD | CRUD | CRUD | R |
| `profile_club` | CRUD | CRUD | R | R | R |
| `dokumen` | CRUD | CRUD | CRUD | R | — |
| `lokasi` | CRUD | CRUD | R | R | — |
| `jadwal` | CRUD | CRUD | CRUD | R | — |

(`—` = no access; `R` = read only; letters = subset of Create/Read/Update/Delete.)

---

## Cross-cutting reminders for module authors

- Every UI permission **must** have a matching Go `RequirePermission` check on the route — hiding a button is not security.
- Permissions are cached per request keyed `perm:role:{id}`; `is_super` bypasses; `perm_version` in the JWT forces reload on RBAC change.
- No self-registration and no super-admin HTTP endpoint — first super admin is created only by `slamctl create-superadmin` (refuses if one already exists).
- Multipart routes (`/files`, `/absensi/check-in`, `/absensi/check-out`): follow the Bab 6.6 checklist verbatim.
- Server time is authoritative for absensi; device time is evidence only.
