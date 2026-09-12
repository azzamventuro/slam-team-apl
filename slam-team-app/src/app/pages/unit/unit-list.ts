import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';

import { ToastService } from '../../core/services/toast.service';
import { UnitService } from './unit.service';
import { Unit, UnitQuery, AnggotaOpsi } from './unit.model';
import { Page, emptyPage } from '../../core/models/api.model';
import { HasPermissionDirective } from '../../shared/directives/has-permission.directive';

@Component({
  selector: 'app-unit-list',
  imports: [RouterLink, FormsModule, TranslatePipe, HasPermissionDirective],
  templateUrl: './unit-list.html',
  styleUrl: './unit-list.scss',
})
export class UnitList {
  private svc = inject(UnitService);
  private toast = inject(ToastService);

  readonly loading = signal(true);
  readonly page = signal<Page<Unit>>(emptyPage<Unit>());
  readonly q = signal('');
  readonly anggotaFilter = signal<number | null>(null);
  readonly disetujuiFilter = signal<string>('');
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
    const query: UnitQuery = {
      page: this.currentPage(),
      per_page: this.perPage(),
      q: this.q(),
    };
    const af = this.anggotaFilter();
    if (af) query.anggota_id = af;
    const df = this.disetujuiFilter();
    if (df !== '') query.disetujui = df === 'true';

    this.svc.list(query).subscribe({
      next: (p) => {
        this.page.set(p);
        this.loading.set(false);
      },
      error: () => {
        this.toast.error('UNIT.LOAD_ERROR');
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

  onDisetujuiChange(val: string): void {
    this.disetujuiFilter.set(val);
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

  deleteItem(u: Unit): void {
    if (!confirm('Yakin ingin menghapus unit ini?')) return;
    this.svc.remove(u.id).subscribe({
      next: () => {
        this.toast.success('UNIT.DELETE_SUCCESS');
        this.load();
      },
      error: () => this.toast.error('UNIT.DELETE_ERROR'),
    });
  }

  /** Display anggota name or dash. */
  anggotaDisplay(u: Unit): string {
    return u.anggota_nama ?? '—';
  }

  /** Badge class for approval status. */
  disetujuiClass(d: boolean): string {
    return d ? 'badge bg-success' : 'badge bg-secondary';
  }
}
