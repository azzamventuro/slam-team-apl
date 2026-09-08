import { Routes } from '@angular/router';
import { authGuard } from './core/guards/auth.guard';

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
    ],
  },
  { path: '**', redirectTo: '' },
];
