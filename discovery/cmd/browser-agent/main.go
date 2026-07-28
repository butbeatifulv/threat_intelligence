// Package main is a local healthcheck shim; production images run index.mjs (Playwright).
package main

import "os"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(0)
	}
	_, _ = os.Stderr.WriteString("discovery browser-agent: use Docker image (node index.mjs) or run: node index.mjs\n")
	os.Exit(1)
}
