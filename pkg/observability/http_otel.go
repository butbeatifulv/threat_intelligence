package observability

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// ChainHTTP applies RED metrics and optional OTEL HTTP instrumentation.
func ChainHTTP(service string, next http.Handler) http.Handler {
	chain := InstrumentHTTP(service, next)
	if Enabled() {
		return otelhttp.NewHandler(chain, service)
	}
	return chain
}
