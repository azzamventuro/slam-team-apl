import { Pipe, PipeTransform, inject } from '@angular/core';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';

/** Marks a trusted HTML string as safe for [innerHTML]. Standalone pipe. */
@Pipe({ name: 'safe' })
export class SafePipe implements PipeTransform {
  private sanitizer = inject(DomSanitizer);

  transform(value: string): SafeHtml {
    return this.sanitizer.bypassSecurityTrustHtml(value);
  }
}
