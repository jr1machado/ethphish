# EthPhish v0.5.0

EthPhish é um **fork independente** do Anglerphish 1.3.0 (que por sua vez
deriva do Gophish), evoluído como uma **plataforma corporativa completa e
autônoma** para testes éticos, autorizados e mensuráveis de phishing e
quishing (simulações via QR Code). Não é uma distribuição, um branch, nem um
produto oficial do Anglerphish ou do Gophish — segue caminho próprio de
versionamento, arquitetura, banco de dados, identidade visual e governança de
release, documentado neste arquivo e em [CHANGELOG.md](CHANGELOG.md).

O produto apoia programas de conscientização, GRC e redução de risco humano.
Toda campanha deve possuir autorização, escopo, período, público e domínios
aprovados. O EthPhish não contempla evasão de controles, coleta de credenciais
reais, payloads, exploração de vulnerabilidades ou uso não autorizado.

## Visão comercial

Para C-levels e executivos, o EthPhish converte o risco humano em evidência
operacional: mostra onde a organização é mais suscetível a engenharia social,
mede a evolução após treinamentos e dá governança sobre campanhas de
conscientização, com trilha de aprovação e isolamento de dados por cliente ou
unidade de negócio.

Desde a v0.4.0, essa governança deixou de ser apenas um processo manual ao
redor da ferramenta e virou **controle técnico dentro dela**: nenhuma
campanha vinculada a um contrato sai — nem na criação, nem no disparo pelo
worker — sem uma aprovação formal, registrada com hora, autor e escopo exato
aceito, e revogada automaticamente se o escopo mudar depois. Isso transforma
"confiamos que o time seguiu o processo" em "o sistema não deixa sair sem
aprovação documentada".

A v0.5.0 fecha o outro lado do programa: quando alguém cai na simulação, o
EthPhish já entrega o treinamento corretivo automaticamente — sem depender
de outra ferramenta ou de um processo manual de follow-up — e também permite
atribuir conscientização recorrente a qualquer grupo, com nota, aprovação e
tentativas registradas por pessoa. O programa deixa de ser só "medir quem
clicou" e passa a fechar o ciclo: medir, treinar, comprovar.

**Casos de uso**

- Programas recorrentes de security awareness com métricas comparáveis entre
  ciclos e áreas.
- Validação de controles humanos antes de auditorias (ISO 27001, PCI-DSS,
  LGPD/GDPR) com evidência exportável — incluindo, agora, o próprio
  comprovante de aprovação do escopo testado (quem aprovou, quando, qual
  versão exata do documento).
- Avaliação de risco por diretoria, unidade de negócio ou aquisição recente
  (due diligence de segurança pós-M&A).
- Simulações aprovadas de resposta a incidente real, sem reexpor a
  organização a coleta de credenciais verdadeiras.
- Provedores de serviços gerenciados (MSSPs) operando múltiplos clientes na
  mesma plataforma, com isolamento de dados garantido no banco (RLS), não
  apenas na aplicação — e agora com um **portal próprio para o cliente final
  aprovar o escopo** sem precisar de conta no sistema nem acesso ao painel do
  MSSP.
- Times jurídico/compliance que precisam assinar formalmente o escopo de um
  pentest de engenharia social antes da execução, com registro de quem
  autorizou e qual versão do contrato foi de fato testada — sem depender de
  e-mail avulso ou planilha de controle paralela.
- Programas de conscientização que precisam provar que quem clicou também
  foi treinado — não só identificar o clique — com nota, aprovação e
  histórico de tentativas por colaborador, prontos para auditoria.
- Clientes finais (via portal) acompanhando a evolução do próprio programa
  sem precisar pedir relatório ao time de segurança toda vez.

**Dores que o EthPhish resolve**

- Falta de métricas consistentes e comparáveis de comportamento humano ao
  longo do tempo.
- Campanhas sem controle de escopo, autorização ou rastreabilidade de
  aprovação — risco jurídico e reputacional para quem executa o teste. Agora
  isso é bloqueado no próprio sistema, não apenas em processo.
- "A campanha já estava aprovada quando foi criada, mas o escopo mudou
  depois e ninguém percebeu" — resolvido pela invalidação automática da
  aprovação quando uma nova versão do contrato é enviada, e pela checagem
  repetida no momento do disparo, não só na criação.
- Relatórios manuais, sem padronização, difíceis de apresentar a auditoria ou
  ao conselho.
- Ausência de uma base técnica que suporte múltiplos clientes/áreas sem risco
  de vazamento cruzado de dados entre eles.
- Dependência de operação manual do time de segurança para disparar e
  monitorar cada campanha, e para perseguir aprovadores por e-mail —
  lembretes e expiração de aprovações pendentes agora rodam sozinhos.
- "Quem clicou recebeu algum treinamento depois?" — hoje isso costuma ser
  outra ferramenta, outra planilha ou simplesmente não acontece. Agora é
  automático: clicar (ou submeter) redireciona direto para a lição
  configurada, sem etapa manual.
- Cliente final sem visibilidade contínua do programa — antes só via
  aprovação pontual de contrato, agora tem portal próprio com indicadores,
  histórico e exportação, sem precisar abrir chamado com o time de
  segurança.

A plataforma preserva as fronteiras éticas em todas as camadas: mede eventos
de interação autorizados (abertura, clique, submissão simulada, reporte) e
não armazena senhas, OTPs ou outras credenciais reais.

## O que mudou neste fork

Desde a base Anglerphish 1.3.0:

- identidade, versionamento e documentação próprios do EthPhish, sem vínculo
  com releases do Anglerphish ou do Gophish;
- PostgreSQL como único banco de runtime suportado (SQLite permanece apenas
  para testes e importação legada controlada);
- migrations PostgreSQL executadas por um role privilegiado dedicado
  (`db-migrate`, passo do Compose que roda e termina antes do `server`
  subir), com pool configurável e advisory lock contra concorrência;
- **fundação multitenant**: `tenants`, `companies`, `tenant_users` e um
  contexto de requisição tipado (`TenantScope`) que barra qualquer fluxo
  tenant-owned sem `tenant_id`/`user_id` validados;
- **escopo por tenant aplicado** a campanhas, grupos, alvos, templates,
  landing pages, perfis SMTP, perfis e templates SMS, IMAP, webhooks e
  relatórios — incluindo os fluxos que não passam por uma requisição HTTP
  (monitor IMAP em background, MFA por SMS, geração de relatórios em fila);
- **PostgreSQL Row-Level Security (RLS)** com `FORCE ROW LEVEL SECURITY` em
  toda tabela tenant-owned, aplicada por uma policy que lê
  `ethphish.tenant_id` da transação; a aplicação roda em runtime sob um role
  restrito (`ethphish_app`, sem `SUPERUSER`/`BYPASSRLS`) porque o role
  padrão do Postgres é superusuário e ignora RLS silenciosamente;
- **fila durável RabbitMQ no caminho crítico de e-mail**: o disparo de
  e-mail de campanha passou de canal Go in-process para publicação por
  `MailLog` na fila `mail.send`, com fila de retry por TTL/DLX e fila morta
  terminal para falhas de processamento (não confundir com o retry SMTP que
  já existia via backoff em `MailLog`); cai de volta ao canal direto quando
  `ETHPHISH_RABBITMQ_URL` não está definido;
- **admin UI publicada por proxy reverso** em listener próprio (9444),
  segregada da web pública de campanhas (9443), sem expor a porta 3333
  diretamente no host;
- Docker multi-stage, usuário não-root, capabilities removidas e health
  checks em todos os serviços do Compose;
- TLS autoassinado para desenvolvimento, roteado pelo Caddy, com emissão sob
  demanda liberada para acesso externo nas portas 9443/9444;
- CI com formatação, vet, testes, integração PostgreSQL, secret scan,
  vulnerabilidades, scan de imagem e SBOM;
- backup PostgreSQL, restore ensaiado e documentação de segurança,
  arquitetura e governança;
- **workflow de contrato e aprovação de campanha** (v0.4.0): contratos
  versionados, aprovadores nomeados, aprovação por magic link de uso único,
  Central de Aprovações com comentários e exportação de evidência, portal
  próprio para o cliente aprovador (sem conta no sistema), e bloqueio de
  campanha reforçado tanto na criação quanto no disparo pelo worker;
- **perfis de participante ampliados** (v0.4.0): departamento, empresa,
  cidade, estado, país, unidade e tags, com importação CSV/XLSX no
  navegador, validação, preview e filtros dinâmicos na tela de Grupos;
- **identidade visual própria** (v0.4.0): logos e temas claro/escuro
  EthPhish, substituindo o seletor herdado do Anglerphish/Gophish;
- correção de segurança: criação de usuário passa a exigir `tenant_id`
  explícito, sem fallback para usuário sem tenant;
- todos os binários, artefatos de release e nomes de arquivo gerados pela
  aplicação (relatórios, workflow de release do GitHub) usam **EthPhish**,
  nunca `anglerphish`/`gophish` — ver [Convenção de nomes](#convenção-de-nomes);
- **portal do cliente ampliado** (v0.5.0): além de decidir aprovações, o
  cliente agora acompanha todas as campanhas do tenant com indicadores
  agregados, exporta CSV e entra por link de login self-service (sem
  esperar uma aprovação pendente) — nunca expõe dado nomeado por alvo;
- **treinamento e quiz** (v0.5.0): lições sequenciais em HTML, quiz misto
  (múltipla escolha + verdadeiro/falso) com nota mínima e limite de
  tentativas configuráveis, entregue por atribuição direta a um grupo
  (e-mail automático) ou por redirecionamento automático pós-clique/submit
  de uma campanha ("teachable moment"), configurável no wizard de campanha.

## Arquitetura

```text
                         rede administrativa privada
                                      │
                          VPN / Zero Trust / OIDC
                                      │
                    internet ── HTTPS 9443/9444 ── Caddy (reverse-proxy)
                                      │ TLS interno
                           ┌──────────▼──────────┐
                           │ servidor central     │
                           │ admin (3333) + web   │
                           │ (8080) + API         │
                           │ scheduler, IMAP,     │
                           │ relatórios, worker    │
                           │ interno (pool mail)   │
                           └───┬───────┬──────┬───┘
                               │       │      │
                     PostgreSQL│  AMQP │      │futuro: AMQP TLS 5671
                          (RLS)│ 5672  │      ▼
                          ┌────▼───┐ ┌─▼──────────┐  ┌───────────────┐
                          │ dados  │ │ RabbitMQ    │─▶│ worker node 1 │
                          │(ethphish│ │ mail.send   │  │ worker node N │
                          │_app role│ │ + retry/DLQ │  │  (futuro)     │
                          └────────┘ └─────────────┘  └───────────────┘
```

Na v0.5.0, o servidor central concentra admin, API, web de campanhas,
**portal do cliente aprovador** (`/approvals/*` — decisão pontual),
**portal do cliente completo** (`/portal/*` — indicadores contínuos,
login self-service), **entrega de treinamento/quiz** (`/training/{token}`),
scheduler (incluindo o cron de lembrete/expiração de aprovações) e um
**pool de goroutines interno** que consome a fila `mail.send`; não existe
ainda um processo worker externo ao servidor. A separação real em worker
nodes (identidade e segredos próprios, sem acesso direto ao PostgreSQL,
comunicação por AMQP TLS 5671) permanece como arquitetura-alvo — RabbitMQ
hoje já está no caminho crítico de entrega de e-mail (campanha, aprovação
**e** atribuição de treinamento), mas isso não deve ser confundido com
workers distribuídos.

Todas as três áreas do cliente (`/approvals/*`, `/portal/*` e
`/training/{token}`) são servidas pelo mesmo processo e pela mesma porta
pública (9443) da web de campanhas — não abrem porta nova. `/approvals/*` e
`/portal/*` compartilham sessão/cookie (`ethphish_client`) com proteção
CSRF própria; `/training/{token}` não tem sessão — o token na URL identifica
o acesso diretamente, o mesmo modelo de confiança do `rid` de campanha, já
que o treinamento é revisitado várias vezes (não é uma decisão única como
uma aprovação). Nenhuma dessas três áreas tem CSRF compartilhado com o
restante da porta 9443, que por natureza não tem proteção CSRF (recebe
POSTs cross-origin legítimos de formulários de captura de credenciais
simuladas).

Consulte o detalhamento em [arquitetura alvo](docs/architecture/target-architecture.md).

## Containers e portas

| Componente | Papel | Portas | Exposição |
| --- | --- | --- | --- |
| `reverse-proxy` (Caddy) | TLS público e roteamento web + admin + portal/treinamento do cliente | 9443/TCP web (campanhas + `/approvals/*` + `/portal/*` + `/training/*`), 9444/TCP admin, publicadas no host | somente para redes autorizadas (VPN/Zero Trust); nunca publicar 9444 na internet aberta; 9443 pode ser exposta ao público-alvo autorizado de uma campanha, aos aprovadores de contrato e a quem recebeu treinamento |
| `server` | administração, API, campanhas, portal/aprovação/treinamento do cliente, worker interno de e-mail | 3333/TCP admin, 8080/TCP web (inclui `/approvals/*`, `/portal/*`, `/training/*`) | somente redes Docker internas (`admin_internal`, `application_internal`) |
| `db-migrate` | aplica migrations com role privilegiado, roda uma vez e termina | nenhuma publicada | rede de dados interna, sem rede externa |
| `postgres` | dados, RLS e migrations | 5432/TCP | rede de dados interna (`data_internal`) |
| `rabbitmq` | fila durável de e-mail em produção | 5672/TCP AMQP, 5671/TCP AMQP TLS (worker nodes futuros) | rede de dados interna |
| `tls-init` | prepara volume privado de certificados | nenhuma | execução única, sem rede |
| `worker-node` (futuro) | entrega aprovada e observabilidade, sem painel nem banco direto | AMQP TLS 5671 de saída; SMTP/HTTPS conforme escopo | rede privada dedicada a workers |

## Comunicação servidor central ↔ workers

- **Hoje**: o "worker" é um pool de goroutines dentro do processo `server`,
  que consome `mail.send` do RabbitMQ pela rede de dados interna (5672/TCP,
  sem TLS em desenvolvimento). Não há uma porta de comunicação central↔worker
  externa porque não há processo worker externo.
- **Arquitetura-alvo**: workers como processos/containers independentes,
  sem acesso administrativo nem ao PostgreSQL, autenticando e consumindo
  jobs assinados via AMQP TLS na porta 5671, em rede privada dedicada. O
  central valida tenant, aprovação, domínio, quota e janela de execução
  antes de publicar o job; o retorno de resultado (idempotente) usa uma API
  interna autenticada ou fila de retorno, a definir em ADR futuro.
- **Crescimento como nodes**: cada worker node novo é idêntico aos demais
  (mesma imagem, sem estado próprio além de credenciais/identidade),
  consome da mesma fila e escala horizontalmente sem replicar banco ou
  painel administrativo. Backlog, taxa de sucesso, retries, DLQ, latência,
  CPU e memória por node devem ser monitorados para decidir quando adicionar
  o próximo node.

## Recursos implementados até a v0.5.0

**Herdados do baseline Anglerphish 1.3.0** (ver detalhamento completo em
[FEATURES.md](FEATURES.md)):

- campanhas por e-mail, SMS, QR Code e canal genérico, campaign sets;
- gestão de grupos, participantes, templates, landing pages e perfis de
  envio (SMTP e SMS — Twilio/Vonage);
- eventos de abertura, clique, submissão simulada, reporte e resposta;
- simulação de MFA, landing pages com Basic Auth, variáveis de template e
  variáveis globais;
- monitor IMAP com leitura de mensagens reportadas e aba de respostas;
- relatórios Word/Excel com opção de anonimização;
- OIDC para SSO administrativo, CSRF, autenticação e testes de
  caracterização do baseline;
- criptografia AES-256-GCM de campos sensíveis em banco.

**Entregues nas sprints 01–04 (EthPhish)**:

- runtime e imagem exclusivos para PostgreSQL 17, sem CGO, sem Goose;
- migrations idempotentes, pool configurável, advisory lock, TLS de banco
  opcional (`ETHPHISH_DB_REQUIRE_TLS=true`) e health/readiness sem dados de
  conexão (`/healthz`, `/readyz`);
- backup lógico e restore ensaiado em banco isolado;
- pré-flight, preparador de schema isolado e importador transacional
  SQLite→PostgreSQL, validados em CI contra PostgreSQL efêmero;
- fundação multitenant (`tenants`, `companies`, `tenant_users`,
  `TenantScope`) e escopo por tenant em **todas** as entidades de negócio
  prioritárias (campanhas, grupos, alvos, templates, landing pages, SMTP,
  SMS, IMAP, webhooks, relatórios);
- **PostgreSQL RLS com `FORCE ROW LEVEL SECURITY`** e role de runtime
  restrito (`ethphish_app`), com teste de integração que abre uma segunda
  conexão sob o role restrito e comprova bloqueio de leitura/escrita
  cruzada entre tenants;
- **fila RabbitMQ durável** para disparo de e-mail de campanha, com retry
  por TTL/DLX e fila morta terminal, isolada do retry SMTP existente;
- admin UI acessível por proxy reverso dedicado (9444), sem publicar a
  porta 3333 diretamente;
- CI/CD e controles de supply chain descritos em
  [release notes](RELEASE_NOTES.md).

**Entregues na v0.4.0 (EthPhish)**:

- **contratos**: cadastro, status (draft/active/archived), aprovadores
  nomeados (nome + e-mail) e versionamento de documento de escopo — cada
  upload gera uma nova versão, sem sobrescrever a anterior;
- **workflow de aprovação**: emissão de aprovação por versão do contrato,
  magic link de token único (hash SHA-256, nunca armazenado em texto claro),
  expiração por tempo e por decisão, reenvio de lembrete, thread de
  comentários e **exportação de evidência em JSON** auditável;
- **cron de lembrete/expiração** rodando junto do admin server, sem
  intervenção manual;
- **portal do cliente aprovador**, público em `/approvals/*` na porta 9443,
  com sessão e CSRF próprios, sem exigir conta nem senha do cliente —
  identidade vem do token do magic link;
- **bloqueio de campanha por aprovação em dois pontos**: na criação
  (`Campaign.Validate`) e no **disparo pelo worker** (`processCampaigns`,
  `processSMSCampaigns`), então uma campanha já enfileirada é pausada, não
  só barrada, se a aprovação expirar ou for invalidada enquanto esperava;
- **invalidação automática por nova versão**: subir um novo documento do
  contrato invalida qualquer aprovação anterior para fins de gate;
- **perfis de participante ampliados**: departamento, empresa, cidade,
  estado, país, unidade e tags, reconhecidos na importação CSV, com
  **importação XLSX no navegador** (SheetJS vendorizado), validação, preview
  e filtros dinâmicos na tela de Grupos;
- **identidade visual EthPhish**: logos e temas claro/escuro próprios;
- correção de segurança: criação de usuário exige `tenant_id` explícito.

**Entregues na v0.5.0 (EthPhish)**:

- **portal do cliente completo** (`/portal/*`): dashboard com todas as
  campanhas do tenant e indicadores agregados (enviados/abertos/clicados/
  submetidos/reportados) — nunca dado nomeado por alvo; detalhe por
  campanha; exportação CSV; login self-service por e-mail (token de uso
  único, 15 min), sem depender de uma aprovação pendente para entrar;
  sessão compartilhada com o portal de aprovação;
- **treinamento e quiz**: treinamentos com lições HTML sequenciais e quiz
  opcional (perguntas de múltipla escolha e verdadeiro/falso misturadas
  livremente), nota mínima e limite de tentativas configuráveis por
  treinamento;
- **atribuição direta**: tela admin "Trainings" atribui um treinamento a
  um grupo inteiro, gera um link único por alvo e dispara e-mail
  automático (mesmo mecanismo de SMTP por tenant já usado em aprovações);
- **teachable moment**: campanha ganha campo opcional de treinamento e
  gatilho (clique, submissão, ou ambos) no wizard de criação — quem
  cai na simulação é redirecionado automaticamente para o treinamento
  configurado, sem etapa manual; o clique/submit da campanha continua
  sendo rastreado normalmente antes do redirecionamento;
- correção de um deadlock latente pré-existente (não introduzido nesta
  sprint, só exposto por ela): `getCampaignStats` usava a conexão global
  do banco em vez da transação ativa, travando sob pool de conexão
  limitado — afetava também `GetCampaignSummariesForTenant`, já em
  produção desde a v0.4.0.

## Funções e recursos futuros

- workers externos ao processo do servidor, distribuídos em nodes próprios,
  sem acesso direto ao banco, comunicando por AMQP TLS 5671;
- extensão do padrão de fila durável para SMS e geração de relatórios (hoje
  ainda em polling de banco, fora de escopo desta release por design);
- transactional outbox para garantir publicação exatamente-uma-vez entre
  commit de banco e publicação na fila;
- conta de emergência (break-glass) para acesso administrativo quando o IdP
  OIDC estiver indisponível — gap identificado no escopo do Sprint 5, ainda
  não implementado (ver [Issues conhecidos](ISSUES_CONHECIDOS.md));
- log de auditoria dedicado para tentativas de login do portal do cliente
  aprovador (sucesso/falha), hoje não confirmado em log estruturado;
- aviso proativo na tela de Contratos quando o tenant não tem perfil SMTP
  configurado, antes de tentar emitir uma aprovação;
- portal de clientes multitenant **self-service completo** (onboarding
  próprio, gestão de múltiplos contratos, recuperação de acesso), além do
  fluxo atual de magic link emitido pelo admin;
- dashboard operacional e executivo, métricas de capacidade e
  observabilidade (backlog, DLQ, latência, node saúde);
- certificado de conclusão de treinamento (PDF ou equivalente) — fora de
  escopo da v0.5.0 por decisão de corte, ver
  [Issues conhecidos](ISSUES_CONHECIDOS.md);
- dashboard de indicadores de treinamento (taxa de início/conclusão, nota
  média, evolução entre tentativas, departamentos com menor adesão,
  reincidência pós-treinamento, impacto nas campanhas seguintes — Sprint08
  §14.4); os dados brutos já são gravados (`training_assignments`,
  `quiz_attempts`), falta a camada de agregação e visualização;
- treinamento exposto no portal do cliente (`/portal/*` já reserva o
  conceito, mas a listagem de treinamentos do cliente ainda não existe);
- conteúdo de vídeo/SCORM em lições de treinamento;
- bundles versionados de campanhas e conteúdos de treinamento, com
  importação/exportação controlada;
- backup automatizado, retenção, restauração recorrente e storage externo;
- alta disponibilidade, assinatura de imagens e política formal de
  atualização de dependências.

## Fora de escopo

- coleta, retenção ou recuperação de senhas, OTPs, cartões ou credenciais
  reais;
- evasão de filtros, antivírus, EDR, gateways ou mecanismos de proteção;
- payloads, exploração, comprometimento de contas ou acesso não autorizado;
- campanhas sem aprovação formal, domínios fora do escopo ou público não
  autorizado;
- exposição pública do painel administrativo fora de rede autorizada
  (VPN/Zero Trust), mesmo estando disponível via proxy reverso;
- o workflow de contrato/aprovação **não substitui** um instrumento jurídico
  de contratação — é evidência operacional de que um responsável nomeado
  aprovou um escopo técnico específico, não uma assinatura contratual;
- onboarding self-service de clientes (cadastro próprio, gestão de conta) —
  o portal do cliente aprovador só aceita quem já foi cadastrado como
  aprovador de um contrato pelo admin.

## Requisitos

- Go 1.25.12 e compilador C (compatibilidade SQLite legada nos testes);
- Docker Engine com Compose v2;
- Node 22 para reconstrução de assets;
- PostgreSQL 17 para o ambiente Compose;
- RabbitMQ 4 para o ambiente Compose (fila durável de e-mail; a aplicação
  cai de volta ao envio direto in-process se `ETHPHISH_RABBITMQ_URL` não for
  definido, mas isso não é recomendado além de desenvolvimento local).

O runtime de produção exige PostgreSQL. SQLite permanece somente para testes
e migração legada explicitamente controlada; não inicie uma imagem de
produção com `ETHPHISH_DB_DRIVER=sqlite3`.

Referência de capacidade inicial: servidor central com 2 vCPU, 4 GB RAM e
100 GB de disco (eleve a memória para 8 GB antes de ampliar volume de
tenants, eventos ou relatórios; eleve o disco se o volume de documentos de
contrato for alto — cada versão de contrato é um arquivo em
`/var/lib/ethphish/contracts`, sem expurgo automático ainda); cada worker
futuro com 1 vCPU, 1 GB RAM e 50 GB de disco. A capacidade real deve ser
dimensionada por volume, taxa de entrega aprovada, retenção de eventos e
geração de relatórios.

## Convenção de nomes

Todo artefato **gerado ou publicado** por este projeto usa o nome
**EthPhish**, nunca `anglerphish` ou `gophish`:

- binários de release (`ethphish`, `ethphish.exe`) e pacotes ZIP publicados
  (`ethphish-vX.Y.Z-<os>-<bits>.zip`) via
  [`.github/workflows/release.yml`](.github/workflows/release.yml);
- título do GitHub Release (`EthPhish vX.Y.Z`);
- nome de arquivo de relatórios exportados pela aplicação (ex.:
  `ethphish_campaign_set_report.xlsx`);
- tags de versão (`vX.Y.Z`) e entradas de [CHANGELOG.md](CHANGELOG.md).

Exceções deliberadas, que **permanecem** referenciando o upstream por
design — não são inconsistência, são rastreabilidade de proveniência:

- o arquivo `ANGLERPHISH_VERSION` e a variável interna
  `AnglerPhishVersion`/`ANGLERPHISH_ENCRYPTION_KEY`, que registram a versão
  do Anglerphish 1.3.0 usada como base do fork (ver
  [CHANGELOG.md](CHANGELOG.md): "changes below `[1.3.0]` belong to the
  upstream Anglerphish project");
- valores de exemplo em testes automatizados (`auth/oidc_test.go`,
  `controllers/controllers_test.go`) que usam `anglerphish-admins` como
  nome de grupo OIDC fictício — são dados de teste, não identidade do
  produto.

## Desenvolvimento local

```sh
docker compose build
docker compose up -d
docker compose ps
```

- Web de campanhas/quishing: `https://localhost:9443`.
- Painel administrativo: `https://localhost:9444` (publicado apenas em
  desenvolvimento local; em produção deve ficar restrito a rede
  administrativa/VPN, nunca exposto na internet aberta).

A credencial temporária de administrador é gerada apenas no primeiro início
e deve ser alterada imediatamente.

```sh
CGO_ENABLED=1 go test ./...
./scripts/backup-postgres.sh
```

Consulte [desenvolvimento local](docs/runbooks/local-development.md),
[release notes](RELEASE_NOTES.md), [issues conhecidos](ISSUES_CONHECIDOS.md) e
[política de segurança](SECURITY.md).

## Documentação

- [Arquitetura alvo](docs/architecture/target-architecture.md)
- [Inventário de dependências](docs/architecture/dependency-inventory.md)
- [Threat model](docs/security/threat-model.md)
- [Uso aceitável](docs/product/acceptable-use.md)
- [Status das Sprints 0–2](docs/project/sprint-0-2-status.md)
- [Sprint 02 — endurecimento PostgreSQL](docs/project/sprint-02.md)
- [Sprint 04 — fundação multitenant](docs/project/sprint-04.md)
- [Sprint 05 — escopo planejado](INFO/Sprint05.md) /
  [Sprint 06 — escopo planejado](INFO/Sprint06.md)
- [Validação manual Sprint 05/06 — evidências e correções](INFO/Validacao-S05-S06/RELATORIO.md)
- [Sprint 07 — escopo planejado](INFO/Sprint07.md) /
  [Sprint 08 — escopo planejado](INFO/Sprint08.md)
- [Design: portal do cliente](docs/superpowers/specs/2026-08-06-client-portal-dashboard-design.md)
- [Design: treinamento e quiz](docs/superpowers/specs/2026-08-06-training-quiz-design.md)
- [Funções e recursos herdados detalhados](FEATURES.md)
- [Release notes v0.5.0](RELEASE_NOTES.md)
- [Issues conhecidos](ISSUES_CONHECIDOS.md)

## Licença e atribuição

O EthPhish mantém a atribuição ao Gophish e ao Anglerphish nos termos da
licença MIT herdada. As mudanças deste repositório formam um fork independente
e não implicam endosso dos projetos upstream.
