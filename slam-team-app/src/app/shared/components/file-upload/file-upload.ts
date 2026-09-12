import {
  Component,
  DestroyRef,
  computed,
  effect,
  inject,
  input,
  output,
  signal,
  untracked,
} from '@angular/core';
import { TranslatePipe } from '@ngx-translate/core';

import {
  FileKategori,
  FileRef,
  FileVarian,
  MAX_UPLOAD_MB,
  UploadedFile,
  formatBytes,
  isImageMime,
  validateFile,
} from '../../services/file.model';
import { FileService, fileErrorText } from '../../services/file.service';
import { SecureImage } from '../secure-image/secure-image';

/** Unique per instance so the visually-hidden input and its label stay paired. */
let seq = 0;

/**
 * The upload control every module's form reuses.
 *
 * It is dumb on purpose: it uploads to POST /files and emits the resulting
 * {id, uuid}. Writing that reference into a `*_file_id` column is the owning
 * form's job — this component never knows which module it is serving.
 *
 * Usage: <app-file-upload kategori="foto_profil" reffType="anggota"
 *          [value]="fotoUuid()" (uploaded)="onFoto($event)" (cleared)="onCleared()" />
 *
 * `file.create` belongs to every signed-in role, so the control needs no
 * permission gate of its own; wrap it in `*hasPermission` only where the
 * surrounding action is itself restricted.
 */
@Component({
  selector: 'app-file-upload',
  imports: [TranslatePipe, SecureImage],
  templateUrl: './file-upload.html',
  styleUrl: './file-upload.scss',
})
export class FileUpload {
  private files = inject(FileService);

  /** `mst_file.kategori` — required, and validated again by the server's `oneof`. */
  readonly kategori = input.required<FileKategori>();
  /** Native `accept`, also used for the client-side pre-flight. `.pdf` for documents. */
  readonly accept = input('image/*');
  readonly reffType = input<string | null>(null);
  readonly reffId = input<number | null>(null);
  /** Leave null to let the server derive `is_publik` from the kategori. */
  readonly isPublik = input<boolean | null>(null);
  /** Mirrors FILE_MAX_UPLOAD_MB; 0 disables the client check. */
  readonly maxMb = input(MAX_UPLOAD_MB);
  readonly disabled = input(false);
  /** uuid of an already-stored file, for an edit form. Setting it to null clears the control. */
  readonly value = input<string | null>(null);
  /** Thumbnail size: `low` here, `medium` where the preview is large. */
  readonly previewVarian = input<FileVarian>('low');
  /**
   * Whether "remove" also soft-deletes the file. Leave true for a fresh upload;
   * pass false when the parent only wants to detach a file it still references
   * elsewhere.
   */
  readonly deleteOnRemove = input(true);
  /** i18n key for the field label above the dropzone. */
  readonly labelKey = input<string | null>(null);

  /** The stored reference the parent writes into its form control. */
  readonly uploaded = output<FileRef>();
  /** The file was removed or detached; the parent should null its `*_file_id`. */
  readonly cleared = output<void>();
  readonly errored = output<unknown>();

  readonly inputId = `slam-file-${++seq}`;
  readonly current = signal<UploadedFile | null>(null);
  readonly busy = signal(false);
  readonly dragging = signal(false);
  /** i18n key of the current problem, or null. */
  readonly errorKey = signal<string | null>(null);
  /** Server-supplied detail shown instead of the generic key when present. */
  readonly errorText = signal('');
  readonly localPreview = signal<string | null>(null);

  /** Plain field, not a signal: revoked from inside an effect (see SecureImage). */
  private localUrl: string | null = null;

  /** Images get a thumbnail; documents get a name + size chip. */
  readonly showsThumbnail = computed(() => isImageMime(this.current()?.mime_type));
  readonly fileSize = computed(() => {
    const file = this.current();
    return file ? formatBytes(file.ukuran_byte) : '';
  });
  /** `menunggu` means medium/low are still being written server-side. */
  readonly processing = computed(() => this.current()?.status_proses === 'menunggu');
  readonly interactive = computed(() => !this.disabled() && !this.busy());

  constructor() {
    // Keeps the control in step with a parent that owns the uuid (edit forms,
    // reset-after-save). `current` is read untracked so the effect depends on
    // `value` alone and cannot re-fire on its own writes.
    effect((onCleanup) => {
      const uuid = this.value();
      const shown = untracked(() => this.current());

      if (!uuid) {
        if (shown) this.current.set(null);
        return;
      }
      if (shown?.uuid === uuid) return;

      const sub = this.files.detail(uuid).subscribe({
        next: (file) => this.current.set(file),
        // A uuid that is gone (soft-deleted) or not ours: show the empty
        // dropzone rather than a box that can never load.
        error: () => this.current.set(null),
      });
      onCleanup(() => sub.unsubscribe());
    });

    inject(DestroyRef).onDestroy(() => this.releaseLocal());
  }

  onChange(event: Event): void {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0] ?? null;
    // Reset so re-picking the same file fires `change` again.
    input.value = '';
    this.pick(file);
  }

  onDragOver(event: DragEvent): void {
    if (!this.interactive()) return;
    event.preventDefault();
    this.dragging.set(true);
  }

  onDragLeave(event: DragEvent): void {
    event.preventDefault();
    this.dragging.set(false);
  }

  onDrop(event: DragEvent): void {
    event.preventDefault();
    this.dragging.set(false);
    if (!this.interactive()) return;
    this.pick(event.dataTransfer?.files?.[0] ?? null);
  }

  /** Removes the current file: soft-deletes it unless `deleteOnRemove` is false. */
  remove(): void {
    const file = this.current();
    this.clearError();
    this.releaseLocal();

    if (!file || !this.deleteOnRemove()) {
      this.current.set(null);
      this.cleared.emit();
      return;
    }

    this.busy.set(true);
    this.files.remove(file.uuid).subscribe({
      next: () => {
        this.busy.set(false);
        this.current.set(null);
        this.cleared.emit();
      },
      error: (err: unknown) => this.fail(err),
    });
  }

  private pick(file: File | null): void {
    if (!file || !this.interactive()) return;
    this.clearError();

    const problem = validateFile(file, this.accept(), this.maxMb());
    if (problem) {
      this.releaseLocal();
      this.errorKey.set(problem);
      return;
    }

    // A local object URL fills the box while the bytes are in flight; it is
    // replaced by the server's variant (via <app-secure-image>) on success,
    // and revoked either way.
    this.setLocalPreview(isImageMime(file.type) ? URL.createObjectURL(file) : null);
    this.current.set(null);
    this.busy.set(true);

    this.files
      .uploadFor(this.kategori(), file, {
        reffType: this.reffType(),
        reffId: this.reffId(),
        isPublik: this.isPublik(),
      })
      .subscribe({
        next: (stored) => {
          this.busy.set(false);
          this.releaseLocal();
          this.current.set(stored);
          this.uploaded.emit({ id: stored.id, uuid: stored.uuid });
        },
        error: (err: unknown) => {
          this.releaseLocal();
          this.fail(err);
        },
      });
  }

  private fail(err: unknown): void {
    this.busy.set(false);
    this.errorKey.set('FILE.FAILED');
    this.errorText.set(fileErrorText(err));
    this.errored.emit(err);
  }

  private clearError(): void {
    this.errorKey.set(null);
    this.errorText.set('');
  }

  private setLocalPreview(url: string | null): void {
    this.releaseLocal();
    this.localUrl = url;
    this.localPreview.set(url);
  }

  private releaseLocal(): void {
    if (this.localUrl) URL.revokeObjectURL(this.localUrl);
    this.localUrl = null;
    this.localPreview.set(null);
  }
}
