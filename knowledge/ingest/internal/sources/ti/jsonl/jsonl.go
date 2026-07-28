package jsonl

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
)

// Envelope is a single JSONL record.
// Input format: JSON lines. Each line can be:
// - {"ioc": {...}}
// - {"campaign": {...}}
// - {"cluster": {...}}
type Envelope struct {
	IOC      *domain.IOC      `json:"ioc,omitempty"`
	Campaign *domain.Campaign `json:"campaign,omitempty"`
	Cluster  *domain.Cluster  `json:"cluster,omitempty"`
	Actor    *domain.Actor    `json:"actor,omitempty"`
	Report   *domain.Report   `json:"report,omitempty"`
}

type Stream struct {
	r io.Reader
}

func NewStreamFromFile(path string) (*Stream, func() error, error) {
	f, err := os.Open(path) // #nosec G304 -- path is a caller-supplied local file argument (sole caller passes a deploy-time env var), not network/request input
	if err != nil {
		return nil, nil, err
	}
	return &Stream{r: f}, f.Close, nil
}

func (s *Stream) Walk(ctx context.Context, fn func(env Envelope) error) (int, error) {
	sc := bufio.NewScanner(s.r)
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	count := 0
	for sc.Scan() {
		select {
		case <-ctx.Done():
			return count, ctx.Err()
		default:
		}

		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var env Envelope
		if err := json.Unmarshal(line, &env); err != nil {
			return count, fmt.Errorf("invalid jsonl at line %d: %w", count+1, err)
		}
		if env.IOC == nil && env.Campaign == nil && env.Cluster == nil && env.Actor == nil && env.Report == nil {
			continue
		}
		if err := fn(env); err != nil {
			return count, err
		}
		count++
	}
	if err := sc.Err(); err != nil {
		return count, err
	}
	return count, nil
}
