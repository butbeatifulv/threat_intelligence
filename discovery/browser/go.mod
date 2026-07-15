module github.com/butbeautifulv/veil/discovery/browser

go 1.25.0

toolchain go1.25.12

require (
	github.com/butbeautifulv/veil/discovery/connector v0.0.0
	github.com/butbeautifulv/veil/discovery/pkg v0.0.0
	github.com/butbeautifulv/veil/pkg v0.0.0
)

require (
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/nats-io/nats.go v1.48.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
)

replace github.com/butbeautifulv/veil/discovery/connector => ../connector

replace github.com/butbeautifulv/veil/discovery/pkg => ../pkg

replace github.com/butbeautifulv/veil/pkg => ../../pkg
