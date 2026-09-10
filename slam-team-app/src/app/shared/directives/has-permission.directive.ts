import { Directive, Input, OnDestroy, TemplateRef, ViewContainerRef, effect, inject } from '@angular/core';

import { PermissionService } from '../../core/services/permission.service';

/**
 * Structural directive that conditionally renders the host element based on
 * a permission code. Uses `effect()` for zoneless-safe reactivity — when the
 * permission set changes (e.g. after `PermissionService.load()`), the view
 * is shown or hidden automatically.
 *
 * Usage:
 * ```html
 * <button *hasPermission="'anggota.create'" class="btn btn-danger">
 *   Tambah Anggota
 * </button>
 * ```
 *
 * Hiding UI elements is UX only — the Go handler enforces permissions.
 */
@Directive({ selector: '[hasPermission]', standalone: true })
export class HasPermissionDirective implements OnDestroy {
  private tpl = inject(TemplateRef<unknown>);
  private vcr = inject(ViewContainerRef);
  private perms = inject(PermissionService);
  private viewRef: unknown = null;

  private effectRef = effect(() => {
    const code = this._code;
    if (!code) return;
    const allow = this.perms.can(code);
    if (allow && !this.viewRef) {
      this.viewRef = this.vcr.createEmbeddedView(this.tpl);
    } else if (!allow && this.viewRef) {
      this.vcr.clear();
      this.viewRef = null;
    }
  });

  private _code = '';

  @Input() set hasPermission(code: string) {
    this._code = code;
  }

  ngOnDestroy(): void {
    this.effectRef.destroy();
  }
}
