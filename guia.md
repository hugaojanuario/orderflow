# Guia DevOps — Do código ao Kubernetes (versão para primeira jornada)

> **Como este guia funciona:** ele não te dá código pronto — te dá o caminho.
> Cada fase é dividida em passos pequenos. Em cada passo você encontra:
> - **O que fazer** (a tarefa)
> - 🔍 **Pesquise:** os termos exatos para jogar no Google/YouTube/doc oficial
> - ✔️ **Confira:** como saber que aquele passo funcionou antes de seguir
>
> **Regra de ouro:** um passo de cada vez. Não leia a fase inteira tentando entender tudo —
> faça o passo 1, confira, vá para o passo 2. Entendimento vem fazendo.

## Antes de começar: como pesquisar e como travar do jeito certo

1. **Prefira a documentação oficial** (docs.docker.com, kubernetes.io, etc.) e vídeos curtos para o primeiro contato com um conceito. Blog post aleatório só depois.
2. **Quando travar:** releia a mensagem de erro com calma (ela quase sempre diz o problema), pesquise a mensagem entre aspas, tente por até ~1 hora. Travou de verdade? Aí sim pergunte (para mim, fórum, etc.) — mas levando o que você já tentou. Isso é treino de DevOps também: 70% do trabalho é diagnosticar.
3. **Caderno de bordo:** crie um arquivo `DIARIO.md` no repo de infra. Toda vez que resolver um problema, anote em 3 linhas: o que quebrou, qual era a causa, como resolveu. Em 3 meses isso vira ouro para entrevistas.
4. **Não existe "fase quase pronta".** Só avance quando o checklist final da fase estiver 100%.

---

# FASE 0 — Preparar o terreno (repositório + servidor)

**Em palavras simples:** antes de qualquer Docker ou pipeline, você precisa de um repositório organizado e de um servidor seguro. Servidor exposto na internet sem proteção é invadido em questão de horas (sério — robôs escaneiam a internet inteira o tempo todo procurando porta aberta).

### Parte A — O repositório

**0.1 — Crie o repositório no GitHub** com o código do OrderFlow (que o Claude Code gerou), público (é portfólio!). Estrutura: `api/`, `worker/`, `web/` na raiz.
✔️ Confira: `git clone` numa pasta nova baixa tudo e o README explica como rodar.

**0.2 — Garanta que nenhum segredo está no git.** Abra o `.gitignore` e confirme que `.env` está lá. Procure no histórico se alguma senha já vazou.
🔍 Pesquise: `gitignore golang node`, `como verificar se commitei senha no git`
✔️ Confira: `git log --all -p | grep -i password` não encontra nada real.

**0.3 — Proteja a branch main.** Configure para que ninguém (nem você) consiga dar push direto na `main` — tudo entra por Pull Request. Parece burocracia trabalhando sozinho, mas é assim que você treina o fluxo profissional, e vai ser necessário para o CI da Fase 2.
🔍 Pesquise: `github branch protection rules require pull request`
✔️ Confira: `git push origin main` direto é rejeitado.

**0.4 — Adote um padrão de commit.** Use Conventional Commits: `feat: ...`, `fix: ...`, `chore: ...`, `docs: ...`. Só isso por enquanto.
🔍 Pesquise: `conventional commits cheat sheet`

### Parte B — O servidor

**0.5 — Crie um usuário para você trabalhar** (ex: `deploy`), em vez de usar root para tudo. Dê a ele sudo e, mais tarde, permissão de docker.
🔍 Pesquise: `ubuntu adduser sudo group`
✔️ Confira: você loga como `deploy` e consegue rodar `sudo apt update`.

**0.6 — Troque senha por chave SSH.** Gere um par de chaves na SUA máquina, coloque a chave pública no servidor, e teste o login sem senha. Depois desligue o login por senha e o login direto de root no arquivo de configuração do SSH.
🔍 Pesquise: `ssh-keygen ed25519`, `ssh-copy-id`, `sshd_config PasswordAuthentication no PermitRootLogin no`
⚠️ **Importante:** só desative a senha DEPOIS de confirmar que o login por chave funciona, senão você se tranca para fora.
✔️ Confira: `ssh deploy@servidor` entra sem pedir senha; tentar logar com senha falha.

**0.7 — Ligue o firewall.** Política: bloquear tudo que entra, liberar só o necessário — portas 22 (SSH), 80 (HTTP) e 443 (HTTPS).
🔍 Pesquise: `ufw allow ssh http https enable`
⚠️ Libere a 22 ANTES de ativar o firewall, pelo mesmo motivo do passo anterior.
✔️ Confira: de outra rede (4G do celular), um scan mostra só 22/80/443 abertas. 🔍 Pesquise: `nmap scan portas abertas ip`

**0.8 — Instale o Docker pela documentação oficial** (o do repositório padrão do Ubuntu costuma ser velho). Adicione o usuário `deploy` ao grupo docker.
🔍 Pesquise: `install docker engine ubuntu official`, `docker without sudo`
✔️ Confira: `docker run hello-world` funciona logado como `deploy`, sem sudo.

**0.9 — Proteções automáticas:** instale `fail2ban` (bane IPs que ficam errando senha no SSH) e ative `unattended-upgrades` (atualizações de segurança automáticas).
🔍 Pesquise: `fail2ban ubuntu ssh`, `unattended-upgrades ubuntu`

**0.10 — Aponte um domínio para o servidor.** Pode ser um domínio barato (.com.br ~R$40/ano) ou um subdomínio gratuito (DuckDNS). Crie um registro DNS tipo A apontando para o IP do servidor. Você vai precisar disso para ter HTTPS na Fase 3.
🔍 Pesquise: `duckdns tutorial` ou `registro A dns apontar para ip`
✔️ Confira: `ping seudominio.com` responde com o IP do servidor.

### ✅ Fase 0 pronta quando
- [ ] Login só por chave SSH, root e senha desativados
- [ ] Firewall ativo com só 22/80/443
- [ ] `docker run hello-world` funciona sem sudo
- [ ] Branch main protegida, `.env` fora do git
- [ ] Domínio apontando para o servidor

---

# FASE 1 — Colocar tudo em containers

**Em palavras simples:** hoje o OrderFlow roda com `go run` e `npm run dev`. Nesta fase cada serviço vira uma **imagem Docker** (um pacote com tudo que o serviço precisa para rodar em qualquer lugar) e o ambiente completo — api, worker, front, banco, redis — sobe com um único comando.

**1.1 — Rode tudo na mão primeiro.** Antes de containerizar, suba a aplicação localmente seguindo o README (Postgres e Redis podem ser containers avulsos: `docker run postgres`, `docker run redis`). Crie um pedido e veja o status avançar.
✔️ Confira: você entende o que cada serviço precisa para subir (variáveis, portas, ordem). Anote — isso é literalmente o conteúdo do Dockerfile e do compose.

**1.2 — Dockerfile da API, versão simples.** Escreva o primeiro Dockerfile do jeito ingênuo: imagem do Go, copia o código, compila, roda. Funcionar primeiro, otimizar depois.
🔍 Pesquise: `dockerfile golang example`
✔️ Confira: `docker build` gera a imagem e `docker run` (com as envs via `-e` e `--network` apontando para o postgres) sobe a API. Anote o tamanho da imagem (`docker images`) — vai ser engraçado depois.

**1.3 — Agora otimize com multi-stage build.** A imagem do passo anterior deve ter ~1GB, porque carrega o compilador Go inteiro. Multi-stage = um estágio compila, outro estágio (uma imagem mínima) só recebe o binário pronto. Meta: **menos de 30MB**.
🔍 Pesquise: `dockerfile multi-stage golang distroless`
✔️ Confira: `docker images` mostra a nova imagem com ~5% do tamanho da antiga. Sinta o orgulho.

**1.4 — Boas práticas no Dockerfile da API:** rodar como usuário não-root (containers rodando como root são risco de segurança), criar um `.dockerignore` (irmão do .gitignore — evita copiar lixo para a imagem), e copiar `go.mod`/`go.sum` + baixar dependências ANTES de copiar o código (assim o Docker reaproveita cache e o build fica rápido quando só o código muda).
🔍 Pesquise: `dockerfile non root user`, `dockerignore`, `docker layer caching go mod download`
✔️ Confira: mudar uma linha de código e rebuildar leva segundos (não re-baixa dependências); `docker run ... whoami` não retorna root.

**1.5 — Dockerfile do worker.** Mesma receita da API. Vai ser rápido — é o mesmo padrão.

**1.6 — Dockerfile do front.** Dois estágios: Node compila o React (gera arquivos estáticos), Nginx serve esses arquivos. Detalhe importante da spec: o entrypoint gera um `env.js` com a URL da API lida de variável de ambiente — assim a MESMA imagem serve em qualquer ambiente (dev, prod), que é uma regra de ouro de DevOps: **build uma vez, configura em runtime**.
🔍 Pesquise: `dockerfile react vite nginx multi-stage`, `react runtime environment variables nginx entrypoint`
✔️ Confira: rodar a imagem com `-e API_URL=http://teste1` e depois `-e API_URL=http://teste2` muda a URL no `env.js` servido, sem rebuildar.

**1.7 — docker-compose.yml: o ambiente inteiro num comando.** O compose descreve os 5 serviços (api, worker, web, postgres, redis) num arquivo YAML. Pontos que você precisa implementar:
- **Rede interna**: os serviços se enxergam pelo nome (a API acessa o banco como `postgres:5432`, não localhost). Só web e api expõem porta para fora.
- **Volume nomeado** no postgres, para os dados sobreviverem ao `docker compose down`.
- **Ordem de subida com saúde**: a API não pode subir antes do banco estar PRONTO (não só "ligado"). Use healthcheck no postgres/redis + `depends_on` com `condition: service_healthy`.
- **Migrations**: um serviço one-shot que roda as migrations e termina, antes da API subir.

🔍 Pesquise: `docker compose depends_on condition service_healthy postgres`, `docker compose named volumes`, `docker compose run migrations golang-migrate`
✔️ Confira: `docker compose up` numa máquina limpa sobe tudo funcionando; `docker compose down` e `up` de novo NÃO perde os pedidos criados.

**1.8 — Quebre de propósito (exercício):** derrube só o redis (`docker compose stop redis`) e veja o `/readyz` da API virar 503. Suba de volta e veja voltar a 200. Você acabou de ver na prática para que serve o readiness — e vai reusar esse conceito direto no Kubernetes.

### ✅ Fase 1 pronta quando
- [ ] 3 Dockerfiles multi-stage, imagens Go < 30MB, todos não-root
- [ ] `docker compose up` sobe o sistema completo numa máquina limpa
- [ ] Dados do banco sobrevivem a down/up
- [ ] Mesma imagem do front funciona com APIs diferentes via env

---

# FASE 2 — CI: o robô que revisa seu código (GitHub Actions)

**Em palavras simples:** CI (Integração Contínua) é um robô que, a cada Pull Request, roda automaticamente lint, testes e build. Se algo falha, o PR fica bloqueado. Você nunca mais depende de "lembrar de rodar os testes".

**2.1 — Primeiro workflow: o hello world.** Crie um workflow que roda em todo PR e só imprime uma mensagem. Objetivo: entender a anatomia do arquivo (gatilho → job → steps) e ver a bolinha verde no PR.
🔍 Pesquise: `github actions workflow basics on pull_request`
✔️ Confira: abrir um PR dispara o workflow e ele aparece verde na aba Actions.

**2.2 — Job de lint.** Faça o workflow rodar o golangci-lint na api e no worker, e o ESLint no front.
🔍 Pesquise: `golangci-lint github action`, `eslint github actions`
✔️ Confira: um PR com código feio (ex: variável não usada) fica com ❌.

**2.3 — Job de testes.** Rode `go test ./...` nos serviços Go. Os testes de integração precisam de Postgres e Redis — o Actions resolve isso com **service containers** (bancos descartáveis que sobem só durante o job).
🔍 Pesquise: `github actions service containers postgres redis go test`
✔️ Confira: quebrar um teste de propósito bloqueia o PR; consertar libera.

**2.4 — Job de build das imagens.** Em PR, só construir as 3 imagens Docker (validar que buildam — sem publicar).
🔍 Pesquise: `docker build-push-action push false`

**2.5 — Scan de segurança com Trivy.** Trivy examina sua imagem procurando vulnerabilidades conhecidas nas dependências e no sistema base. Configure para falhar o job se achar vulnerabilidade HIGH ou CRITICAL.
🔍 Pesquise: `trivy github action scan docker image severity`
✔️ Confira: o job roda e passa (se reprovar, geralmente trocar a versão da imagem base resolve — investigue!).

**2.6 — Não rode o que não mudou.** Hoje, mudar um CSS roda o pipeline do Go inteiro. Configure **path filters**: workflow do backend dispara quando `api/**` ou `worker/**` mudam; do front, quando `web/**` muda.
🔍 Pesquise: `github actions on push paths filter monorepo`
✔️ Confira: um PR só de front não roda jobs de Go.

**2.7 — Publicar imagens no merge.** Quando o PR é mergeado na main, o pipeline deve construir e **publicar** as imagens no GHCR (o registry gratuito do GitHub), tagueadas com o SHA do commit (ex: `ghcr.io/voce/orderflow-api:a1b2c3d`). A tag pelo SHA é o que vai te permitir saber exatamente qual código está em produção — e voltar atrás.
🔍 Pesquise: `github actions push image ghcr GITHUB_TOKEN packages write`, `docker metadata-action tags sha`
✔️ Confira: após o merge, as imagens aparecem na aba Packages do repo, e `docker pull` delas funciona no servidor.

**2.8 — Trave a porta.** Volte na branch protection e marque os jobs de lint/teste/build como **obrigatórios** para merge.
✔️ Confira: é impossível mergear um PR com check vermelho.

**2.9 — Acelere com cache (opcional, mas ensina muito).** Configure cache para módulos Go e camadas Docker, e compare o tempo do pipeline antes/depois.
🔍 Pesquise: `github actions cache go modules`, `buildx cache gha`

### ✅ Fase 2 pronta quando
- [ ] PR roda lint + testes + build + Trivy automaticamente
- [ ] Check vermelho = merge bloqueado
- [ ] Merge na main publica imagens no GHCR com tag do commit
- [ ] Pipeline do front não roda para mudança de backend (e vice-versa)

---

# FASE 3 — Primeiro deploy de verdade (produção com Compose + HTTPS)

**Em palavras simples:** aqui o OrderFlow vai para o ar, no seu servidor, com endereço próprio e cadeado verde no navegador — e atualizado automaticamente a cada merge. Kubernetes fica para a Fase 5; primeiro você aprende a operar produção do jeito simples.

**3.1 — Estrutura de produção no servidor.** Crie um diretório (ex: `/opt/orderflow`) com: um `docker-compose.prod.yml` (diferença para o de dev: usa as imagens prontas do GHCR com `image:`, em vez de `build:`) e o `.env` de produção, com senhas fortes e permissão restrita (`chmod 600` — só o dono lê).
✔️ Confira: `docker compose -f docker-compose.prod.yml up -d` sobe a aplicação puxando as imagens do GHCR. Por enquanto acesse pelo IP:porta, sem HTTPS — um passo de cada vez.

**3.2 — Entenda o problema do HTTPS antes de resolver.** Você tem 2 serviços web (front e api) e só uma porta 443. Quem resolve é um **proxy reverso**: um serviço na frente de todos, que recebe TODA requisição e roteia pelo domínio (`app.seudominio.com` → front, `api.seudominio.com` → api), além de cuidar do certificado TLS. Use o **Caddy** — ele emite e renova o certificado Let's Encrypt sozinho, com um arquivo de configuração de ~10 linhas (Nginx + certbot é o caminho clássico, mas tem mais peças; Caddy te dá o conceito com menos atrito).
🔍 Pesquise: `caddy docker compose reverse proxy automatic https`, `o que é proxy reverso`
✔️ Confira: `https://app.seudominio.com` abre o front com cadeado válido; o site redireciona HTTP→HTTPS; as portas dos serviços internos NÃO estão mais expostas (só o Caddy expõe 80/443).

**3.3 — Deploy automático (CD).** Crie o workflow de CD: depois que a imagem é publicada (fase 2.7), ele conecta no servidor via SSH e roda os comandos de atualização (pull das imagens novas + recriar containers). Para isso: gere um par de chaves SSH NOVO só para o deploy, e guarde a chave privada nos **GitHub Secrets** (cofre de segredos do Actions).
🔍 Pesquise: `github actions deploy ssh docker compose pull up`, `appleboy ssh-action secrets`
✔️ Confira: você muda um texto no front, abre PR, mergeia... e ~5 min depois a mudança está no ar sem você tocar no servidor. Esse momento é mágico — comemore.

**3.4 — Migrations no deploy.** Garanta que as migrations rodam como etapa do deploy, ANTES de trocar o container da API.
✔️ Confira: um PR que adiciona uma coluna nova chega em produção com a coluna criada.

**3.5 — Rollback: o plano para quando der ruim.** Documente (e TESTE) o procedimento de voltar para a versão anterior: como descobrir a tag anterior, e como subir ela de novo. Meta: menos de 2 minutos.
Exercício: faça um deploy propositalmente quebrado (ex: API que não sobe) e execute seu rollback cronometrado.
🔍 Pesquise: `docker compose rollback image tag strategy`
✔️ Confira: existe um `ROLLBACK.md` no repo e você já executou ele uma vez de verdade.

**3.6 — Higiene de produção (o que separa amador de profissional):**
- `restart: unless-stopped` em todos os serviços (sobrevivem a reboot do servidor)
- Limites de memória nos serviços do compose (um vazamento de memória não pode derrubar o servidor inteiro)
- **Rotação de logs do Docker** — sem isso os logs crescem até encher o disco, o incidente nº 1 de quem começa
  🔍 Pesquise: `docker compose restart policy`, `docker compose mem_limit`, `docker daemon json log max-size max-file`
  ✔️ Confira: `sudo reboot` no servidor e a aplicação volta sozinha.

**3.7 — Backup do banco (inegociável).** Um script com `pg_dump` rodando todo dia via cron, guardando os últimos 7 dias, e enviando uma cópia para FORA do servidor (se o disco morrer, o backup não pode morrer junto — pode ser um bucket S3-compatível gratuito, como Cloudflare R2).
E a parte que 90% pula: **teste o restore**. Restaure o dump num banco vazio e confira os dados. Backup nunca testado é só um arquivo que te dá falsa confiança.
🔍 Pesquise: `pg_dump docker cron backup script`, `restore pg_dump`, `rclone cloudflare r2`
✔️ Confira: existe um backup de ontem fora do servidor, e você já restaurou um com sucesso.

### ✅ Fase 3 pronta quando
- [ ] App no ar com HTTPS válido e redirect
- [ ] Merge na main → produção atualizada sem tocar no servidor
- [ ] Rollback documentado e testado em < 2 min
- [ ] Servidor reinicia e tudo volta sozinho
- [ ] Backup diário fora do servidor, restore testado

---

# FASE 4 — Observabilidade: enxergar o sistema

**Em palavras simples:** hoje, se a API cair às 3h da manhã, você só descobre quando alguém reclamar. Observabilidade é montar os "instrumentos do painel do carro": **métricas** (números ao longo do tempo: latência, requisições/s), **logs** (o diário de bordo de cada serviço) e **alertas** (o sistema te chama quando algo foge do normal). A aplicação já expõe tudo isso (o `/metrics` e os logs JSON da spec) — agora você monta quem consome.

**As peças (conheça antes de montar):** **Prometheus** coleta e armazena métricas, visitando o `/metrics` de cada serviço a cada X segundos. **Grafana** transforma essas métricas em dashboards visuais. **Loki** armazena logs (e o **Promtail/Alloy** é quem coleta dos containers e envia pra ele). **node_exporter** expõe métricas da máquina (CPU, disco, RAM) e **cAdvisor** as dos containers.
🔍 Pesquise: `prometheus grafana loki explicado` (vale um vídeo de 15 min antes de começar)

**4.1 — Suba Prometheus + Grafana** num compose separado (ex: `/opt/monitoring`). Configure o Prometheus para coletar (`scrape`) a API e o worker do OrderFlow, pela rede interna do Docker. Grafana acessível só via Caddy (ex: `grafana.seudominio.com`) com senha forte.
🔍 Pesquise: `prometheus grafana docker compose`, `prometheus scrape_configs`
✔️ Confira: em Status→Targets no Prometheus, api e worker aparecem como UP; no Grafana, você consegue plotar uma métrica da API.

**4.2 — Métricas da máquina e dos containers:** adicione node_exporter e cAdvisor ao compose e ao scrape do Prometheus. Importe dashboards prontos da comunidade no Grafana para eles (importar pronto aqui é ok — o dashboard da SUA aplicação você vai fazer na mão).
🔍 Pesquise: `node exporter docker compose`, `cadvisor prometheus`, `grafana import dashboard node exporter 1860`
✔️ Confira: você vê CPU/RAM/disco do servidor e o consumo de cada container.

**4.3 — Logs centralizados com Loki.** Suba Loki + Promtail (ou Alloy) coletando os logs de todos os containers. Como a aplicação loga em JSON, configure o parse dos campos (level, component, request_id).
🔍 Pesquise: `loki promtail docker compose docker logs`, `promtail pipeline json`
✔️ Confira: no Grafana (Explore→Loki) você filtra logs por serviço e por level=error; pega um `request_id` de um log da API e encontra a requisição inteira.

**4.4 — Dashboard do OrderFlow (feito por você).** Monte um dashboard com os **4 sinais de ouro** da API: latência (p95 — o tempo que 95% das requisições levam), tráfego (req/s), erros (% de respostas 5xx) e saturação (CPU/RAM dos containers). Adicione um segundo bloco de negócio: pedidos por status, tamanho da fila, tempo de processamento do worker. Você vai precisar aprender o básico de **PromQL** (a linguagem de consulta do Prometheus) — é a parte mais difícil da fase, reserve um tempo.
🔍 Pesquise: `golden signals sre`, `promql rate histogram_quantile tutorial`
✔️ Confira: você gera carga criando vários pedidos e VÊ os gráficos reagirem em tempo real.

**4.5 — Alertas no seu celular.** Configure alertas (pelo alerting do próprio Grafana, que é mais simples para começar) enviando para Telegram ou Discord, para: API fora do ar, erro 5xx acima de um limite, disco > 80%, fila do worker só crescendo.
🔍 Pesquise: `grafana alerting telegram contact point`
✔️ Confira: derrube a API de propósito → alerta chega no celular em < 2 min → suba → chega o "resolved".

**4.6 — Vigia externo:** se o servidor INTEIRO cair, seu Grafana cai junto e ninguém te avisa. Crie um monitor gratuito de fora (UptimeRobot) apontando para a URL pública.
✔️ Confira: pare o servidor por 5 min e o UptimeRobot te avisa.

**4.7 — Exercício de bombeiro:** peça para alguém (ou simule) quebrar algo aleatório: parar o redis, encher o disco, derrubar o worker. Sua missão: descobrir O QUE quebrou usando SÓ Grafana/Loki, sem dar SSH. Depois escreva seu primeiro **runbook**: um doc "sintoma → onde olhar → como resolver".
✔️ Confira: existe um `runbooks/` no repo com pelo menos 2 cenários escritos por você.

### ✅ Fase 4 pronta quando
- [ ] Dashboard com golden signals + métricas de negócio, feito por você
- [ ] Logs de todos os serviços pesquisáveis no Grafana
- [ ] Alerta no celular em < 2 min quando a API cai (testado!)
- [ ] Monitor externo ativo
- [ ] 2+ runbooks escritos

---

# FASE 5 — Kubernetes

**Em palavras simples:** o Compose executa o que você manda, uma vez. O Kubernetes trabalha diferente: você **declara o estado desejado** ("quero 2 réplicas da API, com esses recursos, saudáveis") e ele trabalha 24/7 para manter isso verdade — container morreu, ele recria; nó reiniciou, ele realoca. É o padrão da indústria para orquestrar containers, e é a skill mais pedida em vagas DevOps. Você vai usar o **k3s**: um Kubernetes completo e leve, perfeito para um servidor só.

**5.1 — Instale o k3s no servidor** e configure o `kubectl` (a ferramenta de linha de comando) na SUA máquina, falando com o cluster remotamente.
🔍 Pesquise: `k3s install`, `k3s kubeconfig remote access`
⚠️ k3s vai querer as portas 80/443 para o ingress dele. Conflito com o Caddy da Fase 3 — decida: ou migre tudo de uma vez, ou suba o k3s em portas alternativas durante a transição.
✔️ Confira: `kubectl get nodes` na sua máquina mostra o servidor como Ready.

**5.2 — Brinque antes de migrar (importante!).** Esqueça o OrderFlow por um dia. Suba um nginx de teste e pratique o vocabulário: **Pod** (a menor unidade — um container rodando), **Deployment** (gerencia N réplicas de um pod), **Service** (endereço fixo interno na frente dos pods). Delete o pod e veja o Deployment recriar sozinho. Escale para 3 réplicas com um comando.
🔍 Pesquise: `kubernetes deployment service tutorial kubectl`, `kubectl scale delete pod`
✔️ Confira: você consegue explicar com suas palavras a diferença entre Pod, Deployment e Service. (Sério — explique em voz alta.)

**5.3 — Migre a API (o primeiro serviço).** Escreva os YAMLs: um namespace `orderflow`, um Deployment da API e um Service. No Deployment, capriche em 4 coisas que você plantou lá na spec do projeto:
- **liveness probe** → `/healthz` (K8s reinicia o pod se falhar)
- **readiness probe** → `/readyz` (K8s só manda tráfego se estiver pronto)
- **resources** (requests/limits de CPU e memória)
- rodar como não-root (securityContext)
  🔍 Pesquise: `kubernetes deployment yaml liveness readiness probe`, `kubernetes resources requests limits explained`
  ✔️ Confira: `kubectl get pods -n orderflow` mostra a API Running e READY 1/1.

**5.4 — Config e segredos.** As envs não-sensíveis vão num **ConfigMap**; senhas e JWT_SECRET vão em **Secrets**. O Deployment referencia os dois.
🔍 Pesquise: `kubernetes configmap secret envfrom deployment`
✔️ Confira: nenhum valor sensível escrito direto no YAML do Deployment.

**5.5 — Banco e Redis no cluster.** Aqui tem decisão de arquitetura: bancos têm ESTADO, e estado no K8s exige **StatefulSet + PVC** (PersistentVolumeClaim — o pedido de disco persistente). Recomendo rodar dentro do cluster justamente para aprender storage, sabendo que em produção real o comum é banco gerenciado (RDS etc.) — saiba explicar esse trade-off.
🔍 Pesquise: `kubernetes statefulset postgres pvc tutorial`, `k3s local-path storage`
✔️ Confira: deletar o pod do postgres NÃO perde os dados (o PVC segura).

**5.6 — Worker, front e migrations.** Worker: Deployment sem Service de negócio (ninguém chama ele — ele consome a fila). Front: Deployment + Service. Migrations: um **Job** do K8s (roda até completar e para) executado a cada release.
🔍 Pesquise: `kubernetes job migrations`
✔️ Confira: o sistema completo funciona dentro do cluster (teste com `kubectl port-forward` no front).

**5.7 — Exponha para o mundo: Ingress + TLS.** O **Ingress** é o "Caddy do Kubernetes": roteia domínios para Services (o k3s já traz o Traefik embutido). O certificado quem emite/renova é o **cert-manager** com Let's Encrypt.
🔍 Pesquise: `k3s traefik ingress example`, `cert-manager letsencrypt http01 tutorial`
✔️ Confira: `https://app.seudominio.com` agora é servido pelo cluster, cadeado válido.

**5.8 — Aponte o CD para o cluster.** Atualize o workflow de deploy: em vez de SSH+compose, ele atualiza a tag da imagem nos manifests e aplica (`kubectl apply` ou kustomize), esperando o `kubectl rollout status` confirmar. Pesquise como dar ao Actions um acesso ao cluster com permissão mínima.
🔍 Pesquise: `github actions deploy kubernetes kubectl set image rollout status`, `kustomize edit set image`
✔️ Confira: merge na main → pods novos sobem, velhos morrem, site nunca sai do ar (o rolling update em ação).

**5.9 — Treino de operação (até virar reflexo):** `kubectl get/describe/logs/exec`, `kubectl rollout undo` (o rollback de 1 comando!), escalar a API para 2 réplicas e ver o Service balancear, matar um pod no meio de um pedido e ver o sistema se recuperar.
✔️ Confira: você faz deploy quebrado de propósito e o readiness probe SEGURA o tráfego nos pods velhos (usuário nem percebe) — entenda por que isso aconteceu.

**5.10 — Migre o monitoramento para o cluster** com o kube-prometheus-stack (chart que instala Prometheus+Grafana+exporters integrados ao K8s) e configure a coleta da api/worker via ServiceMonitor. (Esse passo usa Helm — se preferir, faça depois da Fase 6.)
🔍 Pesquise: `kube-prometheus-stack install`, `servicemonitor example`

### ✅ Fase 5 pronta quando
- [ ] OrderFlow completo no k3s com HTTPS
- [ ] Probes, resources e secrets configurados em tudo
- [ ] Deploy via pipeline com rolling update sem downtime
- [ ] `kubectl rollout undo` testado
- [ ] Você explica o caminho de um request: internet → ingress → service → pod

---

# FASE 6 — Helm: empacotar a aplicação

**Em palavras simples:** você tem uma pilha de YAMLs com valores repetidos. Se quiser um ambiente de staging, teria que duplicar tudo. O **Helm** transforma seus YAMLs em **templates** com variáveis: um "pacote" (chart) + um arquivo de valores por ambiente.

**6.1 — Estude um chart pronto** antes de criar o seu: rode `helm create teste` e explore o esqueleto gerado (templates/, values.yaml). Entenda como `{{ .Values.image.tag }}` funciona.
🔍 Pesquise: `helm chart tutorial values templates`

**6.2 — Converta o OrderFlow num chart.** Transforme seus manifests em templates, extraindo para o `values.yaml` tudo que varia: tags de imagem, número de réplicas, recursos, hosts do ingress.
✔️ Confira: `helm template .` gera YAML idêntico (em essência) ao que você tinha; `helm install` sobe a aplicação igualzinha.

**6.3 — O superpoder: dois ambientes, um chart.** Crie `values-prod.yaml` e `values-staging.yaml` (staging: 1 réplica, menos recursos, outro subdomínio) e suba os dois em namespaces separados no mesmo cluster.
✔️ Confira: subir um staging completo do zero é UM comando. Sinta o poder.

**6.4 — Migration como hook.** Configure o Job de migrations como hook `pre-upgrade` do Helm (roda automaticamente antes de cada release).
🔍 Pesquise: `helm hooks pre-upgrade job`

**6.5 — Atualize o CD para `helm upgrade --install` com a flag `--atomic`** — se o release falhar, o Helm desfaz sozinho. Pratique também `helm rollback` e `helm history`.
✔️ Confira: um deploy com imagem inexistente (de propósito) volta sozinho para a versão anterior.

### ✅ Fase 6 pronta quando
- [ ] Chart próprio do OrderFlow versionado no repo
- [ ] Staging e prod do mesmo chart, valores diferentes
- [ ] Deploy do pipeline via helm com --atomic

---

# FASE 7 — Terraform: infraestrutura escrita em código

**Em palavras simples:** até aqui, tudo que existe "fora" da aplicação (servidor, firewall, configurações do GitHub) foi criado na mão, clicando. **IaC** (Infrastructure as Code) descreve essa infraestrutura em arquivos de código: versionável, revisável por PR, e recriável do zero com um comando. Terraform é a ferramenta padrão. O ciclo é: `terraform plan` (mostra O QUE vai mudar) → você revisa → `terraform apply` (executa).

**7.1 — Primeiro contato com algo que você JÁ tem: o GitHub.** Em vez de começar criando servidores, use o **provider do GitHub** para gerenciar via Terraform o que você fez na Fase 0: o repositório, a branch protection, os secrets. Use `terraform import` para trazer o que já existe para o controle do Terraform.
🔍 Pesquise: `terraform github provider repository branch protection`, `terraform import`
✔️ Confira: mudar uma regra de branch protection editando `.tf` + `apply` funciona; `terraform plan` sem mudanças mostra "no changes" (= a realidade bate com o código).

**7.2 — Entenda o STATE (o conceito mais importante do Terraform).** O Terraform guarda num arquivo de estado o mapa "o que eu criei e onde". Perder ou corromper esse arquivo é o pior acidente possível. Por isso ele nunca fica só na sua máquina: configure um **remote state** num bucket S3-compatível (Cloudflare R2 tem nível gratuito).
🔍 Pesquise: `terraform state explained`, `terraform backend s3 cloudflare r2`
✔️ Confira: o state está no bucket; rodar terraform de outra máquina enxerga a mesma infra.

**7.3 — Agora sim, cloud de verdade.** Crie uma conta num free tier (Oracle Cloud Always Free é generoso — VMs ARM gratuitas para sempre; AWS free tier também serve) e provisione via Terraform: a rede (VCN/VPC), as regras de firewall (security list), e uma VM com IP público. Vá recurso por recurso, rodando `plan` antes de cada `apply` e LENDO o plano.
🔍 Pesquise: `terraform oracle cloud always free compute instance tutorial` (ou `terraform aws ec2 vpc tutorial`)
✔️ Confira: `terraform apply` cria a VM e você consegue SSH nela; `terraform destroy` apaga tudo; `apply` de novo recria.

**7.4 — cloud-init: a VM nasce pronta.** Configure o user_data/cloud-init da VM para, no primeiro boot, instalar Docker (ou k3s) e o necessário automaticamente. Combinado com o passo anterior: você destrói e recria um servidor FUNCIONAL em minutos, sem clicar em nada.
🔍 Pesquise: `terraform cloud-init user_data install docker`
✔️ Confira: VM recém-criada já responde `docker ps` sem você instalar nada na mão.

**7.5 — Organize em módulo.** Extraia a receita "VM + rede + firewall" para um **módulo** reutilizável e use-o para criar dois ambientes (staging/prod) com variáveis diferentes.
🔍 Pesquise: `terraform modules tutorial`

**7.6 — Terraform no pipeline.** Crie um workflow que roda `terraform fmt -check`, `validate` e `plan` em todo PR que mexe na infra — e postar o plano como comentário do PR é o toque profissional. `apply` só manual/aprovado.
🔍 Pesquise: `terraform plan github actions pull request comment`

### ✅ Fase 7 pronta quando
- [ ] GitHub gerenciado por Terraform (repo, proteção, secrets)
- [ ] Remote state em bucket com a infra cloud
- [ ] `destroy` + `apply` recria a VM funcional do zero (cloud-init)
- [ ] `plan` rodando em PR no Actions

---

# FASE 8 — GitOps com ArgoCD (a cereja do bolo)

**Em palavras simples:** hoje seu pipeline EMPURRA mudanças para o cluster (push). No **GitOps**, inverte: um agente DENTRO do cluster (ArgoCD) fica observando um repositório git e PUXA as mudanças, mantendo o cluster sempre igual ao que está no git. O git vira a única fonte de verdade — quer saber o que está em produção? Olhe o repo. Quer voltar atrás? `git revert`.

**8.1 — Crie o repositório de "estado desejado":** um repo separado contendo o chart Helm (ou referência a ele) + os values de cada ambiente. Código da aplicação e configuração de deploy agora moram em repos diferentes — essa separação é o coração do GitOps.

**8.2 — Instale o ArgoCD no k3s** e acesse a interface web dele (via ingress ou port-forward).
🔍 Pesquise: `argocd install k3s getting started`
✔️ Confira: você loga na UI do ArgoCD.

**8.3 — Crie a Application:** o objeto do ArgoCD que diz "observe esse repo/pasta e mantenha o namespace X sincronizado com ele". Ative sync automático + self-heal.
🔍 Pesquise: `argocd application helm values automated sync self heal`
✔️ Confira: a aplicação aparece verde/Synced na UI, com o mapa de recursos desenhado.

**8.4 — Inverta o fluxo do CI.** O pipeline agora termina diferente: builda a imagem, publica no GHCR e **abre um commit/PR no repo de manifests atualizando a tag da imagem**. O ArgoCD percebe o commit e sincroniza o cluster. Você não roda mais `helm upgrade`/`kubectl` em pipeline nenhum.
🔍 Pesquise: `ci update image tag gitops repository pattern`
✔️ Confira: merge no repo do app → minutos depois, ArgoCD mostra o sync acontecendo → versão nova no ar.

**8.5 — Sinta os superpoderes:**
- **Anti-gambiarra:** mude algo direto no cluster (`kubectl edit deployment`, aumente réplicas na mão) e veja o ArgoCD acusar o drift e REVERTER sozinho (self-heal). Mudança agora só por git.
- **Rollback por git:** `git revert` do commit da tag + push = produção volta à versão anterior, com auditoria completa de quem/quando/por quê.
  ✔️ Confira: você fez os dois experimentos e viu acontecer.

### ✅ Fase 8 pronta quando
- [ ] ArgoCD sincronizando produção a partir do repo de manifests
- [ ] CI só builda e atualiza tag via commit — ninguém deploya "por fora"
- [ ] Drift manual é revertido sozinho
- [ ] Rollback via git revert testado

---

# FASE 9 — Polimento e portfólio

**9.1 — Documentação que vende:** README do projeto com um **diagrama da arquitetura completa** (excalidraw/draw.io: usuário → ingress → serviços → banco/fila → monitoramento → pipelines) e uma seção de decisões: "por que k3s", "por que monorepo", "por que GitOps". Recrutador técnico lê isso e já sabe que você pensa, não só copia.

**9.2 — Conte a história:** 2-3 posts (LinkedIn/dev.to) da jornada: "containerizando minha primeira aplicação", "meu primeiro deploy com HTTPS", "do compose ao GitOps". Quem documenta a jornada aparece — e em começo de carreira, aparecer importa.

**9.3 — Simule incidentes e cronometre:** disco cheio, pod morrendo por falta de memória (OOMKill), certificado expirado, migration que quebra. Cada incidente vira um runbook novo na pasta `runbooks/`.

**9.4 — Conecte seus projetos:** coloque o Sentinel monitorando o servidor, mandando alertas para o mesmo canal. Dois projetos seus conversando em produção é portfólio raro de júnior.

**9.5 — Revisão para entrevistas:** com tudo rodando, revise a teoria por trás do que você FEZ (fica 10x mais fácil): como funciona DNS, o handshake TLS, redes no Docker e no K8s, o ciclo de vida de um pod. Seu diferencial é a base de Linux/redes que você já tem — afie ela.

---

## Mapa: fase → competência no currículo

| Fase | O que você pode escrever no currículo |
|---|---|
| 0 | Hardening de servidores Linux, SSH, firewall, Git workflow |
| 1 | Docker avançado: multi-stage, otimização de imagens, Compose |
| 2 | CI com GitHub Actions, scan de vulnerabilidades (Trivy), GHCR |
| 3 | CD, proxy reverso, TLS/Let's Encrypt, backups, rollback |
| 4 | Prometheus, Grafana, Loki, PromQL, alerting, runbooks |
| 5 | Kubernetes: deployments, probes, ingress, storage, rolling updates |
| 6 | Helm: charts próprios, multi-ambiente |
| 7 | Terraform: IaC, remote state, módulos, cloud (OCI/AWS) |
| 8 | GitOps com ArgoCD |
| 9 | Documentação técnica, resposta a incidentes |

## As 5 regras de ouro (cole na parede)

1. **Um passo por vez.** Confira antes de avançar.
2. **Quebre de propósito.** O aprendizado mora no que falha.
3. **Anote tudo que resolver** no DIARIO.md.
4. **Checklist 100% antes da próxima fase.** Sem "quase pronto".
5. **Segredo nunca vai para o git.** Nunca. Nem "só dessa vez".