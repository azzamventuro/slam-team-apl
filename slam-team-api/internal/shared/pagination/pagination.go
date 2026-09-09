// Package pagination holds the standard list-query and paged-envelope shapes
// used by every master-data list endpoint, so `page`/`per_page`/`total` mean
// the same thing across all modules.
package pagination

import "math"

// MaxPerPage caps how many rows a client may request in one page.
const MaxPerPage = 100

// DefaultPerPage is applied when the client omits per_page.
const DefaultPerPage = 20

// ListQuery is the shared query DTO for list endpoints:
//
//	GET /<mod>?page=1&per_page=20&q=&sort=-created_at
//
// Bind it with c.ShouldBindQuery, then call Normalize before use.
type ListQuery struct {
	Page    int    `form:"page,default=1" binding:"min=1"`
	PerPage int    `form:"per_page,default=20" binding:"min=1,max=100"`
	Q       string `form:"q"`
	Sort    string `form:"sort"` // e.g. "-created_at" (leading - = DESC)
}

// Normalize clamps the query to safe values. Binding tags already reject
// out-of-range input, but a ListQuery built in code (or bound without tags)
// still goes through here, so per_page can never exceed MaxPerPage.
func (q *ListQuery) Normalize() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 {
		q.PerPage = DefaultPerPage
	}
	if q.PerPage > MaxPerPage {
		q.PerPage = MaxPerPage
	}
}

// Offset is the SQL OFFSET for the normalized query.
func (q ListQuery) Offset() int { return (q.Page - 1) * q.PerPage }

// Limit is the SQL LIMIT for the normalized query.
func (q ListQuery) Limit() int { return q.PerPage }

// SortColumn splits Sort into a column name and direction, accepting only
// columns present in allowed (a whitelist — `sort` is never interpolated into
// SQL). It returns ok=false when Sort is empty or not whitelisted, in which
// case the repository applies its own default ordering.
func (q ListQuery) SortColumn(allowed ...string) (col string, desc bool, ok bool) {
	if q.Sort == "" {
		return "", false, false
	}
	col, desc = q.Sort, false
	if col[0] == '-' {
		col, desc = col[1:], true
	}
	for _, a := range allowed {
		if a == col {
			return col, desc, true
		}
	}
	return "", false, false
}

// Paginated is the `data` payload of a list response:
//
//	{"success":true,"data":{"items":[...],"page":1,"per_page":20,"total":123,"last_page":7}}
type Paginated[T any] struct {
	Items    []T   `json:"items"`
	Page     int   `json:"page"`
	PerPage  int   `json:"per_page"`
	Total    int64 `json:"total"`
	LastPage int   `json:"last_page"`
}

// New builds the paged envelope, computing last_page from total and per_page.
// A nil items slice is emitted as [] rather than null.
func New[T any](items []T, q ListQuery, total int64) Paginated[T] {
	q.Normalize()
	if items == nil {
		items = []T{}
	}
	last := int(math.Ceil(float64(total) / float64(q.PerPage)))
	if last < 1 {
		last = 1
	}
	return Paginated[T]{
		Items:    items,
		Page:     q.Page,
		PerPage:  q.PerPage,
		Total:    total,
		LastPage: last,
	}
}
