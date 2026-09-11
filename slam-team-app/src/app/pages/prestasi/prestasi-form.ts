import { Component, inject, OnInit, OnDestroy, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import {
  FormBuilder,
  FormGroup,
  ReactiveFormsModule,
  Validators,
} from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';

import { ToastService } from '../../core/services/toast.service';
import { FileService } from '../../core/services/file.service';
import { PrestasiService } from './prestasi.service';
import { PrestasiForm as PrestasiFormBody, AnggotaOpsi } from './prestasi.model';

@Component({
  selector: 'app-prestasi-form',
  imports: [ReactiveFormsModule, RouterLink, TranslatePipe],
  templateUrl: './prestasi-form.html',
  styleUrl: './prestasi-form.scss',
})
export class PrestasiFormComponent implements OnInit, OnDestroy {
  private fb = inject(FormBuilder);
  private svc = inject(PrestasiService);
  private fileSvc = inject(FileService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private toast = inject(ToastService);

  readonly saving = signal(false);
  readonly isEdit = signal(false);
  readonly anggotaList = signal<AnggotaOpsi[]>([]);

  prestasiId: number | null = null;

  /** Object URLs for preview — must be revoked on destroy. */
  private flyerUrl = signal<string | null>(null);
  readonly flyerPreview = this.flyerUrl;
  private fotoUrl = signal<string | null>(null);
  readonly fotoPreview = this.fotoUrl;

  /** File IDs resolved from FileService.upload — sent to API. */
  private flyerFileId: number | null = null;
  private fotoFileId: number | null = null;

  form: FormGroup = this.fb.nonNullable.group({
    anggota_id: [null as number | null, [Validators.required]],
    judul_kompetisi: ['', [Validators.required, Validators.maxLength(200)]],
    kode: ['', [Validators.maxLength(50)]],
    peringkat: ['', [Validators.maxLength(50)]],
    tingkat: ['', [Validators.maxLength(50)]],
    tanggal_kompetisi: [''],
    alamat_kompetisi: [''],
    keterangan: [''],
  });

  ngOnInit(): void {
    this.svc.anggotaTersedia().subscribe({
      next: (list) => this.anggotaList.set(list),
    });

    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.prestasiId = +id;
      this.isEdit.set(true);
      this.svc.detail(this.prestasiId).subscribe({
        next: (p) => {
          this.form.patchValue({
            anggota_id: p.anggota_id,
            judul_kompetisi: p.judul_kompetisi,
            kode: p.kode ?? '',
            peringkat: p.peringkat ?? '',
            tingkat: p.tingkat ?? '',
            tanggal_kompetisi: p.tanggal_kompetisi ?? '',
            alamat_kompetisi: p.alamat_kompetisi ?? '',
            keterangan: p.keterangan ?? '',
          });

          // Load existing flyer preview.
          if (p.flyer?.uuid) {
            this.flyerFileId = p.flyer.file_id;
            this.fileSvc.imageUrl(p.flyer.uuid, 'original').subscribe({
              next: (url) => this.flyerUrl.set(url),
            });
          }

          // Load existing foto sampul preview.
          if (p.foto_sampul?.uuid) {
            this.fotoFileId = p.foto_sampul.file_id;
            this.fileSvc.imageUrl(p.foto_sampul.uuid, 'medium').subscribe({
              next: (url) => this.fotoUrl.set(url),
            });
          }
        },
        error: () => {
          this.toast.error('PRESTASI.LOAD_ERROR');
          this.router.navigate(['/prestasi']);
        },
      });
    }
  }

  ngOnDestroy(): void {
    if (this.flyerUrl()) URL.revokeObjectURL(this.flyerUrl()!);
    if (this.fotoUrl()) URL.revokeObjectURL(this.fotoUrl()!);
  }

  // ── File pickers ──

  onFlyerPicked(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    const file = input.files[0];
    this.fileSvc.upload('file', file, { kategori: 'flyer', reff_type: 'prestasi' }).subscribe({
      next: (res) => {
        this.flyerFileId = res.id;
        if (this.flyerUrl()) URL.revokeObjectURL(this.flyerUrl()!);
        this.flyerUrl.set(URL.createObjectURL(file));
      },
      error: () => this.toast.error('COMMON.SERVER_ERROR'),
    });
    input.value = '';
  }

  onFotoPicked(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    const file = input.files[0];
    this.fileSvc.upload('file', file, { kategori: 'foto', reff_type: 'prestasi' }).subscribe({
      next: (res) => {
        this.fotoFileId = res.id;
        if (this.fotoUrl()) URL.revokeObjectURL(this.fotoUrl()!);
        this.fotoUrl.set(URL.createObjectURL(file));
      },
      error: () => this.toast.error('COMMON.SERVER_ERROR'),
    });
    input.value = '';
  }

  // ── Submit ──

  submit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.saving.set(true);

    const val = this.form.getRawValue();
    const body: PrestasiFormBody = {
      anggota_id: val.anggota_id,
      judul_kompetisi: val.judul_kompetisi,
      kode: val.kode || null,
      peringkat: val.peringkat || null,
      tingkat: val.tingkat || null,
      tanggal_kompetisi: val.tanggal_kompetisi || null,
      alamat_kompetisi: val.alamat_kompetisi || null,
      keterangan: val.keterangan || null,
      flyer_file_id: this.flyerFileId,
      foto_sampul_file_id: this.fotoFileId,
    };

    const obs = this.isEdit()
      ? this.svc.update(this.prestasiId!, body)
      : this.svc.create(body);

    obs.subscribe({
      next: () => {
        this.toast.success(this.isEdit() ? 'PRESTASI.UPDATE_SUCCESS' : 'PRESTASI.CREATE_SUCCESS');
        this.router.navigate(['/prestasi']);
      },
      error: (err) => {
        this.saving.set(false);
        if (err?.errors && typeof err.errors === 'object') {
          this.applyServerErrors(err.errors as Record<string, string[]>);
        } else {
          this.toast.error('PRESTASI.SAVE_ERROR');
        }
      },
    });
  }

  private applyServerErrors(errors: Record<string, string[]>): void {
    for (const [field, msgs] of Object.entries(errors ?? {})) {
      this.form.get(field)?.setErrors({ server: msgs[0] });
    }
  }

  hasError(field: string, error: string): boolean {
    const ctrl = this.form.get(field);
    if (!ctrl) return false;
    return ctrl.hasError(error) && (ctrl.dirty || ctrl.touched);
  }
}
