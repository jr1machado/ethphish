# Desenvolvimento local

## Pré-requisitos

Antes de iniciar qualquer sprint de código, instale e valide as ferramentas de
build abaixo. Este é um gate inicial obrigatório: nenhuma alteração deve ser
considerada pronta sem conseguir executar a suíte local.

- Go **1.25.12** (`go version` deve retornar `go1.25.12`);
- compilador C compatível com GCC (`gcc --version`), exigido pelo driver SQLite
  legado enquanto a Sprint 2 não concluir a transição para PostgreSQL;
- Docker com Compose v2;
- Node 22 para reconstruir os assets do frontend.

Em ambientes com diretório pessoal somente-leitura, execute os comandos Go com
caches graváveis fora do repositório:

```sh
export GOCACHE=/tmp/ethphish-go-build
export GOPATH=/tmp/ethphish-go
go test ./...
```

## Inicialização

```sh
docker compose build
docker compose up -d
docker compose ps
```

A superfície pública fica em `https://localhost:9443` com certificado local do
Caddy. O Caddy e as duas interfaces da aplicação usam TLS no desenvolvimento:
na primeira inicialização, o servidor gera pares autoassinados separados para
administração e web em um volume Docker privado (`ethphish-tls`). As chaves não
são gravadas no repositório. A porta administrativa 3333 não é publicada no
host. Para depuração, acesse-a somente por uma rede administrativa
deliberadamente configurada.

Os workers atuais são goroutines do processo administrativo e não possuem uma
interface HTTP própria; por isso não recebem um certificado independente.
Conexões SMTP e IMAP continuam validando os certificados remotos por padrão.
Não habilite a opção de ignorar erros de certificado fora de um ambiente de
teste explicitamente autorizado.

PostgreSQL e RabbitMQ são provisionados como fundação das próximas sprints, mas
o servidor usa PostgreSQL no Compose desde a Sprint 2. Nenhuma entrega externa
deve ser configurada neste ambiente.

As imagens de runtime definem `ETHPHISH_RUNTIME_ENV=production` e recusam
SQLite. Para executar uma migração legada local, use uma ferramenta de migração
em ambiente isolado; não altere esse ambiente na imagem ou no deploy de
produção.

Em ambientes fora do desenvolvimento, use um DSN PostgreSQL com TLS e defina
`ETHPHISH_DB_REQUIRE_TLS=true`. Com essa opção, a aplicação se recusa a iniciar
se o DSN usar `sslmode=disable`.

## Migrations e recuperação PostgreSQL

O servidor executa as migrations ao iniciar. Para validar uma base vazia, use
`docker compose up -d` e consulte `docker compose logs server`. A repetição da
inicialização é segura: o Goose registra as versões aplicadas.

Antes de uma migration destrutiva fora do desenvolvimento, crie um backup:

```sh
./scripts/backup-postgres.sh
```

O script produz um dump PostgreSQL em formato customizado fora do repositório
(`$ETHPHISH_BACKUP_DIR`, ou `/tmp/ethphish-backups` por padrão), com permissão
restrita ao usuário atual. Para restaurar somente em uma base vazia e validada:

```sh
pg_restore --clean --if-exists --no-owner --no-privileges \
  -h localhost -U ethphish -d ethphish /caminho/ethphish-AAAAMMDDTHHMMSSZ.dump
```

O rollback de uma versão deve ser ensaiado primeiro em uma cópia do backup; não
execute migrations `down` diretamente em produção. Na inicialização,
instâncias PostgreSQL usam um advisory lock para impedir migrations
concorrentes.

## Reversão

Pare os containers com `docker compose down`. A remoção dos volumes é destrutiva
e deve ser solicitada explicitamente; não faz parte do procedimento normal.
