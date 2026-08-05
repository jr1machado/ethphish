# Inventário inicial de dependências críticas

Data de baseline: 2026-08-04. Fonte: `go.mod`, `package.json` e Dockerfile
herdado.

| Área | Dependências/tecnologia | Situação e tratamento |
| --- | --- | --- |
| Persistência | GORM 1.9, Goose legado, SQLite e MySQL | PostgreSQL será introduzido com testes de caracterização antes de uma atualização de ORM. |
| HTTP | Gorilla Mux, sessions e CSRF | Manter durante a fundação; atualizar somente em alteração isolada e testada. |
| Autenticação | bcrypt, OAuth2, CoreOS OIDC | OIDC existente será mantido; MFA administrativo entra em sprint posterior. |
| Entrega | gomail, Twilio SDK | Providers continuam no control plane até a extração de workers. |
| Frontend | jQuery, Bootstrap, Gulp 4, Webpack 4, CKEditor | Build foi fixado em Node 22; modernização de interface não é pré-requisito para PostgreSQL. |
| Relatórios | Python e dependências de DOCX/XLSX | Continuam como subprocesso; storage externo será tratado antes de produção. |
| Container | Debian slim, Node e Go multi-stage | Imagens possuem tags explícitas; digest e assinatura são gates de hardening posteriores. |

## Riscos conhecidos

- GORM v1 e Goose antigo podem gerar SQL incompatível com PostgreSQL.
- O frontend usa ferramentas em fim de vida e deve permanecer congelado enquanto
  a camada de dados é estabilizada.
- O driver SQLite exige CGO e explica o tempo de compilação do baseline.
- O runtime de relatórios requer Python, aumentando a superfície da imagem.
- O módulo do Go continua com o caminho histórico `github.com/gophish/gophish`;
  a mudança desse identificador será uma decisão separada para evitar quebra de
  imports e integrações durante as primeiras sprints.

## Decisão de curto prazo

Não haverá atualização massiva de dependências na fundação. Cada atualização
deverá ter justificativa, teste de regressão e rollback independente.
