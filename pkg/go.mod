module github.com/butbeautifulv/veil/pkg

go 1.25.0

require (
	github.com/butbeautifulv/veil/pkg/engage v0.0.0
	github.com/nats-io/nats-server/v2 v2.10.29
	github.com/nats-io/nats.go v1.48.0
)

replace github.com/butbeautifulv/veil/pkg/engage => ./engage

require (
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/minio/highwayhash v1.0.3 // indirect
	github.com/nats-io/jwt/v2 v2.7.4 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/time v0.10.0 // indirect
)
