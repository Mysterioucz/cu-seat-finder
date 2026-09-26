package authn

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIssueParseRoundTrip(t *testing.T) {
	tok, err := Issue("secret", "u1", "admin", "Ann")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := Parse("secret", tok)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "u1" || claims.Role != "admin" || claims.Name != "Ann" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	tok, _ := Issue("secret", "u1", "customer", "Ann")
	if _, err := Parse("other-secret", tok); err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestRequireRoleRejectsMismatch(t *testing.T) {
	tok, _ := Issue("secret", "u1", "customer", "Ann")
	handlerCalled := false
	h := RequireRole("secret", "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	if handlerCalled {
		t.Fatal("handler should not run for wrong role")
	}
}

func TestRequireRoleAllowsMatch(t *testing.T) {
	tok, _ := Issue("secret", "u1", "admin", "Ann")
	h := RequireRole("secret", "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := FromContext(r.Context())
		if !ok || claims.UserID != "u1" {
			t.Error("claims not propagated to context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireRejectsMissingToken(t *testing.T) {
	h := Require("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not run")
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
