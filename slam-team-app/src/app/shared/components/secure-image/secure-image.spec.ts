import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { Component, signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideTranslateService } from '@ngx-translate/core';

import { SecureImage } from './secure-image';

@Component({
  imports: [SecureImage],
  template: `<app-secure-image [uuid]="uuid()" varian="low" />`,
})
class Host {
  readonly uuid = signal<string | null>('aaa');
}

describe('SecureImage', () => {
  let http: HttpTestingController;
  let fixture: ComponentFixture<Host>;
  let created: string[];
  let revoked: string[];

  beforeEach(async () => {
    created = [];
    revoked = [];
    let n = 0;
    URL.createObjectURL = ((): string => {
      const url = `blob:test/${++n}`;
      created.push(url);
      return url;
    }) as typeof URL.createObjectURL;
    URL.revokeObjectURL = ((url: string) => revoked.push(url)) as typeof URL.revokeObjectURL;

    await TestBed.configureTestingModule({
      imports: [Host],
      providers: [provideHttpClient(), provideHttpClientTesting(), provideTranslateService()],
    }).compileComponents();

    http = TestBed.inject(HttpTestingController);
    fixture = TestBed.createComponent(Host);
    fixture.detectChanges();
  });

  afterEach(() => http.verify());

  /** Private bytes must be fetched, never linked — an <img src> carries no token. */
  it('fetches the requested variant as a blob', () => {
    const req = http.expectOne('/api/v1/files/aaa/low');
    expect(req.request.method).toBe('GET');
    expect(req.request.responseType).toBe('blob');
    req.flush(new Blob(['x']));
    fixture.detectChanges();

    expect(created).toEqual(['blob:test/1']);
    expect(fixture.nativeElement.querySelector('img').src).toContain('blob:test/1');
  });

  it('revokes the previous object URL when the uuid changes', () => {
    http.expectOne('/api/v1/files/aaa/low').flush(new Blob(['x']));
    fixture.detectChanges();

    fixture.componentInstance.uuid.set('bbb');
    fixture.detectChanges();
    http.expectOne('/api/v1/files/bbb/low').flush(new Blob(['y']));
    fixture.detectChanges();

    expect(revoked).toContain('blob:test/1');
    expect(created).toEqual(['blob:test/1', 'blob:test/2']);
  });

  it('revokes on destroy, so a scrolled list leaks nothing', () => {
    http.expectOne('/api/v1/files/aaa/low').flush(new Blob(['x']));
    fixture.detectChanges();

    fixture.destroy();
    expect(revoked).toEqual(['blob:test/1']);
  });

  it('renders the placeholder and issues no request without a uuid', () => {
    http.expectOne('/api/v1/files/aaa/low').flush(new Blob(['x']));
    fixture.detectChanges();

    fixture.componentInstance.uuid.set(null);
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('img')).toBeNull();
    expect(revoked).toContain('blob:test/1');
    http.expectNone(() => true);
  });

  it('shows the error box instead of a broken image when the file is gone', () => {
    http
      .expectOne('/api/v1/files/aaa/low')
      .error(new ProgressEvent('error'), { status: 404, statusText: 'Not Found' });
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('img')).toBeNull();
    expect(fixture.nativeElement.querySelector('.secure-image__state--error')).toBeTruthy();
    expect(created.filter((u) => !revoked.includes(u))).toEqual([]);
  });
});
