import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';

import { ToastService } from '../../core/services/toast.service';
import { PrestasiService } from './prestasi.service';
import { Prestasi, PrestasiQuery, AnggotaOpsi } from './prestasi.model';
import { Page, emptyPage } from '../../core/models/api.model';
import { HasPermissionDirective } from '../../shared/directives/has-permission.directive';

@Component({
  selector: 'app-prestasi-list',
  imports: [RouterLink, FormsModule, TranslatePipe, HasPermissionDirective],
  templateUrl: './prestasi-list.html',
  styleUrl: './prestasi-list.scss',
})
export class PrestasiList {
  private svc = inject(PrestasiService);
  private toast = inject(ToastService);

  readonly loading = signal(true);
  readonly page = signal<Page<Prestasi>>(emptyPage<Prestasi>());
  readonly q = signal('');
  readonly anggotaFilter = signal<number | null>(null);
  readonly tingkatFilter = signal<string>('');
  readonly currentPage = signal(1);
  readonly perPage = signal(20);

  readonly items = computed(() => this.page().items);
  readonly total = computed(() => this.page().total);
  readonly totalPages = computed(() => this.page().last_page);

  readonly anggotaList = signal<AnggotaOpsi[]>([]);

  constructor() {
    this.loadAnggota();
    this.load();
  }

  loadAnggota(): void {
    this.svc.anggotaTersedia().subscribe({
      next: (list) => this.anggotaList.set(list),
    });
  }

  load(): void {
    this.loading.set(true);
    const query: PrestasiQuery = {
      page: this.currentPage(),
      per_page: this.perPage(),
      q: this.q(),
    };
    const af = this.anggotaFilter();
    if (af) query.anggota_id = af;
    const tf = this.tingkatFilter();
    if (tf !== '') query.tingkat = tf;

    this.svc.list(query).subscribe({
      next: (p) => {
        this.page.set(p);
        this.loading.set(false);
      },
      error: () => {
        this.toast.error('PRESTASI.LOAD_ERROR');
        this.loading.set(false);
      },
    });
  }

  onSearch(): void {
    this.currentPage.set(1);
    this.load();
  }

  onAnggotaChange(val: string): void {
    this.anggotaFilter.set(val === '' ? null : +val);
    this.currentPage.set(1);
    this.load();
  }

  onTingkatChange(val: string): void {
    this.tingkatFilter.set(val);
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

  deleteItem(p: Prestasi): void {
    if (!confirm('Yakin ingin menghapus prestasi ini?')) return;
    this.svc.remove(p.id).subscribe({
      next: () => {
        this.toast.success('PRESTASI.DELETE_SUCCESS');
        this.load();
      },
      error: () => this.toast.error('PRESTASI.DELETE_ERROR'),
    });
  }

  anggotaDisplay(p: Prestasi): string {
    return p.anggota?.nama_lengkap ?? '—';
  }
}
