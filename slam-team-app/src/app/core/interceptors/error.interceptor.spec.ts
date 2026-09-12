import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';

import { AuthResponse } from '../models/user.model';
import { AuthService } from '../services/auth.service';
import { ToastService } from '../services/toast.service';
import { authInterceptor } from './auth.interceptor';
import { errorInterceptor } from './error.interceptor';

const session: AuthResponse = {
  access_token: 'acc-old',
  refresh_token: 'ref-old',
  token_type: 'Bearer',
  expires_in: 900,
  user: {
    id: 1,
    username: 'azzz',
    email: 'azzz@slam.id',
    anggota_id: null,
    nama: 'Achmad',
    role: { id: 2, level: 10, is_super: false },
    perm_version: 1,
  },
};

const rotated = {
  success: true,
  message: 'OK',
  data: { ...session, access_token: 'acc-new', refresh_token: 'ref-new' },
};

const unauthorized = { status: 401, statusText: 'Unauthorized' };

describe('errorInterceptor — silent refresh on 401', () => {
  let http: HttpClient;
  let ctrl: HttpTestingController;
  let auth: AuthService;
  let toast: ToastService;
  let router: Router;

  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([authInterceptor, errorInterceptor])),
        provideHttpClientTesting(),
        provideRouter([]),
      ],
    });
    http = TestBed.inject(HttpClient);
    ctrl = TestBed.inject(HttpTestingController);
    auth = TestBed.inject(AuthService);
    toast = TestBed.inject(ToastService);
    router = TestBed.inject(Router);
    vi.spyOn(router, 'navigate').mockResolvedValue(true);
    auth.setSession(session);
  });

  afterEach(() => ctrl.verify());

  it('refreshes ONCE and replays the request with the new Bearer', () => {
    let body: unknown;
    http.get('/api/v1/anggota').subscribe((b) => (body = b));

    const first = ctrl.expectOne('/api/v1/anggota');
    expect(first.request.headers.get('Authorization')).toBe('Bearer acc-old');
    first.flush({ success: false, message: 'token expired' }, unauthorized);

    const refresh = ctrl.expectOne('/api/v1/auth/refresh');
    expect(refresh.request.body).toEqual({ refresh_token: 'ref-old' });
    refresh.flush(rotated);

    const retry = ctrl.expectOne('/api/v1/anggota');
    expect(retry.request.headers.get('Authorization')).toBe('Bearer acc-new');
    retry.flush({ success: true, message: 'OK', data: { items: [] } });

    expect(body).toEqual({ success: true, message: 'OK', data: { items: [] } });
    expect(auth.token()).toBe('acc-new');
    expect(router.navigate).not.toHaveBeenCalled();
  });

  it('two requests failing together trigger a single refresh and both recover', () => {
    http.get('/api/v1/anggota').subscribe();
    http.get('/api/v1/unit').subscribe();

    ctrl.expectOne('/api/v1/anggota').flush(null, unauthorized);
    ctrl.expectOne('/api/v1/unit').flush(null, unauthorized);

    ctrl.expectOne('/api/v1/auth/refresh').flush(rotated);

    expect(ctrl.expectOne('/api/v1/anggota').request.headers.get('Authorization')).toBe(
      'Bearer acc-new',
    );
    expect(ctrl.expectOne('/api/v1/unit').request.headers.get('Authorization')).toBe(
      'Bearer acc-new',
    );
  });

  it('a failed refresh logs out, redirects, and surfaces the ORIGINAL 401 to the caller', () => {
    let status = 0;
    http.get('/api/v1/anggota').subscribe({ error: (e) => (status = e.status) });

    ctrl.expectOne('/api/v1/anggota').flush(null, unauthorized);
    ctrl
      .expectOne('/api/v1/auth/refresh')
      .flush({ success: false, message: 'sesi dicabut' }, unauthorized);

    // No second refresh, no retry — the session is over.
    ctrl.expectNone('/api/v1/auth/refresh');
    ctrl.expectNone('/api/v1/anggota');
    expect(status).toBe(401);
    expect(auth.token()).toBeNull();
    expect(auth.isLoggedIn()).toBe(false);
    expect(router.navigate).toHaveBeenCalledWith(['/auth/login']);
    expect(toast.toasts().some((t) => t.text === 'COMMON.SESSION_EXPIRED')).toBe(true);
  });

  it('a 401 with no refresh token stored logs out straight away', () => {
    localStorage.removeItem('slam_refresh');
    http.get('/api/v1/anggota').subscribe({ error: () => undefined });
    ctrl.expectOne('/api/v1/anggota').flush(null, unauthorized);

    ctrl.expectNone('/api/v1/auth/refresh');
    expect(router.navigate).toHaveBeenCalledWith(['/auth/login']);
  });

  it('a 401 from /auth/login is a wrong password: passed through, no refresh, no logout', () => {
    let status = 0;
    http
      .post('/api/v1/auth/login', { identifier: 'x', password: 'y' })
      .subscribe({ error: (e) => (status = e.status) });
    ctrl.expectOne('/api/v1/auth/login').flush(null, unauthorized);

    ctrl.expectNone('/api/v1/auth/refresh');
    expect(status).toBe(401);
    expect(router.navigate).not.toHaveBeenCalled();
    expect(toast.toasts()).toEqual([]);
  });

  it('a 403 from /auth/login is a locked account: no FORBIDDEN toast', () => {
    http.post('/api/v1/auth/login', {}).subscribe({ error: () => undefined });
    ctrl.expectOne('/api/v1/auth/login').flush(null, { status: 403, statusText: 'Forbidden' });
    expect(toast.toasts()).toEqual([]);
  });

  it('a 403 elsewhere is an RBAC denial: toast, stay signed in', () => {
    http.delete('/api/v1/anggota/1').subscribe({ error: () => undefined });
    ctrl.expectOne('/api/v1/anggota/1').flush(null, { status: 403, statusText: 'Forbidden' });

    expect(toast.toasts().some((t) => t.text === 'COMMON.FORBIDDEN')).toBe(true);
    expect(auth.isLoggedIn()).toBe(true);
    expect(router.navigate).not.toHaveBeenCalled();
  });
});
