module github.com/butbeautifulv/veil/platform/gateway

go 1.25.0

toolchain go1.25.12

require github.com/butbeautifulv/veil/pkg/api v0.0.0

require github.com/butbeautifulv/veil/pkg/auth v0.0.0 // indirect

replace (
	github.com/butbeautifulv/veil/pkg/api => ../../pkg/api
	github.com/butbeautifulv/veil/pkg/auth => ../../pkg/auth
)
