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
- próximos itens: contexto de tenant, RLS e testes negativos de acesso cruzado.
