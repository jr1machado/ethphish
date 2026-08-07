# Release notes — EthPhish v0.6.0

Data: 2026-08-07

## Resumo

Redesenho unificado das três telas públicas de login/autenticação do
EthPhish (admin, portal de aprovação de contrato e portal do cliente), que
até aqui usavam três estéticas diferentes: Bootstrap genérico herdado do
Gophish/Anglerphish no admin, e CSS inline duplicado (copiado e colado) nas
duas telas do cliente. Release **puramente de apresentação** — nenhuma rota,
sessão, token, CSRF ou regra de autenticação/autorização foi alterada.

Substitui as notas da v0.5.0 (tag `v0.5.0`, commit `899e5c8`) como release
corrente; as notas anteriores permanecem disponíveis no histórico do
repositório e na tag correspondente.

## Entregas

### Design system de autenticação único

- Novo arquivo `static/css/ethphish-auth.css` (286 linhas): variáveis de
  tema (`--auth-bg`, `--auth-panel`, `--auth-card`, `--auth-accent`,
  `--auth-danger` etc.), tipografia própria (Space Grotesk para títulos,
  JetBrains Mono para rótulos/elementos técnicos, Source Sans Pro para
  corpo de texto), grid responsivo de duas colunas (painel de marca +
  cartão de formulário) que colapsa para uma coluna abaixo de 860px de
  largura.
- Estética "console de operações": fundo escuro, grade sutil, linha de
  scan animada (`auth-scanline`) no painel de marca — reforça a leitura de
  ferramenta de segurança, não SaaS genérico.
- Cada uma das três telas ganhou headline e texto de contexto próprios no
  painel de marca, em vez de um card genérico repetido:
  - Admin (`templates/login.html`): "Authorized access to the operations
    console."
  - Aprovador de contrato, tela de erro de link (`templates/client_login.html`):
    "Client sign-off, tied to an exact scope."
  - Cliente, login self-service (`templates/portal_login.html`): "Your
    program, measured continuously."

### Limpeza de débito técnico associado

- Removido o bloco `<style>` inline duplicado que existia em
  `client_login.html` e `portal_login.html` (a mesma folha de estilo,
  copiada em cada template) — agora ambos referenciam a folha
  compartilhada.
- Removida a dependência da tela de login admin no `navbar`/`form-signin`
  do Bootstrap herdado do Gophish/Anglerphish; o formulário de SSO/OIDC
  (`{{if .OIDCEnabled}}`) foi preservado, só a marcação ao redor mudou.

### Sem mudança de comportamento

- `csrf_token`/`gorilla.csrf.Token`, ação dos formulários (`POST`), nomes
  de campo (`username`, `password`, `email`), sessão `ethphish_client` e o
  fluxo `/auth/oidc/login` continuam idênticos — só o HTML/CSS ao redor
  mudou. Não há migration, endpoint novo ou dependência de backend nesta
  release.

## Integrações

| Integração | Estado v0.6.0 | Uso |
| --- | --- | --- |
| Google Fonts (Space Grotesk, JetBrains Mono, Source Sans Pro) | já usado no login admin (Source Sans Pro); agora ampliado e consolidado num único `@import url(...)` em `ethphish-auth.css` | tipografia das três telas de login; requer as instâncias de admin/web terem saída de rede para `fonts.googleapis.com`/`fonts.gstatic.com` — ver [Issues conhecidos](ISSUES_CONHECIDOS.md) |
| PostgreSQL, RabbitMQ, SMTP, Caddy | inalterados desde a v0.5.0 | ver notas da release anterior na tag `v0.5.0` do histórico Git |

## Upgrade e rollback

1. Sem migration de banco nesta release — nenhum passo de `db-migrate`
   novo.
2. `docker compose build && docker compose up -d` reconstrói a imagem com
   `static/css/ethphish-auth.css` e os três templates atualizados; assets
   estáticos servidos pelo `server`/Caddy, sem cache externo a invalidar
   além do cache de navegador do usuário final.
3. Rollback é trivial: reverter para a imagem/tag anterior restaura os
   templates antigos, sem efeito colateral em dado persistido (nenhuma
   tabela nova, nenhuma coluna nova).

## Limitações de release

Consulte [ISSUES_CONHECIDOS.md](ISSUES_CONHECIDOS.md). O redesenho não
alcança `templates/reset_password.html` (ainda no Bootstrap herdado) nem
qualquer tela autenticada do painel admin — escopo desta release foi
deliberadamente as telas públicas de entrada visitadas por terceiros
(cliente/aprovador) e a tela de login do admin.
