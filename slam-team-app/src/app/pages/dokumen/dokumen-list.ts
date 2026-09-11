import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';

import { ToastService } from '../../core/services/toast.service';
import { DokumenService } from './dokumen.service';
import { Dokumen, DokumenQuery, VALID_REFF_TYPES } from './dokumen.model';
import { Page, emptyPage } from '../../core/models/api.model';
import { HasPermissionDirective } from '../../shared/directives/has-permission.directive';

@Component({
  selector: 'app-dokumen-list',
  imports: [RouterLink, FormsModule, TranslatePipe, HasPermissionDirective],
  templateUrl: './dokumen-list.html',
  styleUrl: './dokumen-list.scss',
})
export class DokumenList {
  private svc = inject(DokumenService);
  private toast = inject(ToastService);

  readonly loading = signal(true);
  readonly page = signal<Page<Dokumen>>(emptyPage<Dokumen>());
  readonly q = signal('');
  readonly reffTypeFilter = signal('');
  readonly tipeFilter = signal('');
  readonly currentPage = signal(1);
  readonly perPage = signal(20);
  readonly reffTypes = VALID_REFF_TYPES;

  readonly items = computed(() => this.page().items);
  readonly total = computed(() => this.page().total);
  readonly totalPages = computed(() => this.page().last_page);

  constructor() {
    this.load();
  }

  load(): void {
    this.loading.set(true);
    const query: DokumenQuery = {
      page: this.currentPage(),
      per_page: this.perPage(),
      q: this.q(),
    };
    const rt = this.reffTypeFilter();
    if (rt !== '') query.reff_type = rt;
    const tp = this.tipeFilter();
    if (tp !== '') query.tipe = tp;

    this.svc.list(query).subscribe({
      next: (p) => {
        this.page.set(p);
        this.loading.set(false);
      },
      error: () => {
        this.toast.error('DOKUMEN.LOAD_ERROR');
        this.loading.set(false);
      },
    });
  }

  onSearch(): void {
    this.currentPage.set(1);
    this.load();
  }

  onReffTypeChange(val: string): void {
    this.reffTypeFilter.set(val);
    this.currentPage.set(1);
    this.load();
  }

  onTipeChange(val: string): void {
    this.tipeFilter.set(val);
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

  deleteItem(item: Dokumen): void {
    if (!confirm('Yakin ingin menghapus dokumen ini?')) return;
    this.svc.remove(item.id).subscribe({
      next: () => {
        this.toast.success('DOKUMEN.DELETE_SUCCESS');
        this.load();
      },
      error: () => this.toast.error('DOKUMEN.DELETE_ERROR'),
    });
  }

  downloadFile(item: Dokumen): void {
    if (!item.file?.uuid) return;
    this.svc.downloadBlob(item.file.uuid).subscribe({
      next: (blob) => {
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = item.file!.nama_asli;
        a.click();
        URL.revokeObjectURL(url);
      },
      error: () => this.toast.error('DOKUMEN.DOWNLOAD_ERROR'),
    });
  }

  /** Format file size for display. */
  formatSize(bytes: number): string {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / 1048576).toFixed(1) + ' MB';
  }
}
