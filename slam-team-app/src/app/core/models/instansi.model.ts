export interface Instansi {
  id: number;
  kode: string;
  nama: string;
  nama_club: string;
  alamat: string;
  no_telepon: string;
  logo_utama_file_id: number | null;
  logo_utama_uuid: string | null;
  logo_tambahan_file_id: number | null;
  logo_tambahan_uuid: string | null;
  tanggal_bergabung: string | null;
  status: number;
  jumlah_anggota: number;
  created_at: string;
  modified_at: string | null;
}

export interface InstansiForm {
  kode: string;
  nama: string;
  nama_club: string;
  alamat: string;
  no_telepon: string;
  logo_utama_file_id: number | null;
  logo_tambahan_file_id: number | null;
  tanggal_bergabung: string | null;
  status: number;
}

export interface InstansiQuery {
  page: number;
  per_page: number;
  q: string;
  sort?: string;
  status?: number | null;
}
