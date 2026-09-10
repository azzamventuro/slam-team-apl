import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';

import { AuthService } from '../../core/services/auth.service';
import { ToastService } from '../../core/services/toast.service';
import { SlamIconComponent } from '../../shared/components/slam-icon/slam-icon.component';
import { Role } from './hak-akses.model';
import { HakAksesService } from './hak-akses.service';

/**
 * RoleList displays all roles with level, is_super, and CRUD actions.
 * Anti-escalation rules (mirrored from backend, UX only):
 * - Cannot edit a role whose level <= your level (unless you are super).
 * - Cannot delete the last super admin.
 */
@Component({
  selector: 'app-role-list',
  imports: [RouterLink, TranslatePipe, SlamIconComponent],
  templateUrl: './role-list.html',
  styleUrl: './role-list.scss',
})
export class RoleList {
  private svc = inject(HakAksesService);
  private auth = inject(AuthService);
  private toast = inject(ToastService);

  readonly loading = signal(true);
  readonly roles = signal<Role[]>([]);

  readonly myLevel = computed(() => this.auth.user()?.role?.level ?? 999);
  readonly isSuper = computed(() => this.auth.user()?.role?.is_super ?? false);

  /** Count of super-admin roles — prevents deleting the last one. */
  readonly superAdminCount = computed(() =>
    this.roles().filter((r) => r.is_super).length,
  );

  constructor() {
    this.load();
  }

  load(): void {
    this.loading.set(true);
    this.svc.listRoles().subscribe({
      next: (roles) => {
        this.roles.set(roles);
        this.loading.set(false);
      },
      error: () => {
        this.toast.error('HAK_AKSES.LOAD_ERROR');
        this.loading.set(false);
      },
    });
  }

  /** Can the current user edit this role? */
  canEdit(role: Role): boolean {
    if (this.isSuper()) return true;
    return role.level > this.myLevel();
  }

  /** Can the current user delete this role? */
  canDelete(role: Role): boolean {
    if (!this.canEdit(role)) return false;
    // Prevent deleting the last super admin.
    if (role.is_super && this.superAdminCount() <= 1) return false;
    return true;
  }

  delete(role: Role): void {
    if (!confirm(`Hapus role "${role.nama}"?`)) return;
    this.svc.deleteRole(role.id).subscribe({
      next: () => {
        this.toast.success('COMMON.DELETED');
        this.load();
      },
      error: () => this.toast.error('COMMON.SERVER_ERROR'),
    });
  }

  /** Level badge color class. */
  levelClass(level: number): string {
    if (level === 0) return 'text-danger';
    if (level <= 10) return 'text-warning';
    return 'text-muted';
  }
}
