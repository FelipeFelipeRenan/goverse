#!/usr/bin/env bash
# Teste de fumaça do monolito (identity + rooms) contra a API em execução.
#   ./deploy/smoke.sh        roda o fluxo completo
#   ./deploy/smoke.sh wait   só espera a API ficar pronta
# Requer: curl e jq. Compatível com bash 3.2 (macOS).
set -euo pipefail

BASE="${BASE:-http://localhost:${GOVERSE_PUBLIC_PORT:-80}}"
DIR="$(mktemp -d)"; trap 'rm -rf "$DIR"' EXIT
ok()   { echo "  ✔ $*"; }
fail() { echo "  ✘ $*" >&2; exit 1; }
for bin in curl jq; do command -v "$bin" >/dev/null || fail "preciso do '$bin' instalado"; done

echo "Aguardando $BASE/readyz ..."
for i in $(seq 1 60); do
  if curl -fsS -o /dev/null "$BASE/readyz" 2>/dev/null; then ok "API pronta"; break; fi
  [ "$i" = 60 ] && fail "API não ficou pronta em 60s (veja: make logs)"
  sleep 1
done
[ "${1:-}" = "wait" ] && exit 0

PASS="senha-de-teste-123"
R=""; S=""

# do_call USUARIO METODO CAMINHO [CORPO]  -> preenche $S (status) e $R (corpo). USUARIO vazio = anônimo.
do_call() {
  local u=$1 m=$2 p=$3 b=${4:-} out
  local args=(-s -w '\n%{http_code}' -X "$m" "$BASE$p" -H 'Content-Type: application/json')
  if [ -n "$u" ]; then
    args+=(-b "$DIR/$u.jar" -H "X-CSRF-TOKEN: $(awk '$6=="csrf_token"{print $7}' "$DIR/$u.jar")")
  fi
  [ -n "$b" ] && args+=(-d "$b")
  out=$(curl "${args[@]}")
  S=${out##*$'\n'}; R=${out%$'\n'*}
}
expect() { [ "$S" = "$1" ] && ok "$2 -> $1" || fail "$2 -> $S (esperado $1): $R"; }

# signup NOME -> cadastra, loga (cookies em $DIR/NOME.jar) e imprime o id do usuário
signup() {
  local n=$1 email="smoke-$1-$(date +%s)-$RANDOM@example.com"
  curl -fsS -o /dev/null -X POST "$BASE/user" -H 'Content-Type: application/json' \
    -d "{\"username\":\"$n\",\"email\":\"$email\",\"password\":\"$PASS\"}"
  curl -fsS -c "$DIR/$n.jar" -X POST "$BASE/auth/login" -H 'Content-Type: application/json' \
    -d "{\"email\":\"$email\",\"password\":\"$PASS\"}" | jq -r .user.id
}

echo; echo "== identity =="
EMAIL="smoke-main-$(date +%s)-$RANDOM@example.com"
do_call "" POST /user "{\"username\":\"Smoke\",\"email\":\"$EMAIL\",\"password\":\"$PASS\"}"; expect 201 "POST /user"
do_call "" POST /user "{\"username\":\"Smoke\",\"email\":\"$EMAIL\",\"password\":\"$PASS\"}"; expect 409 "e-mail duplicado"
do_call "" POST /auth/login "{\"email\":\"$EMAIL\",\"password\":\"errada-123\"}";               expect 401 "senha errada"
curl -fsS -c "$DIR/main.jar" -X POST "$BASE/auth/login" -H 'Content-Type: application/json' \
  -d "{\"type\":\"password\",\"email\":\"$EMAIL\",\"password\":\"$PASS\"}" | jq -e '.csrf_token | length > 0' >/dev/null \
  && ok "login devolve csrf_token e cookies" || fail "login sem csrf_token"
do_call ""     GET /auth/me;                                          expect 401 "/auth/me sem sessão"
do_call main   GET /auth/me;                                          expect 200 "/auth/me com sessão"
echo "$R" | jq -e --arg e "$EMAIL" '.email == $e' >/dev/null && ok "/auth/me devolve o usuário" || fail "/auth/me: $R"
# PUT sem o header CSRF (usa curl direto para não enviar o header)
code=$(curl -s -o /dev/null -w '%{http_code}' -b "$DIR/main.jar" -X PUT "$BASE/user/me" -H 'Content-Type: application/json' -d '{"username":"X"}')
[ "$code" = 403 ] && ok "PUT /user/me sem CSRF -> 403" || fail "PUT sem CSRF -> $code (esperado 403)"
do_call main PUT /user/me '{"username":"Novo Nome"}';                 expect 200 "PUT /user/me com CSRF"
echo "$R" | jq -e '.username == "Novo Nome"' >/dev/null && ok "perfil atualizado" || fail "perfil: $R"
do_call main GET /users;                                              expect 200 "GET /users com sessão"
echo "$R" | grep -qi 'password\|\$2[aby]\$' && fail "/users vaza senha/hash" || ok "/users não vaza senha"
do_call ""   GET /users;                                              expect 401 "GET /users sem sessão"

echo; echo "== rooms =="
A=$(signup a); B=$(signup b); C=$(signup c)
[ -n "$A" ] && [ -n "$B" ] && [ -n "$C" ] && ok "3 usuários criados" || fail "signup falhou"

do_call "" POST /rooms '{"name":"x"}';                                expect 401 "criar sala sem sessão"
do_call a POST /rooms '{"name":"Sala Smoke","description":"d","is_public":true,"max_members":2}'; expect 201 "A cria sala pública (max 2)"
ROOM=$(echo "$R" | jq -r .id)
echo "$R" | jq -e '.name == "Sala Smoke" and .member_count == 1 and .max_members == 2' >/dev/null && ok "formato id/name e contagem" || fail "sala: $R"
do_call b POST /rooms '{"name":"Sala Secreta","is_public":false}';   expect 201 "B cria sala privada"
SECRET=$(echo "$R" | jq -r .id)

do_call "" GET "/rooms?limit=100";                                    expect 200 "listar salas sem login"
echo "$R" | jq -e --arg id "$ROOM" 'map(.room_id) | index($id) != null' >/dev/null && ok "sala pública aparece (formato room_id)" || fail "lista: $R"
echo "$R" | grep -q "$SECRET" && fail "sala privada vazou na listagem" || ok "sala privada NÃO aparece"
do_call "" GET "/rooms?public_only=false&limit=100";                  expect 200 "public_only=false"
echo "$R" | grep -q "$SECRET" && fail "public_only=false vaza sala privada" || ok "public_only=false não vaza"
do_call a GET "/rooms/$SECRET";                                       expect 404 "A não enxerga a sala privada de B"
do_call a POST "/rooms/$SECRET/join";                                 expect 403 "A não entra em sala privada"

do_call a GET "/rooms/$ROOM/members";                                 expect 200 "membros da sala"
echo "$R" | jq -e 'length == 1 and .[0].role == "owner" and (.[0].user.email | length > 0)' >/dev/null \
  && ok "dono listado com dados do usuário (1 consulta, sem N+1)" || fail "membros: $R"
do_call b POST "/rooms/$ROOM/join";                                   expect 200 "B entra"
do_call b POST "/rooms/$ROOM/join";                                   expect 200 "B entra de novo (idempotente)"
do_call c POST "/rooms/$ROOM/join";                                   expect 409 "C não entra: sala lotada"
do_call a GET "/rooms/$ROOM/members"
echo "$R" | jq -e 'length == 2 and all(.[]; .user.email | length > 0)' >/dev/null && ok "2 membros com usuário" || fail "membros: $R"

do_call b PATCH "/rooms/$ROOM" '{"name":"hack"}';                     expect 403 "membro comum não edita"
do_call b DELETE "/rooms/$ROOM";                                      expect 403 "membro comum não apaga"
do_call a POST "/rooms/$ROOM/members" "{\"user_id\":\"$C\",\"role\":\"owner\"}"; expect 400 "não dá para criar um 2º dono"
do_call a PUT "/rooms/$ROOM/members/$B/role" '{"new_role":"admin"}';  expect 200 "A promove B a admin"
do_call b PATCH "/rooms/$ROOM" '{"name":"Renomeada","max_members":5}'; expect 200 "admin edita a sala"
echo "$R" | jq -e '.room_name == "Renomeada" and .max_members == 5' >/dev/null && ok "PATCH aplica name e max_members" || fail "patch: $R"
do_call c POST "/rooms/$ROOM/join";                                   expect 200 "C entra após aumentar a sala"
do_call a DELETE "/rooms/$ROOM/members/$A";                           expect 409 "dono não pode ser removido"
do_call b DELETE "/rooms/$ROOM/members/$C";                           expect 403 "admin não expulsa outro membro (só o dono)"
do_call a DELETE "/rooms/$ROOM/members/$C";                           expect 200 "dono expulsa C"
do_call b GET /user/rooms;                                            expect 200 "salas de B"
echo "$R" | jq -e --arg id "$ROOM" 'map(.room_id) | index($id) != null' >/dev/null && ok "B vê a sala em /user/rooms" || fail "user/rooms: $R"
do_call a GET /rooms/mine;                                            expect 200 "salas do dono"
echo "$R" | jq -e --arg id "$ROOM" 'map(.room_id) | index($id) != null' >/dev/null && ok "A vê a sala em /rooms/mine" || fail "rooms/mine: $R"
do_call a DELETE "/rooms/$ROOM";                                      expect 200 "dono apaga a sala"
do_call a GET "/rooms/$ROOM";                                         expect 404 "sala apagada some"

echo; echo "Tudo certo ✅"
