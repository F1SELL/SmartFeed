package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	jwtpkg "SmartFeed/pkg/jwt"
)

func TestAuthUnauthorizedWithoutBearer(t *testing.T) {
	h := Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthOKWithValidToken(t *testing.T) {
	token, err := jwtpkg.GenerateToken(7, "user", "secret")
	if err != nil {
		t.Fatalf("token generate failed: %v", err)
	}

	h := Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := UserIDFromContext(r.Context())
		if !ok || uid != 7 {
			t.Fatalf("missing user id in context")
		}
		role, ok := RoleFromContext(r.Context())
		if !ok || role != "user" {
			t.Fatalf("missing role in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestRequireRoles(t *testing.T) {
	base := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := RequireRoles("admin")(base)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without role, got %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2 = req2.WithContext(context.WithValue(req2.Context(), ctxKeyRole, "user"))
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusForbidden {
		t.Fatalf("expected 403 with wrong role, got %d", rr2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3 = req3.WithContext(context.WithValue(req3.Context(), ctxKeyRole, "admin"))
	rr3 := httptest.NewRecorder()
	h.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Fatalf("expected 200 with allowed role, got %d", rr3.Code)
	}
}
