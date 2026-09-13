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
            loadComponent: () =>
              import('./pages/instansi/instansi-list').then((m) => m.InstansiList),
          },
          {
            path: 'create',
            loadComponent: () =>
              import('./pages/instansi/instansi-form').then((m) => m.InstansiFormComponent),
          },
          {
            path: ':id/edit',
            loadComponent: () =>
              import('./pages/instansi/instansi-form').then((m) => m.InstansiFormComponent),
          },
        ],
      },
      // ── Anggota (Master Data) ──
      {
        path: 'anggota',
        canActivate: [permissionGuard],
        data: { permission: 'anggota.read' },
        children: [
          {
            path: '',
            loadComponent: () => import('./pages/anggota/anggota-list').then((m) => m.AnggotaList),
          },
          {
            path: 'create',
            loadComponent: () =>
              import('./pages/anggota/anggota-form').then((m) => m.AnggotaFormComponent),
          },
          {
            path: ':id/edit',
            loadComponent: () =>
              import('./pages/anggota/anggota-form').then((m) => m.AnggotaFormComponent),
          },
          {
            path: ':id',
            loadComponent: () =>
              import('./pages/anggota/anggota-detail').then((m) => m.AnggotaDetail),
          },
        ],
      },
      // ── Unit (Armament) ──
      {
        path: 'unit',
        canActivate: [permissionGuard],
        data: { permission: 'unit.read' },
        children: [
          {
            path: '',
            loadComponent: () => import('./pages/unit/unit-list').then((m) => m.UnitList),
          },
          {
            path: 'create',
            loadComponent: () => import('./pages/unit/unit-form').then((m) => m.UnitFormComponent),
          },
          {
            path: ':id/edit',
            loadComponent: () => import('./pages/unit/unit-form').then((m) => m.UnitFormComponent),
          },
        ],
      },
      // ── Lokasi (Operasional — training/activity locations, geofence + timezone) ──
      {
        path: 'lokasi',
        canActivate: [permissionGuard],
        data: { permission: 'lokasi.read' },
        children: [
          {
            path: '',
            loadComponent: () => import('./pages/lokasi/lokasi-list').then((m) => m.LokasiList),
          },
          {
            path: 'new',
            canActivate: [permissionGuard],
            data: { permission: 'lokasi.create' },
            loadComponent: () =>
              import('./pages/lokasi/lokasi-form').then((m) => m.LokasiFormComponent),
          },
          {
            path: ':id/edit',
            canActivate: [permissionGuard],
            data: { permission: 'lokasi.update' },
            loadComponent: () =>
              import('./pages/lokasi/lokasi-form').then((m) => m.LokasiFormComponent),
          },
        ],
      },
      // ── Prestasi (Achievements) ──
      {
        path: 'prestasi',
        canActivate: [permissionGuard],
        data: { permission: 'prestasi.read' },
        children: [
          {
            path: '',
            loadComponent: () =>
              import('./pages/prestasi/prestasi-list').then((m) => m.PrestasiList),
          },
          {
            path: 'create',
            loadComponent: () =>
              import('./pages/prestasi/prestasi-form').then((m) => m.PrestasiFormComponent),
          },
          {
            path: ':id/edit',
            loadComponent: () =>
              import('./pages/prestasi/prestasi-form').then((m) => m.PrestasiFormComponent),
          },
        ],
      },
      // ── Inorga (Kepengurusan) ──
      {
        path: 'inorga',
        canActivate: [permissionGuard],
        data: { permission: 'inorga.read' },
        children: [
          {
            path: '',
            loadComponent: () => import('./pages/inorga/inorga-list').then((m) => m.InorgaList),
          },
          {
            path: 'create',
            loadComponent: () =>
              import('./pages/inorga/inorga-form').then((m) => m.InorgaFormComponent),
          },
          {
            path: ':id/edit',
            loadComponent: () =>
              import('./pages/inorga/inorga-form').then((m) => m.InorgaFormComponent),
          },
        ],
      },
      // ── Medsos (Social Media Links) ──
      {
        path: 'medsos',
        canActivate: [permissionGuard],
        data: { permission: 'medsos.read' },
        children: [
          {
            path: '',
            loadComponent: () => import('./pages/medsos/medsos-list').then((m) => m.MedsosList),
          },
          {
            path: 'create',
            loadComponent: () =>
              import('./pages/medsos/medsos-form').then((m) => m.MedsosFormComponent),
          },
          {
            path: ':id/edit',
            loadComponent: () =>
              import('./pages/medsos/medsos-form').then((m) => m.MedsosFormComponent),
          },
        ],
      },
      // ── Dokumen (Documents) ──
      {
        path: 'dokumen',
        canActivate: [permissionGuard],
        data: { permission: 'dokumen.read' },
        children: [
          {
            path: '',
            loadComponent: () => import('./pages/dokumen/dokumen-list').then((m) => m.DokumenList),
          },
          {
            path: 'create',
            loadComponent: () =>
              import('./pages/dokumen/dokumen-form').then((m) => m.DokumenFormComponent),
          },
          {
            path: ':id/edit',
            loadComponent: () =>
              import('./pages/dokumen/dokumen-form').then((m) => m.DokumenFormComponent),
          },
        ],
      },
      // ── Profile Club (Singleton — one form, no list) ──
      {
        path: 'profile-club',
        canActivate: [permissionGuard],
        data: { permission: 'profile_club.read' },
        loadComponent: () =>
          import('./pages/profile-club/profile-club').then((m) => m.ProfileClubComponent),
      },
      // ── User Management — 3 surfaces sharing same components ──
      {
        path: 'admin',
        canActivate: [permissionGuard],
        data: { permission: 'admin.read', base: 'admin' },
        children: [
          {
            path: '',
            loadComponent: () =>
              import('./pages/user-management/user-list').then((m) => m.UserListComponent),
          },
          {
            path: 'create',
            loadComponent: () =>
              import('./pages/user-management/user-form').then((m) => m.UserFormComponent),
          },
          {
            path: ':id/edit',
            loadComponent: () =>
              import('./pages/user-management/user-form').then((m) => m.UserFormComponent),
          },
        ],
      },
      {
        path: 'moderator',
        canActivate: [permissionGuard],
        data: { permission: 'moderator.read', base: 'moderator' },
        children: [
          {
            path: '',
            loadComponent: () =>
              import('./pages/user-management/user-list').then((m) => m.UserListComponent),
          },
          {
            path: 'create',
            loadComponent: () =>
              import('./pages/user-management/user-form').then((m) => m.UserFormComponent),
          },
          {
            path: ':id/edit',
            loadComponent: () =>
              import('./pages/user-management/user-form').then((m) => m.UserFormComponent),
          },
        ],
      },
      {
        path: 'user',
        canActivate: [permissionGuard],
        data: { permission: 'user.read', base: 'user' },
        children: [
          {
            path: '',
            loadComponent: () =>
              import('./pages/user-management/user-list').then((m) => m.UserListComponent),
          },
          {
            path: 'create',
            loadComponent: () =>
              import('./pages/user-management/user-form').then((m) => m.UserFormComponent),
          },
          {
            path: ':id/edit',
            loadComponent: () =>
              import('./pages/user-management/user-form').then((m) => m.UserFormComponent),
          },
        ],
      },
    ],
  },
  { path: '**', redirectTo: '' },
];
