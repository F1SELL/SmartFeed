package hash

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	h, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("unexpected hash error: %v", err)
	}
	if h == "" {
		t.Fatal("hash is empty")
	}
	if !CheckPassword("secret123", h) {
		t.Fatal("expected password to match hash")
	}
	if CheckPassword("bad", h) {
		t.Fatal("expected wrong password not to match")
	}
}
