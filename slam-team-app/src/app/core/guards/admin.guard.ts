import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { AuthService } from '../services/auth.service';
import { Role } from '../models/user.model';

/** mst_role.level for Admin — mirrors internal/middleware/role.go's LevelAdmin. */
const LEVEL_ADMIN = 10;

/**
 * Role-based gate for system pages that have no seeded modul.aksi permission
 * (pengaturan is not one of the 19 mst_modul rows, so permissionGuard has
 * nothing to look up). Allows super admin, or a role at Admin level or above —
 * mirrors the backend's middleware.RequireAdmin exactly. Hiding here is UX
 * only; the API enforces this again.
 */
export const adminGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);
  return isAdmin(auth.user()?.role) ? true : router.createUrlTree(['/dashboard']);
};

/** A missing role (claims-less token) must be denied, never treated as admin. */
export function isAdmin(role: Role | undefined): boolean {
  if (!role) return false;
  return role.is_super || (role.level > 0 && role.level <= LEVEL_ADMIN);
}
