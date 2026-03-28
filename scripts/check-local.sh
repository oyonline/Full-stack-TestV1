#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/go-admin"
FRONTEND_DIR="$ROOT_DIR/vue-vben-admin"
WEB_DIR="$FRONTEND_DIR/apps/web-antd"

echo "==> go test ./..."
cd "$BACKEND_DIR"
go test ./...

echo "==> pnpm typecheck"
cd "$WEB_DIR"
pnpm typecheck

echo "==> pnpm build:local"
pnpm build:local

echo "==> local checks passed"
