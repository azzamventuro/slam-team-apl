// Types for the centralised file layer (mst_file + mst_file_varian).
//
// Mirrors internal/modules/core/file/dto — keep the field names identical to
// the JSON on the wire. No module stores a path of its own: it stores a
// `*_file_id` pointing at mst_file.id, and every URL/variant lives here.

/** The three rendered sizes. `original` is always present; the other two are written asynchronously. */
export type FileVarian = 'original' | 'medium' | 'low';

/** `mst_file.kategori` — decides the disk folder, the accepted MIME types and the default `is_publik`. */
export type FileKategori =
  'foto_profil' | 'foto_formal' | 'absensi' | 'banner' | 'logo' | 'dokumen' | 'flyer' | 'kta';

/** Enum `status_proses_file`: `menunggu` while medium/low are still being written. */
export type StatusProsesFile = 'menunggu' | 'selesai' | 'gagal';

/** The kategori values the backend's `oneof` accepts, in DTO order. */
export const FILE_KATEGORI: readonly FileKategori[] = [
  'foto_profil',
  'foto_formal',
  'absensi',
  'banner',
  'logo',
  'dokumen',
  'flyer',
  'kta',
] as const;

/**
 * Client mirror of the server's `FILE_MAX_UPLOAD_MB` (config default 15). It
 * only saves a round trip — the Go handler caps the body with MaxBytesReader
 * and is the authoritative limit.
 */
export const MAX_UPLOAD_MB = 15;

/** One rendered size on the wire (`dto.VarianResp`). */
export interface FileVarianInfo {
  varian: FileVarian;
  /** Authenticated serve route — works for public and private files alike. */
  url: string;
  /** Only set for public files, which may also be fetched without a token. */
  url_publik?: string | null;
  lebar_px: number | null;
  tinggi_px: number | null;
  ukuran_byte: number;
  mime_type: string;
  kualitas?: number | null;
}

/** `dto.UploadResp` — the body of POST /files and of GET /files/{uuid}. */
export interface UploadedFile {
  /**
   * mst_file.id, which an owning form writes into its `*_file_id` column.
   *
   * SEAM: the Fase 0 API returns only `uuid`, so this is optional today and
   * `FileRef.id` comes back undefined. When the API adds `id` to UploadResp,
   * owning forms get the numeric key without a change here.
   */
  id?: number;
  /** The public identity in every URL — never the numeric id. */
  uuid: string;
  nama_asli: string;
  nama_slug: string;
  ekstensi: string;
  mime_type: string;
  ukuran_byte: number;
  kategori: FileKategori;
  reff_type?: string | null;
  reff_id?: number | null;
  is_publik: boolean;
  status_proses: StatusProsesFile;
  hash_sha256: string;
  lebar_px: number | null;
  tinggi_px: number | null;
  variants: FileVarianInfo[];
  /** The size an owning form should display: medium for images, original otherwise. */
  url: string;
}

/** What an upload component hands back to the parent form control. */
export interface FileRef {
  id?: number;
  uuid: string;
}

/** Non-file parts of the multipart body. All optional except `kategori`. */
export interface UploadOptions {
  reffType?: string | null;
  reffId?: number | null;
  isPublik?: boolean | null;
}

/** True for anything the variant pipeline treats as an image. */
export function isImageMime(mime: string | null | undefined): boolean {
  return !!mime && mime.startsWith('image/');
}

/**
 * Does the file satisfy an `accept` attribute? Handles the three forms a
 * browser accepts: `image/*`, an exact `image/png`, and a `.pdf` extension.
 * An empty accept means "anything".
 */
export function matchesAccept(file: File, accept: string): boolean {
  const rules = accept
    .split(',')
    .map((r) => r.trim().toLowerCase())
    .filter(Boolean);
  if (!rules.length) return true;

  const mime = (file.type || '').toLowerCase();
  const name = file.name.toLowerCase();

  return rules.some((rule) => {
    if (rule.startsWith('.')) return name.endsWith(rule);
    if (rule.endsWith('/*')) return mime.startsWith(rule.slice(0, -1));
    return mime === rule;
  });
}

/**
 * Client-side pre-flight. Returns the i18n key of the problem, or null when the
 * file may be sent. The server validates the type from the first 512 bytes, so
 * this catches the obvious mistakes only — it is UX, not a security boundary.
 */
export function validateFile(file: File, accept: string, maxMb: number): string | null {
  if (!matchesAccept(file, accept)) return 'FILE.WRONG_TYPE';
  if (maxMb > 0 && file.size > maxMb * 1024 * 1024) return 'FILE.TOO_LARGE';
  return null;
}

/** Human-readable byte size for the file chip (`1.8 MB`). */
export function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 KB';
  const mb = bytes / (1024 * 1024);
  if (mb >= 1) return `${mb.toFixed(1)} MB`;
  return `${Math.max(1, Math.round(bytes / 1024))} KB`;
}
