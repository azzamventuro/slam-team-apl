# PROMPT — Build APP settings page (auto-generated from typed settings) (Fase 0)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce the SLAM tokens in `_shared/design-tokens.md` (dark, SLAM red, UPPERCASE headings, red focus ring), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
- `D:/xampp/htdocs/slam-team-apl/workflow/modules/03-pengaturan-log.md`.
- `_shared/conventions-app.md` (§3 ApiService, §6 reactive forms).

## 2. Scope
An **auto-generated settings admin page**: each setting renders by its `tipe_nilai`. Contract:

| Method | Path | Permission |
|--------|------|-----------|
| GET | `/pengaturan` | `pengaturan.read` (grouped by `grup`) |
| PUT | `/pengaturan` | `pengaturan.update` (`{items:[{kunci,nilai}]}`) |

Permission band: **SA/Admin only.**

## 3. Files under `src/app/pages/pengaturan/`
- `pengaturan.model.ts` — `Setting { kunci; grup; label; nilai; tipe_nilai: 'string'|'integer'|'boolean'|'json'|'date'; nilai_bawaan; opsi; satuan; is_terkunci; is_publik }`.
- `pengaturan.service.ts` — `list()`, `bulkUpdate(items)`.
- `pengaturan.ts/.html` — group settings by `grup`; for each render the right control by `tipe_nilai`:
  - `string` → text input; `integer` → number; `boolean` → toggle/switch; `date` → native `<input type="date">`; `json` → textarea (validate JSON); `opsi` present → `<select>`.
  - Show `label` + `satuan`; a **"reset to default"** button per row using `nilai_bawaan`.
  - **Disable** `is_terkunci` rows unless the user is super admin (mirror backend; backend still enforces).
  - Collect changed rows → `PUT /pengaturan {items}`; toast + reload; map `422` errors inline.

## 4. Routing
`{ path:'pengaturan', canActivate:[authGuard, permissionGuard], data:{permission:'pengaturan.read'}, loadComponent: ... }`.

## 5. i18n
Namespace `PENGATURAN` (`TITLE`, `RESET_DEFAULT`, `LOCKED`, group labels) + `COMMON.SAVE`. IND & ENG in sync.

## 6. Verification checklist
- [ ] `npm run build:local` compiles.
- [ ] Each setting renders by `tipe_nilai`; `opsi` becomes a select; `boolean` a toggle.
- [ ] Changing values + Save issues one `PUT /pengaturan {items}` and persists (verify via reload/DB).
- [ ] `is_terkunci` rows disabled for non-super; reset-to-default works.
- [ ] SA/Admin only (route guarded); dark theme + red accent + focus ring; IND/ENG keys match; taste-skill announced.
