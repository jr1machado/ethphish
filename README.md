# EthPhish v0.2.0

EthPhish é um fork independente do Anglerphish 1.3.0, evoluído como uma
plataforma corporativa completa para testes éticos, autorizados e mensuráveis
de phishing e quishing (simulações via QR Code). Não é uma distribuição nem um
produto oficial do Anglerphish ou do Gophish.

O produto apoia programas de conscientização, GRC e redução de risco humano.
Toda campanha deve possuir autorização, escopo, período, público e domínios
aprovados. O EthPhish não contempla evasão de controles, coleta de credenciais
reais, payloads, exploração de vulnerabilidades ou uso não autorizado.

## Visão comercial

Para C-levels e executivos, o EthPhish converte o risco humano em evidência
operacional: mostra onde a organização é mais suscetível a engenharia social,
mede a evolução após treinamentos e dá governança sobre campanhas de
conscientização. Casos de uso incluem programas recorrentes de awareness,
validação de controles antes de auditorias, avaliação de risco por área,
simulações aprovadas após incidentes e prestação de serviços gerenciados para
múltiplos clientes.

As dores endereçadas são a falta de métricas consistentes de comportamento,
campanhas sem controle de escopo, relatórios manuais, baixa rastreabilidade de
aprovações e ausência de uma base evolutiva para segmentação multitenant. A
plataforma preserva as fronteiras éticas: mede eventos de interação autorizados
e não armazena senhas, OTPs ou outras credenciais reais.

## O que mudou neste fork

- identidade, versionamento e documentação próprios do EthPhish;
- PostgreSQL como banco de desenvolvimento e base de produção planejada;
- migrations PostgreSQL, pool configurável e lock de concorrência;
- configuração por variáveis de ambiente, sem editar `config.json` em runtime;
- Docker multi-stage, usuário não-root, capabilities removidas e health checks;
- TLS autoassinado para desenvolvimento, Caddy e painel administrativo sem
  publicação direta no host;
- CI com formatação, vet, testes, integração PostgreSQL, secret scan,
  vulnerabilidades, scan de imagem e SBOM;
- backup PostgreSQL, restore ensaiado e documentação de segurança, arquitetura
  e governança.

## Arquitetura

```text
                         rede administrativa privada
                                      │
                          VPN / Zero Trust / OIDC
                                      │
                           ┌──────────▼──────────┐
                           │ servidor central    │
                           │ administração/API   │
                           │ scheduler e jobs    │
                           └───────┬───────┬─────┘
                                   │       │
                         PostgreSQL│       │AMQP (futuro)
                                   │       ▼
                              ┌────▼───┐ ┌───────────────┐
                              │ dados  │ │ RabbitMQ       │
                              └────────┘ └───┬─────┬─────┘
                                               │     │
                                      ┌────────▼┐ ┌──▼─────────┐
                                      │ worker  │ │ worker     │
                                      │ node 1  │ │ node N     │
                                      └─────────┘ └────────────┘

 internet ── HTTPS 443 ── Caddy ── TLS interno ── web de campanhas/quishing
```

Na v0.2.0, scheduler, mailer e workers são componentes internos do processo do
servidor central; não existe uma porta de comunicação central-worker. A
separação em nodes é a arquitetura-alvo: workers sem acesso administrativo nem
ao PostgreSQL consumirão jobs assinados de RabbitMQ, em rede privada, por AMQP
TLS (`5671`). A porta AMQP não será publicada no host.

Consulte o detalhamento em [arquitetura alvo](docs/architecture/target-architecture.md).

## Containers e portas

| Componente | Papel | Portas | Exposição |
| --- | --- | --- | --- |
| `reverse-proxy` | TLS público e roteamento web | 9443/TCP local (443/TCP em produção) | publicada somente para conteúdo web autorizado |
| `server` | administração, API, campanhas e worker interno | 3333/TCP admin, 8080/TCP web | somente redes Docker internas |
| `postgres` | dados e migrations | 5432/TCP | rede de dados interna |
| `rabbitmq` | fundação para distribuição futura | 5672/5671/TCP | rede de dados interna |
| `tls-init` | prepara volume privado de certificados | nenhuma | execução única, sem rede |
| `worker-node` (futuro) | entrega aprovada e observabilidade | AMQP TLS 5671 de saída; SMTP/HTTPS conforme escopo | sem painel e sem banco direto |

## Recursos implementados na v0.2.0

- campanhas legadas por e-mail, SMS, QR Code e canal genérico, sempre sob uso
  autorizado;
- gestão de grupos, participantes, templates, landing pages e perfis de envio;
- eventos de abertura, clique, submissão simulada e reporte, conforme o
  comportamento herdado;
- OIDC existente, CSRF, autenticação administrativa e testes de caracterização;
- relatórios herdados Word/Excel, storage de relatórios e anonimização já
  existente no baseline;
- PostgreSQL, migrations idempotentes, readiness e health checks;
- TLS de desenvolvimento e configuração de TLS obrigatório para PostgreSQL
  quando `ETHPHISH_DB_REQUIRE_TLS=true`;
- backups lógicos e restore ensaiado em banco isolado;
- runtime e imagem PostgreSQL-only, migrations sem Goose/CGO e importação
  SQLite→PostgreSQL com pré-flight, transação e reconciliação por contagem;
- paridade PostgreSQL para campanhas, SMS, IMAP persistido, criptografia,
  webhooks e transições de relatórios, sem chamadas externas nos testes;
- CI/CD e controles de supply chain descritos em [release notes](RELEASE_NOTES.md).

## Próximas capacidades

- transactional outbox, filas AMQP, retries, DLQ e workers distribuídos;
- nodes de worker escaláveis horizontalmente, com identidade própria e sem
  acesso direto ao banco;
- multitenancy com RLS, portal de clientes e fluxo auditável de aprovação;
- dashboard operacional e executivo, métricas de capacidade e observabilidade;
- bundles versionados de campanhas, conteúdos de treinamento e importação/
  exportação controlada;
- backup automatizado, retenção, restauração recorrente e storage externo;
- alta disponibilidade, assinatura de imagens e política de atualização de
  dependências.

## Fora de escopo

- coleta, retenção ou recuperação de senhas, OTPs, cartões ou credenciais reais;
- evasão de filtros, antivírus, EDR, gateways ou mecanismos de proteção;
- payloads, exploração, comprometimento de contas ou acesso não autorizado;
- campanhas sem aprovação formal, domínios fora do escopo ou público não
  autorizado;
- exposição pública do painel administrativo.

## Requisitos

- Go 1.25.12 e compilador C (compatibilidade SQLite legada nos testes);
- Docker Engine com Compose v2;
- Node 22 para reconstrução de assets;
- PostgreSQL 17 para o ambiente Compose.

O runtime de produção exige PostgreSQL. SQLite permanece somente para testes e
migração legada explicitamente controlada; não inicie uma imagem de produção
com `ETHPHISH_DB_DRIVER=sqlite3`.

Referência de capacidade inicial: servidor central com 2 vCPU, 4 GB RAM e
100 GB de disco; cada worker futuro com 1 vCPU, 1 GB RAM e 50 GB de disco. A
capacidade real deve ser dimensionada por volume, taxa de entrega aprovada,
retenção de eventos e geração de relatórios.

## Desenvolvimento local

```sh
docker compose build
docker compose up -d
docker compose ps
```

Acesse `https://localhost:9443` para a superfície web de desenvolvimento. O
painel administrativo não é publicado no host. A credencial temporária é
gerada apenas no primeiro início e deve ser alterada imediatamente.

```sh
CGO_ENABLED=1 go test ./...
./scripts/backup-postgres.sh
```

Consulte [desenvolvimento local](docs/runbooks/local-development.md),
[release notes](RELEASE_NOTES.md), [issues conhecidos](ISSUES_CONHECIDOS.md) e
[política de segurança](SECURITY.md).

## Documentação

- [Arquitetura alvo](docs/architecture/target-architecture.md)
- [Inventário de dependências](docs/architecture/dependency-inventory.md)
- [Threat model](docs/security/threat-model.md)
- [Uso aceitável](docs/product/acceptable-use.md)
- [Status das Sprints 0–2](docs/project/sprint-0-2-status.md)
- [Sprint 02 — endurecimento PostgreSQL](docs/project/sprint-02.md)
- [Release notes v0.2.0](RELEASE_NOTES.md)
- [Issues conhecidos](ISSUES_CONHECIDOS.md)

## Licença e atribuição

O EthPhish mantém a atribuição ao Gophish e ao Anglerphish nos termos da
licença MIT herdada. As mudanças deste repositório formam um fork independente
e não implicam endosso dos projetos upstream.
