package mcpserver

import (
	"context"
	"testing"

	playbookuc "github.com/butbeautifulv/veil/knowledge/serve/internal/usecase/playbook"
)

func TestHandlePlaybookSearch_diskImaging(t *testing.T) {
	pb, err := playbookuc.NewService()
	if err != nil {
		t.Skip(err)
	}
	srv := &Server{playbook: pb}
	_, err = srv.handlePlaybookSearch(context.Background(), map[string]any{
		"query": "disk imaging",
		"limit": float64(10),
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range pb.Search("disk imaging", "", 10) {
		if s.ID == "acquiring-disk-image-with-dd-and-dcfldd" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected acquiring-disk-image-with-dd-and-dcfldd in search hits")
	}
}
