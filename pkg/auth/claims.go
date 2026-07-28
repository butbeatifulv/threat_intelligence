package auth

func SubjectFromClaims(sub string, claims map[string]any, clientID string) *Subject {
	email, _ := claims["email"].(string)
	return &Subject{
		Sub:    sub,
		Email:  email,
		Roles:  extractRoles(claims, clientID),
		Claims: claims,
	}
}

func extractRoles(claims map[string]any, clientID string) []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(rs []string) {
		for _, r := range rs {
			if r == "" {
				continue
			}
			if _, ok := seen[r]; ok {
				continue
			}
			seen[r] = struct{}{}
			out = append(out, r)
		}
	}
	if ra, ok := claims["realm_access"].(map[string]any); ok {
		add(stringSlice(ra["roles"]))
	}
	if res, ok := claims["resource_access"].(map[string]any); ok {
		if ent, ok := res[clientID].(map[string]any); ok {
			add(stringSlice(ent["roles"]))
		}
	}
	return out
}

func stringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, x := range arr {
		if s, ok := x.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
