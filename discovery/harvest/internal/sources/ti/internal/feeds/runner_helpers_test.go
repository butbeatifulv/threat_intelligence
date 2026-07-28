package feeds

import (
	"testing"
	"time"
)

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("a", "b"); got != "a" {
		t.Fatalf("got %q", got)
	}
	if got := firstNonEmpty("  ", "b"); got != "b" {
		t.Fatalf("got %q", got)
	}
}

func TestParseDelayEnv(t *testing.T) {
	def := 2 * time.Second
	if got := parseDelayEnv("", def); got != def {
		t.Fatalf("empty: got %v", got)
	}
	if got := parseDelayEnv("500ms", def); got != 500*time.Millisecond {
		t.Fatalf("duration: got %v", got)
	}
	if got := parseDelayEnv("1500", def); got != 1500*time.Millisecond {
		t.Fatalf("ms int: got %v", got)
	}
	if got := parseDelayEnv("bogus", def); got != def {
		t.Fatalf("invalid: got %v", got)
	}
}

func TestStripHTML(t *testing.T) {
	in := "<p>hello <b>world</b></p>"
	if got := StripHTML(in); got != "hello world" {
		t.Fatalf("got %q", got)
	}
	if got := StripHTML("plain"); got != "plain" {
		t.Fatalf("got %q", got)
	}
}
