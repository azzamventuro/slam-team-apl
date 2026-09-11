/** Mirrors backend dto.UserResp, AnggotaInfo, RoleInfo, RoleOpsi, AnggotaOpsi. */

export interface AnggotaInfo {
  id: number;
  nama_lengkap: string;
  no_induk?: string;
  instansi_id: number;
  foto_url?: string;
}

export interface RoleInfo {
  id: number;
  nama: string;
  level: number;
  is_super: boolean;
}

export interface UserRow {
  id: number;
  anggota: AnggotaInfo;
  username: string;
  email: string;
  role: RoleInfo;
  timezone: string;
  is_aktif: boolean;
  login_terakhir?: string;
  terkunci_sampai?: string;
  created_at: string;
  modified_at?: string;
}

export interface UserForm {
  anggota_id: number;
  username: string;
  email: string;
  password?: string;
  role_id: number;
  timezone?: string;
  is_aktif?: boolean;
  buka_kunci?: boolean;
}

export interface RoleOpsi {
  id: number;
  nama: string;
  level: number;
}

export interface AnggotaOpsi {
  id: number;
  nama_lengkap: string;
  no_induk?: string;
}

export interface UserListQuery {
  page?: number;
  per_page?: number;
  q?: string;
  sort?: string;
  is_aktif?: boolean;
  instansi_id?: number;
}
