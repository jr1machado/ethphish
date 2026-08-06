# Release notes — EthPhish v0.4.0

Data: 2026-08-06

## Resumo

Release do workflow de contratos e aprovação de campanhas, com portal
próprio para o cliente aprovador, além de perfis de participantes ampliados
e identidade visual EthPhish. Fecha a lacuna apontada em
[ISSUES_CONHECIDOS.md](ISSUES_CONHECIDOS.md) da v0.3.0: "multitenancy ainda
não é auto-serviço" — agora existe um fluxo formal de contrato → escopo →
aprovação → liberação de campanha, auditável e com evidência exportável.
Não é ainda um portal de onboarding self-service completo — ver Issues
conhecidos desta versão.

Substitui as notas da v0.3.0 (tag `v0.3.0`, commit `b4f54c9`) como release
corrente; as notas anteriores permanecem disponíveis no histórico do
repositório e na tag correspondente.

## Entregas

### Workflow de contratos e aprovação de campanhas

- **Contratos**: cadastro com nome, cliente e status (draft/active/archived),
  responsáveis por aprovação (nome + e-mail) e versionamento de documento —
  cada upload de escopo gera uma nova `ContractVersion`, sem sobrescrever a
  anterior (`models/contract.go`, `controllers/api/contract.go`,
  `templates/contracts.html`).
- **Central de Aprovações**: emissão de aprovação por versão do contrato,
  status (pending/approved/rejected/changes_requested/expired), prazo de
  expiração, reenvio de lembrete, thread de comentários entre admin e
  cliente, e **exportação de evidência em JSON** (contrato + versão +
  request + thread completo) para auditoria (`models/approval.go`,
  `controllers/api/approval.go`).
- **Token de aprovação de uso único**: link mágico com token opaco de 32
  caracteres, armazenado apenas como hash SHA-256, amarrado a uma versão
  exata do contrato e a um aprovador específico — a identidade do aprovador
  vem do token, nunca de input do cliente. Expira por tempo e por decisão
  (uma vez decidido, o link para de funcionar mesmo dentro da validade).
- **Portal do cliente aprovador**: rotas públicas `/approvals/*` no servidor
  de phishing (porta 9443), completamente segregadas do painel
  administrativo — sessão própria (`ethphish_client`, cookie `HttpOnly`,
  `Secure`, `SameSite=Lax`), CSRF dedicado, sem exigir conta nem senha do
  cliente. Ações: aprovar, rejeitar, solicitar alterações, comentar.
- **Bloqueio de campanha por aprovação, em dois pontos**: a criação da
  campanha (`Campaign.Validate`) e o **envio pelo worker** (`processCampaigns`
  / `processSMSCampaigns`) checam `IsCampaignApproved` — uma campanha já
  enfileirada é **pausada no disparo**, não só barrada na criação, se a
  aprovação expirar ou for invalidada por uma nova versão do contrato
  enquanto ela esperava na fila.
- **Invalidação automática por nova versão**: subir uma nova versão do
  documento do contrato invalida, para fins de gate de campanha, qualquer
  aprovação anterior — a aprovação é sempre amarrada à versão exata que foi
  revisada.
- Cron de lembrete/expiração (`approvals.StartScheduler`, iniciado junto do
  admin server) expira automaticamente aprovações pendentes vencidas e
  reenvia lembretes para as que passam do intervalo configurado, sem
  intervenção manual.

### Participantes e grupos

- Novos campos de perfil no cadastro de participante: departamento, empresa,
  cidade, estado, país, unidade e tags (`BaseRecipient`).
- Reconhecimento automático dessas colunas na importação CSV, mais
  **importação XLSX no navegador** (SheetJS vendorizado, sem round-trip ao
  servidor para conversão), com validação e preview antes de confirmar.
- Filtros dinâmicos por empresa, departamento, cidade, estado, país e tag na
  tela de Grupos.

### Identidade visual

- Logos e temas claro/escuro próprios do EthPhish, substituindo o seletor de
  tema herdado do Anglerphish/Gophish.

### Plataforma e configuração

- Novas variáveis: `ETHPHISH_PHISH_CSRF_KEY` (protege o CSRF do portal do
  cliente — sem ela, cada restart do processo gera uma chave nova e invalida
  formulários de aprovação em voo) e `ETHPHISH_APPROVAL_PORTAL_BASE_URL`
  (base pública usada para montar o link mágico nos e-mails de aprovação).
- Emissão de certificado sob demanda liberada para acesso externo nas
  portas 9443/9444 do reverse proxy (antes restrita a loopback/rede
  interna em alguns cenários de desenvolvimento).
- Correção de segurança: criação de usuário passa a **exigir** escopo de
  tenant explícito — o fallback anterior que criava usuário sem tenant
  associado foi removido.

### Correções encontradas na validação manual desta sprint

Ver relatório completo com evidências (prints e export real de aprovação)
em [`INFO/Validacao-S05-S06/RELATORIO.md`](INFO/Validacao-S05-S06/RELATORIO.md).

- A listagem de contratos (`GetContractsForTenant`) não trazia versões nem
  aprovadores — a tela nunca mostrava o botão "Request Approval" embora o
  backend estivesse correto (só o endpoint de contrato único trazia os
  dados). Corrigido com preload na listagem.
- Emitir ou reenviar uma aprovação sempre reportava sucesso na tela, mesmo
  quando nenhum e-mail saía por falta de perfil SMTP no tenant. Agora a
  resposta inclui quantos aprovadores foram de fato notificados, e a UI
  avisa quando o número é menor que o total.
- Reenvio manual de lembrete não gravava `last_reminder_sent_at` (diferente
  do cron automático) — a coluna "Last Reminder" na Central de Aprovações
  nunca era atualizada. Corrigido para gravar em ambos os caminhos.

## Integrações

| Integração | Estado v0.4.0 | Uso |
| --- | --- | --- |
| PostgreSQL | ativo, com RLS forçado (herdado da v0.3.0) | dados, migrations, isolamento por tenant, agora incluindo `contracts`, `contract_versions`, `contract_approvers`, `approval_requests`, `approval_comments`, `client_users`, `client_sessions` |
| RabbitMQ | ativo no caminho crítico de e-mail (herdado da v0.3.0) | fila `mail.send` + retry/DLQ; e-mails de aprovação/lembrete/decisão usam o mesmo mecanismo de envio das campanhas (perfil SMTP do tenant), não uma fila própria |
| Caddy | ativo no Compose | TLS e proxy web + admin (9443/9444); 9443 agora também serve o portal do cliente aprovador, sem porta nova |
| OIDC | recurso herdado configurável | autenticação administrativa; **não** cobre o login do cliente aprovador, que usa magic link, não SSO |
| SMTP | recurso herdado, agora também usado pelo workflow de aprovação | e-mails de solicitação de aprovação, lembrete e notificação de decisão dependem de um perfil SMTP configurado no tenant — sem ele, a aprovação é criada mas ninguém é notificado (ver Issues conhecidos) |
| GitHub Actions | workflow de release corrigido para nomear artefatos/binários como `ethphish-*` em vez de `anglerphish-*`/`gophish` | build e publicação de release |

## Upgrade e rollback

1. Faça backup com `./scripts/backup-postgres.sh` antes de atualizar.
2. Aplique as novas migrations (`20260806100000_add_contracts_approvals`,
   `20260806110000_add_approver_tokens`) via o passo `db-migrate` do Compose
   — automático em `docker compose up -d`.
3. Defina `ETHPHISH_APPROVAL_PORTAL_BASE_URL` com a URL pública real do
   servidor de phishing (porta 9443) e `ETHPHISH_PHISH_CSRF_KEY` com uma
   chave fixa antes de expor o portal do cliente em produção — sem a
   segunda, um restart invalida qualquer aprovação em andamento.
4. Configure ao menos um perfil SMTP por tenant que for usar o workflow de
   aprovação; sem ele, o botão "Request Approval" cria o request mas não
   notifica ninguém (a UI agora avisa isso, ver Correções acima).
5. Em falha, interrompa a atualização e restaure somente em banco isolado a
   partir do dump validado; não execute migrations `down` diretamente em
   produção.

## Limitações de release

Consulte [ISSUES_CONHECIDOS.md](ISSUES_CONHECIDOS.md). Conta de emergência
(gap remanescente do escopo do Sprint 5), auditoria de login do portal do
cliente, workers distribuídos em nodes externos e portal de onboarding
self-service completo permanecem fora desta release.
