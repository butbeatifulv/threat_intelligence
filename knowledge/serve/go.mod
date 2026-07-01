module github.com/butbeautifulv/veil/knowledge/serve

go 1.25.0

require (
	github.com/MicahParks/keyfunc/v3 v3.8.0
	github.com/butbeautifulv/veil/knowledge/connector v0.0.0
	github.com/butbeautifulv/veil/pkg/api v0.0.0
	github.com/butbeautifulv/veil/pkg/auth v0.0.0
	github.com/butbeautifulv/veil/pkg/mcp v0.0.0
	github.com/butbeautifulv/veil/pkg/observability v0.0.0
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4
)

require (
	github.com/MicahParks/jwkset v0.11.0 // indirect
	golang.org/x/time v0.9.0 // indirect
)

replace (
	github.com/butbeautifulv/veil/knowledge/connector => ../connector
	github.com/butbeautifulv/veil/pkg/api => ../../pkg/api
	github.com/butbeautifulv/veil/pkg/auth => ../../pkg/auth
	github.com/butbeautifulv/veil/pkg/mcp => ../../pkg/mcp
	github.com/butbeautifulv/veil/pkg/observability => ../../pkg/observability
)
