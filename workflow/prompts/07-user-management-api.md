# PROMPT — Build API user-management surfaces `admin`/`moderator`/`user` (Fase 2)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

Build the **user-management** surfaces (`admin`, `moderator`, `user`) over `users` + `user_role`, filtered by role band. NOT separate tables.

## 0. Read first
1. `D:/xampp/htdocs/slam-team-apl/slamteam_db.dbml` — **`users`** (424), **`user_role`** (248), **`mst_role`** (205). DBML WINS.
2. `D:/xampp/htdocs/slam-team-apl/workflow/modules/07-user-management.md`.
3. `_shared/permission-matrix.md` (bands: `admin` SA-CRUD/others-R; `moderator` SA+Admin-CRUD; `user` SA+Admin-CRUD, Mod-CRU) + `_shared/api-endpoints.md` §13 user-management + `_shared/conventions-api.md` §8 (anti-escalation).

## 1. Scope + endpoints
Same CRUD shape at three route bases, each filtered to a target role band:
| Base | Modul | Bands (SA/Admin/Mod/User) |
|------|-------|---------------------------|
| `/admin` | `admin` | CRUD / R / R / R |
| `/moderator` | `moderator` | CRUD / CRUD / R / R |
| `/user` | `user` | CRUD / CRUD / CRU / R |

Each: `GET /<base>`, `GET /<base>/:id`, `POST /<base>`, `PUT /<base>/:id`, `DELETE /<base>/:id`, guarded `RequirePermission("<modul>.<aksi>")`.

## 2. Files under `internal/modules/core/usermgmt/`
- **domain**: reuse `User` + `UserRole` (or import from auth domain).
- **dto**: `CreateUserReq{anggota_id required,gt=0; username required,max=50; email required,email; password required,min=8; role_id required,gt=0; timezone; is_aktif}`, `UpdateUserReq` (password optional; reset flow), `UserResp` (no password; include `anggota_nama`, `role`).
- **repository**: `List(band, query)` — join `mst_role` filter by target band's level range; `Create` (create `users` + `user_role`); `Update`, `SoftDelete`, `SetActive`, `ResetPassword` (bcrypt via `pkg/utils`). Partial-unique username/email `WHERE is_deleted=false`.
- **service — ANTI-ESCALATION (enforced here, not just UI):** cannot create/edit/assign a user whose `role.level <= claims.role.level`; cannot grant a role you cannot manage; the **last super admin cannot be deleted/disabled/demoted**; usernames/emails unique. Hash password on create/reset. `audit.Write` on all changes. No self-registration — only reachable here.
- **handler/main/router**: three route groups (`admin`,`moderator`,`user`) sharing the service with a band param; each guarded by its own module permission.

## 3. Migration + seeder
- No new tables (uses `users`/`user_role` from auth/RBAC). Confirm permissions `admin.*`,`moderator.*`,`user.*` seeded per matrix.

## 4. Conventions
Envelope; never return password; anti-escalation server-side; bcrypt; soft-delete; `audit.Write`.

## 5. Verification checklist
- [ ] `go build`/`go vet` pass.
- [ ] Each base lists only users in its target role band; create makes a `users` + `user_role` row; password hashed.
- [ ] Anti-escalation: cannot create/edit a user with `role.level <= yours`; last super admin cannot be deleted/disabled.
- [ ] Bands enforced: e.g. Moderator on `/user` = CRU (no delete); Admin on `/admin` = read-only.
- [ ] Username/email uniqueness (partial); no password leaked; `audit.Write` on changes.
