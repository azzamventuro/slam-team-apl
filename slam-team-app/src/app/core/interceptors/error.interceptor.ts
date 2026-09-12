import { HttpErrorResponse, HttpInterceptorFn, HttpRequest } from '@angular/common/http';
import { inject } from '@angular/core';
import { catchError, switchMap, throwError } from 'rxjs';

import { AuthService } from '../services/auth.service';
import { ToastService } from '../services/toast.service';

/**
 * One place for the cross-cutting HTTP failures:
 *   401 — the access token is stale: try ONE silent refresh and replay the
 *         request with the new token; if that fails, the session is gone —
 *         clear it and bounce to login.
 *   403 — RBAC denial from the Go handler: tell the user, keep them signed in.
 *   0   — no response at all (API down / CORS / offline).
 * Everything else is re-thrown for the caller to handle (e.g. a 422 mapped onto
 * form controls).
 *
 * The auth endpoints themselves are exempt from the 401/403 handling: a 401
 * from /auth/login is a wrong password (the form shows it), a 401 from
 * /auth/refresh is the end of the line (refreshing again would loop), and a
 * 403 from /auth/login is a locked account, not an RBAC denial.
 */
export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const toast = inject(ToastService);

  return next(req).pipe(
    catchError((err: HttpErrorResponse) => {
      const authEndpoint = isAuthEndpoint(req);

      switch (err.status) {
        case 401:
          if (authEndpoint) break;
          return recover(req, err);
        case 403:
          if (authEndpoint) break;
          toast.error('COMMON.FORBIDDEN');
          break;
        case 0:
          toast.error('COMMON.NETWORK_ERROR');
          break;
      }
      return throwError(() => err);
    }),
  );

  /**
   * Refresh, then replay. `next` here is the rest of the chain AFTER this
   * interceptor, so authInterceptor does not run again — the new Bearer has to
   * be set on the retried request by hand. Concurrent 401s share one refresh
   * (see AuthService.refresh), so a page with several lists recovers with a
   * single round trip.
   */
  function recover(original: HttpRequest<unknown>, cause: HttpErrorResponse) {
    if (!auth.refreshToken()) {
      expire();
      return throwError(() => cause);
    }
    return auth.refresh().pipe(
      switchMap((res) =>
        next(original.clone({ setHeaders: { Authorization: `Bearer ${res.access_token}` } })),
      ),
      catchError(() => {
        expire();
        // The caller sees the original 401, not the refresh's failure.
        return throwError(() => cause);
      }),
    );
  }

  function expire(): void {
    auth.logout();
    toast.warning('COMMON.SESSION_EXPIRED');
  }
};

/**
 * login / refresh / logout — the routes whose 401/403 mean something else.
 * /auth/me is deliberately NOT here: a stale token on boot hydration should
 * refresh like any other request.
 */
function isAuthEndpoint(req: HttpRequest<unknown>): boolean {
  return /\/auth\/(login|refresh|logout)(\?|$)/.test(req.url);
}
