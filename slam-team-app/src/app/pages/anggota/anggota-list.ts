import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';

import { ToastService } from '../../core/services/toast.service';
import { FileService } from '../../core/services/file.service';
import { SlamIconComponent } from '../../shared/components/slam-icon/slam-icon.component';
import { AnggotaService } from './anggota.service';
import { Anggota, AnggotaQuery } from '../../core/models/anggota.model';
import { Page, emptyPage } from '../../core/models/api.model';
import { HasPermissionDirective } from '../../shared/directives/has-permission.directive';

@Component({
  selector: 'app-anggota-list',
  imports: [RouterLink, FormsModule, TranslatePipe, HasPermissionDirective, SlamIconComponent],
  templateUrl: './anggota-list.html',
  styleUrl: './anggota-list.scss',
})
export class AnggotaList {
  private svc = inject(AnggotaService);
  private toast = inject(ToastService);

  readonly loading = signal(true);
  readonly page = signal<Page<Anggota>>(emptyPage<Anggota>());
  readonly q = signal('');
  readonly instansiFilter = signal<number | null>(null);
  readonly jenisFilter = signal<string>('');
  readonly statusFilter = signal<string>('');
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
    const query: AnggotaQuery = {
      page: this.currentPage(),
      per_page: this.perPage(),
      q: this.q(),
    };
    const inst = this.instansiFilter();
    if (inst !== null) query.instansi_id = inst;
    const jenis = this.jenisFilter();
    if (jenis) query.jenis = jenis;
    const status = this.statusFilter();
    if (status) query.status = status;

    this.svc.list(query).subscribe({
      next: (p) => {
        this.page.set(p);
        this.loading.set(false);
      },
      error: () => {
        this.toast.error('ANGGOTA.LOAD_ERROR');
        this.loading.set(false);
      },
    });
  }

  onSearch(): void {
    this.currentPage.set(1);
    this.load();
  }

  onInstansiChange(val: string): void {
    this.instansiFilter.set(val === '' ? null : +val);
    this.currentPage.set(1);
    this.load();
  }

  onJenisChange(val: string): void {
    this.jenisFilter.set(val);
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

  deleteItem(a: Anggota): void {
    if (!confirm('Yakin ingin menghapus anggota ini?')) return;
    this.svc.remove(a.id).subscribe({
      next: () => {
        this.toast.success('ANGGOTA.DELETE_SUCCESS');
        this.load();
      },
      error: () => this.toast.error('ANGGOTA.DELETE_ERROR'),
    });
  }

  /** Label for no_induk — shows NRA or dash if null. */
  noInduk(a: Anggota): string {
    return a.no_induk ?? '—';
  }

  /** CSS class for status badge. */
  statusClass(s: string): string {
    return s === 'aktif' ? 'badge bg-success' : 'badge bg-secondary';
  }

  /** Human-readable jenis_anggota label. */
  jenisLabel(j: string): string {
    switch (j) {
      case 'siswa': return 'Siswa';
      case 'dewasa': return 'Dewasa';
      case 'siswa_ke_dewasa': return 'Siswa → Dewasa';
      default: return j;
    }
  }
}
