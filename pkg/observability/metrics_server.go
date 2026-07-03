package observability

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// StartMetricsServer serves GET /metrics until ctx is canceled.
func StartMetricsServer(ctx context.Context, listen string, log *slog.Logger) {
	if listen == "" {
		listen = ":9090"
	}
	if log == nil {
		log = slog.Default()
	}
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	srv := &http.Server{
		Addr:              listen,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shCtx)
	}()
	go func() {
		log.Info("metrics server listening", slog.String("addr", listen))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Warn("metrics server stopped", slog.String("err", err.Error()))
		}
	}()
}
