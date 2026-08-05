# ADR-002: PostgreSQL como banco de produção

- Status: aceito
- Data: 2026-08-04

PostgreSQL será a fonte de verdade em produção. SQLite existirá somente durante
a transição, em testes legados e como origem da ferramenta de migração. A adoção
será precedida por testes de caracterização e terá migrations, backup, restore e
rollback documentados.
