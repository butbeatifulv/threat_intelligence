package config

import (
	"fmt"
	"net/url"
	"os"
)

const serviceName = "veil-gateway"

// Config holds upstream targets and listen address for the HTTP edge proxy.
type Config struct {
	ListenAddr   string
	GraphAPIURL  *url.URL
	EngageAPIURL *url.URL
	GraphMCPURL  *url.URL
	EngageMCPURL *url.URL
}

// Load reads gateway settings from the environment.
func Load() (Config, error) {
	listen := envOr("VEIL_GATEWAY_LISTEN", ":8080")
	graphAPI, err := parseURL("VEIL_GRAPH_API_URL", "http://127.0.0.1:8090")
	if err != nil {
		return Config{}, err
	}
	engageAPI, err := parseURL("VEIL_ENGAGE_API_URL", "http://127.0.0.1:8890")
	if err != nil {
		return Config{}, err
	}
	graphMCP, err := parseURL("VEIL_GRAPH_MCP_URL", "http://127.0.0.1:8091")
	if err != nil {
		return Config{}, err
	}
	engageMCP, err := parseURL("VEIL_ENGAGE_MCP_URL", "http://127.0.0.1:8892")
	if err != nil {
		return Config{}, err
	}
	return Config{
		ListenAddr:   listen,
		GraphAPIURL:  graphAPI,
		EngageAPIURL: engageAPI,
		GraphMCPURL:  graphMCP,
		EngageMCPURL: engageMCP,
	}, nil
}

// ServiceName is the gateway service identifier for health responses.
func ServiceName() string { return serviceName }

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseURL(key, fallback string) (*url.URL, error) {
	raw := envOr(key, fallback)
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("%s: must be an absolute URL with scheme and host", key)
	}
	return u, nil
}
