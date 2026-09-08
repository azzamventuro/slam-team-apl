# PROMPT — Build APP UI for `anggota` (Master Data Anggota) (Fase 2)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce the SLAM tokens in `_shared/design-tokens.md` (dark, SLAM red, UPPERCASE headings, red focus ring, 44px targets), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
- `D:/xampp/htdocs/slam-team-apl/workflow/modules/06-anggota.md`.
- `_shared/conventions-app.md` (§9 signal list, §6 reactive forms, §8 FileService). Reuse `ApiService/PermissionService/HasPermissionDirective/FileService`.

## 2. Scope (envelope `{success,message,data,errors}`)
| Method | Path | Permission |
|--------|------|-----------|
| GET | `/anggota` (`page,per_page,q,sort,instansi_id,jenis_anggota,status_anggota`) | `anggota.read` |
| GET/POST/PUT/DELETE | `/anggota[/:id]` | `anggota.read/create/update/delete` |

Band: **SA/Admin/Mod CRUD, User read-only, Guest none.**

## 3. Files under `src/app/pages/anggota/`
- `anggota.model.ts` — `Anggota` (mirror `AnggotaResp`: `no_induk`, `nama_lengkap`, `jenis_anggota`, `status_anggota`, `instansi_id/instansi_nama`, `tanggal_lahir`, three `*_file_id`+`*_uuid`), `AnggotaForm`, `AnggotaQuery`.
- `anggota.service.ts` — thin wrapper over `ApiService`.
- `anggota-list.ts/.html/.scss` — signal-based list (conventions-app §9): columns NRA (`no_induk`), Nama, Instansi, Jenis, Status badge, actions; filters `q` + instansi + jenis + status selects; empty/loading states; **Tambah** `*hasPermission="'anggota.create'"`.
- `anggota-form.ts/.html` — one component for create/edit (mode from `:id`). Typed `FormBuilder.nonNullable` mirroring the DTO:
  - `nama_lengkap` required maxLength 150; `instansi_id` required (dropdown from `/instansi`); `jenis_anggota` select (siswa/dewasa/siswa_ke_dewasa); `status_anggota` select (aktif/non_aktif); `tanggal_lahir` **required** native date; plus `nama_panggilan`, `jenis_kelamin`, identitas, alamat, kode_pos, tempat_lahir, pekerjaan.
  - **Three uploads** via `<app-file-upload>`: `foto_profil` (kategori foto_profil), `foto_formal` (foto_formal — note "red-bg photo for KTA"), `file_identitas` (kategori dokumen, PRIVATE). Store returned ids into form controls; previews via FileService (revoke on destroy).
  - Submit: invalid → markAllAsTouched; create/update; map `422` `res.errors` inline; toast + navigate on success.

## 4. Routing
`{ path:'anggota', canActivate:[authGuard, permissionGuard], data:{permission:'anggota.read'}, children:[list, new(create), :id/edit(update)] }`. Dynamic sidebar surfaces it automatically.

## 5. i18n
Namespace `ANGGOTA` (`TITLE`, `FORM.*` for each field, `JENIS.*`, `STATUS.AKTIF/NON_AKTIF`) + `COMMON.*`/`VALIDATION.*`. IND & ENG in sync.

## 6. Verification checklist
- [ ] `npm run build:local` compiles.
- [ ] Create/edit/soft-delete round-trip; `no_induk` shown read-only (filled later by KTA).
- [ ] `jenis_anggota`/`status_anggota` selects match enum; `tanggal_lahir` required; instansi picker required.
- [ ] Three uploads post FormData w/o manual Content-Type; `file_identitas` treated as private (blob fetch); previews revoke.
- [ ] User sees list but no C/U/D buttons AND backend blocks writes (403).
- [ ] Responsive; dark theme + red accent + UPPERCASE + focus ring; IND/ENG keys match; taste-skill announced.
