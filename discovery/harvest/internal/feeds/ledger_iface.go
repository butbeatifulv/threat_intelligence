package feeds

import (
	"context"
	"time"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"
)

// CrawlLedger is the ledger surface used by FetchIfDue (*ledger.Store implements it).
type CrawlLedger interface {
	ShouldFetch(ctx context.Context, key string, policy ledger.FetchPolicy, minRefetch time.Duration, force bool) (bool, error)
	GetContentSHA(ctx context.Context, key string) (string, error)
	RecordFetch(ctx context.Context, key, source, url string, policy ledger.FetchPolicy, contentSHA256 string) error
}
