# PROMPT — Build APP UI for `prestasi` (Fase 2)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
`workflow/modules/10-prestasi.md`; `_shared/conventions-app.md` (§9 list, §6 form, §8 FileService).

## 2. Scope
CRUD `/prestasi` (`page,per_page,q,sort,anggota_id,tingkat`). Band: SA/Admin/Mod CRUD, User R, Guest R (public via landing).

## 3. Files under `src/app/pages/prestasi/`
- `prestasi.model.ts` (incl. `flyer_uuid`, `foto_uuid`, `anggota_nama`), `prestasi.service.ts`.
- `prestasi-list.ts/.html` — signal list (Judul Kompetisi, Peringkat, Tingkat, Tanggal, actions); filters; **Tambah** `*hasPermission="'prestasi.create'"`.
- `prestasi-form.ts/.html` — typed form: `anggota_id` picker, `judul_kompetisi`, `peringkat`, `tingkat`, `tanggal_kompetisi` native date, `alamat_kompetisi`, `keterangan`, two uploads (`flyer`, `foto_sampul`) via `<app-file-upload>`. Map `422` inline.

## 4. Routing
`{ path:'prestasi', canActivate:[authGuard, permissionGuard], data:{permission:'prestasi.read'}, children:[list,new,:id/edit] }`.

## 5. i18n
Namespace `PRESTASI` (`TITLE`, `FORM.JUDUL/PERINGKAT/TINGKAT/TANGGAL/ALAMAT/FLYER/FOTO/KETERANGAN`) + `COMMON.*`. IND/ENG in sync.

## 6. Verification checklist
- [ ] build compiles; CRUD round-trip; two uploads (blob preview + revoke).
- [ ] User read-only (buttons hidden + backend 403); dark theme + red accent + focus ring; IND/ENG keys match; taste-skill announced.
