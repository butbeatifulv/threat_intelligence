package hostnorm

import "testing"

func TestNormalizeHost(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"https://Foo.COM", "foo.com"},
		{"http://example.com:8080/path", "example.com"},
		{"  ", ""},
		{"CVE-2024-1234", "cve-2024-1234"},
	}
	for _, tc := range tests {
		if got := NormalizeHost(tc.in); got != tc.want {
			t.Errorf("NormalizeHost(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
