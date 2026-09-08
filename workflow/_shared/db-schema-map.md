# SLAM Team — Database Schema Map (`slamteam_db`)

Quick-reference map of all **31 tables** in `slamteam_db`, grouped into the 7 TableGroups (kelompok).
Stack: PostgreSQL, accessed by Go + GORM. ERD Rev. 3 (greenfield rebuild — no migration from the old Laravel/MySQL app).

> **Source of truth:** `D:/xampp/htdocs/slam-team-apl/slamteam_db.dbml`. This file is a navigation aid only. If this map and the `.dbml` disagree, **the `.dbml` wins**. Column names / relations / enum labels are verbatim from the DBML.

**Legend:** `M` = master/reference table (`mst_` prefix, plus the RBAC identity/junction tables) · `T` = transactional/operational table. Tables carrying the `mst_` prefix are marked master by convention.

Table count per kelompok: 1·HakAkses (5) · 2·Sistem (4) · 3·Identitas&KTA (5) · 4·File (2) · 5·Jadwal (4) · 6·Absensi (3) · 7·Konten&Master (8) = **31**.

---

## Cross-cutting conventions (Bab 8.2)

These apply to (nearly) every table — assume them unless a table note says otherwise:

- **Timestamps** — all time columns are `timestamptz` stored in **UTC**. There is no `datetime`. Go imports `_ "time/tzdata"`; display in the schedule/location IANA timezone (e.g. `07:00 WIB`), never the browser tz.
- **Soft delete** — `is_deleted boolean` + `deleted_at timestamptz` + `deleted_by bigint`. Rows are hidden, not physically removed. (Reference tables like `mst_modul`, `mst_permission`, `mst_wilayah`, `mst_file_varian`, and the log/session tables have no soft-delete triplet.)
- **Primary key** — `id bigint` identity (`increment`).
- **Public identity** — anything exposed in a URL uses a separate non-sequential column (`uuid` or `token_publik`), never the raw `id`. Present on `notifikasi`, `mst_file`, `absensi`, `kta`.
- **Coordinates** — `decimal(10,7)` (≈ 1 cm precision).
- **Audit stamp** — most tables carry `created_at`/`created_by` and `modified_at`/`modified_by` (bigint → `users.id`).
- **`mst_` prefix** — reserved for master/reference tables only.
- **File references** — every `*_file_id bigint` column is an FK → `mst_file.id`. Private files (identity docs, selfies, izin attachments, KTA PNG) served only through the authenticated Go file handler.

---

## Enum reference (Bab 8.3)

All enums are Postgres `ENUM` types. Values verbatim:

| Enum | Values |
|---|---|
| `aksi_permission` | `create`, `read`, `update`, `delete`, `approve`, `assign`, `export`, `override`, `print`, `cabut`, `batal_sesi` |
| `cakupan_permission` | `semua`, `instansi_sendiri`, `milik_sendiri` |
| `jenis_anggota` | `siswa`, `dewasa`, `siswa_ke_dewasa` (siswa_ke_dewasa → 2 KTA: pelajar arsip + dewasa aktif) |
| `status_anggota` | `aktif`, `non_aktif` (set **manually**; KTA expiry does NOT change it) |
| `jenis_kta` | `pelajar`, `dewasa` |
| `status_kta` | `aktif`, `arsip`, `dicabut`, `kadaluarsa` |
| `mode_absen` | `masuk_saja`, `masuk_pulang` |
| `tipe_absensi` | `masuk`, `pulang` |
| `metode_absensi` | `selfie`, `qr`, `manual` (manual = moderator override for account-holders only) |
| `status_kehadiran` | `hadir`, `terlambat`, `pulang_cepat`, `hadir_luar_radius`, `izin`, `sakit`, `dinas`, `alfa` |
| `jenis_izin` | `izin`, `sakit`, `dinas`, `pulang_cepat` |
| `status_izin` | `menunggu`, `disetujui`, `ditolak` |
| `status_sesi` | `terjadwal`, `berlangsung`, `selesai`, `dibatalkan` |
| `status_tugas` | `ditugaskan`, `diterima`, `ditolak`, `izin` |
| `pola_ulang` | `tidak_berulang`, `harian`, `mingguan`, `bulanan`, `kustom` |
| `varian_file` | `original`, `medium`, `low` |
| `status_proses_file` | `menunggu`, `selesai`, `gagal` |
| `tipe_nilai_pengaturan` | `string`, `integer`, `boolean`, `json`, `date` |
| `prioritas_notifikasi` | `rendah`, `normal`, `tinggi` |

> Note: `jadwal.status` and `kegiatan.status_kegiatan` / `artikel.status_kegiatan` are **not** enums — `jadwal.status` is a `varchar(20)` with convention `draft / terbit / dibatalkan / selesai`; the kegiatan/artikel status columns are `int`/`varchar` free-form.

---

## Kelompok 1 · Hak Akses (5 tables)

Dynamic RBAC. The permission matrix lives entirely in the DB; Go middleware reads `role_permission` and caches per request. `is_super` roles bypass all checks.

### `mst_modul` — M
- **Purpose:** the 19 application modules that drive the sidebar and permission grid.
- **Key columns:** `kode` (unique), `nama`, `grup`, `route`, `icon`, `urutan`, `tampil_di_menu`, `is_aktif`.
- **Constraints/idx:** `kode` unique. 19 seed rows (anggota, admin, moderator, user, kegiatan, medsos, profile_club, inorga, unit, prestasi, instansi, artikel, dokumen, hak_akses, lokasi, jadwal, absensi, izin, kta). No `registrasi` module — no self-registration.
- **Relations:** ← `mst_permission.modul_id`.

### `mst_permission` — M
- **Purpose:** one row per (module × action) — the atomic permissions.
- **Key columns:** `modul_id`, `aksi` (`aksi_permission`), `kode` (unique, format `modul.aksi` e.g. `anggota.create`), `is_berbahaya`.
- **Constraints/idx:** `kode` unique; `(modul_id, aksi)` unique.
- **Relations:** `modul_id` → `mst_modul.id`; ← `role_permission.permission_id`.

### `mst_role` — M
- **Purpose:** roles with a numeric `level` for the anti-escalation hierarchy.
- **Key columns:** `kode` (unique), `nama`, `level` (SA=0, Admin=10, Moderator=20, User=30, Guest=99), `is_sistem`, `is_super` (bypass all checks), `is_aktif`. Soft-delete triplet present.
- **Enforced rules (backend):** super roles can't be deleted/disabled/demoted; last super admin can't be deleted; can't grant a permission you lack; can't edit a role with `level <=` your own; only super admins manage roles.
- **Relations:** ← `role_permission.role_id`, ← `user_role.role_id`, ← `users.role_id`.

### `role_permission` — M (junction, the matrix heart)
- **Purpose:** maps role → permission, with the **`cakupan`** scope dimension distinguishing "see all" from "see own".
- **Key columns:** `role_id`, `permission_id`, `cakupan` (`cakupan_permission`, default `semua`).
- **Constraints/idx:** `(role_id, permission_id)` unique; index on `role_id`.
- **Relations:** `role_id` → `mst_role.id` (**cascade delete**); `permission_id` → `mst_permission.id`.

### `user_role` — M (junction)
- **Purpose:** assigns roles to users. One role per user for now, but code reads through this table so multi-role later needs no migration.
- **Key columns:** `user_id`, `role_id`, `is_utama`, `berlaku_sampai`.
- **Constraints/idx:** `(user_id, role_id)` unique.
- **Relations:** `user_id` → `users.id` (**cascade**); `role_id` → `mst_role.id`.

---

## Kelompok 2 · Sistem (4 tables)

### `mst_pengaturan` — M
- **Purpose:** typed key-value settings that auto-generate the settings page (add a setting = one INSERT). Source of all tunable values (NRA length, KTA dimensions, absensi radius/tolerance, file variant px sizes).
- **Key columns:** `kunci` (unique, format `grup.nama`), `grup` (umum/nra/kta/absensi/file/notifikasi), `label`, `nilai` (text), `tipe_nilai` (`tipe_nilai_pengaturan`), `nilai_bawaan`, `opsi` (jsonb), `satuan`, `is_publik` (readable unauthenticated), `is_terkunci` (super-admin only).
- **Constraints/idx:** `kunci` unique; `(grup, urutan)`. See DBML note for the full seed list.
- **Relations:** `modified_by` → `users.id`.

### `notifikasi` — T
- **Purpose:** in-app notifications, one row per recipient (fan-out). Account-holders only.
- **Key columns:** `uuid` (public), `user_id` (recipient), `tipe`, `judul`, `isi`, `route`, `reff_type`/`reff_id` (polymorphic), `prioritas` (`prioritas_notifikasi`), `is_dibaca`, `dibaca_pada`, `kedaluwarsa_pada`.
- **Constraints/idx:** `uuid` unique; `(user_id, is_dibaca, created_at)` (badge/list query); `(reff_type, reff_id)`; `kedaluwarsa_pada`.
- **Relations:** `user_id` → `users.id` (**cascade**); `created_by` → `users.id`.

### `log_aktivitas` — T
- **Purpose:** generic audit log (replaces the old narrow `log_hak_akses`) — covers RBAC, settings, absensi override, KTA cabut/cetak, login, etc.
- **Key columns:** `aktor_user_id` (null for system/jobs), `modul`, `aksi`, `reff_type`/`reff_id`, `ringkasan`, `nilai_lama`/`nilai_baru` (jsonb), `ip_address` (inet), `user_agent`.
- **Constraints/idx:** `(modul, created_at)`, `(reff_type, reff_id)`, `(aktor_user_id, created_at)`. No soft-delete.
- **Relations:** `aktor_user_id` → `users.id`.

### `sesi_login` — T
- **Purpose:** refresh-token sessions; enables force-logout and active-device listing (PWA keeps sessions long).
- **Key columns:** `user_id`, `refresh_token` (unique), `ip_address`, `user_agent`, `info_perangkat` (jsonb), `berlaku_sampai`, `dicabut_pada`.
- **Constraints/idx:** `refresh_token` unique; index on `user_id`.
- **Relations:** `user_id` → `users.id` (**cascade**).

---

## Kelompok 3 · Identitas & KTA (5 tables)

### `users` — M
- **Purpose:** login accounts. Every user IS an anggota; not every anggota has an account. **No self-registration** — created only via admin panel by holders of `user.create`.
- **Key columns:** `anggota_id`, `username`, `email`, `password` (bcrypt/argon2), `role_id`, `timezone`, `is_aktif`, `login_terakhir`, `gagal_login`, `terkunci_sampai`. Soft-delete triplet.
- **Constraints/idx:** `username` unique (partial `WHERE is_deleted=false`), `email` unique (partial); indexes on `anggota_id`, `role_id`. Login accepts username OR email (single query, both unique). Named `users` (not `user`) because `USER` is reserved in Postgres.
- **Relations:** `anggota_id` → `anggota.id` (**restrict**); `role_id` → `mst_role.id` (**restrict**). Referenced as the actor by nearly every `*_by`/`user_id` column across the schema.

### `mst_wilayah` — M
- **Purpose:** BPS region codes supplying the first 4 NRA digits. Currently fixed at `3573` (Kota Malang); table exists so future regions need no schema change.
- **Key columns:** `kode` (unique, 4-digit BPS), `nama`, `provinsi`, `tingkat`, `is_aktif`.
- **Relations:** ← `anggota.wilayah_id`.

### `anggota` — M
- **Purpose:** the member — core master record. Members created only via admin panel.
- **Key columns:** `instansi_id` (not null), `wilayah_id`, `no_induk` (unique, **nullable** — a copy of the active NRA from `kta`, empty until first card), `nama_lengkap`, `nama_panggilan`, `foto_profil_file_id`, `foto_formal_file_id` (formal red-bg photo for KTA), `jenis_anggota` (`jenis_anggota`), `jenis_kelamin`, `no_identitas` + `file_identitas_file_id` (PRIVATE), `alamat` (printed on KTA), `tempat_lahir` (printed, NOT used for NRA region), `tanggal_lahir` (not null — supplies NRA digits 5-8), `tanggal_bergabung`, `status_anggota` (`status_anggota`, manual). Soft-delete triplet.
- **Constraints/idx:** `no_induk` unique; indexes on `instansi_id`, `jenis_anggota`, `status_anggota`.
- **Key rule:** NRA belongs to the CARD (`kta.no_kta`), not the member. `no_induk` is just a copy of the active NRA. `siswa_ke_dewasa` members have 2 KTA rows and 2 NRAs; `no_induk` points to the dewasa one.
- **Relations:** `instansi_id` → `mst_instansi.id` (**restrict**); `wilayah_id` → `mst_wilayah.id`; three `*_file_id` → `mst_file.id`. Referenced by `users`, `kta`, `jadwal_peserta`, `absensi`, `absensi_izin`, `absensi_rekap`, `unit`, `prestasi`, `mst_medsos`.

### `kta` — T
- **Purpose:** the member card. Owns the NRA. Not deleted — revoked (`cabut`) keeps the row.
- **Key columns:** `anggota_id`, `jenis_kta` (`jenis_kta`), `no_kta` (unique — the NRA printed on this card, globally unique), `no_urut_kartu` (global running number, from a Postgres SEQUENCE, manually editable with setval guard), `token_publik` (unique, QR URL), `qr_url`, `versi_cetak` (bumps on reprint, NRA unchanged), `status` (`status_kta`), `berlaku_dari`/`berlaku_sampai` (auto: pelajar +12mo, dewasa +24mo from `mst_pengaturan`, editable), `file_kta_id` (PNG, PRIVATE), `alasan_cetak`, `dicetak_pada`/`dicetak_oleh`, `batch_cetak_id`, `dicabut_pada`/`dicabut_oleh`/`alasan_dicabut`. Soft-delete triplet.
- **Constraints/idx:** `no_kta` unique, `token_publik` unique, `(anggota_id, jenis_kta, versi_cetak)` unique, `(anggota_id, status)`, `batch_cetak_id`, `berlaku_sampai`.
- **NRA format:** 11 digits = wilayah(4, paten 3573) + birth-month(2) + birth-year(2) + running-number(3, expands to 4 after 999 → 12 digits, hence `varchar(20)`). Global numbering across all cards. `arsip` status applied to pelajar card once dewasa card issues; its QR stops validating and the public page shows a "replaced" notice without leaking the replacement token.
- **Relations:** `anggota_id` → `anggota.id` (**restrict**); `file_kta_id` → `mst_file.id`; `dicetak_oleh`/`dicabut_oleh` → `users.id`.

### `mst_instansi` — M
- **Purpose:** the institution/club a member belongs to (used for RBAC `instansi_sendiri` scope).
- **Key columns:** `kode` (unique), `nama`, `nama_club`, `alamat`, `no_telepon`, `logo_utama_file_id`, `logo_tambahan_file_id`, `tanggal_bergabung`, `status`. Soft-delete triplet.
- **Relations:** ← `anggota.instansi_id`; two `*_file_id` → `mst_file.id`.

---

## Kelompok 4 · File (2 tables)

Built once in Fase 0 before any module touches files. Every image → 3 variants.

### `mst_file` — M
- **Purpose:** logical file record (one per uploaded asset). URL and disk path deliberately separated.
- **Key columns:** `uuid` (public), `nama_asli`, `nama_slug`, `ekstensi`, `mime_type`, `ukuran_byte`, `hash_sha256`, `lebar_px`/`tinggi_px`, `kategori` (foto_profil/foto_formal/absensi/banner/logo/dokumen/flyer/kta), `reff_type`/`reff_id` (polymorphic owner), `storage_driver`, `path_dasar`, `is_publik` (false = must go through auth endpoint), `status_proses` (`status_proses_file`), `metadata_exif` (jsonb). Soft-delete triplet.
- **Constraints/idx:** `uuid` unique; `(reff_type, reff_id)`, `hash_sha256`, `kategori`.
- **Serving:** URL `slamteam.id/{modul}/{reff_id}/{nama_slug}-{varian}.{ext}`; disk `storage/{kategori}/{tahun}/{bulan}/{uuid}/{varian}.{ext}`. Private files served only by the Go handler `GET /api/v1/files/{uuid}/{varian}`, never static nginx.
- **Relations:** ← `mst_file_varian.file_id`; referenced by every `*_file_id` column across the schema.

### `mst_file_varian` — M
- **Purpose:** the 3 rendered variants per image.
- **Key columns:** `file_id`, `varian` (`varian_file`), `path`, `url_publik`, `lebar_px`/`tinggi_px`, `ukuran_byte`, `mime_type`, `kualitas`.
- **Constraints/idx:** `(file_id, varian)` unique. Sizes/quality: original (≤4000px, q90, source fmt), medium (1200px, q80, WebP+JPEG), low (400px, q70, WebP+JPEG) — all px from `mst_pengaturan` grup `file`; images never upscaled.
- **Relations:** `file_id` → `mst_file.id` (**cascade**).

---

## Kelompok 5 · Jadwal (4 tables)

### `mst_lokasi` — M
- **Purpose:** attendance locations with geofence.
- **Key columns:** `kode` (unique), `nama`, `jenis_lokasi`, `alamat`, `latitude`/`longitude` (CHECK ±90 / ±180), `radius_meter` (default 100, CHECK > 0), `timezone` (IANA), `foto_file_id`, `is_aktif`. Soft-delete triplet.
- **Relations:** `foto_file_id` → `mst_file.id`; ← `jadwal.lokasi_id`.

### `jadwal` — T
- **Purpose:** a schedule/event definition with recurrence + attendance rules.
- **Key columns:** `kode` (unique), `nama`, `jenis_jadwal`, `lokasi_id`, `latitude`/`longitude`/`radius_meter` (snapshot/override of location), `timezone`, `tanggal_mulai`/`tanggal_selesai`, `jam_mulai`/`jam_selesai` (CHECK selesai > mulai), `is_berulang`, `pola_ulang` (`pola_ulang`), `hari_ulang` (`int[]` 0-6, 0=Minggu), `interval_ulang`, `tanggal_akhir_ulang`, `mode_absen` (`mode_absen`), `wajib_absen`, `butuh_selfie`, `butuh_lokasi`, `izinkan_luar_radius`, `toleransi_telat_mnt`, `buka_absen_mnt`, `tutup_absen_mnt`, `kuota`, `kegiatan_id`, `inorga_id`, `status` (varchar: draft/terbit/dibatalkan/selesai). Soft-delete triplet.
- **Constraints/idx:** `kode` unique; `(tanggal_mulai, status)`, `lokasi_id`.
- **Relations:** `lokasi_id` → `mst_lokasi.id` (**restrict**); `kegiatan_id` → `kegiatan.id`; `inorga_id` → `mst_inorga.id`. ← `jadwal_sesi`, `jadwal_peserta`, `absensi`.

### `jadwal_sesi` — T
- **Purpose:** concrete session instances generated from recurrence — the anchor point for attendance. Single-date schedules still get one row for a uniform absensi flow.
- **Key columns:** `jadwal_id`, `tanggal_lokal`, `mulai_utc`/`selesai_utc`, `timezone`, `absen_buka_utc`/`absen_tutup_utc`, `status` (`status_sesi`), `alasan_batal`, `jml_ditugaskan`/`jml_hadir` (counters), `catatan`.
- **Constraints/idx:** `(jadwal_id, tanggal_lokal)` unique; `(tanggal_lokal, status)`; `(absen_buka_utc, absen_tutup_utc)`.
- **Relations:** `jadwal_id` → `jadwal.id` (**restrict**); ← `jadwal_peserta.sesi_id`, `absensi.sesi_id`, `absensi_izin.sesi_id`.

### `jadwal_peserta` — T
- **Purpose:** assigns members to a schedule/session. Uses `anggota_id` (not `user_id`) so account-less members can still be assigned.
- **Key columns:** `jadwal_id`, `sesi_id` (null = all sessions), `anggota_id`, `peran_peserta` (peserta/pelatih/panitia/pengawas), `wajib_absen` (auto false if member has no user account — session-closer marks `alfa` only for `true` rows), `status_tugas` (`status_tugas`), `ditugaskan_oleh`, `ditugaskan_pada`, `direspon_pada`. Soft-delete (`is_deleted`/`deleted_at`/`deleted_by`).
- **Constraints/idx:** `(anggota_id, jadwal_id)`, `(jadwal_id, sesi_id)`.
- **Relations:** `jadwal_id` → `jadwal.id` (**restrict**); `sesi_id` → `jadwal_sesi.id`; `anggota_id` → `anggota.id` (**restrict**); `ditugaskan_oleh` → `users.id`.

---

## Kelompok 6 · Absensi (3 tables)

Only account-holding members participate. Server time is authoritative; device clock is evidence only.

### `absensi` — T
- **Purpose:** a check-in/check-out record bound to a session.
- **Key columns:** `uuid` (public), `sesi_id`, `jadwal_id`, `anggota_id` (subject), `user_id` (not null — only account-holders attend), `tipe` (`tipe_absensi`), `waktu_server_utc` (**official reference**), `waktu_perangkat` (evidence), `timezone_jadwal`/`timezone_perangkat`, `selisih_jam_detik`, `latitude`/`longitude`/`akurasi_meter`/`jarak_meter`/`radius_berlaku`/`dalam_radius` (null only for manual override), `konfirmasi_luar_radius`, `foto_file_id` (selfie, PRIVATE, null for manual), `status_kehadiran` (`status_kehadiran`), `menit_telat`, `menit_pulang_cepat`, `izin_id`, `metode` (`metode_absensi`), `info_perangkat` (jsonb), `ip_address`, `flag_mencurigakan` (jsonb: jam_perangkat_beda/akurasi_rendah/exif_tidak_cocok/perangkat_dipakai_banyak_akun), `is_override`/`override_oleh`/`override_pada`/`alasan_override`. Soft-delete triplet.
- **Constraints/idx:** `uuid` unique; `(sesi_id, anggota_id, tipe)` unique; `(anggota_id, waktu_server_utc)`; `(jadwal_id, status_kehadiran)`; `sesi_id`.
- **Two creation paths:** `selfie` (normal — coords/foto/dalam_radius filled) and `manual` (moderator override of account-holder's failed/missed absen — coords/foto empty, `is_override=true`, `alasan_override` required).
- **Relations:** `sesi_id` → `jadwal_sesi.id` (**restrict**); `jadwal_id` → `jadwal.id` (**restrict**); `anggota_id` → `anggota.id` (**restrict**); `user_id` → `users.id` (**restrict**); `foto_file_id` → `mst_file.id`; `override_oleh` → `users.id`; `izin_id` → `absensi_izin.id`.

### `absensi_izin` — T
- **Purpose:** leave/permission requests (izin/sakit/dinas/pulang_cepat) tied to a session.
- **Key columns:** `sesi_id`, `jadwal_id`, `anggota_id`, `jenis` (`jenis_izin`), `alasan` (not null), `waktu_pulang_diminta` (pulang_cepat only), `lampiran_file_id` (e.g. doctor's note, PRIVATE), `status` (`status_izin`), `diproses_oleh`/`diproses_pada`/`catatan_peninjau`. Soft-delete triplet.
- **Constraints/idx:** `(sesi_id, anggota_id)`, `(status, created_at)`.
- **Relations:** `sesi_id` → `jadwal_sesi.id` (**restrict**); `jadwal_id` → `jadwal.id`; `anggota_id` → `anggota.id` (**restrict**); `lampiran_file_id` → `mst_file.id`; `diproses_oleh` → `users.id`; ← `absensi.izin_id`.

### `absensi_rekap` — T (OPTIONAL)
- **Purpose:** per-member per-period attendance rollup. Only built if monthly reports slow down; initially compute directly from `absensi`.
- **Key columns:** `anggota_id`, `periode` (date), `jml_wajib`/`jml_hadir`/`jml_telat`/`jml_pulang_cepat`/`jml_izin`/`jml_sakit`/`jml_alfa`, `total_menit_telat`, `persen_kehadiran` (decimal 5,2), `dihitung_pada`.
- **Constraints/idx:** `(anggota_id, periode)` unique.
- **Relations:** `anggota_id` → `anggota.id` (**restrict**).

---

## Kelompok 7 · Konten & Master (8 tables)

Public-facing content and per-member master data. All have the standard soft-delete triplet + audit stamps and follow standard REST CRUD.

### `kegiatan` — T
- **Purpose:** activities/events for the public landing page.
- **Key columns:** `user_id` (author, not null), `kode`, `judul`, `gambar_highlight_file_id`, `konten`, `tanggal_mulai`/`tanggal_selesai`, `status_kegiatan` (int).
- **Relations:** `user_id` → `users.id`; `gambar_highlight_file_id` → `mst_file.id`; ← `jadwal.kegiatan_id`.

### `artikel` — T
- **Purpose:** blog/news articles.
- **Key columns:** `user_id` (author), `kode`, `judul`, `gambar_highlight_file_id`, `konten`, `status_kegiatan` (varchar here).
- **Relations:** `user_id` → `users.id`; `gambar_highlight_file_id` → `mst_file.id`.

### `unit` — M
- **Purpose:** a member's airsoft gun/unit registry.
- **Key columns:** `anggota_id`, `kode`, `model`, `panjang`/`panjang_inbar`/`lebar`/`berat`/`berat_bb`/`fps` (decimals), `deskripsi_warna`, `foto_sampul_file_id`, `disetujui`/`disetujui_oleh`/`disetujui_pada` (approval workflow).
- **Relations:** `anggota_id` → `anggota.id`; `foto_sampul_file_id` → `mst_file.id`; `disetujui_oleh` → `users.id`.

### `prestasi` — T
- **Purpose:** a member's competition achievements (public-readable incl. Guest).
- **Key columns:** `anggota_id`, `kode`, `peringkat`, `tingkat`, `judul_kompetisi`, `flyer_file_id`, `tanggal_kompetisi`, `alamat_kompetisi`, `foto_sampul_file_id`, `keterangan`.
- **Relations:** `anggota_id` → `anggota.id`; `flyer_file_id`/`foto_sampul_file_id` → `mst_file.id`.

### `profile_club` — M
- **Purpose:** the club's public profile (name, logos, banner, address).
- **Key columns:** `nama`, `singkatan`, `banner_file_id`, `logo_simple_file_id`, `logo_besar_file_id`, `alamat`, `keterangan`.
- **Relations:** three `*_file_id` → `mst_file.id`.

### `mst_inorga` — M
- **Purpose:** internal organization / structure (inorga) periods with SK document.
- **Key columns:** `kode`, `nama`, `logo_file_id`, `banner_file_id`, `tanggal_mulai`/`tanggal_selesai`, `file_sk_file_id`, `konten`.
- **Relations:** three `*_file_id` → `mst_file.id`; ← `jadwal.inorga_id`.

### `mst_medsos` — M
- **Purpose:** a member's social-media links.
- **Key columns:** `anggota_id`, `kode`, `tipe` (int), `icon` (icon-set name like "instagram", NOT an upload), `jenis_medsos`, `konten_medsos`.
- **Relations:** `anggota_id` → `anggota.id`.

### `mst_dokumen` — M
- **Purpose:** polymorphic document attachments for any entity.
- **Key columns:** `kode`, `file_id`, `tipe`, `format`, `reff_id`/`reff_type` (polymorphic owner), `jenis` (int), `keterangan`.
- **Relations:** `file_id` → `mst_file.id`.

---

## Delete-behavior quick list

- **Cascade:** `role_permission.role_id`, `user_role.user_id`, `notifikasi.user_id`, `sesi_login.user_id`, `mst_file_varian.file_id`.
- **Restrict:** `users.role_id`, `users.anggota_id`, `anggota.instansi_id`, `kta.anggota_id`, `jadwal.lokasi_id`, `jadwal_sesi.jadwal_id`, `jadwal_peserta.jadwal_id`, `jadwal_peserta.anggota_id`, all four `absensi.*` FKs, `absensi_izin.sesi_id`, `absensi_izin.anggota_id`, `absensi_rekap.anggota_id`.
- All `*_file_id` FKs → `mst_file.id` use the default (no explicit action).
