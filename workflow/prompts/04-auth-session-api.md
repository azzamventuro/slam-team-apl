# PROMPT — Complete `04-auth-session` API: real users domain + sesi_login + refresh/logout (Fase 0)

> **Repo-accurate (revised 2026-09-12).** This module is **PARTIALLY built** — the real `users` table + RBAC + `create-superadmin` already exist; the auth *module code* is still a placeholder and `sesi_login` is missing. You are FILLING THE GAP, not starting over. Read the existing code before editing.

Working directory: `D:/xampp/htdocs/slam-team-apl/slam-team-api`.

## 0. Read first (authoritative)
1. `slamteam_db.dbml` — `users` (line 424, ALREADY migrated in `0005`) and `sesi_login` (line 400, NOT yet created). DBML WINS.
2. `workflow/modules/04-auth-session.md`; `_shared/api-endpoints.md` §1; `_shared/conventions-api.md` §2/§3/§6/§12.
3. Existing code you MUST align with (read first):
   - `internal/modules/core/auth/{domain,dto,repository,service,handler,main.auth.go}` — currently a PLACEHOLDER (`domain.User` = `Name`/`Email` only; login by email; no refresh/logout/sesi_login). `service.MyPermissions` is RBAC-integrated — KEEP it. Routes are mounted under the `/auth` group (`/auth/login`, `/auth/me`, `/auth/me/permissions`).
   - `pkg/jwt/jwt.go` — `Claims{UserID,Email,Name,RoleID,RoleLevel,IsSuper,PermVersion}` + `Issue`/`IssueWithRole`. **No `AnggotaID` yet.**
   - `internal/shared/{audit,model,response,apperr}`, `internal/middleware/{auth,permission}`.
   - `migrations/0004` (user_role), `0005` (users), `0006` (users.anggota_id FK); `cmd/slamctl/superadmin.go`.

## 1. Already done — DO NOT recreate
`users` table (full schema, 0005) + all FKs; `user_role` (0004); RBAC + PermGuard + `seed_rbac`; `slamctl create-superadmin`; auth wired with hakakses repo + permGuard; `GET /auth/me`, `GET /auth/me/permissions`.

## 2. Gap to fill

### 2a. Real `users` domain (replace placeholder)
Rewrite `auth/domain/user.go` to match `users` EXACTLY: `anggota_id *int64`, `username`, `email`, `password (json:"-")`, `role_id int64`, `timezone`, `is_aktif bool`, `login_terakhir *time.Time`, `password_diubah *time.Time`, `gagal_login int`, `terkunci_sampai *time.Time`, embed `internal/shared/model.Audit`. **Remove the `Name` field** (name lives on `anggota`). Add a `SesiLogin` entity for `sesi_login`. Fix every reference (`toUserResponse`, service): `nama` now comes from a join to `anggota.nama_lengkap`. If any OTHER module imports `auth/domain.User`, reconcile (usermgmt must use its own users entity).

### 2b. Migration `0013_sesi_login.up.sql` (+ `.down.sql`)
Create `sesi_login` per `.dbml`: `id`, `user_id` → `users.id` **ON DELETE CASCADE**, `refresh_token varchar(255)` UNIQUE, `ip_address inet`, `user_agent text`, `info_perangkat jsonb`, `berlaku_sampai timestamptz NOT NULL`, `dicabut_pada timestamptz`, `created_at timestamptz DEFAULT now()`. Index `user_id`. `.down` drops it. **`users` already exists — do NOT touch it.** Use the next free number `0013`.

### 2c. Repository
`FindByUsernameOrEmail(id string)` → `WHERE (username=? OR email=?) AND is_deleted=false` (single query); `TouchLogin`, `IncrementGagalLogin`, `ResetGagalLogin`, `LockUntil(userID,t)`; `CreateSesi`, `FindActiveSesiByToken` (not revoked, `berlaku_sampai > now`), `RevokeSesi(token)`, `RevokeAllSesi(userID)`.

### 2d. DTO (rancangan Bab 9 / api-endpoints §1)
- `LoginRequest{ Identifier string binding:"required"; Password string binding:"required,min=8" }` (username OR email — replaces email-only).
- `RefreshRequest{ RefreshToken string binding:"required" }`.
- `AuthResponse{ AccessToken, RefreshToken, TokenType:"Bearer", ExpiresIn int, User AuthUser }`, `AuthUser{ ID, Username, Email, AnggotaID, Nama, Role{ID,Nama,Level,IsSuper}, PermVersion }`.

### 2e. JWT claims — add `AnggotaID`
Add `AnggotaID *int64` to `jwt.Claims` (json `anggota_id,omitempty`) and thread it through `IssueWithRole` (or a new `IssueAccess`). Additive/backward-compatible; unblocks `milik_sendiri` cakupan in other modules. Pass the user's `anggota_id` at login.

### 2f. Service
- **Login(identifier,password):** `FindByUsernameOrEmail`; if `terkunci_sampai > now` → forbidden ("akun terkunci"); if `!is_aktif` → forbidden; `utils.CheckPassword`; on fail → `IncrementGagalLogin` (count ≥ 5 → `LockUntil(now+15m)`), return **generic** `ErrInvalidCredentials` (jangan sebut field mana). On success → `ResetGagalLogin` + `TouchLogin`; issue access JWT (WITH `anggota_id` + KEEP existing role/permVersion via hakakses) + opaque refresh (`crypto/rand`, ≥32B) + `CreateSesi` (ip/user_agent dari `*gin.Context`, `berlaku_sampai = now + refreshTTL`); `audit.Write(modul="auth", aksi="login")`. Return `AuthResponse`.
- **Refresh:** `FindActiveSesiByToken`; invalid → `ErrUnauthorized`; issue new access + new refresh; `RevokeSesi(old)` (rotation).
- **Logout:** revoke the current session — accept the `refresh_token` in the body and `RevokeSesi` (or embed a `sid` claim and revoke by it).
- Lockout threshold: constant `N=5` / `15m` (optionally read a new `mst_pengaturan` key — not required; note it if added).

### 2g. Handler + routes (main.auth.go)
Add `POST /auth/refresh` (public), `POST /auth/logout` (bearer). KEEP `POST /auth/login`, `GET /auth/me`, `GET /auth/me/permissions`. Enrich `/auth/me` to return user + linked **anggota** summary (`nama_lengkap`, foto uuid) via join.

## 3. Conventions
Envelope via `response`; sentinels via `apperr` + `response.FromError`; never return `password`; generic login error; migration numbered **0013**; no AutoMigrate; `audit.Write` for login.

## 4. Verification checklist
- [ ] `go build ./...` + `go vet ./...` pass; `slamctl migrate up`/`down` clean (0013 up/down; `users` untouched).
- [ ] `create-superadmin`, then **login by username AND by email** both succeed (single query).
- [ ] Wrong password → 401 generic; 5 failures → `terkunci_sampai` set + login blocked until expiry.
- [ ] Access token carries `user_id`, **`anggota_id`**, `role_id`, `is_super`, `perm_version`; a `sesi_login` row is created on login.
- [ ] `POST /auth/refresh` rotates (old revoked, new works); `POST /auth/logout` revokes the active session.
- [ ] `/auth/me` returns user + anggota summary; `/auth/me/permissions` behavior unchanged.
- [ ] Password never leaked; `audit.Write` logged login; no `Name`-column references remain.
