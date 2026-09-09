import { Component, inject } from '@angular/core';
import { TranslatePipe } from '@ngx-translate/core';

import { ToastService, ToastVariant } from '../../../core/services/toast.service';

/**
 * Renders the ToastService queue. Mounted once at the app root so messages
 * survive route changes and show on the login page too.
 */
@Component({
  selector: 'app-toast-host',
  imports: [TranslatePipe],
  templateUrl: './toast-host.html',
  styleUrl: './toast-host.scss',
})
export class ToastHost {
  private toastService = inject(ToastService);
  readonly toasts = this.toastService.toasts;

  dismiss(id: number): void {
    this.toastService.dismiss(id);
  }

  /** Status is never colour-only: each variant carries a glyph and a label. */
  icon(variant: ToastVariant): string {
    switch (variant) {
      case 'success': return '✓';
      case 'warning': return '⚠';
      case 'danger':  return '✕';
      default:        return 'ℹ';
    }
  }

  labelKey(variant: ToastVariant): string {
    return `COMMON.STATUS.${variant.toUpperCase()}`;
  }
}
