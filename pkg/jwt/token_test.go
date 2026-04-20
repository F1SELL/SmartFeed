package jwt

import "testing"

func TestGenerateAndParseToken(t *testing.T) {
	token, err := GenerateToken(42, "admin", "secret")
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}

	claims, err := ParseToken(token, "secret")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if claims.UserID != 42 || claims.Role != "admin" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestParseTokenInvalid(t *testing.T) {
	token, err := GenerateToken(1, "user", "secret")
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}

	if _, err := ParseToken(token, "wrong"); err == nil {
		t.Fatal("expected parse error with wrong secret")
	}
	if _, err := ParseToken("not-a-token", "secret"); err == nil {
		t.Fatal("expected parse error with malformed token")
	}
}
