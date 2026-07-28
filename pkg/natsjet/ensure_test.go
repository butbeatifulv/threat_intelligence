package natsjet

import (
	"testing"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	natsgo "github.com/nats-io/nats.go"
)

func jetStreamForTest(t *testing.T) natsgo.JetStreamContext {
	t.Helper()
	opts := &server.Options{JetStream: true, StoreDir: t.TempDir(), Port: -1}
	srv := test.RunServer(opts)
	t.Cleanup(srv.Shutdown)
	nc, err := natsgo.Connect(srv.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(nc.Close)
	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	return js
}

func TestEnsureStream_createsAndWidens(t *testing.T) {
	js := jetStreamForTest(t)
	if err := EnsureStream(js, "EVENTS", []string{"events.ti.>"}); err != nil {
		t.Fatal(err)
	}
	if err := EnsureStream(js, "EVENTS", []string{"events.ti.>", "events.vuln.>"}); err != nil {
		t.Fatal(err)
	}
	info, err := js.StreamInfo("EVENTS")
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Config.Subjects) != 2 {
		t.Fatalf("subjects %#v", info.Config.Subjects)
	}
}

func TestEnsureStream_idempotentNoWiden(t *testing.T) {
	js := jetStreamForTest(t)
	subs := []string{"events.ti.>"}
	if err := EnsureStream(js, "EVENTS2", subs); err != nil {
		t.Fatal(err)
	}
	if err := EnsureStream(js, "EVENTS2", subs); err != nil {
		t.Fatal(err)
	}
}
