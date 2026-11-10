#!/usr/bin/env bash
# Backup lógico do Postgres da stack de desenvolvimento (pg_dump formato custom).
# Uso: deploy/backup.sh [diretório]   (padrão: ./backups; mantém os 14 mais recentes)
set -euo pipefail
cd "$(dirname "$0")/.."
DIR="${1:-backups}"
KEEP="${GOVERSE_BACKUP_KEEP:-14}"
COMPOSE=(docker compose $([ -f .env ] && echo --env-file .env) -f deploy/compose.yml)

mkdir -p "$DIR"
OUT="$DIR/goverse-$(date +%Y%m%d-%H%M%S).dump"
"${COMPOSE[@]}" exec -T postgres pg_dump -U goverse -d goverse -Fc > "$OUT.tmp"
# um dump vazio ou truncado não pode virar "backup": valida que o pg_restore consegue listá-lo
if ! "${COMPOSE[@]}" exec -T postgres pg_restore --list < "$OUT.tmp" >/dev/null; then
  rm -f "$OUT.tmp"; echo "backup inválido, descartado" >&2; exit 1
fi
mv "$OUT.tmp" "$OUT"
ls -1t "$DIR"/goverse-*.dump | tail -n +$((KEEP + 1)) | xargs -r rm -f
echo "backup: $OUT ($(wc -c < "$OUT") bytes)"
