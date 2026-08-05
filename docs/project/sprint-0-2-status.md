# Status de entrega — Sprints 0 a 2

Atualizado em 2026-08-05.

## Sprint 0 — concluída no repositório

- Baseline Anglerphish 1.3.0, inventário de dependências, arquitetura e ADRs
  registrados em `docs/architecture/` e `docs/adrs/`.
- Threat model, política de uso, Definition of Done, política de segurança,
  CODEOWNERS e estratégia de ambientes registrados.
- Matriz de testes de caracterização registrada em
  `tests/characterization/README.md`; a suíte cobre os fluxos herdados sem
  executar entregas externas.

## Sprint 1 — concluída no repositório

- Imagem multi-stage, usuário não-root, capabilities removidas, health check,
  Compose isolado e painel administrativo sem porta publicada.
- Configuração por ambiente, TLS de desenvolvimento autoassinado, Caddy e CI
  com formatação, vet, testes, scan de segredos, vulnerabilidades, imagem,
  Trivy e SBOM.

## Sprint 2 — concluída no repositório

- Driver PostgreSQL, pool configurável, migrations específicas por dialeto,
  health/readiness e teste de integração com PostgreSQL efêmero no CI.
- Advisory lock PostgreSQL para serializar migrations concorrentes.
- Backup lógico local e procedimento de restauração documentados.
- `ETHPHISH_DB_REQUIRE_TLS=true` impede DSN PostgreSQL com
  `sslmode=disable` em ambientes que exigem transporte protegido.

## Ações administrativas externas pendentes

- Configurar no GitHub a proteção de `main`/`develop`: pull request, revisão
  obrigatória de CODEOWNERS e checks de CI bloqueantes.
- Confirmar que o repositório remoto corporativo e a equipe definida em
  `CODEOWNERS` existem e possuem as permissões descritas.
- Ensaiar restauração do dump em infraestrutura aprovada e isolada. O backup
  local é criado e protegido, mas a restauração não é executada contra bases
  ativas.
