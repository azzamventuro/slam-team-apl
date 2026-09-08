package utils

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("s3cret123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !CheckPassword(hash, "s3cret123") {
		t.Fatal("correct password rejected")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatal("wrong password accepted")
	}
}
