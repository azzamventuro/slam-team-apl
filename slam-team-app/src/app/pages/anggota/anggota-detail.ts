import { Component, inject, OnInit, OnDestroy, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';

import { ToastService } from '../../core/services/toast.service';
import { FileService } from '../../core/services/file.service';
import { AnggotaService } from './anggota.service';
import { Anggota } from '../../core/models/anggota.model';
import { HasPermissionDirective } from '../../shared/directives/has-permission.directive';

@Component({
  selector: 'app-anggota-detail',
  imports: [RouterLink, TranslatePipe, HasPermissionDirective],
  templateUrl: './anggota-detail.html',
  styleUrl: './anggota-detail.scss',
})
export class AnggotaDetail implements OnInit, OnDestroy {
  private svc = inject(AnggotaService);
  private fileSvc = inject(FileService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private toast = inject(ToastService);

  readonly loading = signal(true);
  readonly anggota = signal<Anggota | null>(null);

  readonly fotoProfilUrl = signal<string | null>(null);
  readonly fotoFormalUrl = signal<string | null>(null);
  readonly fileIdentitasUrl = signal<string | null>(null);

  private urls: string[] = [];

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (!id) {
      this.router.navigate(['/anggota']);
      return;
    }
    this.svc.detail(+id).subscribe({
      next: (a) => {
        this.anggota.set(a);
        this.loading.set(false);

        // Load images via blob (private files need Bearer header).
        if (a.foto_profil_uuid) {
          this.fileSvc.imageUrl(a.foto_profil_uuid, 'medium').subscribe({
            next: (url) => { this.fotoProfilUrl.set(url); this.urls.push(url); },
          });
        }
        if (a.foto_formal_uuid) {
          this.fileSvc.imageUrl(a.foto_formal_uuid, 'medium').subscribe({
            next: (url) => { this.fotoFormalUrl.set(url); this.urls.push(url); },
          });
        }
        if (a.file_identitas_uuid) {
          this.fileSvc.imageUrl(a.file_identitas_uuid, 'medium').subscribe({
            next: (url) => { this.fileIdentitasUrl.set(url); this.urls.push(url); },
          });
        }
      },
      error: () => {
        this.toast.error('ANGGOTA.LOAD_ERROR');
        this.router.navigate(['/anggota']);
      },
    });
  }

  ngOnDestroy(): void {
    for (const url of this.urls) URL.revokeObjectURL(url);
  }

  jenisLabel(j: string): string {
    switch (j) {
      case 'siswa': return 'Siswa';
      case 'dewasa': return 'Dewasa';
      case 'siswa_ke_dewasa': return 'Siswa → Dewasa';
      default: return j;
    }
  }

  jkLabel(v: number | null): string {
    if (v === 1) return 'Laki-laki';
    if (v === 2) return 'Perempuan';
    return '—';
  }

  jenisIdentitasLabel(v: number | null): string {
    switch (v) {
      case 1: return 'KTP';
      case 2: return 'SIM';
      case 3: return 'Paspor';
      case 4: return 'Lainnya';
      default: return '—';
    }
  }

  statusClass(s: string): string {
    return s === 'aktif' ? 'badge bg-success' : 'badge bg-secondary';
  }
}
