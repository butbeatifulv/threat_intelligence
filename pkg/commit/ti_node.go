package commit

import (
	"fmt"
	"strings"
)

// TI upsert payloads are normalized by NED before publish. Graph ingest uses
// IdempotencyKey as the Neo4j node id (see IOCNodeID, ActorNodeID, etc.).

const (
	tiIOCKeyPrefix    = "ti:ioc:"
	tiActorKeyPrefix  = "ti:actor:"
	tiReportKeyPrefix = "ti:report:"
)

// IOCNodeID returns the Neo4j IOC node id from a ti:ioc: idempotency key.
func IOCNodeID(idempotencyKey string) (string, error) {
	return tiKeySuffix(idempotencyKey, tiIOCKeyPrefix)
}

// IOCLinkNodeID returns the IOC node id suffix from ti:lc: or ti:lrmi: link keys.
func IOCLinkNodeID(idempotencyKey string) (string, error) {
	parts := strings.Split(idempotencyKey, ":")
	if len(parts) < 2 {
		return "", fmt.Errorf("commit: invalid TI link idempotency key %q", idempotencyKey)
	}
	id := parts[len(parts)-1]
	if id == "" {
		return "", fmt.Errorf("commit: empty IOC id in link key %q", idempotencyKey)
	}
	return id, nil
}

// ActorNodeID returns the Neo4j Actor node id from a ti:actor: idempotency key.
func ActorNodeID(idempotencyKey string) (string, error) {
	return tiKeySuffix(idempotencyKey, tiActorKeyPrefix)
}

// ReportNodeID returns the Neo4j Report node id from a ti:report: idempotency key.
func ReportNodeID(idempotencyKey string) (string, error) {
	return tiKeySuffix(idempotencyKey, tiReportKeyPrefix)
}

// TILinkSuffix returns the last colon-separated segment of a TI link idempotency key.
func TILinkSuffix(idempotencyKey string) string {
	if i := strings.LastIndex(idempotencyKey, ":"); i >= 0 && i+1 < len(idempotencyKey) {
		return idempotencyKey[i+1:]
	}
	return ""
}

func tiKeySuffix(idempotencyKey, prefix string) (string, error) {
	if !strings.HasPrefix(idempotencyKey, prefix) {
		return "", fmt.Errorf("commit: idempotency key %q does not have prefix %q", idempotencyKey, prefix)
	}
	id := strings.TrimPrefix(idempotencyKey, prefix)
	if id == "" {
		return "", fmt.Errorf("commit: empty id in key %q", idempotencyKey)
	}
	return id, nil
}
