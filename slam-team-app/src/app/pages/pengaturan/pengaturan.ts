import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import {
  AbstractControl,
  FormControl,
  ReactiveFormsModule,
  ValidationErrors,
} from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';

import { ApiResponse } from '../../core/models/api.model';
import { AuthService } from '../../core/services/auth.service';
import { ToastService } from '../../core/services/toast.service';
import { Setting, SettingGroup, SettingKV } from './pengaturan.model';
import { PengaturanService } from './pengaturan.service';

/** A malformed JSON textarea blocks submit; empty is left to the server ("nilai wajib diisi"). */
function jsonValidator(control: AbstractControl): ValidationErrors | null {
  const value = control.value;
  if (value === null || value === undefined || value === '') return null;
  try {
    JSON.parse(String(value));
    return null;
  } catch {
    return { json: true };
  }
}

@Component({
  selector: 'app-pengaturan',
  imports: [ReactiveFormsModule, TranslatePipe],
  templateUrl: './pengaturan.html',
  styleUrl: './pengaturan.scss',
})
export class Pengaturan {
  private svc = inject(PengaturanService);
  private auth = inject(AuthService);
  private toast = inject(ToastService);

  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly groups = signal<SettingGroup[]>([]);
  readonly activeGrup = signal<string>('');

  /** is_terkunci rows are only editable by a super admin — mirrors the service rule. */
  readonly isSuper = computed(() => this.auth.user()?.role?.is_super ?? false);

  private controls = new Map<string, FormControl>();

  constructor() {
    this.load();
  }

  controlFor(kunci: string): FormControl {
    // Populated synchronously in buildForm before groups() ever renders a row.
    return this.controls.get(kunci) as FormControl;
  }

  resetToDefault(setting: Setting): void {
    const ctrl = this.controlFor(setting.kunci);
    if (ctrl.disabled) return;
    ctrl.setValue(this.toControlValue(setting.tipe_nilai, setting.nilai_bawaan));
    ctrl.markAsDirty();
  }

  save(): void {
    if (this.saving()) return;

    let hasInvalid = false;
    for (const ctrl of this.controls.values()) {
      if (ctrl.enabled && ctrl.invalid) {
        ctrl.markAsTouched();
        hasInvalid = true;
      }
    }
    if (hasInvalid) {
      this.toast.error('VALIDATION.JSON');
      return;
    }

    const items = this.collectChanges();
    if (!items.length) {
      this.toast.info('PENGATURAN.NO_CHANGES');
      return;
    }

    this.saving.set(true);
    this.svc.bulkUpdate(items).subscribe({
      next: (groups) => {
        this.applyGroups(groups);
        this.saving.set(false);
        this.toast.success('COMMON.SAVED');
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.applyServerErrors(err);
      },
    });
  }

  private load(): void {
    this.loading.set(true);
    this.svc.list().subscribe({
      next: (groups) => {
        this.applyGroups(groups);
        this.loading.set(false);
      },
      error: () => {
        this.toast.error('COMMON.SERVER_ERROR');
        this.loading.set(false);
      },
    });
  }

  private applyGroups(groups: SettingGroup[]): void {
    this.groups.set(groups);
    this.buildForm(groups);
    if (!this.activeGrup() && groups.length) this.activeGrup.set(groups[0].grup);
  }

  private buildForm(groups: SettingGroup[]): void {
    this.controls.clear();
    for (const group of groups) {
      for (const setting of group.items) {
        const ctrl = new FormControl(this.toControlValue(setting.tipe_nilai, setting.nilai), {
          nonNullable: false,
          validators: setting.tipe_nilai === 'json' ? [jsonValidator] : [],
        });
        if (setting.is_terkunci && !this.isSuper()) ctrl.disable();
        this.controls.set(setting.kunci, ctrl);
      }
    }
  }

  private collectChanges(): SettingKV[] {
    const items: SettingKV[] = [];
    for (const group of this.groups()) {
      for (const setting of group.items) {
        const ctrl = this.controlFor(setting.kunci);
        if (ctrl.disabled || !ctrl.dirty) continue;
        items.push({ kunci: setting.kunci, nilai: this.toPayloadValue(setting.tipe_nilai, ctrl.value) });
      }
    }
    return items;
  }

  /** Cast → form control value. json becomes pretty text for the textarea. */
  private toControlValue(tipe: Setting['tipe_nilai'], value: unknown): unknown {
    if (tipe === 'json') return value === null || value === undefined ? '' : JSON.stringify(value, null, 2);
    return value;
  }

  /** Form control value → PUT payload. json is parsed back into an object. */
  private toPayloadValue(tipe: Setting['tipe_nilai'], value: unknown): unknown {
    if (tipe === 'json') {
      try {
        return JSON.parse(String(value));
      } catch {
        return value;
      }
    }
    return value;
  }

  private applyServerErrors(err: HttpErrorResponse): void {
    const body = err.error as ApiResponse<unknown> | undefined;
    const errors = body?.errors;
    if (errors && typeof errors === 'object' && !Array.isArray(errors)) {
      for (const [kunci, msg] of Object.entries(errors as Record<string, unknown>)) {
        const ctrl = this.controls.get(kunci);
        if (!ctrl) continue;
        ctrl.setErrors({ server: Array.isArray(msg) ? String(msg[0]) : String(msg) });
        ctrl.markAsTouched();
      }
    }
    this.toast.error(err.status === 403 ? 'COMMON.FORBIDDEN' : 'COMMON.SERVER_ERROR');
  }
}
