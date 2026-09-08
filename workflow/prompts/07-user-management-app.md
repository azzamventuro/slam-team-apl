# PROMPT — Build APP user-management (`admin`/`moderator`/`user`) (Fase 2)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce SLAM tokens (`_shared/design-tokens.md`: dark, SLAM red, UPPERCASE headings, red focus ring), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
- `D:/xampp/htdocs/slam-team-apl/workflow/modules/07-user-management.md` + `_shared/permission-matrix.md`.
- `_shared/conventions-app.md` §5 (anti-escalation UI), §6 (forms), §9 (list).

## 2. Scope
Three surfaces over `users` (route bases `/admin`, `/moderator`, `/user`) — same UI parameterized by role band. Contract per base: `GET/POST/PUT/DELETE /<base>[/:id]`. Bands per matrix (admin: SA CRUD, others R; moderator: SA/Admin CRUD; user: SA/Admin CRUD, Mod CRU).

## 3. Files under `src/app/pages/user-management/`
- `user.model.ts` — `UserRow` (id, username, email, anggota_nama, role{id,nama,level}, is_aktif), `UserForm`.
- `user.service.ts` — parameterized by base (`admin|moderator|user`).
- `user-list.ts/.html` — signal list (username, email, anggota, role, active toggle, actions); one component reused for all three bases (base from route data).
- `user-form.ts/.html` — create/edit: **anggota picker** (from `/anggota`), username/email, password (set on create / reset on edit), **role select limited to bands below the actor** (mirror anti-escalation — disable out-of-range roles), active toggle. Map `422` inline.

## 4. Routing
Three routes, each `canActivate:[authGuard, permissionGuard]` with `data:{permission:'admin.read'|'moderator.read'|'user.read', base:'admin'|'moderator'|'user'}`.

## 5. i18n
Namespace `USERMGMT` (`ADMIN.TITLE`, `MODERATOR.TITLE`, `USER.TITLE`, `FORM.USERNAME/EMAIL/PASSWORD/ROLE/ANGGOTA/AKTIF`, `RESET_PASSWORD`) + `COMMON.*`. IND & ENG in sync.

## 6. Permission-gating reminder
Disabling out-of-range roles is UX; the Go handler enforces anti-escalation (`role.level <= yours` blocked, last super admin protected).

## 7. Verification checklist
- [ ] `npm run build:local` compiles.
- [ ] Each base lists its band; create makes a user (verify) with hashed password; edit/reset works.
- [ ] Role select disables roles at/above the actor's level; last super admin cannot be deleted in UI (and backend blocks).
- [ ] Band buttons honor the matrix (e.g. Admin on `/admin` read-only); backend 403 on out-of-band writes.
- [ ] Dark theme + red accent + focus ring; IND/ENG keys match; taste-skill announced.
