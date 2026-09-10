import { Component, computed, inject, OnInit, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';

import { AuthService } from '../../core/services/auth.service';
import { ToastService } from '../../core/services/toast.service';
import { PermissionService } from '../../core/services/permission.service';
import {
  ModulCatalogue,
  Permission,
  Role,
  RolePermissionEntry,
} from './hak-akses.model';
import { HakAksesService } from './hak-akses.service';

/** All possible aksi values that appear as columns in the grid. */
const AksiColumns = ['create', 'read', 'update', 'delete', 'approve', 'assign', 'export', 'override', 'print', 'cabut', 'batal_sesi'] as const;

/** Cakupan options for the select dropdown on each checked cell. */
const CakupanOptions = ['semua', 'instansi_sendiri', 'milik_sendiri'] as const;

/**
 * RoleMatrix displays the role × modul permission grid.
 * Rows = modules, columns = aksi. Each checked cell also carries a cakupan select.
 * Anti-escalation (UX only, enforced server-side):
 * - Permissions the current actor does not hold are disabled (greyed out).
 * - Roles with level <= yours cannot be edited (redirect back).
 */
@Component({
  selector: 'app-role-matrix',
  imports: [RouterLink, TranslatePipe],
  templateUrl: './role-matrix.html',
  styleUrl: './role-matrix.scss',
})
export class RoleMatrix implements OnInit {
  private svc = inject(HakAksesService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private auth = inject(AuthService);
  private perms = inject(PermissionService);
  private toast = inject(ToastService);

  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly role = signal<Role | null>(null);
  readonly moduls = signal<ModulCatalogue[]>([]);
  readonly permissions = signal<Permission[]>([]);

  /** Currently active permission grants for this role, keyed by "modul_kode.aksi". */
  private readonly activeGrants = signal<Map<string, RolePermissionEntry>>(new Map());

  /** The grid aksi columns to display. */
  readonly aksiColumns = AksiColumns;
  readonly cakupanOptions = CakupanOptions;

  /** My level for anti-escalation check. */
  readonly myLevel = computed(() => this.auth.user()?.role?.level ?? 999);
  readonly isSuper = computed(() => this.auth.user()?.role?.is_super ?? false);

  /** Can the current user edit this role's permissions? */
  readonly canEditMatrix = computed(() => {
    const r = this.role();
    if (!r) return false;
    if (this.isSuper()) return true;
    return r.level > this.myLevel();
  });

  /** The roleId from the route. */
  private roleId = 0;

  ngOnInit(): void {
    this.roleId = +this.route.snapshot.paramMap.get('id')!;
    this.load();
  }

  load(): void {
    this.loading.set(true);
    // Fetch role detail, moduls, and permissions in parallel.
    let pending = 3;
    const done = () => {
      if (--pending === 0) {
        this.loading.set(false);
        // Redirect if user cannot edit this role.
        if (!this.canEditMatrix()) {
          this.toast.error('COMMON.FORBIDDEN');
          this.router.navigate(['/hak-akses']);
        }
      }
    };

    this.svc.getRole(this.roleId).subscribe({
      next: (role) => {
        this.role.set(role);
        // Build the active grants map from the role's permissions.
        const map = new Map<string, RolePermissionEntry>();
        for (const p of role.permissions) {
          map.set(`${p.modul_kode}.${p.aksi}`, p);
        }
        this.activeGrants.set(map);
        done();
      },
      error: () => {
        this.toast.error('HAK_AKSES.LOAD_ERROR');
        this.router.navigate(['/hak-akses']);
      },
    });

    this.svc.listModuls().subscribe({
      next: (m) => { this.moduls.set(m); done(); },
      error: done,
    });

    this.svc.listPermissions().subscribe({
      next: (p) => { this.permissions.set(p); done(); },
      error: done,
    });
  }

  /** Check if a permission cell is checked. */
  isChecked(modulKode: string, aksi: string): boolean {
    return this.activeGrants().has(`${modulKode}.${aksi}`);
  }

  /** Get the cakupan for a permission cell. */
  getCakupan(modulKode: string, aksi: string): string {
    return this.activeGrants().get(`${modulKode}.${aksi}`)?.cakupan ?? 'semua';
  }

  /** Check if the user has a specific permission (for disabling cells). */
  userHasPermission(modulKode: string, aksi: string): boolean {
    return this.perms.can(`${modulKode}.${aksi}`);
  }

  /** Toggle a permission cell on/off. */
  toggle(modulKode: string, aksi: string): void {
    if (!this.canEditMatrix()) return;
    const key = `${modulKode}.${aksi}`;
    const grants = new Map(this.activeGrants());
    if (grants.has(key)) {
      grants.delete(key);
    } else {
      // Find the permission_id.
      const perm = this.permissions().find(
        (p) => p.modul_kode === modulKode && p.aksi === aksi,
      );
      if (perm) {
        grants.set(key, {
          permission_id: perm.id,
          modul_kode: modulKode,
          aksi,
          cakupan: 'semua',
        });
      }
    }
    this.activeGrants.set(grants);
  }

  /** Update the cakupan for a permission cell. */
  setCakupan(modulKode: string, aksi: string, cakupan: string): void {
    const key = `${modulKode}.${aksi}`;
    const grants = new Map(this.activeGrants());
    const entry = grants.get(key);
    if (entry) {
      grants.set(key, { ...entry, cakupan });
      this.activeGrants.set(grants);
    }
  }

  /** Save all permissions for this role. */
  save(): void {
    if (!this.canEditMatrix()) return;
    this.saving.set(true);
    const permissions = Array.from(this.activeGrants().values()).map((g) => ({
      permission_id: g.permission_id,
      cakupan: g.cakupan,
    }));
    this.svc.setRolePermissions(this.roleId, { permissions }).subscribe({
      next: () => {
        this.toast.success('COMMON.SAVED');
        // Reload permissions so the sidebar and other UIs reflect changes.
        this.perms.load().subscribe();
        this.router.navigate(['/hak-akses']);
      },
      error: (e) => {
        this.toast.error(e?.message ?? 'COMMON.SERVER_ERROR');
        this.saving.set(false);
      },
    });
  }

  /** Get the role name for the heading. */
  roleName = computed(() => this.role()?.nama ?? '');
}
