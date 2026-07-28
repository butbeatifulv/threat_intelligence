package auth

import "testing"

func TestSubjectFromClaims_roles(t *testing.T) {
	claims := map[string]any{
		"email": "u@example.com",
		"realm_access": map[string]any{
			"roles": []any{"veil-reader", "veil-reader"},
		},
		"resource_access": map[string]any{
			"veil-api": map[string]any{
				"roles": []any{"veil-admin"},
			},
		},
	}
	sub := SubjectFromClaims("user-1", claims, "veil-api")
	if sub.Sub != "user-1" || sub.Email != "u@example.com" {
		t.Fatalf("subject: %+v", sub)
	}
	if len(sub.Roles) != 2 {
		t.Fatalf("roles: %v", sub.Roles)
	}
	if !sub.HasRole("veil-reader") || !sub.HasRole("veil-admin") {
		t.Fatalf("missing roles: %v", sub.Roles)
	}
}

func TestExtractRoles_empty(t *testing.T) {
	sub := SubjectFromClaims("x", map[string]any{}, "veil-api")
	if len(sub.Roles) != 0 {
		t.Fatalf("expected no roles, got %v", sub.Roles)
	}
}

func TestExtractRoles_skipsEmptyAndBadTypes(t *testing.T) {
	claims := map[string]any{
		"realm_access": map[string]any{
			"roles": []any{"", "veil-reader", 42},
		},
		"resource_access": map[string]any{
			"veil-api": map[string]any{
				"roles": "not-a-slice",
			},
		},
	}
	sub := SubjectFromClaims("u", claims, "veil-api")
	if len(sub.Roles) != 1 || sub.Roles[0] != "veil-reader" {
		t.Fatalf("roles: %v", sub.Roles)
	}
}
