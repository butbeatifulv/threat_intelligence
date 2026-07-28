package domain

import "testing"

func TestIOC_NodeID_stable(t *testing.T) {
	ioc := IOC{Type: IOCIP, Value: "203.0.113.1"}
	id1 := ioc.NodeID()
	id2 := ioc.NodeID()
	if id1 == "" || id1 != id2 {
		t.Fatalf("NodeID not stable: %q %q", id1, id2)
	}
	if len(id1) != 64 {
		t.Fatalf("expected sha256 hex length 64, got %d", len(id1))
	}
}

func TestIOC_NodeID_differsByType(t *testing.T) {
	a := IOC{Type: IOCIP, Value: "203.0.113.1"}
	b := IOC{Type: IOCDomain, Value: "203.0.113.1"}
	if a.NodeID() == b.NodeID() {
		t.Fatal("same value different type should yield different NodeID")
	}
}

func TestIOC_NodeID_differsByValue(t *testing.T) {
	a := IOC{Type: IOCIP, Value: "203.0.113.1"}
	b := IOC{Type: IOCIP, Value: "203.0.113.2"}
	if a.NodeID() == b.NodeID() {
		t.Fatal("different values should yield different NodeID")
	}
}
