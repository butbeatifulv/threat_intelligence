// Command indexbuild builds Bleve playbook search index from chunks JSONL.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

type chunkRow struct {
	ChunkID      string   `json:"chunk_id"`
	SkillID      string   `json:"skill_id"`
	Subdomain    string   `json:"subdomain"`
	SectionTitle string   `json:"section_title"`
	ChunkIndex   int      `json:"chunk_index"`
	AttackIDs    []string `json:"attack_ids"`
	Text         string   `json:"text"`
	Kind         string   `json:"kind"`
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: indexbuild <chunks.jsonl> <index_dir>\n")
		os.Exit(1)
	}
	rows, err := loadChunks(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "load chunks: %v\n", err)
		os.Exit(1)
	}
	indexDir := os.Args[2]
	if err := os.RemoveAll(indexDir); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "remove index: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(indexDir), 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}
	idx, err := bleve.New(indexDir, indexMapping())
	if err != nil {
		fmt.Fprintf(os.Stderr, "bleve new: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = idx.Close() }()
	batch := idx.NewBatch()
	for _, row := range rows {
		doc := map[string]any{
			"skill_id":      row.SkillID,
			"subdomain":     row.Subdomain,
			"section_title": row.SectionTitle,
			"chunk_index":   row.ChunkIndex,
			"attack_ids":    strings.Join(row.AttackIDs, " "),
			"text":          row.Text,
			"kind":          row.Kind,
		}
		if err := batch.Index(row.ChunkID, doc); err != nil {
			fmt.Fprintf(os.Stderr, "batch index: %v\n", err)
			os.Exit(1)
		}
	}
	if batch.Size() > 0 {
		if err := idx.Batch(batch); err != nil {
			fmt.Fprintf(os.Stderr, "batch: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Printf("OK: indexed %d chunks -> %s\n", len(rows), indexDir)
}

func indexMapping() mapping.IndexMapping {
	im := bleve.NewIndexMapping()
	im.DefaultAnalyzer = "en"
	doc := bleve.NewDocumentMapping()
	text := bleve.NewTextFieldMapping()
	text.Store = true
	text.IncludeInAll = true
	doc.AddFieldMappingsAt("text", text)
	kw := bleve.NewKeywordFieldMapping()
	kw.Store = true
	doc.AddFieldMappingsAt("skill_id", kw)
	doc.AddFieldMappingsAt("subdomain", kw)
	st := bleve.NewTextFieldMapping()
	st.Store = true
	doc.AddFieldMappingsAt("section_title", st)
	doc.AddFieldMappingsAt("attack_ids", st)
	doc.AddFieldMappingsAt("kind", kw)
	im.DefaultMapping = doc
	return im
}

func loadChunks(path string) ([]chunkRow, error) {
	raw, err := os.ReadFile(path) // #nosec G304 -- path is os.Args[1], an operator-supplied CLI argument to this offline indexing tool, not network input
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(path, ".jsonl") {
		var rows []chunkRow
		for _, line := range strings.Split(string(raw), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var row chunkRow
			if err := json.Unmarshal([]byte(line), &row); err != nil {
				return nil, err
			}
			rows = append(rows, row)
		}
		return rows, nil
	}
	var rows []chunkRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}
