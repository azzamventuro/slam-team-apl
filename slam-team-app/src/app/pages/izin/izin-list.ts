import { DatePipe } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';

import { emptyPage, Page } from '../../core/models/api.model';
import { PermissionService } from '../../core/services/permission.service';
import { ToastService } from '../../core/services/toast.service';
import { HasPermissionDirective } from '../../shared/directives/has-permission.directive';
import { Izin, IzinQuery, JENIS_IZIN, STATUS_IZIN, StatusIzin, hhmm } from './izin.model';
import { IzinJenisChip, IzinLampiranButton, IzinStatusChip } from './izin-parts';
import { IzinService } from './izin.service';

/**
 * Every request the caller may see. Ownership is not a filter here — the
 * API applies the caller's cakupan (User: own rows; Admin/Mod: all), so the
 * same page serves both. Approvers additionally get a link to the queue.
 */
@Component({
  selector: 'app-izin-list',
  imports: [
    RouterLink,
    FormsModule,
    TranslatePipe,
    DatePipe,
    HasPermissionDirective,
    IzinStatusChip,
    IzinJenisChip,
    IzinLampiranButton,
  ],
  templateUrl: './izin-list.html',
  styleUrl: './izin-list.scss',
})
export class IzinList {
  private svc = inject(IzinService);
  private toast = inject(ToastService);
  private perms = inject(PermissionService);

  readonly statusOpsi = STATUS_IZIN;
  readonly jenisOpsi = JENIS_IZIN;

  readonly loading = signal(true);
  readonly page = signal<Page<Izin>>(emptyPage<Izin>());
  readonly statusFilter = signal<StatusIzin | ''>('');
  readonly jenisFilter = signal<string>('');
  readonly currentPage = signal(1);
  readonly perPage = signal(20);

  readonly items = computed(() => this.page().items);
  readonly total = computed(() => this.page().total);
  readonly totalPages = computed(() => this.page().last_page);
  /** The requester column only says something when the caller can see other people's rows. */
  readonly showsAnggota = computed(() => this.perms.can('izin.approve'));

  constructor() {
    this.load();
  }

  load(): void {
    this.loading.set(true);
    const query: IzinQuery = {
      page: this.currentPage(),
      per_page: this.perPage(),
      sort: '-created_at',
    };
    if (this.statusFilter()) query.status = this.statusFilter();
    if (this.jenisFilter()) query.jenis = this.jenisFilter() as IzinQuery['jenis'];

    this.svc.list(query).subscribe({
      next: (p) => {
        this.page.set(p);
        this.loading.set(false);
      },
      error: () => {
        this.toast.error('IZIN.LOAD_ERROR');
        this.loading.set(false);
      },
    });
  }

  onStatusChange(val: StatusIzin | ''): void {
    this.statusFilter.set(val);
    this.currentPage.set(1);
    this.load();
  }

  onJenisChange(val: string): void {
    this.jenisFilter.set(val);
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

  hhmm(t: string | null | undefined): string {
    return hhmm(t);
  }
}
