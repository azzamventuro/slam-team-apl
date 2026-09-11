import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';

import { ToastService } from '../../core/services/toast.service';
import { MedsosService } from './medsos.service';
import { Medsos, MedsosQuery, AnggotaOpsi } from './medsos.model';
import { Page, emptyPage } from '../../core/models/api.model';
import { HasPermissionDirective } from '../../shared/directives/has-permission.directive';

@Component({
  selector: 'app-medsos-list',
  imports: [RouterLink, FormsModule, TranslatePipe, HasPermissionDirective],
  templateUrl: './medsos-list.html',
  styleUrl: './medsos-list.scss',
})
export class MedsosList {
  private svc = inject(MedsosService);
  private toast = inject(ToastService);

  readonly loading = signal(true);
  readonly page = signal<Page<Medsos>>(emptyPage<Medsos>());
  readonly q = signal('');
  readonly jenisFilter = signal('');
  readonly anggotaFilter = signal<number | null>(null);
  readonly currentPage = signal(1);
  readonly perPage = signal(20);
  readonly anggotaList = signal<AnggotaOpsi[]>([]);

  readonly items = computed(() => this.page().items);
  readonly total = computed(() => this.page().total);
  readonly totalPages = computed(() => this.page().last_page);

  constructor() {
    this.load();
    this.loadAnggota();
  }

  loadAnggota(): void {
    this.svc.anggotaTersedia().subscribe({
      next: (list) => this.anggotaList.set(list),
    });
  }

  load(): void {
    this.loading.set(true);
    const query: MedsosQuery = {
      page: this.currentPage(),
      per_page: this.perPage(),
      q: this.q(),
    };
    const jf = this.jenisFilter();
    if (jf !== '') query.jenis_medsos = jf;
    const af = this.anggotaFilter();
    if (af) query.anggota_id = af;

    this.svc.list(query).subscribe({
      next: (p) => {
        this.page.set(p);
        this.loading.set(false);
      },
      error: () => {
        this.toast.error('MEDSOS.LOAD_ERROR');
        this.loading.set(false);
      },
    });
  }

  onSearch(): void {
    this.currentPage.set(1);
    this.load();
  }

  onJenisChange(val: string): void {
    this.jenisFilter.set(val);
    this.currentPage.set(1);
    this.load();
  }

  onAnggotaChange(val: string): void {
    this.anggotaFilter.set(val === '' ? null : +val);
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

  deleteItem(item: Medsos): void {
    if (!confirm('Yakin ingin menghapus medsos ini?')) return;
    this.svc.remove(item.id).subscribe({
      next: () => {
        this.toast.success('MEDSOS.DELETE_SUCCESS');
        this.load();
      },
      error: () => this.toast.error('MEDSOS.DELETE_ERROR'),
    });
  }

  anggotaDisplay(m: Medsos): string {
    const a = this.anggotaList().find((x) => x.id === m.anggota_id);
    return a?.nama_lengkap ?? String(m.anggota_id);
  }
}
