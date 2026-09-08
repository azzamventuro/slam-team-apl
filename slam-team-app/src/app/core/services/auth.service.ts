import { Injectable, computed, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';

import { environment } from '../../../environments/environment';
import { ApiResponse, LoginResponse, User } from '../models/user.model';

const TOKEN_KEY = 'slam_token';
const USER_KEY = 'slam_user';

/**
 * AuthService owns authentication state: the JWT (persisted to localStorage)
 * and the current user (a signal, so zoneless change detection just works).
 */
@Injectable({ providedIn: 'root' })
export class AuthService {
  private http = inject(HttpClient);

  private readonly _user = signal<User | null>(readStoredUser());
  readonly user = this._user.asReadonly();
  readonly isLoggedIn = computed(() => this._user() !== null);

  login(email: string, password: string): Observable<ApiResponse<LoginResponse>> {
    return this.http
      .post<ApiResponse<LoginResponse>>(`${environment.apiURL}/auth/login`, { email, password })
      .pipe(
        tap((res) => {
          if (res.success && res.data) {
            localStorage.setItem(TOKEN_KEY, res.data.token);
            localStorage.setItem(USER_KEY, JSON.stringify(res.data.user));
            this._user.set(res.data.user);
          }
        }),
      );
  }

  logout(): void {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    this._user.set(null);
  }

  token(): string | null {
    return localStorage.getItem(TOKEN_KEY);
  }
}

function readStoredUser(): User | null {
  const raw = localStorage.getItem(USER_KEY);
  return raw ? (JSON.parse(raw) as User) : null;
}
