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

## Segundo incremento — entidades prioritárias e templates

Validado no mesmo banco isolado:

- a migração cria o tenant legado, retropreenche `tenant_id` e aplica a coluna
  a `campaigns`, `groups`, `templates`, `pages`, `smtp`, `sms_profiles`,
  `sms_templates`, `imap`, `webhooks` e `reports`;
- um template criado com `tenant_id` explícito preserva o vínculo ao ser relido
  no PostgreSQL;
- o mesmo usuário recebeu concessões em dois tenants e um template do segundo
  não foi retornado por `GetTemplateForTenant` sob o escopo do primeiro;
- a seleção de tenant continua rejeitando uma concessão inexistente com 403.

Resultado do teste: `PASS`. A mensagem `record not found` no log foi a
verificação negativa esperada para o template de outro tenant.

## Preparação transacional para RLS

O fluxo tenant-scoped de templates passou a abrir uma transação por operação e
executar `set_config('ethphish.tenant_id', ..., true)` em PostgreSQL. O escopo
é local à transação, portanto não persiste em uma conexão reutilizada pelo pool.
O teste isolado foi repetido após essa alteração com resultado `PASS`.

## Terceiro incremento — landing pages

As operações de criação, listagem, busca, alteração e exclusão de landing
pages da API agora recebem `TenantScope` e usam o executor transacional. No
PostgreSQL isolado, uma página criada no tenant A não foi retornada quando a
consulta usou o tenant B. Resultado: `PASS`; os dois logs `record not found`
correspondem às verificações negativas de template e página.

## Limite atual

O middleware de tenant é aplicado à API e o fluxo de templates já consome o
escopo em criação, leitura, alteração e exclusão. Os demais fluxos de negócio
ainda precisam migrar para consultas escopadas; RLS PostgreSQL só será ativado
quando essas transações receberem o contexto de tenant de forma segura.
