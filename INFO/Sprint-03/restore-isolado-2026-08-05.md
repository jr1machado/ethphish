# Ensaio de backup e restore isolado

- Data: 2026-08-05 (UTC)
- Escopo: ambiente local Docker Compose
- Origem: dump lógico PostgreSQL criado por `scripts/backup-postgres.sh`
- Destino: banco descartável `ethphish_restore_verify`
- Base ativa `ethphish`: não alterada

## Comandos executados

```sh
./scripts/backup-postgres.sh
./scripts/restore-postgres-isolated.sh <dump-gerado>
```

## Resultado

O restore isolado foi concluído sem erro. A consulta de validação retornou a
versão mais recente aplicada: `20260805000002`.

Esta evidência não contém o caminho do dump, credenciais, dados pessoais ou
conteúdo de campanhas. O banco de destino pode ser descartado após a revisão.
