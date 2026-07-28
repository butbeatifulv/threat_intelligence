package auth

// Stack groups verifier and RBAC enforcer for HTTP and MCP.
type Stack struct {
	Config   Config
	Verifier Verifier
	Enforcer *Enforcer
}

func NewStack(verifier Verifier, cfg Config) *Stack {
	return &Stack{
		Config:   cfg,
		Verifier: verifier,
		Enforcer: NewEnforcer(cfg),
	}
}
