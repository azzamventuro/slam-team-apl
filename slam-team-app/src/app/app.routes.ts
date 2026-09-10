import { Routes } from '@angular/router';
import { authGuard } from './core/guards/auth.guard';
import { adminGuard } from './core/guards/admin.guard';
import { permissionGuard } from './core/guards/permission.guard';

export const routes: Routes = [
  {
    path: 'auth/login',
    loadComponent: () => import('./pages/auth/login/login').then((m) => m.Login),
  },
  {
    path: 'forbidden',
    loadComponent: () => import('./pages/forbidden/forbidden').then((m) => m.Forbidden),
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
      // ── Hak Akses (RBAC) — SA only ──
      {
        path: 'hak-akses',
        canActivate: [permissionGuard],
        data: { permission: 'hak_akses.read' },
        children: [
          {
            path: '',
            loadComponent: () => import('./pages/hak-akses/role-list').then((m) => m.RoleList),
          },
          {
            path: 'create',
            loadComponent: () => import('./pages/hak-akses/role-form').then((m) => m.RoleForm),
          },
          {
            path: ':id/edit',
            loadComponent: () => import('./pages/hak-akses/role-form').then((m) => m.RoleForm),
          },
          {
            path: ':id/matrix',
            loadComponent: () => import('./pages/hak-akses/role-matrix').then((m) => m.RoleMatrix),
          },
        ],
      },
      // ── Instansi (Master Data) ──
      {
        path: 'instansi',
        canActivate: [permissionGuard],
        data: { permission: 'instansi.read' },
        children: [
          {
            path: '',
            loadComponent: () => import('./pages/instansi/instansi-list').then((m) => m.InstansiList),
          },
          {
            path: 'create',
            loadComponent: () => import('./pages/instansi/instansi-form').then((m) => m.InstansiFormComponent),
          },
          {
            path: ':id/edit',
            loadComponent: () => import('./pages/instansi/instansi-form').then((m) => m.InstansiFormComponent),
          },
        ],
      },
    ],
  },
  { path: '**', redirectTo: '' },
];
