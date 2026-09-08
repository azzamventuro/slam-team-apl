# PROMPT — Build API module `auth` + `sesi_login` (Fase 0)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

You are completing the **auth** module of slam-team-api. A toy `auth` module already exists (login + `/me`) — **replace its placeholder `users` table with the real DBML schema** and add `sesi_login`, refresh, logout, and permissions.

## 0. Read first
1. `D:/xampp/htdocs/slam-team-apl/slamteam_db.dbml` — **`users`** (line 424) and **`sesi_login`** (line 400); relations (line 1167–1176: `users.anggota_id → anggota.id` restrict, `users.role_id → mst_role.id` restrict, `sesi_login.user_id → users.id` cascade). DBML WINS.
2. `D:/xampp/htdocs/slam-team-apl/workflow/modules/04-auth-session.md`.
3. `_shared/api-endpoints.md` §1 (login/refresh/logout/me/permissions) + `_shared/conventions-api.md`.
4. The existing `internal/modules/core/auth/` (reference/extend, don't rebuild the wiring).

## 1. Scope + endpoints
| Method | Path | Auth |
|--------|------|------|
| POST | `/auth/login` | public — `{identifier, password}` (username OR email) |
| POST | `/auth/refresh` | public — `{refresh_token}` (rotates) |
| POST | `/auth/logout` | bearer (revoke current `sesi_login`) |
| GET | `/me` | bearer |
| GET | `/me/permissions` | bearer → `{is_super, permissions[]}` |

## 2. Files under `internal/modules/core/auth/`
- **domain**: `User` (`users`, `Password` tagged `json:"-"`), `SesiLogin` (`sesi_login`).
- **dto**: `LoginReq{Identifier required; Password required,min=8}`, `RefreshReq{RefreshToken required}`, `AuthResp` (access_token, refresh_token, token_type, expires_in, user{...role, perm_version}), `MeResp`, `MePermissionsResp`.
- **repository**: `FindByUsernameOrEmail(id)` → `WHERE (username=? OR email=?) AND is_deleted=false` (single query); `TouchLogin`, `IncrementGagalLogin`, `ResetGagalLogin`, `LockUntil`; `CreateSesi`, `FindSesiByToken` (active, not revoked/expired), `RevokeSesi`, `RevokeAllSesi(userID)`; `EffectivePermissions(roleID)` (or delegate to the Fase-1 PermGuard cache).
- **service**:
  - **Login**: locked (`terkunci_sampai > now`) → `ErrForbidden`; `!is_aktif` → `ErrForbidden`; `utils.CheckPassword`; on fail `IncrementGagalLogin` (lock after N from `mst_pengaturan`) → **generic** `ErrUnauthorized` ("kredensial salah" — never reveal which field); on success reset gagal, `TouchLogin`, issue access JWT (embed `user_id, anggota_id, role_id, perm_version, is_super, exp`) + opaque refresh, `CreateSesi`; `audit.Write(modul="auth", aksi="login")`.
  - **Refresh**: `FindSesiByToken` → issue new access + new refresh, `RevokeSesi` old (rotation); invalid → `ErrUnauthorized`.
  - **Logout**: `RevokeSesi` current.
  - **/me**: join `anggota` (nama, foto uuid) + `mst_role`. **/me/permissions**: effective set + `is_super`.
- **main.auth.go + router**: add the refresh/logout/me/permissions routes (login already wired).

## 3. Migration + seeder
- `users` + `sesi_login` EXACTLY per `.dbml`: **partial-unique** `username`/`email` `WHERE is_deleted=false`; FK `anggota_id`/`role_id` **restrict**; `sesi_login.user_id → users.id` **cascade**. Order **after** `mst_role` (Fase 1 RBAC) and `anggota`.
- `slamctl create-superadmin` (from Fase 0) creates the first super admin; there is **no HTTP register** and **no super-admin endpoint**.

## 4. Conventions
Envelope; no AutoMigrate; never return `password`; generic login errors; JWT secret stable; `audit.Write` for login.

## 5. Verification checklist
- [ ] `go build`/`go vet` pass; migration up/down clean; matches `.dbml` (partial-unique, FK actions).
- [ ] Login by **username** AND by **email** both work (single query).
- [ ] Wrong credentials → 401 generic; N failures → account locked (`terkunci_sampai`).
- [ ] Access token embeds `user_id/anggota_id/role_id/perm_version/is_super`; a `sesi_login` row is created on login.
- [ ] Refresh rotates (old revoked); logout revokes the active session.
- [ ] `/me` returns user + anggota + role; `/me/permissions` returns `is_super` + codes.
- [ ] Password hash never leaked; `audit.Write` logged login.
