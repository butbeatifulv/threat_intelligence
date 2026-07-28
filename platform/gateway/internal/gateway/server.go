package gateway

import (
	"net/http"

	"github.com/butbeautifulv/veil/platform/gateway/internal/config"
)

// NewMux returns a handler mux that proxies graph and engage APIs and exposes composite health.
func NewMux(cfg config.Config) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/v1/", newReverseProxy(cfg.GraphAPIURL))
	mux.Handle("/api/", newReverseProxy(cfg.EngageAPIURL))
	registerCompositeHealth(mux, cfg)
	return mux
}
