package auth

// Subject is an authenticated principal from a validated JWT.
type Subject struct {
	Sub    string
	Email  string
	Roles  []string
	Claims map[string]any
}

func (s *Subject) HasRole(role string) bool {
	for _, r := range s.Roles {
		if r == role {
			return true
		}
	}
	return false
}
