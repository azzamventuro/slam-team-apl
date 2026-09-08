# PROMPT — Build APP UI for `inorga` (Fase 2)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
`workflow/modules/11-inorga.md`; `_shared/conventions-app.md` (§9 list, §6 form, §8 FileService).

## 2. Scope
CRUD `/inorga` (`page,per_page,q,sort`). Band: SA/Admin CRUD, Mod/User/Guest R.

## 3. Files under `src/app/pages/inorga/`
- `inorga.model.ts` (incl. `logo_uuid`, `banner_uuid`, `file_sk_uuid`), `inorga.service.ts`.
- `inorga-list.ts/.html` — signal list (Kode, Nama, Periode, actions); **Tambah** `*hasPermission="'inorga.create'"`.
- `inorga-form.ts/.html` — typed form: `kode`, `nama`, `tanggal_mulai/selesai` native dates, `konten` textarea, three uploads (`logo`, `banner`, `file_sk` — SK as a document upload) via `<app-file-upload>`. Map `422` inline.

## 4. Routing
`{ path:'inorga', canActivate:[authGuard, permissionGuard], data:{permission:'inorga.read'}, children:[list,new,:id/edit] }`.

## 5. i18n
Namespace `INORGA` (`TITLE`, `FORM.KODE/NAMA/PERIODE/LOGO/BANNER/SK/KONTEN`) + `COMMON.*`. IND/ENG in sync.

## 6. Verification checklist
- [ ] build compiles; CRUD round-trip; three uploads (logo/banner/SK); previews revoke.
- [ ] Mod/User read-only (buttons hidden + backend 403); dark theme + red accent + focus ring; IND/ENG keys match; taste-skill announced.
