import { Component } from '@angular/core';

@Component({
  selector: 'app-footer',
  template: `<footer class="border-top text-center text-muted small py-3">
    © {{ year }} Slam Team
  </footer>`,
})
export class Footer {
  readonly year = new Date().getFullYear();
}
