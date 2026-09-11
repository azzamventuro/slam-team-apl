package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"slam-team-api/internal/shared/timeutil"
	"slam-team-api/pkg/utils"

	"gorm.io/gorm"
)

// passwordEnv lets an operator pass the password without putting it in shell
// history or the process list.
const passwordEnv = "SLAMCTL_SUPERADMIN_PASSWORD"

// requiredTables must exist before the bootstrap can run; they are created by
// the RBAC and identity migrations, not by this command.
var requiredTables = []string{"mst_role", "mst_instansi", "anggota", "users"}

type superadminFlags struct {
	nama         string
	username     string
	email        string
	password     string
	tanggalLahir string
	instansiID   int64
	timezone     string
}

// cmdCreateSuperadmin bootstraps the very first super admin. It exists only as
// a CLI command — never as an HTTP endpoint — because it is the one write that
// cannot be authorised by an existing permission holder (chicken-and-egg). It
// refuses to run once any super admin exists, so it can never become a
// post-launch backdoor.
func cmdCreateSuperadmin(args []string) error {
	f, err := parseSuperadminFlags(args)
	if err != nil {
		return err
	}

	cfg, db, err := openDB()
	if err != nil {
		return err
	}
	defer closeDB(db)

	if f.timezone == "" {
		f.timezone = cfg.App.Timezone
	}
	if _, tzErr := time.LoadLocation(f.timezone); tzErr != nil {
		return fmt.Errorf("zona waktu %q tidak dikenal (contoh: %s)", f.timezone, timeutil.DefaultZone)
	}

	for _, t := range requiredTables {
		ok, err := tableExists(db, t)
		if err != nil {
			return fmt.Errorf("periksa tabel %s: %w", t, err)
		}
		if !ok {
			return fmt.Errorf("tabel %q belum ada — jalankan `slamctl migrate up` dulu", t)
		}
	}

	// Guard first, before anything is written.
	var existing int64
	if err := db.Raw(`
		SELECT count(*) FROM users u
		JOIN mst_role r ON r.id = u.role_id
		WHERE r.is_super = true AND r.is_deleted = false AND u.is_deleted = false
	`).Scan(&existing).Error; err != nil {
		return fmt.Errorf("hitung super admin yang ada: %w", err)
	}
	if existing > 0 {
		return fmt.Errorf("super admin sudah ada (%d akun) — perintah dibatalkan; "+
			"akun berikutnya dibuat lewat panel admin oleh pemegang izin user.create", existing)
	}

	var roleID int64
	err = db.Raw(`
		SELECT id FROM mst_role
		WHERE is_super = true AND is_deleted = false AND is_aktif = true
		ORDER BY level ASC, id ASC LIMIT 1
	`).Scan(&roleID).Error
	if err != nil {
		return fmt.Errorf("cari peran super admin: %w", err)
	}
	if roleID == 0 {
		return errors.New("peran super admin (mst_role.is_super = true) belum ada — jalankan `slamctl seed` dulu")
	}

	instansiID := f.instansiID
	if instansiID == 0 {
		if err := db.Raw(`SELECT id FROM mst_instansi WHERE is_deleted = false ORDER BY id LIMIT 1`).
			Scan(&instansiID).Error; err != nil {
			return fmt.Errorf("cari instansi: %w", err)
		}
		if instansiID == 0 {
			return errors.New("belum ada baris mst_instansi (anggota.instansi_id NOT NULL) — " +
				"jalankan `slamctl seed` dulu atau berikan --instansi-id")
		}
	} else {
		var found int64
		if err := db.Raw(`SELECT count(*) FROM mst_instansi WHERE id = ? AND is_deleted = false`, instansiID).
			Scan(&found).Error; err != nil {
			return fmt.Errorf("periksa instansi: %w", err)
		}
		if found == 0 {
			return fmt.Errorf("instansi id=%d tidak ditemukan", instansiID)
		}
	}

	var taken int64
	if err := db.Raw(`
		SELECT count(*) FROM users
		WHERE is_deleted = false AND (lower(username) = lower(?) OR lower(email) = lower(?))
	`, f.username, f.email).Scan(&taken).Error; err != nil {
		return fmt.Errorf("periksa username/email: %w", err)
	}
	if taken > 0 {
		return fmt.Errorf("username %q atau email %q sudah dipakai", f.username, f.email)
	}

	hash, err := utils.HashPassword(f.password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	var userID, anggotaID int64
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(`
			INSERT INTO anggota (instansi_id, nama_lengkap, jenis_anggota, tanggal_lahir,
			                     status_anggota, tanggal_bergabung, created_at)
			VALUES (?, ?, CAST(? AS jenis_anggota), CAST(? AS date),
			        CAST(? AS status_anggota), CURRENT_DATE, now())
			RETURNING id
		`, instansiID, f.nama, "dewasa", f.tanggalLahir, "aktif").Scan(&anggotaID).Error; err != nil {
			return fmt.Errorf("insert anggota: %w", err)
		}

		if err := tx.Raw(`
			INSERT INTO users (anggota_id, username, email, password, role_id,
			                   timezone, is_aktif, created_at)
			VALUES (?, ?, ?, ?, ?, ?, true, now())
			RETURNING id
		`, anggotaID, f.username, f.email, hash, roleID, f.timezone).Scan(&userID).Error; err != nil {
			return fmt.Errorf("insert users: %w", err)
		}

		// Link user to role in the user_role junction table so that
		// GetUserPrimaryRole (used by JWT issue) finds the role assignment.
		if err := tx.Exec(`
			INSERT INTO user_role (user_id, role_id, is_utama, created_at)
			VALUES (?, ?, true, now())
		`, userID, roleID).Error; err != nil {
			return fmt.Errorf("insert user_role: %w", err)
		}

		// Audit trail, best effort: log_aktivitas belongs to a later migration.
		ok, err := tableExists(tx, "log_aktivitas")
		if err != nil {
			return fmt.Errorf("periksa log_aktivitas: %w", err)
		}
		if ok {
			// aktor_user_id stays NULL: this is a system/CLI action, not one
			// performed by a signed-in user.
			if err := tx.Exec(`
				INSERT INTO log_aktivitas (aktor_user_id, modul, aksi, reff_type, reff_id, ringkasan, created_at)
				VALUES (NULL, 'sistem', 'buat', 'users', ?, 'Membuat super admin pertama', now())
			`, userID).Error; err != nil {
				return fmt.Errorf("insert log_aktivitas: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("super admin dibuat: users.id=%d anggota.id=%d username=%s role_id=%d instansi_id=%d\n",
		userID, anggotaID, f.username, roleID, instansiID)
	return nil
}

func parseSuperadminFlags(args []string) (*superadminFlags, error) {
	f := &superadminFlags{}
	fs := flag.NewFlagSet("create-superadmin", flag.ContinueOnError)
	fs.StringVar(&f.nama, "nama", "", "nama lengkap (wajib)")
	fs.StringVar(&f.username, "username", "", "username login, unik (wajib)")
	fs.StringVar(&f.email, "email", "", "email login, unik (wajib)")
	fs.StringVar(&f.password, "password", "", "password; kosongkan dan pakai env "+passwordEnv)
	fs.StringVar(&f.tanggalLahir, "tanggal-lahir", "", "tanggal lahir YYYY-MM-DD (wajib)")
	fs.Int64Var(&f.instansiID, "instansi-id", 0, "id mst_instansi; bawaan: instansi pertama")
	fs.StringVar(&f.timezone, "timezone", "", "zona IANA user; bawaan: APP_TIMEZONE")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if f.password == "" {
		f.password = os.Getenv(passwordEnv)
	}

	f.nama = strings.TrimSpace(f.nama)
	f.username = strings.TrimSpace(f.username)
	f.email = strings.TrimSpace(f.email)

	var missing []string
	if f.nama == "" {
		missing = append(missing, "--nama")
	}
	if f.username == "" {
		missing = append(missing, "--username")
	}
	if f.email == "" {
		missing = append(missing, "--email")
	}
	if f.tanggalLahir == "" {
		missing = append(missing, "--tanggal-lahir")
	}
	if f.password == "" {
		missing = append(missing, "--password (atau env "+passwordEnv+")")
	}
	if len(missing) > 0 {
		fs.SetOutput(os.Stderr)
		fs.Usage()
		return nil, fmt.Errorf("flag wajib belum diisi: %s", strings.Join(missing, ", "))
	}

	if len(f.username) < 3 || len(f.username) > 50 || strings.ContainsAny(f.username, " \t") {
		return nil, errors.New("username harus 3–50 karakter tanpa spasi")
	}
	if !strings.Contains(f.email, "@") || len(f.email) > 150 {
		return nil, errors.New("email tidak valid")
	}
	if len(f.nama) > 150 {
		return nil, errors.New("nama lengkap maksimal 150 karakter")
	}
	if len([]rune(f.password)) < 8 {
		return nil, errors.New("password minimal 8 karakter")
	}
	if _, err := time.Parse("2006-01-02", f.tanggalLahir); err != nil {
		return nil, fmt.Errorf("tanggal lahir harus format YYYY-MM-DD: %q", f.tanggalLahir)
	}
	return f, nil
}
