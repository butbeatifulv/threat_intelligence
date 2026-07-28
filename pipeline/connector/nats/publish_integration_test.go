package nats

import (
	"testing"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	natsgo "github.com/nats-io/nats.go"
)

func startJetStreamTestServer(t *testing.T) natsgo.JetStreamContext {
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

func TestEnsureIngestStream_createsStream(t *testing.T) {
	js := startJetStreamTestServer(t)
	if err := EnsureIngestStream(js); err != nil {
		t.Fatal(err)
	}
	info, err := js.StreamInfo("INGEST")
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Config.Subjects) != 1 || info.Config.Subjects[0] != "ingest.>" {
		t.Fatalf("subjects %#v", info.Config.Subjects)
	}
}

func TestEnsureBothStreams_scrapeAndIngest(t *testing.T) {
	js := startJetStreamTestServer(t)
	if err := EnsureBothStreams(js); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"SCRAPE", "INGEST"} {
		if _, err := js.StreamInfo(name); err != nil {
			t.Fatalf("stream %s: %v", name, err)
		}
	}
}
