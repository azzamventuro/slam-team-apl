import { Component, computed, inject, OnInit, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';

import { ToastService } from '../../core/services/toast.service';
import { PermissionService } from '../../core/services/permission.service';
import { SlamIconComponent } from '../../shared/components/slam-icon/slam-icon.component';
import { UserService } from './user-management.service';
import { UserRow, UserListQuery } from './user-management.model';
import { Page, emptyPage } from '../../core/models/api.model';

@Component({
  selector: 'app-user-list',
  imports: [RouterLink, FormsModule, TranslatePipe, SlamIconComponent],
  templateUrl: './user-list.html',
  styleUrl: './user-list.scss',
})
export class UserListComponent implements OnInit {
  private svc = inject(UserService);
  private toast = inject(ToastService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private perms = inject(PermissionService);

  /** Surface base from route data: 'admin' | 'moderator' | 'user'. */
  readonly base = signal('');

  readonly loading = signal(true);
  readonly page = signal<Page<UserRow>>(emptyPage<UserRow>());
  readonly q = signal('');
  readonly statusFilter = signal<string>('');
  readonly currentPage = signal(1);
  readonly perPage = signal(20);

  readonly items = computed(() => this.page().items);
  readonly total = computed(() => this.page().total);
  readonly totalPages = computed(() => this.page().last_page);

  /** Title key based on surface. */
  readonly titleKey = computed(() => {
    const b = this.base();
    if (b === 'admin') return 'USERMGMT.ADMIN.TITLE';
    if (b === 'moderator') return 'USERMGMT.MODERATOR.TITLE';
    return 'USERMGMT.USER.TITLE';
  });

  /** Permission prefix based on surface. */
  readonly permPrefix = computed(() => `${this.base()}.`);

  /** Whether the current user can create in this surface. */
  readonly canCreate = computed(() => this.perms.can(`${this.base()}.create`));

  /** Whether the current user can update in this surface. */
  readonly canUpdate = computed(() => this.perms.can(`${this.base()}.update`));

  /** Whether the current user can delete in this surface. */
  readonly canDelete = computed(() => this.perms.can(`${this.base()}.delete`));

  ngOnInit(): void {
    // Read base from route data.
    const data = this.route.snapshot.data;
    this.base.set(data['base'] || 'admin');
    this.load();
  }

  load(): void {
    this.loading.set(true);
    const query: UserListQuery = {
      page: this.currentPage(),
      per_page: this.perPage(),
      q: this.q(),
    };
    const s = this.statusFilter();
    if (s !== '') query.is_aktif = s === '1';

    this.svc.list(this.base(), query).subscribe({
      next: (p) => {
        this.page.set(p);
        this.loading.set(false);
      },
      error: () => {
        this.toast.show('USERMGMT.LOAD_ERROR', 'danger');
        this.loading.set(false);
      },
    });
  }

  onSearch(): void {
    this.currentPage.set(1);
    this.load();
  }

  onStatusChange(val: string): void {
    this.statusFilter.set(val);
    this.currentPage.set(1);
    this.load();
  }

  prevPage(): void {
    if (this.currentPage() > 1) {
      this.currentPage.update((p) => p - 1);
      this.load();
    }
  }

  nextPage(): void {
    if (this.currentPage() < this.totalPages()) {
      this.currentPage.update((p) => p + 1);
      this.load();
    }
  }

  /** Map a user row's lock status to a display label. */
  lockLabel(user: UserRow): string {
    if (user.terkunci_sampai) return '🔒';
    return '';
  }

  deleteUser(user: UserRow): void {
    if (!confirm(`Hapus user "${user.username}"?`)) return;
    this.svc.remove(this.base(), user.id).subscribe({
      next: () => {
        this.toast.show('USERMGMT.DELETE_SUCCESS', 'success');
        this.load();
      },
      error: () => this.toast.show('USERMGMT.DELETE_ERROR', 'danger'),
    });
  }
}
