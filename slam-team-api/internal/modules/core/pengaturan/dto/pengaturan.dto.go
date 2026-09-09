// Package dto holds the pengaturan request/response shapes. Domain entities are
// never bound or serialised directly.
package dto

import (
	"encoding/json"
	"time"
)

// SettingItem is one row as the settings page consumes it: every piece of
// metadata needed to render a control, plus Nilai/NilaiBawaan already cast to
// their JSON type (integer → number, boolean → bool, json → object/array,
// string/date → string) so the client never parses text itself.
type SettingItem struct {
	ID          int64           `json:"id"`
	Kunci       string          `json:"kunci"`
	Grup        string          `json:"grup"`
	Label       string          `json:"label"`
	Nilai       any             `json:"nilai"`
	TipeNilai   string          `json:"tipe_nilai"`
	NilaiBawaan any             `json:"nilai_bawaan"`
	Opsi        json.RawMessage `json:"opsi"`
	Satuan      *string         `json:"satuan"`
	Keterangan  *string         `json:"keterangan"`
	Urutan      int             `json:"urutan"`
	IsPublik    bool            `json:"is_publik"`
	IsTerkunci  bool            `json:"is_terkunci"`
	ModifiedAt  *time.Time      `json:"modified_at"`
	ModifiedBy  *int64          `json:"modified_by"`
}

// SettingGroup is one tab of the settings page: a grup and its rows in urutan
// order. GET /pengaturan returns these so the client does not have to group.
type SettingGroup struct {
	Grup  string        `json:"grup"`
	Items []SettingItem `json:"items"`
}

// SettingKV is one entry of a bulk update.
//
// Nilai is a *json.RawMessage rather than a plain value on purpose. The setting
// values 0, false and "" are all legitimate, so `binding:"required"` on a
// non-pointer would reject them as "empty"; against a pointer it only rejects
// an absent field or an explicit null. RawMessage also preserves the literal
// the client sent, so a large integer never round-trips through float64.
type SettingKV struct {
	Kunci string           `json:"kunci" binding:"required,max=100"`
	Nilai *json.RawMessage `json:"nilai" binding:"required"`
}

// UpdatePengaturanReq is the PUT /pengaturan body. The update is atomic: every
// item validates, or nothing is written.
type UpdatePengaturanReq struct {
	Items []SettingKV `json:"items" binding:"required,min=1,dive"`
}

// PublicSettings is the GET /public/pengaturan payload: a flat kunci → cast
// value map of the is_publik rows only.
type PublicSettings map[string]any
