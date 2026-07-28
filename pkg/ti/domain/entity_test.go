package domain

import (
	"encoding/json"
	"testing"
)

func TestIOC_JSONRoundTrip(t *testing.T) {
	in := IOC{
		Type:  IOCDomain,
		Value: "evil.example",
		Tags:  []string{"c2"},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out IOC
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.Type != in.Type || out.Value != in.Value || len(out.Tags) != 1 {
		t.Fatalf("got %+v", out)
	}
}

func TestIOCType_constants(t *testing.T) {
	cases := []IOCType{IOCIP, IOCDomain, IOCURL, IOCHash}
	for _, c := range cases {
		if string(c) == "" {
			t.Fatalf("empty IOCType constant")
		}
	}
}

func TestActor_Campaign_Report_zeroSafe(t *testing.T) {
	var a Actor
	var c Campaign
	var r Report
	if a.Name != "" || c.Name != "" || r.Title != "" {
		t.Fatal("zero values should be empty")
	}
}
