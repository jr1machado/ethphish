# ADR-001: baseline tecnológico

- Status: aceito
- Data: 2026-08-04

## Decisão

Anglerphish 1.3.0 é o upstream funcional congelado. O EthPhish mantém identidade,
versionamento e ciclo de releases próprios. Alterações do upstream serão trazidas
de forma explícita, revisadas e testadas.

## Consequências

O histórico herdado permanece rastreável e atualizações não entram
automaticamente. Correções próprias não devem ser feitas diretamente no branch
que representa o upstream.
