import { Injectable, inject, signal } from '@angular/core';
import { Observable, tap } from 'rxjs';

import { MePermissions } from '../models/user.model';
import { ApiService } from './api.service';

/**
 * PermissionService owns the user's effective permission set.
 *
 * After login, the app calls `load()` which fetches `GET /auth/me/permissions`.
 * The response contains a flat list of `"<modul>.<aksi>"` codes (e.g.
 * `"anggota.read"`, `"file.delete"`) and a `perm_version` counter that
 * increases whenever the role_permission matrix changes.
 *
 * `can(code)` is the single gate used by the sidebar, permissionGuard, and
 * the `*hasPermission` directive. Super admins (`is_super = true`) bypass
 * every check — `can()` returns `true` for any code.
 *
 * Hiding UI elements is UX only — the Go handler enforces permissions on
 * every request.
 */
@Injectable({ providedIn: 'root' })
export class PermissionService {
  private api = inject(ApiService);

  private readonly _perms = signal<Set<string>>(new Set());
  private readonly _super = signal(false);
  private readonly _permVersion = signal(0);

  /** After load(), the permission set is ready for `can()` checks. */
  readonly ready = signal(false);

  /** Current permission version from the backend. */
  readonly permVersion = this._permVersion.asReadonly();

  /** Fetch the effective permission set from the API. */
  load(): Observable<MePermissions> {
    return this.api.get<MePermissions>('/auth/me/permissions').pipe(
      tap((p) => {
        this._super.set(p.is_super);
        this._perms.set(new Set(p.permissions));
        this._permVersion.set(p.perm_version);
        this.ready.set(true);
      }),
    );
  }

  /** Returns `true` when the current user holds the given permission code. */
  can = (code: string): boolean => this._super() || this._perms().has(code);

  /** Clear the permission state (on logout). */
  clear(): void {
    this._perms.set(new Set());
    this._super.set(false);
    this._permVersion.set(0);
    this.ready.set(false);
  }
}
