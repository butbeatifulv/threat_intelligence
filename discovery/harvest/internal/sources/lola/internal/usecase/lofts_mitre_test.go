package usecase

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestExtractLoftsLinks(t *testing.T) {
	const sample = `<html><body>
<h2>Category A</h2>
<a href="/x/foo">Foo Link</a>
<a href="https://example.com/y">External</a>
</body></html>`
	doc, err := html.Parse(strings.NewReader(sample))
	if err != nil {
		t.Fatal(err)
	}
	links := extractLoftsLinks(doc)
	if len(links) < 2 {
		t.Fatalf("got %d links", len(links))
	}
}

