# Release notes — EthPhish v0.5.0

Data: 2026-08-06

## Resumo

Release do portal do cliente ampliado (de decisão pontual pra dashboard
contínuo) e do subsistema completo de treinamento e quiz — fecha o ciclo
"medir → treinar → comprovar" que a v0.4.0 abriu com contratos e aprovação.
Certificado de conclusão e dashboard agregado de indicadores de treinamento
ficam fora desta release por decisão de corte — ver
[ISSUES_CONHECIDOS.md](ISSUES_CONHECIDOS.md).

Substitui as notas da v0.4.0 (tag `v0.4.0`, commit `2d77197`) como release
corrente; as notas anteriores permanecem disponíveis no histórico do
repositório e na tag correspondente.

## Entregas

### Portal do cliente completo

- Extensão do portal do cliente aprovador existente: nova área `/portal/*`
  na porta pública 9443, sessão compartilhada com `/approvals/*`
  (`ethphish_client`).
- Dashboard lista **todas as campanhas do tenant** (não só as vinculadas a
  contratos que o cliente aprova), com indicadores agregados
  (enviados/abertos/clicados/submetidos/reportados) — nunca um dado
  nomeado por alvo, seguindo o perfil "Cliente executivo" descrito no
  escopo do Sprint 7.
- Detalhe por campanha e exportação CSV das mesmas métricas agregadas.
- **Login self-service**: cliente digita o e-mail, sistema verifica se é
  um aprovador de contrato conhecido em algum tenant e manda um link de
  uso único (15 minutos); resposta é sempre a mesma mensagem genérica,
  independente de o e-mail existir ou não (anti-enumeração).
- Novas funções de modelo "tenant-wide" (`GetCampaignSummariesForTenantAllUsers`,
  `GetCampaignSummaryForTenantAllUsers`, `GetAnySMTPForTenant`), distintas
  das equivalentes já existentes que filtram por admin dono do recurso.

### Treinamento e quiz

- **Autoria**: nova tela admin "Trainings" — lições HTML sequenciais e
  quiz opcional com perguntas de múltipla escolha e verdadeiro/falso
  livremente misturadas, nota mínima e limite de tentativas configuráveis
  por treinamento.
- **Atribuição direta**: admin escolhe um grupo e o sistema cria um
  acesso único por alvo, com e-mail automático (mesmo mecanismo de perfil
  SMTP por tenant já usado nas aprovações).
- **Teachable moment**: campanha ganha campo opcional de treinamento e
  gatilho (clique, submissão de dados, ou ambos) no wizard de criação,
  editável depois. Quando o evento configurado dispara, o alvo é
  redirecionado automaticamente para o treinamento — o clique/submit
  continua sendo registrado normalmente na campanha antes do redirect, sem
  perda de métrica.
- **Entrega pública**: `/training/{token}` — lições em sequência (marcadas
  como vistas), quiz no fim, nota calculada no servidor com comparação
  tolerante a maiúsculas/espaços, aprovação/reprovação conforme o
  percentual mínimo configurado, bloqueio ao atingir o limite de
  tentativas.
- Token de acesso ao treinamento é **texto puro**, não hash — diferente
  dos magic links de aprovação/login (decisão única), o acesso ao
  treinamento é revisitado várias vezes ao longo de lições e tentativas de
  quiz, seguindo o mesmo modelo do `rid` de campanha, não o de um link de
  aprovação.

### Correções encontradas na validação manual desta sprint

- **Deadlock pré-existente, não introduzido nesta sprint**: `getCampaignStats`
  usava a conexão global do banco em vez do handle de transação recebido
  de `withTenantTransaction`; sob um pool de conexão limitado a uma
  (banco de teste sqlite3), a segunda consulta trava esperando por uma
  conexão que a transação aberta já segura. Já afetava
  `GetCampaignSummariesForTenant`/`GetCampaignSummaryForTenant`, em
  produção desde a v0.4.0, só nunca tinha sido exercitado por teste.
  Corrigido passando o handle correto (`tx` ou `db`) em cada chamador.
- Perguntas de quiz salvavam com opções e resposta correta vazias — os
  campos estavam marcados `json:"-"`, então a API de autoria descartava
  esses dados silenciosamente em toda criação.
- Formulário de quiz renderizava sem nenhuma pergunta, sem erro visível —
  `OptionsList()` tinha receiver ponteiro, não chamável pelo
  `html/template` numa variável de `range` não endereçável.
- Rotas públicas de treinamento devolviam 403 em tudo — faltava o
  fallback de chave CSRF vazia presente nas outras duas áreas do
  portal.

Todas confirmadas com evidência de antes/depois (print + log de erro) antes
da correção — nenhuma foi hipotética.

## Integrações

| Integração | Estado v0.5.0 | Uso |
| --- | --- | --- |
| PostgreSQL | ativo, com RLS forçado nas tabelas originais (herdado) | dados, agora incluindo `trainings`, `training_lessons`, `training_quizzes`, `quiz_questions`, `training_assignments`, `training_lesson_views`, `quiz_attempts`, `portal_login_tokens` |
| RabbitMQ | ativo no caminho crítico de e-mail (herdado) | e-mails de atribuição de treinamento usam o mesmo caminho de envio de campanha/aprovação |
| SMTP | recurso herdado | e-mails de atribuição de treinamento e de login self-service do portal dependem de um perfil SMTP configurado no tenant, mesma limitação já documentada para aprovações |
| Caddy | ativo no Compose | 9443 agora também serve `/portal/*` e `/training/*`, sem porta nova |

## Upgrade e rollback

1. Faça backup com `./scripts/backup-postgres.sh` antes de atualizar.
2. Aplique as novas migrations (`20260806120000_add_portal_login_tokens`,
   `20260806130000_add_training_and_quiz`) via o passo `db-migrate` do
   Compose — automático em `docker compose up -d`.
3. Configure ao menos um perfil SMTP por tenant que for usar atribuição
   direta de treinamento ou login self-service do portal; sem ele, o
   recurso funciona no banco mas ninguém recebe e-mail.
4. Em falha, interrompa a atualização e restaure somente em banco isolado a
   partir do dump validado; não execute migrations `down` diretamente em
   produção.

## Limitações de release

Consulte [ISSUES_CONHECIDOS.md](ISSUES_CONHECIDOS.md). Certificado de
conclusão, dashboard de indicadores de treinamento, treinamento visível no
portal do cliente, e conta de emergência (gap remanescente da v0.4.0)
permanecem fora desta release.
