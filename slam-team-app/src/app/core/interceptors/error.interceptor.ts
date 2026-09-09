import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';

import { AuthService } from '../services/auth.service';
import { ToastService } from '../services/toast.service';

/**
 * One place for the cross-cutting HTTP failures:
 *   401 — the session is gone: clear it and bounce to login.
 *   403 — RBAC denial from the Go handler: tell the user, keep them signed in.
 *   0   — no response at all (API down / CORS / offline).
 * Everything else is re-thrown for the caller to handle (e.g. a 422 mapped onto
 * form controls).
 */
export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const router = inject(Router);
  const toast = inject(ToastService);

  return next(req).pipe(
    catchError((err: HttpErrorResponse) => {
      switch (err.status) {
        case 401:
          // Already on the login page? Then this is a failed sign-in attempt,
          // not an expired session — let the form show the message itself.
          if (!router.url.startsWith('/auth/login')) {
            auth.logout();
            toast.warning('COMMON.SESSION_EXPIRED');
            router.navigate(['/auth/login']);
          }
          break;
        case 403:
          toast.error('COMMON.FORBIDDEN');
          break;
        case 0:
          toast.error('COMMON.NETWORK_ERROR');
          break;
      }
      return throwError(() => err);
    }),
  );
};
