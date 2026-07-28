package auth

import (
	"errors"
	"fmt"
	"testing"
)

func TestEnforcer_RBAC(t *testing.T) {
	cfg := Config{
		Enabled:     true,
		RBACEnabled: true,
		RoleReader:  "veil-reader",
		RoleAdmin:   "veil-admin",
	}
	e := NewEnforcer(cfg)

	reader := &Subject{Roles: []string{"veil-reader"}}
	if err := e.Enforce(reader, PermGraphRead); err != nil {
		t.Fatalf("reader: %v", err)
	}

	none := &Subject{Roles: []string{"other"}}
	if err := e.Enforce(none, PermGraphRead); err != ErrForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}

	off := NewEnforcer(Config{Enabled: true, RBACEnabled: false})
	if err := off.Enforce(none, PermGraphRead); err != nil {
		t.Fatalf("rbac off: %v", err)
	}
}

func TestEnforcer_Enforce_table(t *testing.T) {
	cfg := Config{
		RBACEnabled:      true,
		RoleReader:       "veil-reader",
		RoleAdmin:        "veil-admin",
		RoleEngageRunner: "veil-engage-runner",
		RoleEngageAdmin:  "veil-engage-admin",
	}
	e := NewEnforcer(cfg)

	reader := &Subject{Roles: []string{"veil-reader"}}
	admin := &Subject{Roles: []string{"veil-admin"}}
	runner := &Subject{Roles: []string{"veil-engage-runner"}}
	engageAdmin := &Subject{Roles: []string{"veil-engage-admin"}}
	none := &Subject{Roles: []string{"other"}}

	tests := []struct {
		name    string
		sub     *Subject
		perm    string
		wantErr error
	}{
		{"nil subject", nil, PermGraphRead, ErrUnauthorized},
		{"graph read reader", reader, PermGraphRead, nil},
		{"graph read admin", admin, PermGraphRead, nil},
		{"graph read runner denied", runner, PermGraphRead, ErrForbidden},
		{"graph read engage admin denied", engageAdmin, PermGraphRead, ErrForbidden},
		{"graph read none", none, PermGraphRead, ErrForbidden},
		{"engage tool run runner", runner, PermEngageToolRun, nil},
		{"engage tool run engage admin", engageAdmin, PermEngageToolRun, nil},
		{"engage tool run admin", admin, PermEngageToolRun, nil},
		{"engage tool run reader denied", reader, PermEngageToolRun, ErrForbidden},
		{"engage tool run none", none, PermEngageToolRun, ErrForbidden},
		{"engage job create runner", runner, PermEngageJobCreate, nil},
		{"engage job create admin", admin, PermEngageJobCreate, nil},
		{"engage report read runner", runner, PermEngageReportRead, nil},
		{"engage report read engage admin", engageAdmin, PermEngageReportRead, nil},
		{"engage admin engage admin", engageAdmin, PermEngageAdmin, nil},
		{"engage admin admin", admin, PermEngageAdmin, nil},
		{"engage admin runner denied", runner, PermEngageAdmin, ErrForbidden},
		{"engage admin reader denied", reader, PermEngageAdmin, ErrForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := e.Enforce(tt.sub, tt.perm)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Enforce() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnforcer_Enforce_unknownPermission(t *testing.T) {
	e := NewEnforcer(Config{RBACEnabled: true, RoleReader: "veil-reader"})
	sub := &Subject{Roles: []string{"veil-reader"}}
	err := e.Enforce(sub, "bogus:perm")
	if err == nil || err.Error() != fmt.Sprintf("unknown permission: %s", "bogus:perm") {
		t.Fatalf("got %v", err)
	}
}

func TestEnforcer_RBACDisabled_allowsAnyAuthenticated(t *testing.T) {
	e := NewEnforcer(Config{RBACEnabled: false})
	sub := &Subject{Roles: []string{"other"}}
	for _, perm := range []string{
		PermGraphRead,
		PermEngageToolRun,
		PermEngageJobCreate,
		PermEngageReportRead,
		PermEngageAdmin,
	} {
		if err := e.Enforce(sub, perm); err != nil {
			t.Fatalf("%s: %v", perm, err)
		}
	}
}

func TestEnforcer_RBACDisabled_nilSubjectDenied(t *testing.T) {
	e := NewEnforcer(Config{RBACEnabled: false})
	if err := e.Enforce(nil, PermGraphRead); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("got %v", err)
	}
}
