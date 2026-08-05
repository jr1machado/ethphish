# Arquitetura alvo e operação — v0.1.0

## Escopo desta arquitetura

O EthPhish separa claramente o estado atual da arquitetura-alvo. A v0.1.0 é um
monólito modular seguro para desenvolvimento, com PostgreSQL e broker
provisionados. A distribuição real de workers é futura e não deve ser assumida
como pronta apenas porque RabbitMQ está no Compose.

## Componentes atuais

```text
                    ┌─────────────────────────────────┐
                    │ Caddy / reverse proxy            │
internet ─ HTTPS ──►│ TLS público de desenvolvimento    │
                    └───────────────┬─────────────────┘
                                    │ TLS interno
                    ┌───────────────▼─────────────────┐
                    │ EthPhish server central          │
                    │ admin + API + web + scheduler    │
                    │ mailer/SMS worker interno + IMAP │
                    └───────┬──────────────┬──────────┘
                            │              │
                      PostgreSQL       RabbitMQ
                      migrations       provisionado
```

O servidor central contém interfaces administrativa e web, scheduler, worker
herdado, IMAP, relatórios e API. O painel administrativo só deve existir na
rede administrativa. A superfície web é a única roteada pelo proxy local.

## Arquitetura de workers planejada

```text
central server ── outbox transacional ──► RabbitMQ TLS ──► worker node 1
       │                                      │             worker node 2
       └─ PostgreSQL (única fonte de verdade) └───────────► worker node N
```

Cada node receberá apenas trabalhos autorizados e assinados, terá identidade e
segredos próprios, e não terá porta administrativa nem acesso direto ao
PostgreSQL. O central validará tenant, aprovação, domínio, quota e janela de
execução antes de publicar o job. Workers reportarão resultado idempotente ao
central por uma API interna autenticada ou fila de retorno, ainda a definir em
ADR futuro.

## Redes e portas

| Origem | Destino | Porta/protocolo | Estado |
| --- | --- | --- | --- |
| navegador autorizado | Caddy | 443/TCP (9443 local) HTTPS | atual |
| Caddy | web EthPhish | 8080/TCP HTTPS interno | atual |
| health checker | admin EthPhish | 3333/TCP HTTPS interno | atual |
| central | PostgreSQL | 5432/TCP PostgreSQL | atual, rede de dados |
| central/worker node | RabbitMQ | 5671/TCP AMQP TLS | futuro |
| worker node | relay SMTP/provedor HTTPS | 465/587/TCP ou HTTPS | futuro, somente destinos aprovados |
| worker node | central | API/fila de retorno privada | futuro |

Não publicar 3333, 5432, 5671/5672, métricas, debug ou profiler no host. A
porta 5672 não deverá ser usada por workers em produção; AMQP deve usar TLS em
5671 e autenticação por credenciais/identidade de node.

## Distribuição e crescimento

1. Inicie com um servidor central e worker interno, sem entregas externas.
2. Extraia transactional outbox e consumidor de jobs.
3. Adicione um worker-node em zona/rede aprovada, com limite de concorrência e
   quota próprios.
4. Escale horizontalmente adicionando nodes idênticos, nunca replicando o
   banco ou o painel administrativo nos workers.
5. Monitore backlog, taxa de sucesso, retries, DLQ, latência, CPU, memória e
   erros por node; pause automaticamente quando ultrapassar limites aprovados.

Referência inicial: central 2 vCPU/4 GB RAM/100 GB de disco; nodes 1 vCPU/1 GB
RAM/50 GB. Eleve a memória do central para 8 GB antes de ampliar volume de
relatórios, eventos ou tenants. Essas referências não substituem capacity
planning por campanha, retenção e SLO.

## Ferramentas e integrações

| Camada | Ferramentas | Função |
| --- | --- | --- |
| aplicação | Go 1.24.5, GORM legado, Goose | servidor, domínio e migrations |
| interface | Node 22, Gulp, Webpack | assets do frontend herdado |
| dados | PostgreSQL 17 | persistência e migrations |
| mensageria | RabbitMQ 4 | fundação de distribuição futura |
| borda | Caddy 2 | TLS e proxy web |
| relatórios | Python 3 | geração herdada Word/Excel |
| entrega | SMTP, SMS, IMAP | recursos herdados sob escopo autorizado |
| entrega de software | Docker Compose e GitHub Actions | ambiente reproduzível e CI |
| segurança | Gitleaks, govulncheck, Trivy, SBOM | scans e evidência de supply chain |

## Operação segura

- variáveis de ambiente e secret store para segredos, nunca arquivos versionados;
- TLS obrigatório para banco fora do desenvolvimento;
- backup antes de migration destrutiva e restore ensaiado em base isolada;
- logs sem DSN, senhas ou dados pessoais desnecessários;
- aprovação e autorização de campanhas antes de qualquer entrega;
- nenhum worker deve ter acesso administrativo ou ao banco.
