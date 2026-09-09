import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';

import { ToastHost } from './shared/components/toast-host/toast-host';

@Component({
  selector: 'app-root',
  // The toast host lives at the root so messages (403s, network failures) show
  // on every route, including the unauthenticated login page.
  imports: [RouterOutlet, ToastHost],
  templateUrl: './app.html',
  styleUrl: './app.scss',
})
export class App {}
