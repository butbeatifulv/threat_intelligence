package feeds

import (
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
)

func TestExtractIOCsFromText_dedupes(t *testing.T) {
	sha := strings.Repeat("a", 64)
	text := "seen " + sha + " again " + sha
	out := ExtractIOCsFromText(text)
	var hashes int
	for _, ioc := range out {
		if ioc.Type == domain.IOCHash && ioc.Value == sha {
			hashes++
		}
	}
	if hashes != 1 {
		t.Fatalf("expected one sha256, got %d from %+v", hashes, out)
	}
}

func TestExtractIOCsFromText_mixedTypes(t *testing.T) {
	text := "ip 203.0.113.10 url https://evil.example/x md5 deadbeefdeadbeefdeadbeefdeadbeef"
	out := ExtractIOCsFromText(text)
	types := map[domain.IOCType]bool{}
	for _, ioc := range out {
		types[ioc.Type] = true
	}
	for _, want := range []domain.IOCType{domain.IOCIP, domain.IOCURL, domain.IOCHash} {
		if !types[want] {
			t.Fatalf("missing %s in %+v", want, out)
		}
	}
}
