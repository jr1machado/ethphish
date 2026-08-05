# Sprint 02 — endurecimento do runtime PostgreSQL

## Contexto

As Sprints 0–2 do roadmap original estabeleceram a fundação PostgreSQL. Esta
iteração de desenvolvimento usa a branch `feature/sprint02` para fechar as
pendências operacionais antes da paridade funcional e da migração de dados.

## Objetivo

Garantir que uma imagem EthPhish de produção não possa iniciar com SQLite.
SQLite continua suportado somente pela suíte legada e por operações de migração
executadas em ambiente isolado e autorizado.

## Entregas desta iteração

- `ETHPHISH_RUNTIME_ENV=production` definido pela imagem runtime e pelo
  Compose;
- recusa explícita de `ETHPHISH_DB_DRIVER=sqlite3` em runtime de produção;
- teste de regressão para essa recusa;
- documentação operacional da separação entre PostgreSQL de produção e SQLite
  legado.

## Próximos itens

- ferramenta controlada para migração SQLite → PostgreSQL, com preflight,
  relatório de reconciliação e origem somente leitura;
- testes de paridade dos fluxos críticos em PostgreSQL;
- remover o driver SQLite da imagem de produção após a ferramenta de migração
  e a paridade serem aprovadas;
- ensaio de restore PostgreSQL com evidência anexada ao processo de release.

## Critérios de aceite deste incremento

- a imagem runtime falha de forma clara ao receber SQLite em produção;
- o Compose de desenvolvimento continua iniciando com PostgreSQL;
- testes de configuração e fluxos administrativos permanecem aprovados;
- nenhuma configuração de produção precisa de SQLite.
