package jwt

import (
	"testing"
	"time"
)

// Issue then Verify must round-trip the claims; a wrong secret must fail.
func TestIssueVerifyRoundTrip(t *testing.T) {
	m := New("test-secret", "slam-team-api", time.Hour)

	token, err := m.Issue(42, "a@b.com", "Alice")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	claims, err := m.Verify(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.UserID != 42 || claims.Email != "a@b.com" || claims.Name != "Alice" {
		t.Fatalf("claims mismatch: %+v", claims)
	}

	// Wrong secret must be rejected.
	if _, err := New("other-secret", "x", time.Hour).Verify(token); err == nil {
		t.Fatal("expected verify failure with wrong secret")
	}
}

func TestVerifyExpired(t *testing.T) {
	m := New("test-secret", "slam-team-api", -time.Hour) // already expired
	token, _ := m.Issue(1, "a@b.com", "A")
	if _, err := m.Verify(token); err == nil {
		t.Fatal("expected expired token to fail verification")
	}
}

// IssueAccess must carry the RBAC claims and anggota_id through Verify; a
// token issued without an anggota keeps the claim nil (omitted, not zero).
func TestIssueAccessCarriesAnggotaID(t *testing.T) {
	m := New("test-secret", "slam-team-api", time.Hour)

	anggotaID := int64(12)
	token, err := m.IssueAccess(7, "a@b.com", "Alice", &anggotaID, 2, 10, false, 7)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	claims, err := m.Verify(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.AnggotaID == nil || *claims.AnggotaID != 12 {
		t.Fatalf("anggota_id = %v, want 12", claims.AnggotaID)
	}
	if claims.RoleID != 2 || claims.RoleLevel != 10 || claims.IsSuper || claims.PermVersion != 7 {
		t.Fatalf("rbac claims mismatch: %+v", claims)
	}

	legacy, _ := m.IssueWithRole(7, "a@b.com", "Alice", 2, 10, false, 7)
	claims, _ = m.Verify(legacy)
	if claims.AnggotaID != nil {
		t.Fatalf("legacy issuer must leave anggota_id nil, got %d", *claims.AnggotaID)
	}
}
