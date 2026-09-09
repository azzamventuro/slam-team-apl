package migrations_test

import (
	"strings"
	"testing"

	"slam-team-api/migrations"

	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Every committed migration must be a complete, rollback-able pair; a missing
// .down.sql only shows up when a rollback is already needed.
func TestEveryMigrationHasBothDirections(t *testing.T) {
	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		t.Fatalf("read embedded migrations: %v", err)
	}

	seen := map[string]map[string]bool{}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		switch {
		case strings.HasSuffix(name, ".up.sql"):
			base := strings.TrimSuffix(name, ".up.sql")
			if seen[base] == nil {
				seen[base] = map[string]bool{}
			}
			seen[base]["up"] = true
		case strings.HasSuffix(name, ".down.sql"):
			base := strings.TrimSuffix(name, ".down.sql")
			if seen[base] == nil {
				seen[base] = map[string]bool{}
			}
			seen[base]["down"] = true
		default:
			t.Errorf("%s: expected a .up.sql / .down.sql suffix", name)
		}
	}

	if len(seen) == 0 {
		t.Fatal("no migrations were embedded")
	}
	for base, dirs := range seen {
		if !dirs["up"] {
			t.Errorf("%s: missing .up.sql", base)
		}
		if !dirs["down"] {
			t.Errorf("%s: missing .down.sql", base)
		}
	}
}

// iofs is what the runner parses the embedded files with, so a malformed
// version prefix fails here rather than against a live database.
func TestSourceParsesAndStartsAtOne(t *testing.T) {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatalf("iofs.New: %v", err)
	}
	defer src.Close()

	first, err := src.First()
	if err != nil {
		t.Fatalf("First: %v", err)
	}
	if first != 1 {
		t.Errorf("first migration version = %d, want 1 (0001_enums)", first)
	}
}
