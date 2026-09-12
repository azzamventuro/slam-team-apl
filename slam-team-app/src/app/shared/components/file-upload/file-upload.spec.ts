import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { Component, signal, viewChild } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideTranslateService } from '@ngx-translate/core';

import { ApiResponse } from '../../../core/models/api.model';
import { FileRef, UploadedFile } from '../../services/file.model';
import { FileUpload } from './file-upload';

@Component({
  imports: [FileUpload],
  template: `<app-file-upload
    kategori="foto_profil"
    reffType="anggota"
    [reffId]="reffId()"
    [value]="value()"
    [maxMb]="15"
    (uploaded)="refs.push($event)"
    (cleared)="clears = clears + 1"
  />`,
})
class Host {
  readonly upload = viewChild.required(FileUpload);
  readonly reffId = signal<number | null>(7);
  readonly value = signal<string | null>(null);
  refs: FileRef[] = [];
  clears = 0;
}

function stored(over: Partial<UploadedFile> = {}): ApiResponse<UploadedFile> {
  return {
    success: true,
    message: 'created',
    data: {
      uuid: '8f3c',
      nama_asli: 'IMG_2201.jpg',
      nama_slug: 'img-2201',
      ekstensi: 'jpg',
      mime_type: 'image/jpeg',
      ukuran_byte: 1843201,
      kategori: 'foto_profil',
      is_publik: false,
      status_proses: 'menunggu',
      hash_sha256: 'a'.repeat(64),
      lebar_px: 3024,
      tinggi_px: 4032,
      variants: [],
      url: '/api/v1/files/8f3c/medium',
      ...over,
    },
  };
}

/** Drives the same code path as the file input, without a real FileList. */
function drop(fixture: ComponentFixture<Host>, file: File): void {
  fixture.componentInstance.upload().onDrop({
    preventDefault: () => undefined,
    dataTransfer: { files: [file] },
  } as unknown as DragEvent);
  fixture.detectChanges();
}

describe('FileUpload', () => {
  let http: HttpTestingController;
  let fixture: ComponentFixture<Host>;
  let created: string[];
  let revoked: string[];
  const jpeg = () => new File(['bytes'], 'IMG_2201.jpg', { type: 'image/jpeg' });

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

  it('uploads the dropped file with its kategori and owner, then emits the stored ref', () => {
    drop(fixture, jpeg());

    const req = http.expectOne('/api/v1/files');
    expect(req.request.headers.has('Content-Type')).toBe(false);
    const body = req.request.body as FormData;
    expect(body.get('kategori')).toBe('foto_profil');
    expect(body.get('reff_type')).toBe('anggota');
    expect(body.get('reff_id')).toBe('7');

    req.flush(stored());
    fixture.detectChanges();

    expect(fixture.componentInstance.refs).toEqual([{ id: undefined, uuid: '8f3c' }]);
    // The stored preview comes from the authenticated low variant, never a
    // token-less <img src>.
    http.expectOne('/api/v1/files/8f3c/low').flush(new Blob(['x']));
  });

  it('revokes the pre-upload preview once the server copy takes over', () => {
    drop(fixture, jpeg());
    expect(created).toEqual(['blob:test/1']);

    http.expectOne('/api/v1/files').flush(stored());
    fixture.detectChanges();

    expect(revoked).toContain('blob:test/1');
    http.expectOne('/api/v1/files/8f3c/low').flush(new Blob(['y']));
    fixture.detectChanges();
    fixture.destroy();

    expect(created.filter((u) => !revoked.includes(u))).toEqual([]);
  });

  it('rejects an oversized file locally, without a request', () => {
    const big = jpeg();
    Object.defineProperty(big, 'size', { value: 16 * 1024 * 1024 });
    drop(fixture, big);

    expect(fixture.componentInstance.upload().errorKey()).toBe('FILE.TOO_LARGE');
    expect(fixture.nativeElement.querySelector('.file-upload__error')).toBeTruthy();
    expect(fixture.componentInstance.refs).toEqual([]);
  });

  it('rejects a file the accept rule excludes', () => {
    drop(fixture, new File(['x'], 'surat.pdf', { type: 'application/pdf' }));

    expect(fixture.componentInstance.upload().errorKey()).toBe('FILE.WRONG_TYPE');
    expect(created).toEqual([]);
  });

  it('surfaces the server sentence when the upload is refused', () => {
    drop(fixture, jpeg());
    http
      .expectOne('/api/v1/files')
      .flush(
        { success: false, message: 'validasi gagal', errors: { file: 'gambar tidak bisa dibaca' } },
        { status: 422, statusText: 'Unprocessable Entity' },
      );
    fixture.detectChanges();

    expect(fixture.componentInstance.upload().errorText()).toBe('gambar tidak bisa dibaca');
    expect(revoked).toContain('blob:test/1');
  });

  it('soft-deletes on remove and tells the parent to clear its file id', () => {
    drop(fixture, jpeg());
    http.expectOne('/api/v1/files').flush(stored());
    fixture.detectChanges();
    http.expectOne('/api/v1/files/8f3c/low').flush(new Blob(['x']));
    fixture.detectChanges();

    fixture.componentInstance.upload().remove();
    const del = http.expectOne('/api/v1/files/8f3c');
    expect(del.request.method).toBe('DELETE');
    del.flush({ success: true, message: 'OK' });
    fixture.detectChanges();

    expect(fixture.componentInstance.clears).toBe(1);
    expect(fixture.componentInstance.upload().current()).toBeNull();
  });

  it('loads an already-stored uuid so an edit form opens filled in', () => {
    fixture.componentInstance.value.set('8f3c');
    fixture.detectChanges();

    http.expectOne('/api/v1/files/8f3c').flush(stored({ status_proses: 'selesai' }));
    fixture.detectChanges();

    expect(fixture.componentInstance.upload().current()?.nama_asli).toBe('IMG_2201.jpg');
    http.expectOne('/api/v1/files/8f3c/low').flush(new Blob(['x']));
  });
});
