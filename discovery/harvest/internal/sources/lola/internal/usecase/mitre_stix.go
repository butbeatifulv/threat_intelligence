package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"
)

const defaultMitreSTIXURL = "https://raw.githubusercontent.com/mitre-attack/attack-stix-data/master/enterprise-attack/enterprise-attack.json"

type stixRel struct {
	src, tgt, rel string
}

func mitreExternalID(m map[string]any) string {
	refs, ok := m["external_references"].([]any)
	if !ok {
		return ""
	}
	for _, r := range refs {
		rm, ok := r.(map[string]any)
		if !ok {
			continue
		}
		src, _ := rm["source_name"].(string)
		ext, _ := rm["external_id"].(string)
		if strings.EqualFold(src, "mitre-attack") && ext != "" {
			return ext
		}
	}
	return ""
}

func (u *ScraperUsecase) IngestMITREEnterprise(ctx context.Context) error {
	url := strings.TrimSpace(os.Getenv("LOLA_MITRE_STIX_URL"))
	if url == "" {
		url = defaultMitreSTIXURL
	}
	maxRels := envIntLola("LOLA_MITRE_MAX_RELATIONSHIPS", 80000)
	maxTech := envIntLola("LOLA_MITRE_MAX_TECHNIQUES", 5000)
	u.logger.Info("ingesting MITRE ATT&CK STIX", slog.String("url", url))

	cacheRel := "mitre/enterprise-attack.json"
	res, err := feeds.FetchIfDue(ctx, u.feeds, u.ledger, "lola:mitre:enterprise-stix", "lola", url, ledger.PolicyStatic, cacheRel, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "veil-lola/1.0")
		return req, nil
	})
	if err != nil {
		return err
	}
	if res.Unchanged {
		u.logger.Info("MITRE STIX unchanged, skip publish")
		return nil
	}
	if res.Skipped && len(res.Body) == 0 {
		return fmt.Errorf("lola:mitre skipped by ledger without cache")
	}
	stixReader := bytes.NewReader(res.Body)

	stixToExt := map[string]string{}
	var rels []stixRel

	dec := json.NewDecoder(stixReader)
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return fmt.Errorf("mitre stix: expected object")
	}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := keyTok.(string)
		if !ok {
			return fmt.Errorf("mitre stix: bad key")
		}
		if key == "objects" {
			arr, err := dec.Token()
			if err != nil {
				return err
			}
			if d, ok := arr.(json.Delim); !ok || d != '[' {
				return fmt.Errorf("mitre stix: expected array")
			}
			techCount := 0
			for dec.More() {
				var raw json.RawMessage
				if err := dec.Decode(&raw); err != nil {
					return err
				}
				var m map[string]any
				if err := json.Unmarshal(raw, &m); err != nil {
					continue
				}
				typ, _ := m["type"].(string)
				id, _ := m["id"].(string)
				switch typ {
				case "attack-pattern":
					ext := mitreExternalID(m)
					if ext != "" && strings.HasPrefix(ext, "T") {
						stixToExt[id] = ext
						if techCount < maxTech {
							name, _ := m["name"].(string)
							desc, _ := m["description"].(string)
							md := fmt.Sprintf("# %s (`%s`)\n\n%s\n", name, ext, desc)
							if err := u.repo.UpsertAttackTechnique(ctx, ext, name, desc, md); err != nil {
								return err
							}
							techCount++
						}
					}
				case "x-mitre-tactic":
					ext := mitreExternalID(m)
					if ext != "" && strings.HasPrefix(ext, "TA") {
						stixToExt[id] = ext
						name, _ := m["name"].(string)
						desc, _ := m["description"].(string)
						md := fmt.Sprintf("# %s (`%s`)\n\n%s\n", name, ext, desc)
						if err := u.repo.UpsertAttackTactic(ctx, ext, name, desc, md); err != nil {
							return err
						}
					}
				case "relationship":
					if len(rels) >= maxRels {
						continue
					}
					rel, _ := m["relationship_type"].(string)
					src, _ := m["source_ref"].(string)
					tgt, _ := m["target_ref"].(string)
					if src == "" || tgt == "" || rel == "" {
						continue
					}
					rels = append(rels, stixRel{src: src, tgt: tgt, rel: rel})
				}
			}
			closeTok, err := dec.Token()
			if err != nil {
				return err
			}
			if d, ok := closeTok.(json.Delim); !ok || d != ']' {
				return fmt.Errorf("mitre stix: expected ] closing objects")
			}
			_, _ = dec.Token() // ]
		} else {
			var skip json.RawMessage
			if err := dec.Decode(&skip); err != nil {
				return err
			}
		}
	}

	for _, r := range rels {
		switch r.rel {
		case "subtechnique-of":
			childExt := stixToExt[r.src]
			parentExt := stixToExt[r.tgt]
			if childExt != "" && parentExt != "" {
				if err := u.repo.MergeSubtechniqueOf(ctx, parentExt, childExt); err != nil {
					return err
				}
			}
		case "supports":
			a := stixToExt[r.src]
			b := stixToExt[r.tgt]
			var tacExt, techExt string
			if strings.HasPrefix(a, "TA") && strings.HasPrefix(b, "T") {
				tacExt, techExt = a, b
			} else if strings.HasPrefix(b, "TA") && strings.HasPrefix(a, "T") {
				tacExt, techExt = b, a
			}
			if tacExt != "" && techExt != "" {
				if err := u.repo.MergeTacticIncludesTechnique(ctx, tacExt, techExt); err != nil {
					return err
				}
			}
		}
	}

	if err := u.repo.LinkArtifactsAndCommandsToTechniques(ctx); err != nil {
		return err
	}
	u.logger.Info("MITRE STIX ingest done", slog.Int("stix_ids", len(stixToExt)), slog.Int("relationships", len(rels)))
	return nil
}

func envIntLola(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return def
	}
	return n
}
