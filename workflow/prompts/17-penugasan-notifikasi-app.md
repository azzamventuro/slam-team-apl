# PROMPT — Build APP participant assignment + notification bell (Fase 4)

> **Repo-accurate (revised 2026-09-13).** App shell, RBAC, topbar, and jadwal/anggota/instansi pages built — mirror them; reuse. Add the bell into the EXISTING topbar.

Working directory: `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce the SLAM tokens in `_shared/design-tokens.md` (dark, SLAM red, UPPERCASE headings, red focus ring, 44px targets), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first / mirror existing code
- `workflow/modules/17-penugasan-notifikasi.md`; `_shared/conventions-app.md` §9 (signal list), §5 (permissions).
- REUSE `api.service.ts`, `permission.service.ts`, `HasPermissionDirective`, `SlamIcon`, toast. Anggota picker ← `GET /anggota`; instansi filter ← `GET /instansi`. **Add the bell into the existing `layouts/topbar/` component** — do not create a new topbar.

## 2. Scope
Assignment (`jadwal.assign`, Admin/Mod) + own notifikasi (bearer). Endpoints from #33:
`POST /jadwal/:id/peserta`, `POST /jadwal/:id/peserta/bulk` → `{ditugaskan,dilewati}`, `GET /jadwal/:id/peserta`, `DELETE /jadwal/:id/peserta/:anggotaId`, `PATCH /peserta/:id/respon`; `GET /notifikasi`, `GET /notifikasi/jumlah-belum-dibaca`, `PATCH /notifikasi/:id/baca`, `PATCH /notifikasi/baca-semua`.

## 3. Files
- `src/app/pages/jadwal/peserta-panel.ts/.html` — mounted on a schedule (`/jadwal/:id/peserta`). List participants (anggota nama, peran_peserta, `wajib_absen` **read-only** + note "otomatis nonaktif — anggota tanpa akun" when applicable, `status_tugas` badge). **Single assign** (anggota picker + peran + wajib_absen) and **bulk assign** (multi-select + a by-instansi filter that expands to that instansi's anggota → `anggota_ids[]`). Assign/remove `*hasPermission="'jadwal.assign'"`; bulk toasts `{ditugaskan, dilewati}`.
- `src/app/pages/jadwal/respon.ts` (or inline) — for the **assignee**: **Terima / Tolak** their own assignment (`PATCH /peserta/:id/respon`). Bearer only.
- `src/app/layouts/topbar/notif-bell.ts/.html` — a bell in the existing topbar: unread **badge** from `/notifikasi/jumlah-belum-dibaca` (refresh on interval + after actions), a dropdown list, **mark-read** + **mark-all-read**, click → `router.navigate(notif.route)`. Signal-based, zoneless-clean.
- Services: `penugasan.service.ts`, `notifikasi.service.ts`.

## 4. i18n
Namespaces `PESERTA` (`ASSIGN`, `BULK`, `PERAN.*`, `WAJIB_ABSEN`, `WAJIB_ABSEN_AUTO_OFF`, `STATUS.*`, `ACCEPT`, `REJECT`) and `NOTIF` (`TITLE`, `MARK_READ`, `MARK_ALL`, `EMPTY`) + `COMMON.*`. IND & ENG in sync.

## 5. Verification checklist
- [ ] `npm run build:local` compiles.
- [ ] Single + bulk assign work; bulk shows `{ditugaskan, dilewati}`; `wajib_absen` read-only for account-less anggota.
- [ ] Assignee can Terima/Tolak; assign/remove gated `jadwal.assign` (hidden for User + backend 403).
- [ ] Notification bell (existing topbar) shows unread badge, lists notifikasi, mark-read/all works, click navigates; count refreshes after actions.
- [ ] Dark theme + red accent + focus ring; IND/ENG keys match; "Using design-taste-frontend" announced.
