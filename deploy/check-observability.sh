#!/usr/bin/env bash
# Confere que a telemetria chega ao Grafana (LGTM) e que o dashboard "Goverse" foi provisionado
# com métricas que existem de verdade. Pressupõe a stack no ar e algum tráfego (make smoke).
set -euo pipefail
GRAFANA="${GRAFANA:-http://localhost:3000}"
AUTH="admin:admin"
fail() { echo "✘ $*" >&2; exit 1; }

echo "== aguardando Grafana"
for _ in $(seq 1 60); do
  curl -fsS -u "$AUTH" "$GRAFANA/api/health" >/dev/null 2>&1 && break; sleep 2
done
curl -fsS -u "$AUTH" "$GRAFANA/api/health" >/dev/null || fail "Grafana não respondeu"

UID_PROM=$(curl -fsS -u "$AUTH" "$GRAFANA/api/datasources" | jq -r '[.[]|select(.type=="prometheus")][0].uid')
[ -n "$UID_PROM" ] && [ "$UID_PROM" != "null" ] || fail "datasource Prometheus não encontrado"
names() { curl -fsS -u "$AUTH" "$GRAFANA/api/datasources/proxy/uid/$UID_PROM/api/v1/label/__name__/values" | jq -r '.data[]'; }

REQUIRED=(http_server_request_duration_seconds_count http_server_request_duration_seconds_bucket chat_spool_pending)
echo "== aguardando métricas (export a cada 10s)"
for _ in $(seq 1 30); do
  all=$(names 2>/dev/null || true); missing=()
  for m in "${REQUIRED[@]}"; do grep -qx "$m" <<<"$all" || missing+=("$m"); done
  [ ${#missing[@]} -eq 0 ] && break; sleep 5
done
if [ ${#missing[@]} -ne 0 ]; then
  echo "métricas ausentes: ${missing[*]}" >&2
  echo "--- métricas http_/chat_/go_ que EXISTEM:" >&2
  grep -E '^(http_|chat_|go_|db_|pgx)' <<<"$all" >&2 || true
  exit 1
fi
echo "✔ métricas presentes"

echo "== dashboard"
curl -fsS -u "$AUTH" "$GRAFANA/api/search?query=Goverse" | jq -e 'map(select(.uid=="goverse-overview")) | length == 1' >/dev/null \
  || fail "dashboard goverse-overview não foi provisionado"
echo "✔ dashboard provisionado"

echo "== toda métrica usada nas queries do dashboard existe"
used=$(jq -r '.. | .expr? // empty' deploy/grafana/goverse.json | grep -oE '(http_server_[a-z_]+|chat_[a-z_]+)' | sort -u)
bad=0
for m in $used; do grep -qx "$m" <<<"$all" || { echo "  ✘ $m (usada no dashboard, mas não existe)"; bad=1; }; done
[ $bad -eq 0 ] || { echo "$all" | grep -E '^(http_|chat_)' >&2; exit 1; }
echo "✔ queries do dashboard referenciam métricas reais"
