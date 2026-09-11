export interface Inorga {
  id: number;
  kode: string | null;
  nama: string;
  tanggal_mulai: string | null;
  tanggal_selesai: string | null;
  konten: string | null;
  aktif: boolean;
  logo: FileRef | null;
  banner: FileRef | null;
  file_sk: FileRef | null;
  created_at: string;
  modified_at: string | null;
}

export interface FileRef {
  id: number;
  uuid: string;
  nama_asli?: string;
}

export interface InorgaForm {
  kode?: string | null;
  nama: string;
  tanggal_mulai?: string | null;
  tanggal_selesai?: string | null;
  logo_file_id?: number | null;
  banner_file_id?: number | null;
  file_sk_file_id?: number | null;
  konten?: string | null;
}

export interface InorgaQuery {
  page: number;
  per_page: number;
  q: string;
  sort?: string;
  aktif?: string | null;
}
