package proxypool

import (
	"strings"
	"testing"
	"time"
)

func TestSplitEnvList(t *testing.T) {
	if SplitEnvList("") != nil {
		t.Fatal("empty")
	}
	got := SplitEnvList(" http://a:1 , http://b:2\nhttp://c:3 ")
	if len(got) != 3 || got[0] != "http://a:1" {
		t.Fatalf("%v", got)
	}
}

func TestNew_rejectsEmptyAndBadScheme(t *testing.T) {
	if _, err := New(nil, time.Minute); err == nil {
		t.Fatal("nil list")
	}
	if _, err := New([]string{"socks5://x"}, time.Minute); err == nil {
		t.Fatal("bad scheme")
	}
	if _, err := New([]string{"not-a-url"}, time.Minute); err == nil {
		t.Fatal("bad url")
	}
}

func TestNew_ok(t *testing.T) {
	p, err := New([]string{"http://127.0.0.1:8080"}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if p == nil {
		t.Fatal("nil pool")
	}
}

func TestSplitEnvList_trimsCommas(t *testing.T) {
	got := SplitEnvList(",,  http://h:1  ,,")
	if len(got) != 1 || !strings.HasPrefix(got[0], "http://") {
		t.Fatalf("%v", got)
	}
}
