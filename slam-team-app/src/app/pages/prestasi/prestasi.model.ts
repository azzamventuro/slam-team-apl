export interface Prestasi {
  id: number;
  anggota_id: number;
  anggota: AnggotaRef;
  kode: string | null;
  peringkat: string | null;
  tingkat: string | null;
  judul_kompetisi: string;
  flyer: FileRef | null;
  tanggal_kompetisi: string | null;
  alamat_kompetisi: string | null;
  foto_sampul: FileRef | null;
  keterangan: string | null;
  created_at: string;
  modified_at: string | null;
}

export interface AnggotaRef {
  id: number;
  nama_lengkap: string;
  nama_panggilan?: string | null;
  no_induk?: string | null;
}

export interface FileRef {
  file_id: number;
  uuid: string;
  is_publik: boolean;
}

export interface PrestasiForm {
  anggota_id: number;
  kode?: string | null;
  peringkat?: string | null;
  tingkat?: string | null;
  judul_kompetisi: string;
  flyer_file_id?: number | null;
  tanggal_kompetisi?: string | null;
  alamat_kompetisi?: string | null;
  foto_sampul_file_id?: number | null;
  keterangan?: string | null;
}

export interface PrestasiQuery {
  page: number;
  per_page: number;
  q: string;
  sort?: string;
  anggota_id?: number | null;
  tingkat?: string | null;
  peringkat?: string | null;
  tanggal_dari?: string | null;
  tanggal_sampai?: string | null;
}

export interface AnggotaOpsi {
  id: number;
  nama_lengkap: string;
  no_induk: string;
}
