package main

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func init() {
	seedOrder = append(seedOrder, "rbac")
	seeders["rbac"] = seedRBAC
}

// --- Modul rows (19 menu + 1 non-menu "file") ---

type modulRow struct {
	kode     string
	nama     string
	grup     string
	route    string
	icon     string
	urutan   int
	menu     bool
}

var seedModulRows = []modulRow{
	// Master Data
	{kode: "anggota", nama: "Anggota", grup: "Master Data", route: "/anggota", icon: "users", urutan: 1, menu: true},
	{kode: "admin", nama: "Admin", grup: "Master Data", route: "/admin", icon: "shield", urutan: 2, menu: true},
	{kode: "moderator", nama: "Moderator", grup: "Master Data", route: "/moderator", icon: "shield-check", urutan: 3, menu: true},
	{kode: "user", nama: "User", grup: "Master Data", route: "/user", icon: "person", urutan: 4, menu: true},
	{kode: "instansi", nama: "Instansi / Sekolah", grup: "Master Data", route: "/instansi", icon: "building", urutan: 5, menu: true},
	{kode: "unit", nama: "Unit", grup: "Master Data", route: "/unit", icon: "layers", urutan: 6, menu: true},
	{kode: "prestasi", nama: "Prestasi", grup: "Master Data", route: "/prestasi", icon: "trophy", urutan: 7, menu: true},
	{kode: "inorga", nama: "Inorga", grup: "Master Data", route: "/inorga", icon: "award", urutan: 8, menu: true},
	// Konten
	{kode: "kegiatan", nama: "Kegiatan", grup: "Konten", route: "/kegiatan", icon: "calendar", urutan: 9, menu: true},
	{kode: "artikel", nama: "Artikel", grup: "Konten", route: "/artikel", icon: "newspaper", urutan: 10, menu: true},
	{kode: "medsos", nama: "Medsos", grup: "Konten", route: "/medsos", icon: "share", urutan: 11, menu: true},
	{kode: "profile_club", nama: "Profile Club", grup: "Konten", route: "/profile-club", icon: "info-circle", urutan: 12, menu: true},
	// Sistem
	{kode: "dokumen", nama: "Dokumen", grup: "Sistem", route: "/dokumen", icon: "folder", urutan: 13, menu: true},
	{kode: "hak_akses", nama: "Hak Akses", grup: "Sistem", route: "/hak-akses", icon: "key", urutan: 14, menu: true},
	// Operasional
	{kode: "lokasi", nama: "Lokasi", grup: "Operasional", route: "/lokasi", icon: "map-pin", urutan: 15, menu: true},
	{kode: "jadwal", nama: "Jadwal", grup: "Operasional", route: "/jadwal", icon: "calendar-check", urutan: 16, menu: true},
	{kode: "absensi", nama: "Absensi", grup: "Operasional", route: "/absensi", icon: "clipboard-check", urutan: 17, menu: true},
	{kode: "izin", nama: "Izin", grup: "Operasional", route: "/izin", icon: "envelope-paper", urutan: 18, menu: true},
	{kode: "kta", nama: "KTA", grup: "Operasional", route: "/kta", icon: "credit-card", urutan: 19, menu: true},
	// Non-menu (file layer — not shown in sidebar, but needed for file.create/file.delete permissions)
	{kode: "file", nama: "File", grup: "Sistem", route: "", icon: "", urutan: 0, menu: false},
}

// --- Permission rows per modul ---

type permissionSeed struct {
	modulKode   string
	aksi        string
	nama        string
	keterangan  string
	berbahaya   bool
}

// seedPermissions returns ALL permission rows for every module.
// The aksi values must match the Postgres enum `aksi_permission`.
var seedPermissions = []permissionSeed{
	// === Master Data: CRuD (4 actions) ===
	{modulKode: "anggota", aksi: "create", nama: "Buat Anggota"},
	{modulKode: "anggota", aksi: "read", nama: "Lihat Anggota"},
	{modulKode: "anggota", aksi: "update", nama: "Ubah Anggota"},
	{modulKode: "anggota", aksi: "delete", nama: "Hapus Anggota", berbahaya: true},

	{modulKode: "admin", aksi: "create", nama: "Buat Admin"},
	{modulKode: "admin", aksi: "read", nama: "Lihat Admin"},
	{modulKode: "admin", aksi: "update", nama: "Ubah Admin"},
	{modulKode: "admin", aksi: "delete", nama: "Hapus Admin", berbahaya: true},

	{modulKode: "moderator", aksi: "create", nama: "Buat Moderator"},
	{modulKode: "moderator", aksi: "read", nama: "Lihat Moderator"},
	{modulKode: "moderator", aksi: "update", nama: "Ubah Moderator"},
	{modulKode: "moderator", aksi: "delete", nama: "Hapus Moderator", berbahaya: true},

	{modulKode: "user", aksi: "create", nama: "Buat User"},
	{modulKode: "user", aksi: "read", nama: "Lihat User"},
	{modulKode: "user", aksi: "update", nama: "Ubah User"},
	{modulKode: "user", aksi: "delete", nama: "Hapus User", berbahaya: true},

	{modulKode: "instansi", aksi: "create", nama: "Buat Instansi"},
	{modulKode: "instansi", aksi: "read", nama: "Lihat Instansi"},
	{modulKode: "instansi", aksi: "update", nama: "Ubah Instansi"},
	{modulKode: "instansi", aksi: "delete", nama: "Hapus Instansi", berbahaya: true},

	{modulKode: "unit", aksi: "create", nama: "Buat Unit"},
	{modulKode: "unit", aksi: "read", nama: "Lihat Unit"},
	{modulKode: "unit", aksi: "update", nama: "Ubah Unit"},
	{modulKode: "unit", aksi: "delete", nama: "Hapus Unit", berbahaya: true},

	{modulKode: "prestasi", aksi: "create", nama: "Buat Prestasi"},
	{modulKode: "prestasi", aksi: "read", nama: "Lihat Prestasi"},
	{modulKode: "prestasi", aksi: "update", nama: "Ubah Prestasi"},
	{modulKode: "prestasi", aksi: "delete", nama: "Hapus Prestasi", berbahaya: true},

	{modulKode: "inorga", aksi: "create", nama: "Buat Inorga"},
	{modulKode: "inorga", aksi: "read", nama: "Lihat Inorga"},
	{modulKode: "inorga", aksi: "update", nama: "Ubah Inorga"},
	{modulKode: "inorga", aksi: "delete", nama: "Hapus Inorga", berbahaya: true},

	// === Konten: CRuD (4 actions) ===
	{modulKode: "kegiatan", aksi: "create", nama: "Buat Kegiatan"},
	{modulKode: "kegiatan", aksi: "read", nama: "Lihat Kegiatan"},
	{modulKode: "kegiatan", aksi: "update", nama: "Ubah Kegiatan"},
	{modulKode: "kegiatan", aksi: "delete", nama: "Hapus Kegiatan", berbahaya: true},

	{modulKode: "artikel", aksi: "create", nama: "Buat Artikel"},
	{modulKode: "artikel", aksi: "read", nama: "Lihat Artikel"},
	{modulKode: "artikel", aksi: "update", nama: "Ubah Artikel"},
	{modulKode: "artikel", aksi: "delete", nama: "Hapus Artikel", berbahaya: true},

	{modulKode: "medsos", aksi: "create", nama: "Buat Medsos"},
	{modulKode: "medsos", aksi: "read", nama: "Lihat Medsos"},
	{modulKode: "medsos", aksi: "update", nama: "Ubah Medsos"},
	{modulKode: "medsos", aksi: "delete", nama: "Hapus Medsos", berbahaya: true},

	{modulKode: "profile_club", aksi: "create", nama: "Buat Profile Club"},
	{modulKode: "profile_club", aksi: "read", nama: "Lihat Profile Club"},
	{modulKode: "profile_club", aksi: "update", nama: "Ubah Profile Club"},
	{modulKode: "profile_club", aksi: "delete", nama: "Hapus Profile Club", berbahaya: true},

	// === Sistem: CRuD ===
	{modulKode: "dokumen", aksi: "create", nama: "Buat Dokumen"},
	{modulKode: "dokumen", aksi: "read", nama: "Lihat Dokumen"},
	{modulKode: "dokumen", aksi: "update", nama: "Ubah Dokumen"},
	{modulKode: "dokumen", aksi: "delete", nama: "Hapus Dokumen", berbahaya: true},

	{modulKode: "hak_akses", aksi: "create", nama: "Buat Hak Akses", berbahaya: true},
	{modulKode: "hak_akses", aksi: "read", nama: "Lihat Hak Akses", berbahaya: true},
	{modulKode: "hak_akses", aksi: "update", nama: "Ubah Hak Akses", berbahaya: true},
	{modulKode: "hak_akses", aksi: "delete", nama: "Hapus Hak Akses", berbahaya: true},

	// === File: create + delete ===
	{modulKode: "file", aksi: "create", nama: "Unggah Berkas"},
	{modulKode: "file", aksi: "delete", nama: "Hapus Berkas", berbahaya: true},

	// === Operasional: Lokasi CRuD ===
	{modulKode: "lokasi", aksi: "create", nama: "Buat Lokasi"},
	{modulKode: "lokasi", aksi: "read", nama: "Lihat Lokasi"},
	{modulKode: "lokasi", aksi: "update", nama: "Ubah Lokasi"},
	{modulKode: "lokasi", aksi: "delete", nama: "Hapus Lokasi", berbahaya: true},

	// === Jadwal: CRuD + assign + batal_sesi ===
	{modulKode: "jadwal", aksi: "create", nama: "Buat Jadwal"},
	{modulKode: "jadwal", aksi: "read", nama: "Lihat Jadwal"},
	{modulKode: "jadwal", aksi: "update", nama: "Ubah Jadwal"},
	{modulKode: "jadwal", aksi: "delete", nama: "Hapus Jadwal", berbahaya: true},
	{modulKode: "jadwal", aksi: "assign", nama: "Tugaskan Jadwal"},
	{modulKode: "jadwal", aksi: "batal_sesi", nama: "Batalkan Sesi Jadwal", berbahaya: true},

	// === Absensi: create, read, override, delete, export ===
	{modulKode: "absensi", aksi: "create", nama: "Lakukan Absensi"},
	{modulKode: "absensi", aksi: "read", nama: "Lihat Absensi"},
	{modulKode: "absensi", aksi: "override", nama: "Override Absensi", berbahaya: true},
	{modulKode: "absensi", aksi: "delete", nama: "Hapus Absensi", berbahaya: true},
	{modulKode: "absensi", aksi: "export", nama: "Ekspor Absensi"},

	// === Izin: create, read, approve ===
	{modulKode: "izin", aksi: "create", nama: "Ajukan Izin"},
	{modulKode: "izin", aksi: "read", nama: "Lihat Izin"},
	{modulKode: "izin", aksi: "approve", nama: "Setujui Izin"},

	// === KTA: create, read, print, cabut ===
	{modulKode: "kta", aksi: "create", nama: "Terbitkan KTA"},
	{modulKode: "kta", aksi: "read", nama: "Lihat KTA"},
	{modulKode: "kta", aksi: "print", nama: "Cetak KTA"},
	{modulKode: "kta", aksi: "cabut", nama: "Cabut KTA", berbahaya: true},
}

// --- Role rows (5 built-in) ---

type roleRow struct {
	kode   string
	nama   string
	level  int
	super  bool
	sistem bool
}

var seedRoleRows = []roleRow{
	{kode: "super_admin", nama: "Super Admin", level: 0, super: true, sistem: true},
	{kode: "admin", nama: "Admin", level: 10, super: false, sistem: true},
	{kode: "moderator", nama: "Moderator", level: 20, super: false, sistem: true},
	{kode: "user", nama: "User", level: 30, super: false, sistem: true},
	{kode: "guest", nama: "Guest", level: 99, super: false, sistem: true},
}

// --- Permission matrix (Bab 3.5) ---
// Maps role kode → set of permission kode that role receives.
// Super Admin is intentionally NOT seeded (is_super bypass).

type matrixEntry struct {
	role    string
	perm    string
	cakupan string // "" means "semua" (default)
}

// buildMatrix returns the full Bab 3.5 matrix as a list of entries.
func buildMatrix() []matrixEntry {
	var m []matrixEntry

	// Helper: add an entry for all listed roles with cakupan "semua".
	perm := func(kode string, roles ...string) {
		for _, r := range roles {
			m = append(m, matrixEntry{role: r, perm: kode, cakupan: "semua"})
		}
	}
	// Helper: add with explicit cakupan.
	permScope := func(kode, cakupan string, roles ...string) {
		for _, r := range roles {
			m = append(m, matrixEntry{role: r, perm: kode, cakupan: cakupan})
		}
	}

	// ── Tabel 1: Master Data & Konten ──

	// anggota: SA(bypass) AdminCRUD ModCRUD UserR Guest-
	perm("anggota.create", "admin", "moderator")
	perm("anggota.read", "admin", "moderator", "user")
	perm("anggota.update", "admin", "moderator")
	perm("anggota.delete", "admin")

	// admin: SA(bypass) AdminR ModR UserR Guest-
	perm("admin.read", "admin", "moderator", "user")

	// moderator: SA(bypass) AdminCRUD ModR UserR Guest-
	perm("moderator.create", "admin")
	perm("moderator.read", "admin", "moderator", "user")
	perm("moderator.update", "admin")
	perm("moderator.delete", "admin")

	// user: SA(bypass) AdminCRUD ModCRU UserR Guest-
	perm("user.create", "admin", "moderator")
	perm("user.read", "admin", "moderator", "user")
	perm("user.update", "admin", "moderator")
	perm("user.delete", "admin")

	// instansi: SA(bypass) AdminCRUD ModR UserR Guest-
	perm("instansi.create", "admin")
	perm("instansi.read", "admin", "moderator", "user")
	perm("instansi.update", "admin")
	perm("instansi.delete", "admin")

	// unit: SA(bypass) AdminCRUD ModCRUD UserCRUD Guest-
	perm("unit.create", "admin", "moderator", "user")
	perm("unit.read", "admin", "moderator", "user")
	perm("unit.update", "admin", "moderator", "user")
	perm("unit.delete", "admin", "moderator", "user")

	// prestasi: SA(bypass) AdminCRUD ModCRUD UserR GuestR
	perm("prestasi.create", "admin", "moderator")
	perm("prestasi.read", "admin", "moderator", "user", "guest")
	perm("prestasi.update", "admin", "moderator")
	perm("prestasi.delete", "admin")

	// inorga: SA(bypass) AdminCRUD ModR UserR GuestR
	perm("inorga.create", "admin")
	perm("inorga.read", "admin", "moderator", "user", "guest")
	perm("inorga.update", "admin")
	perm("inorga.delete", "admin")

	// kegiatan: SA(bypass) AdminCRUD ModCRUD UserR GuestR
	perm("kegiatan.create", "admin", "moderator")
	perm("kegiatan.read", "admin", "moderator", "user", "guest")
	perm("kegiatan.update", "admin", "moderator")
	perm("kegiatan.delete", "admin")

	// artikel: SA(bypass) AdminCRUD ModCRUD UserR GuestR
	perm("artikel.create", "admin", "moderator")
	perm("artikel.read", "admin", "moderator", "user", "guest")
	perm("artikel.update", "admin", "moderator")
	perm("artikel.delete", "admin")

	// medsos: SA(bypass) AdminCRUD ModCRUD UserCRUD GuestR
	perm("medsos.create", "admin", "moderator", "user")
	perm("medsos.read", "admin", "moderator", "user", "guest")
	perm("medsos.update", "admin", "moderator", "user")
	perm("medsos.delete", "admin")

	// profile_club: SA(bypass) AdminCRUD ModR UserR GuestR
	perm("profile_club.create", "admin")
	perm("profile_club.read", "admin", "moderator", "user", "guest")
	perm("profile_club.update", "admin")
	perm("profile_club.delete", "admin")

	// dokumen: SA(bypass) AdminCRUD ModCRUD UserR Guest-
	perm("dokumen.create", "admin", "moderator")
	perm("dokumen.read", "admin", "moderator", "user")
	perm("dokumen.update", "admin", "moderator")
	perm("dokumen.delete", "admin")

	// ── Tabel 2: Operasional & Sistem ──

	// lokasi: SA(bypass) AdminCRUD ModR UserR Guest-
	perm("lokasi.create", "admin")
	perm("lokasi.read", "admin", "moderator", "user")
	perm("lokasi.update", "admin")
	perm("lokasi.delete", "admin")

	// jadwal: SA(bypass) AdminCRUD ModCRUD UserR Guest-
	perm("jadwal.create", "admin", "moderator")
	perm("jadwal.read", "admin", "moderator", "user")
	perm("jadwal.update", "admin", "moderator")
	perm("jadwal.delete", "admin")
	perm("jadwal.assign", "admin", "moderator")
	perm("jadwal.batal_sesi", "admin", "moderator")

	// absensi: special — read has cakupan for User
	perm("absensi.create", "admin", "moderator", "user")
	permScope("absensi.read", "semua", "admin", "moderator")
	permScope("absensi.read", "milik_sendiri", "user")
	perm("absensi.override", "admin")
	perm("absensi.delete", "admin")
	perm("absensi.export", "admin", "moderator")

	// izin: create (all logged), read (user=milik_sendiri), approve
	perm("izin.create", "admin", "moderator", "user")
	permScope("izin.read", "semua", "admin", "moderator")
	permScope("izin.read", "milik_sendiri", "user")
	perm("izin.approve", "admin", "moderator")

	// kta: SA(bypass) Admin create+read+print+cabut
	perm("kta.create", "admin")
	perm("kta.read", "admin")
	perm("kta.print", "admin")
	perm("kta.cabut", "admin")

	// file: create (all logged), delete (Mod+User = milik_sendiri)
	perm("file.create", "admin", "moderator", "user")
	permScope("file.delete", "semua", "admin")
	permScope("file.delete", "milik_sendiri", "moderator", "user")

	// hak_akses: SA only (SA bypass, so no rows needed for SA)

	return m
}

// seedRBAC is the idempotent seeder for all RBAC tables.
func seedRBAC(ctx context.Context, db *gorm.DB) error {
	// Check prerequisite tables exist.
	for _, tbl := range []string{"mst_modul", "mst_role"} {
		var count int64
		if err := db.WithContext(ctx).Raw(
			"SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?", tbl,
		).Scan(&count).Error; err != nil {
			return fmt.Errorf("periksa tabel %s: %w", tbl, err)
		}
		if count == 0 {
			return fmt.Errorf("tabel %s belum ada — jalankan `slamctl migrate up` dulu", tbl)
		}
	}

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Seed mst_modul
		if err := seedModul(tx); err != nil {
			return fmt.Errorf("seed modul: %w", err)
		}

		// 2. Seed mst_permission (needs modul IDs)
		modulID, err := loadModulIDs(tx)
		if err != nil {
			return fmt.Errorf("load modul ids: %w", err)
		}
		if err := seedPermission(tx, modulID); err != nil {
			return fmt.Errorf("seed permission: %w", err)
		}

		// 3. Seed mst_role
		if err := seedRole(tx); err != nil {
			return fmt.Errorf("seed role: %w", err)
		}

		// 4. Seed role_permission matrix
		permID, err := loadPermIDs(tx)
		if err != nil {
			return fmt.Errorf("load permission ids: %w", err)
		}
		roleID, err := loadRoleIDs(tx)
		if err != nil {
			return fmt.Errorf("load role ids: %w", err)
		}
		if err := seedMatrix(tx, permID, roleID); err != nil {
			return fmt.Errorf("seed matrix: %w", err)
		}

		// 5. Set rbac.perm_version = 1 if not set
		if err := ensurePermVersion(tx); err != nil {
			return fmt.Errorf("set perm_version: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	total := len(seedModulRows) + len(seedPermissions) + len(seedRoleRows)
	fmt.Printf("(%d modul, %d permission, %d role, %d matrix rows) ", total, len(seedPermissions), len(seedRoleRows), len(buildMatrix()))
	return nil
}

func seedModul(tx *gorm.DB) error {
	const q = `
		INSERT INTO mst_modul (kode, nama, grup, route, icon, urutan, tampil_di_menu)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (kode) DO UPDATE SET
			nama           = EXCLUDED.nama,
			grup           = EXCLUDED.grup,
			route          = EXCLUDED.route,
			icon           = EXCLUDED.icon,
			urutan         = EXCLUDED.urutan,
			tampil_di_menu = EXCLUDED.tampil_di_menu`
	for _, r := range seedModulRows {
		if err := tx.Exec(q, r.kode, r.nama, r.grup, r.route, r.icon, r.urutan, r.menu).Error; err != nil {
			return fmt.Errorf("modul %q: %w", r.kode, err)
		}
	}
	return nil
}

func loadModulIDs(tx *gorm.DB) (map[string]int64, error) {
	type row struct {
		Kode string
		ID   int64
	}
	var rows []row
	if err := tx.Raw("SELECT kode, id FROM mst_modul").Scan(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[string]int64, len(rows))
	for _, r := range rows {
		m[r.Kode] = r.ID
	}
	return m, nil
}

func seedPermission(tx *gorm.DB, modulID map[string]int64) error {
	const q = `
		INSERT INTO mst_permission (modul_id, aksi, kode, nama, keterangan, is_berbahaya)
		VALUES (?, CAST(? AS aksi_permission), ?, ?, NULLIF(?, ''), ?)
		ON CONFLICT (kode) DO UPDATE SET
			modul_id      = EXCLUDED.modul_id,
			aksi          = EXCLUDED.aksi,
			nama          = EXCLUDED.nama,
			keterangan    = EXCLUDED.keterangan,
			is_berbahaya  = EXCLUDED.is_berbahaya`
	for _, p := range seedPermissions {
		mid, ok := modulID[p.modulKode]
		if !ok {
			return fmt.Errorf("modul %q tidak ditemukan di mst_modul", p.modulKode)
		}
		kode := p.modulKode + "." + p.aksi
		if err := tx.Exec(q, mid, p.aksi, kode, p.nama, p.keterangan, p.berbahaya).Error; err != nil {
			return fmt.Errorf("permission %q: %w", kode, err)
		}
	}
	return nil
}

func seedRole(tx *gorm.DB) error {
	const q = `
		INSERT INTO mst_role (kode, nama, level, is_super, is_sistem)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (kode) DO UPDATE SET
			nama      = EXCLUDED.nama,
			level     = EXCLUDED.level,
			is_super  = EXCLUDED.is_super,
			is_sistem = EXCLUDED.is_sistem`
	for _, r := range seedRoleRows {
		if err := tx.Exec(q, r.kode, r.nama, r.level, r.super, r.sistem).Error; err != nil {
			return fmt.Errorf("role %q: %w", r.kode, err)
		}
	}
	return nil
}

func loadPermIDs(tx *gorm.DB) (map[string]int64, error) {
	type row struct {
		Kode string
		ID   int64
	}
	var rows []row
	if err := tx.Raw("SELECT kode, id FROM mst_permission").Scan(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[string]int64, len(rows))
	for _, r := range rows {
		m[r.Kode] = r.ID
	}
	return m, nil
}

func loadRoleIDs(tx *gorm.DB) (map[string]int64, error) {
	type row struct {
		Kode string
		ID   int64
	}
	var rows []row
	if err := tx.Raw("SELECT kode, id FROM mst_role").Scan(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[string]int64, len(rows))
	for _, r := range rows {
		m[r.Kode] = r.ID
	}
	return m, nil
}

func seedMatrix(tx *gorm.DB, permID, roleID map[string]int64) error {
	const q = `
		INSERT INTO role_permission (role_id, permission_id, cakupan)
		VALUES (?, ?, CAST(? AS cakupan_permission))
		ON CONFLICT (role_id, permission_id) DO UPDATE SET
			cakupan = EXCLUDED.cakupan`
	for _, e := range buildMatrix() {
		rid, ok := roleID[e.role]
		if !ok {
			return fmt.Errorf("role %q tidak ditemukan", e.role)
		}
		pid, ok := permID[e.perm]
		if !ok {
			return fmt.Errorf("permission %q tidak ditemukan", e.perm)
		}
		cakupan := e.cakupan
		if cakupan == "" {
			cakupan = "semua"
		}
		if err := tx.Exec(q, rid, pid, cakupan).Error; err != nil {
			return fmt.Errorf("matrix %s×%s: %w", e.role, e.perm, err)
		}
	}
	return nil
}

func ensurePermVersion(tx *gorm.DB) error {
	// Only set if not already present — don't overwrite an existing counter.
	var exists bool
	if err := tx.Raw(
		"SELECT EXISTS(SELECT 1 FROM mst_pengaturan WHERE kunci = 'rbac.perm_version')",
	).Scan(&exists).Error; err != nil {
		return err
	}
	if exists {
		return nil
	}
	// Insert a new row for perm_version tracking.
	const q = `
		INSERT INTO mst_pengaturan (kunci, grup, label, nilai, tipe_nilai, nilai_bawaan, urutan, is_terkunci)
		VALUES ('rbac.perm_version', 'rbac', 'Permission version counter', '1', 'integer', '1', 99, true)
		ON CONFLICT (kunci) DO NOTHING`
	return tx.Exec(q).Error
}
