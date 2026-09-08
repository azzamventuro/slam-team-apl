# PROMPT — Build APP laporan/rekap dashboard (Fase 7)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`). For charts, use the **status colors** from the tokens (hadir=success, terlambat=warning, alfa=danger, izin=info) and follow the dataviz guidance. Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
`workflow/modules/21-laporan-rekap.md`; `_shared/conventions-app.md` (§9 list, §8 FileService for downloads).

## 2. Scope
GET `/absensi/rekap` (`absensi.read`, cakupan), POST `/absensi/export` (`absensi.export`).

## 3. Files under `src/app/pages/laporan/`
- `laporan.model.ts`, `laporan.service.ts` (rekap; export → blob download).
- `rekap.ts/.html` — a rekap table per anggota/periode (count columns + a `persen_kehadiran` bar), summary cards (total hadir/telat/alfa), and simple charts (bar/stacked) using status colors. Filters: periode (month) + jadwal.
- **Export** buttons (Excel / PDF) `*hasPermission="'absensi.export'"` → download the returned file.

## 4. Routing
`{ path:'laporan', canActivate:[authGuard, permissionGuard], data:{permission:'absensi.read'}, loadComponent: ... }`.

## 5. i18n
Namespace `LAPORAN` (`REKAP`, `PERIODE`, `PERSEN`, `EXPORT_XLSX`, `EXPORT_PDF`, status column labels) + `COMMON.*`. IND/ENG in sync.

## 6. Verification checklist
- [ ] build compiles; rekap table + charts render with token status colors; filters work.
- [ ] Export downloads a valid xlsx/pdf; export gated `absensi.export`.
- [ ] User sees own rekap only (cakupan); dark theme + red accent + focus ring; IND/ENG keys match; taste-skill announced.
