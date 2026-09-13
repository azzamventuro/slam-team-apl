import { SlicePipe } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';

import { emptyPage, Page } from '../../core/models/api.model';
import { ToastService } from '../../core/services/toast.service';
import { SlamIconComponent } from '../../shared/components/slam-icon/slam-icon.component';
import { HasPermissionDirective } from '../../shared/directives/has-permission.directive';
import { Izin, IzinQuery, hhmm } from './izin.model';
import { IzinJenisChip, IzinLampiranButton } from './izin-parts';
import { IzinService } from './izin.service';

/** Maximum length the API accepts for catatan_peninjau (dto.TolakReq). */
const MAX_CATATAN = 2000;

/**
 * The approver's queue: every `menunggu` request, oldest first, each with
 * Setujui / Tolak. Rejecting opens an inline note box because the API
 * requires `catatan_peninjau` — the requester reads it in their notification.
 *
 * The buttons are behind `izin.approve`; the route is too. The Go handler
 * enforces it regardless (403 for User).
 */
@Component({
  selector: 'app-izin-approval',
  imports: [
    RouterLink,
    FormsModule,
    TranslatePipe,
    SlicePipe,
    HasPermissionDirective,
    SlamIconComponent,
    IzinJenisChip,
    IzinLampiranButton,
  ],
  templateUrl: './izin-approval.html',
  styleUrl: './izin-approval.scss',
})
export class IzinApproval {
  private svc = inject(IzinService);
  private toast = inject(ToastService);

  readonly maxCatatan = MAX_CATATAN;
  readonly loading = signal(true);
  readonly page = signal<Page<Izin>>(emptyPage<Izin>());
  readonly currentPage = signal(1);
  readonly perPage = signal(20);

  /** id of the row whose reject note is open; one at a time keeps the queue readable. */
  readonly rejecting = signal<number | null>(null);
  readonly catatan = signal('');
  /** id of the row with a request in flight — its buttons are disabled, the rest stay usable. */
  readonly busyId = signal<number | null>(null);

  readonly items = computed(() => this.page().items);
  readonly total = computed(() => this.page().total);
  readonly totalPages = computed(() => this.page().last_page);
  readonly catatanValid = computed(() => {
    const n = this.catatan().trim().length;
    return n > 0 && n <= MAX_CATATAN;
  });

  constructor() {
    this.load();
  }

  load(): void {
    this.loading.set(true);
    const query: IzinQuery = {
      page: this.currentPage(),
      per_page: this.perPage(),
      status: 'menunggu',
      sort: 'created_at',
    };
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

  approve(z: Izin): void {
    this.busyId.set(z.id);
    this.svc.approve(z.id).subscribe({
      next: () => {
        this.busyId.set(null);
        this.toast.success('IZIN.APPROVED');
        this.removeRow(z.id);
      },
      error: (err: HttpErrorResponse) => this.fail(err),
    });
  }

  openReject(z: Izin): void {
    this.rejecting.set(z.id);
    this.catatan.set('');
  }

  cancelReject(): void {
    this.rejecting.set(null);
    this.catatan.set('');
  }

  confirmReject(z: Izin): void {
    if (!this.catatanValid()) return;
    this.busyId.set(z.id);
    this.svc.tolak(z.id, { catatan_peninjau: this.catatan().trim() }).subscribe({
      next: () => {
        this.busyId.set(null);
        this.cancelReject();
        this.toast.success('IZIN.REJECTED');
        this.removeRow(z.id);
      },
      error: (err: HttpErrorResponse) => this.fail(err),
    });
  }

  /**
   * Drop the processed row locally; refetch only when the page would be left
   * empty so the next batch slides in.
   */
  private removeRow(id: number): void {
    const p = this.page();
    const items = p.items.filter((z) => z.id !== id);
    const remaining = Math.max(0, p.total - 1);
    if (!items.length && remaining > 0) {
      // This page emptied but others remain: later rows shift into this
      // page's slot, unless this was the last page, in which case step back.
      if (this.currentPage() >= p.last_page) this.currentPage.update((n) => Math.max(1, n - 1));
      this.load();
      return;
    }
    this.page.set({ ...p, items, total: remaining });
  }

  /** 409 = already processed by someone else — reload so the queue tells the truth. */
  private fail(err: HttpErrorResponse): void {
    this.busyId.set(null);
    if (err.status === 409) {
      this.toast.warning('IZIN.ALREADY_PROCESSED');
      this.load();
    } else if (err.status === 403) {
      this.toast.error('COMMON.FORBIDDEN');
    } else {
      this.toast.error('IZIN.PROCESS_ERROR');
    }
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
