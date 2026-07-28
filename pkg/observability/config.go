package observability

import (
	"os"
	"strconv"
	"strings"
)

// Config holds OpenTelemetry and logging settings from environment.
type Config struct {
	Enabled       bool
	ServiceName   string
	OTLPEndpoint  string
	SamplerRatio  float64
	MetricsListen string
	LogFormat     string
	LogLevel      string
}

// LoadConfigFromEnv reads observability settings; serviceName defaults OTEL_SERVICE_NAME.
func LoadConfigFromEnv(serviceName string) Config {
	if serviceName == "" {
		serviceName = envOr("OTEL_SERVICE_NAME", "veil")
	}
	ratio := 1.0
	if v := os.Getenv("OTEL_TRACES_SAMPLER_ARG"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			ratio = f
		}
	}
	return Config{
		Enabled:       envBool("OTEL_ENABLED", false),
		ServiceName:   serviceName,
		OTLPEndpoint:  envOr("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317"),
		SamplerRatio:  ratio,
		MetricsListen: envOr("METRICS_LISTEN", ":9090"),
		LogFormat:     strings.ToLower(envOr("LOG_FORMAT", "json")),
		LogLevel:      strings.ToLower(envOr("LOG_LEVEL", "info")),
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}
