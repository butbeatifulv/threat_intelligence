package embed

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllama_Embed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"embedding": []float32{0.1, 0.2, 0.3},
		})
	}))
	defer srv.Close()
	o := &Ollama{BaseURL: srv.URL, Model: "test"}
	vecs, err := o.Embed(context.Background(), []string{"hello"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != 1 || len(vecs[0]) != 3 {
		t.Fatalf("unexpected vectors: %+v", vecs)
	}
}

func TestNoop_EmbedFails(t *testing.T) {
	_, err := NoopProvider{}.Embed(context.Background(), []string{"x"})
	if err == nil {
		t.Fatal("expected error")
	}
}
