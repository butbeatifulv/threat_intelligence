package normalize

import (
	"testing"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
	"github.com/butbeautifulv/veil/pkg/ti/ids"
)

func TestNormalizeIOC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		in     domain.IOC
		want   domain.IOC
		wantOK bool
	}{
		{
			name:   "ip trims and canonicalizes",
			in:     domain.IOC{Type: domain.IOCIP, Value: " 203.0.113.1 ", Source: "feed"},
			want:   domain.IOC{Type: domain.IOCIP, Value: "203.0.113.1", Source: "feed", Sources: []string{"feed"}},
			wantOK: true,
		},
		{name: "ip invalid", in: domain.IOC{Type: domain.IOCIP, Value: "not-an-ip"}, wantOK: false},
		{
			name:   "domain lowercases and strips trailing dot",
			in:     domain.IOC{Type: domain.IOCDomain, Value: " Evil.COM. ", Source: "a", Sources: []string{"b", "a"}},
			want:   domain.IOC{Type: domain.IOCDomain, Value: "evil.com", Source: "a", Sources: []string{"a", "b"}},
			wantOK: true,
		},
		{name: "domain rejects path-like value", in: domain.IOC{Type: domain.IOCDomain, Value: "evil.com/foo"}, wantOK: false},
		{
			name:   "url normalizes scheme host and strips https :443",
			in:     domain.IOC{Type: domain.IOCURL, Value: "HTTPS://Evil.COM:443/path"},
			want:   domain.IOC{Type: domain.IOCURL, Value: "https://evil.com/path"},
			wantOK: true,
		},
		{name: "url missing host", in: domain.IOC{Type: domain.IOCURL, Value: "https:///path"}, wantOK: false},
		{
			name:   "hash accepts sha256",
			in:     domain.IOC{Type: domain.IOCHash, Value: " ABCDEF0123456789abcdef0123456789abcdef0123456789abcdef0123456789 "},
			want:   domain.IOC{Type: domain.IOCHash, Value: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"},
			wantOK: true,
		},
		{name: "hash rejects bad length", in: domain.IOC{Type: domain.IOCHash, Value: "abc123"}, wantOK: false},
		{name: "unknown type", in: domain.IOC{Type: "email", Value: "a@b.com"}, wantOK: false},
		{name: "empty value", in: domain.IOC{Type: domain.IOCIP, Value: "  "}, wantOK: false},
		{
			name:   "sources skip blanks and dedupe",
			in:     domain.IOC{Type: domain.IOCIP, Value: "203.0.113.1", Source: "a", Sources: []string{"", "b", "a"}},
			want:   domain.IOC{Type: domain.IOCIP, Value: "203.0.113.1", Source: "a", Sources: []string{"a", "b"}},
			wantOK: true,
		},
		{
			name:   "url strips http default port",
			in:     domain.IOC{Type: domain.IOCURL, Value: "http://Evil.COM:80/path"},
			want:   domain.IOC{Type: domain.IOCURL, Value: "http://evil.com/path"},
			wantOK: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := NormalizeIOC(tt.in)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v want %v", ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			if got.Type != tt.want.Type || got.Value != tt.want.Value || got.Source != tt.want.Source {
				t.Fatalf("got %#v want %#v", got, tt.want)
			}
			if len(got.Sources) != len(tt.want.Sources) {
				t.Fatalf("sources %#v want %#v", got.Sources, tt.want.Sources)
			}
			for i := range got.Sources {
				if got.Sources[i] != tt.want.Sources[i] {
					t.Fatalf("sources[%d] = %q want %q", i, got.Sources[i], tt.want.Sources[i])
				}
			}
		})
	}
}

func TestNormalizeCampaign(t *testing.T) {
	c := domain.Campaign{
		ID: " c1 ", Name: " Camp ", Summary: " s ", Source: " src ",
		Actors: []string{" Alice ", "Bob"},
	}
	got := NormalizeCampaign(c)
	if got.ID != "c1" || got.Name != "Camp" || got.Actors[0] != "Alice" {
		t.Fatalf("got %#v", got)
	}
}

func TestNormalizeCluster(t *testing.T) {
	cl := NormalizeCluster(domain.Cluster{ID: " c1 ", Name: " N ", Description: " d ", Source: " s "})
	if cl.ID != "c1" || cl.Name != "N" || cl.Description != "d" || cl.Source != "s" {
		t.Fatalf("got %#v", cl)
	}
}

func TestActorStableID_and_ReportStableID(t *testing.T) {
	if ActorStableID(" APT29 ") == "" {
		t.Fatal("actor id empty")
	}
	if ReportStableID(" https://r.example ") == "" {
		t.Fatal("report id empty")
	}
	if ActorStableID("") != ids.ActorStableID("") {
		t.Fatal("forward mismatch")
	}
}

func TestCanonicalID_matchesNodeID(t *testing.T) {
	ioc, ok := NormalizeIOC(domain.IOC{Type: domain.IOCIP, Value: "203.0.113.1"})
	if !ok {
		t.Fatal("normalize failed")
	}
	if CanonicalID(ioc) != ids.CanonicalIOCID(ioc) || CanonicalID(ioc) != ioc.NodeID() {
		t.Fatal("id mismatch")
	}
}

func TestNormalizeValidatedIOC_unknownType(t *testing.T) {
	_, ok := normalizeValidatedIOC(domain.IOC{Type: domain.IOCType("other"), Value: "x"})
	if ok {
		t.Fatal("expected false for unhandled type")
	}
}
