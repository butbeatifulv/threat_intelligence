package browser

import (
	"context"
	"encoding/json"

	"github.com/butbeautifulv/veil/pkg/harvest"
)

// HarvestPublisher publishes browser inspect events to scrape.>.
type HarvestPublisher interface {
	Publish(ctx context.Context, kind, contentKey string, payload any) error
}

// PublishInspect emits harvest.KindBrowserInspectRaw when inspect succeeded.
func PublishInspect(ctx context.Context, pub HarvestPublisher, url string, result InspectResult) error {
	if pub == nil || !result.Success {
		return nil
	}
	raw, err := json.Marshal(result)
	if err != nil {
		return err
	}
	pl := harvest.BrowserInspectRaw{
		URL:       url,
		RawJSON:   string(raw),
		Timestamp: result.Timestamp,
	}
	return pub.Publish(ctx, harvest.KindBrowserInspectRaw, harvest.BrowserContentKey(url), pl)
}
