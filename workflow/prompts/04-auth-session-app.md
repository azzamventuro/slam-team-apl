# PROMPT — Complete `04-auth-session` APP: real login + refresh/logout wiring (Fase 0)

> **Repo-accurate (revised 2026-09-12).** Auth UI is PARTIALLY built (toy login by email). You are ALIGNING it to the new API contract from `04-auth-session` (API) — **extend, don't rebuild**. `PermissionService`, the guards, and `HasPermissionDirective` already exist (from `05-hak-akses`) — reuse them.

Working directory: `D:/xampp/htdocs/slam-team-apl/slam-team-app`.

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce the SLAM tokens in `_shared/design-tokens.md` (dark, SLAM red, UPPERCASE headings, red focus ring, 44px targets), Bootstrap 5. *If `blog-fe` restored, mirror it; else follow tokens.*

## 1. Read first / align with existing code
- `workflow/modules/04-auth-session.md`; `_shared/conventions-app.md` §4 (interceptors) + §5 (auth + RBAC).
- Existing (DO NOT rebuild): `core/services/auth.service.ts` (toy: posts `email`, reads `token`), `core/services/permission.service.ts` (loads perms — KEEP, just reload after login), `core/guards/{auth,permission,admin}.guard.ts`, `shared/directives` `HasPermissionDirective`, `pages/auth/login`, `core/interceptors/{auth,error}`.

## 2. New backend contract (from #7) — routes are under `/auth`
- POST `/auth/login` `{identifier, password}` → `{access_token, refresh_token, token_type, expires_in, user{id, username, email, anggota_id, nama, role{id,nama,level,is_super}, perm_version}}`
- POST `/auth/refresh` `{refresh_token}` → same shape (rotates the refresh token)
- POST `/auth/logout` (bearer, body `{refresh_token}`) → revoke session
- GET `/auth/me`, GET `/auth/me/permissions` (already consumed by PermissionService)

## 3. Gap to fill

### 3a. AuthService (extend `core/services/auth.service.ts`)
- `login(identifier, password)` → POST `/auth/login`; store `slam_token` (access), `slam_refresh` (refresh), `slam_user`; set the `user` signal. Expose `token()` and `refreshToken()`.
- `refresh()` → POST `/auth/refresh {refresh_token}`; rotate the stored access + refresh; on failure clear + throw.
- `logout()` → POST `/auth/logout {refresh_token}` (best-effort), then clear storage + `PermissionService.clear()` + navigate `/auth/login`.
- **Boot hydration**: if `slam_token` present on startup, load `/auth/me` into `user` + `PermissionService.load()`.

### 3b. Login page (`pages/auth/login`, extend)
- Replace the `email` control with **`identifier`** (username OR email); keep password (`min 8`). On submit → `AuthService.login`; on success store tokens + `PermissionService.load()` → navigate `/dashboard`. On **401** show generic `AUTH.LOGIN.ERROR_INVALID` (jangan sebut field mana); on **403 locked** show `AUTH.LOGIN.ERROR_LOCKED`. Disable submit while `saving()`.

### 3c. Interceptors
- `errorInterceptor`: on **401** for a non-`/auth/*` request, attempt **one** silent `AuthService.refresh()`, then retry the original request with the new access token; if refresh fails → `logout()` + redirect. **403** → toast, no logout (unchanged).
- `authInterceptor`: keep attaching Bearer from `slam_token`.

## 4. i18n
`AUTH.LOGIN.TITLE`, `.IDENTIFIER`, `.PASSWORD`, `.SUBMIT`, `.ERROR_INVALID`, `.ERROR_LOCKED` + `COMMON.*`. IND & ENG in sync.

## 5. Verification checklist
- [ ] `npm run build:local` compiles.
- [ ] Login with **username** and with **email** both work; access + refresh tokens stored; `PermissionService` reloaded after login; dynamic sidebar + `*hasPermission` reflect the role.
- [ ] Wrong credentials → generic inline error; locked account → locked message.
- [ ] An expired access token triggers **one** silent `/auth/refresh` that recovers the in-flight request; a failed refresh logs out + redirects to login.
- [ ] Logout calls `/auth/logout`, clears tokens + permissions, redirects to `/auth/login`.
- [ ] Existing `PermissionService`/guards/`HasPermissionDirective` reused (not rebuilt); dark theme + red accent + focus ring; IND/ENG key sets match; "Using design-taste-frontend" announced.
