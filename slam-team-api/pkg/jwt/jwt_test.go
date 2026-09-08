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
