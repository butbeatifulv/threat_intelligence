// Package ids provides stable TI entity identifiers for pipeline dedup and graph ingest.
package ids

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
)

// CanonicalIOCID is the Neo4j IOC node id for a normalized IOC (type:value hash).
func CanonicalIOCID(i domain.IOC) string {
	return i.NodeID()
}

// ActorStableID returns a deterministic actor id from display name.
func ActorStableID(name string) string {
	n := strings.TrimSpace(strings.ToLower(name))
	h := sha256.Sum256([]byte("actor:" + n))
	return hex.EncodeToString(h[:])
}

// ReportStableID returns a deterministic report id from canonical link.
func ReportStableID(link string) string {
	h := sha256.Sum256([]byte("report:" + strings.TrimSpace(link)))
	return hex.EncodeToString(h[:])
}
