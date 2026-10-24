#!/usr/bin/env bash
# Runs every verification gate (design D13). Exits non-zero on the first failure.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
export PATH="$HOME/go/bin:$PATH"

step() { printf '\n== %s\n' "$*"; }

cd "$root/golang"
step "go vet (native)"
go vet ./...
step "go vet (js/wasm)"
GOOS=js GOARCH=wasm go vet ./...
step "go test"
go test ./...
step "Go license list"
go run ./internal/licensegen -check

cd "$root"
step "fluffy generate --check"
fluffy generate --check
step "flutter analyze"
fvm flutter analyze
step "flutter test"
fvm flutter test

printf '\nAll checks passed.\n'
