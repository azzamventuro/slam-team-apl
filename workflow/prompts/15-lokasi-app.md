# PROMPT — Build APP UI for `lokasi` (map picker) (Fase 3)

> **Repo-accurate (revised 2026-09-13).** App shell, RBAC, shared file components, and many CRUD pages are already built — mirror them; reuse. `lokasi` is a seeded sidebar module.

Working directory: `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce the SLAM tokens in `_shared/design-tokens.md` (dark, SLAM red, UPPERCASE headings, red focus ring, 44px targets), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first / mirror existing code
- `workflow/modules/15-lokasi.md`; `_shared/conventions-app.md` (§9 signal list, §6 reactive form, §8 FileService).
- Mirror `src/app/pages/instansi/` or `pages/unit/`. REUSE `core/services/api.service.ts`, `shared/components/file-upload`, `HasPermissionDirective`, `core/services/permission.service.ts`, `core/guards/{auth,permission}.guard.ts`, `SlamIcon`, toast. Match the file-reference pattern those pages use.

## 2. Scope
CRUD `/lokasi` (`page,per_page,q,sort,is_aktif`). Band: SA/Admin CRUD, Mod/User R. `lokasi` is a seeded sidebar module — do NOT hard-code a menu entry.
Contract (from #29): `GET/POST/PUT/DELETE /lokasi` — `kode, nama, jenis_lokasi, alamat, latitude, longitude, radius_meter, timezone, foto (per convention), is_aktif`.

## 3. Files under `src/app/pages/lokasi/`
- `lokasi.model.ts` — mirror the API response (incl. resolved foto uuid; `jumlah_jadwal` may be 0 for now).
- `lokasi.service.ts` — thin wrapper over `ApiService`.
- `lokasi-list.ts/.html` — signal-based list (like instansi/unit): columns Kode, Nama, Jenis, Radius, Timezone, Status, actions; filters `q` + `is_aktif`; **Tambah** `*hasPermission="'lokasi.create'"`; Edit/Hapus gated `lokasi.update`/`lokasi.delete`. On 409 delete → `LOKASI.DELETE_IN_USE` toast (defensive — guard lands with `16-jadwal`).
- `lokasi-form.ts/.html` — typed reactive form (create/edit by `:id`) with a **MAP PICKER**:
  - **Leaflet + OpenStreetMap tiles** (no API key; `npm i leaflet @types/leaflet --legacy-peer-deps`, import Leaflet CSS, keep OSM attribution). Click to set `latitude`/`longitude`; a **draggable circle** whose radius reflects `radius_meter` (two-way). Keep numeric lat/long inputs in sync as fallback.
  - `timezone` = `<select>` IANA (default Asia/Jakarta); `jenis_lokasi`, `alamat`, `is_aktif`; foto via `<app-file-upload>`.
  - Validate `radius_meter >= 1`, coord ranges; map `422` inline. Dispose the Leaflet map + revoke object URLs on destroy.

## 4. Routing
`{ path:'lokasi', canActivate:[authGuard, permissionGuard], data:{permission:'lokasi.read'}, children:[ '' → list, 'new' → form (lokasi.create), ':id/edit' → form (lokasi.update) ] }`.

## 5. i18n
Namespace `LOKASI` (`TITLE`, `FORM.KODE/NAMA/JENIS/ALAMAT/RADIUS/TIMEZONE/FOTO`, `STATUS.AKTIF/NONAKTIF`, `DELETE_IN_USE`) + `COMMON.*`/`VALIDATION.*`. IND & ENG in sync.

## 6. Verification checklist
- [ ] `npm run build:local` compiles (Leaflet imported + CSS loaded).
- [ ] Create/edit/soft-delete round-trip; map click sets lat/long; radius circle ↔ `radius_meter` in sync; timezone select works.
- [ ] Coord/radius client validation mirrors the DTO; `422` maps inline.
- [ ] Mod/User: Tambah/Edit/Hapus hidden AND backend 403; list still loads.
- [ ] Leaflet map disposed + object URLs revoked on destroy.
- [ ] Appears in the dynamic sidebar under Operasional; dark theme + red accent + focus ring; IND/ENG keys match; "Using design-taste-frontend" announced.
