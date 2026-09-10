// Package dto holds request/response structs for the hakakses (RBAC) endpoints.
package dto

import (
	"slam-team-api/internal/modules/core/hakakses/domain"
)

// ── Role CRUD ──

// CreateRoleReq is the body for POST /hakakses/roles.
type CreateRoleReq struct {
	Kode       string `json:"kode"       binding:"required,max=50"`
	Nama       string `json:"nama"       binding:"required,max=100"`
	Level      int    `json:"level"      binding:"required,min=10,max=98"`
	Keterangan string `json:"keterangan" binding:"max=500"`
}

// UpdateRoleReq is the body for PATCH /hakakses/roles/:id.
type UpdateRoleReq struct {
	Nama       string `json:"nama"       binding:"omitempty,max=100"`
	Level      int    `json:"level"      binding:"omitempty,min=10,max=98"`
	Keterangan string `json:"keterangan" binding:"max=500"`
}

// RoleResp is the standard role JSON shape returned by every endpoint.
type RoleResp struct {
	ID         int64               `json:"id"`
	Kode       string              `json:"kode"`
	Nama       string              `json:"nama"`
	Level      int                 `json:"level"`
	Keterangan string              `json:"keterangan,omitempty"`
	IsSistem   bool                `json:"is_sistem"`
	IsSuper    bool                `json:"is_super"`
	IsAktif    bool                `json:"is_aktif"`
	Persediaan bool                `json:"persediaan"` // computed in service
	CountUsers int64               `json:"count_users"`
	CountPerms int                 `json:"count_perms"`
}

// RoleDetailResp includes the full permission set for a single role.
type RoleDetailResp struct {
	RoleResp
	Permissions []RolePermissionEntry `json:"permissions"`
}

// RolePermissionEntry is a single permission row inside a role's detail.
type RolePermissionEntry struct {
	PermissionID  int64                    `json:"permission_id"`
	Kode          string                   `json:"kode"`
	Nama          string                   `json:"nama"`
	Aksi          domain.AksiPermission    `json:"aksi"`
	Cakupan       domain.CakupanPermission `json:"cakupan"`
	IsBerbahaya   bool                     `json:"is_berbahaya"`
}

// ── Permission matrix ──

// SetPermissionsReq is the body for PUT /hakakses/roles/:id/permissions.
// The entire matrix for a role is replaced atomically.
type SetPermissionsReq struct {
	Permissions []PermissionGrant `json:"permissions" binding:"required,min=1,dive"`
}

// PermissionGrant is one permission assignment inside a SetPermissionsReq.
type PermissionGrant struct {
	PermissionID int64                    `json:"permission_id" binding:"required"`
	Cakupan      domain.CakupanPermission `json:"cakupan"       binding:"required,oneof=semua instansi_sendiri milik_sendiri"`
}

// ── Permission catalogue ──

// ModulResp is a modul row as returned by the GET /hakakses/permissions endpoint.
type ModulResp struct {
	ID     int64           `json:"id"`
	Kode   string          `json:"kode"`
	Nama   string          `json:"nama"`
	Grup   string          `json:"grup"`
	Urutan int             `json:"urutan"`
	Perms  []PermResp      `json:"permissions"`
}

// PermResp is a single permission inside a ModulResp.
type PermResp struct {
	ID          int64  `json:"id"`
	Aksi        string `json:"aksi"`
	Kode        string `json:"kode"`
	Nama        string `json:"nama"`
	IsBerbahaya bool   `json:"is_berbahaya"`
}

// ── GET /me/permissions ──

// MyPermissionsResp is the shape returned by GET /auth/me/permissions.
type MyPermissionsResp struct {
	Permissions []string `json:"permissions"` // e.g. ["anggota.read", "file.delete"]
	PermVersion int64    `json:"perm_version"`
	IsSuper     bool     `json:"is_super"`
}
