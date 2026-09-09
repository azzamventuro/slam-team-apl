import { Component, DestroyRef, effect, inject, input, signal, untracked } from '@angular/core';
import { TranslatePipe } from '@ngx-translate/core';

import { FileVarian } from '../../services/file.model';
import { FileService } from '../../services/file.service';

/** What the box is showing right now — each state carries a glyph, never colour alone. */
type ImageState = 'empty' | 'loading' | 'ready' | 'error';

/**
 * An `<img>` for a **private** file.
 *
 * `GET /files/:uuid/:varian` requires the Bearer token, which a plain
 * `<img src>` cannot send — so the variant is fetched as a blob (the
 * `authInterceptor` attaches the header) and bound as an object URL. The URL is
 * revoked when `uuid`/`varian` change and on destroy, so a long list that
 * scrolls does not leak a blob per row.
 *
 * Prefer `medium` in lists and cards, `low` for thumbnails, `original` only
 * when full resolution is actually needed.
 */
@Component({
  selector: 'app-secure-image',
  imports: [TranslatePipe],
  templateUrl: './secure-image.html',
  styleUrl: './secure-image.scss',
})
export class SecureImage {
  private files = inject(FileService);

  /** `mst_file.uuid`. Null/empty renders the placeholder instead of firing a request. */
  readonly uuid = input<string | null | undefined>(null);
  readonly varian = input<FileVarian>('medium');
  /** Empty alt is correct for decorative thumbnails; pass a description otherwise. */
  readonly alt = input('');
  /** CSS `aspect-ratio` for the box (e.g. `1 / 1`, `16 / 9`). Null keeps the image's own. */
  readonly ratio = input<string | null>(null);
  /** `cover` crops to fill the box; `contain` fits the whole image inside it. */
  readonly fit = input<'cover' | 'contain'>('cover');

  readonly state = signal<ImageState>('empty');
  readonly src = signal<string | null>(null);

  /**
   * Held as a plain field, not a signal: `release` runs inside the effect, and
   * reading a signal there would make the effect depend on its own output.
   */
  private objectUrl: string | null = null;

  constructor() {
    effect((onCleanup) => {
      const uuid = this.uuid();
      const varian = this.varian();

      untracked(() => this.release());

      if (!uuid) {
        this.state.set('empty');
        return;
      }

      this.state.set('loading');
      const sub = this.files.imageUrl(uuid, varian).subscribe({
        next: (url) => {
          this.objectUrl = url;
          this.src.set(url);
          this.state.set('ready');
        },
        // 404 (deleted), 403 (not yours) and a dropped connection all land here:
        // show the broken-image box rather than an <img> that will never paint.
        error: () => this.state.set('error'),
      });

      // A uuid that changes mid-flight cancels the previous request, so a fast
      // scroll cannot resolve an old blob over the new one.
      onCleanup(() => sub.unsubscribe());
    });

    inject(DestroyRef).onDestroy(() => this.release());
  }

  /** Revokes the current object URL and clears the binding. Safe to call twice. */
  private release(): void {
    if (this.objectUrl) URL.revokeObjectURL(this.objectUrl);
    this.objectUrl = null;
    this.src.set(null);
  }
}
