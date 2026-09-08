# PROMPT — Build APP UI for `kegiatan` (Fase 8)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
`workflow/modules/22-kegiatan.md`; `_shared/conventions-app.md` (§9 list, §6 form, §8 FileService). (Public rendering lives in module 24 — this is the ADMIN CRUD.)

## 2. Scope
CRUD `/kegiatan` (`page,per_page,q,sort`). Band: SA/Admin/Mod CRUD, User R, Guest R (public via landing).

## 3. Files under `src/app/pages/kegiatan/`
- `kegiatan.model.ts` (incl. `gambar_uuid`), `kegiatan.service.ts`.
- `kegiatan-list.ts/.html` — signal list (Judul, Tanggal, Status badge, actions); **Tambah** `*hasPermission="'kegiatan.create'"`.
- `kegiatan-form.ts/.html` — typed form: `judul` (required), `konten` textarea, `gambar_highlight` upload, `tanggal_mulai/selesai` native dates, `status_kegiatan` select (0 draft / 1 terbit). `user_id` is NOT a form field (set server-side). Map `422` inline.

## 4. Routing
`{ path:'kegiatan', canActivate:[authGuard, permissionGuard], data:{permission:'kegiatan.read'}, children:[list,new,:id/edit] }`.

## 5. i18n
Namespace `KEGIATAN` (`TITLE`, `FORM.JUDUL/KONTEN/GAMBAR/TANGGAL/STATUS`, `STATUS.DRAFT/TERBIT`) + `COMMON.*`. IND/ENG in sync.

## 6. Verification checklist
- [ ] build compiles; CRUD round-trip; highlight upload (blob preview + revoke); status select 0/1.
- [ ] User read-only (buttons hidden + backend 403); dark theme + red accent + focus ring; IND/ENG keys match; taste-skill announced.
