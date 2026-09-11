export interface Anggota {
  id: number;
  instansi_id: number;
  instansi_nama: string | null;
  wilayah_id: number | null;
  wilayah_nama: string | null;
  no_induk: string | null;
  nama_lengkap: string;
  nama_panggilan: string | null;
  foto_profil_file_id: number | null;
  foto_profil_uuid: string | null;
  foto_formal_file_id: number | null;
  foto_formal_uuid: string | null;
  jenis_anggota: 'siswa' | 'dewasa' | 'siswa_ke_dewasa';
  jenis_kelamin: number | null;
  jenis_identitas: number | null;
  no_identitas: string | null;
  file_identitas_file_id: number | null;
  file_identitas_uuid: string | null;
  pekerjaan: string | null;
  alamat: string | null;
  kode_pos: string | null;
  tempat_lahir: string | null;
  tanggal_lahir: string;
  tanggal_bergabung: string | null;
  status_anggota: 'aktif' | 'non_aktif';
  created_at: string;
  modified_at: string | null;
}

export interface AnggotaForm {
  instansi_id: number;
  wilayah_id?: number | null;
  nama_lengkap: string;
  nama_panggilan?: string | null;
  jenis_anggota: string;
  jenis_kelamin?: number | null;
  jenis_identitas?: number | null;
  no_identitas?: string | null;
  pekerjaan?: string | null;
  alamat?: string | null;
  kode_pos?: string | null;
  tempat_lahir?: string | null;
  tanggal_lahir: string;
  tanggal_bergabung?: string | null;
  status_anggota?: string;
  foto_profil_file_id?: number | null;
  foto_formal_file_id?: number | null;
  file_identitas_file_id?: number | null;
}

export interface AnggotaQuery {
  page: number;
  per_page: number;
  q: string;
  sort?: string;
  instansi_id?: number | null;
  jenis?: string | null;
  status?: string | null;
}
