#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

export DATABASE_URL="${DATABASE_URL:-postgres://agentos:agentos@localhost:5432/agentos?sslmode=disable}"

echo "==> build"
make build-primary

echo "==> unit tests"
go test ./...

echo "==> policy tests"
make policy

if command -v docker >/dev/null 2>&1; then
  echo "==> starting docker-compose infra"
  (cd deploy/docker && docker compose up -d)
  sleep 3
fi

echo "==> agentosd smoke (requires Postgres + OPA)"
if curl -sf http://localhost:8181/health >/dev/null 2>&1 && \
   pg_isready -h localhost -U agentos -d agentos >/dev/null 2>&1; then
  ./bin/agentosd serve --config deploy/agentos.yaml &
  PID=$!
  trap 'kill $PID 2>/dev/null || true' EXIT
  sleep 2
  curl -sf http://localhost:8080/healthz
  curl -sf http://localhost:8080/readyz
  ./bin/agentctl status
  echo "demo: v0.2 operable core endpoints reachable"
else
  echo "skip e2e: Postgres or OPA not available on localhost"
fi

echo "==> done"
