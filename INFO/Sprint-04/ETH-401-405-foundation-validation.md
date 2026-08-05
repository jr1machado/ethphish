# Sprint 04 — evidência ETH-401 a ETH-405

Data: 2026-08-05

## Escopo validado

- ETH-401: cadastro persistente de tenant ativo;
- ETH-402: empresa vinculada exclusivamente ao tenant criado;
- ETH-403: concessão explícita de usuário para tenant e empresa;
- ETH-405 (incremento inicial): contexto de tenant tipado e seleção
  autorizada via `X-EthPhish-Tenant-ID`.

## Ambiente isolado

Banco temporário: `ethphish_sprint04_test`, no serviço PostgreSQL interno do
`compose.yaml`. Nenhuma integração de SMTP, IMAP, SMS ou webhook foi acionada.
As migrations usadas foram as presentes no commit em validação, copiadas para
um diretório temporário do ambiente de teste.

## Comando executado

```text
go test ./tests/integration -run '^TestPostgresTenantFoundation$' -count=1
```

O binário foi executado no contêiner já pertencente à rede privada do banco,
com `ETHPHISH_TEST_POSTGRES_DSN` apontando para o banco temporário.

## Resultado

```text
Please login with the username admin and the password integration-test-password
PASS
```

O teste cria tenant, empresa e `tenant_user`; relê as relações no PostgreSQL;
aceita a seleção do tenant concedido e rejeita a seleção de outro tenant com
HTTP 403. Os testes unitários de `context` e `middleware` também passaram.

## Limite deste incremento

As rotas de negócio ainda não consomem o `TenantScope`. Portanto, este registro
não afirma isolamento completo de campanhas, templates, grupos ou resultados.
Essa proteção será adicionada junto de `tenant_id`, consultas escopadas e RLS.
