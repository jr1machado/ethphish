# Release notes — EthPhish v0.3.0

Data: 2026-08-06

## Resumo

Release de fundação multitenant e confiabilidade de entrega do EthPhish,
fork independente para testes éticos de phishing e quishing. Introduz
isolamento de dados por tenant garantido no banco (PostgreSQL RLS) e o
primeiro caminho de entrega assíncrono durável (RabbitMQ) para e-mail de
campanha. Não é uma autorização para uso em produção pública nem para
portal de clientes self-service — ver [Issues conhecidos](ISSUES_CONHECIDOS.md).

Substitui as notas da v0.2.0 (tag `v0.2.0`, commit `335b5b9`) como release
corrente; a v0.2.0 permanece disponível no histórico do repositório.

## Entregas

### Multitenancy e isolamento de dados

- fundação multitenant: tabelas `tenants`, `companies`, `tenant_users` e
  contexto de requisição tipado `TenantScope`, que bloqueia qualquer fluxo
  tenant-owned sem `tenant_id`/`user_id` validados;
- escopo por tenant aplicado a campanhas, grupos, alvos, templates, landing
  pages, perfis SMTP, perfis e templates SMS, IMAP, webhooks e relatórios,
  incluindo fluxos fora de requisição HTTP (monitor IMAP em background, MFA
  por SMS, geração de relatórios em fila);
- entrega de webhook de campanha corrigida para resolver o tenant dono antes
  do disparo, em vez de notificar webhooks de todos os tenants (bug
  pré-existente identificado durante os testes de integração desta sprint);
- **PostgreSQL RLS** com `FORCE ROW LEVEL SECURITY` em toda tabela
  tenant-owned, policy `tenant_isolation` baseada em `ethphish.tenant_id`
  (via `set_config` local à transação, em `withTenantTransaction`);
- role de runtime restrito **`ethphish_app`** (sem `SUPERUSER`/`BYPASSRLS`),
  criado por migration dedicada; o role privilegiado `ethphish` fica restrito
  ao passo `db-migrate` do Compose, que roda e termina antes do `server`
  subir — necessário porque o role padrão do Postgres é superusuário e
  ignora RLS mesmo com `FORCE ROW LEVEL SECURITY`;
- teste de integração que abre uma segunda conexão sob o role restrito e
  comprova, em PostgreSQL real, que leitura/escrita cruzada entre tenants é
  bloqueada — não apenas configurada.

### Confiabilidade de entrega

- disparo de e-mail de campanha passa de canal Go in-process para
  publicação por `MailLog` na fila durável `mail.send` (RabbitMQ), consumida
  por um pool de goroutines dentro do próprio processo `server`;
- fila de retry por TTL + dead-letter-exchange (sem plugin RabbitMQ) e fila
  morta terminal para falhas de processamento Go (crash de consumidor, banco
  inacessível), independente do retry SMTP já existente via backoff em
  `MailLog`;
- redelivery idempotente: sucesso e erro removem a linha `MailLog`, então
  uma mensagem redelivrada sem linha correspondente é um ack sem efeito;
- fallback automático para o canal direto quando `ETHPHISH_RABBITMQ_URL` não
  está definido — nenhum teste existente foi afetado;
- SMS e geração de relatórios permanecem, por decisão de escopo, no caminho
  de polling de banco já existente.

### Exposição administrativa

- admin UI passa a ser roteável pelo proxy reverso, em listener HTTPS
  dedicado (9444), segregado da web pública de campanhas (9443); a porta
  3333 do servidor nunca é publicada diretamente no host;
- correção de roteamento: uma tentativa inicial de montar o admin sob
  `/admin` no listener 9443 quebrou navegação pós-login (a aplicação emite
  redirects e links raiz-relativos); a solução final usa listener próprio.

### Plataforma

- migrations executadas por um passo `db-migrate` dedicado no Compose, com
  o role privilegiado, antes do `server` iniciar;
- correção de bug de boot: `MigrationsPath` resolvia para `db/db_<driver>`
  mas os arquivos SQL vivem em `db/db_<driver>/migrations`, causando falha
  "no SQL migrations found" após reconstrução da árvore de trabalho;
- RabbitMQ 4 com credenciais de desenvolvimento dedicadas (o usuário `guest`
  padrão só aceita conexões locais ao container).

## Integrações

| Integração | Estado v0.3.0 | Uso |
| --- | --- | --- |
| PostgreSQL | ativo, com RLS forçado | dados, migrations, isolamento por tenant |
| RabbitMQ | ativo no caminho crítico de e-mail | fila `mail.send` + retry/DLQ; SMS e relatórios ainda em polling |
| Caddy | ativo no Compose | TLS e proxy web + admin (9443/9444) |
| OIDC | recurso herdado configurável | autenticação administrativa, mediante IdP aprovado |
| SMTP, SMS, IMAP | recursos herdados, agora escopados por tenant | apenas em escopo autorizado; não configurados automaticamente |
| GitHub Actions | workflow incluído | exige publicação do workflow e token com escopo `workflow` |

## Upgrade e rollback

1. Faça backup com `./scripts/backup-postgres.sh` antes de atualizar.
2. Atualize a imagem e execute `docker compose up -d`; o passo `db-migrate`
   roda automaticamente antes do `server` subir.
3. Confirme que o `server` conecta como `ethphish_app` (não `ethphish`) —
   uma DSN apontando para o role privilegiado desativa a proteção de RLS
   silenciosamente.
4. Consulte `docker compose logs server db-migrate` para migrations e
   health.
5. Em falha, interrompa a atualização e restaure somente em banco isolado a
   partir do dump validado; não execute migrations `down` diretamente em
   produção.

## Limitações de release

Consulte [ISSUES_CONHECIDOS.md](ISSUES_CONHECIDOS.md). Workers distribuídos
em nodes externos, extensão da fila durável para SMS/relatórios, portal de
clientes self-service e operação externa permanecem fora desta release.
