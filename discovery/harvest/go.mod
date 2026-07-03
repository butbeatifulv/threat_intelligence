module github.com/butbeautifulv/veil/discovery/harvest

go 1.25.0

require (
	github.com/butbeautifulv/veil/pkg v0.0.0
	github.com/butbeautifulv/veil/pkg/observability v0.0.0
	github.com/butbeautifulv/veil/discovery/connector v0.0.0
	github.com/butbeautifulv/veil/discovery/pkg v0.0.0
	github.com/go-sql-driver/mysql v1.9.3
	golang.org/x/net v0.49.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/nats-io/nats.go v1.48.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
)

replace github.com/butbeautifulv/veil/discovery/connector => ../connector

replace github.com/butbeautifulv/veil/pkg => ../../pkg

replace github.com/butbeautifulv/veil/pkg/observability => ../../pkg/observability

replace github.com/butbeautifulv/veil/discovery/pkg => ../pkg
