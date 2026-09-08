# PROMPT — Build APP UI for `medsos` (member social links) (Fase 2)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`: dark, SLAM red, UPPERCASE headings, red focus ring), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
- `workflow/modules/12-medsos.md` (module spec) AND `workflow/prompts/12-medsos-api.md` (the API contract already defined — mirror its endpoints/fields EXACTLY).
- `_shared/conventions-app.md` (§9 list, §6 form). Reuse `ApiService/PermissionService/HasPermissionDirective`.

## 2. Scope (envelope `{success,message,data,errors}`)
`mst_medsos` = a member's social-media links. Contract:
| Method | Path | Permission |
|--------|------|-----------|
| GET | `/medsos` (`page,per_page,q,sort,anggota_id,jenis_medsos`) | `medsos.read` |
| GET/POST/PUT/DELETE | `/medsos[/:id]` | `medsos.read/create/update/delete` |
Band: SA/Admin/Mod CRUD (`semua`), **User CRUD (`milik_sendiri`)**, Guest R. **No file upload** — `icon` is an icon-set NAME (e.g. "instagram"), not an upload.

## 3. Files under `src/app/pages/medsos/`
- `medsos.model.ts` — `Medsos { id; anggota_id; kode; tipe; icon; jenis_medsos; konten_medsos }`, `MedsosForm`, `MedsosQuery`.
- `medsos.service.ts` — thin wrapper.
- `medsos-list.ts/.html` — signal list (Jenis, Konten [link], Icon rendered, actions); filter `q` + `jenis_medsos`; for a User-scope view fix `anggota_id` to self.
- `medsos-form.ts/.html` — typed form: `jenis_medsos` (required), `konten_medsos` (required, maxLength 255 — the URL/handle), `icon` (text or select of icon-set names; render the matching icon), `kode`, `tipe`. For User scope, `anggota_id` = self (hidden/fixed); for admins, an anggota picker. Map `422` inline.

## 4. Routing
`{ path:'medsos', canActivate:[authGuard, permissionGuard], data:{permission:'medsos.read'}, children:[list,new,:id/edit] }`.

## 5. i18n
Namespace `MEDSOS` (`TITLE`, `FORM.JENIS/KONTEN/ICON`) + `COMMON.*`. IND/ENG in sync.

## 6. Verification checklist
- [ ] `npm run build:local` compiles.
- [ ] CRUD round-trip; `konten_medsos` maxLength 255; icon renders from the icon-set name (no upload).
- [ ] User (`milik_sendiri`) manages only their own links (anggota_id fixed to self); admins can filter by anggota.
- [ ] Buttons hidden without permission AND backend 403; dark theme + red accent + focus ring; IND/ENG keys match; taste-skill announced.
