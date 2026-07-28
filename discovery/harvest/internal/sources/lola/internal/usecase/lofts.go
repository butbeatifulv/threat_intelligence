package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"

	"golang.org/x/net/html"
)

func (u *ScraperUsecase) IngestLOFTS(ctx context.Context) error {
	url := strings.TrimSpace(os.Getenv("LOFTS_URL"))
	if url == "" {
		url = "https://lofts.galeal.com/"
	}
	skip := strings.EqualFold(strings.TrimSpace(os.Getenv("LOFTS_SKIP_ON_ERROR")), "true")
	u.logger.Info("ingesting LOFTS", slog.String("url", url))

	var body []byte
	res, err := feeds.FetchIfDue(ctx, u.feeds, u.ledger, "lola:lofts:index", "lola", url, ledger.PolicyPeriodic, "lofts/index.html", func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "veil-lola/1.0")
		return req, nil
	})
	if err != nil {
		if skip {
			u.logger.Warn("lofts fetch failed; skipping", slog.String("err", err.Error()))
			return nil
		}
		return err
	}
	if res.Unchanged {
		u.logger.Info("LOFTS index unchanged, skip ingest")
		return nil
	}
	if res.Skipped && len(res.Body) == 0 {
		if skip {
			u.logger.Warn("lofts skipped by ledger without cache")
			return nil
		}
		return fmt.Errorf("lola:lofts skipped by ledger without cache")
	}
	body = res.Body

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	entries := extractLoftsLinks(doc)
	maxE := 200
	if v := os.Getenv("LOFTS_MAX_ENTRIES"); v != "" {
		var n int
		_, _ = fmt.Sscanf(v, "%d", &n)
		if n > 0 {
			maxE = n
		}
	}
	n := 0
	for _, e := range entries {
		if n >= maxE {
			break
		}
		if e.href == "" || strings.HasPrefix(e.href, "#") {
			continue
		}
		link := e.href
		if strings.HasPrefix(link, "/") {
			link = strings.TrimRight(url, "/") + link
		}
		md := fmt.Sprintf("# %s\n\n**Link:** %s\n", e.text, link)
		if err := u.repo.UpsertLoftsEntry(ctx, e.text, e.category, link, md); err != nil {
			return err
		}
		n++
	}
	u.logger.Info("LOFTS ingest done", slog.Int("entries", n))
	return nil
}

type loftsLink struct {
	href, text, category string
}

func extractLoftsLinks(n *html.Node) []loftsLink {
	var out []loftsLink
	var walk func(*html.Node, string)
	walk = func(node *html.Node, category string) {
		if node.Type == html.ElementNode && node.Data == "h2" {
			category = textContent(node)
		}
		if node.Type == html.ElementNode && node.Data == "a" {
			var href string
			for _, a := range node.Attr {
				if a.Key == "href" {
					href = strings.TrimSpace(a.Val)
					break
				}
			}
			if href != "" {
				t := strings.TrimSpace(textContent(node))
				if t != "" {
					out = append(out, loftsLink{href: href, text: t, category: category})
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c, category)
		}
	}
	walk(n, "")
	return out
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(b.String())
}
