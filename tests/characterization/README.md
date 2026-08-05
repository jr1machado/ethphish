# Testes de caracterização

Estes testes fixam o comportamento herdado que deve sobreviver às mudanças de
infraestrutura das Sprints 2 e 3. Eles vivem próximos aos pacotes que exercitam,
para compartilhar os fixtures e evitar uma segunda suíte de integração frágil.

| Fluxo | Testes de referência |
| --- | --- |
| Login administrativo, CSRF e OIDC | `controllers/route_test.go` |
| Criação de grupo, template, página e SMTP | `models/group_test.go`, `models/template_test.go`, `models/page_test.go`, `models/smtp_test.go` |
| Criação, agendamento e resultados de campanhas | `models/campaign_test.go`, `models/result_test.go`, `worker/worker_test.go` |
| Abertura, clique, reporte e submissão | `controllers/phish_test.go` |
| Criptografia de campos | `models/campaign_test.go` (leitura de dados cifrados) |
| API administrativa e isolamento por usuário legado | `controllers/api/*_test.go` |
| Relatórios e jobs | `reports/` e `controllers/api/reports.go` |

## Execução

```sh
docker build --target test -t ethphish-test:local .
```

Todo comportamento novo ou corrigido precisa ser coberto no pacote responsável e
incluído nesta matriz quando for relevante à migração para PostgreSQL.
