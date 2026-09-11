import { Component, computed, inject, OnInit, OnDestroy, signal } from '@angular/core';
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
import { InorgaService } from './inorga.service';
import { InorgaForm as InorgaFormBody } from './inorga.model';

@Component({
  selector: 'app-inorga-form',
  imports: [ReactiveFormsModule, RouterLink, TranslatePipe],
  templateUrl: './inorga-form.html',
  styleUrl: './inorga-form.scss',
})
export class InorgaFormComponent implements OnInit, OnDestroy {
  private fb = inject(FormBuilder);
  private svc = inject(InorgaService);
  private fileSvc = inject(FileService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private toast = inject(ToastService);

  readonly saving = signal(false);
  readonly isEdit = signal(false);

  inorgaId: number | null = null;

  /** Object URLs for preview — must be revoked on destroy. */
  private logoUrl = signal<string | null>(null);
  readonly logoPreview = this.logoUrl;
  private bannerUrl = signal<string | null>(null);
  readonly bannerPreview = this.bannerUrl;
  private skUrl = signal<string | null>(null);
  readonly skPreview = this.skUrl;

  /** File IDs resolved from FileService.upload — sent to API. */
  private _logoFileId: number | null = null;
  private _bannerFileId: number | null = null;
  private _skFileId = signal<number | null>(null);
  readonly skUploaded = computed(() => this._skFileId() !== null);

  form: FormGroup = this.fb.nonNullable.group({
    kode: ['', [Validators.maxLength(50)]],
    nama: ['', [Validators.required, Validators.maxLength(150)]],
    tanggal_mulai: ['', [Validators.required]],
    tanggal_selesai: [''],
    konten: [''],
  });

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.inorgaId = +id;
      this.isEdit.set(true);
      this.svc.detail(this.inorgaId).subscribe({
        next: (item) => {
          this.form.patchValue({
            kode: item.kode ?? '',
            nama: item.nama,
            tanggal_mulai: item.tanggal_mulai
              ? new Date(item.tanggal_mulai).toISOString().slice(0, 10)
              : '',
            tanggal_selesai: item.tanggal_selesai
              ? new Date(item.tanggal_selesai).toISOString().slice(0, 10)
              : '',
            konten: item.konten ?? '',
          });

          // Load existing logo preview.
          if (item.logo?.uuid) {
            this._logoFileId = item.logo.id;
            this.fileSvc.imageUrl(item.logo.uuid, 'medium').subscribe({
              next: (url) => this.logoUrl.set(url),
            });
          }

          // Load existing banner preview.
          if (item.banner?.uuid) {
            this._bannerFileId = item.banner.id;
            this.fileSvc.imageUrl(item.banner.uuid, 'medium').subscribe({
              next: (url) => this.bannerUrl.set(url),
            });
          }

          // Load existing SK preview.
          if (item.file_sk?.uuid) {
            this._skFileId.set(item.file_sk.id);
          }
        },
        error: () => {
          this.toast.error('INORGA.LOAD_ERROR');
          this.router.navigate(['/inorga']);
        },
      });
    }
  }

  ngOnDestroy(): void {
    if (this.logoUrl()) URL.revokeObjectURL(this.logoUrl()!);
    if (this.bannerUrl()) URL.revokeObjectURL(this.bannerUrl()!);
    if (this.skUrl()) URL.revokeObjectURL(this.skUrl()!);
  }

  // ── File pickers ──

  onLogoPicked(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    const file = input.files[0];
    this.fileSvc.upload('file', file, { kategori: 'logo', reff_type: 'mst_inorga' }).subscribe({
      next: (res) => {
        this._logoFileId = res.id;
        if (this.logoUrl()) URL.revokeObjectURL(this.logoUrl()!);
        this.logoUrl.set(URL.createObjectURL(file));
      },
      error: () => this.toast.error('COMMON.SERVER_ERROR'),
    });
    input.value = '';
  }

  onBannerPicked(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    const file = input.files[0];
    this.fileSvc.upload('file', file, { kategori: 'banner', reff_type: 'mst_inorga' }).subscribe({
      next: (res) => {
        this._bannerFileId = res.id;
        if (this.bannerUrl()) URL.revokeObjectURL(this.bannerUrl()!);
        this.bannerUrl.set(URL.createObjectURL(file));
      },
      error: () => this.toast.error('COMMON.SERVER_ERROR'),
    });
    input.value = '';
  }

  onSkPicked(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    const file = input.files[0];
    this.fileSvc.upload('file', file, { kategori: 'dokumen', reff_type: 'mst_inorga' }).subscribe({
      next: (res) => {
        this._skFileId.set(res.id);
        this.toast.success('INORGA.SK_UPLOAD_SUCCESS');
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
    const body: InorgaFormBody = {
      kode: val.kode || null,
      nama: val.nama,
      tanggal_mulai: val.tanggal_mulai || null,
      tanggal_selesai: val.tanggal_selesai || null,
      logo_file_id: this._logoFileId,
      banner_file_id: this._bannerFileId,
      file_sk_file_id: this._skFileId(),
      konten: val.konten || null,
    };

    const obs = this.isEdit()
      ? this.svc.update(this.inorgaId!, body)
      : this.svc.create(body);

    obs.subscribe({
      next: () => {
        this.toast.success(this.isEdit() ? 'INORGA.UPDATE_SUCCESS' : 'INORGA.CREATE_SUCCESS');
        this.router.navigate(['/inorga']);
      },
      error: (err) => {
        this.saving.set(false);
        if (err?.errors && typeof err.errors === 'object') {
          this.applyServerErrors(err.errors as Record<string, string[]>);
        } else {
          this.toast.error('INORGA.SAVE_ERROR');
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
