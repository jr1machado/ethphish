# Arquitetura alvo e operação — EthPhish v0.3.0

## Escopo desta arquitetura

O EthPhish separa claramente o estado atual da arquitetura-alvo. A v0.3.0 é
um monólito modular com isolamento multitenant no banco (RLS) e um primeiro
caminho de entrega assíncrono durável via RabbitMQ. A distribuição real de
workers em nodes externos é futura e não deve ser assumida como pronta
apenas porque a fila já está no caminho crítico de e-mail.

## Componentes atuais

```text
                    ┌─────────────────────────────────────┐
                    │ Caddy / reverse proxy                │
internet ─ HTTPS ──►│ 9443 web pública, 9444 admin (dev)   │
                    └───────────────┬───────────────┬─────┘
                            TLS interno         TLS interno
                    ┌───────────────▼───────────────▼─────┐
                    │ EthPhish server central              │
                    │ admin (3333) + web (8080) + API      │
                    │ scheduler + IMAP + relatórios         │
                    │ pool de goroutines consumindo         │
                    │ mail.send (worker interno)            │
                    └───────┬──────────────┬───────────────┘
                            │              │
                      PostgreSQL       RabbitMQ
                      RLS forçado,     mail.send +
                      role ethphish_app retry/DLQ
```

O servidor central contém interfaces administrativa e web, scheduler, IMAP,
relatórios, API e o consumidor da fila de e-mail. O painel administrativo
deve existir apenas na rede administrativa (VPN/Zero Trust), mesmo estando
acessível por um listener HTTPS próprio no proxy (9444). A superfície web de
campanhas é a única destinada a tráfego público.

Migrations rodam sob o role privilegiado `ethphish` apenas no passo
`db-migrate` do Compose, que completa antes do `server` subir; o `server` em
runtime conecta sempre com o role restrito `ethphish_app` (sem
`SUPERUSER`/`BYPASSRLS`), porque o role padrão criado pela imagem oficial do
Postgres é superusuário e ignora RLS mesmo com `FORCE ROW LEVEL SECURITY`.

## Isolamento multitenant (RLS)

- toda tabela tenant-owned (campanhas, grupos, targets, templates, pages,
  smtp, sms_profiles, sms_templates, imap, webhooks, reports) tem
  `FORCE ROW LEVEL SECURITY` e a policy `tenant_isolation`;
- `withTenantTransaction` define `ethphish.tenant_id` via `set_config` local
  à transação antes de qualquer leitura/escrita tenant-owned;
- sessões sem `ethphish.tenant_id` definido enxergam todas as linhas — é
  intencional, para preservar workers legados que ainda não migraram
  (monitor IMAP externo, drenagem da fila de relatórios, limpeza agendada);
- validado por teste de integração que abre uma segunda conexão sob o role
  restrito e comprova bloqueio de leitura/escrita cruzada entre tenants em
  PostgreSQL real, não apenas configuração declarada.

## Fila durável de e-mail (RabbitMQ)

- o mailer publica um evento por `MailLog` na fila `mail.send`; um pool de
  goroutines dentro do processo `server` consome e envia;
- falha de processamento (não um erro SMTP modelado) vai para
  `mail.send.retry` (TTL + dead-letter-exchange, sem plugin) e retorna a
  `mail.send`; após `MaxProcessingRetries` esgotados, vai para
  `mail.send.dead` para triagem manual;
- redelivery é idempotente: sucesso e erro removem a linha `MailLog`, então
  reentrega sem linha correspondente é um ack sem efeito;
- sem `ETHPHISH_RABBITMQ_URL` definido, o worker cai de volta ao canal Go
  in-process original;
- SMS e geração de relatórios permanecem no polling de banco já existente,
  fora de escopo desta release.

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
ADR futuro. O outbox transacional (publicação atômica com o commit do
`MailLog`) é pré-requisito antes de tratar a entrega como exactly-once.

## Redes e portas

| Origem | Destino | Porta/protocolo | Estado |
| --- | --- | --- | --- |
| navegador autorizado | Caddy | 9443/TCP HTTPS (web pública) | atual |
| operador autorizado (VPN/Zero Trust) | Caddy | 9444/TCP HTTPS (admin) | atual |
| Caddy | web EthPhish | 8080/TCP HTTPS interno | atual |
| Caddy | admin EthPhish | 3333/TCP HTTPS interno | atual |
| central (worker interno) | RabbitMQ | 5672/TCP AMQP (rede de dados interna) | atual |
| central | PostgreSQL (`ethphish_app`) | 5432/TCP PostgreSQL | atual, rede de dados |
| `db-migrate` | PostgreSQL (`ethphish`, privilegiado) | 5432/TCP PostgreSQL | atual, execução única |
| worker node externo | RabbitMQ | 5671/TCP AMQP TLS | futuro |
| worker node | relay SMTP/provedor HTTPS | 465/587/TCP ou HTTPS | futuro, somente destinos aprovados |
| worker node | central | API/fila de retorno privada | futuro |

Não publicar 3333, 5432, 5672 (exceto dentro da rede de dados interna),
métricas, debug ou profiler no host. A porta 9444 (admin) é publicada pelo
Compose de desenvolvimento apenas por conveniência local; em produção deve
ser restrita por firewall/security group à rede administrativa. Workers
futuros devem usar AMQP TLS na porta 5671, nunca 5672, com autenticação por
credenciais/identidade de node.

## Distribuição e crescimento

1. Servidor central com worker interno (pool de goroutines) consumindo
   `mail.send` — estado atual desta release.
2. Adicionar transactional outbox para publicação exatamente-uma-vez.
3. Extrair um worker-node externo em zona/rede aprovada, com limite de
   concorrência e quota próprios, consumindo por AMQP TLS 5671.
4. Escalar horizontalmente adicionando nodes idênticos, nunca replicando o
   banco ou o painel administrativo nos workers.
5. Monitorar backlog, taxa de sucesso, retries, DLQ, latência, CPU, memória e
   erros por node; pausar automaticamente quando ultrapassar limites
   aprovados.

Referência inicial: central 2 vCPU/4 GB RAM/100 GB de disco; nodes 1 vCPU/1 GB
RAM/50 GB. Eleve a memória do central para 8 GB antes de ampliar volume de
relatórios, eventos ou tenants. Essas referências não substituem capacity
planning por campanha, retenção e SLO.

## Ferramentas e integrações

| Camada | Ferramentas | Função |
| --- | --- | --- |
| aplicação | Go 1.25.12, GORM legado | servidor, domínio e migrations |
| interface | Node 22, Gulp, Webpack | assets do frontend herdado |
| dados | PostgreSQL 17 com RLS forçado | persistência, migrations e isolamento por tenant |
| mensageria | RabbitMQ 4 | fila durável de e-mail (`mail.send`); base para workers distribuídos futuros |
| borda | Caddy 2 | TLS e proxy web + admin |
| relatórios | Python 3 | geração herdada Word/Excel |
| entrega | SMTP, SMS, IMAP | recursos herdados sob escopo autorizado, agora escopados por tenant |
| entrega de software | Docker Compose e GitHub Actions | ambiente reproduzível e CI |
| segurança | Gitleaks, govulncheck, Trivy, SBOM | scans e evidência de supply chain |

## Operação segura

- variáveis de ambiente e secret store para segredos, nunca arquivos versionados;
- TLS obrigatório para banco fora do desenvolvimento;
- runtime do `server` sempre com o role restrito (`ethphish_app`); role
  privilegiado (`ethphish`) restrito ao passo `db-migrate`;
- backup antes de migration destrutiva e restore ensaiado em base isolada;
- logs sem DSN, senhas ou dados pessoais desnecessários;
- aprovação e autorização de campanhas antes de qualquer entrega;
- porta administrativa (9444) restrita a rede autorizada, nunca exposta na
  internet aberta;
- nenhum worker deve ter acesso administrativo ou ao banco.
