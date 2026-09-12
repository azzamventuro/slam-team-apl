// Package dto holds the auth module's request/response shapes with binding tags.
package dto

// LoginRequest is the POST /auth/login body.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// UserResponse is the public projection of a user.
type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// LoginResponse is returned on successful authentication.
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// MyPermissionsResp is returned by GET /auth/me/permissions.
type MyPermissionsResp struct {
	Permissions []string `json:"permissions"` // e.g. ["anggota.read", "file.delete"] or ["*"] for super
	PermVersion int64    `json:"perm_version"`
	IsSuper     bool     `json:"is_super"`
}
