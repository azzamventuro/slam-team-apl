import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';

import { ApiResponse } from '../../core/models/api.model';
import { AuthService } from '../../core/services/auth.service';
import { ToastService } from '../../core/services/toast.service';
import { FileUpload } from '../../shared/components/file-upload/file-upload';
import { FileRef } from '../../shared/services/file.model';
import {
  CreateIzinReq,
  JENIS_IZIN,
  JadwalOpsi,
  JenisIzin,
  SesiOpsi,
  hhmm,
  jadwalLabel,
} from './izin.model';
import { IzinService } from './izin.service';

/**
 * "Ajukan izin" — one request for one session. The session is picked in two
 * steps (jadwal → its materialised sesi) because there is no
 * "my assignments" endpoint; cancelled sessions are hidden since the API
 * refuses them anyway (422 `sesi_id`).
 *
 * `waktu_pulang_diminta` is shown — and required — only for
 * `jenis = pulang_cepat`; for every other jenis the field is left out of the
 * body entirely, mirroring the service's cross-field rule.
 */
@Component({
  selector: 'app-izin-form',
  imports: [ReactiveFormsModule, RouterLink, TranslatePipe, FileUpload],
  templateUrl: './izin-form.html',
  styleUrl: './izin-form.scss',
})
export class IzinForm {
  private fb = inject(FormBuilder);
  private svc = inject(IzinService);
  private router = inject(Router);
  private toast = inject(ToastService);
  private auth = inject(AuthService);

  readonly jenisOpsi = JENIS_IZIN;
  readonly saving = signal(false);
  readonly jadwalList = signal<JadwalOpsi[]>([]);
  readonly sesiList = signal<SesiOpsi[]>([]);
  readonly loadingSesi = signal(false);
  readonly jenis = signal<JenisIzin>('izin');
  readonly isPulangCepat = computed(() => this.jenis() === 'pulang_cepat');

  /**
   * An account not yet linked to an anggota cannot request anything — the
   * service answers 403. Say so up front instead of after the round trip.
   */
  readonly noAnggota = computed(() => !this.auth.user()?.anggota_id);

  /** uuid of the uploaded lampiran — bound to <app-file-upload>'s [value]. */
  readonly lampiranUuid = signal<string | null>(null);
  /** Numeric FK resolved from the upload's {id, uuid} — what goes on the wire. */
  private lampiranFileId: number | null = null;

  form: FormGroup = this.fb.nonNullable.group({
    jadwal_id: [null as number | null, [Validators.required]],
    sesi_id: [null as number | null, [Validators.required]],
    jenis: ['izin' as JenisIzin, [Validators.required]],
    alasan: ['', [Validators.required, Validators.maxLength(2000)]],
    waktu_pulang_diminta: [''],
  });

  constructor() {
    this.svc.jadwalOpsi().subscribe({
      next: (p) => this.jadwalList.set(p.items),
      error: () => this.toast.error('IZIN.LOAD_JADWAL_ERROR'),
    });

    this.form.controls['jadwal_id'].valueChanges
      .pipe(takeUntilDestroyed())
      .subscribe((id: number | null) => this.loadSesi(id));

    // The time field is required exactly when pulang_cepat is chosen; the
    // signal drives the template, the validator drives submit.
    this.form.controls['jenis'].valueChanges
      .pipe(takeUntilDestroyed())
      .subscribe((j: JenisIzin) => {
        this.jenis.set(j);
        const waktu = this.form.controls['waktu_pulang_diminta'];
        if (j === 'pulang_cepat') {
          waktu.setValidators([Validators.required]);
        } else {
          waktu.clearValidators();
          waktu.setValue('');
        }
        waktu.updateValueAndValidity();
      });
  }

  private loadSesi(jadwalId: number | null): void {
    this.form.controls['sesi_id'].setValue(null);
    this.sesiList.set([]);
    if (!jadwalId) return;
    this.loadingSesi.set(true);
    this.svc.sesiOpsi(jadwalId).subscribe({
      next: (rows) => {
        this.sesiList.set(rows.filter((s) => s.status !== 'dibatalkan'));
        this.loadingSesi.set(false);
      },
      error: () => {
        this.loadingSesi.set(false);
        this.toast.error('IZIN.LOAD_SESI_ERROR');
      },
    });
  }

  jadwalLabel(j: JadwalOpsi): string {
    return jadwalLabel(j);
  }

  sesiLabel(s: SesiOpsi): string {
    return `${s.tanggal_lokal} · ${hhmm(s.mulai_lokal)}–${hhmm(s.selesai_lokal)}`;
  }

  // ── Lampiran ──

  onLampiranUploaded(ref: FileRef): void {
    this.lampiranFileId = ref.id ?? null;
    this.lampiranUuid.set(ref.uuid);
  }

  onLampiranCleared(): void {
    this.lampiranFileId = null;
    this.lampiranUuid.set(null);
  }

  // ── Submit ──

  submit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.saving.set(true);
    const raw = this.form.getRawValue();
    const body: CreateIzinReq = {
      sesi_id: raw.sesi_id,
      jenis: raw.jenis,
      alasan: raw.alasan.trim(),
      lampiran_file_id: this.lampiranFileId,
    };
    if (raw.jenis === 'pulang_cepat') body.waktu_pulang_diminta = raw.waktu_pulang_diminta;

    this.svc.create(body).subscribe({
      next: () => {
        this.toast.success('IZIN.CREATE_SUCCESS');
        this.router.navigate(['/izin']);
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.applyServerErrors(err);
      },
    });
  }

  /**
   * 422 → the matching control (inline); 409 (an open request already exists
   * for this session) and 403 (account without anggota) get their own toasts.
   */
  private applyServerErrors(err: HttpErrorResponse): void {
    const body = err.error as ApiResponse<unknown> | undefined;
    const errors = body?.errors;
    let mapped = false;
    if (errors && typeof errors === 'object' && !Array.isArray(errors)) {
      for (const [field, msg] of Object.entries(errors as Record<string, unknown>)) {
        const text = Array.isArray(msg) ? String(msg[0]) : String(msg);
        const ctrl = this.form.get(field);
        if (ctrl) {
          ctrl.setErrors({ server: text });
          ctrl.markAsTouched();
          mapped = true;
        } else if (field === 'lampiran_file_id') {
          // No control of its own — the upload box is the field.
          this.toast.error(text);
          mapped = true;
        }
      }
    }
    if (err.status === 409) {
      this.toast.error('IZIN.DUPLICATE');
    } else if (err.status === 403) {
      this.toast.error('IZIN.NO_ANGGOTA');
    } else if (!mapped) {
      this.toast.error('IZIN.SAVE_ERROR');
    }
  }

  hasError(field: string, error: string): boolean {
    const ctrl = this.form.get(field);
    if (!ctrl) return false;
    return ctrl.hasError(error) && (ctrl.dirty || ctrl.touched);
  }

  serverError(field: string): string | null {
    const ctrl = this.form.get(field);
    if (!ctrl || !(ctrl.dirty || ctrl.touched)) return null;
    return (ctrl.getError('server') as string | undefined) ?? null;
  }
}
