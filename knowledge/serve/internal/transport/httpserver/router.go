package httpserver

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/butbeautifulv/veil/knowledge/serve/internal/domain"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/usecase"
	playbookuc "github.com/butbeautifulv/veil/knowledge/serve/internal/usecase/playbook"
	procedureuc "github.com/butbeautifulv/veil/knowledge/serve/internal/usecase/procedure"
	frameworkuc "github.com/butbeautifulv/veil/knowledge/serve/internal/usecase/framework"
	"github.com/butbeautifulv/veil/pkg/api"
)

// SetProdMode toggles generic API error messages (no internal details).
func SetProdMode(prod bool) {
	api.SetProdMode(prod)
}

func Register(mux *http.ServeMux, uc *usecase.ReadUsecase, pb *playbookuc.Service, proc *procedureuc.Service, fw *frameworkuc.Service) {
	api.RegisterHealth(mux, "veil-api", nil)
	mux.HandleFunc("GET /v1/categories", func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, map[string]any{"categories": uc.ListCategoryMeta()})
	})
	mux.HandleFunc("GET /v1/categories/{category}/kinds", func(w http.ResponseWriter, r *http.Request) {
		cat := r.PathValue("category")
		kinds, err := uc.ListKindsInCategory(r.Context(), cat)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"category": cat, "kinds": kinds})
	})
	mux.HandleFunc("GET /v1/categories/{category}/nodes", func(w http.ResponseWriter, r *http.Request) {
		cat := r.PathValue("category")
		kind := r.URL.Query().Get("kind")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		nodes, err := uc.NodesByCategory(r.Context(), cat, kind, limit, offset)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"category": cat, "kind": kind, "nodes": nodes})
	})
	mux.HandleFunc("GET /v1/categories/{category}/search", func(w http.ResponseWriter, r *http.Request) {
		cat := r.PathValue("category")
		q := r.URL.Query().Get("q")
		kind := r.URL.Query().Get("kind")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		nodes, err := uc.SearchInCategory(r.Context(), cat, q, kind, limit)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"category": cat, "query": q, "kind": kind, "nodes": nodes})
	})
	mux.HandleFunc("GET /v1/categories/engage/context", func(w http.ResponseWriter, r *http.Request) {
		host := r.URL.Query().Get("q")
		if host == "" {
			host = r.URL.Query().Get("host")
		}
		ctx, err := uc.EngageTargetContext(r.Context(), host)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		if ctx == nil {
			api.WriteJSON(w, http.StatusOK, map[string]any{"host": host, "found": false})
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"host": host, "found": ctx.Target != nil, "context": ctx})
	})
	mux.HandleFunc("GET /v1/nodes/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		n, err := uc.GetNodeForAPI(r.Context(), id)
		if err != nil {
			if errors.Is(err, domain.ErrNodeNotFound) {
				writeErr(w, http.StatusNotFound, err)
				return
			}
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"node": n})
	})
	mux.HandleFunc("GET /v1/nodes/{id}/neighbors", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		depth, _ := strconv.Atoi(r.URL.Query().Get("depth"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		g, err := uc.Neighbors(r.Context(), id, depth, limit)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"graph": g})
	})
	mux.HandleFunc("GET /v1/kinds", func(w http.ResponseWriter, r *http.Request) {
		kinds, err := uc.ListKinds(r.Context())
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"kinds": kinds})
	})
	if pb != nil {
		registerPlaybookRoutes(mux, uc, pb, proc, fw)
	}
}

func registerPlaybookRoutes(mux *http.ServeMux, uc *usecase.ReadUsecase, pb *playbookuc.Service, proc *procedureuc.Service, fw *frameworkuc.Service) {
	if fw != nil {
		mux.HandleFunc("GET /v1/playbooks/framework/mitre-layer", func(w http.ResponseWriter, r *http.Request) {
			raw, err := fw.RawMitreLayerJSON()
			if err != nil {
				writeErr(w, http.StatusInternalServerError, err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(raw)
		})
		mux.HandleFunc("GET /v1/playbooks/framework/coverage", func(w http.ResponseWriter, r *http.Request) {
			out, err := fw.MitreCoverage()
			if err != nil {
				writeErr(w, http.StatusInternalServerError, err)
				return
			}
			api.WriteJSON(w, http.StatusOK, out)
		})
		mux.HandleFunc("GET /v1/playbooks/framework/docs", func(w http.ResponseWriter, r *http.Request) {
			docs, err := fw.ListMappingDocs()
			if err != nil {
				writeErr(w, http.StatusInternalServerError, err)
				return
			}
			api.WriteJSON(w, http.StatusOK, map[string]any{"mappings_dir": "pkg/playbook/corpus/mappings", "files": docs})
		})
	}
	mux.HandleFunc("GET /v1/playbooks/subdomains", func(w http.ResponseWriter, r *http.Request) {
		meta := pb.IndexMeta()
		api.WriteJSON(w, http.StatusOK, map[string]any{
			"subdomain_counts": meta.SubdomainCounts,
			"skill_count":      meta.SkillCount,
		})
	})
	if proc != nil {
		mux.HandleFunc("GET /v1/playbooks/ontology/subdomains", func(w http.ResponseWriter, r *http.Request) {
			subs, err := proc.Subdomains()
			if err != nil {
				writeErr(w, http.StatusInternalServerError, err)
				return
			}
			api.WriteJSON(w, http.StatusOK, map[string]any{"subdomains": subs})
		})
		mux.HandleFunc("GET /v1/playbooks/ontology/technique/{technique_id}/skills", func(w http.ResponseWriter, r *http.Request) {
			tid := r.PathValue("technique_id")
			ids, err := proc.TechniqueSkillIDs(tid)
			if err != nil {
				writeErr(w, http.StatusBadRequest, err)
				return
			}
			api.WriteJSON(w, http.StatusOK, map[string]any{
				"technique_id": tid,
				"skill_ids":    ids,
				"catalog_tools": proc.CatalogToolsForTechnique(tid),
			})
		})
	}
	mux.HandleFunc("GET /v1/playbooks/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		sub := r.URL.Query().Get("subdomain")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		hits := pb.Search(q, sub, limit)
		api.WriteJSON(w, http.StatusOK, map[string]any{
			"query": q, "subdomain": sub, "skills": usecase.Summaries(hits), "count": len(hits),
		})
	})
	mux.HandleFunc("GET /v1/techniques/{technique_id}/playbooks", func(w http.ResponseWriter, r *http.Request) {
		tid := r.PathValue("technique_id")
		out, err := uc.ForTechnique(r.Context(), tid, pb.Catalog())
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, out)
	})
	if proc != nil {
		mux.HandleFunc("GET /v1/playbooks/{id}/procedure", func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("id")
			spec, err := proc.GetSpec(id)
			if err != nil {
				writeErr(w, http.StatusNotFound, err)
				return
			}
			api.WriteJSON(w, http.StatusOK, map[string]any{"procedure": spec})
		})
		mux.HandleFunc("GET /v1/playbooks/{id}/recommend-tools", func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("id")
			tools, err := proc.RecommendTools(id)
			if err != nil {
				writeErr(w, http.StatusNotFound, err)
				return
			}
			api.WriteJSON(w, http.StatusOK, map[string]any{"id": id, "catalog_tools": tools})
		})
	}
	mux.HandleFunc("GET /v1/playbooks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		detail, err := pb.Get(id)
		if err != nil {
			writeErr(w, http.StatusNotFound, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"skill": detail})
	})
}

func writeErr(w http.ResponseWriter, status int, err error) {
	api.WriteError(w, status, err)
}
