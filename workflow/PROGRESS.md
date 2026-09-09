# PROGRESS — laporan pengerjaan modul

Tracker per-prompt untuk build SLAM Team. Satu baris = satu prompt yang di-paste ke Claude Code (API atau APP). Urutan = build order (`EXECUTION.md` §3). Tandai selesai setiap kali sebuah prompt tuntas + terverifikasi + ter-commit.

**Progress: 4 / 49 selesai** · 🟡 0 proses · ⬜ 45 belum

Status: ⬜ Belum · 🟡 Proses · ✅ Selesai · ⛔ Terblokir

### Cara update (perintah ke AI)
Saat sebuah prompt selesai, minta AI: **"update PROGRESS: `<modul>` `<api|app>` selesai, commit `<hash>`"**. AI akan:
1. Ubah **Status** baris itu ⬜ → ✅ (atau 🟡 saat sedang dikerjakan).
2. Isi **Tanggal** (YYYY-MM-DD), **Commit** (hash pendek), dan **Catatan** bila ada.
3. Perbarui angka **Progress: N / 49** di atas.
4. Tambahkan satu baris ke **Riwayat** di bawah.
5. (Opsional) tandai **V** modul terkait di tabel `00-general-tasks.md`.

Aturan tetap: API dulu, baru APP. Verifikasi = checklist prompt + `make build && go vet ./...` (API) / `npm run build:local` (APP) + satu round-trip **form → API → DB** (row masuk Postgres, envelope benar, role terlarang → 403). Commit per prompt.

---

## Fase 0 — Fondasi

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 1 | 01-foundation | API | Opus · high | ✅ | 2026-09-09 | `46e69f6` | envelope+`apperr`+`FromError`, `model.Audit`, `pagination`, tzdata, CORS config, `slamctl` + migrator |
| 2 | 01-foundation | APP | Opus · high | ✅ | 2026-09-09 | `46e69f6` | `ApiService` + `Page<T>`, interceptor 401/403, design tokens (dark + merah), i18n `COMMON.*`/`VALIDATION.*` |
| 3 | 03-pengaturan-log | API | Sonnet · high | ✅ | 2026-09-09 | `c90d083` | migrasi `0002` + seeder 24 baris idempoten, `shared/audit` (writer lintas modul), gate `RequireAdmin` + `is_terkunci` SA-only di service |
| 4 | 03-pengaturan-log | APP | Sonnet · high | ✅ | 2026-09-09 | `62f1a0d` | halaman self-generating per `tipe_nilai`, `adminGuard` meniru `RequireAdmin` (belum ada `pengaturan.*`), `User.role` ditambah (opsional — auth Fase 0 belum menerbitkan klaim peran) |
| 5 | 02-file-management | API | Opus · high | ⬜ |  |  | butuh varian px (mst_pengaturan) + log_aktivitas |
| 6 | 02-file-management | APP | Opus · high | ⬜ |  |  |  |
| 7 | 04-auth-session | API | Opus · high | ⬜ |  |  |  |
| 8 | 04-auth-session | APP | Opus · high | ⬜ |  |  |  |

## Fase 1 — Hak akses dinamis

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 9 | 05-hak-akses | API | Opus · xhigh ★ | ⬜ |  |  |  |
| 10 | 05-hak-akses | APP | Opus · xhigh ★ | ⬜ |  |  | RBAC grid — layar tersulit |

## Fase 2 — Anggota + master data

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 11 | 08-instansi | API | Sonnet · high | ⬜ |  |  | migrasi sebelum anggota |
| 12 | 08-instansi | APP | Sonnet · high | ⬜ |  |  |  |
| 13 | 06-anggota | API | Sonnet · high | ⬜ |  |  |  |
| 14 | 06-anggota | APP | Sonnet · high | ⬜ |  |  |  |
| 15 | 07-user-management | API | Opus · high | ⬜ |  |  | anti-eskalasi |
| 16 | 07-user-management | APP | Opus · high | ⬜ |  |  |  |
| 17 | 09-unit | API | Sonnet · high | ⬜ |  |  |  |
| 18 | 09-unit | APP | Sonnet · high | ⬜ |  |  |  |
| 19 | 10-prestasi | API | Sonnet · high | ⬜ |  |  |  |
| 20 | 10-prestasi | APP | Sonnet · high | ⬜ |  |  |  |
| 21 | 11-inorga | API | Sonnet · high | ⬜ |  |  |  |
| 22 | 11-inorga | APP | Sonnet · high | ⬜ |  |  |  |
| 23 | 12-medsos | API | Sonnet · high | ⬜ |  |  |  |
| 24 | 12-medsos | APP | Sonnet · high | ⬜ |  |  |  |
| 25 | 13-dokumen | API | Sonnet · high | ⬜ |  |  |  |
| 26 | 13-dokumen | APP | Sonnet · high | ⬜ |  |  |  |
| 27 | 14-profile-club | API | Sonnet · high | ⬜ |  |  | singleton |
| 28 | 14-profile-club | APP | Sonnet · high | ⬜ |  |  |  |

## Fase 3 — Lokasi + jadwal

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 29 | 15-lokasi | API | Sonnet · high | ⬜ |  |  | migrasi sebelum jadwal |
| 30 | 15-lokasi | APP | Sonnet · high | ⬜ |  |  | map picker |
| 31 | 16-jadwal | API | Opus · xhigh ★ | ⬜ |  |  | generate-sesi |
| 32 | 16-jadwal | APP | Opus · xhigh ★ | ⬜ |  |  | kalender sadar timezone |

## Fase 4 — Penugasan + notifikasi

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 33 | 17-penugasan-notifikasi | API | Opus · high | ⬜ |  |  | fan-out + wajib_absen |
| 34 | 17-penugasan-notifikasi | APP | Opus · high | ⬜ |  |  | lonceng notifikasi |

## Fase 5 — Absensi + izin

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 35 | 19-izin | API | Opus · high | ⬜ |  |  | sebelum absensi (izin_id) |
| 36 | 19-izin | APP | Opus · high | ⬜ |  |  |  |
| 37 | 18-absensi | API | Opus · xhigh ★ | ⬜ |  |  | server-time authority |
| 38 | 18-absensi | APP | Opus · xhigh ★ | ⬜ |  |  | kamera + GPS |

## Fase 6 — KTA / NRA / QR

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 39 | 20-kta-nra-qr | API | Opus · xhigh ★ | ⬜ |  |  | NRA sequence + PNG |
| 40 | 20-kta-nra-qr | APP | Opus · xhigh ★ | ⬜ |  |  | halaman QR publik |

## Fase 7 — Laporan / rekap

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 41 | 21-laporan-rekap | API | Opus · high | ⬜ |  |  | export xlsx/pdf |
| 42 | 21-laporan-rekap | APP | Opus · high | ⬜ |  |  |  |

## Fase 8 — Landing + konten publik

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 43 | 22-kegiatan | API | Sonnet · high | ⬜ |  |  |  |
| 44 | 22-kegiatan | APP | Sonnet · high | ⬜ |  |  |  |
| 45 | 23-artikel | API | Sonnet · high | ⬜ |  |  |  |
| 46 | 23-artikel | APP | Sonnet · high | ⬜ |  |  |  |
| 47 | 24-landing-publik | API | Sonnet · high | ⬜ |  |  | endpoint publik |
| 48 | 24-landing-publik | APP | Sonnet · high | ⬜ |  |  | rute di luar authGuard |

## Fase 9 — PWA

| # | Modul | Bagian | Model · Effort | Status | Tanggal | Commit | Catatan |
|---|-------|--------|----------------|--------|---------|--------|---------|
| 49 | 25-pwa | APP | Sonnet · high | ⬜ |  |  | manifest + SW (tanpa API) |

---

## Riwayat

- **2026-09-09** — #1 `01-foundation` API selesai (Opus · high) — commit `46e69f6`.
- **2026-09-09** — #2 `01-foundation` APP selesai (Opus · high) — commit `46e69f6` (satu commit bersama API, karena kerja API belum ter-commit saat APP dikerjakan).
- **2026-09-09** — #3 `03-pengaturan-log` API selesai (Opus · high) — commit `c90d083`. Gate Fase 0 memakai `middleware.RequireAdmin` (peran), bukan `pengaturan.read/update`: tidak ada izin `pengaturan.*` yang di-seed karena pengaturan bukan salah satu dari 19 `mst_modul`. FK `modified_by`/`aktor_user_id` → `users.id` ditunda ke migrasi yang membuat tabel `users`.
- **2026-09-09** — #4 `03-pengaturan-log` APP selesai (Sonnet · high) — commit `62f1a0d`. Mengikuti API: bukan `permissionGuard`, tapi `adminGuard` baru yang meniru `middleware.RequireAdmin` persis (`is_super OR level<=10`). Auth module (scaffold Fase 0) belum menerbitkan `role_id`/`role_level`/`is_super` di `LoginResponse` maupun `GET /auth/me` walau `pkg/jwt.Claims` sudah punya field-nya — jadi `User.role` di frontend opsional dan setiap akun hari ini otomatis ditolak `adminGuard`, sama seperti `RequireAdmin` menolak token tanpa klaim peran. Diverifikasi manual di Chrome dengan sesi & respons API dipalsukan (bukan lewat login sungguhan, karena belum ada user admin yang bisa lolos gate).
