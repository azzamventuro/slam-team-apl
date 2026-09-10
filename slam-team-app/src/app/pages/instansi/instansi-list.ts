import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';

import { ToastService } from '../../core/services/toast.service';
import { FileService } from '../../core/services/file.service';
import { SlamIconComponent } from '../../shared/components/slam-icon/slam-icon.component';
import { InstansiService } from './instansi.service';
import { Instansi, InstansiQuery } from '../../core/models/instansi.model';
import { Page, emptyPage } from '../../core/models/api.model';
import { HasPermissionDirective } from '../../shared/directives/has-permission.directive';

@Component({
  selector: 'app-instansi-list',
  imports: [RouterLink, FormsModule, TranslatePipe, HasPermissionDirective, SlamIconComponent],
  templateUrl: './instansi-list.html',
  styleUrl: './instansi-list.scss',
})
export class InstansiList {
  private svc = inject(InstansiService);
  private fileSvc = inject(FileService);
  private toast = inject(ToastService);

  readonly loading = signal(true);
  readonly page = signal<Page<Instansi>>(emptyPage<Instansi>());
  readonly q = signal('');
  readonly statusFilter = signal<number | null>(null);
  readonly currentPage = signal(1);
  readonly perPage = signal(20);

  readonly items = computed(() => this.page().items);
  readonly total = computed(() => this.page().total);
  readonly totalPages = computed(() => this.page().last_page);

  /** Map of instansi.id → object URL for logo_utama. */
  private logoUrls = signal<Map<number, string>>(new Map());
  readonly logos = computed(() => this.logoUrls());

  constructor() {
    this.load();
  }

  load(): void {
    this.loading.set(true);
    const query: InstansiQuery = {
      page: this.currentPage(),
      per_page: this.perPage(),
      q: this.q(),
    };
    const s = this.statusFilter();
    if (s !== null) query.status = s;

    this.svc.list(query).subscribe({
      next: (p) => {
        this.page.set(p);
        this.loading.set(false);
      },
      error: () => {
        this.toast.error('INSTANSI.LOAD_ERROR');
        this.loading.set(false);
      },
    });
  }

  onSearch(): void {
    this.currentPage.set(1);
    this.load();
  }

  onStatusChange(val: string): void {
    this.statusFilter.set(val === '' ? null : +val);
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

  delete(inst: Instansi): void {
    const msg = `Hapus instansi "${inst.nama}"?`;
    if (!confirm(msg)) return;
    this.svc.remove(inst.id).subscribe({
      next: () => {
        this.toast.success('COMMON.DELETED');
        this.load();
      },
      error: () => this.toast.error('COMMON.SERVER_ERROR'),
    });
  }

  loadLogo(id: number, uuid: string | null): void {
    if (!uuid) return;
    if (this.logoUrls().has(id)) return;
    this.fileSvc.imageUrl(uuid, 'medium').subscribe((url) => {
      this.logoUrls.update((m) => {
        const next = new Map(m);
        next.set(id, url);
        return next;
      });
    });
  }

  statusLabel(s: number): string {
    return s === 1 ? 'Aktif' : 'Nonaktif';
  }

  statusClass(s: number): string {
    return s === 1 ? 'badge bg-success' : 'badge bg-secondary';
  }
}
