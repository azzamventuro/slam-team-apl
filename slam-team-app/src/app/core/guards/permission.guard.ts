import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { PermissionService } from '../services/permission.service';

/**
 * Data-driven route guard. Reads `route.data['permission']` and checks
 * whether the current user holds that permission code.
 *
 * Usage in routes.ts:
 * ```ts
 * { path: 'anggota', canActivate: [authGuard, permissionGuard],
 *   data: { permission: 'anggota.read' }, ... }
 * ```
 *
 * Hiding the link is UX only — the Go handler enforces permissions on every
 * request.
 */
export const permissionGuard: CanActivateFn = (route) => {
  const perms = inject(PermissionService);
  const router = inject(Router);
  const need = route.data?.['permission'] as string | undefined;
  if (!need || perms.can(need)) return true;
  return router.createUrlTree(['/forbidden']);
};
