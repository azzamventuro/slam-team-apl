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
import { CreateRoleReq, UpdateRoleReq } from './hak-akses.model';
import { HakAksesService } from './hak-akses.service';

/**
 * RoleForm handles both create (POST) and edit (PUT) for roles.
 * Anti-escalation is enforced server-side — the form does not block editing
 * by level (that's the list's job).
 */
@Component({
  selector: 'app-role-form',
  imports: [ReactiveFormsModule, RouterLink, TranslatePipe],
  templateUrl: './role-form.html',
  styleUrl: './role-form.scss',
})
export class RoleForm implements OnInit {
  private fb = inject(FormBuilder);
  private svc = inject(HakAksesService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private toast = inject(ToastService);

  readonly saving = signal(false);
  readonly isEdit = signal(false);

  roleId: number | null = null;

  form: FormGroup = this.fb.nonNullable.group({
    kode: ['', [Validators.required, Validators.maxLength(30), Validators.pattern(/^[a-z_]+$/)]],
    nama: ['', [Validators.required, Validators.maxLength(60)]],
    level: [20, [Validators.required, Validators.min(0), Validators.max(100)]],
    keterangan: [''],
  });

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.roleId = +id;
      this.isEdit.set(true);
      // Kode is not editable on update.
      this.form.get('kode')?.disable();
      this.svc.getRole(this.roleId).subscribe({
        next: (role) => {
          this.form.patchValue({
            kode: role.kode,
            nama: role.nama,
            level: role.level,
            keterangan: role.keterangan,
          });
        },
        error: () => {
          this.toast.error('HAK_AKSES.LOAD_ERROR');
          this.router.navigate(['/hak-akses']);
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
    const raw = this.form.getRawValue();

    const req$ = this.isEdit()
      ? this.svc.updateRole(this.roleId!, raw as UpdateRoleReq)
      : this.svc.createRole(raw as CreateRoleReq);

    req$.subscribe({
      next: () => {
        this.toast.success('COMMON.SAVED');
        this.router.navigate(['/hak-akses']);
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

  /** Check if a field has a specific error type. */
  hasError(field: string, error: string): boolean {
    const ctrl = this.form.get(field);
    if (!ctrl) return false;
    return ctrl.hasError(error) && (ctrl.dirty || ctrl.touched);
  }
}
