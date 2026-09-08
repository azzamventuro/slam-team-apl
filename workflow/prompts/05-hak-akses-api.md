# PROMPT — Build API dynamic RBAC (`hak_akses`, 5 tables + PermGuard) (Fase 1)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-api`.

---

You are building the **dynamic RBAC** that every later module depends on. Fase 1.

## 0. Read first
1. `D:/xampp/htdocs/slam-team-apl/slamteam_db.dbml` — **`mst_modul`** (168), **`mst_permission`** (191), **`mst_role`** (205), **`role_permission`** (232), **`user_role`** (248), relations (1164–1169). DBML WINS.
2. `D:/xampp/htdocs/slam-team-apl/workflow/modules/05-hak-akses.md`.
3. `D:/xampp/htdocs/slam-team-apl/workflow/_shared/permission-matrix.md` — **the SEED matrix** (Bab 3.5) you will insert.
4. `_shared/conventions-api.md` §8 (RequirePermission, perm cache, perm_version, cakupan) + `_shared/api-endpoints.md` §2.

## 1. Scope + endpoints
| Method | Path | Permission |
|--------|------|-----------|
| GET | `/modul` | `hak_akses.read` (also feeds sidebar) |
| GET | `/permissions` | `hak_akses.read` |
| GET | `/roles` | `hak_akses.read` |
| POST | `/roles` | `hak_akses.create` |
| PUT | `/roles/:id` | `hak_akses.update` (anti-escalation) |
| DELETE | `/roles/:id` | `hak_akses.delete` (last super admin protected) |
| GET | `/roles/:id/permissions` | `hak_akses.read` |
| PUT | `/roles/:id/permissions` | `hak_akses.update` (replace matrix; bump perm_version) |

`hak_akses` = **Super Admin only**.

## 2. Deliverables
- **PermGuard middleware** (`internal/middleware/permission.go`): `Require(perm)` — `is_super` bypass; load effective `map[perm]cakupan` cached per role (`perm:role:{id}`, Redis if present else in-process), keyed with `perm_version`; deny → `response.Forbidden`; set `c.Set("cakupan:"+perm, cak)` for handlers. Add `perm_version` bump on any RBAC change.
- **response additions** (if not already from Fase 0): `Forbidden`, `NotFound`, `FromError`.
- **module `internal/modules/core/hakakses/`**: domain (`Modul`, `Permission`, `Role`, `RolePermission`, `UserRole`); repository; service with **anti-escalation** (cannot grant a permission you lack; cannot edit a role whose `level <= yours`; last `is_super` role cannot be deleted/disabled/demoted; only super admins manage roles); handler; `RequirePermission("hak_akses.*")`.
- `PUT /roles/:id/permissions` request: `{permissions:[{permission_id, cakupan}]}` → replace matrix, bump `perm_version`, `audit.Write`.

## 3. Migration + seeder
- 5 tables EXACTLY per `.dbml` (enum `aksi_permission`, `cakupan_permission`; `role_permission.role_id` cascade + `UNIQUE(role_id,permission_id)`; `user_role.user_id` cascade + `UNIQUE(user_id,role_id)`; `mst_role` soft-delete). Order after `0001_enums`, before `users` (FK `users.role_id`).
- **Idempotent seeder** (`slamctl seed rbac`): 19 `mst_modul`, all `mst_permission` (`modul.aksi`), 5 `mst_role` (levels SA=0/Admin=10/Mod=20/User=30/Guest=99, `is_super` on SA), and `role_permission` rows for the **full matrix in `_shared/permission-matrix.md`** with the correct `cakupan` (e.g. `absensi.read` = `milik_sendiri` for User; `file.delete` = `milik_sendiri` for Mod/User). `ON CONFLICT` keyed on natural unique columns.

## 4. Conventions
Envelope; no AutoMigrate; cakupan enforced in services (reads AND writes); `is_super` bypass; `perm_version` invalidation; anti-escalation in service (not just UI).

## 5. Verification checklist
- [ ] `go build`/`go vet` pass; migration up/down clean; matches `.dbml`.
- [ ] Seeder run twice → 19 modul, full permission + role_permission matrix, no dupes; cakupan values correct.
- [ ] `PermGuard.Require` denies 403 without the permission; `is_super` bypasses; cached per role.
- [ ] `PUT /roles/:id/permissions` replaces the matrix and **bumps `perm_version`**; stale tokens reload.
- [ ] Anti-escalation enforced: cannot grant a permission you lack; cannot edit a role with `level <= yours`; last super admin cannot be deleted.
- [ ] `hak_akses.*` allowed only for Super Admin; `audit.Write` on changes.
