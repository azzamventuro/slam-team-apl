// Package service contains the business rules for the user-management module.
//
// ANTI-ESCALATION RULES (enforced here, not just UI):
//   - Cannot create/edit a user whose role.level <= actor's role.level.
//   - Cannot grant a role with level <= actor's level.
//   - The last super admin cannot be deleted/disabled/demoted.
//   - Super Admin (is_super=true) can never be created via HTTP.
//   - Usernames and emails are unique across active (non-deleted) rows.
//   - Passwords are bcrypt-hashed on create/reset.
//   - All changes write audit log entries.
package service

import (
	"context"
	"fmt"
	"time"

	"slam-team-api/internal/modules/core/usermgmt/domain"
	"slam-team-api/internal/modules/core/usermgmt/dto"
	"slam-team-api/internal/modules/core/usermgmt/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/utils"
)

// Band defines the role-level range for a user-management surface.
type Band struct {
	Min int // inclusive lower bound of role.level
	Max int // inclusive upper bound of role.level
}

// Predefined bands matching the permission matrix.
var (
	BandAdmin     = Band{Min: 1, Max: 19}
	BandModerator = Band{Min: 20, Max: 29}
	BandUser      = Band{Min: 30, Max: 98}
)

// Actor captures the authenticated user making a request.
type Actor struct {
	UserID    int64
	RoleLevel int
	IsSuper   bool
	IPAddress string
	UserAgent string
}

// UsermgmtService composes repository + audit for business operations.
type UsermgmtService struct {
	repo    *repository.UsermgmtRepository
	auditor *audit.Writer
}

// NewUsermgmtService creates a service bound to the repository and audit writer.
func NewUsermgmtService(repo *repository.UsermgmtRepository, auditor *audit.Writer) *UsermgmtService {
	return &UsermgmtService{repo: repo, auditor: auditor}
}

// ── Create ──

// Create validates anti-escalation rules, inserts a user + user_role, hashes
// the password, writes an audit log, and returns the populated DTO.
func (s *UsermgmtService) Create(ctx context.Context, req dto.CreateUserReq, band Band, actor Actor) (*dto.UserResp, error) {
	// 1. Validate role exists and is within band.
	role, err := s.validateRoleForBand(ctx, req.RoleID, band)
	if err != nil {
		return nil, err
	}

	// 2. Anti-escalation: cannot create user with role.level <= actor's level
	//    (unless actor is super admin).
	if !actor.IsSuper && role.Level <= actor.RoleLevel {
		return nil, fmt.Errorf("tidak dapat menetapkan peran dengan level %d (setara/lebih rendah dari level anda %d): %w",
			role.Level, actor.RoleLevel, apperr.ErrForbidden)
	}

	// 3. Anti-escalation: cannot create a super admin via HTTP.
	if role.IsSuper {
		return nil, fmt.Errorf("peran super admin hanya dapat dibuat melalui CLI: %w", apperr.ErrForbidden)
	}

	// 4. Anggota FK exists and doesn't already have an account.
	ok, err := s.repo.AnggotaExists(ctx, req.AnggotaID)
	if err != nil {
		return nil, fmt.Errorf("cek anggota: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("anggota tidak ditemukan: %w", apperr.ErrValidation)
	}
	hasAccount, err := s.repo.AnggotaHasAccount(ctx, req.AnggotaID)
	if err != nil {
		return nil, fmt.Errorf("cek akun anggota: %w", err)
	}
	if hasAccount {
		return nil, fmt.Errorf("anggota ini sudah memiliki akun user: %w", apperr.ErrConflict)
	}

	// 5. Username uniqueness.
	exists, err := s.repo.ExistsUsername(ctx, req.Username, 0)
	if err != nil {
		return nil, fmt.Errorf("cek username: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("username '%s' sudah digunakan: %w", req.Username, apperr.ErrConflict)
	}

	// 6. Email uniqueness.
	exists, err = s.repo.ExistsEmail(ctx, req.Email, 0)
	if err != nil {
		return nil, fmt.Errorf("cek email: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email '%s' sudah digunakan: %w", req.Email, apperr.ErrConflict)
	}

	// 7. Hash password.
	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// 8. Default timezone.
	tz := "Asia/Jakarta"
	if req.Timezone != nil && *req.Timezone != "" {
		tz = *req.Timezone
	}

	// 9. Default is_aktif = true.
	isAktif := true
	if req.IsAktif != nil {
		isAktif = *req.IsAktif
	}

	// 10. Build entity.
	u := domain.User{
		AnggotaID: req.AnggotaID,
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashed,
		RoleID:    req.RoleID,
		Timezone:  tz,
		IsAktif:   isAktif,
	}
	actorID := actor.UserID
	u.CreatedBy = &actorID

	// 11. Persist (user + user_role in transaction).
	if err := s.repo.Create(ctx, &u, req.RoleID, &actorID); err != nil {
		return nil, err
	}

	// 12. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actorID,
		Modul:       "user-management",
		Aksi:        "create",
		ReffType:    "users",
		ReffID:      &u.ID,
		Ringkasan:   fmt.Sprintf("Membuat akun user '%s' dengan peran %s", u.Username, role.Nama),
		NilaiBaru:   map[string]any{"username": u.Username, "email": u.Email, "role_id": u.RoleID},
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, &actorID))

	// 13. Fetch back with joins for the response.
	created, err := s.repo.FindByID(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	return toResp(created), nil
}

// ── Update ──

// Update validates anti-escalation rules, applies changes, and returns the
// updated DTO. Password and role changes are optional.
func (s *UsermgmtService) Update(ctx context.Context, id int64, req dto.UpdateUserReq, band Band, actor Actor) (*dto.UserResp, error) {
	// 1. Load existing user.
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Anti-escalation: cannot edit a user whose role.level <= actor's level
	//    (unless actor is super admin).
	if !actor.IsSuper && existing.RoleLevel <= actor.RoleLevel {
		return nil, fmt.Errorf("tidak dapat mengedit akun dengan peran level %d (setara/lebih rendah dari level anda %d): %w",
			existing.RoleLevel, actor.RoleLevel, apperr.ErrForbidden)
	}

	// 3. Anti-escalation: last super admin cannot be edited for demotion.
	if existing.RoleIsSuper {
		isLast, err := s.isLastSuperAdmin(ctx, existing.ID)
		if err != nil {
			return nil, err
		}
		if isLast {
			// Allow editing other fields but block role change and demotion.
			if req.RoleID != nil {
				return nil, fmt.Errorf("super admin terakhir tidak dapat dipindahkan perannya: %w", apperr.ErrForbidden)
			}
			if req.IsAktif != nil && !*req.IsAktif {
				return nil, fmt.Errorf("super admin terakhir tidak dapat dinonaktifkan: %w", apperr.ErrForbidden)
			}
		}
	}

	// 4. Snapshot for audit.
	before := *existing

	// 5. Apply username change (if provided).
	if req.Username != nil && *req.Username != existing.Username {
		exists, err := s.repo.ExistsUsername(ctx, *req.Username, id)
		if err != nil {
			return nil, fmt.Errorf("cek username: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("username '%s' sudah digunakan: %w", *req.Username, apperr.ErrConflict)
		}
		existing.Username = *req.Username
	}

	// 6. Apply email change (if provided).
	if req.Email != nil && *req.Email != existing.Email {
		exists, err := s.repo.ExistsEmail(ctx, *req.Email, id)
		if err != nil {
			return nil, fmt.Errorf("cek email: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("email '%s' sudah digunakan: %w", *req.Email, apperr.ErrConflict)
		}
		existing.Email = *req.Email
	}

	// 7. Apply password change (if provided) — rehash.
	if req.Password != nil && *req.Password != "" {
		hashed, err := utils.HashPassword(*req.Password)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		existing.Password = hashed
		now := time.Now().UTC()
		existing.PasswordDiubah = &now
	}

	// 8. Apply role change (if provided) — anti-escalation check.
	if req.RoleID != nil && *req.RoleID != existing.RoleID {
		role, err := s.validateRoleForBand(ctx, *req.RoleID, band)
		if err != nil {
			return nil, err
		}
		if !actor.IsSuper && role.Level <= actor.RoleLevel {
			return nil, fmt.Errorf("tidak dapat menetapkan peran dengan level %d (setara/lebih rendah dari level anda %d): %w",
				role.Level, actor.RoleLevel, apperr.ErrForbidden)
		}
		if role.IsSuper {
			return nil, fmt.Errorf("peran super admin hanya dapat ditetapkan melalui CLI: %w", apperr.ErrForbidden)
		}
		// Sync user_role mirror.
		if err := s.repo.UpdateRoleSync(ctx, id, *req.RoleID, &actor.UserID); err != nil {
			return nil, fmt.Errorf("sinkronisasi user_role: %w", err)
		}
		existing.RoleID = *req.RoleID
	}

	// 9. Apply timezone change.
	if req.Timezone != nil {
		existing.Timezone = *req.Timezone
	}

	// 10. Apply is_aktif change.
	if req.IsAktif != nil {
		existing.IsAktif = *req.IsAktif
	}

	// 11. Apply buka_kunci (unlock account).
	if req.BukaKunci != nil && *req.BukaKunci {
		existing.GagalLogin = 0
		existing.TerkunciSampai = nil
	}

	// 12. Set audit fields.
	actorID := actor.UserID
	existing.ModifiedBy = &actorID
	now := time.Now().UTC()
	existing.ModifiedAt = &now

	// 13. Persist.
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	// 14. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actorID,
		Modul:       "user-management",
		Aksi:        "update",
		ReffType:    "users",
		ReffID:      &id,
		Ringkasan:   fmt.Sprintf("Mengubah akun user '%s'", existing.Username),
		NilaiLama:   before,
		NilaiBaru:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, &actorID))

	// 15. Fetch back with joins.
	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(updated), nil
}

// ── Delete ──

// SoftDelete checks guard conditions, then soft-deletes and writes an audit log.
func (s *UsermgmtService) SoftDelete(ctx context.Context, id int64, actor Actor) error {
	// 1. Load existing row.
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Anti-escalation: cannot delete a user whose role.level <= actor's level.
	if !actor.IsSuper && existing.RoleLevel <= actor.RoleLevel {
		return fmt.Errorf("tidak dapat menghapus akun dengan peran level %d (setara/lebih rendah dari level anda %d): %w",
			existing.RoleLevel, actor.RoleLevel, apperr.ErrForbidden)
	}

	// 3. Last super admin protection.
	if existing.RoleIsSuper {
		isLast, err := s.isLastSuperAdmin(ctx, existing.ID)
		if err != nil {
			return err
		}
		if isLast {
			return fmt.Errorf("super admin terakhir tidak dapat dihapus: %w", apperr.ErrForbidden)
		}
	}

	// 4. Soft-delete.
	actorID := actor.UserID
	if err := s.repo.SoftDelete(ctx, id, &actorID); err != nil {
		return err
	}

	// 5. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actorID,
		Modul:       "user-management",
		Aksi:        "delete",
		ReffType:    "users",
		ReffID:      &id,
		Ringkasan:   fmt.Sprintf("Menghapus akun user '%s'", existing.Username),
		NilaiLama:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, &actorID))

	return nil
}

// ── Read-only ──

// List returns a paginated list of users within the given band.
func (s *UsermgmtService) List(ctx context.Context, query dto.ListUserQuery, band Band) ([]dto.UserResp, int64, error) {
	page, perPage := query.Page, query.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	q := repository.ListQuery{
		Q:          query.Q,
		IsAktif:    query.IsAktif,
		InstansiID: query.InstansiID,
		Sort:       query.Sort,
		Offset:     offset,
		Limit:      perPage,
	}

	items, total, err := s.repo.List(ctx, q, band.Min, band.Max)
	if err != nil {
		return nil, 0, err
	}

	out := make([]dto.UserResp, len(items))
	for i := range items {
		out[i] = *toResp(&items[i])
	}
	return out, total, nil
}

// FindByID returns a single user as DTO.
func (s *UsermgmtService) FindByID(ctx context.Context, id int64) (*dto.UserResp, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(u), nil
}

// AvailableRoles returns roles within the band that the actor can manage
// (level > actor's level, is_super=false, is_aktif=true).
func (s *UsermgmtService) AvailableRoles(ctx context.Context, band Band, actorLevel int) ([]dto.RoleOpsi, error) {
	roles, err := s.repo.RolesByBand(ctx, band.Min, band.Max, actorLevel)
	if err != nil {
		return nil, err
	}
	out := make([]dto.RoleOpsi, len(roles))
	for i, r := range roles {
		out[i] = dto.RoleOpsi{ID: r.ID, Nama: r.Nama, Level: r.Level}
	}
	return out, nil
}

// AvailableAnggota returns anggota without user accounts, optionally filtered.
func (s *UsermgmtService) AvailableAnggota(ctx context.Context, q string, instansiID *int64) ([]dto.AnggotaOpsi, error) {
	items, err := s.repo.AnggotaWithoutAccount(ctx, q, instansiID, 20)
	if err != nil {
		return nil, err
	}
	out := make([]dto.AnggotaOpsi, len(items))
	for i, a := range items {
		out[i] = dto.AnggotaOpsi{ID: a.ID, NamaLengkap: a.NamaLengkap, NoInduk: a.NoInduk}
	}
	return out, nil
}

// ── Private helpers ──

// validateRoleForBand checks that the role exists, is active, is within the band,
// and is not a super admin role.
func (s *UsermgmtService) validateRoleForBand(ctx context.Context, roleID int64, band Band) (*domain.Role, error) {
	ok, err := s.repo.RoleExists(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("cek role: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("peran tidak ditemukan atau tidak aktif: %w", apperr.ErrValidation)
	}

	// Fetch the role to check level and is_super.
	var role domain.Role
	if err := s.repo.DB().WithContext(ctx).Where("id = ? AND is_deleted = false", roleID).First(&role).Error; err != nil {
		return nil, fmt.Errorf("ambil data peran: %w", err)
	}

	// Must be within the surface band.
	if role.Level < band.Min || role.Level > band.Max {
		return nil, fmt.Errorf("peran '%s' (level %d) tidak termasuk dalam band ini (%d–%d): %w",
			role.Nama, role.Level, band.Min, band.Max, apperr.ErrValidation)
	}

	return &role, nil
}

// isLastSuperAdmin checks if the given user is the only active super admin.
func (s *UsermgmtService) isLastSuperAdmin(ctx context.Context, userID int64) (bool, error) {
	count, err := s.repo.CountSuperAdmin(ctx)
	if err != nil {
		return false, fmt.Errorf("hitung super admin: %w", err)
	}
	if count > 1 {
		return false, nil
	}
	// count == 1 — check if it's this user.
	isSA, err := s.repo.IsSuperAdmin(ctx, userID)
	if err != nil {
		return false, err
	}
	return isSA, nil
}

// toResp maps a domain entity to the response DTO.
func toResp(u *domain.User) *dto.UserResp {
	var loginStr *string
	if u.LoginTerakhir != nil {
		s := u.LoginTerakhir.Format(time.RFC3339)
		loginStr = &s
	}
	var terkunciStr *string
	if u.TerkunciSampai != nil {
		s := u.TerkunciSampai.Format(time.RFC3339)
		terkunciStr = &s
	}
	var modStr *string
	if u.ModifiedAt != nil {
		s := u.ModifiedAt.Format(time.RFC3339)
		modStr = &s
	}

	var fotoURL *string
	if u.AnggotaFotoUUID != nil && *u.AnggotaFotoUUID != "" {
		s := "/api/v1/files/" + *u.AnggotaFotoUUID + "/low"
		fotoURL = &s
	}

	return &dto.UserResp{
		ID: u.ID,
		Anggota: dto.AnggotaInfo{
			ID:          u.AnggotaID,
			NamaLengkap: derefStr(u.AnggotaNamaLengkap, ""),
			NoInduk:     u.AnggotaNoInduk,
			InstansiID:  derefInt64(u.AnggotaInstansiID, 0),
			FotoURL:     fotoURL,
		},
		Username:       u.Username,
		Email:          u.Email,
		Role:           dto.RoleInfo{ID: u.RoleID, Nama: u.RoleNama, Level: u.RoleLevel, IsSuper: u.RoleIsSuper},
		Timezone:       u.Timezone,
		IsAktif:        u.IsAktif,
		LoginTerakhir:  loginStr,
		TerkunciSampai: terkunciStr,
		CreatedAt:      u.CreatedAt.Format(time.RFC3339),
		ModifiedAt:     modStr,
	}
}

func derefStr(p *string, fallback string) string {
	if p != nil {
		return *p
	}
	return fallback
}

func derefInt64(p *int64, fallback int64) int64 {
	if p != nil {
		return *p
	}
	return fallback
}
