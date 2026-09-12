import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';

import { AuthResponse, User } from '../models/user.model';
import { AuthService } from './auth.service';
import { PermissionService } from './permission.service';

const user: User = {
  id: 1,
  username: 'azzz',
  email: 'azzz@slam.id',
  anggota_id: 12,
  nama: 'Achmad',
  role: { id: 2, nama: 'Admin', level: 10, is_super: false },
  perm_version: 7,
};

function pair(
  access: string,
  refresh: string,
): { success: true; message: string; data: AuthResponse } {
  return {
    success: true,
    message: 'OK',
    data: {
      access_token: access,
      refresh_token: refresh,
      token_type: 'Bearer',
      expires_in: 900,
      user,
    },
  };
}

const perms = {
  success: true,
  message: 'OK',
  data: { permissions: ['anggota.read'], perm_version: 7, is_super: false },
};

describe('AuthService', () => {
  let svc: AuthService;
  let http: HttpTestingController;
  let router: Router;

  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])],
    });
    svc = TestBed.inject(AuthService);
    http = TestBed.inject(HttpTestingController);
    router = TestBed.inject(Router);
    vi.spyOn(router, 'navigate').mockResolvedValue(true);
  });

  afterEach(() => http.verify());

  it('login posts {identifier, password}, stores the pair + user, then loads permissions', () => {
    let done: User | undefined;
    svc.login('azzz', 'rahasia123').subscribe((u) => (done = u));

    const req = http.expectOne('/api/v1/auth/login');
    expect(req.request.body).toEqual({ identifier: 'azzz', password: 'rahasia123' });
    req.flush(pair('acc-1', 'ref-1'));

    // Permissions are loaded BEFORE login emits — the caller may navigate immediately.
    expect(done).toBeUndefined();
    http.expectOne('/api/v1/auth/me/permissions').flush(perms);

    expect(done?.nama).toBe('Achmad');
    expect(localStorage.getItem('slam_token')).toBe('acc-1');
    expect(localStorage.getItem('slam_refresh')).toBe('ref-1');
    expect(JSON.parse(localStorage.getItem('slam_user')!).username).toBe('azzz');
    expect(svc.isLoggedIn()).toBe(true);
    expect(TestBed.inject(PermissionService).can('anggota.read')).toBe(true);
  });

  it('login works with an email identifier just the same (one field, server matches both)', () => {
    svc.login('azzz@slam.id', 'rahasia123').subscribe();
    const req = http.expectOne('/api/v1/auth/login');
    expect(req.request.body).toEqual({ identifier: 'azzz@slam.id', password: 'rahasia123' });
    req.flush(pair('acc-1', 'ref-1'));
    http.expectOne('/api/v1/auth/me/permissions').flush(perms);
    expect(svc.token()).toBe('acc-1');
  });

  it('refresh rotates the stored pair', () => {
    svc.setSession(pair('acc-1', 'ref-1').data);

    svc.refresh().subscribe();
    const req = http.expectOne('/api/v1/auth/refresh');
    expect(req.request.body).toEqual({ refresh_token: 'ref-1' });
    req.flush(pair('acc-2', 'ref-2'));

    expect(svc.token()).toBe('acc-2');
    expect(svc.refreshToken()).toBe('ref-2');
  });

  it('concurrent refresh calls share ONE request (a rotated token is single-use)', () => {
    svc.setSession(pair('acc-1', 'ref-1').data);
    let a = '';
    let b = '';
    svc.refresh().subscribe((r) => (a = r.access_token));
    svc.refresh().subscribe((r) => (b = r.access_token));

    http.expectOne('/api/v1/auth/refresh').flush(pair('acc-2', 'ref-2'));
    expect(a).toBe('acc-2');
    expect(b).toBe('acc-2');

    // After completion a new call is a new request, not a replay.
    svc.refresh().subscribe();
    http.expectOne('/api/v1/auth/refresh').flush(pair('acc-3', 'ref-3'));
    expect(svc.token()).toBe('acc-3');
  });

  it('a failed refresh clears the session and re-throws', () => {
    svc.setSession(pair('acc-1', 'ref-1').data);
    let failed = false;
    svc.refresh().subscribe({ error: () => (failed = true) });

    http
      .expectOne('/api/v1/auth/refresh')
      .flush(
        { success: false, message: 'refresh token tidak valid' },
        { status: 401, statusText: 'Unauthorized' },
      );

    expect(failed).toBe(true);
    expect(svc.token()).toBeNull();
    expect(svc.refreshToken()).toBeNull();
    expect(svc.isLoggedIn()).toBe(false);
  });

  it('refresh without a stored refresh token fails without a request', () => {
    let failed = false;
    svc.refresh().subscribe({ error: () => (failed = true) });
    expect(failed).toBe(true);
    http.expectNone('/api/v1/auth/refresh');
  });

  it('logout revokes the session with the Bearer still attached, then clears and redirects', () => {
    svc.setSession(pair('acc-1', 'ref-1').data);
    TestBed.inject(PermissionService).load().subscribe();
    http.expectOne('/api/v1/auth/me/permissions').flush(perms);

    svc.logout();

    const req = http.expectOne('/api/v1/auth/logout');
    expect(req.request.body).toEqual({ refresh_token: 'ref-1' });
    // Storage is already cleared by the time the response lands…
    expect(svc.token()).toBeNull();
    expect(svc.isLoggedIn()).toBe(false);
    expect(TestBed.inject(PermissionService).can('anggota.read')).toBe(false);
    expect(router.navigate).toHaveBeenCalledWith(['/auth/login']);
    // …and a failure from the API must not resurrect it.
    req.flush({ success: false, message: 'x' }, { status: 500, statusText: 'Internal' });
    expect(svc.token()).toBeNull();
  });

  it('logout without a refresh token skips the API call but still clears locally', () => {
    localStorage.setItem('slam_token', 'acc-1');
    svc.logout();
    http.expectNone('/api/v1/auth/logout');
    expect(svc.token()).toBeNull();
    expect(router.navigate).toHaveBeenCalledWith(['/auth/login']);
  });

  it('hydrate verifies a surviving token against /auth/me and loads permissions', () => {
    localStorage.setItem('slam_token', 'acc-1');
    svc.hydrate();

    http.expectOne('/api/v1/auth/me').flush({
      success: true,
      message: 'OK',
      data: {
        ...user,
        timezone: 'Asia/Jakarta',
        is_aktif: true,
        anggota: { id: 12, nama_lengkap: 'Achmad', foto_url: '/api/v1/files/u/low' },
      },
    });
    http.expectOne('/api/v1/auth/me/permissions').flush(perms);

    expect(svc.user()?.nama).toBe('Achmad');
    expect(svc.user()?.foto_url).toBe('/api/v1/files/u/low');
    expect(JSON.parse(localStorage.getItem('slam_user')!)).not.toHaveProperty('timezone');
    expect(TestBed.inject(PermissionService).ready()).toBe(true);
  });

  it('hydrate is a no-op without a token', () => {
    svc.hydrate();
    http.expectNone('/api/v1/auth/me');
  });
});
