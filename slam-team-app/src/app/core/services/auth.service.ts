import { Injectable, computed, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import {
  EMPTY,
  Observable,
  catchError,
  finalize,
  map,
  share,
  switchMap,
  tap,
  throwError,
} from 'rxjs';

import { AuthResponse, Me, User } from '../models/user.model';
import { ApiService } from './api.service';
import { PermissionService } from './permission.service';

const TOKEN_KEY = 'slam_token';
const REFRESH_KEY = 'slam_refresh';
const USER_KEY = 'slam_user';

/**
 * AuthService owns the session: the access JWT and the opaque refresh token
 * (both persisted to localStorage) and the current user (a signal, so zoneless
 * change detection just works). HTTP goes through ApiService, which unwraps the
 * envelope — every method here emits the `data` payload, not the envelope.
 *
 * Lifecycle, in the order it happens:
 *   login    → POST /auth/login    → store pair + user → PermissionService.load()
 *   refresh  → POST /auth/refresh  → rotate the pair (errorInterceptor calls this once on a 401)
 *   logout   → POST /auth/logout   → best-effort revoke, then clear + redirect
 *   hydrate  → GET  /auth/me       → on boot, when a token survived a reload
 *
 * The Bearer header itself is attached by authInterceptor, never here.
 */
@Injectable({ providedIn: 'root' })
export class AuthService {
  private api = inject(ApiService);
  private perms = inject(PermissionService);
  private router = inject(Router);

  private readonly _user = signal<User | null>(readStoredUser());
  readonly user = this._user.asReadonly();
  readonly isLoggedIn = computed(() => this._user() !== null);

  /**
   * The refresh in flight, if any. Several requests failing with 401 at the
   * same moment (a page with three lists on it) must share ONE refresh: a
   * rotated token is single-use, so a second concurrent refresh would be
   * refused and log everyone out.
   */
  private refreshing: Observable<AuthResponse> | null = null;

  /**
   * `identifier` is the username OR the email — the repository matches both
   * in one query. The login payload already carries the user with its role,
   * so no second round trip to /auth/me is needed; permissions are loaded
   * before the observable completes, so the caller can navigate straight away
   * and the sidebar / `*hasPermission` are correct on first paint.
   */
  login(identifier: string, password: string): Observable<User> {
    return this.api.post<AuthResponse>('/auth/login', { identifier, password }).pipe(
      tap((res) => this.setSession(res)),
      switchMap(() => this.perms.load()),
      map(() => this._user()!),
    );
  }

  /**
   * Rotates the token pair. On any failure the session is cleared and the
   * error re-thrown, so the interceptor's fallback (logout + redirect) fires.
   * Concurrent callers get the same in-flight request (see `refreshing`).
   */
  refresh(): Observable<AuthResponse> {
    if (this.refreshing) return this.refreshing;

    const refreshToken = this.refreshToken();
    if (!refreshToken) return throwError(() => new Error('no refresh token'));

    this.refreshing = this.api
      .post<AuthResponse>('/auth/refresh', { refresh_token: refreshToken })
      .pipe(
        tap((res) => this.setSession(res)),
        catchError((err: unknown) => {
          this.clearSession();
          return throwError(() => err);
        }),
        finalize(() => (this.refreshing = null)),
        share(),
      );
    return this.refreshing;
  }

  /**
   * Revokes the session server-side (best effort — an expired token or a dead
   * network must not keep the user signed in locally), then clears storage
   * and permissions and returns to the login page.
   *
   * The request is subscribed BEFORE storage is cleared: interceptors run on
   * subscribe, so that is when authInterceptor reads the Bearer.
   */
  logout(): void {
    const refreshToken = this.refreshToken();
    if (refreshToken) {
      this.api
        .post<void>('/auth/logout', { refresh_token: refreshToken })
        .pipe(catchError(() => EMPTY))
        .subscribe();
    }
    this.clearSession();
    this.router.navigate(['/auth/login']);
  }

  /**
   * Boot hydration: a token that survived a reload is verified against
   * /auth/me and the permission set is loaded, without blocking first paint.
   * A 401 here goes through the interceptor's refresh → logout path, so a
   * stale session ends up on the login page on its own.
   */
  hydrate(): void {
    if (!this.token()) return;
    this.fetchMe().subscribe({ error: () => undefined });
  }

  /** GET /auth/me → merge into `user`, then load permissions. */
  fetchMe(): Observable<User> {
    return this.api.get<Me>('/auth/me').pipe(
      tap((me) => {
        // Project /me down to the AuthUser shape login stores, so `slam_user`
        // has one shape whichever endpoint filled it.
        const user: User = {
          id: me.id,
          username: me.username,
          email: me.email,
          anggota_id: me.anggota_id,
          nama: me.nama,
          foto_url: me.anggota?.foto_url ?? me.foto_url ?? null,
          role: me.role,
          perm_version: me.perm_version,
        };
        this._user.set(user);
        this.persistUser(user);
      }),
      switchMap(() => this.perms.load()),
      map(() => this._user()!),
    );
  }

  /** Persists a token pair + user. Called by login and by every refresh. */
  setSession(res: AuthResponse): void {
    localStorage.setItem(TOKEN_KEY, res.access_token);
    localStorage.setItem(REFRESH_KEY, res.refresh_token);
    this.persistUser(res.user);
    this._user.set(res.user);
  }

  /** Forgets the session locally only; `logout()` is the user-facing version. */
  clearSession(): void {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_KEY);
    localStorage.removeItem(USER_KEY);
    this._user.set(null);
    this.perms.clear();
  }

  token(): string | null {
    return localStorage.getItem(TOKEN_KEY);
  }

  refreshToken(): string | null {
    return localStorage.getItem(REFRESH_KEY);
  }

  private persistUser(user: User): void {
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  }
}

function readStoredUser(): User | null {
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as User;
  } catch {
    // A corrupted entry would otherwise throw during service construction and
    // take the whole app down at bootstrap.
    localStorage.removeItem(USER_KEY);
    return null;
  }
}
