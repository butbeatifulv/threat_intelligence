package config

import "testing"

func TestValidateSecurity_requireAuth(t *testing.T) {
	sec := SecurityConfig{RequireAuth: true}
	if err := ValidateSecurity(sec, false); err == nil {
		t.Fatal("expected error")
	}
	if err := ValidateSecurity(sec, true); err != nil {
		t.Fatal(err)
	}
}

func TestLoadSecurityForEnv_prod(t *testing.T) {
	t.Setenv("VEIL_REQUIRE_AUTH", "1")
	sec := LoadSecurityForEnv("prod")
	if !sec.Prod || !sec.RequireAuth {
		t.Fatalf("sec: %+v", sec)
	}
}
