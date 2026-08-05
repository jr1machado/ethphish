# Baseline arquitetural

O baseline é o Anglerphish 1.3.0, fork do Gophish 0.12.1. O binário atual reúne
servidor administrativo, API, servidor público, mailer, SMS, IMAP e relatórios.
GORM v1 acessa SQLite ou MySQL, e filas de entrega são tabelas consultadas por
polling.

## Fronteiras atuais

| Pacote | Responsabilidade |
| --- | --- |
| `controllers` | UI administrativa e superfície pública |
| `controllers/api` | API autenticada por chave |
| `models` | domínio, validação e persistência |
| `worker` / `mailer` | agendamento e entrega de email/SMS |
| `imap` | reportes e respostas recebidas |
| `reports` | geração assíncrona via Python |
| `auth`, `middleware`, `crypto` | identidade e controles de segurança |

## Baseline de execução

Os modos `all`, `admin` e `phish` permitem separar superfícies, mas compartilham
o mesmo código e banco. A evolução preservará comportamento por testes antes de
introduzir PostgreSQL, tenancy e workers externos.
