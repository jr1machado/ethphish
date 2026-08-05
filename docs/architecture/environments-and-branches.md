# Ambientes e branches

## Branches

- `upstream/*`: espelho controlado do Anglerphish;
- `main`: produto estável;
- `develop`: integração;
- `feature/*`, `fix/*`, `docs/*`: trabalho revisável.

Pull requests exigem CI aprovado e revisão dos owners aplicáveis. Proteções no
provedor Git devem ser configuradas administrativamente e não são representadas
somente por arquivos deste repositório.

## Ambientes

- Desenvolvimento: Compose, dados sintéticos e nenhum envio externo.
- Integração: efêmero por PR, banco isolado e providers falsos.
- Homologação: topologia próxima da produção, dois tenants e dois workers falsos.
- Produção: bloqueada até isolamento, backup/restore, OIDC, aprovação, hardening,
  pentest e piloto controlado serem aceitos.
