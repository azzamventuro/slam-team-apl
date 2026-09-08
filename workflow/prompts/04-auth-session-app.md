# PROMPT — Extend APP login + AuthService/PermissionService (Fase 0)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit on the existing login screen, enforce the SLAM tokens in `_shared/design-tokens.md` (dark, SLAM red, UPPERCASE headings, red focus ring, 44px targets), Bootstrap 5. *If `blog-fe` restored mirror it; else follow tokens.*

## 1. Read first
- `D:/xampp/htdocs/slam-team-apl/workflow/modules/04-auth-session.md`.
- `_shared/conventions-app.md` §5 (auth + dynamic RBAC: AuthService, PermissionService, guards) + §4 (interceptors).
- The existing `pages/auth/login`, `core/services/auth.service.ts`, `core/interceptors/*` — **extend, don't rebuild**.

## 2. Scope
Wire the real auth flow against the backend (envelope `{success,message,data,errors}`):

| Method | Path |
|--------|------|
| POST | `/auth/login` `{identifier, password}` |
| POST | `/auth/refresh` `{refresh_token}` |
| POST | `/auth/logout` |
| GET | `/me`, `/me/permissions` |

## 3. Work items
- **`pages/auth/login`** (extend): reactive form `identifier` + `password` (required, min 8); submit → `AuthService.login`; on success store `slam_token`/`slam_user`, call `PermissionService.load()`, navigate `/dashboard`; on 401 show a **generic** error (`AUTH.LOGIN.ERROR_INVALID`); on locked show `AUTH.LOGIN.ERROR_LOCKED`. Disable submit while `saving()`.
- **`AuthService`** (extend): `login(identifier, password)`, `refresh()`, `logout()`, `user` signal, `token()`. Persist token/user in `localStorage` (`slam_token`/`slam_user`); hydrate `user` on boot.
- **`PermissionService`** (`core/services/permission.service.ts`): `load()` from `/me/permissions`, holds a `Set<string>` + `is_super`, exposes `can(code)`; called after login and on boot; `clear()` on logout.
- **interceptors**: `authInterceptor` attaches Bearer (existing). In `errorInterceptor`, optionally attempt **one** `refresh()` on 401 before logout+redirect; 403 → toast, no logout.

## 4. i18n
Namespace `AUTH` (`LOGIN.TITLE`, `.IDENTIFIER`, `.PASSWORD`, `.SUBMIT`, `.ERROR_INVALID`, `.ERROR_LOCKED`) + `COMMON.*`. IND & ENG in sync.

## 5. Verification checklist
- [ ] `npm run build:local` compiles.
- [ ] Login with username OR email works; token + user stored; `PermissionService.load()` runs after login.
- [ ] Wrong credentials → generic inline error (no field disclosure); locked account → locked message.
- [ ] Refresh path works (token rotates); logout clears token/user + permissions and redirects to login.
- [ ] Boot hydrates user + permissions from storage/`/me`.
- [ ] Dark theme + red accent + focus ring; IND/ENG keys match; existing files extended; taste-skill announced.
