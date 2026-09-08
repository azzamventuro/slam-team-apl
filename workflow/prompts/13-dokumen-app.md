# PROMPT — Build APP UI for `dokumen` (Fase 2)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
`workflow/modules/13-dokumen.md`; `_shared/conventions-app.md` (§9 list, §6 form, §8 FileService).

## 2. Scope
CRUD `/dokumen` (`page,per_page,q,sort,reff_type,reff_id,tipe,jenis`). Band: SA/Admin/Mod CRUD, User R, Guest none. Documents are private (any file type).

## 3. Files under `src/app/pages/dokumen/`
- `dokumen.model.ts` (incl. `file_uuid`, `nama_asli`), `dokumen.service.ts`.
- `dokumen-list.ts/.html` — signal list (Kode, Tipe, Reff, Keterangan, download action); filter by `reff_type`/`tipe`; **Tambah** `*hasPermission="'dokumen.create'"`.
- `dokumen-form.ts/.html` — typed form: **document upload** (any file type, not just images — `<app-file-upload>` with `kategori:'dokumen'`), `tipe`, `format`, `reff_type`/`reff_id` (polymorphic link), `jenis`, `keterangan`. Download via the authenticated file endpoint (blob). Map `422` inline.

## 4. Routing
`{ path:'dokumen', canActivate:[authGuard, permissionGuard], data:{permission:'dokumen.read'}, children:[list,new,:id/edit] }`.

## 5. i18n
Namespace `DOKUMEN` (`TITLE`, `FORM.TIPE/FORMAT/REFF/JENIS/KETERANGAN/FILE`) + `COMMON.*`. IND/ENG in sync.

## 6. Verification checklist
- [ ] build compiles; CRUD round-trip; upload accepts non-image files; download via authenticated blob.
- [ ] Polymorphic reff link works; User read-only (buttons hidden + backend 403); dark theme + red accent + focus ring; IND/ENG keys match; taste-skill announced.
