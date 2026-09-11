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
import { UnitService } from './unit.service';
import { UnitForm as UnitFormBody, AnggotaOpsi } from './unit.model';
import { HasPermissionDirective } from '../../shared/directives/has-permission.directive';

@Component({
  selector: 'app-unit-form',
  imports: [ReactiveFormsModule, RouterLink, TranslatePipe, HasPermissionDirective],
  templateUrl: './unit-form.html',
  styleUrl: './unit-form.scss',
})
export class UnitFormComponent implements OnInit, OnDestroy {
  private fb = inject(FormBuilder);
  private svc = inject(UnitService);
  private fileSvc = inject(FileService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private toast = inject(ToastService);

  readonly saving = signal(false);
  readonly isEdit = signal(false);
  readonly anggotaList = signal<AnggotaOpsi[]>([]);

  unitId: number | null = null;

  /** Object URL for foto preview — must be revoked on destroy. */
  private fotoUrl = signal<string | null>(null);
  readonly fotoPreview = this.fotoUrl;

  /** UUID resolved from FileService.upload — sent to API. */
  private fotoUuid: string | null = null;

  form: FormGroup = this.fb.nonNullable.group({
    anggota_id: [null as number | null, [Validators.required]],
    kode: ['', [Validators.maxLength(50)]],
    model: ['', [Validators.maxLength(100)]],
    panjang: [null as number | null],
    panjang_inbar: [null as number | null],
    lebar: [null as number | null],
    berat: [null as number | null],
    berat_bb: [null as number | null],
    fps: [null as number | null],
    deskripsi_warna: ['', [Validators.maxLength(150)]],
    disetujui: [false],
  });

  ngOnInit(): void {
    // Load anggota dropdown.
    this.svc.anggotaTersedia().subscribe({
      next: (list) => this.anggotaList.set(list),
    });

    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.unitId = +id;
      this.isEdit.set(true);
      this.svc.detail(this.unitId).subscribe({
        next: (u) => {
          this.form.patchValue({
            anggota_id: u.anggota_id,
            kode: u.kode ?? '',
            model: u.model ?? '',
            panjang: u.panjang ?? null,
            panjang_inbar: u.panjang_inbar ?? null,
            lebar: u.lebar ?? null,
            berat: u.berat ?? null,
            berat_bb: u.berat_bb ?? null,
            fps: u.fps ?? null,
            deskripsi_warna: u.deskripsi_warna ?? '',
            disetujui: u.disetujui,
          });

          // Load existing foto preview.
          if (u.foto_sampul?.uuid) {
            this.fotoUuid = u.foto_sampul.uuid;
            this.fileSvc.imageUrl(u.foto_sampul.uuid, 'medium').subscribe({
              next: (url) => this.fotoUrl.set(url),
            });
          }
        },
        error: () => {
          this.toast.error('UNIT.LOAD_ERROR');
          this.router.navigate(['/unit']);
        },
      });
    }
  }

  ngOnDestroy(): void {
    if (this.fotoUrl()) URL.revokeObjectURL(this.fotoUrl()!);
  }

  // ── File picker ──

  onFotoPicked(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    const file = input.files[0];
    this.fileSvc.upload('file', file, { kategori: 'foto', reff_type: 'unit' }).subscribe({
      next: (res) => {
        this.fotoUuid = res.uuid;
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
    const body: UnitFormBody = {
      anggota_id: val.anggota_id,
      kode: val.kode || null,
      model: val.model || null,
      panjang: val.panjang ?? null,
      panjang_inbar: val.panjang_inbar ?? null,
      lebar: val.lebar ?? null,
      berat: val.berat ?? null,
      berat_bb: val.berat_bb ?? null,
      fps: val.fps ?? null,
      deskripsi_warna: val.deskripsi_warna || null,
      foto_sampul_uuid: this.fotoUuid,
    };

    // Only send disetujui on edit and only if user has permission.
    if (this.isEdit()) {
      body.disetujui = val.disetujui;
    }

    const obs = this.isEdit()
      ? this.svc.update(this.unitId!, body)
      : this.svc.create(body);

    obs.subscribe({
      next: () => {
        this.toast.success(this.isEdit() ? 'UNIT.UPDATE_SUCCESS' : 'UNIT.CREATE_SUCCESS');
        this.router.navigate(['/unit']);
      },
      error: (err) => {
        this.saving.set(false);
        if (err?.errors && typeof err.errors === 'object') {
          this.applyServerErrors(err.errors as Record<string, string[]>);
        } else {
          this.toast.error('UNIT.SAVE_ERROR');
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
