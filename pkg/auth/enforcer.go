package auth

import "fmt"

const (
	PermGraphRead       = "graph:read"
	PermEngageToolRun   = "engage:tool:run"
	PermEngageJobCreate = "engage:job:create"
	PermEngageReportRead = "engage:report:read"
	PermEngageAdmin     = "engage:admin"
)

// Enforcer applies RBAC when enabled.
type Enforcer struct {
	enabled          bool
	roleReader       string
	roleAdmin        string
	roleEngageRunner string
	roleEngageAdmin  string
}

func NewEnforcer(cfg Config) *Enforcer {
	return &Enforcer{
		enabled:          cfg.RBACEnabled,
		roleReader:       cfg.RoleReader,
		roleAdmin:        cfg.RoleAdmin,
		roleEngageRunner: cfg.RoleEngageRunner,
		roleEngageAdmin:  cfg.RoleEngageAdmin,
	}
}

// Enforce checks permission for the subject. When RBAC is disabled, any authenticated subject is allowed.
func (e *Enforcer) Enforce(sub *Subject, perm string) error {
	if sub == nil {
		return ErrUnauthorized
	}
	if !e.enabled {
		return nil
	}
	switch perm {
	case PermGraphRead:
		if sub.HasRole(e.roleReader) || sub.HasRole(e.roleAdmin) {
			return nil
		}
		return ErrForbidden
	case PermEngageToolRun, PermEngageJobCreate, PermEngageReportRead:
		if sub.HasRole(e.roleEngageRunner) || sub.HasRole(e.roleEngageAdmin) || sub.HasRole(e.roleAdmin) {
			return nil
		}
		return ErrForbidden
	case PermEngageAdmin:
		if sub.HasRole(e.roleEngageAdmin) || sub.HasRole(e.roleAdmin) {
			return nil
		}
		return ErrForbidden
	default:
		return fmt.Errorf("unknown permission: %s", perm)
	}
}
