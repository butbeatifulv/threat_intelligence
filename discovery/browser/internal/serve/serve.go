package serve

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	discbrowser "github.com/butbeautifulv/veil/discovery/pkg/browser"
	connats "github.com/butbeautifulv/veil/discovery/connector/nats"
	"github.com/butbeautifulv/veil/pkg/harvest"
)

// Config for the discovery browser HTTP service.
type Config struct {
	Addr          string
	Sidecar       *discbrowser.Sidecar
	HarvestPub    discbrowser.HarvestPublisher
	HarvestSource string
	Logger        *slog.Logger
}

// Handler returns the HTTP handler for /health, /inspect, /exec.
func Handler(cfg Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"service": "discovery-browser",
			"sidecar": cfg.Sidecar != nil && cfg.Sidecar.Enabled(),
		})
	})
	mux.HandleFunc("POST /inspect", func(w http.ResponseWriter, r *http.Request) {
		writeInspect(w, r, cfg)
	})
	mux.HandleFunc("POST /exec", func(w http.ResponseWriter, r *http.Request) {
		writeExec(w, r, cfg)
	})
	return mux
}

func writeInspect(w http.ResponseWriter, r *http.Request, cfg Config) {
	body, err := readJSONBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	url := inspectURL(body)
	if url == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "error": "url or target required"})
		return
	}
	req := inspectReqFromBody(url, body)
	if cfg.Sidecar == nil || !cfg.Sidecar.Enabled() {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(discbrowser.InspectResult{
			Success: false,
			Error:   "browser sidecar not configured (DISCOVERY_BROWSER_SIDECAR_URL)",
		})
		return
	}
	out := cfg.Sidecar.Inspect(r.Context(), req)
	if out.Success && cfg.HarvestPub != nil {
		if err := discbrowser.PublishInspect(r.Context(), cfg.HarvestPub, url, out); err != nil && cfg.Logger != nil {
			cfg.Logger.Warn("harvest publish failed", slog.String("url", url), slog.String("err", err.Error()))
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func writeExec(w http.ResponseWriter, r *http.Request, cfg Config) {
	body, err := readJSONBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	target := inspectURL(body)
	if cfg.Sidecar == nil || !cfg.Sidecar.Enabled() {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stdout": "", "stderr": "browser sidecar not configured", "exit_code": 1, "error": "sidecar not configured",
		})
		return
	}
	out := cfg.Sidecar.Inspect(r.Context(), discbrowser.InspectRequest{
		URL: target, WaitTime: 2, Headless: true, Screenshot: bodyBool(body, "screenshot", true),
	})
	if !out.Success {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stdout": "", "stderr": out.Error, "exit_code": 1, "error": out.Error,
		})
		return
	}
	raw, _ := json.Marshal(out)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"stdout": string(raw) + "\n", "stderr": "", "exit_code": 0, "error": "",
	})
}

func inspectURL(body map[string]any) string {
	if body == nil {
		return ""
	}
	if u := strings.TrimSpace(toString(body["url"])); u != "" {
		return u
	}
	return strings.TrimSpace(toString(body["target"]))
}

func inspectReqFromBody(url string, body map[string]any) discbrowser.InspectRequest {
	params := map[string]string{}
	for k, v := range body {
		if s, ok := v.(string); ok {
			params[k] = s
		}
	}
	req := discbrowser.InspectFromParams(url, params)
	if body != nil {
		if v, ok := body["wait_time"]; ok {
			switch n := v.(type) {
			case float64:
				req.WaitTime = int(n)
			case int:
				req.WaitTime = n
			}
		}
		if body["active_tests"] == true {
			req.ActiveTests = true
		}
		if body["screenshot"] == false {
			req.Screenshot = false
		}
	}
	return req
}

func bodyBool(body map[string]any, key string, def bool) bool {
	if body == nil {
		return def
	}
	v, ok := body[key]
	if !ok {
		return def
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return def
}

func readJSONBody(r *http.Request) (map[string]any, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	return body, nil
}

func toString(v any) string {
	s, _ := v.(string)
	return s
}

// LoadConfigFromEnv builds service config from environment.
func LoadConfigFromEnv(logger *slog.Logger) (Config, func(), error) {
	addr := strings.TrimSpace(os.Getenv("DISCOVERY_BROWSER_LISTEN"))
	if addr == "" {
		addr = ":8920"
	}
	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}
	cfg := Config{
		Addr:    addr,
		Sidecar: discbrowser.NewSidecarFromEnv(),
		Logger:  logger,
	}
	var cleanup func()
	natsURL := strings.TrimSpace(os.Getenv("NATS_URL"))
	if natsURL != "" {
		pub, err := connats.ConnectJetStreamAndStream(natsURL)
		if err != nil {
			return Config{}, nil, err
		}
		subject := strings.TrimSpace(os.Getenv("BROWSER_SCRAPE_SUBJECT"))
		if subject == "" {
			subject = "scrape.browser.events"
		}
		cfg.HarvestPub = connats.NewDomainPublisher(pub, harvest.SourceBrowser, subject)
		cleanup = func() { pub.Close() }
	}
	return cfg, cleanup, nil
}
