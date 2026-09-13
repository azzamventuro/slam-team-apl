import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';

import { emptyPage, Page } from '../../core/models/api.model';
import { ToastService } from '../../core/services/toast.service';
import { HasPermissionDirective } from '../../shared/directives/has-permission.directive';
import { Lokasi, LokasiQuery } from './lokasi.model';
import { LokasiService } from './lokasi.service';

@Component({
  selector: 'app-lokasi-list',
  imports: [RouterLink, FormsModule, TranslatePipe, HasPermissionDirective],
  templateUrl: './lokasi-list.html',
  styleUrl: './lokasi-list.scss',
})
export class LokasiList {
  private svc = inject(LokasiService);
  private toast = inject(ToastService);

  readonly loading = signal(true);
  readonly page = signal<Page<Lokasi>>(emptyPage<Lokasi>());
  readonly q = signal('');
  /** Tri-state as a string so an untouched <select> can represent "all". */
  readonly aktifFilter = signal<string>('');
  readonly currentPage = signal(1);
  readonly perPage = signal(20);

  readonly items = computed(() => this.page().items);
  readonly total = computed(() => this.page().total);
  readonly totalPages = computed(() => this.page().last_page);

  constructor() {
    this.load();
  }

  load(): void {
    this.loading.set(true);
    const query: LokasiQuery = {
      page: this.currentPage(),
      per_page: this.perPage(),
      q: this.q(),
    };
    const af = this.aktifFilter();
    if (af !== '') query.is_aktif = af === 'true';

    this.svc.list(query).subscribe({
      next: (p) => {
        this.page.set(p);
        this.loading.set(false);
      },
      error: () => {
        this.toast.error('LOKASI.LOAD_ERROR');
        this.loading.set(false);
      },
    });
  }

  onSearch(): void {
    this.currentPage.set(1);
    this.load();
  }

  onAktifChange(val: string): void {
    this.aktifFilter.set(val);
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

  deleteItem(l: Lokasi): void {
    if (!confirm(`Yakin ingin menghapus lokasi "${l.nama}"?`)) return;
    this.svc.remove(l.id).subscribe({
      next: () => {
        this.toast.success('LOKASI.DELETE_SUCCESS');
        this.load();
      },
      // The service guards a lokasi still referenced by active jadwal with a
      // 409 — that is expected, not a server failure, so it gets its own
      // message instead of the generic one.
      error: (err: HttpErrorResponse) => {
        this.toast.error(err.status === 409 ? 'LOKASI.DELETE_IN_USE' : 'LOKASI.DELETE_ERROR');
      },
    });
  }

  statusClass(aktif: boolean): string {
    return aktif ? 'badge bg-success' : 'badge bg-secondary';
  }
}
