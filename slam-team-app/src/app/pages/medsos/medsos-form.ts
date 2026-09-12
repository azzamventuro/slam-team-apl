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
import { MedsosService } from './medsos.service';
import { MedsosForm as MedsosFormBody, AnggotaOpsi } from './medsos.model';

@Component({
  selector: 'app-medsos-form',
  imports: [ReactiveFormsModule, RouterLink, TranslatePipe],
  templateUrl: './medsos-form.html',
  styleUrl: './medsos-form.scss',
})
export class MedsosFormComponent implements OnInit {
  private fb = inject(FormBuilder);
  private svc = inject(MedsosService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private toast = inject(ToastService);

  readonly saving = signal(false);
  readonly isEdit = signal(false);
  readonly anggotaList = signal<AnggotaOpsi[]>([]);

  medsosId: number | null = null;

  form: FormGroup = this.fb.nonNullable.group({
    anggota_id: [null as number | null, [Validators.required]],
    jenis_medsos: ['', [Validators.required, Validators.maxLength(50)]],
    konten_medsos: ['', [Validators.required, Validators.maxLength(255)]],
    icon: ['', [Validators.maxLength(50)]],
    kode: ['', [Validators.maxLength(50)]],
    tipe: [0, [Validators.min(0)]],
  });

  ngOnInit(): void {
    this.svc.anggotaTersedia().subscribe({
      next: (list) => this.anggotaList.set(list),
    });

    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.medsosId = +id;
      this.isEdit.set(true);
      this.svc.detail(this.medsosId).subscribe({
        next: (item) => {
          this.form.patchValue({
            anggota_id: item.anggota_id,
            jenis_medsos: item.jenis_medsos,
            konten_medsos: item.konten_medsos,
            icon: item.icon ?? '',
            kode: item.kode ?? '',
            tipe: item.tipe ?? 0,
          });
        },
        error: () => {
          this.toast.error('MEDSOS.LOAD_ERROR');
          this.router.navigate(['/medsos']);
        },
      });
    }
  }

  submit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.saving.set(true);

    const val = this.form.getRawValue();
    const body: MedsosFormBody = {
      anggota_id: val.anggota_id!,
      jenis_medsos: val.jenis_medsos,
      konten_medsos: val.konten_medsos,
      icon: val.icon || null,
      kode: val.kode || null,
      tipe: val.tipe ?? 0,
    };

    const obs = this.isEdit()
      ? this.svc.update(this.medsosId!, body)
      : this.svc.create(body);

    obs.subscribe({
      next: () => {
        this.toast.success(this.isEdit() ? 'MEDSOS.UPDATE_SUCCESS' : 'MEDSOS.CREATE_SUCCESS');
        this.router.navigate(['/medsos']);
      },
      error: (err) => {
        this.saving.set(false);
        if (err?.errors && typeof err.errors === 'object') {
          this.applyServerErrors(err.errors as Record<string, string[]>);
        } else {
          this.toast.error('MEDSOS.SAVE_ERROR');
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
