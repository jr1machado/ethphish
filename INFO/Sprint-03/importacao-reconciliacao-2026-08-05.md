# Validação de importação SQLite → PostgreSQL

- Data: 2026-08-05 (UTC)
- Commit validado: `50c6b8c`
- Workflow: GitHub Actions `31041290294`
- Dados: exclusivamente sintéticos

## Cenário aprovado

1. Criar SQLite temporária com uma tabela `users` e um registro sintético.
2. Criar um banco PostgreSQL descartável.
3. Aplicar apenas o schema PostgreSQL, sem bootstrap do servidor.
4. Abrir SQLite em modo somente leitura e executar a importação aprovada.
5. Verificar reconciliação de contagem e preservação do ID `7` no PostgreSQL.

## Resultado

O job `PostgreSQL migrations` foi aprovado. O workflow completo também aprovou
qualidade/testes, scan de segredos, build/scan de container e SBOM.

O teste não usa dados de clientes, credenciais nem conteúdo de campanhas.
