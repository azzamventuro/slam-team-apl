import { Injectable, signal } from '@angular/core';

export type ToastVariant = 'success' | 'warning' | 'danger' | 'info';

export interface Toast {
  id: number;
  /** An i18n key (preferred) or a literal message — the host runs it through `translate`. */
  text: string;
  variant: ToastVariant;
}

/** Default lifetime of a toast, in ms. */
const DEFAULT_TIMEOUT = 5000;

/**
 * App-wide transient messages. A signal list rendered once by <app-toast-host>,
 * so anything (interceptor, service, page) can report without owning UI.
 */
@Injectable({ providedIn: 'root' })
export class ToastService {
  private readonly _toasts = signal<Toast[]>([]);
  readonly toasts = this._toasts.asReadonly();
  private nextId = 1;

  show(text: string, variant: ToastVariant = 'info', timeout = DEFAULT_TIMEOUT): number {
    const id = this.nextId++;
    this._toasts.update((list) => [...list, { id, text, variant }]);
    if (timeout > 0) setTimeout(() => this.dismiss(id), timeout);
    return id;
  }

  success(text: string) { return this.show(text, 'success'); }
  warning(text: string) { return this.show(text, 'warning'); }
  error(text: string) { return this.show(text, 'danger'); }
  info(text: string) { return this.show(text, 'info'); }

  dismiss(id: number): void {
    this._toasts.update((list) => list.filter((t) => t.id !== id));
  }

  clear(): void {
    this._toasts.set([]);
  }
}
