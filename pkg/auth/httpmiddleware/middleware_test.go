package httpmiddleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/auth/static"
)

func testStack(t *testing.T, rbac bool) *auth.Stack {
	t.Helper()
	v := static.New("test-token", "runner-1", []string{"veil-reader", "veil-engage-runner"})
	cfg := auth.Config{
		Enabled:          true,
		RBACEnabled:      rbac,
		RoleReader:       "veil-reader",
		RoleAdmin:        "veil-admin",
		RoleEngageRunner: "veil-engage-runner",
		RoleEngageAdmin:  "veil-engage-admin",
	}
	return auth.NewStack(v, cfg)
}

func okHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sub, ok := auth.SubjectFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sub.Sub))
	})
}

func TestAuth_disabled_passesThrough(t *testing.T) {
	stack := testStack(t, true)
	stack.Config.Enabled = false
	h := Auth(stack, true, false, auth.PermGraphRead, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/tools", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestAuth_nilStack_passesThrough(t *testing.T) {
	h := Auth(nil, true, false, auth.PermGraphRead, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/secret", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestAuth_missingBearer_unauthorized(t *testing.T) {
	h := Auth(testStack(t, true), false, false, auth.PermGraphRead, okHandler(t))
	req := httptest.NewRequest(http.MethodGet, "/api/tools", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestAuth_validBearer_setsSubject(t *testing.T) {
	h := Auth(testStack(t, true), false, false, auth.PermEngageToolRun, okHandler(t))
	req := httptest.NewRequest(http.MethodPost, "/api/tools/nmap", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "runner-1" {
		t.Fatalf("status %d body %q", rec.Code, rec.Body.String())
	}
}

func TestAuth_forbiddenWithoutRole(t *testing.T) {
	v := static.New("tok", "u", []string{"veil-reader"})
	cfg := auth.Config{Enabled: true, RBACEnabled: true, RoleReader: "veil-reader", RoleEngageRunner: "veil-engage-runner"}
	stack := auth.NewStack(v, cfg)
	h := Auth(stack, false, false, auth.PermEngageToolRun, okHandler(t))
	req := httptest.NewRequest(http.MethodPost, "/api/tools/x", nil)
	req.Header.Set("Authorization", "Bearer tok")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestAuth_nonStrict_healthPublic(t *testing.T) {
	h := Auth(testStack(t, true), false, false, auth.PermGraphRead, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestAuth_strict_healthGETPublic(t *testing.T) {
	h := Auth(testStack(t, true), true, false, auth.PermGraphRead, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestAuth_strict_healthPOSTRequiresAuth(t *testing.T) {
	h := Auth(testStack(t, true), true, false, auth.PermGraphRead, okHandler(t))
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestAuth_prodForbiddenBody(t *testing.T) {
	v := static.New("tok", "u", []string{"veil-reader"})
	cfg := auth.Config{Enabled: true, RBACEnabled: true, RoleReader: "veil-reader", RoleEngageRunner: "veil-engage-runner"}
	stack := auth.NewStack(v, cfg)
	h := Auth(stack, false, true, auth.PermEngageToolRun, okHandler(t))
	req := httptest.NewRequest(http.MethodPost, "/api/tools/x", nil)
	req.Header.Set("Authorization", "Bearer tok")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusForbidden || body["error"] != "forbidden" {
		t.Fatalf("status %d body %v", rec.Code, body)
	}
}

func TestAuth_bearerTokenVariants(t *testing.T) {
	if bearerToken("Bearer tok") != "tok" || bearerToken("bearer") != "" {
		t.Fatal("bearer parsing")
	}
}

func TestAuth_invalidBearer(t *testing.T) {
	h := Auth(testStack(t, true), false, false, auth.PermGraphRead, okHandler(t))
	req := httptest.NewRequest(http.MethodGet, "/api/tools", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestAuth_prodMasksErrorBody(t *testing.T) {
	h := Auth(testStack(t, true), false, true, auth.PermGraphRead, okHandler(t))
	req := httptest.NewRequest(http.MethodGet, "/api/tools", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"] != "unauthorized" {
		t.Fatalf("body %v", body)
	}
}
