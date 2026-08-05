# Evidências — Sprint 03: PostgreSQL, paridade funcional e migração

Este diretório reúne somente evidências reproduzíveis da Sprint 03. Um item só
passa para **validado** após o comando indicado concluir com sucesso e o
resultado ser registrado em arquivo versionado ou no CI vinculado ao commit.

| ID | Evidência exigida | Estado atual |
| --- | --- | --- |
| ETH-301 | migrations PostgreSQL e build sem CGO | em validação pelo CI do commit `05b640b` |
| ETH-302 | campanhas e resultados em PostgreSQL | teste de integração existente; ampliar reconciliação |
| ETH-303 | templates e landing pages em PostgreSQL | coberto pelo teste de campanha; evidência dedicada pendente |
| ETH-304 | SMTP, SMS e IMAP em PostgreSQL | SMS coberto; SMTP/IMAP pendentes |
| ETH-305 | criptografia com chave correta/incorreta | pendente |
| ETH-306 | importação SQLite → PostgreSQL | implementada; validação PostgreSQL efêmera pendente |
| ETH-307 | contagens, hashes, órfãos e referências | contagens pós-importação implementadas; hashes/órfãos pendentes |
| ETH-308 | backup PostgreSQL | ensaio registrado em `restore-isolado-2026-08-05.md` |
| ETH-309 | restore em banco isolado | aprovado no banco descartável; evidência registrada |
| ETH-310 | imagem sem SQLite/CGO | build local sem CGO aprovado; validar CI do commit `05b640b` |

## Comandos de verificação

```sh
CGO_ENABLED=0 go build -trimpath -o /tmp/ethphish-postgres-only .
go test ./tests/integration -run TestPostgresMigrationsAndReadiness -count=1
./scripts/backup-postgres.sh
```

Para um destino isolado de importação, use `postgres-schema-prepare`; ele não
executa o bootstrap do servidor nem cria usuários administrativos.

Os relatórios de pré-flight, importação, reconciliação e restore devem ser
gravados nesta pasta sem credenciais, dados pessoais, conteúdo de campanhas ou
outros dados de produção.
