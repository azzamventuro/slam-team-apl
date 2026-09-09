package pagination_test

import (
	"testing"

	"slam-team-api/internal/shared/pagination"
)

func TestNormalizeCapsPerPage(t *testing.T) {
	tests := []struct {
		name          string
		in            pagination.ListQuery
		page, perPage int
	}{
		{"zero values get defaults", pagination.ListQuery{}, 1, pagination.DefaultPerPage},
		{"per_page capped at 100", pagination.ListQuery{Page: 2, PerPage: 5000}, 2, pagination.MaxPerPage},
		{"negative page clamped", pagination.ListQuery{Page: -3, PerPage: 20}, 1, 20},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			q := tc.in
			q.Normalize()
			if q.Page != tc.page || q.PerPage != tc.perPage {
				t.Errorf("Normalize() = page %d, per_page %d; want %d, %d",
					q.Page, q.PerPage, tc.page, tc.perPage)
			}
		})
	}
}

func TestOffset(t *testing.T) {
	q := pagination.ListQuery{Page: 4, PerPage: 25}
	if got := q.Offset(); got != 75 {
		t.Errorf("Offset() = %d, want 75", got)
	}
}

func TestNewComputesLastPage(t *testing.T) {
	q := pagination.ListQuery{Page: 1, PerPage: 20}
	p := pagination.New([]string{"a"}, q, 123)
	if p.LastPage != 7 {
		t.Errorf("LastPage = %d, want 7", p.LastPage)
	}
	// An empty result set is still page 1 of 1, with items [] not null.
	empty := pagination.New[string](nil, q, 0)
	if empty.LastPage != 1 || empty.Items == nil {
		t.Errorf("empty page = %+v, want last_page 1 and non-nil items", empty)
	}
}

func TestSortColumnWhitelist(t *testing.T) {
	allowed := []string{"created_at", "nama"}

	if col, desc, ok := (pagination.ListQuery{Sort: "-created_at"}).SortColumn(allowed...); !ok || col != "created_at" || !desc {
		t.Errorf("SortColumn(-created_at) = %q, %v, %v", col, desc, ok)
	}
	if _, _, ok := (pagination.ListQuery{Sort: "password"}).SortColumn(allowed...); ok {
		t.Error("non-whitelisted column was accepted")
	}
	if _, _, ok := (pagination.ListQuery{}).SortColumn(allowed...); ok {
		t.Error("empty sort should fall through to the repository default")
	}
}
