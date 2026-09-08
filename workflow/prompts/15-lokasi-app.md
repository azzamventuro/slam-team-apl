# PROMPT — Build APP UI for `lokasi` (map picker) (Fase 3)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
`workflow/modules/15-lokasi.md`; `_shared/conventions-app.md` (§9 list, §6 form, §8 FileService).

## 2. Scope
CRUD `/lokasi` (`page,per_page,q,sort,is_aktif`). Band: SA/Admin CRUD, Mod/User R, Guest none.

## 3. Files under `src/app/pages/lokasi/`
- `lokasi.model.ts` (incl. `foto_uuid`, `jumlah_jadwal`), `lokasi.service.ts`.
- `lokasi-list.ts/.html` — signal list (Kode, Nama, Jenis, Radius, Timezone, Status, Jumlah Jadwal, actions); filters `q`+`is_aktif`; on 409 delete → `LOKASI.DELETE_IN_USE` toast.
- `lokasi-form.ts/.html` — typed form with a **MAP PICKER**: click map to set `latitude/longitude`; a draggable radius circle reflects `radius_meter`. Use **Leaflet + OpenStreetMap tiles (no API key)**; if a map lib is not desired, fall back to numeric lat/long inputs — do NOT add a heavy UI kit. `timezone` = `<select>` of IANA zones (default Asia/Jakarta). `jenis_lokasi`, `alamat`, foto upload. Validate `radius_meter>=1`, coord ranges. Map `422` inline.

## 4. Routing
`{ path:'lokasi', canActivate:[authGuard, permissionGuard], data:{permission:'lokasi.read'}, children:[list,new,:id/edit] }`.

## 5. i18n
Namespace `LOKASI` (`TITLE`, `FORM.KODE/NAMA/JENIS/ALAMAT/RADIUS/TIMEZONE/FOTO`, `STATUS.*`, `DELETE_IN_USE`) + `COMMON.*`. IND/ENG in sync.

## 6. Verification checklist
- [ ] build compiles; CRUD round-trip; map click sets lat/long; radius circle reflects value.
- [ ] Timezone select (IANA); coord/radius validation mirrors DTO; delete-in-use surfaces 409 toast.
- [ ] Mod/User read-only (buttons hidden + backend 403); dark theme + red accent + focus ring; IND/ENG keys match; taste-skill announced.
