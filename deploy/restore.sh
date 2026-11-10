#!/usr/bin/env bash
# Restaura um backup no Postgres da stack de desenvolvimento. APAGA os dados atuais.
# Uso: deploy/restore.sh backups/goverse-AAAAMMDD-HHMMSS.dump
set -euo pipefail
cd "$(dirname "$0")/.."
FILE="${1:?uso: deploy/restore.sh <arquivo.dump>}"
[ -f "$FILE" ] || { echo "arquivo não existe: $FILE" >&2; exit 1; }
COMPOSE=(docker compose $([ -f .env ] && echo --env-file .env) -f deploy/compose.yml)

if [ "${GOVERSE_ASSUME_YES:-}" != "1" ]; then
  read -r -p "Isto APAGA os dados atuais e restaura $FILE. Digite 'sim' para continuar: " ans
  [ "$ans" = "sim" ] || { echo "cancelado"; exit 1; }
fi
"${COMPOSE[@]}" stop goverse
"${COMPOSE[@]}" exec -T postgres pg_restore -U goverse -d goverse --clean --if-exists --no-owner < "$FILE"
"${COMPOSE[@]}" start goverse
echo "restaurado: $FILE"
