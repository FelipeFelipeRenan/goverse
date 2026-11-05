#!/usr/bin/env bash
# Teste de fumaça do monolito: cadastro -> login -> /auth/me -> CSRF -> editar perfil.
#   ./deploy/smoke.sh        roda o fluxo completo
#   ./deploy/smoke.sh wait   só espera a API ficar pronta
set -euo pipefail

BASE="${BASE:-http://localhost:8080}"
JAR="$(mktemp)"; trap 'rm -f "$JAR"' EXIT
ok()   { echo "  ✔ $*"; }
fail() { echo "  ✘ $*" >&2; exit 1; }

echo "Aguardando $BASE/readyz ..."
for i in $(seq 1 60); do
  if curl -fsS -o /dev/null "$BASE/readyz" 2>/dev/null; then ok "API pronta"; break; fi
  [ "$i" = 60 ] && fail "API não ficou pronta em 60s (veja: make logs)"
  sleep 1
done
[ "${1:-}" = "wait" ] && exit 0

EMAIL="smoke-$(date +%s)-$RANDOM@example.com"; PASS="senha-de-teste-123"

code=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE/user" -H 'Content-Type: application/json' \
  -d "{\"username\":\"Smoke\",\"email\":\"$EMAIL\",\"password\":\"$PASS\"}")
[ "$code" = 201 ] && ok "POST /user -> 201" || fail "POST /user -> $code (esperado 201)"

code=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE/user" -H 'Content-Type: application/json' \
  -d "{\"username\":\"Smoke\",\"email\":\"$EMAIL\",\"password\":\"$PASS\"}")
[ "$code" = 409 ] && ok "e-mail duplicado -> 409" || fail "duplicado -> $code (esperado 409)"

code=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE/auth/login" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"errada-123\"}")
[ "$code" = 401 ] && ok "senha errada -> 401" || fail "senha errada -> $code (esperado 401)"

CSRF=$(curl -fsS -c "$JAR" -X POST "$BASE/auth/login" -H 'Content-Type: application/json' \
  -d "{\"type\":\"password\",\"email\":\"$EMAIL\",\"password\":\"$PASS\"}" | jq -r .csrf_token)
[ -n "$CSRF" ] && [ "$CSRF" != null ] && ok "login -> cookies + csrf_token" || fail "login sem csrf_token"

code=$(curl -s -o /dev/null -w '%{http_code}' "$BASE/auth/me"); [ "$code" = 401 ] \
  && ok "/auth/me sem sessão -> 401" || fail "/auth/me sem sessão -> $code"

curl -fsS -b "$JAR" "$BASE/auth/me" | jq -e --arg e "$EMAIL" '.email == $e' >/dev/null \
  && ok "/auth/me com sessão devolve o usuário" || fail "/auth/me com sessão"

code=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR" -X PUT "$BASE/user/me" \
  -H 'Content-Type: application/json' -d '{"username":"Novo Nome"}')
[ "$code" = 403 ] && ok "PUT /user/me sem header CSRF -> 403" || fail "PUT sem CSRF -> $code (esperado 403)"

curl -fsS -b "$JAR" -X PUT "$BASE/user/me" -H 'Content-Type: application/json' -H "X-CSRF-TOKEN: $CSRF" \
  -d '{"username":"Novo Nome"}' | jq -e '.username == "Novo Nome"' >/dev/null \
  && ok "PUT /user/me com CSRF atualiza o perfil" || fail "PUT /user/me com CSRF"

body=$(curl -fsS -b "$JAR" "$BASE/users")
echo "$body" | grep -qi 'password\|\$2[aby]\$' && fail "/users vaza senha/hash" || ok "/users não vaza senha"

code=$(curl -s -o /dev/null -w '%{http_code}' "$BASE/users"); [ "$code" = 401 ] \
  && ok "/users sem sessão -> 401" || fail "/users sem sessão -> $code"

echo "Tudo certo ✅"
