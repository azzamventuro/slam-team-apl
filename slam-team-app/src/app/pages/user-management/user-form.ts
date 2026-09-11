import { Component, inject, OnInit, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';

import { ToastService } from '../../core/services/toast.service';
import { AuthService } from '../../core/services/auth.service';
import { PermissionService } from '../../core/services/permission.service';
import { SlamIconComponent } from '../../shared/components/slam-icon/slam-icon.component';
import { UserService } from './user-management.service';
import {
  AnggotaOpsi,
  RoleOpsi,
  UserForm,
  UserRow,
} from './user-management.model';

@Component({
  selector: 'app-user-form',
  imports: [ReactiveFormsModule, RouterLink, TranslatePipe, SlamIconComponent],
  templateUrl: './user-form.html',
  styleUrl: './user-form.scss',
})
export class UserFormComponent implements OnInit {
  private fb = inject(FormBuilder);
  private svc = inject(UserService);
  private toast = inject(ToastService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private auth = inject(AuthService);
  private perms = inject(PermissionService);

  /** Surface base from route data. */
  readonly base = signal('admin');

  /** Whether we're editing (has :id param). */
  readonly isEdit = signal(false);

  /** Editing user's ID. */
  private editId = 0;

  /** Current form. */
  form!: FormGroup;

  /** Loading state for initial data fetch. */
  readonly loading = signal(true);

  /** Submitting state. */
  readonly submitting = signal(false);

  /** Available roles from API. */
  readonly roles = signal<RoleOpsi[]>([]);

  /** Available anggota from API. */
  readonly anggotaList = signal<AnggotaOpsi[]>([]);

  /** Actor's own role level for anti-escalation. */
  private actorLevel = 99;

  /** Title key based on surface + mode. */
  get titleKey(): string {
    const prefix = this.base() === 'admin' ? 'USERMGMT.ADMIN'
      : this.base() === 'moderator' ? 'USERMGMT.MODERATOR'
      : 'USERMGMT.USER';
    return this.isEdit() ? `${prefix}.EDIT_TITLE` : `${prefix}.CREATE_TITLE`;
  }

  ngOnInit(): void {
    this.base.set(this.route.snapshot.data['base'] || 'admin');

    // Actor's role level for anti-escalation filtering.
    const role = this.auth.user()?.role;
    this.actorLevel = role?.level ?? 99;

    // Build form.
    this.form = this.fb.group({
      anggota_id: [null, Validators.required],
      username: ['', [Validators.required, Validators.minLength(3)]],
      email: ['', [Validators.required, Validators.email]],
      password: ['', this.isEdit() ? [] : [Validators.required, Validators.minLength(6)]],
      role_id: [null, Validators.required],
      timezone: ['Asia/Jakarta'],
      is_aktif: [true],
      buka_kunci: [false],
    });

    // Check if editing.
    const idParam = this.route.snapshot.paramMap.get('id');
    if (idParam) {
      this.isEdit.set(true);
      this.editId = +idParam;
      this.form.get('password')?.clearValidators();
      this.form.get('password')?.updateValueAndValidity();
    }

    // Load reference data.
    this.loadRoles();
    this.loadAnggota();

    // If editing, load the existing user.
    if (this.isEdit()) {
      this.loadUser();
    } else {
      this.loading.set(false);
    }
  }

  private loadRoles(): void {
    this.svc.availableRoles(this.base()).subscribe({
      next: (roles) => this.roles.set(roles),
      error: () => this.toast.show('USERMGMT.ROLES_ERROR', 'danger'),
    });
  }

  private loadAnggota(): void {
    this.svc.availableAnggota(this.base()).subscribe({
      next: (list) => this.anggotaList.set(list),
      error: () => this.toast.show('USERMGMT.ANGGOTA_ERROR', 'danger'),
    });
  }

  private loadUser(): void {
    this.svc.detail(this.base(), this.editId).subscribe({
      next: (u) => {
        this.form.patchValue({
          anggota_id: u.anggota?.id,
          username: u.username,
          email: u.email,
          role_id: u.role?.id,
          timezone: u.timezone || 'Asia/Jakarta',
          is_aktif: u.is_aktif,
        });
        this.loading.set(false);
      },
      error: () => {
        this.toast.show('USERMGMT.LOAD_ERROR', 'danger');
        this.loading.set(false);
      },
    });
  }

  /** Whether the given role is available to the current actor (anti-escalation). */
  isRoleDisabled(r: RoleOpsi): boolean {
    // Cannot assign a role whose level is >= actor's level (anti-escalation).
    return r.level >= this.actorLevel;
  }

  hasError(field: string): boolean {
    const ctrl = this.form.get(field);
    return !!(ctrl && ctrl.invalid && (ctrl.dirty || ctrl.touched));
  }

  onSubmit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.submitting.set(true);
    const val = { ...this.form.value } as UserForm;

    // Don't send password if empty (edit mode without reset).
    if (!val.password) delete val.password;

    if (this.isEdit()) {
      this.svc.update(this.base(), this.editId, val).subscribe({
        next: () => {
          this.toast.show('USERMGMT.UPDATE_SUCCESS', 'success');
          this.router.navigate([`/${this.base()}`]);
        },
        error: (err) => {
          this.handleServerError(err);
          this.submitting.set(false);
        },
      });
    } else {
      this.svc.create(this.base(), val).subscribe({
        next: () => {
          this.toast.show('USERMGMT.CREATE_SUCCESS', 'success');
          this.router.navigate([`/${this.base()}`]);
        },
        error: (err) => {
          this.handleServerError(err);
          this.submitting.set(false);
        },
      });
    }
  }

  private handleServerError(err: unknown): void {
    const body = (err as { error?: { errors?: Record<string, string[]>; message?: string } })?.error;
    if (body?.errors) {
      for (const [field, msgs] of Object.entries(body.errors)) {
        const ctrl = this.form.get(field);
        if (ctrl) ctrl.setErrors({ server: msgs[0] });
      }
    }
    this.toast.show(body?.message || 'USERMGMT.SAVE_ERROR', 'danger');
  }
}
