import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { ApiResponse } from '../../core/models/api.model';
import { UploadedFile, matchesAccept, validateFile } from './file.model';
import { FileService, fileErrorText } from './file.service';

function stored(over: Partial<UploadedFile> = {}): UploadedFile {
  return {
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
  };
}

function envelope(data: UploadedFile): ApiResponse<UploadedFile> {
  return { success: true, message: 'created', data };
}

describe('FileService', () => {
  let svc: FileService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    svc = TestBed.inject(FileService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  /** Bab 6.6: the browser must generate the multipart boundary itself. */
  it('posts FormData without a manual Content-Type', () => {
    const file = new File(['bytes'], 'IMG_2201.jpg', { type: 'image/jpeg' });
    let got: UploadedFile | undefined;
    svc.upload('file', file, { kategori: 'foto_profil' }).subscribe((f) => (got = f));

    const req = http.expectOne('/api/v1/files');
    expect(req.request.method).toBe('POST');
    expect(req.request.headers.has('Content-Type')).toBe(false);
    expect(req.request.body instanceof FormData).toBe(true);

    const body = req.request.body as FormData;
    expect(body.get('kategori')).toBe('foto_profil');
    expect((body.get('file') as File).name).toBe('IMG_2201.jpg');

    req.flush(envelope(stored()));
    expect(got?.uuid).toBe('8f3c');
  });

  it('maps upload options onto the multipart fields and drops the absent ones', () => {
    const file = new File(['bytes'], 'a.jpg', { type: 'image/jpeg' });
    svc.uploadFor('absensi', file, { reffType: 'absensi', reffId: 12 }).subscribe();

    const body = http.expectOne('/api/v1/files').request.body as FormData;
    expect(body.get('kategori')).toBe('absensi');
    expect(body.get('reff_type')).toBe('absensi');
    expect(body.get('reff_id')).toBe('12');
    // Absent is_publik means "let the server derive it from the kategori" —
    // it must not arrive as the string "null".
    expect(body.has('is_publik')).toBe(false);

    http.expectNone('/api/v1/files/8f3c');
  });

  it('sends is_publik when it is explicitly false', () => {
    const file = new File(['bytes'], 'a.jpg', { type: 'image/jpeg' });
    svc.uploadFor('banner', file, { isPublik: false }).subscribe();

    const body = http.expectOne('/api/v1/files').request.body as FormData;
    expect(body.get('is_publik')).toBe('false');
  });

  it('unwraps the envelope and surfaces the stored uuid', () => {
    let got: UploadedFile | undefined;
    svc.detail('8f3c').subscribe((f) => (got = f));
    http.expectOne('/api/v1/files/8f3c').flush(envelope(stored({ status_proses: 'selesai' })));
    expect(got?.status_proses).toBe('selesai');
  });

  it('soft-deletes through DELETE /files/:uuid', () => {
    let done = false;
    svc.remove('8f3c').subscribe(() => (done = true));
    const req = http.expectOne('/api/v1/files/8f3c');
    expect(req.request.method).toBe('DELETE');
    req.flush({ success: true, message: 'OK' });
    expect(done).toBe(true);
  });
});

describe('file pre-flight helpers', () => {
  const jpeg = new File(['x'], 'a.jpg', { type: 'image/jpeg' });
  const pdf = new File(['x'], 'surat.pdf', { type: 'application/pdf' });

  it('matches wildcard, exact and extension accept rules', () => {
    expect(matchesAccept(jpeg, 'image/*')).toBe(true);
    expect(matchesAccept(pdf, 'image/*')).toBe(false);
    expect(matchesAccept(pdf, '.pdf')).toBe(true);
    expect(matchesAccept(jpeg, 'image/png')).toBe(false);
    expect(matchesAccept(pdf, '')).toBe(true);
  });

  it('names the problem with an i18n key, or null when the file may be sent', () => {
    expect(validateFile(jpeg, 'image/*', 15)).toBeNull();
    expect(validateFile(pdf, 'image/*', 15)).toBe('FILE.WRONG_TYPE');

    const big = new File(['x'], 'big.jpg', { type: 'image/jpeg' });
    Object.defineProperty(big, 'size', { value: 16 * 1024 * 1024 });
    expect(validateFile(big, 'image/*', 15)).toBe('FILE.TOO_LARGE');
    expect(validateFile(big, 'image/*', 0)).toBeNull();
  });
});

describe('fileErrorText', () => {
  it('prefers the first field error, then the envelope message', () => {
    expect(fileErrorText({ error: { errors: { file: 'berkas wajib diunggah' } } })).toBe(
      'berkas wajib diunggah',
    );
    expect(fileErrorText({ error: { errors: { file: ['terlalu besar'] } } })).toBe('terlalu besar');
    expect(fileErrorText({ error: { errors: 'ditolak' } })).toBe('ditolak');
    expect(fileErrorText({ error: { message: 'validasi gagal', errors: null } })).toBe(
      'validasi gagal',
    );
    expect(fileErrorText(new Error('boom'))).toBe('');
  });
});
