# PROMPT — Build APP UI for `jadwal` (timezone-aware calendar) (Fase 3) ★

> **Repo-accurate (revised 2026-09-13).** App shell, RBAC, lokasi/anggota/instansi pages built. `kegiatan` picker deferred (Fase 8). Use `Intl.DateTimeFormat` for tz display (no library).

Working directory: `D:/xampp/htdocs/slam-team-apl/slam-team-app`. **The schedule calendar is one of the 3 hardest screens — build it first.**

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce the SLAM tokens in `_shared/design-tokens.md` (dark, SLAM red, UPPERCASE headings, red focus ring, 44px targets), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first / mirror existing code
- `workflow/modules/16-jadwal.md`; `_shared/conventions-app.md` (§9 signal list, §6 reactive form).
- Mirror `pages/lokasi/` and `pages/instansi/`. REUSE `api.service.ts`, `permission.service.ts`, `core/guards/{auth,permission}.guard.ts`, `HasPermissionDirective`, `SlamIcon`, toast. Lokasi dropdown ← `GET /lokasi`; inorga dropdown ← `GET /inorga`.

## 2. Scope (real RBAC)
| Method | Path | Permission |
|--------|------|-----------|
| GET/POST/PUT/DELETE | `/jadwal[/:id]` | `jadwal.read/create/update/delete` |
| POST | `/jadwal/:id/generate-sesi` | `jadwal.update` |
| GET | `/jadwal/:id/sesi` | `jadwal.read` |
| PATCH | `/sesi/:id/batalkan` | `jadwal.batal_sesi` |
Band: SA/Admin/Mod CRUD, User read. `jadwal` is a seeded sidebar module — no hard-coded menu.

## 3. Files under `src/app/pages/jadwal/`
- `jadwal.model.ts` (Jadwal + Sesi; `kegiatan_id` optional/nullable — no kegiatan picker yet), `jadwal.service.ts` (list/detail/create/update/remove/generateSesi/listSesi/batalkanSesi).
- `jadwal-calendar.ts/.html` — **timezone-aware calendar** (month/week). **Display session times in the schedule `timezone`** (07:00 WIB) via `new Intl.DateTimeFormat('id-ID', { timeZone: jadwal.timezone, ... })` — no timezone library. Lightweight custom month grid styled per tokens (no heavy calendar UI kit).
- `jadwal-form.ts/.html` — typed reactive form with a **recurrence sub-form**: `pola_ulang` toggles conditional controls (`hari_ulang` Sun–Sat checkboxes when `mingguan`, `interval_ulang`, `tanggal_akhir_ulang`). Attendance config (`mode_absen`, `toleransi_telat_mnt`, buka/tutup menit, flags, `kuota`). `lokasi_id` dropdown; `inorga_id` optional dropdown; `timezone` select; `status` select. Validate `jam_selesai>jam_mulai`. Map `422` inline. (Omit the kegiatan picker — Fase 8.)
- `jadwal-sesi-list.ts/.html` — sessions per schedule + **Generate Sesi** action + **Batalkan Sesi** (`*hasPermission="'jadwal.batal_sesi'"`, requires reason). Times in schedule tz.

## 4. Routing
`{ path:'jadwal', canActivate:[authGuard, permissionGuard], data:{permission:'jadwal.read'}, children:[ '' → calendar/list, 'new' → form (jadwal.create), ':id/edit' → form (jadwal.update), ':id/sesi' → sesi-list ] }`.

## 5. i18n
Namespace `JADWAL` (`TITLE`, `FORM.*`, `RECUR.POLA/HARI/INTERVAL/AKHIR`, `SESI.*`, `GENERATE`, `CANCEL_SESI`, `CANCEL_REASON`) + `COMMON.*`. IND & ENG in sync.

## 6. Verification checklist
- [ ] `npm run build:local` compiles.
- [ ] Calendar shows session times in the **schedule** timezone, not browser tz (verify by switching the schedule tz).
- [ ] Recurrence form toggles the right controls; Generate Sesi materializes sessions; re-running adds none (idempotent).
- [ ] `jam_selesai>jam_mulai` validated; lokasi/inorga dropdowns load; `422` maps inline.
- [ ] Cancel-session requires a reason and is gated `jadwal.batal_sesi`; User read-only + backend 403.
- [ ] Appears in the dynamic sidebar; dark theme + red accent + focus ring; calendar built first; IND/ENG keys match; taste-skill announced.
