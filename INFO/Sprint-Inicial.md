# EthPhish — Planejamento Inicial de Sprints

**Data-base:** 4 de agosto de 2026
**Duração:** 2 semanas por sprint
**Horizonte inicial:** 10 semanas
**Base:** Anglerphish 1.3.0
**Objetivo do ciclo:** produzir uma fundação segura, reproduzível, baseada em PostgreSQL e preparada para multitenancy.

---

# 1. Premissas de planejamento

## 1.0 Gate inicial de ambiente

Antes do início da Sprint 0, o ambiente de desenvolvimento deve ter Go 1.25.12,
Docker Compose v2 e Node 22 instalados e validados. A primeira evidência de
execução é `go version` e `go test ./...`; dependências ausentes ou cache sem
permissão de escrita devem ser corrigidos antes de aceitar qualquer item da
sprint.

## 1.1 Capacidade de referência

Planejamento inicial para um time com funções equivalentes a:

* 1 liderança técnica ou arquiteto;
* 1 desenvolvedor backend Go;
* 1 desenvolvedor full stack;
* 1 profissional DevOps ou plataforma compartilhado;
* QA e segurança compartilhados.

Capacidade sugerida:

```text
Capacidade bruta:       35 a 40 pontos por sprint
Capacidade planejada:   28 a 32 pontos
Reserva para riscos:    20%
```

Caso o time tenha apenas duas pessoas, cada sprint deverá ter seu escopo reduzido ou sua duração ampliada.

## 1.2 Escala de estimativa

```text
1 ponto  — alteração pequena e conhecida
2 pontos — tarefa simples
3 pontos — pequena complexidade ou integração
5 pontos — complexidade média
8 pontos — complexidade alta ou grande incerteza
13 pontos — tarefa que deve ser dividida
```

Itens estimados em 13 pontos não deverão entrar em desenvolvimento sem decomposição.

## 1.3 Regras gerais

Toda história deverá incluir:

* critérios de aceite;
* testes;
* autorização;
* logs;
* métricas quando aplicável;
* documentação;
* análise multitenant;
* análise de segurança;
* rollback ou procedimento de reversão.

Nenhuma sprint deverá colocar a plataforma em produção externa.

---

# 2. Visão das primeiras sprints

| Sprint   | Tema                  | Resultado esperado                                   |
| -------- | --------------------- | ---------------------------------------------------- |
| Sprint 0 | Descoberta e baseline | Projeto formalizado, riscos e arquitetura conhecidos |
| Sprint 1 | Fork e CI/CD          | Build seguro, reproduzível e conteinerizado          |
| Sprint 2 | PostgreSQL — fundação | Conexão, schema e migrations básicas funcionando     |
| Sprint 3 | PostgreSQL — paridade | Funcionalidades principais operando sem SQLite       |
| Sprint 4 | Fundação multitenant  | Tenants segregados no banco e na aplicação           |

---

# 3. Sprint 0 — Descoberta, governança e baseline

## Objetivo

Formalizar o projeto EthPhish, congelar o baseline do Anglerphish e reduzir incertezas antes das alterações estruturais.

## Meta da sprint

> Ao final da sprint, o time deverá conhecer o comportamento atual da plataforma, os riscos de segurança, as dependências críticas e as decisões arquiteturais necessárias para começar a implementação.

## Backlog

| ID      | História ou atividade                                | Pontos | Responsável principal |
| ------- | ---------------------------------------------------- | -----: | --------------------- |
| ETH-001 | Criar repositório corporativo do EthPhish            |      2 | DevOps                |
| ETH-002 | Registrar Anglerphish 1.3.0 como upstream controlado |      3 | Tech Lead             |
| ETH-003 | Inventariar funcionalidades existentes               |      5 | Produto/Backend       |
| ETH-004 | Mapear estrutura do código e principais módulos      |      5 | Backend               |
| ETH-005 | Criar testes de caracterização das funções críticas  |      8 | Backend/QA            |
| ETH-006 | Elaborar threat model inicial                        |      5 | Segurança/Arquitetura |
| ETH-007 | Criar política de uso aceitável                      |      3 | Produto/Segurança     |
| ETH-008 | Definir ambientes e estratégia de branches           |      3 | DevOps                |
| ETH-009 | Criar ADRs arquiteturais iniciais                    |      5 | Arquitetura           |
| ETH-010 | Definir Definition of Done e processo de revisão     |      2 | Tech Lead             |

**Estimativa total:** 41 pontos.

A sprint deverá priorizar aproximadamente 30 pontos. O inventário funcional e os ADRs podem ser concluídos paralelamente, enquanto testes de caracterização devem se concentrar primeiro nos fluxos críticos.

## ADRs obrigatórios

### ADR-001 — Base tecnológica

```text
Decisão:
Usar Anglerphish 1.3.0 como upstream funcional congelado.

Estratégia:
Manter upstream separado do branch principal do EthPhish.
```

### ADR-002 — Banco de dados

```text
Decisão:
PostgreSQL será o banco primário de produção.

SQLite será permitido apenas como origem temporária de migração
ou em testes legados durante a transição.
```

### ADR-003 — Modelo arquitetural

```text
Decisão:
Começar como monólito modular com workers distribuídos.

Motivo:
Reduzir complexidade operacional sem impedir extrações futuras.
```

### ADR-004 — Administração

```text
Decisão:
O painel administrativo nunca será publicado diretamente na internet.

Acesso:
VPN, Zero Trust, bastion ou rede administrativa privada.
```

### ADR-005 — Dados sensíveis

```text
Decisão:
Não armazenar senhas, OTPs ou credenciais reais de participantes.

Persistir somente o evento de tentativa de preenchimento.
```

## Testes de caracterização prioritários

* criação de usuário e grupo;
* criação de template;
* criação de landing page;
* criação e início de campanha;
* registro de abertura;
* registro de clique;
* submissão simulada;
* geração de relatório;
* login administrativo;
* criptografia e leitura de campos protegidos;
* seleção e uso de configuração SMTP.

## Critérios de aceite

* repositório criado e protegido;
* versão upstream registrada;
* funcionalidades principais inventariadas;
* riscos críticos documentados;
* arquitetura inicial aprovada;
* ambientes definidos;
* Definition of Done publicada;
* fluxo crítico coberto por testes básicos;
* nenhuma funcionalidade de evasão incluída no escopo.

## Riscos

| Risco                           | Mitigação                                         |
| ------------------------------- | ------------------------------------------------- |
| Ausência de testes no legado    | Criar testes de caracterização antes de refatorar |
| Dependências antigas            | Inventariar e classificar por criticidade         |
| Comportamentos não documentados | Registrar evidências durante execução local       |
| Escopo excessivo                | Manter Sprint 0 focada em redução de incerteza    |

## Entregáveis

```text
/docs/architecture/
/docs/security/
/docs/product/
/docs/adrs/
/docs/runbooks/
/tests/characterization/
CONTRIBUTING.md
SECURITY.md
CODEOWNERS
```

---

# 4. Sprint 1 — Fundação do fork e CI/CD

## Objetivo

Criar uma base de desenvolvimento e entrega segura, reproduzível e auditável.

## Meta da sprint

> Qualquer commit deverá ser compilado, testado e analisado automaticamente, produzindo uma imagem conteinerizada que não exponha o painel administrativo.

## Backlog

| ID      | História ou atividade                                 | Pontos | Responsável principal |
| ------- | ----------------------------------------------------- | -----: | --------------------- |
| ETH-101 | Renomear produto e definir versionamento EthPhish     |      3 | Backend/Produto       |
| ETH-102 | Criar Dockerfile multi-stage do servidor              |      5 | DevOps                |
| ETH-103 | Executar container como usuário não root              |      3 | DevOps                |
| ETH-104 | Remover alteração automática de listen address        |      3 | Backend               |
| ETH-105 | Externalizar configurações por ambiente               |      5 | Backend               |
| ETH-106 | Criar Docker Compose de desenvolvimento               |      5 | DevOps                |
| ETH-107 | Implementar pipeline de build e testes                |      5 | DevOps                |
| ETH-108 | Adicionar `gofmt`, `go vet` e análise estática        |      3 | Backend               |
| ETH-109 | Adicionar Gitleaks e `govulncheck`                    |      3 | Segurança             |
| ETH-110 | Adicionar scan da imagem com Trivy                    |      3 | DevOps                |
| ETH-111 | Gerar SBOM inicial                                    |      3 | DevOps                |
| ETH-112 | Configurar proteção de branches e revisão obrigatória |      2 | Tech Lead             |

**Estimativa total:** 43 pontos.
**Escopo recomendado:** 30 a 32 pontos.

Os itens ETH-111 e parte das análises mais avançadas podem ser transferidos para a sprint seguinte caso a atualização do build apresente incompatibilidades.

## Estrutura inicial de containers

```text
ethphish-server
postgresql
rabbitmq
reverse-proxy
```

Nesta sprint, PostgreSQL e RabbitMQ poderão existir no Compose, mesmo que ainda não estejam integrados completamente ao código.

## Requisitos do container do servidor

* usuário não root;
* imagem mínima;
* filesystem somente leitura quando possível;
* diretórios temporários em `tmpfs`;
* capabilities removidas;
* health check;
* tag imutável;
* sem compilador no runtime;
* sem código-fonte no runtime;
* sem publicação direta da porta administrativa.

## Publicação de portas

```text
Permitido externamente:
443/TCP — landing pages e conteúdo público futuro

Somente rede administrativa:
porta interna do painel administrativo

Nunca publicado:
5432/TCP — PostgreSQL
5672/5671 — broker, salvo rede privada autorizada
métricas internas
debug
profiler
```

## Pipeline mínimo

```text
Checkout
   ↓
Validação de formatação
   ↓
Análise estática
   ↓
Testes unitários
   ↓
Testes de caracterização
   ↓
Dependency scan
   ↓
Secret scan
   ↓
Build dos binários
   ↓
Build da imagem
   ↓
Scan da imagem
   ↓
SBOM
   ↓
Publicação no registry de desenvolvimento
```

## Critérios de aceite

* build reproduzível;
* imagem executada sem root;
* painel administrativo não publicado externamente;
* configuração sensível não armazenada no repositório;
* pipeline executado em pull requests;
* falhas críticas bloqueiam merge;
* imagem identificada pela versão e commit;
* health check funcional;
* documentação para execução local disponível.

## Métricas da sprint

* tempo médio do pipeline;
* percentual de builds bem-sucedidos;
* quantidade de vulnerabilidades identificadas;
* cobertura dos testes existentes;
* tamanho da imagem;
* tempo de inicialização do container.

## Riscos

| Risco                                      | Mitigação                                     |
| ------------------------------------------ | --------------------------------------------- |
| Build frontend legado quebrar              | Fixar versões e criar cache controlado        |
| Dependências sem suporte                   | Atualização incremental, não massiva          |
| Imagem crescer excessivamente              | Multi-stage e remoção de ferramentas de build |
| Painel ficar acessível por erro de Compose | Teste automatizado de portas publicadas       |

---

# 5. Sprint 2 — PostgreSQL: conexão, schema e migrations

## Objetivo

Criar a fundação PostgreSQL sem remover imediatamente o suporte legado necessário para comparação.

## Meta da sprint

> A aplicação deverá iniciar conectada ao PostgreSQL e executar migrations controladas em um ambiente de desenvolvimento.

## Backlog

| ID      | História ou atividade                                | Pontos | Responsável principal |
| ------- | ---------------------------------------------------- | -----: | --------------------- |
| ETH-201 | Mapear schema SQLite atual                           |      5 | Backend/DB            |
| ETH-202 | Identificar SQL e migrations específicos por dialeto |      5 | Backend               |
| ETH-203 | Implementar configuração PostgreSQL                  |      5 | Backend               |
| ETH-204 | Implementar conexão e pool                           |      5 | Backend               |
| ETH-205 | Criar migration inicial PostgreSQL                   |      8 | Backend/DB            |
| ETH-206 | Criar mecanismo seguro de execução de migrations     |      5 | Backend               |
| ETH-207 | Adicionar health check do banco                      |      3 | Backend               |
| ETH-208 | Criar testes de integração PostgreSQL                |      8 | Backend/QA            |
| ETH-209 | Documentar rollback e recuperação                    |      3 | DevOps                |
| ETH-210 | Criar backup básico no ambiente de desenvolvimento   |      3 | DevOps                |

**Estimativa total:** 50 pontos.

Esta sprint possui alta incerteza. O compromisso recomendado é concluir a conexão, migrations essenciais e os primeiros testes de integração. A migração completa das funcionalidades fica para a Sprint 3.

## Configurações mínimas

```text
ETHPHISH_DB_DRIVER=postgres
ETHPHISH_DB_DSN="host=postgres port=5432 user=ethphish password=<secret> dbname=ethphish sslmode=disable"
ETHPHISH_DB_MAX_OPEN_CONNECTIONS=10
ETHPHISH_DB_MAX_IDLE_CONNECTIONS=5
ETHPHISH_DB_CONNECTION_MAX_LIFETIME=30m
```

Em produção, `sslmode=disable` não será permitido quando o banco estiver em outro host ou rede não confiável.

## Regras para migrations

* ordem determinística;
* execução idempotente;
* identificação de versão;
* checksum das migrations;
* bloqueio contra execução concorrente;
* logs sem credenciais;
* rollback documentado;
* backup antes de migrations destrutivas;
* migrations de schema separadas de cargas extensas de dados.

## Tabelas prioritárias

* usuários administrativos;
* configurações;
* grupos;
* participantes;
* templates de e-mail;
* landing pages;
* configurações de envio;
* campanhas;
* resultados;
* eventos.

## Critérios de aceite

* aplicação inicia usando PostgreSQL;
* migrations funcionam em banco vazio;
* execução repetida não corrompe o schema;
* pool de conexão é configurável;
* indisponibilidade do banco é reportada no health check;
* testes são executados em banco efêmero no CI;
* logs não exibem DSN completo ou senha;
* rollback está documentado.

## Testes obrigatórios

```text
Banco vazio
Banco parcialmente migrado
Migration executada duas vezes
Banco indisponível
Senha inválida
Conexão encerrada durante operação
Limite de conexões
Rollback de versão
Concorrência na execução de migrations
```

## Riscos

| Risco                              | Mitigação                                   |
| ---------------------------------- | ------------------------------------------- |
| GORM legado gerar SQL incompatível | Testes por repositório e queries explícitas |
| IDs e sequences divergentes        | Definir estratégia única de identidade      |
| Tipos booleanos e timestamps       | Normalização no schema PostgreSQL           |
| Migration muito extensa            | Separar schema de migração de dados         |

---

# 6. Sprint 3 — PostgreSQL: paridade funcional e migração

## Objetivo

Concluir a transição das funcionalidades críticas para PostgreSQL e remover a dependência de SQLite no runtime de produção.

## Meta da sprint

> Os principais fluxos do Anglerphish deverão funcionar no PostgreSQL com resultados equivalentes aos fluxos legados.

## Backlog

| ID      | História ou atividade                        | Pontos | Responsável principal |
| ------- | -------------------------------------------- | -----: | --------------------- |
| ETH-301 | Adaptar repositórios e queries incompatíveis |      8 | Backend               |
| ETH-302 | Validar campanhas e resultados no PostgreSQL |      5 | Backend/QA            |
| ETH-303 | Validar templates e landing pages            |      3 | Backend/QA            |
| ETH-304 | Validar configurações SMTP, SMS e IMAP       |      5 | Backend/QA            |
| ETH-305 | Validar criptografia de campos sensíveis     |      5 | Segurança/Backend     |
| ETH-306 | Criar ferramenta SQLite para PostgreSQL      |      8 | Backend               |
| ETH-307 | Implementar reconciliação de registros       |      5 | Backend/QA            |
| ETH-308 | Criar backup com `pgBackRest` ou equivalente |      5 | DevOps                |
| ETH-309 | Executar teste de restauração                |      3 | DevOps/QA             |
| ETH-310 | Remover SQLite da imagem de produção         |      3 | DevOps                |

**Estimativa total:** 50 pontos.

A ferramenta de migração poderá ser entregue em versão inicial, limitada às entidades já homologadas.

## Fluxo de migração

```text
Bloquear alterações no ambiente de origem
        ↓
Criar backup do SQLite
        ↓
Executar validações de integridade
        ↓
Extrair dados em lotes
        ↓
Normalizar tipos e referências
        ↓
Importar no PostgreSQL
        ↓
Reconciliar quantidades e hashes
        ↓
Executar testes funcionais
        ↓
Gerar relatório de migração
```

## Validações mínimas

* quantidade de usuários;
* grupos e participantes;
* templates;
* landing pages;
* campanhas;
* resultados;
* eventos;
* configurações de envio;
* relacionamentos;
* registros órfãos;
* timestamps;
* conteúdo criptografado.

## Critérios de aceite

* fluxos críticos operam no PostgreSQL;
* SQLite não existe na imagem de produção;
* migração gera relatório;
* diferenças de contagem são apresentadas;
* backup e restauração são testados;
* dados criptografados continuam legíveis somente com a chave correta;
* falhas de migração não alteram a origem;
* procedimento de rollback está documentado.

## Gate para avançar

A Sprint 4 não deverá alterar as entidades principais para multitenancy até que:

```text
Migrations PostgreSQL estejam estáveis
Testes críticos estejam passando
Backup tenha sido restaurado com sucesso
Inconsistências de dados estejam resolvidas
```

---

# 7. Sprint 4 — Fundação multitenant

## Objetivo

Introduzir tenants, empresas e isolamento de dados como requisito estrutural do EthPhish.

## Meta da sprint

> Dois tenants deverão utilizar a mesma instalação sem conseguir consultar, alterar ou inferir dados um do outro.

## Backlog

| ID      | História ou atividade                           | Pontos | Responsável principal |
| ------- | ----------------------------------------------- | -----: | --------------------- |
| ETH-401 | Criar entidade `tenant`                         |      3 | Backend               |
| ETH-402 | Criar entidade `company`                        |      3 | Backend               |
| ETH-403 | Vincular usuários administrativos a tenants     |      5 | Backend               |
| ETH-404 | Adicionar `tenant_id` às entidades prioritárias |      8 | Backend/DB            |
| ETH-405 | Implementar contexto de tenant por requisição   |      5 | Backend               |
| ETH-406 | Implementar PostgreSQL Row-Level Security       |      8 | Backend/DB            |
| ETH-407 | Criar papéis iniciais do tenant                 |      5 | Backend               |
| ETH-408 | Criar política de empresa por usuário           |      5 | Backend               |
| ETH-409 | Particionar cache e storage por tenant          |      5 | Full Stack/Backend    |
| ETH-410 | Criar suíte de testes de isolamento             |      8 | QA/Segurança          |
| ETH-411 | Criar trilha inicial de auditoria               |      5 | Backend               |

**Estimativa total:** 60 pontos.

Essa sprint deverá priorizar isolamento do domínio antes de funcionalidades de interface avançadas. Dependendo do tamanho do time, poderá ser dividida em:

```text
Sprint 4A — Modelo e contexto de tenant
Sprint 4B — RLS, autorização e testes de isolamento
```

## Estratégia de isolamento

```text
Banco compartilhado
Tabelas compartilhadas
tenant_id obrigatório
PostgreSQL Row-Level Security
Middleware de tenant
Autorização na camada de serviço
Testes de acesso cruzado
Storage particionado
Cache particionado
```

## Entidades prioritárias com `tenant_id`

* usuários;
* empresas;
* participantes;
* departamentos;
* grupos;
* templates;
* landing pages;
* campanhas;
* treinamentos;
* configurações de envio;
* domínios;
* relatórios;
* eventos;
* aprovações;
* workers autorizados.

## Regra de consulta

Uma consulta de negócio deverá combinar:

```text
Identidade autenticada
Tenant ativo
Empresa autorizada
Permissão necessária
Recurso solicitado
```

O identificador do recurso, isoladamente, nunca será suficiente.

## Testes obrigatórios

* usuário do tenant A consulta campanha do tenant B;
* usuário altera identificador na URL;
* token de A é apresentado em B;
* cache retorna recurso de outro tenant;
* exportação inclui dados externos ao tenant;
* relatório agrega empresas não autorizadas;
* job de worker referencia tenant não permitido;
* consulta sem contexto de tenant;
* superadministrador acessa conteúdo sem elevação explícita;
* usuário perde acesso durante uma sessão ativa.

## Critérios de aceite

* tenant A não acessa dados do tenant B;
* consultas sem tenant são negadas;
* RLS permanece ativa mesmo em erro da aplicação;
* permissões são verificadas no backend;
* cache não mistura tenants;
* storage utiliza prefixos ou buckets segregados;
* testes de isolamento fazem parte do CI;
* ações administrativas são auditadas;
* tentativas de acesso cruzado geram evento de segurança.

---

# 8. Dependências entre as primeiras sprints

```text
Sprint 0
   │
   ├── baseline e testes
   ▼
Sprint 1
   │
   ├── build e container seguros
   ▼
Sprint 2
   │
   ├── conexão e schema PostgreSQL
   ▼
Sprint 3
   │
   ├── paridade e migração
   ▼
Sprint 4
       └── multitenancy e isolamento
```

Não é recomendado executar PostgreSQL e multitenancy simultaneamente sobre as mesmas entidades sem primeiro estabilizar o schema.

---

# 9. Marcos de decisão

## Marco 1 — Final da Sprint 0

Decidir:

* quais funcionalidades legadas serão preservadas;
* quais módulos serão descontinuados;
* quais dependências exigem atualização imediata;
* se o frontend atual será mantido temporariamente;
* qual ferramenta será usada para secrets management.

## Marco 2 — Final da Sprint 1

Validar:

* reprodutibilidade da imagem;
* segurança do runtime;
* pipeline;
* estratégia de configuração;
* registry;
* política de releases.

## Marco 3 — Final da Sprint 2

Decidir:

* manter GORM v1 temporariamente;
* atualizar ORM antes de continuar;
* utilizar SQL explícito para módulos críticos;
* estratégia final das migrations.

## Marco 4 — Final da Sprint 3

Autorizar ou bloquear:

* remoção definitiva de SQLite;
* implantação da base multitenant;
* criação dos ambientes de homologação;
* início do desenvolvimento do portal.

## Marco 5 — Final da Sprint 4

Validar:

* isolamento multitenant;
* modelo de autorização;
* impacto no legado;
* estratégia para contratos e workflows de aprovação.

---

# 10. Ambientes iniciais

## Desenvolvimento local

```text
Docker Compose
PostgreSQL
RabbitMQ
EthPhish Server
Reverse proxy
SMTP de desenvolvimento
```

## Integração

```text
Ambiente efêmero por pull request, quando possível
Banco isolado
Dados sintéticos
Testes automatizados
Sem envio externo
```

## Homologação

```text
Infraestrutura semelhante à produção
Domínios de teste
Provedores de sandbox
Dois tenants fictícios
Dois workers de teste
Sem participantes reais
```

## Produção

Não será criada durante as primeiras cinco sprints. Sua habilitação dependerá de:

* multitenancy validada;
* backup restaurado;
* OIDC;
* workflow de aprovação;
* hardening;
* pentest;
* piloto controlado.

---

# 11. Riscos do ciclo inicial

| Risco                                     | Probabilidade | Impacto | Tratamento                                          |
| ----------------------------------------- | ------------: | ------: | --------------------------------------------------- |
| Legado incompatível com PostgreSQL        |          Alta |    Alto | Testes de caracterização e migração em duas sprints |
| Ausência de isolamento multitenant        |         Média | Crítico | RLS, middleware e testes ofensivos                  |
| Dependências obsoletas                    |          Alta |    Alto | Inventário e atualização incremental                |
| Escopo excessivo                          |          Alta |   Médio | Limite de 28–32 pontos por sprint                   |
| Configuração central de 4 GB insuficiente |         Média |   Médio | Limites de containers e observabilidade leve        |
| Falta de QA dedicado                      |         Média |    Alto | Testes automatizados como gate                      |
| Vazamento de segredos no CI               |         Média | Crítico | Secret scanning e cofre de segredos                 |
| Exposição do painel                       |         Baixa | Crítico | Teste automatizado de rede e portas                 |

---

# 12. Indicadores do ciclo

Ao final das primeiras cinco sprints, deverão ser medidos:

```text
Cobertura dos fluxos críticos
Tempo do pipeline
Taxa de sucesso dos builds
Vulnerabilidades abertas por severidade
Tamanho da imagem
Tempo de inicialização
Tempo de execução das migrations
Quantidade de queries incompatíveis
Resultado do teste de restauração
Quantidade de testes multitenant
Tentativas de acesso cruzado bloqueadas
```

---

# 13. Resultado esperado após dez semanas

Ao final deste ciclo, o EthPhish deverá possuir:

* repositório e governança próprios;
* baseline rastreável do Anglerphish;
* container endurecido;
* CI/CD com gates básicos;
* PostgreSQL como banco principal;
* migrations controladas;
* backup e restauração testados;
* SQLite removido do runtime de produção;
* modelo inicial de tenants e empresas;
* isolamento de dados no banco e na aplicação;
* trilha de auditoria inicial;
* base segura para iniciar autenticação, workflows de aprovação e portal do cliente.

O ciclo seguinte deverá começar por:

```text
OIDC e proteção administrativa
        ↓
Cadastro ampliado de participantes
        ↓
Contratos e workflow de aprovação
        ↓
Portal multitenant
        ↓
Provedores de entrega e workers
```
