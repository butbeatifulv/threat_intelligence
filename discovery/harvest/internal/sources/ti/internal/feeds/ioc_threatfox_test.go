package feeds

import (
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
)

func TestIocFromThreatFoxExport(t *testing.T) {
	cases := []struct {
		val  string
		typ  string
		want domain.IOCType
	}{
		{"evil.com", "domain", domain.IOCDomain},
		{"https://evil.com/a", "url", domain.IOCURL},
		{"1.2.3.4:443", "ip:port", domain.IOCIP},
		{"2001:db8::1", "ipv6", domain.IOCIP},
		{"deadbeefdeadbeefdeadbeefdeadbeef", "md5_hash", domain.IOCHash},
		{strings.Repeat("a", 64), "sha256", domain.IOCHash},
	}
	for _, tc := range cases {
		ioc, ok := IOCFromThreatFoxExport(tc.val, tc.typ)
		if !ok {
			t.Fatalf("expected ok for %q %q", tc.val, tc.typ)
		}
		if ioc.Type != tc.want {
			t.Fatalf("%q %q: got type %s want %s", tc.val, tc.typ, ioc.Type, tc.want)
		}
	}
}
