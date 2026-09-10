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
import { InstansiService } from './instansi.service';
import { InstansiForm } from '../../core/models/instansi.model';

@Component({
  selector: 'app-instansi-form',
  imports: [ReactiveFormsModule, RouterLink, TranslatePipe],
  templateUrl: './instansi-form.html',
  styleUrl: './instansi-form.scss',
})
export class InstansiFormComponent implements OnInit, OnDestroy {
  private fb = inject(FormBuilder);
  private svc = inject(InstansiService);
  private fileSvc = inject(FileService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private toast = inject(ToastService);

  readonly saving = signal(false);
  readonly isEdit = signal(false);

  instId: number | null = null;

  /** Object URLs for preview — must be revoked on destroy. */
  private logoUtamaUrl = signal<string | null>(null);
  private logoTambahanUrl = signal<string | null>(null);
  readonly logoUtamaPreview = this.logoUtamaUrl;
  readonly logoTambahanPreview = this.logoTambahanUrl;

  /** File IDs resolved from FileService.upload after picking a file. */
  private logoUtamaFileId: number | null = null;
  private logoTambahanFileId: number | null = null;

  form: FormGroup = this.fb.nonNullable.group({
    kode: ['', [Validators.required, Validators.maxLength(50)]],
    nama: ['', [Validators.required, Validators.maxLength(150)]],
    nama_club: ['', [Validators.maxLength(150)]],
    alamat: ['', [Validators.maxLength(500)]],
    no_telepon: ['', [Validators.maxLength(20), Validators.pattern(/^[0-9+\-() ]*$/)]],
    tanggal_bergabung: [''],
    status: [1],
  });

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.instId = +id;
      this.isEdit.set(true);
      this.form.get('kode')?.disable();
      this.svc.detail(this.instId).subscribe({
        next: (inst) => {
          this.form.patchValue({
            kode: inst.kode,
            nama: inst.nama,
            nama_club: inst.nama_club ?? '',
            alamat: inst.alamat ?? '',
            no_telepon: inst.no_telepon ?? '',
            tanggal_bergabung: inst.tanggal_bergabung ?? '',
            status: inst.status,
          });
          this.logoUtamaFileId = inst.logo_utama_file_id;
          this.logoTambahanFileId = inst.logo_tambahan_file_id;
          if (inst.logo_utama_uuid) {
            this.fileSvc.imageUrl(inst.logo_utama_uuid, 'medium').subscribe((url) => {
              this.logoUtamaUrl.set(url);
            });
          }
          if (inst.logo_tambahan_uuid) {
            this.fileSvc.imageUrl(inst.logo_tambahan_uuid, 'medium').subscribe((url) => {
              this.logoTambahanUrl.set(url);
            });
          }
        },
        error: () => {
          this.toast.error('INSTANSI.LOAD_ERROR');
          this.router.navigate(['/instansi']);
        },
      });
    }
  }

  ngOnDestroy(): void {
    if (this.logoUtamaUrl()) URL.revokeObjectURL(this.logoUtamaUrl()!);
    if (this.logoTambahanUrl()) URL.revokeObjectURL(this.logoTambahanUrl()!);
  }

  onLogoUtamaPicked(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    const file = input.files[0];
    this.fileSvc.upload('file', file, { kategori: 'logo', reff_type: 'instansi' }).subscribe({
      next: (res) => {
        this.logoUtamaFileId = res.id;
        if (this.logoUtamaUrl()) URL.revokeObjectURL(this.logoUtamaUrl()!);
        this.logoUtamaUrl.set(URL.createObjectURL(file));
      },
      error: () => this.toast.error('COMMON.SERVER_ERROR'),
    });
    input.value = '';
  }

  onLogoTambahanPicked(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    const file = input.files[0];
    this.fileSvc.upload('file', file, { kategori: 'logo', reff_type: 'instansi' }).subscribe({
      next: (res) => {
        this.logoTambahanFileId = res.id;
        if (this.logoTambahanUrl()) URL.revokeObjectURL(this.logoTambahanUrl()!);
        this.logoTambahanUrl.set(URL.createObjectURL(file));
      },
      error: () => this.toast.error('COMMON.SERVER_ERROR'),
    });
    input.value = '';
  }

  submit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.saving.set(true);
    const raw = this.form.getRawValue();
    const body: InstansiForm = {
      ...raw,
      logo_utama_file_id: this.logoUtamaFileId,
      logo_tambahan_file_id: this.logoTambahanFileId,
    };

    const req$ = this.isEdit()
      ? this.svc.update(this.instId!, body)
      : this.svc.create(body);

    req$.subscribe({
      next: () => {
        this.toast.success('COMMON.SAVED');
        this.router.navigate(['/instansi']);
      },
      error: (e) => {
        this.applyServerErrors(e?.errors);
        this.saving.set(false);
      },
    });
  }

  private applyServerErrors(errors: Record<string, string[]> | undefined): void {
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
