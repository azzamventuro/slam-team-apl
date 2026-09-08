# PROMPT — Build APP participant assignment + notification bell (Fase 4)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
`workflow/modules/17-penugasan-notifikasi.md`; `_shared/conventions-app.md` (§9 list, §5 permissions).

## 2. Scope
Assignment (`jadwal.assign`) + own notifikasi (bearer). Endpoints per module doc §3.

## 3. Files
- `src/app/pages/jadwal/peserta-panel.ts/.html` — list a schedule's participants + **single assign** (anggota picker) and **bulk** (multi-select / filter by instansi). Columns: anggota, peran_peserta, `wajib_absen` (read-only + note when anggota has no account), `status_tugas` badge. Assign/remove `*hasPermission="'jadwal.assign'"`.
- `src/app/pages/jadwal/respon.ts/.html` — for an assignee: **Terima / Tolak** their assignment (`PATCH /peserta/:id/respon`).
- `src/app/layouts/topbar/notif-bell.ts/.html` — a bell with an unread **badge** (from `/notifikasi/jumlah-belum-dibaca`), a dropdown list of notifikasi, **mark-read** + **mark-all-read**, and click → `router.navigate(route)`. Poll/refresh the count on an interval or after actions.
- Services: `penugasan.service.ts`, `notifikasi.service.ts`.

## 4. i18n
Namespaces `PESERTA` (`ASSIGN`, `BULK`, `PERAN.*`, `WAJIB_ABSEN`, `STATUS.*`, `ACCEPT/REJECT`) and `NOTIF` (`TITLE`, `MARK_READ`, `MARK_ALL`, `EMPTY`) + `COMMON.*`. IND/ENG in sync.

## 5. Verification checklist
- [ ] build compiles; single + bulk assign work; `wajib_absen` read-only for account-less anggota.
- [ ] Assignee can accept/reject; assignment `jadwal.assign`-gated (buttons hidden + backend 403).
- [ ] Notification bell shows unread badge, lists notifikasi, mark-read/all works, click navigates.
- [ ] Dark theme + red accent + focus ring; IND/ENG keys match; taste-skill announced.
