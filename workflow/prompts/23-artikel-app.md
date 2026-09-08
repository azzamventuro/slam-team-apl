# PROMPT — Build APP UI for `artikel` (Fase 8)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
`workflow/modules/23-artikel.md`; `_shared/conventions-app.md` (§9 list, §6 form, §8 FileService). (Public blog rendering lives in module 24 — this is the ADMIN CRUD.)

## 2. Scope
CRUD `/artikel` (`page,per_page,q,sort`). Band: SA/Admin/Mod CRUD, User R, Guest R. NOTE the status field is `status_kegiatan` (values `draft`/`terbit`).

## 3. Files under `src/app/pages/artikel/`
- `artikel.model.ts` (incl. `gambar_uuid`; status field name `status_kegiatan`), `artikel.service.ts`.
- `artikel-list.ts/.html` — signal list (Judul, Status, actions); **Tambah** `*hasPermission="'artikel.create'"`.
- `artikel-form.ts/.html` — typed form: `judul` (required), `konten` **content editor** (textarea or a lightweight rich-text — **no heavy new dependency**), `gambar_highlight` upload, `status_kegiatan` select (draft/terbit). `user_id` is not a form field. Map `422` inline.

## 4. Routing
`{ path:'artikel', canActivate:[authGuard, permissionGuard], data:{permission:'artikel.read'}, children:[list,new,:id/edit] }`.

## 5. i18n
Namespace `ARTIKEL` (`TITLE`, `FORM.JUDUL/KONTEN/GAMBAR/STATUS`, `STATUS.DRAFT/TERBIT`) + `COMMON.*`. IND/ENG in sync.

## 6. Verification checklist
- [ ] build compiles; CRUD round-trip; content editor works; highlight upload (blob preview + revoke).
- [ ] Status select uses `draft`/`terbit`; User read-only (buttons hidden + backend 403); dark theme + red accent + focus ring; IND/ENG keys match; taste-skill announced.
