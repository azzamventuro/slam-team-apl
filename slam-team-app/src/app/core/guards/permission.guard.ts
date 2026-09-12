import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { filter, map, race, timer } from 'rxjs';

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
  if (!need) return true;

  // If permissions are already loaded, check immediately.
  if (perms.ready()) return perms.can(need);

  // Permissions not loaded yet (page reload or first navigation after login).
  // Trigger a load if not already in-flight, then wait up to 3 s.
  perms.ensureLoaded();

  return race(
    perms.ready$.pipe(filter(Boolean)),
    timer(3000),
  ).pipe(map(() => perms.can(need) || router.createUrlTree(['/forbidden'])));
};
