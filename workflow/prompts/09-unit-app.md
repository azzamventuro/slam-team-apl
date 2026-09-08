# PROMPT — Build APP UI for `unit` (Fase 2)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`: dark, SLAM red, UPPERCASE headings, red focus ring), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
`workflow/modules/09-unit.md`; `_shared/conventions-app.md` (§9 list, §6 form, §8 FileService). Reuse `ApiService/PermissionService/HasPermissionDirective/FileService`.

## 2. Scope
CRUD `/unit` (`page,per_page,q,sort,anggota_id,disetujui`). Band: **all logged-in roles CRUD** (User `milik_sendiri`). Guest none.

## 3. Files under `src/app/pages/unit/`
- `unit.model.ts` (mirror response incl. `foto_uuid`, `anggota_nama`, `disetujui`), `unit.service.ts`.
- `unit-list.ts/.html` — signal list (Kode, Model, FPS, Berat, Disetujui badge, actions); filters; **Tambah** `*hasPermission="'unit.create'"`.
- `unit-form.ts/.html` — typed form: `anggota_id` (self for User; picker for admins), `model`, numeric spec fields (panjang/lebar/berat/berat_bb/fps — number inputs), `deskripsi_warna`, `foto_sampul` upload (`<app-file-upload>`). Approve control (`disetujui`) shown only to the approver permission. Map `422` inline.

## 4. Routing
`{ path:'unit', canActivate:[authGuard, permissionGuard], data:{permission:'unit.read'}, children:[list,new,:id/edit] }`.

## 5. i18n
Namespace `UNIT` (`TITLE`, `FORM.MODEL/PANJANG/LEBAR/BERAT/BERAT_BB/FPS/WARNA/FOTO`, `DISETUJUI`) + `COMMON.*`. IND/ENG in sync.

## 6. Verification checklist
- [ ] build compiles; CRUD round-trip; decimals validated; foto preview via blob (revoke).
- [ ] User manages own units; approve control gated; buttons hidden without permission AND backend 403.
- [ ] Dark theme + red accent + focus ring; IND/ENG keys match; taste-skill announced.
