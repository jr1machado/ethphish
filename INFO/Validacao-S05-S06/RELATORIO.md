# Relatório de Validação — Sprint 05 e Sprint 06

**Data da validação:** 2026-08-06
**Ambiente:** stack local via `docker compose` (postgres, rabbitmq, server, reverse-proxy Caddy), branch `feature/sprint05`, build da imagem `ethphish-dev-server` a partir do working tree atual (inclui todo o código não commitado de contratos/aprovações).
**Método:** acesso direto à aplicação — navegador headless (Chromium via Playwright) para telas administrativas e portal do cliente, e chamadas HTTP diretas (`curl`) contra a API para os cenários que a UI não permitiu completar (ver Bugs). Evidências em `screenshots/` (numeradas na ordem de execução) e export real de evidência de aprovação em `evidencia-exportada-pendente.json`.

Credencial de teste criada para esta validação: usuário `admin`, senha resetada via banco (`Valida@2026Sprint!`) — não havia senha inicial documentada/logada disponível.

---

## Resumo executivo

| Sprint | Escopo | Resultado |
|---|---|---|
| Sprint 5 (auth/OIDC + portal do aprovador) | Login OIDC/RBAC/proxy admin | ✅ já validado em sprints anteriores (fora do escopo desta rodada) |
| Sprint 5 | Conta de emergência | ❌ **não implementada** — não encontrada no código nem testável |
| Sprint 5 | Auth do aprovador (token único, sessão própria, e-mail validado) | ✅ **aprovado** |
| Sprint 6 | Contratos, versionamento, escopo, approvers | ⚠️ **aprovado no backend, quebrado na UI** (bug bloqueante — ver abaixo) |
| Sprint 6 | Central de Aprovações (emitir, comentar, reenviar, exportar) | ✅ **aprovado** |
| Sprint 6 | Bloqueio de campanha sem aprovação / liberação com aprovação / invalidação por nova versão | ✅ **aprovado** |

---

## Sprint 6 — Contratos, escopo e workflow de aprovação

### 1. Criar contrato, aprovadores e versão (`screenshots/05` a `09`)

Login (`01-login-page.png`, `02-dashboard.png`) → Contratos vazio (`05`) → formulário preenchido (`06`) → contrato salvo com sucesso (`07`, POST 201) → aprovador adicionado (`08`, POST 201) → versão PDF enviada com sucesso (`09`, POST 201, arquivo salvo em `/var/lib/ethphish/contracts/...`).

Todas as chamadas de API retornaram sucesso (confirmado nos logs do container). **Critério de aceite atendido no backend.**

### 2. 🔴 BUG BLOQUEANTE — UI de Contratos não mostra versões/aprovadores

Ao reabrir o modal do contrato (`10-contracts-reopened-with-data.png`) — inclusive após reload completo de página — as tabelas de **Approvers** e **Versions** aparecem vazias, e o botão **"Request Approval"** nunca é renderizado (ele só existe dentro da linha de uma versão listada).

**Causa raiz identificada:**
- `models.GetContractsForTenant` (`models/contract.go:93`) — usada pelo endpoint de listagem `GET /api/contracts/` — **não faz** `Preload("Versions")`/`Preload("Approvers")`, ao contrário de `GetContractForTenant` (singular, com preload, usada só no endpoint `GET /api/contracts/{id}`).
- `static/js/src/app/contracts.js` → `edit(id)` reaproveita o array `contracts` já carregado pela listagem (sem versões/aprovadores) e **nunca chama** `GET /api/contracts/{id}` para buscar o contrato completo.

Resultado: nenhum usuário consegue emitir uma aprovação pela tela — o fluxo documentado no Sprint 6 (upload de versão → "Request Approval" → Central de Aprovações) é **inacessível via UI** no estado atual, apesar do backend estar correto (confirmado com `curl` direto no endpoint singular, ver abaixo).

```
GET /api/contracts/        -> [{"id":1,...}]                      (sem versions/approvers)
GET /api/contracts/1       -> {"id":1,...,"versions":[...],"approvers":[...]}  (completo)
```

**Correção sugerida:** adicionar `.Preload("Versions").Preload("Approvers")` em `GetContractsForTenant`, ou trocar `edit()` no front-end para consumir `GET /api/contracts/{id}`.

Para prosseguir com a validação dos demais critérios, os passos seguintes (emitir aprovação, criar segunda contrato/versão) foram executados via API direta — a mesma rota que a UI usaria se o botão estivesse visível.

### 3. Emissão de aprovação e alerta de e-mail enganoso (achado secundário)

Emiti a aprovação via `POST /api/approvals/` (contract_version_id=1) → sucesso, request criado com `token_expires_at` = +7 dias.

**Ambiente de teste não tinha perfil SMTP configurado** para o tenant. Nesse cenário, `approvals.SendApprovalRequestEmail` falha silenciosamente (`log.Error` + `continue` em `controllers/api/approval.go:118-124`), mas a API retorna `201` e o front-end exibe **"Approval request sent to configured approvers"** — mensagem de sucesso mesmo quando nenhum e-mail foi de fato enviado. Recomenda-se refletir falha de envio na resposta (ex.: contagem de e-mails enviados vs. aprovadores).

Para validar o portal do cliente sem um servidor SMTP real, o token do aprovador foi substituído diretamente no banco (mesmo algoritmo de hash usado pela aplicação — SHA-256) para simular a entrega do link por e-mail. Isso é uma limitação do ambiente de teste, não da aplicação.

### 4. Central de Aprovações (`screenshots/12` a `15`, `22`)

- Listagem mostra a aprovação pendente com status, prazos (`12`).
- Modal de detalhe abre corretamente (`13`).
- Comentário do lado admin adicionado com sucesso e refletido no thread (`14`).
- **Export Evidence** baixa JSON real com contrato, versão, request e thread de comentários (`evidencia-exportada-pendente.json`) — **critério "comprovante pode ser consultado e exportado" atendido**.
- **Resend** funciona (flash de sucesso, `15`) — porém a coluna "Last Reminder" não é atualizada na tabela sem recarregar a página (achado menor, cosmético).
- Após decisão do cliente, status vira **approved** (verde) na tela (`22`) — sincronização entre portal e admin confirmada.

### 5. Portal do cliente (`screenshots/17` a `20`)

- Link mágico (`/approvals/login?token=...`) redireciona direto para a tela de decisão da versão específica (`17`), mostrando escopo, documento e comentários do admin — **isso confirma que o e-mail validado do aprovador vira a identidade da sessão, sem input do cliente**.
- Decisão **Approve** registrada com sucesso, tela atualiza para status "approved" (`18`).
- Token **sem parâmetro** e token **inválido** mostram a mesma página de erro genérica "Link invalid or expired" (`19`, `20`) — sem vazar detalhes se o link nunca existiu vs. expirou (bom para segurança).
- Reuso do mesmo token **após a decisão** foi tentado (`curl`) e devolve a mesma página de erro — **token de uso único efetivamente aplicado** (a validade é amarrada ao status `pending` do approval request, não apenas ao token).

### 6. Bloqueio/liberação de campanha por aprovação (via API, backend puro)

| Cenário | Resultado |
|---|---|
| Campanha vinculada a contrato **sem** aprovação | `400` — `"Campaign requires an approved contract before it can run"` |
| Campanha vinculada a contrato **com versão aprovada** | `201` — campanha criada e **"In progress"** (`21-campaigns-list-with-approved-campaign.png`) |
| Upload de **nova versão** do contrato já aprovado, depois nova tentativa de campanha | `400` novamente — aprovação anterior **invalidada automaticamente** pela troca de versão |

Todos os três critérios de aceite de bloqueio do Sprint 6 foram confirmados na prática, batendo com os testes automatizados já existentes em `models/contract_approval_test.go`.

---

## Sprint 5 — Autenticação e portal do aprovador

| Critério | Status | Evidência |
|---|---|---|
| Token sozinho não concede acesso a outros recursos | ✅ | sessão limitada ao path `/approvals`, cookie `HttpOnly`+`Secure`+`SameSite=Lax` |
| Tokens expirados/usados/revogados são rejeitados | ✅ | `19`, `20` + reuso pós-decisão testado via curl |
| Todas as tentativas são auditadas | ⚠️ parcial | tentativas de login admin aparecem no log estruturado do server (`level=info`); não foi encontrado log dedicado para tentativas de login do **portal do cliente** (sucesso/falha) — recomenda revisão |
| Conta de emergência | ❌ | não localizada no código (`grep` por "emergency"/"break-glass" não retornou nada) nem no histórico de commits |

---

## Arquivo de evidências

```
INFO/Validacao-S05-S06/
├── RELATORIO.md                        (este arquivo)
├── evidencia-exportada-pendente.json   (export real do endpoint /api/approvals/1/export)
└── screenshots/
    01-login-page.png … 25-contracts-list-two-contracts.png
```

## Achados consolidados (por severidade)

1. ~~**[Alto] UI de Contratos não exibe versões/aprovadores**~~ — **CORRIGIDO.**
2. ~~**[Médio] Mensagem de sucesso enganosa ao emitir aprovação sem SMTP configurado**~~ — **CORRIGIDO.**
3. ~~**[Baixo] Coluna "Last Reminder" na Central de Aprovações não atualiza após "Resend"**~~ — **CORRIGIDO.**
4. **[Gap de escopo] "Conta de emergência" do Sprint 5 não implementada.** Não corrigido — é uma feature ausente, não um bug pontual; precisa de definição de escopo antes de implementar (ver seção "Fora de escopo").
5. **[Observação] Log de auditoria do portal do cliente (tentativas de login) não confirmado.** Não é um defeito localizado — mantido como observação para investigação futura.

Todos os critérios de aceite de **workflow de contrato/aprovação** (Sprint 6) foram comprovados funcionando **de ponta a ponta, incluindo pela interface administrativa** após as correções abaixo.

---

## Correções aplicadas (2026-08-06, pós-validação)

### 1. `models/contract.go` — `GetContractsForTenant` sem preload
Adicionado `.Preload("Versions").Preload("Approvers")`, igual ao que já existia em `GetContractForTenant`. A listagem usada pela tela de Contratos agora traz versões e aprovadores, populando o modal e exibindo o botão "Request Approval".

### 2. `controllers/api/approval.go` — sucesso enganoso ao (re)emitir aprovação
`IssueApproval` e `ApprovalResend` agora contam quantos aprovadores foram efetivamente notificados (`approvers_notified`/`approvers_total` na resposta, via novo tipo `approvalNotifyResponse`) em vez de sempre reportar sucesso. `static/js/src/app/contracts.js` e `static/js/src/app/approvals.js` foram ajustados para mostrar um aviso quando `approvers_notified < approvers_total`. JS minificado recompilado via `npx gulp scripts`.

### 3. `controllers/api/approval.go` — "Last Reminder" nunca era gravado no resend manual
`ApprovalResend` não chamava `models.MarkReminderSent`, diferente do cron de lembrete (`approvals/scheduler.go`). Adicionada a chamada após o loop de notificação, alinhando o comportamento do botão "Resend" com o job automático.

**Validação pós-fix:** `go build ./...`, `go vet ./...` e `go test ./models/... ./controllers/...` passam limpos; imagem Docker reconstruída (`docker compose build server db-migrate` + `up -d --force-recreate server`) e as 3 correções foram reconfirmadas ao vivo via API e browser — `screenshots/26-fix-contracts-modal-populated.png` (modal agora populado, botões "Request Approval" visíveis) e `screenshots/27-fix-approvals-last-reminder.png` (coluna "Last Reminder" preenchida após um resend real).

### Fora de escopo desta correção
- **Conta de emergência (Sprint 5):** feature ausente, não um bug — requer decisão de produto (quem é a conta, como é provisionada, como fica de fora do RBAC normal) antes de implementar.
- **Auditoria de login do portal do cliente:** nenhuma linha de código quebrada foi localizada — é uma lacuna de observabilidade a investigar, não uma correção pontual.
