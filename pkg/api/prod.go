package api

import "sync/atomic"

var prodMode atomic.Bool

// SetProdMode toggles generic API error messages (no internal details).
func SetProdMode(prod bool) {
	prodMode.Store(prod)
}

// ProdMode reports whether production-safe error text is enabled.
func ProdMode() bool {
	return prodMode.Load()
}
