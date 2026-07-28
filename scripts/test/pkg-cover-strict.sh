#!/usr/bin/env bash
# T3 gate: 100% statement coverage on logic packages (see docs/development/pkg-test-coverage.md).
export PKG_COVER_STRICT=1
exec "$(cd "$(dirname "$0")" && pwd)/pkg-cover.sh"
