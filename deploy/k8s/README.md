# Deploy em Kubernetes

Um Deployment (1 réplica, ver aviso do chat), Service, Ingress (nginx) e ConfigMap. Kustomize puro, sem Helm.

## Primeira vez

1. Edite `base/configmap.yaml` (origens do frontend, redirect do Google, endpoint OTLP) e `base/ingress.yaml` (host).
2. Crie a chave JWT e os segredos:

```bash
openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
kubectl create namespace goverse
kubectl -n goverse create secret generic goverse-secrets \
  --from-file=jwt-private-key=private.pem \
  --from-literal=database-url='postgres://<usuario>:<senha>@<host>:5432/goverse?sslmode=require' \
  --from-literal=google-client-id='...' --from-literal=google-client-secret='...'
shred -u private.pem   # a chave só deve existir no cluster
```

3. Aplique:

```bash
kubectl apply -k deploy/k8s/base                         # Postgres externo/gerenciado
# ou, para estudo/staging, com Postgres no cluster:
kubectl -n goverse create secret generic goverse-postgres --from-literal=password='...'
kubectl apply -k deploy/k8s/overlays/in-cluster-postgres  # database-url: postgres://goverse:<senha>@postgres:5432/goverse?sslmode=disable
```

As migrations rodam no boot (lock consultivo no Postgres), então não há Job de migração.

## Imagens e releases

O workflow `release.yml` publica `ghcr.io/felipefeliperenan/goverse` a cada push na `main` (tags `sha-<commit>` e `latest`). Para fixar uma versão:

```bash
cd deploy/k8s/base && kustomize edit set image ghcr.io/felipefeliperenan/goverse=ghcr.io/felipefeliperenan/goverse:sha-<commit>
```

Se o pacote no GHCR for privado, crie um `imagePullSecret`.

## Rotacionar a chave JWT

Trocar a chave invalida todas as sessões (o servidor só aceita a chave atual). Atualize o Secret e faça `kubectl -n goverse rollout restart deploy/goverse`.
