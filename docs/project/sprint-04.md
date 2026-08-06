# Sprint 04 — Fundação multitenant

## Início

Branch: `feature/sprint04`. Esta sprint começa pelo modelo de dados: tenants,
empresas e vínculos administrativos. Nenhuma tabela de negócio recebe
`tenant_id` até que o contexto de requisição, RLS e testes de isolamento sejam
introduzidos juntos.

## Primeiro incremento

- migrations PostgreSQL para `tenants`, `companies` e `tenant_users`;
- relação explícita usuário–tenant–empresa, sem alterar ainda as consultas
  legadas;
- modelos de domínio para administrar os vínculos e consulta de concessões
  ativas;
- contexto tipado de requisição (`TenantScope`), que impede a continuação de
  fluxos tenant-owned sem `tenant_id` e `user_id` validados;
- teste de integração PostgreSQL para a criação e leitura das três relações,
  além de testes unitários para rejeitar contexto ausente ou inválido;
- teste negativo em PostgreSQL: a seleção de um tenant sem concessão retorna
  `403 Forbidden`.

Evidência executável: `INFO/Sprint-04/ETH-401-405-foundation-validation.md`.

## Próximos incrementos

1. Aplicação do middleware de seleção autorizada às rotas tenant-owned. Concluído.
2. Consultas escopadas nas entidades prioritárias restantes. `tenant_id` e o
   retropreenchimento controlado foram aplicados a campanhas, grupos,
   templates, landing pages, SMTP, SMS, IMAP, webhooks e relatórios, incluindo
   os fluxos remanescentes que não passavam por request HTTP (monitor IMAP em
   background, fluxo de MFA por SMS, geração de relatórios em fila). Concluído.
3. RLS PostgreSQL transacional e testes negativos de acesso cruzado em banco
   isolado. Concluído — ver `20260806000004_enable_tenant_row_level_security.sql`
   e `20260806000005_create_restricted_app_role.sql`.

O executor transacional para RLS está ativo em todo caminho tenant-owned
(`withTenantTransaction`, que define `ethphish.tenant_id` via `set_config`
local à transação). A policy `tenant_isolation` foi habilitada em campanhas,
grupos, targets, templates, pages, smtp, sms_profiles, sms_templates, imap,
webhooks e reports, com `FORCE ROW LEVEL SECURITY`. Sessões sem
`ethphish.tenant_id` definido continuam vendo todas as linhas — isso é
intencional e preserva os workers legados (monitor IMAP externo, drenagem da
fila de relatórios, limpeza agendada) sem exigir migração deles.

Achado crítico durante a implementação: a imagem oficial do PostgreSQL cria
`POSTGRES_USER` como superusuário, e superusuário ignora RLS mesmo com
`FORCE ROW LEVEL SECURITY`. A aplicação agora conecta em runtime com o role
restrito `ethphish_app` (sem `SUPERUSER`/`BYPASSRLS`, criado pela migration
005); migrations continuam rodando com o role privilegiado `ethphish`, agora
via um passo `db-migrate` dedicado no `compose.yaml` que roda antes do
`server`. Sem essa separação de role, a policy RLS existiria só no papel.

Nenhuma rota de negócio é declarada multitenant enquanto ainda não consumir o
`TenantScope`; essa regra evita uma alegação prematura de isolamento.
