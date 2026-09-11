import { Component, inject, OnInit, OnDestroy, signal } from '@angular/core';
import {
  FormBuilder,
  FormGroup,
  ReactiveFormsModule,
  Validators,
} from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';

import { ToastService } from '../../core/services/toast.service';
import { ProfileClubService } from './profile-club.service';
import { ProfileClubForm } from './profile-club.model';

@Component({
  selector: 'app-profile-club',
  imports: [ReactiveFormsModule, TranslatePipe],
  templateUrl: './profile-club.html',
  styleUrl: './profile-club.scss',
})
export class ProfileClubComponent implements OnInit, OnDestroy {
  private fb = inject(FormBuilder);
  private svc = inject(ProfileClubService);
  private toast = inject(ToastService);

  readonly saving = signal(false);
  readonly loading = signal(true);
  /** null = no profile yet (create mode), number = existing id (edit mode). */
  readonly profileId = signal<number | null>(null);

  /** Object URLs for image previews — revoked on destroy. */
  private bannerUrl = signal<string | null>(null);
  private logoSimpleUrl = signal<string | null>(null);
  private logoBesarUrl = signal<string | null>(null);
  readonly bannerPreview = this.bannerUrl.asReadonly();
  readonly logoSimplePreview = this.logoSimpleUrl.asReadonly();
  readonly logoBesarPreview = this.logoBesarUrl.asReadonly();

  /** File IDs from upload. */
  private bannerFileId: number | null = null;
  private logoSimpleFileId: number | null = null;
  private logoBesarFileId: number | null = null;

  form: FormGroup = this.fb.nonNullable.group({
    nama: ['', [Validators.required, Validators.maxLength(150)]],
    singkatan: ['', [Validators.maxLength(50)]],
    alamat: [''],
    keterangan: [''],
  });

  ngOnInit(): void {
    this.svc.list().subscribe({
      next: (page) => {
        this.loading.set(false);
        if (page.items.length > 0) {
          const row = page.items[0];
          this.profileId.set(row.id);
          this.form.patchValue({
            nama: row.nama,
            singkatan: row.singkatan ?? '',
            alamat: row.alamat ?? '',
            keterangan: row.keterangan ?? '',
          });
          this.bannerFileId = row.banner_file_id;
          this.logoSimpleFileId = row.logo_simple_file_id;
          this.logoBesarFileId = row.logo_besar_file_id;
          if (row.banner_file_uuid) {
            this.svc.imageUrl(row.banner_file_uuid).subscribe((url) => this.bannerUrl.set(url));
          }
          if (row.logo_simple_file_uuid) {
            this.svc.imageUrl(row.logo_simple_file_uuid).subscribe((url) => this.logoSimpleUrl.set(url));
          }
          if (row.logo_besar_file_uuid) {
            this.svc.imageUrl(row.logo_besar_file_uuid).subscribe((url) => this.logoBesarUrl.set(url));
          }
        }
      },
      error: () => {
        this.loading.set(false);
        this.toast.error('PROFILE_CLUB.LOAD_ERROR');
      },
    });
  }

  ngOnDestroy(): void {
    this.revokeAll();
  }

  // ─── Image picks ───

  onBannerPicked(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    const file = input.files[0];
    this.svc.uploadFile(file, 'banner').subscribe({
      next: (res) => {
        this.bannerFileId = res.id;
        this.revoke(this.bannerUrl());
        this.bannerUrl.set(URL.createObjectURL(file));
      },
      error: () => this.toast.error('COMMON.SERVER_ERROR'),
    });
    input.value = '';
  }

  onLogoSimplePicked(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    const file = input.files[0];
    this.svc.uploadFile(file, 'logo').subscribe({
      next: (res) => {
        this.logoSimpleFileId = res.id;
        this.revoke(this.logoSimpleUrl());
        this.logoSimpleUrl.set(URL.createObjectURL(file));
      },
      error: () => this.toast.error('COMMON.SERVER_ERROR'),
    });
    input.value = '';
  }

  onLogoBesarPicked(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    const file = input.files[0];
    this.svc.uploadFile(file, 'logo').subscribe({
      next: (res) => {
        this.logoBesarFileId = res.id;
        this.revoke(this.logoBesarUrl());
        this.logoBesarUrl.set(URL.createObjectURL(file));
      },
      error: () => this.toast.error('COMMON.SERVER_ERROR'),
    });
    input.value = '';
  }

  // ─── Submit ───

  submit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.saving.set(true);
    const raw = this.form.getRawValue();
    const body: ProfileClubForm = {
      ...raw,
      banner_file_id: this.bannerFileId,
      logo_simple_file_id: this.logoSimpleFileId,
      logo_besar_file_id: this.logoBesarFileId,
    };

    const req$ = this.profileId() !== null
      ? this.svc.update(this.profileId()!, body)
      : this.svc.create(body);

    req$.subscribe({
      next: (res) => {
        this.profileId.set(res.id);
        this.toast.success('COMMON.SAVED');
        this.saving.set(false);
      },
      error: (e) => {
        this.applyServerErrors(e?.errors);
        this.saving.set(false);
      },
    });
  }

  // ─── Helpers ───

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

  private revoke(url: string | null): void {
    if (url) URL.revokeObjectURL(url);
  }

  private revokeAll(): void {
    this.revoke(this.bannerUrl());
    this.revoke(this.logoSimpleUrl());
    this.revoke(this.logoBesarUrl());
  }
}
