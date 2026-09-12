import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';

import { ToastService } from '../../core/services/toast.service';
import { InorgaService } from './inorga.service';
import { Inorga, InorgaQuery } from './inorga.model';
import { Page, emptyPage } from '../../core/models/api.model';
import { HasPermissionDirective } from '../../shared/directives/has-permission.directive';

@Component({
  selector: 'app-inorga-list',
  imports: [RouterLink, FormsModule, TranslatePipe, HasPermissionDirective],
  templateUrl: './inorga-list.html',
  styleUrl: './inorga-list.scss',
})
export class InorgaList {
  private svc = inject(InorgaService);
  private toast = inject(ToastService);

  readonly loading = signal(true);
  readonly page = signal<Page<Inorga>>(emptyPage<Inorga>());
  readonly q = signal('');
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
    const query: InorgaQuery = {
      page: this.currentPage(),
      per_page: this.perPage(),
      q: this.q(),
    };
    const af = this.aktifFilter();
    if (af !== '') query.aktif = af;

    this.svc.list(query).subscribe({
      next: (p) => {
        this.page.set(p);
        this.loading.set(false);
      },
      error: () => {
        this.toast.error('INORGA.LOAD_ERROR');
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

  deleteItem(item: Inorga): void {
    if (!confirm('Yakin ingin menghapus periode kepengurusan ini?')) return;
    this.svc.remove(item.id).subscribe({
      next: () => {
        this.toast.success('INORGA.DELETE_SUCCESS');
        this.load();
      },
      error: () => this.toast.error('INORGA.DELETE_ERROR'),
    });
  }

  formatPeriode(item: Inorga): string {
    const start = item.tanggal_mulai ? new Date(item.tanggal_mulai).toLocaleDateString('id-ID', { year: 'numeric', month: 'short' }) : '—';
    const end = item.tanggal_selesai ? new Date(item.tanggal_selesai).toLocaleDateString('id-ID', { year: 'numeric', month: 'short' }) : 'Berjalan';
    return `${start} – ${end}`;
  }
}
