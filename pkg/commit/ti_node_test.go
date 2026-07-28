package commit

import "testing"

func TestIOCNodeID(t *testing.T) {
	id, err := IOCNodeID("ti:ioc:abc123")
	if err != nil {
		t.Fatal(err)
	}
	if id != "abc123" {
		t.Fatalf("got %q want abc123", id)
	}
	if id2, err := IOCNodeID(TIIoCIdempotencyKey("abc123")); err != nil || id2 != id {
		t.Fatalf("stability: got %q err %v", id2, err)
	}
}

func TestIOCNodeID_errors(t *testing.T) {
	cases := []string{"ti:kev:CVE-1", "ti:ioc:", "ioc:abc"}
	for _, key := range cases {
		if _, err := IOCNodeID(key); err == nil {
			t.Fatalf("expected error for %q", key)
		}
	}
}

func TestActorNodeID(t *testing.T) {
	id, err := ActorNodeID("ti:actor:actorhash")
	if err != nil {
		t.Fatal(err)
	}
	if id != "actorhash" {
		t.Fatalf("got %q want actorhash", id)
	}
	if id2, err := ActorNodeID(TIActorIdempotencyKey("actorhash")); err != nil || id2 != id {
		t.Fatalf("stability: got %q err %v", id2, err)
	}
}

func TestActorNodeID_errors(t *testing.T) {
	if _, err := ActorNodeID("ti:actor:"); err == nil {
		t.Fatal("expected error for empty suffix")
	}
	if _, err := ActorNodeID("ti:campaign:c1"); err == nil {
		t.Fatal("expected prefix mismatch error")
	}
}

func TestReportNodeID(t *testing.T) {
	id, err := ReportNodeID("ti:report:report-stable-1")
	if err != nil {
		t.Fatal(err)
	}
	if id != "report-stable-1" {
		t.Fatalf("got %q want report-stable-1", id)
	}
	if id2, err := ReportNodeID(TIReportIdempotencyKey("report-stable-1")); err != nil || id2 != id {
		t.Fatalf("stability: got %q err %v", id2, err)
	}
}

func TestReportNodeID_errors(t *testing.T) {
	if _, err := ReportNodeID("ti:report:"); err == nil {
		t.Fatal("expected error")
	}
}

func TestIOCLinkNodeID(t *testing.T) {
	cases := []struct {
		key, want string
	}{
		{"ti:lc:camp1:deadbeef", "deadbeef"},
		{"ti:lrmi:report-1:deadbeef", "deadbeef"},
		{TILinkCampaignIOCIdempotencyKey("camp1", "deadbeef"), "deadbeef"},
		{TILinkReportMentionsIOCIdempotencyKey("report-1", "deadbeef"), "deadbeef"},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			id, err := IOCLinkNodeID(tc.key)
			if err != nil {
				t.Fatal(err)
			}
			if id != tc.want {
				t.Fatalf("got %q want %q", id, tc.want)
			}
			// stability
			id2, err := IOCLinkNodeID(tc.key)
			if err != nil || id2 != id {
				t.Fatalf("unstable: %q vs %q", id2, id)
			}
		})
	}
}

func TestIOCLinkNodeID_errors(t *testing.T) {
	cases := []string{"nocolon", "ti:lc:camp:", ""}
	for _, key := range cases {
		if _, err := IOCLinkNodeID(key); err == nil {
			t.Fatalf("expected error for %q", key)
		}
	}
}

func TestTILinkSuffix(t *testing.T) {
	cases := []struct {
		key, want string
	}{
		{"ti:lca:camp:actorhash", "actorhash"},
		{"ti:lkc:cluster:campaign", "campaign"},
		{TILinkClusterCampaignIdempotencyKey("cluster", "campaign"), "campaign"},
		{TILinkCampaignActorIdempotencyKey("camp", "actorhash"), "actorhash"},
		{"nocolon", ""},
		{"", ""},
		{"ti:lc:", ""},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			got := TILinkSuffix(tc.key)
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
			if TILinkSuffix(tc.key) != got {
				t.Fatal("TILinkSuffix not stable")
			}
		})
	}
}
