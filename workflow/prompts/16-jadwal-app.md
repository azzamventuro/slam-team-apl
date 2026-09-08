# PROMPT — Build APP UI for `jadwal` (timezone-aware calendar) (Fase 3)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`: dark, SLAM red, UPPERCASE headings, red focus ring, 44px targets), Bootstrap 5. **The schedule calendar is one of the 3 hardest screens — build it first.** *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
`workflow/modules/16-jadwal.md`; `_shared/conventions-app.md` (§9 list, §6 form).

## 2. Scope
| Method | Path |
|--------|------|
| GET/POST/PUT/DELETE | `/jadwal[/:id]` (`jadwal.read/create/update/delete`) |
| POST | `/jadwal/:id/generate-sesi` (`jadwal.update`) |
| GET | `/jadwal/:id/sesi` (`jadwal.read`) |
| PATCH | `/sesi/:id/batalkan` (`jadwal.batal_sesi`) |
Band: SA/Admin/Mod CRUD, User R.

## 3. Files under `src/app/pages/jadwal/`
- `jadwal.model.ts` (Jadwal + Sesi shapes), `jadwal.service.ts` (list/detail/create/update/remove/generateSesi/listSesi/batalkanSesi).
- `jadwal-calendar.ts/.html` — **timezone-aware calendar** (month/week) rendering sessions; **display each session time in the schedule `timezone`** (e.g. 07:00 WIB), NOT the browser tz (convert `mulai_utc` using `jadwal.timezone`). Use a light custom grid or a lightweight calendar consistent with tokens (no heavy UI kit).
- `jadwal-form.ts/.html` — typed form incl. a **recurrence sub-form**: `pola_ulang` select controls conditional controls (`hari_ulang` = Sun–Sat checkboxes when mingguan; `interval_ulang`; `tanggal_akhir_ulang`). Attendance config (mode_absen, toleransi_telat_mnt, buka/tutup menit, flags butuh_selfie/butuh_lokasi/izinkan_luar_radius). `lokasi_id` dropdown (from `/lokasi`). Validate `jam_selesai>jam_mulai`. Map `422` inline.
- `jadwal-sesi-list.ts/.html` — sessions per schedule + **Generate Sesi** action (calls generate-sesi) and **Batalkan Sesi** (`*hasPermission="'jadwal.batal_sesi'"`, requires reason).

## 4. Routing
`{ path:'jadwal', canActivate:[authGuard, permissionGuard], data:{permission:'jadwal.read'}, children:[calendar(list), new, :id/edit, :id/sesi] }`.

## 5. i18n
Namespace `JADWAL` (`TITLE`, `FORM.*`, `RECUR.POLA/HARI/INTERVAL/AKHIR`, `SESI.*`, `GENERATE`, `CANCEL_SESI`, `CANCEL_REASON`) + `COMMON.*`. IND/ENG in sync.

## 6. Verification checklist
- [ ] build compiles; calendar shows sessions with times in the **schedule** timezone (07:00 WIB), not browser tz.
- [ ] Recurrence form: pola_ulang toggles the right conditional controls; generate-sesi materializes sessions; re-run adds none (idempotent).
- [ ] Cancel session requires reason and is gated `jadwal.batal_sesi`; User sees read-only.
- [ ] Dark theme + red accent + focus ring; calendar built first; IND/ENG keys match; taste-skill announced.
