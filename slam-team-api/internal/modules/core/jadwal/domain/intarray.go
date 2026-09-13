package domain

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

// IntArray maps a PostgreSQL int[] column (jadwal.hari_ulang). The pgx stdlib
// driver hands non-intrinsic types to database/sql as their text form
// ("{1,3,5}"), so the type implements sql.Scanner to parse that and
// driver.Valuer to write it back — no lib/pq dependency needed for one array
// column. A nil slice is stored as SQL NULL; an empty slice as '{}'.
type IntArray []int64

// Scan implements sql.Scanner for the text array literal.
func (a *IntArray) Scan(src any) error {
	var s string
	switch v := src.(type) {
	case nil:
		*a = nil
		return nil
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("IntArray: tidak bisa scan dari %T", src)
	}

	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "{") || !strings.HasSuffix(s, "}") {
		return fmt.Errorf("IntArray: literal array tidak valid %q", s)
	}
	body := strings.TrimSpace(s[1 : len(s)-1])
	if body == "" {
		*a = IntArray{}
		return nil
	}

	parts := strings.Split(body, ",")
	out := make(IntArray, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			return fmt.Errorf("IntArray: elemen %q bukan bilangan bulat", p)
		}
		out = append(out, n)
	}
	*a = out
	return nil
}

// Value implements driver.Valuer, producing the text array literal.
func (a IntArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	parts := make([]string, len(a))
	for i, n := range a {
		parts[i] = strconv.FormatInt(n, 10)
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}

// Contains reports whether n is one of the array's elements.
func (a IntArray) Contains(n int64) bool {
	for _, v := range a {
		if v == n {
			return true
		}
	}
	return false
}
