package usecase

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"

	scrapefeeds "github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/ti/internal/feeds"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/ti/internal/repository"
	"github.com/butbeautifulv/veil/pkg/ti/domain"
)

const kevURL = "https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json"

const threatFoxExportURL = "https://threatfox.abuse.ch/export/json/recent/"
const threatFoxAPIURL = "https://threatfox-api.abuse.ch/api/v1/"

// Runner orchestrates TI feed fetch → parse → repository upsert.
type Runner struct {
	Repo    repository.GraphRepository
	Logger  *slog.Logger
	Fetcher *feeds.Fetcher
}

func NewRunner(repo repository.GraphRepository, logger *slog.Logger, fc *scrapefeeds.Client, led scrapefeeds.CrawlLedger) *Runner {
	return &Runner{
		Repo:    repo,
		Logger:  logger,
		Fetcher: feeds.NewFetcher(logger, fc, led),
	}
}

func (r *Runner) Run(ctx context.Context, kinds []string) error {
	if err := r.Repo.EnsureSchema(ctx); err != nil {
		return err
	}
	for _, k := range kinds {
		switch strings.TrimSpace(strings.ToLower(k)) {
		case "kev":
			if err := r.runKEV(ctx); err != nil {
				return err
			}
		case "pt":
			if err := r.runPTRSS(ctx); err != nil {
				return err
			}
		case "urlhaus":
			if err := r.runURLhaus(ctx); err != nil {
				return err
			}
		case "threatfox":
			if err := r.runThreatFox(ctx); err != nil {
				return err
			}
		case "malwarebazaar":
			if err := r.runMalwareBazaar(ctx); err != nil {
				return err
			}
		case "feodo":
			if err := r.runFeodo(ctx); err != nil {
				return err
			}
		case "openphish":
			if err := r.runOpenPhish(ctx); err != nil {
				return err
			}
		default:
			r.Logger.Info("unknown feed kind", slog.String("kind", k))
		}
	}
	return nil
}

type kevFile struct {
	Vulnerabilities []struct {
		CVEID         string `json:"cveID"`
		VendorProject string `json:"vendorProject"`
		Product       string `json:"product"`
		ShortDesc     string `json:"shortDescription"`
		DateAdded     string `json:"dateAdded"`
	} `json:"vulnerabilities"`
}

func (r *Runner) runKEV(ctx context.Context) error {
	r.Logger.Info("ingesting CISA KEV")
	res, err := r.Fetcher.FetchLedger(ctx, "ti:kev", kevURL, "ti/kev.json", ledger.PolicyDaily)
	if err != nil {
		return err
	}
	if res.Unchanged {
		r.Logger.Info("KEV unchanged, skip publish")
		return nil
	}
	if res.Skipped && len(res.Body) == 0 {
		return fmt.Errorf("ti:kev skipped by ledger without cache")
	}
	b := res.Body
	var doc kevFile
	if err := json.Unmarshal(b, &doc); err != nil {
		return err
	}
	limitN := len(doc.Vulnerabilities)
	if v := os.Getenv("TI_KEV_MAX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n < limitN {
			limitN = n
		}
	}
	for i := 0; i < limitN; i++ {
		v := doc.Vulnerabilities[i]
		if err := r.Repo.UpsertKEVVulnerability(ctx, v.CVEID, v.VendorProject, v.Product, v.ShortDesc, v.DateAdded); err != nil {
			return err
		}
	}
	r.Logger.Info("KEV ingest done", slog.Int("count", limitN))
	return nil
}

type rss struct {
	Channel struct {
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

func (r *Runner) runPTRSS(ctx context.Context) error {
	u := os.Getenv("PT_RSS_URL")
	if u == "" {
		u = "https://www.ptsecurity.com/rss/all.xml"
	}
	r.Logger.Info("ingesting PT RSS", slog.String("url", u))
	res, err := r.Fetcher.FetchLedger(ctx, "ti:pt:rss", u, "ti/pt.xml", ledger.PolicyDaily)
	if err != nil {
		return err
	}
	if res.Unchanged {
		r.Logger.Info("PT RSS unchanged, skip publish")
		return nil
	}
	if res.Skipped && len(res.Body) == 0 {
		return fmt.Errorf("ti:pt skipped by ledger without cache")
	}
	b := res.Body
	var doc rss
	if err := xml.Unmarshal(b, &doc); err != nil {
		return err
	}
	limitN := len(doc.Channel.Items)
	if v := os.Getenv("TI_PT_MAX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n < limitN {
			limitN = n
		}
	}
	n := 0
	for i := 0; i < limitN; i++ {
		it := doc.Channel.Items[i]
		if !strings.Contains(strings.ToLower(it.Link), "ptsecurity") && !strings.Contains(strings.ToLower(it.Title), "pt") {
			continue
		}
		rep := domain.Report{
			Title:        it.Title,
			Provider:     "Positive Technologies",
			Link:         it.Link,
			PublishedAt:  it.PubDate,
			BodyMarkdown: feeds.StripHTML(it.Description),
			Source:       "pt-rss",
		}
		if err := r.Repo.UpsertReport(ctx, rep); err != nil {
			return err
		}
		for _, ioc := range feeds.ExtractIOCsFromText(it.Title + "\n" + it.Description + "\n" + rep.BodyMarkdown) {
			if err := r.Repo.UpsertIOC(ctx, ioc); err != nil {
				return err
			}
		}
		n++
	}
	r.Logger.Info("PT RSS ingest done", slog.Int("reports", n))
	return nil
}

func (r *Runner) runURLhaus(ctx context.Context) error {
	u := "https://urlhaus.abuse.ch/downloads/csv_recent/"
	r.Logger.Info("ingesting URLhaus recent CSV", slog.String("url", u))
	res, err := r.Fetcher.FetchLedger(ctx, "ti:urlhaus:recent", u, "ti/urlhaus_recent.csv", ledger.PolicyDaily)
	if err != nil {
		return err
	}
	if res.Unchanged {
		r.Logger.Info("URLhaus unchanged, skip publish")
		return nil
	}
	if res.Skipped && len(res.Body) == 0 {
		return fmt.Errorf("ti:urlhaus skipped by ledger without cache")
	}
	b := res.Body
	lines := strings.Split(string(b), "\n")
	max := 500
	if v := os.Getenv("TI_URLHAUS_MAX"); v != "" {
		var n int
		_, _ = fmt.Sscanf(v, "%d", &n)
		if n > 0 {
			max = n
		}
	}
	count := 0
	for _, ln := range lines {
		if count >= max {
			break
		}
		if strings.HasPrefix(ln, "#") || strings.TrimSpace(ln) == "" {
			continue
		}
		parts := strings.Split(ln, `","`)
		if len(parts) < 3 {
			parts = strings.Split(ln, ",")
		}
		var urlStr string
		for _, p := range parts {
			p = strings.Trim(p, `"`)
			if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
				urlStr = p
				break
			}
		}
		if urlStr == "" {
			continue
		}
		ioc := domain.IOC{Type: domain.IOCURL, Value: urlStr, Source: "urlhaus"}
		if err := r.Repo.UpsertIOC(ctx, ioc); err != nil {
			return err
		}
		count++
	}
	r.Logger.Info("URLhaus ingest done", slog.Int("iocs", count))
	return nil
}

func (r *Runner) runThreatFox(ctx context.Context) error {
	if k := strings.TrimSpace(os.Getenv("THREATFOX_AUTH_KEY")); k != "" {
		return r.runThreatFoxAPI(ctx, k)
	}
	return r.runThreatFoxExport(ctx)
}

func (r *Runner) runThreatFoxAPI(ctx context.Context, authKey string) error {
	days := 3
	if v := os.Getenv("TI_THREATFOX_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 7 {
			days = n
		}
	}
	body := fmt.Sprintf(`{"query":"get_iocs","days":%d}`, days)
	cacheRel := fmt.Sprintf("ti/threatfox_api_days_%d.json", days)
	key := fmt.Sprintf("ti:threatfox:api:days:%d", days)
	r.Logger.Info("ingesting ThreatFox API", slog.Int("days", days))
	res, err := r.Fetcher.FetchLedgerPOST(ctx, key, threatFoxAPIURL, cacheRel, ledger.PolicyDaily, authKey, []byte(body))
	if err != nil {
		return err
	}
	if res.Unchanged {
		r.Logger.Info("ThreatFox API unchanged, skip publish")
		return nil
	}
	if res.Skipped && len(res.Body) == 0 {
		return fmt.Errorf("ti:threatfox:api skipped by ledger without cache")
	}
	b := res.Body
	var doc struct {
		QueryStatus string `json:"query_status"`
		Data        []struct {
			IOC     string `json:"ioc"`
			IOCType string `json:"ioc_type"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return err
	}
	if doc.QueryStatus != "ok" && doc.QueryStatus != "no_results" {
		return fmt.Errorf("threatfox query_status=%s", doc.QueryStatus)
	}
	max := 4000
	if v := os.Getenv("TI_THREATFOX_MAX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			max = n
		}
	}
	count := 0
	for _, row := range doc.Data {
		if count >= max {
			break
		}
		ioc, ok := feeds.IOCFromThreatFoxExport(row.IOC, row.IOCType)
		if !ok {
			continue
		}
		if err := r.Repo.UpsertIOC(ctx, ioc); err != nil {
			return err
		}
		count++
	}
	r.Logger.Info("ThreatFox API ingest done", slog.Int("iocs", count))
	return nil
}

func (r *Runner) runThreatFoxExport(ctx context.Context) error {
	r.Logger.Info("ingesting ThreatFox public export", slog.String("url", threatFoxExportURL))
	res, err := r.Fetcher.FetchLedger(ctx, "ti:threatfox:export", threatFoxExportURL, "ti/threatfox_recent.json", ledger.PolicyDaily)
	if err != nil {
		return err
	}
	if res.Unchanged {
		r.Logger.Info("ThreatFox export unchanged, skip publish")
		return nil
	}
	if res.Skipped && len(res.Body) == 0 {
		return fmt.Errorf("ti:threatfox:export skipped by ledger without cache")
	}
	b := res.Body
	var root map[string]json.RawMessage
	if err := json.Unmarshal(b, &root); err != nil {
		return err
	}
	max := 4000
	if v := os.Getenv("TI_THREATFOX_MAX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			max = n
		}
	}
	count := 0
outer:
	for _, chunk := range root {
		var rows []struct {
			IOCValue string `json:"ioc_value"`
			IOCType  string `json:"ioc_type"`
		}
		if err := json.Unmarshal(chunk, &rows); err != nil {
			continue
		}
		for _, row := range rows {
			if count >= max {
				break outer
			}
			ioc, ok := feeds.IOCFromThreatFoxExport(row.IOCValue, row.IOCType)
			if !ok {
				continue
			}
			if err := r.Repo.UpsertIOC(ctx, ioc); err != nil {
				return err
			}
			count++
		}
	}
	r.Logger.Info("ThreatFox ingest done", slog.Int("iocs", count))
	return nil
}

type malwareBazaarRecent struct {
	QueryStatus string `json:"query_status"`
	Data        []struct {
		Sha256 string `json:"sha256_hash"`
		Md5    string `json:"md5_hash"`
	} `json:"data"`
}

func (r *Runner) runMalwareBazaar(ctx context.Context) error {
	key := strings.TrimSpace(os.Getenv("MALWAREBAZAAR_AUTH_KEY"))
	if key == "" {
		key = strings.TrimSpace(os.Getenv("MALWARE_BAZAAR_API_KEY"))
	}
	if key == "" {
		r.Logger.Warn("malwarebazaar feed skipped", slog.String("reason", "set MALWAREBAZAAR_AUTH_KEY (abuse.ch Auth-Key)"))
		return nil
	}
	const mbURL = "https://mb-api.abuse.ch/api/v1/"
	body := []byte(`{"query":"get_recent","selector":"time"}`)
	r.Logger.Info("ingesting MalwareBazaar recent (API)")
	res, err := r.Fetcher.FetchLedgerPOST(ctx, "ti:malwarebazaar:recent", mbURL, "ti/malwarebazaar_recent.json", ledger.PolicyDaily, key, body)
	if err != nil {
		return err
	}
	if res.Unchanged {
		r.Logger.Info("MalwareBazaar unchanged, skip publish")
		return nil
	}
	if res.Skipped && len(res.Body) == 0 {
		return fmt.Errorf("ti:malwarebazaar skipped by ledger without cache")
	}
	b := res.Body
	var doc malwareBazaarRecent
	if err := json.Unmarshal(b, &doc); err != nil {
		return err
	}
	if doc.QueryStatus != "ok" && doc.QueryStatus != "no_results" {
		return fmt.Errorf("malwarebazaar query_status=%s", doc.QueryStatus)
	}
	max := 500
	if v := os.Getenv("TI_MALWAREBAZAAR_MAX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			max = n
		}
	}
	n := 0
	for _, row := range doc.Data {
		if n >= max {
			break
		}
		if row.Sha256 != "" {
			ioc := domain.IOC{Type: domain.IOCHash, Value: row.Sha256, Source: "malwarebazaar"}
			if err := r.Repo.UpsertIOC(ctx, ioc); err != nil {
				return err
			}
			n++
			continue
		}
		if row.Md5 != "" {
			ioc := domain.IOC{Type: domain.IOCHash, Value: row.Md5, Source: "malwarebazaar"}
			if err := r.Repo.UpsertIOC(ctx, ioc); err != nil {
				return err
			}
			n++
		}
	}
	r.Logger.Info("MalwareBazaar ingest done", slog.Int("iocs", n))
	return nil
}

func (r *Runner) runFeodo(ctx context.Context) error {
	u := strings.TrimSpace(os.Getenv("FEODO_BLOCKLIST_URL"))
	if u == "" {
		u = "https://feodotracker.abuse.ch/downloads/ipblocklist_recommended.txt"
	}
	r.Logger.Info("ingesting Feodo Tracker blocklist", slog.String("url", u))
	res, err := r.Fetcher.FetchLedger(ctx, "ti:feodo:blocklist", u, "ti/feodo_ipblocklist.txt", ledger.PolicyDaily)
	if err != nil {
		return err
	}
	if res.Unchanged {
		r.Logger.Info("Feodo unchanged, skip publish")
		return nil
	}
	if res.Skipped && len(res.Body) == 0 {
		return fmt.Errorf("ti:feodo skipped by ledger without cache")
	}
	b := res.Body
	max := 5000
	if v := os.Getenv("TI_FEODO_MAX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			max = n
		}
	}
	count := 0
	for _, ln := range strings.Split(string(b), "\n") {
		if count >= max {
			break
		}
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		if ip := net.ParseIP(ln); ip == nil {
			continue
		}
		ioc := domain.IOC{Type: domain.IOCIP, Value: ln, Source: "feodo"}
		if err := r.Repo.UpsertIOC(ctx, ioc); err != nil {
			return err
		}
		count++
	}
	r.Logger.Info("Feodo ingest done", slog.Int("iocs", count))
	return nil
}

func (r *Runner) runOpenPhish(ctx context.Context) error {
	u := strings.TrimSpace(os.Getenv("OPENPHISH_FEED_URL"))
	if u == "" {
		u = "https://openphish.com/feed.txt"
	}
	r.Logger.Info("ingesting OpenPhish feed", slog.String("url", u))
	res, err := r.Fetcher.FetchLedger(ctx, "ti:openphish:feed", u, "ti/openphish_feed.txt", ledger.PolicyDaily)
	if err != nil {
		r.Logger.Warn("openphish feed skipped after retries", slog.String("err", err.Error()))
		return nil
	}
	if res.Unchanged {
		r.Logger.Info("OpenPhish unchanged, skip publish")
		return nil
	}
	if res.Skipped && len(res.Body) == 0 {
		r.Logger.Warn("openphish skipped by ledger without cache")
		return nil
	}
	b := res.Body
	max := 800
	if v := os.Getenv("TI_OPENPHISH_MAX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			max = n
		}
	}
	count := 0
	for _, ln := range strings.Split(string(b), "\n") {
		if count >= max {
			break
		}
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		ioc := domain.IOC{Type: domain.IOCURL, Value: ln, Source: "openphish"}
		if err := r.Repo.UpsertIOC(ctx, ioc); err != nil {
			return err
		}
		count++
	}
	r.Logger.Info("OpenPhish ingest done", slog.Int("iocs", count))
	return nil
}
