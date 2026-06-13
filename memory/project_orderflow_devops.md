---
name: OrderFlow DevOps Learning Project
description: Contexto do projeto orderflow — app de pedidos sendo usado para aprender DevOps do zero ao Kubernetes
type: project
---

App de pedidos (restaurante) usado como base de aprendizado DevOps seguindo guia.md.

**Arquitetura:**
- `api/` — Go + Gin, Postgres, Redis, JWT auth, Prometheus metrics em `/metrics`, health em `/healthz` e `/readyz`
- `worker/` — Go, consome fila Redis, processa pedidos (status: pending→accepted→preparing→delivered), também expõe `/healthz` `/readyz` `/metrics` na porta 8081
- `web/` — React/Vite, usa `window.ENV.API_URL` carregado de `public/env.js` (runtime config pattern para Docker)
- Postgres com 5 migrations em `api/db/migrations/`

**Estado atual (2026-06-12):** Fase 1 do guia.
- API Dockerfile existe (multi-stage, alpine, layer cache correto), mas **falta non-root user**
- Worker Dockerfile: não existe ainda
- Web Dockerfile: não existe ainda
- docker-compose.yml: não existe ainda

**Why:** Aprendizado socrático — guia não entrega código pronto, ensina o caminho.
**How to apply:** Não entregar implementações prontas. Fazer perguntas, dar dicas, explicar conceitos. Só escrever código se o usuário pedir explicitamente.
