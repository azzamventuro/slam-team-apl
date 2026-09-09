import { Injectable, computed, inject, signal } from '@angular/core';
import { Observable, tap } from 'rxjs';

import { LoginResponse, User } from '../models/user.model';
import { ApiService } from './api.service';

const TOKEN_KEY = 'slam_token';
const USER_KEY = 'slam_user';

/**
 * AuthService owns authentication state: the JWT (persisted to localStorage)
 * and the current user (a signal, so zoneless change detection just works).
 * HTTP goes through ApiService, which unwraps the envelope — `login` therefore
 * emits the `data` payload, not the envelope.
 */
@Injectable({ providedIn: 'root' })
export class AuthService {
  private api = inject(ApiService);

  private readonly _user = signal<User | null>(readStoredUser());
  readonly user = this._user.asReadonly();
  readonly isLoggedIn = computed(() => this._user() !== null);

  login(email: string, password: string): Observable<LoginResponse> {
    return this.api
      .post<LoginResponse>('/auth/login', { email, password })
      .pipe(tap((data) => this.setSession(data.token, data.user)));
  }

  /** Persists a session. Also the seam for refresh-token flows (modul 04). */
  setSession(token: string, user: User): void {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USER_KEY, JSON.stringify(user));
    this._user.set(user);
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
