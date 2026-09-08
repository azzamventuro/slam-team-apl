# PROMPT — Build APP RBAC grid + guard/directive/sidebar (Fase 1)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce the SLAM tokens in `_shared/design-tokens.md` (dark, SLAM red, UPPERCASE headings, red focus ring, 44px targets), Bootstrap 5. The **role × modul grid is one of the 3 hardest screens — build it first**. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
- `D:/xampp/htdocs/slam-team-apl/workflow/modules/05-hak-akses.md` + `_shared/permission-matrix.md`.
- `_shared/conventions-app.md` §5 (PermissionService, permissionGuard, `*hasPermission`, dynamic sidebar, anti-escalation).

## 2. Scope (contract, `hak_akses.*` = Super Admin only)
| Method | Path |
|--------|------|
| GET | `/modul`, `/permissions`, `/roles`, `/roles/:id/permissions` |
| POST/PUT/DELETE | `/roles`, `/roles/:id` |
| PUT | `/roles/:id/permissions` `{permissions:[{permission_id, cakupan}]}` |

## 3. Deliverables
- **`shared/directives/has-permission.directive.ts`** (`*hasPermission`) — conventions-app §5.4 (effect-based, zoneless-safe).
- **`core/guards/permission.guard.ts`** — `data.permission` gate → `/forbidden` (conventions-app §5.3).
- **dynamic sidebar** (`layouts/sidebar`) — fetch `/modul`, filter to modules with `<kode>.read`, group by `grup` (Master Data / Konten / Sistem / Operasional). Never hard-code entries.
- **`pages/hak-akses/`**:
  - `role-list.*` — roles with level + is_super; create/edit/delete (respect anti-escalation).
  - `role-matrix.*` — **the grid**: rows = modules (from `/modul`+`/permissions`), columns = C/R/U/D (+ any special aksi like `assign/override/print/...`), each checked cell also carries a **cakupan** select (`semua`/`instansi_sendiri`/`milik_sendiri`) where meaningful. **Disable** permissions the current actor does not hold; disable editing roles with `level <= yours`; block deleting the last super admin. Save → `PUT /roles/:id/permissions`.
- After save, reload permissions (`PermissionService.load()`) so the UI reflects the bumped `perm_version`.

## 4. Routing
`{ path:'hak-akses', canActivate:[authGuard, permissionGuard], data:{permission:'hak_akses.read'}, children:[...] }`.

## 5. i18n
Namespace `HAK_AKSES` (`ROLE`, `LEVEL`, `MATRIX`, `CAKUPAN.SEMUA/INSTANSI/MILIK`, `C/R/U/D` labels) + `COMMON.*`. IND & ENG in sync.

## 6. Permission-gating reminder
`*hasPermission`/`permissionGuard` are **UX only** — the Go handler enforces `hak_akses.*` and anti-escalation. Hiding a checkbox is not security.

## 7. Verification checklist
- [ ] `npm run build:local` compiles.
- [ ] Grid renders role × modul with C/R/U/D + cakupan; disabled cells for permissions the actor lacks; can't edit a role with `level <= yours`; last super admin protected.
- [ ] Save issues `PUT /roles/:id/permissions`; permissions reload afterward.
- [ ] Dynamic sidebar shows only modules the user can `.read`, grouped by `grup`.
- [ ] `*hasPermission` shows/hides controls; `permissionGuard` redirects unauthorized.
- [ ] SA-only route; dark theme + red accent + focus ring; IND/ENG keys match; grid built first; taste-skill announced.
