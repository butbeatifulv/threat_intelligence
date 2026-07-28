package domain

import (
	"crypto/sha256"
	"encoding/hex"
)

// NodeID returns the canonical Neo4j IOC node id for a normalized IOC (type:value hash).
// Used by offline JSONL ingest; NATS path should prefer commit.IOCNodeID(idempotency_key).
func (i IOC) NodeID() string {
	h := sha256.Sum256([]byte(string(i.Type) + ":" + i.Value))
	return hex.EncodeToString(h[:])
}
