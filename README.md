# OrderFlow

Sistema de gestão de pedidos para uma rede fictícia de restaurantes — pensado como
"playground DevOps": 3 serviços independentes + banco + cache/fila, com todos os
requisitos operacionais que uma aplicação real precisa (health checks, métricas
Prometheus, logs JSON estruturados, graceful shutdown, config 100% via env).

> **Containerização, CI/CD, Kubernetes, observabilidade etc. ficam de fora de propósito** —
> esse é o exercício de DevOps. A aplicação está pronta para ser containerizada e operada.

## Arquitetura

```
[ Frontend React ]  →  [ API (Go) ]  →  [ PostgreSQL ]
                            │
                            ├──→  [ Redis ]  (cache + fila de eventos)
                            │
                       [ Worker (Go) ]  ← consome a fila e atualiza pedidos
```

| Serviço | Stack | Porta | Descrição |
|---|---|---|---|
| `api/` | Go + Gin | 8080 | API REST de pedidos, cardápio, auth e stats |
| `worker/` | Go | 8081 (health/metrics) | Consome a fila e simula o fluxo da cozinha (`received → preparing → ready → delivered`) |
| `web/` | React + Vite | 5173 (dev) | Login, cardápio, painel de pedidos (polling 5s) e stats |
| postgres | PostgreSQL 16 | 5432 | Banco de dados |
| redis | Redis 7 | 6379 | Cache de leitura + fila de eventos (lista) |

## Pré-requisitos

- Go 1.22+
- Node.js 20+
- PostgreSQL 16 e Redis 7 rodando (localmente ou onde preferir)
- `make` (opcional, mas recomendado)
- `golangci-lint` v2 (apenas para `make lint`)

As migrations rodam via `go run` do CLI do golang-migrate — não precisa instalar nada.

## Subindo tudo localmente (sem Docker)

### 1. Banco e fila

Crie o banco e o usuário no Postgres (exemplo):

```sql
CREATE USER orderflow WITH PASSWORD 'orderflow';
CREATE DATABASE orderflow OWNER orderflow;
```

### 2. Configuração

Cada serviço lê um `.env` próprio (ou as variáveis do ambiente). Copie os exemplos:

```bash
cp api/.env.example api/.env
cp worker/.env.example worker/.env
```

A aplicação **falha na inicialização com mensagem clara** se uma variável
obrigatória (`DATABASE_URL`, `REDIS_URL`, `JWT_SECRET`) estiver ausente.

### 3. Migrations e seed

```bash
export DATABASE_URL='postgres://orderflow:orderflow@localhost:5432/orderflow?sslmode=disable'
make migrate-up
make seed     # popula o cardápio e cria o admin (admin@orderflow.local / admin123)
```

### 4. Serviços

Em terminais separados:

```bash
make run-api      # API em http://localhost:8080
make run-worker   # worker (health/metrics em http://localhost:8081)
make run-web      # frontend em http://localhost:5173 (roda npm install antes na primeira vez)
```

Sem make: `cd api && go run ./cmd/api`, `cd worker && go run ./cmd/worker`,
`cd web && npm install && npm run dev`.

Crie um pedido pelo frontend e veja o status avançar sozinho até `delivered`.

## Variáveis de ambiente

### api

| Variável | Obrigatória | Default | Descrição |
|---|---|---|---|
| `PORT` | não | `8080` | Porta HTTP |
| `DATABASE_URL` | **sim** | — | URL do Postgres |
| `REDIS_URL` | **sim** | — | URL do Redis |
| `JWT_SECRET` | **sim** | — | Segredo de assinatura dos tokens |
| `LOG_LEVEL` | não | `info` | `debug`, `info`, `warn`, `error` |
| `APP_ENV` | não | `development` | `production` ativa o release mode do Gin |

### worker

| Variável | Obrigatória | Default | Descrição |
|---|---|---|---|
| `PORT` | não | `8081` | Porta do health/metrics |
| `DATABASE_URL` | **sim** | — | URL do Postgres |
| `REDIS_URL` | **sim** | — | URL do Redis |
| `LOG_LEVEL` | não | `info` | Nível de log |
| `APP_ENV` | não | `development` | Ambiente |
| `WORKER_CONCURRENCY` | não | `4` | Tamanho do worker pool |
| `ACCEPT_TIME_SECONDS` | não | `2` | Delay `received → preparing` |
| `PREP_TIME_SECONDS` | não | `5` | Delay `preparing → ready` |
| `DELIVERY_TIME_SECONDS` | não | `3` | Delay `ready → delivered` |

### web

A URL da API é lida **em runtime** de `window.ENV.API_URL` (arquivo `/env.js`).
Em desenvolvimento, edite `web/public/env.js`. Em produção, o entrypoint do
container deve gerar o `env.js` a partir da variável `API_URL` — assim a mesma
imagem serve em qualquer ambiente.

## Endpoints principais (api)

| Método | Rota | Auth | Descrição |
|---|---|---|---|
| POST | `/api/v1/auth/register` | — | Cria usuário |
| POST | `/api/v1/auth/login` | — | Login, retorna JWT (24h) |
| GET | `/api/v1/menu` | — | Cardápio |
| POST | `/api/v1/orders` | Bearer | Cria pedido e publica evento na fila |
| GET | `/api/v1/orders?page=&limit=&status=` | — | Lista paginada (cache 10s na 1ª página) |
| GET | `/api/v1/orders/:id` | — | Detalhe com histórico de status |
| GET | `/api/v1/stats` | — | Totais do dia |
| GET | `/api/v1/version` | — | Versão e commit (via `-ldflags`) |
| GET | `/healthz` | — | Liveness |
| GET | `/readyz` | — | Readiness (checa Postgres e Redis, 503 se fora) |
| GET | `/metrics` | — | Métricas Prometheus |

O worker expõe `/healthz`, `/readyz` e `/metrics` na porta 8081.

Os arquivos em `api/hacks/*.http` têm requests prontos para testar os endpoints
(extensão REST Client do VS Code).

## Build com versão

```bash
cd api && go build -ldflags "-X main.version=1.0.0 -X main.commit=$(git rev-parse --short HEAD)" ./cmd/api
cd worker && go build -ldflags "-X main.version=1.0.0 -X main.commit=$(git rev-parse --short HEAD)" ./cmd/worker
```

## Testes e lint

```bash
make test    # go test ./... nos dois serviços Go
make lint    # golangci-lint nos serviços Go + eslint no front
```

## Alvos do Makefile

| Alvo | Descrição |
|---|---|
| `run-api` / `run-worker` / `run-web` | Sobe cada serviço localmente |
| `test` | Testes dos serviços Go |
| `lint` | golangci-lint + eslint |
| `migrate-up` / `migrate-down` | Migrations (usa `DATABASE_URL` do ambiente) |
| `seed` | Popula cardápio + usuário admin |

## Notas de operação

- **Logs**: JSON estruturado no stdout via `slog`, com `component`, `request_id`,
  método, rota, status e latência por request.
- **Graceful shutdown**: SIGTERM/SIGINT param de aceitar conexões e esperam os
  requests/jobs em andamento (timeout 15s). O worker termina o job atual antes de sair.
- **Idempotência do worker**: cada transição só é aplicada se o pedido ainda estiver
  no status esperado — reprocessar o mesmo evento não duplica transições. Na
  inicialização, pedidos não finalizados são recolocados na fila.
- **Fila**: lista do Redis (`orderflow:orders:events`), `LPUSH` na API e `BRPOP` no worker.
