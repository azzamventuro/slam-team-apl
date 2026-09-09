import { Routes } from '@angular/router';
import { authGuard } from './core/guards/auth.guard';
import { adminGuard } from './core/guards/admin.guard';

export const routes: Routes = [
  {
    path: 'auth/login',
    loadComponent: () => import('./pages/auth/login/login').then((m) => m.Login),
  },
  {
    // Authenticated shell: vertical layout wraps all protected pages.
    path: '',
    loadComponent: () => import('./layouts/vertical/vertical').then((m) => m.VerticalLayout),
    canActivate: [authGuard],
    children: [
      { path: '', redirectTo: 'dashboard', pathMatch: 'full' },
      {
        path: 'dashboard',
        loadComponent: () => import('./pages/dashboard/dashboard').then((m) => m.Dashboard),
      },
      {
        // System settings — SA/Admin only (no pengaturan.* permission is
        // seeded; adminGuard mirrors the backend's role-based gate).
        path: 'pengaturan',
        canActivate: [adminGuard],
        loadComponent: () => import('./pages/pengaturan/pengaturan').then((m) => m.Pengaturan),
      },
    ],
  },
  { path: '**', redirectTo: '' },
];
