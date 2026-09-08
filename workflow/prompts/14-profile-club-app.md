# PROMPT — Build APP UI for `profile_club` (singleton) (Fase 2)

Paste into Claude Code, working directory `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`: dark, SLAM red, UPPERCASE headings, red focus ring), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
`workflow/modules/14-profile-club.md`; `_shared/conventions-app.md` (§6 form, §8 FileService).

## 2. Scope
A **single edit-in-place form** (not a list). Contract: `GET /profile_club` (load the row) + `PUT /profile_club` (upsert). Band: SA/Admin edit, others read.

## 3. Files under `src/app/pages/profile-club/`
- `profile-club.model.ts` (incl. `banner_uuid`, `logo_simple_uuid`, `logo_besar_uuid`), `profile-club.service.ts` (`get()`, `save(form)`).
- `profile-club.ts/.html` — load the current row into a typed form; fields `nama`, `singkatan`, `alamat`, `keterangan` + three uploads (`banner`, `logo_simple`, `logo_besar`) via `<app-file-upload>`. Save → `PUT`; toast; map `422` inline. Read-only rendering for non-editors.

## 4. Routing
`{ path:'profile-club', canActivate:[authGuard, permissionGuard], data:{permission:'profile_club.read'}, loadComponent: ... }` (edit gated by `profile_club.update` via `*hasPermission` on the Save button).

## 5. i18n
Namespace `PROFILE_CLUB` (`TITLE`, `FORM.NAMA/SINGKATAN/ALAMAT/KETERANGAN/BANNER/LOGO_SIMPLE/LOGO_BESAR`) + `COMMON.*`. IND/ENG in sync.

## 6. Verification checklist
- [ ] build compiles; loads the single row; Save upserts (no duplicate); three uploads (blob preview + revoke).
- [ ] Non-editors see read-only (Save hidden) AND backend blocks writes (403); dark theme + red accent + focus ring; IND/ENG keys match; taste-skill announced.
