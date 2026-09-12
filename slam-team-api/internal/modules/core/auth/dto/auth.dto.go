// Package dto holds the auth module's request/response shapes with binding tags.
package dto

// LoginRequest is the POST /auth/login body. Identifier is the username OR the
// email — the repository matches both in one query.
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required,min=8"`
}

// RefreshRequest is the POST /auth/refresh body: the opaque refresh token
// handed out by login (or by the previous refresh — tokens rotate).
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest is the POST /auth/logout body. The refresh token names the
// session to revoke; without it the access token alone cannot identify a
// sesi_login row.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthRole is the role summary embedded in AuthUser and MeResponse.
type AuthRole struct {
	ID      int64  `json:"id"`
	Nama    string `json:"nama"`
	Level   int    `json:"level"`
	IsSuper bool   `json:"is_super"`
}

// AuthUser is the user projection returned with a token pair. Nama and
// FotoURL come from the linked anggota row (users has no name column).
type AuthUser struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	AnggotaID   *int64   `json:"anggota_id"`
	Nama        string   `json:"nama"`
	FotoURL     *string  `json:"foto_url,omitempty"`
	Role        AuthRole `json:"role"`
	PermVersion int64    `json:"perm_version"`
}

// AuthResponse is returned by login and refresh (rancangan Bab 9 / api-endpoints §1).
type AuthResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"` // always "Bearer"
	ExpiresIn    int      `json:"expires_in"` // access-token lifetime in seconds
	User         AuthUser `json:"user"`
}

// MeAnggota is the linked-anggota summary on GET /me.
type MeAnggota struct {
	ID          int64   `json:"id"`
	NamaLengkap string  `json:"nama_lengkap"`
	NoInduk     *string `json:"no_induk,omitempty"`
	FotoUUID    *string `json:"foto_uuid,omitempty"`
	FotoURL     *string `json:"foto_url,omitempty"`
}

// MeResponse is returned by GET /me: the account, its role, and the anggota
// behind it (nil when the account is not yet linked to a member).
type MeResponse struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	AnggotaID   *int64     `json:"anggota_id"`
	Nama        string     `json:"nama"`
	Timezone    string     `json:"timezone"`
	IsAktif     bool       `json:"is_aktif"`
	Role        AuthRole   `json:"role"`
	Anggota     *MeAnggota `json:"anggota"`
	PermVersion int64      `json:"perm_version"`
}

// MyPermissionsResp is returned by GET /auth/me/permissions.
type MyPermissionsResp struct {
	Permissions []string `json:"permissions"` // e.g. ["anggota.read", "file.delete"] or ["*"] for super
	PermVersion int64    `json:"perm_version"`
	IsSuper     bool     `json:"is_super"`
}
