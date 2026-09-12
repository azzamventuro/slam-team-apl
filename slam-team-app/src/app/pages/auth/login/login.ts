import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';

import { AuthService } from '../../../core/services/auth.service';

/** Mirrors the Go DTO: `password` is `binding:"required,min=8"`. */
export const PASSWORD_MIN = 8;

@Component({
  selector: 'app-login',
  imports: [ReactiveFormsModule, TranslatePipe],
  templateUrl: './login.html',
  styleUrl: './login.scss',
})
export class Login {
  private fb = inject(FormBuilder);
  private auth = inject(AuthService);
  private router = inject(Router);

  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly passwordMin = PASSWORD_MIN;

  // `identifier` is the username OR the email — one field, matched server-side
  // in a single query — so there is no email validator here on purpose.
  readonly form = this.fb.nonNullable.group({
    identifier: ['', [Validators.required]],
    password: ['', [Validators.required, Validators.minLength(PASSWORD_MIN)]],
  });

  /** Show a field's error only once the user has engaged with it. */
  invalid(name: 'identifier' | 'password'): boolean {
    const control = this.form.controls[name];
    return control.invalid && (control.touched || control.dirty);
  }

  submit(): void {
    if (this.saving()) return;
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.saving.set(true);
    this.error.set(null);
    const { identifier, password } = this.form.getRawValue();
    this.auth.login(identifier.trim(), password).subscribe({
      // Permissions are already loaded by the time login emits, so the
      // sidebar and *hasPermission are right on the dashboard's first paint.
      next: () => this.router.navigate(['/dashboard']),
      error: (err: unknown) => {
        this.error.set(messageFor(err));
        this.saving.set(false);
      },
    });
  }
}

/**
 * 401 is "wrong credentials" — the server deliberately does not say which
 * half, and neither do we. 403 is a locked account (too many failures).
 * Anything else is not the user's fault.
 */
function messageFor(err: unknown): string {
  const status = err instanceof HttpErrorResponse ? err.status : 0;
  switch (status) {
    case 401:
      return 'AUTH.LOGIN.ERROR_INVALID';
    case 403:
      return 'AUTH.LOGIN.ERROR_LOCKED';
    case 0:
      return 'COMMON.NETWORK_ERROR';
    default:
      return 'COMMON.SERVER_ERROR';
  }
}
