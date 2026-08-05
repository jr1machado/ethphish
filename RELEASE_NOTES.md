# Release notes — v0.1.0

Data: 2026-08-05

## Resumo

Primeira release de fundação do EthPhish, fork independente para testes éticos
de phishing e quishing. Esta versão estabiliza a plataforma para desenvolvimento
controlado; não é uma autorização para campanhas externas ou produção pública.

## Entregas

### Produto e governança

- identidade EthPhish, `VERSION=0.1.0`, uso aceitável, segurança, ADRs e
  Definition of Done;
- baseline Anglerphish 1.3.0 rastreado e matriz de caracterização;
- documentação comercial e de arquitetura, incluindo limites éticos.

### Plataforma

- Docker multi-stage com Node 22.18.0, Go 1.25.12 e runtime Debian slim;
- execução como UID/GID 10001, sem capabilities e sem publicação da porta
  administrativa;
- Compose com Caddy, PostgreSQL 17, RabbitMQ, volumes privados e health checks;
- certificados autoassinados de desenvolvimento para administração e web.

### Dados e confiabilidade

- driver e migrations PostgreSQL;
- pool configurável por ambiente;
- advisory lock para impedir migrations concorrentes;
- health (`/healthz`) e readiness (`/readyz`) sem dados de conexão;
- backup lógico e ensaio de restore em banco PostgreSQL isolado;
- opção de exigir TLS de banco com `ETHPHISH_DB_REQUIRE_TLS=true`.

### Qualidade e supply chain

- CI para formatação, vet, testes, PostgreSQL efêmero, Gitleaks,
  govulncheck, Trivy e SBOM;
- testes de configuração, TLS, health checks e integração PostgreSQL.
- correção das CVEs de alta e crítica severidade encontradas na imagem inicial:
  Go 1.25.12, `golang.org/x/{crypto,net,oauth2,text}`, `go-jose` e `logrus`
  foram atualizados para versões corrigidas.
- `gorilla/csrf` foi atualizado para 1.7.3 e o SDK Twilio para 1.30.9,
  removendo a cadeia legada de `github.com/golang-jwt/jwt` sem correção.

## Integrações

| Integração | Estado v0.1.0 | Uso |
| --- | --- | --- |
| PostgreSQL | ativo no Compose | dados, migrations e readiness |
| Caddy | ativo no Compose | TLS e proxy web de desenvolvimento |
| RabbitMQ | provisionado | base para workers futuros; não há jobs distribuídos ainda |
| OIDC | recurso herdado configurável | autenticação administrativa, mediante IdP aprovado |
| SMTP, SMS, IMAP | recursos herdados | apenas em escopo autorizado; não configurados automaticamente |
| GitHub Actions | workflow incluído | exige publicação do workflow e token com escopo `workflow` |

## Upgrade e rollback

1. Faça backup com `./scripts/backup-postgres.sh`.
2. Atualize a imagem e execute `docker compose up -d`.
3. Consulte `docker compose logs server` para migrations e health.
4. Em falha, interrompa a atualização e restaure somente em banco isolado a
   partir do dump validado; não execute migrations `down` diretamente em
   produção.

## Limitações de release

Consulte [ISSUES_CONHECIDOS.md](ISSUES_CONHECIDOS.md). Workers distribuídos,
multitenancy e operação externa permanecem fora desta release.
