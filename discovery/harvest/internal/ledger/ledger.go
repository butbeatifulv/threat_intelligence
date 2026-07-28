// Package ledger stores crawl metadata in MySQL/Vitess (not graph content).
package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// FetchPolicy controls refetch cadence.
type FetchPolicy string

const (
	PolicyStatic   FetchPolicy = "static"
	PolicyPeriodic FetchPolicy = "periodic"
	PolicyDaily    FetchPolicy = "daily"
)

// Store records URL fetch metadata.
type Store struct {
	db *sql.DB
}

func OpenFromEnv() (*Store, error) {
	dsn := strings.TrimSpace(os.Getenv("VITESS_DSN"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("MYSQL_DSN"))
	}
	if dsn == "" {
		return nil, fmt.Errorf("ledger: VITESS_DSN or MYSQL_DSN required")
	}
	dsn = ensureParseTimeDSN(dsn)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) EnsureSchema(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS crawl_resource (
  resource_key VARCHAR(512) PRIMARY KEY,
  source VARCHAR(64) NOT NULL,
  url TEXT NOT NULL,
  etag VARCHAR(255) NULL,
  content_sha256 CHAR(64) NULL,
  last_fetched_at TIMESTAMP NOT NULL,
  last_changed_at TIMESTAMP NULL,
  fetch_policy VARCHAR(16) NOT NULL
)`)
	return err
}

// GetContentSHA returns the last recorded content hash for a resource, or empty if unknown.
func (s *Store) GetContentSHA(ctx context.Context, key string) (string, error) {
	var sha sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT content_sha256 FROM crawl_resource WHERE resource_key = ?`, key).Scan(&sha)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if sha.Valid {
		return sha.String, nil
	}
	return "", nil
}

// ShouldFetch returns true when resource is due for refetch.
func (s *Store) ShouldFetch(ctx context.Context, key string, policy FetchPolicy, minRefetch time.Duration, force bool) (bool, error) {
	if force {
		return true, nil
	}
	var lastFetched time.Time
	err := s.db.QueryRowContext(ctx, `SELECT last_fetched_at FROM crawl_resource WHERE resource_key = ?`, key).Scan(&lastFetched)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	switch policy {
	case PolicyStatic:
		return false, nil
	case PolicyDaily:
		return time.Since(lastFetched) >= 24*time.Hour, nil
	default:
		if minRefetch <= 0 {
			return true, nil
		}
		return time.Since(lastFetched) >= minRefetch, nil
	}
}

// RecordFetch upserts crawl metadata after a successful fetch.
func (s *Store) RecordFetch(ctx context.Context, key, source, url string, policy FetchPolicy, contentSHA256 string) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, `
INSERT INTO crawl_resource (resource_key, source, url, content_sha256, last_fetched_at, last_changed_at, fetch_policy)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  source = VALUES(source),
  url = VALUES(url),
  content_sha256 = VALUES(content_sha256),
  last_fetched_at = VALUES(last_fetched_at),
  last_changed_at = IF(content_sha256 <> VALUES(content_sha256) OR content_sha256 IS NULL, VALUES(last_fetched_at), last_changed_at),
  fetch_policy = VALUES(fetch_policy)
`, key, source, url, nullIfEmpty(contentSHA256), now, now, string(policy))
	return err
}

func ensureParseTimeDSN(dsn string) string {
	if strings.Contains(dsn, "parseTime=") {
		return dsn
	}
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return dsn + sep + "parseTime=true"
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
