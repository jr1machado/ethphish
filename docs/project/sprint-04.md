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

1. Aplicação do middleware de seleção autorizada às rotas tenant-owned.
2. Consultas escopadas nas entidades prioritárias restantes. `tenant_id` e o
   retropreenchimento controlado já foram aplicados a campanhas, grupos,
   templates, landing pages, SMTP, SMS, IMAP, webhooks e relatórios.
3. RLS PostgreSQL transacional e testes negativos de acesso cruzado em banco
   isolado.

O executor transacional para RLS já está ativo nos fluxos de templates e
landing pages. A política RLS só será habilitada quando os fluxos remanescentes
que acessam as mesmas tabelas estiverem migrados, para não interromper workers
legados.

Nenhuma rota de negócio é declarada multitenant enquanto ainda não consumir o
`TenantScope`; essa regra evita uma alegação prematura de isolamento.
