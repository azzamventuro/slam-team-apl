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
import { AnggotaService } from './anggota.service';
import { AnggotaForm } from '../../core/models/anggota.model';
import { Instansi } from '../../core/models/instansi.model';
import { ApiService } from '../../core/services/api.service';
import { Page } from '../../core/models/api.model';

@Component({
  selector: 'app-anggota-form',
  imports: [ReactiveFormsModule, RouterLink, TranslatePipe],
  templateUrl: './anggota-form.html',
  styleUrl: './anggota-form.scss',
})
export class AnggotaFormComponent implements OnInit, OnDestroy {
  private fb = inject(FormBuilder);
  private svc = inject(AnggotaService);
  private fileSvc = inject(FileService);
  private api = inject(ApiService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private toast = inject(ToastService);

  readonly saving = signal(false);
  readonly isEdit = signal(false);
  readonly instansiList = signal<Instansi[]>([]);

  anggotaId: number | null = null;

  /** Object URLs for preview — must be revoked on destroy. */
  private fotoProfilUrl = signal<string | null>(null);
  private fotoFormalUrl = signal<string | null>(null);
  private fileIdentitasUrl = signal<string | null>(null);
  readonly fotoProfilPreview = this.fotoProfilUrl;
  readonly fotoFormalPreview = this.fotoFormalUrl;
  readonly fileIdentitasPreview = this.fileIdentitasUrl;

  /** File IDs resolved from FileService.upload. */
  private fotoProfilFileId: number | null = null;
  private fotoFormalFileId: number | null = null;
  private fileIdentitasFileId: number | null = null;

  form: FormGroup = this.fb.nonNullable.group({
    instansi_id: [null as number | null, [Validators.required]],
    nama_lengkap: ['', [Validators.required, Validators.maxLength(150)]],
    nama_panggilan: ['', [Validators.maxLength(50)]],
    jenis_anggota: ['', [Validators.required]],
    jenis_kelamin: [null as number | null],
    jenis_identitas: [null as number | null],
    no_identitas: ['', [Validators.maxLength(50)]],
    pekerjaan: ['', [Validators.maxLength(100)]],
    alamat: [''],
    kode_pos: ['', [Validators.maxLength(10)]],
    tempat_lahir: ['', [Validators.maxLength(100)]],
    tanggal_lahir: ['', [Validators.required]],
    tanggal_bergabung: [''],
    status_anggota: ['aktif'],
  });

  ngOnInit(): void {
    // Load instansi list for dropdown.
    this.api.get<Page<Instansi>>('/instansi', { per_page: 100 }).subscribe({
      next: (p) => this.instansiList.set(p.items),
    });

    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.anggotaId = +id;
      this.isEdit.set(true);
      this.svc.detail(this.anggotaId).subscribe({
        next: (a) => {
          this.form.patchValue({
            instansi_id: a.instansi_id,
            nama_lengkap: a.nama_lengkap,
            nama_panggilan: a.nama_panggilan ?? '',
            jenis_anggota: a.jenis_anggota,
            jenis_kelamin: a.jenis_kelamin,
            jenis_identitas: a.jenis_identitas,
            no_identitas: a.no_identitas ?? '',
            pekerjaan: a.pekerjaan ?? '',
            alamat: a.alamat ?? '',
            kode_pos: a.kode_pos ?? '',
            tempat_lahir: a.tempat_lahir ?? '',
            tanggal_lahir: a.tanggal_lahir,
            tanggal_bergabung: a.tanggal_bergabung ?? '',
            status_anggota: a.status_anggota,
          });
          this.fotoProfilFileId = a.foto_profil_file_id;
          this.fotoFormalFileId = a.foto_formal_file_id;
          this.fileIdentitasFileId = a.file_identitas_file_id;

          // Load preview images for existing files.
          if (a.foto_profil_uuid) {
            this.fileSvc.imageUrl(a.foto_profil_uuid, 'medium').subscribe((url) => this.fotoProfilUrl.set(url));
          }
          if (a.foto_formal_uuid) {
            this.fileSvc.imageUrl(a.foto_formal_uuid, 'medium').subscribe((url) => this.fotoFormalUrl.set(url));
          }
          if (a.file_identitas_uuid) {
            this.fileSvc.imageUrl(a.file_identitas_uuid, 'medium').subscribe((url) => this.fileIdentitasUrl.set(url));
          }
        },
        error: () => {
          this.toast.error('ANGGOTA.LOAD_ERROR');
          this.router.navigate(['/anggota']);
        },
      });
    }
  }

  ngOnDestroy(): void {
    if (this.fotoProfilUrl()) URL.revokeObjectURL(this.fotoProfilUrl()!);
    if (this.fotoFormalUrl()) URL.revokeObjectURL(this.fotoFormalUrl()!);
    if (this.fileIdentitasUrl()) URL.revokeObjectURL(this.fileIdentitasUrl()!);
  }

  // ── File pickers ──

  onFotoProfilPicked(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    const file = input.files[0];
    this.fileSvc.upload('file', file, { kategori: 'foto', reff_type: 'anggota' }).subscribe({
      next: (res) => {
        this.fotoProfilFileId = res.id;
        if (this.fotoProfilUrl()) URL.revokeObjectURL(this.fotoProfilUrl()!);
        this.fotoProfilUrl.set(URL.createObjectURL(file));
      },
      error: () => this.toast.error('COMMON.SERVER_ERROR'),
    });
    input.value = '';
  }

  onFotoFormalPicked(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    const file = input.files[0];
    this.fileSvc.upload('file', file, { kategori: 'foto', reff_type: 'anggota' }).subscribe({
      next: (res) => {
        this.fotoFormalFileId = res.id;
        if (this.fotoFormalUrl()) URL.revokeObjectURL(this.fotoFormalUrl()!);
        this.fotoFormalUrl.set(URL.createObjectURL(file));
      },
      error: () => this.toast.error('COMMON.SERVER_ERROR'),
    });
    input.value = '';
  }

  onFileIdentitasPicked(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    const file = input.files[0];
    this.fileSvc.upload('file', file, { kategori: 'dokumen', reff_type: 'anggota' }).subscribe({
      next: (res) => {
        this.fileIdentitasFileId = res.id;
        if (this.fileIdentitasUrl()) URL.revokeObjectURL(this.fileIdentitasUrl()!);
        this.fileIdentitasUrl.set(URL.createObjectURL(file));
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
    const body: AnggotaForm = {
      instansi_id: val.instansi_id,
      nama_lengkap: val.nama_lengkap,
      nama_panggilan: val.nama_panggilan || null,
      jenis_anggota: val.jenis_anggota,
      jenis_kelamin: val.jenis_kelamin,
      jenis_identitas: val.jenis_identitas,
      no_identitas: val.no_identitas || null,
      pekerjaan: val.pekerjaan || null,
      alamat: val.alamat || null,
      kode_pos: val.kode_pos || null,
      tempat_lahir: val.tempat_lahir || null,
      tanggal_lahir: val.tanggal_lahir,
      tanggal_bergabung: val.tanggal_bergabung || null,
      status_anggota: val.status_anggota,
      foto_profil_file_id: this.fotoProfilFileId,
      foto_formal_file_id: this.fotoFormalFileId,
      file_identitas_file_id: this.fileIdentitasFileId,
    };

    const obs = this.isEdit()
      ? this.svc.update(this.anggotaId!, body)
      : this.svc.create(body);

    obs.subscribe({
      next: () => {
        this.toast.success(this.isEdit() ? 'ANGGOTA.UPDATE_SUCCESS' : 'ANGGOTA.CREATE_SUCCESS');
        this.router.navigate(['/anggota']);
      },
      error: (err) => {
        this.saving.set(false);
        if (err?.errors && typeof err.errors === 'object') {
          this.applyServerErrors(err.errors as Record<string, string[]>);
        } else {
          this.toast.error('ANGGOTA.SAVE_ERROR');
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
