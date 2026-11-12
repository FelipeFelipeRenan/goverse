# Operação

## Observabilidade

`make up` sobe o Grafana LGTM em <http://localhost:3000> (admin/admin). O dashboard **Goverse** já vem provisionado
(`deploy/grafana/goverse.json`): taxa de requisições, latência p50/p95/p99, erro 5xx, conexões WebSocket, mensagens por
resultado, barramento entre réplicas e o **spool**. O CI (`deploy/check-observability.sh`) confere que toda métrica
usada nas queries existe de verdade no Prometheus.

O que olhar quando algo está errado:

| Sintoma | Painel / métrica | Leitura |
|---|---|---|
| Mensagens "somem" ao recarregar | `chat_spool_pending` > 0, `chat_messages_total{result="spooled"}` | banco fora; as mensagens estão no spool |
| Remetentes recebem erro | `result="failed"` ou `"rejected"` | sem spool configurado, ou sala apagada |
| Usuários em réplicas diferentes não se veem | `chat_bus_events_total{outcome="error"}` | barramento caiu; ver logs "barramento caiu" |
| `chat.dead_letters` com linhas | `result="dead_letter"` | investigue `SELECT * FROM chat.dead_letters` |

## Backup e restore

Desenvolvimento: `make backup` (grava em `./backups`, valida com `pg_restore --list`, guarda os 14 mais recentes) e
`make restore FILE=backups/goverse-AAAAMMDD-HHMMSS.dump` (**apaga** os dados atuais; pede confirmação).
O CI faz a ida e volta (backup, restore, mesma contagem de usuários).

Kubernetes: `CronJob goverse-backup` (03:17, `pg_dump -Fc` para o PVC `goverse-backups`, 7 dias). Isso **não** protege contra
perder o cluster: copie os dumps para fora (bucket) e **teste um restore de verdade** antes de confiar neles:

```bash
kubectl -n goverse exec -it <pod-com-o-pvc> -- ls /backups
pg_restore -d "<URL do banco novo>" --clean --if-exists --no-owner goverse-....dump
```

Postgres gerenciado costuma oferecer PITR; prefira-o em produção e trate o dump como segunda camada.
