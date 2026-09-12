import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';
import { provideTranslateService } from '@ngx-translate/core';
import { of, throwError } from 'rxjs';

import { AuthService } from '../../../core/services/auth.service';
import { Login } from './login';

describe('Login', () => {
  let fixture: ComponentFixture<Login>;
  let login: ReturnType<typeof vi.fn>;
  let router: Router;

  beforeEach(async () => {
    login = vi.fn();
    await TestBed.configureTestingModule({
      imports: [Login],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: AuthService, useValue: { login } },
      ],
    }).compileComponents();
    router = TestBed.inject(Router);
    vi.spyOn(router, 'navigate').mockResolvedValue(true);
    fixture = TestBed.createComponent(Login);
    fixture.detectChanges();
  });

  function fill(identifier: string, password: string): void {
    fixture.componentInstance.form.setValue({ identifier, password });
  }

  it('renders an identifier field (not email) and no email validator', () => {
    expect(fixture.nativeElement.querySelector('#login-identifier')).toBeTruthy();
    expect(fixture.nativeElement.querySelector('#login-email')).toBeNull();

    fill('azzz', 'rahasia123');
    expect(fixture.componentInstance.form.valid).toBe(true);
  });

  it('mirrors the DTO: password shorter than 8 never reaches the API', () => {
    fill('azzz', 'short');
    fixture.componentInstance.submit();
    expect(login).not.toHaveBeenCalled();
    expect(fixture.componentInstance.form.controls.password.hasError('minlength')).toBe(true);
  });

  it('submits the trimmed identifier and navigates to the dashboard on success', () => {
    login.mockReturnValue(of({}));
    fill('  azzz@slam.id ', 'rahasia123');
    fixture.componentInstance.submit();

    expect(login).toHaveBeenCalledWith('azzz@slam.id', 'rahasia123');
    expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
  });

  it('401 → generic invalid-credentials message, submit re-enabled', () => {
    login.mockReturnValue(throwError(() => new HttpErrorResponse({ status: 401 })));
    fill('azzz', 'rahasia123');
    fixture.componentInstance.submit();
    fixture.detectChanges();

    expect(fixture.componentInstance.error()).toBe('AUTH.LOGIN.ERROR_INVALID');
    expect(fixture.componentInstance.saving()).toBe(false);
    expect(fixture.nativeElement.querySelector('[role="alert"]')).toBeTruthy();
    expect(fixture.nativeElement.querySelector('button[type="submit"]').disabled).toBe(false);
  });

  it('403 → locked-account message', () => {
    login.mockReturnValue(throwError(() => new HttpErrorResponse({ status: 403 })));
    fill('azzz', 'rahasia123');
    fixture.componentInstance.submit();

    expect(fixture.componentInstance.error()).toBe('AUTH.LOGIN.ERROR_LOCKED');
    expect(router.navigate).not.toHaveBeenCalled();
  });

  it('disables submit while saving and ignores a second click', () => {
    login.mockReturnValue(of({}));
    fixture.componentInstance.saving.set(true);
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('button[type="submit"]').disabled).toBe(true);

    fill('azzz', 'rahasia123');
    fixture.componentInstance.submit();
    expect(login).not.toHaveBeenCalled();
  });
});
