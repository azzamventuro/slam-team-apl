export interface Medsos {
  id: number;
  anggota_id: number;
  kode: string | null;
  tipe: number;
  icon: string | null;
  jenis_medsos: string;
  konten_medsos: string;
  created_at: string;
  created_by: number | null;
  modified_at: string | null;
  modified_by: number | null;
}

export interface MedsosForm {
  anggota_id: number;
  jenis_medsos: string;
  konten_medsos: string;
  icon?: string | null;
  kode?: string | null;
  tipe?: number | null;
}

export interface MedsosQuery {
  page: number;
  per_page: number;
  q: string;
  sort?: string;
  anggota_id?: number | null;
  jenis_medsos?: string | null;
}

export interface AnggotaOpsi {
  id: number;
  nama_lengkap: string;
}
