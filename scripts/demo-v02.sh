#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

export TOKEN="${AGENTOS_TOKEN:-dev-token}"
export BASE="${AGENTOS_BASE:-http://localhost:8080}"
export DATABASE_URL="${DATABASE_URL:-postgres://agentos:agentos@localhost:5432/agentos?sslmode=disable}"

auth() { echo "Authorization: Bearer $TOKEN"; }

curl_json() {
  local method="$1" url="$2" body="${3:-}"
  local out http_code
  if [[ -n "$body" ]]; then
    out=$(curl -sS -w "\n%{http_code}" -H "$(auth)" -H "Content-Type: application/json" \
      -X "$method" "$url" -d "$body")
  else
    out=$(curl -sS -w "\n%{http_code}" -H "$(auth)" -X "$method" "$url")
  fi
  http_code="${out##*$'\n'}"
  out="${out%$'\n'*}"
  if [[ "$http_code" -ge 300 ]]; then
    echo "HTTP $http_code: $out" >&2
    return 1
  fi
  printf '%s' "$out"
}

wait_ready() {
  local i
  for i in $(seq 1 30); do
    if curl -sf -H "$(auth)" "$BASE/readyz" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "agentosd not ready at $BASE" >&2
  return 1
}

stop_port() {
  local port="$1"
  if command -v lsof >/dev/null 2>&1; then
    local pids
    pids=$(lsof -ti ":$port" 2>/dev/null || true)
    if [[ -n "$pids" ]]; then
      echo "==> stopping process on :$port"
      # shellcheck disable=SC2086
      kill $pids 2>/dev/null || true
      sleep 1
      # shellcheck disable=SC2086
      kill -9 $pids 2>/dev/null || true
    fi
  fi
}

echo "==> build"
make build-primary

if command -v docker >/dev/null 2>&1; then
  if ! curl -sf http://localhost:8181/health >/dev/null 2>&1 || \
     ! pg_isready -h localhost -U agentos -d agentos >/dev/null 2>&1; then
    echo "==> starting docker-compose infra"
    (cd deploy/docker && docker compose up -d)
    sleep 3
  fi
  echo "==> reloading OPA policies"
  (cd deploy/docker && docker compose restart opa)
  for i in $(seq 1 20); do
    if curl -sf http://localhost:8181/health >/dev/null 2>&1; then
      break
    fi
    sleep 1
  done
fi

AGENTOSD_PID=""
cleanup() {
  if [[ -n "$AGENTOSD_PID" ]]; then
    kill "$AGENTOSD_PID" 2>/dev/null || true
    wait "$AGENTOSD_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

echo "==> starting fresh agentosd"
stop_port 8080
./bin/agentosd serve --config deploy/agentos.yaml &
AGENTOSD_PID=$!
wait_ready

echo "==> create task"
TASK_JSON=$(curl_json POST "$BASE/v1/tasks" \
  '{"agentRef":"agent:demo","input":{"goal":"demo payment analysis"},"labels":{"classification":"internal"}}')
TASK_ID=$(echo "$TASK_JSON" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
echo "task=$TASK_ID"

echo "==> tool echo"
curl_json POST "$BASE/v1/tools/tool.echo:invoke" \
  "{\"taskId\":\"$TASK_ID\",\"agentId\":\"agent:demo\",\"arguments\":{\"message\":\"hello\"}}" >/dev/null

echo "==> memory put"
curl_json POST "$BASE/v1/memory/records" "$(cat examples/memory/fact.json)" >/dev/null

echo "==> memory search"
HITS=$(curl_json POST "$BASE/v1/memory/query" '{"query":"payment"}')
HIT_COUNT=$(echo "$HITS" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d) if isinstance(d,list) else 0)")
echo "hits $HIT_COUNT"
if [[ "$HIT_COUNT" -lt 1 ]]; then
  echo "expected at least 1 memory hit for query 'payment'" >&2
  exit 1
fi

echo "==> done"
