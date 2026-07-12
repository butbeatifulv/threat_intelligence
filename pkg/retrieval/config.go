package retrieval

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	EnvSearchEngine   = "VEIL_SEARCH_ENGINE"
	EnvSearchMode     = "VEIL_SEARCH_MODE"
	EnvBleveIndexPath = "VEIL_BLEVE_INDEX_PATH"
	EnvQdrantURL      = "VEIL_QDRANT_URL"
	EnvQdrantCollection = "VEIL_QDRANT_COLLECTION"
	EnvEmbedProvider  = "VEIL_EMBED_PROVIDER"
	EnvEmbedURL       = "VEIL_EMBED_URL"
	EnvEmbedModel     = "VEIL_EMBED_MODEL"
	EnvEmbedAPIKey    = "VEIL_EMBED_API_KEY"
	EnvRerankEnabled  = "VEIL_RERANK_ENABLED"
	EnvRerankURL      = "VEIL_RERANK_URL"
	EnvRepoRoot       = "VEIL_REPO_ROOT"

	DefaultBleveRel     = "docs/skills-index/playbook-search.bleve"
	DefaultQdrantURL    = "http://127.0.0.1:6333"
	DefaultCollection   = "veil_playbooks"
	DefaultSearchLimit  = 20
	DefaultFetchK       = 50
	DefaultRRFK         = 60
)

// Config holds retrieval paths and runtime mode.
type Config struct {
	RepoRoot        string
	BleveIndexPath  string
	QdrantURL       string
	QdrantCollection string
	SearchEngine    string // legacy | bleve
	SearchMode      SearchMode
	EmbedProvider   string
	EmbedURL        string
	EmbedModel      string
	EmbedAPIKey     string
	RerankEnabled   bool
	RerankURL       string
	FetchK          int
	RRFK            int
}

// ConfigFromEnv loads config from environment with repo-relative defaults.
func ConfigFromEnv(repoRoot string) Config {
	root := strings.TrimSpace(repoRoot)
	if root == "" {
		if r, err := VeilRoot(); err == nil {
			root = r
		}
	}
	bleve := os.Getenv(EnvBleveIndexPath)
	if bleve == "" && root != "" {
		bleve = filepath.Join(root, DefaultBleveRel)
	} else if bleve != "" && !filepath.IsAbs(bleve) && root != "" {
		bleve = filepath.Join(root, bleve)
	}
	engine := strings.ToLower(strings.TrimSpace(os.Getenv(EnvSearchEngine)))
	if engine == "" {
		engine = "bleve"
	}
	mode := SearchMode(strings.ToLower(strings.TrimSpace(os.Getenv(EnvSearchMode))))
	if mode == "" {
		mode = ModeHybrid
	}
	cfg := Config{
		RepoRoot:         root,
		BleveIndexPath:   bleve,
		QdrantURL:        envOr(EnvQdrantURL, DefaultQdrantURL),
		QdrantCollection: envOr(EnvQdrantCollection, DefaultCollection),
		SearchEngine:     engine,
		SearchMode:       mode,
		EmbedProvider:    envOr(EnvEmbedProvider, "ollama"),
		EmbedURL:         envOr(EnvEmbedURL, "http://127.0.0.1:11434"),
		EmbedModel:       envOr(EnvEmbedModel, "nomic-embed-text"),
		EmbedAPIKey:      os.Getenv(EnvEmbedAPIKey),
		RerankURL:        os.Getenv(EnvRerankURL),
		FetchK:           DefaultFetchK,
		RRFK:             DefaultRRFK,
	}
	cfg.RerankEnabled = strings.TrimSpace(os.Getenv(EnvRerankEnabled)) == "1"
	if v := os.Getenv("VEIL_SEARCH_FETCH_K"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.FetchK = n
		}
	}
	return cfg
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
