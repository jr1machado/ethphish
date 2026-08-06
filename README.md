# EthPhish v0.3.0

EthPhish é um fork independente do Anglerphish 1.3.0, evoluído como uma
plataforma corporativa completa para testes éticos, autorizados e mensuráveis
de phishing e quishing (simulações via QR Code). Não é uma distribuição nem um
produto oficial do Anglerphish ou do Gophish.

O produto apoia programas de conscientização, GRC e redução de risco humano.
Toda campanha deve possuir autorização, escopo, período, público e domínios
aprovados. O EthPhish não contempla evasão de controles, coleta de credenciais
reais, payloads, exploração de vulnerabilidades ou uso não autorizado.

## Visão comercial

Para C-levels e executivos, o EthPhish converte o risco humano em evidência
operacional: mostra onde a organização é mais suscetível a engenharia social,
mede a evolução após treinamentos e dá governança sobre campanhas de
conscientização, com trilha de aprovação e isolamento de dados por cliente ou
unidade de negócio.

**Casos de uso**

- Programas recorrentes de security awareness com métricas comparáveis entre
  ciclos e áreas.
- Validação de controles humanos antes de auditorias (ISO 27001, PCI-DSS,
  LGPD/GDPR) com evidência exportável.
- Avaliação de risco por diretoria, unidade de negócio ou aquisição recente
  (due diligence de segurança pós-M&A).
- Simulações aprovadas de resposta a incidente real, sem reexpor a
  organização a coleta de credenciais verdadeiras.
- Provedores de serviços gerenciados (MSSPs) operando múltiplos clientes na
  mesma plataforma, com isolamento de dados garantido no banco (RLS), não
  apenas na aplicação.

**Dores que o EthPhish resolve**

- Falta de métricas consistentes e comparáveis de comportamento humano ao
  longo do tempo.
- Campanhas sem controle de escopo, autorização ou rastreabilidade de
  aprovação — risco jurídico e reputacional para quem executa o teste.
- Relatórios manuais, sem padronização, difíceis de apresentar a auditoria ou
  ao conselho.
- Ausência de uma base técnica que suporte múltiplos clientes/áreas sem risco
  de vazamento cruzado de dados entre eles.
- Dependência de operação manual do time de segurança para disparar e
  monitorar cada campanha.

A plataforma preserva as fronteiras éticas em todas as camadas: mede eventos
de interação autorizados (abertura, clique, submissão simulada, reporte) e
não armazena senhas, OTPs ou outras credenciais reais.

## O que mudou neste fork

Desde a base Anglerphish 1.3.0:

- identidade, versionamento e documentação próprios do EthPhish, sem vínculo
  com releases do Anglerphish ou do Gophish;
- PostgreSQL como único banco de runtime suportado (SQLite permanece apenas
  para testes e importação legada controlada);
- migrations PostgreSQL executadas por um role privilegiado dedicado
  (`db-migrate`, passo do Compose que roda e termina antes do `server`
  subir), com pool configurável e advisory lock contra concorrência;
- **fundação multitenant**: `tenants`, `companies`, `tenant_users` e um
  contexto de requisição tipado (`TenantScope`) que barra qualquer fluxo
  tenant-owned sem `tenant_id`/`user_id` validados;
- **escopo por tenant aplicado** a campanhas, grupos, alvos, templates,
  landing pages, perfis SMTP, perfis e templates SMS, IMAP, webhooks e
  relatórios — incluindo os fluxos que não passam por uma requisição HTTP
  (monitor IMAP em background, MFA por SMS, geração de relatórios em fila);
- **PostgreSQL Row-Level Security (RLS)** com `FORCE ROW LEVEL SECURITY` em
  toda tabela tenant-owned, aplicada por uma policy que lê
  `ethphish.tenant_id` da transação; a aplicação roda em runtime sob um role
  restrito (`ethphish_app`, sem `SUPERUSER`/`BYPASSRLS`) porque o role
  padrão do Postgres é superusuário e ignora RLS silenciosamente;
- **fila durável RabbitMQ no caminho crítico de e-mail**: o disparo de
  e-mail de campanha passou de canal Go in-process para publicação por
  `MailLog` na fila `mail.send`, com fila de retry por TTL/DLX e fila morta
  terminal para falhas de processamento (não confundir com o retry SMTP que
  já existia via backoff em `MailLog`); cai de volta ao canal direto quando
  `ETHPHISH_RABBITMQ_URL` não está definido;
- **admin UI publicada por proxy reverso** em listener próprio (9444),
  segregada da web pública de campanhas (9443), sem expor a porta 3333
  diretamente no host;
- Docker multi-stage, usuário não-root, capabilities removidas e health
  checks em todos os serviços do Compose;
- TLS autoassinado para desenvolvimento, roteado pelo Caddy;
- CI com formatação, vet, testes, integração PostgreSQL, secret scan,
  vulnerabilidades, scan de imagem e SBOM;
- backup PostgreSQL, restore ensaiado e documentação de segurança,
  arquitetura e governança.

## Arquitetura

```text
                         rede administrativa privada
                                      │
                          VPN / Zero Trust / OIDC
                                      │
                    internet ── HTTPS 9443/9444 ── Caddy (reverse-proxy)
                                      │ TLS interno
                           ┌──────────▼──────────┐
                           │ servidor central     │
                           │ admin (3333) + web   │
                           │ (8080) + API         │
                           │ scheduler, IMAP,     │
                           │ relatórios, worker    │
                           │ interno (pool mail)   │
                           └───┬───────┬──────┬───┘
                               │       │      │
                     PostgreSQL│  AMQP │      │futuro: AMQP TLS 5671
                          (RLS)│ 5672  │      ▼
                          ┌────▼───┐ ┌─▼──────────┐  ┌───────────────┐
                          │ dados  │ │ RabbitMQ    │─▶│ worker node 1 │
                          │(ethphish│ │ mail.send   │  │ worker node N │
                          │_app role│ │ + retry/DLQ │  │  (futuro)     │
                          └────────┘ └─────────────┘  └───────────────┘
```

Na v0.3.0, o servidor central concentra admin, API, web de campanhas,
scheduler e um **pool de goroutines interno** que consome a fila
`mail.send`; não existe ainda um processo worker externo ao servidor. A
separação real em worker nodes (identidade e segredos próprios, sem acesso
direto ao PostgreSQL, comunicação por AMQP TLS 5671) permanece como
arquitetura-alvo — RabbitMQ hoje já está no caminho crítico de entrega de
e-mail, mas isso não deve ser confundido com workers distribuídos.

Consulte o detalhamento em [arquitetura alvo](docs/architecture/target-architecture.md).

## Containers e portas

| Componente | Papel | Portas | Exposição |
| --- | --- | --- | --- |
| `reverse-proxy` (Caddy) | TLS público e roteamento web + admin | 9443/TCP web, 9444/TCP admin, publicadas no host | somente para redes autorizadas (VPN/Zero Trust); nunca publicar 9444 na internet aberta |
| `server` | administração, API, campanhas, worker interno de e-mail | 3333/TCP admin, 8080/TCP web | somente redes Docker internas (`admin_internal`, `application_internal`) |
| `db-migrate` | aplica migrations com role privilegiado, roda uma vez e termina | nenhuma publicada | rede de dados interna, sem rede externa |
| `postgres` | dados, RLS e migrations | 5432/TCP | rede de dados interna (`data_internal`) |
| `rabbitmq` | fila durável de e-mail em produção | 5672/TCP AMQP, 5671/TCP AMQP TLS (worker nodes futuros) | rede de dados interna |
| `tls-init` | prepara volume privado de certificados | nenhuma | execução única, sem rede |
| `worker-node` (futuro) | entrega aprovada e observabilidade, sem painel nem banco direto | AMQP TLS 5671 de saída; SMTP/HTTPS conforme escopo | rede privada dedicada a workers |

## Comunicação servidor central ↔ workers

- **Hoje**: o "worker" é um pool de goroutines dentro do processo `server`,
  que consome `mail.send` do RabbitMQ pela rede de dados interna (5672/TCP,
  sem TLS em desenvolvimento). Não há uma porta de comunicação central↔worker
  externa porque não há processo worker externo.
- **Arquitetura-alvo**: workers como processos/containers independentes,
  sem acesso administrativo nem ao PostgreSQL, autenticando e consumindo
  jobs assinados via AMQP TLS na porta 5671, em rede privada dedicada. O
  central valida tenant, aprovação, domínio, quota e janela de execução
  antes de publicar o job; o retorno de resultado (idempotente) usa uma API
  interna autenticada ou fila de retorno, a definir em ADR futuro.
- **Crescimento como nodes**: cada worker node novo é idêntico aos demais
  (mesma imagem, sem estado próprio além de credenciais/identidade),
  consome da mesma fila e escala horizontalmente sem replicar banco ou
  painel administrativo. Backlog, taxa de sucesso, retries, DLQ, latência,
  CPU e memória por node devem ser monitorados para decidir quando adicionar
  o próximo node.

## Recursos implementados até a v0.3.0

**Herdados do baseline Anglerphish 1.3.0** (ver detalhamento completo em
[FEATURES.md](FEATURES.md)):

- campanhas por e-mail, SMS, QR Code e canal genérico, campaign sets;
- gestão de grupos, participantes, templates, landing pages e perfis de
  envio (SMTP e SMS — Twilio/Vonage);
- eventos de abertura, clique, submissão simulada, reporte e resposta;
- simulação de MFA, landing pages com Basic Auth, variáveis de template e
  variáveis globais;
- monitor IMAP com leitura de mensagens reportadas e aba de respostas;
- relatórios Word/Excel com opção de anonimização;
- OIDC para SSO administrativo, CSRF, autenticação e testes de
  caracterização do baseline;
- criptografia AES-256-GCM de campos sensíveis em banco.

**Entregues nas sprints 01–04 (EthPhish)**:

- runtime e imagem exclusivos para PostgreSQL 17, sem CGO, sem Goose;
- migrations idempotentes, pool configurável, advisory lock, TLS de banco
  opcional (`ETHPHISH_DB_REQUIRE_TLS=true`) e health/readiness sem dados de
  conexão (`/healthz`, `/readyz`);
- backup lógico e restore ensaiado em banco isolado;
- pré-flight, preparador de schema isolado e importador transacional
  SQLite→PostgreSQL, validados em CI contra PostgreSQL efêmero;
- fundação multitenant (`tenants`, `companies`, `tenant_users`,
  `TenantScope`) e escopo por tenant em **todas** as entidades de negócio
  prioritárias (campanhas, grupos, alvos, templates, landing pages, SMTP,
  SMS, IMAP, webhooks, relatórios);
- **PostgreSQL RLS com `FORCE ROW LEVEL SECURITY`** e role de runtime
  restrito (`ethphish_app`), com teste de integração que abre uma segunda
  conexão sob o role restrito e comprova bloqueio de leitura/escrita
  cruzada entre tenants;
- **fila RabbitMQ durável** para disparo de e-mail de campanha, com retry
  por TTL/DLX e fila morta terminal, isolada do retry SMTP existente;
- admin UI acessível por proxy reverso dedicado (9444), sem publicar a
  porta 3333 diretamente;
- CI/CD e controles de supply chain descritos em
  [release notes](RELEASE_NOTES.md).

## Funções e recursos futuros

- workers externos ao processo do servidor, distribuídos em nodes próprios,
  sem acesso direto ao banco, comunicando por AMQP TLS 5671;
- extensão do padrão de fila durável para SMS e geração de relatórios (hoje
  ainda em polling de banco, fora de escopo desta release por design);
- transactional outbox para garantir publicação exatamente-uma-vez entre
  commit de banco e publicação na fila;
- portal de clientes multitenant self-service, com fluxo auditável de
  aprovação de campanhas por tenant;
- dashboard operacional e executivo, métricas de capacidade e
  observabilidade (backlog, DLQ, latência, node saúde);
- bundles versionados de campanhas e conteúdos de treinamento, com
  importação/exportação controlada;
- backup automatizado, retenção, restauração recorrente e storage externo;
- alta disponibilidade, assinatura de imagens e política formal de
  atualização de dependências.

## Fora de escopo

- coleta, retenção ou recuperação de senhas, OTPs, cartões ou credenciais
  reais;
- evasão de filtros, antivírus, EDR, gateways ou mecanismos de proteção;
- payloads, exploração, comprometimento de contas ou acesso não autorizado;
- campanhas sem aprovação formal, domínios fora do escopo ou público não
  autorizado;
- exposição pública do painel administrativo fora de rede autorizada
  (VPN/Zero Trust), mesmo estando disponível via proxy reverso.

## Requisitos

- Go 1.25.12 e compilador C (compatibilidade SQLite legada nos testes);
- Docker Engine com Compose v2;
- Node 22 para reconstrução de assets;
- PostgreSQL 17 para o ambiente Compose;
- RabbitMQ 4 para o ambiente Compose (fila durável de e-mail; a aplicação
  cai de volta ao envio direto in-process se `ETHPHISH_RABBITMQ_URL` não for
  definido, mas isso não é recomendado além de desenvolvimento local).

O runtime de produção exige PostgreSQL. SQLite permanece somente para testes
e migração legada explicitamente controlada; não inicie uma imagem de
produção com `ETHPHISH_DB_DRIVER=sqlite3`.

Referência de capacidade inicial: servidor central com 2 vCPU, 4 GB RAM e
100 GB de disco (eleve a memória para 8 GB antes de ampliar volume de
tenants, eventos ou relatórios); cada worker futuro com 1 vCPU, 1 GB RAM e
50 GB de disco. A capacidade real deve ser dimensionada por volume, taxa de
entrega aprovada, retenção de eventos e geração de relatórios.

## Desenvolvimento local

```sh
docker compose build
docker compose up -d
docker compose ps
```

- Web de campanhas/quishing: `https://localhost:9443`.
- Painel administrativo: `https://localhost:9444` (publicado apenas em
  desenvolvimento local; em produção deve ficar restrito a rede
  administrativa/VPN, nunca exposto na internet aberta).

A credencial temporária de administrador é gerada apenas no primeiro início
e deve ser alterada imediatamente.

```sh
CGO_ENABLED=1 go test ./...
./scripts/backup-postgres.sh
```

Consulte [desenvolvimento local](docs/runbooks/local-development.md),
[release notes](RELEASE_NOTES.md), [issues conhecidos](ISSUES_CONHECIDOS.md) e
[política de segurança](SECURITY.md).

## Documentação

- [Arquitetura alvo](docs/architecture/target-architecture.md)
- [Inventário de dependências](docs/architecture/dependency-inventory.md)
- [Threat model](docs/security/threat-model.md)
- [Uso aceitável](docs/product/acceptable-use.md)
- [Status das Sprints 0–2](docs/project/sprint-0-2-status.md)
- [Sprint 02 — endurecimento PostgreSQL](docs/project/sprint-02.md)
- [Sprint 04 — fundação multitenant](docs/project/sprint-04.md)
- [Funções e recursos herdados detalhados](FEATURES.md)
- [Release notes v0.3.0](RELEASE_NOTES.md)
- [Issues conhecidos](ISSUES_CONHECIDOS.md)

## Licença e atribuição

O EthPhish mantém a atribuição ao Gophish e ao Anglerphish nos termos da
licença MIT herdada. As mudanças deste repositório formam um fork independente
e não implicam endosso dos projetos upstream.
