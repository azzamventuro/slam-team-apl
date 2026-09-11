import { Component, inject, OnInit, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import {
  FormBuilder,
  FormGroup,
  ReactiveFormsModule,
  Validators,
} from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';

import { ToastService } from '../../core/services/toast.service';
import { DokumenService } from './dokumen.service';
import { DokumenForm as DokumenFormBody, VALID_REFF_TYPES } from './dokumen.model';

@Component({
  selector: 'app-dokumen-form',
  imports: [ReactiveFormsModule, RouterLink, TranslatePipe],
  templateUrl: './dokumen-form.html',
  styleUrl: './dokumen-form.scss',
})
export class DokumenFormComponent implements OnInit {
  private fb = inject(FormBuilder);
  private svc = inject(DokumenService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private toast = inject(ToastService);

  readonly saving = signal(false);
  readonly isEdit = signal(false);
  readonly reffTypes = VALID_REFF_TYPES;
  readonly uploading = signal(false);

  dokumenId: number | null = null;

  /** UUID of the uploaded file. */
  private fileUuid: string | null = null;
  readonly uploadedFileName = signal<string | null>(null);
  readonly uploadedFileSize = signal<string | null>(null);

  form: FormGroup = this.fb.nonNullable.group({
    file_uuid: ['', [Validators.required]],
    reff_type: ['', [Validators.required]],
    reff_id: [null as number | null, [Validators.required, Validators.min(1)]],
    kode: ['', [Validators.maxLength(50)]],
    tipe: ['', [Validators.maxLength(50)]],
    jenis: [null as number | null],
    keterangan: ['', [Validators.maxLength(255)]],
  });

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.dokumenId = +id;
      this.isEdit.set(true);
      this.svc.detail(this.dokumenId).subscribe({
        next: (item) => {
          this.form.patchValue({
            reff_type: item.reff_type,
            reff_id: item.reff_id,
            kode: item.kode ?? '',
            tipe: item.tipe ?? '',
            jenis: item.jenis ?? null,
            keterangan: item.keterangan ?? '',
          });

          // Store existing file info (no re-upload needed on edit unless changing).
          if (item.file) {
            this.fileUuid = item.file.uuid;
            this.uploadedFileName.set(item.file.nama_asli);
            this.uploadedFileSize.set(this.formatSize(item.file.ukuran_byte));
          }
        },
        error: () => {
          this.toast.error('DOKUMEN.LOAD_ERROR');
          this.router.navigate(['/dokumen']);
        },
      });
    }
  }

  onFilePicked(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    const file = input.files[0];
    this.uploading.set(true);

    this.svc.uploadFile(file).subscribe({
      next: (res) => {
        this.fileUuid = res.uuid;
        this.form.get('file_uuid')?.setValue(res.uuid);
        this.form.get('file_uuid')?.markAsDirty();
        this.uploadedFileName.set(file.name);
        this.uploadedFileSize.set(this.formatSize(file.size));
        this.uploading.set(false);
      },
      error: () => {
        this.toast.error('DOKUMEN.UPLOAD_ERROR');
        this.uploading.set(false);
      },
    });
    input.value = '';
  }

  submit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    // Ensure file_uuid is set.
    if (!this.fileUuid && !this.isEdit()) {
      this.toast.error('DOKUMEN.FILE_REQUIRED');
      return;
    }

    this.saving.set(true);
    const val = this.form.getRawValue();

    if (this.isEdit()) {
      // On edit, only send changed fields + file_uuid if a new file was picked.
      const body: Partial<DokumenFormBody> = {
        reff_type: val.reff_type,
        reff_id: val.reff_id!,
        kode: val.kode || null,
        tipe: val.tipe || null,
        jenis: val.jenis ?? null,
        keterangan: val.keterangan || null,
      };
      if (this.fileUuid && this.form.get('file_uuid')?.dirty) {
        body.file_uuid = this.fileUuid;
      }
      this.svc.update(this.dokumenId!, body).subscribe({
        next: () => {
          this.toast.success('DOKUMEN.UPDATE_SUCCESS');
          this.router.navigate(['/dokumen']);
        },
        error: (err) => {
          this.saving.set(false);
          if (err?.errors && typeof err.errors === 'object') {
            this.applyServerErrors(err.errors as Record<string, string[]>);
          } else {
            this.toast.error('DOKUMEN.SAVE_ERROR');
          }
        },
      });
    } else {
      const body: DokumenFormBody = {
        file_uuid: this.fileUuid!,
        reff_type: val.reff_type,
        reff_id: val.reff_id!,
        kode: val.kode || null,
        tipe: val.tipe || null,
        jenis: val.jenis ?? null,
        keterangan: val.keterangan || null,
      };
      this.svc.create(body).subscribe({
        next: () => {
          this.toast.success('DOKUMEN.CREATE_SUCCESS');
          this.router.navigate(['/dokumen']);
        },
        error: (err) => {
          this.saving.set(false);
          if (err?.errors && typeof err.errors === 'object') {
            this.applyServerErrors(err.errors as Record<string, string[]>);
          } else {
            this.toast.error('DOKUMEN.SAVE_ERROR');
          }
        },
      });
    }
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

  private formatSize(bytes: number): string {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / 1048576).toFixed(1) + ' MB';
  }
}
