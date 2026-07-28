package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/auth/static"
)

func TestAuthMiddleware_delegates(t *testing.T) {
	v := static.New("secret", "runner", []string{"veil-reader"})
	stack := auth.NewStack(v, auth.Config{
		Enabled:      true,
		RBACEnabled:  true,
		RoleReader:   "veil-reader",
	})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sub, ok := auth.SubjectFromContext(r.Context())
		if !ok {
			http.Error(w, "no subject", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sub.Sub))
	})
	h := AuthMiddleware(stack, false, false, auth.PermGraphRead, next)

	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "runner" {
		t.Fatalf("status %d body %q", rec.Code, rec.Body.String())
	}
}

func TestProdMode_toggle(t *testing.T) {
	SetProdMode(true)
	if !ProdMode() {
		t.Fatal("expected prod mode on")
	}
	SetProdMode(false)
	if ProdMode() {
		t.Fatal("expected prod mode off")
	}
}
