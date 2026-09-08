# SLAM Team — Glossary (domain terms)

Authoritative reference for all agents. If prose here disagrees with `slamteam_db.dbml`, the DBML wins.

## Members & accounts

**anggota** — A club member (person). Row in `anggota`. NOT every anggota has a login: field members can exist purely as records. This is the master identity for a person.

**user** — A login account (row in `users`, linked to an `anggota_id`). Only account-holding members can log in, be assigned to sessions, and do absensi. `anggota` : `user` is 1:0..1 — many anggota have no user.

**NRA (Nomor Register Anggota)** — 11-digit member registration number: `wilayah(4, paten 3573) + birth-month(2) + birth-year(2) + running(3, expands to 4 after 999)`. Global sequence across all cards (Postgres SEQUENCE, setval-guarded). Belongs to the **card** (`kta.no_kta`), not the person; `anggota.no_induk` is just a copy of the active NRA.

## Cards

**KTA (Kartu Tanda Anggota)** — Physical/digital member card (`kta`). ID-1/CR80 sized, PNG output, carries an NRA.

**KTA pelajar / dewasa / arsip** — Card variants. `jenis_anggota = siswa_ke_dewasa` produces TWO kta rows with DIFFERENT NRAs: a **pelajar** card that becomes **arsip** (archived/replaced) and a **dewasa** card that is aktif. Arsip cards are retained but marked replaced.

## Scheduling

**jadwal** — A schedule/event definition (the recurring or one-off plan). Parent record.

**jadwal_sesi** — A concrete session instance generated from a `jadwal` (one dated occurrence). Absensi binds to a `jadwal_sesi`, never directly to `jadwal`.

**jadwal_peserta** — Assignment of an anggota (with an account) to a jadwal/session. Carries `wajib_absen` and accept/reject response.

**sesi window (absen_buka / absen_tutup)** — The open/close time window on a session during which check-in/check-out is allowed. Server time is the authority for whether the window is open.

**wajib_absen** — Flag on `jadwal_peserta`: if true, the session-closer job marks the participant **alfa** when they never check in. If false, absence is not penalized.

## RBAC & permissions

**cakupan** — Scope dimension on `role_permission`: `semua` (all records) / `instansi_sendiri` (own instansi only) / `milik_sendiri` (own records only). Controls data visibility, not just action access.

**is_super** — Flag marking a role as super admin. is_super roles BYPASS all permission checks. The last super admin cannot be deleted.

**perm_version** — Version counter carried in the JWT. Bumped whenever permissions/roles change so cached per-request permissions (`perm:role:{id}`) are invalidated and clients reload effective permissions.

## Absensi

**override** — Moderator manual correction of an attendance record (`metode = manual`). Requires `alasan_override` (reason mandatory). Permission `absensi.override` (SA/Admin/Mod).

**flag_mencurigakan** — JSONB column on an absensi record flagging suspicious signals (e.g. out-of-radius, device clock mismatch, GPS anomalies) as evidence for review.

## Data conventions

**mst_ prefix** — Marks master/reference tables (`mst_pengaturan`, `mst_lokasi`, `mst_wilayah`, `mst_modul`): configuration and lookup data, not transactional records.

**soft delete** — Never hard-delete. Rows carry `is_deleted` (bool) + `deleted_at` + `deleted_by`. Queries filter `is_deleted = false` by default.

**timestamptz + IANA** — All timestamps stored as `timestamptz` in UTC, paired with an IANA timezone column (e.g. `Asia/Jakarta`). Go imports `_ "time/tzdata"`. Display in the schedule's timezone (07:00 WIB), never the browser timezone.

**file varian (original / medium / low)** — Every uploaded image produces 3 sized variants (dimensions from `mst_pengaturan`). Private files are served only via authenticated `GET /api/v1/files/{uuid}/{varian}`.

**token_publik** — A separate uuid/token column used for public-facing URLs (e.g. `GET /public/kta/{token}`). Never expose guessable sequential PK ids; public identity always goes through this token.
