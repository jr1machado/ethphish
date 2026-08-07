# Issues conhecidos — EthPhish v0.6.0

## Limitações técnicas

| Item | Impacto | Mitigação/planejamento |
| --- | --- | --- |
| **Fontes de autenticação carregadas de CDN externo (Google Fonts)** (novo na v0.6.0) | `static/css/ethphish-auth.css` importa Space Grotesk/JetBrains Mono/Source Sans Pro de `fonts.googleapis.com`; sem saída de rede para esse domínio (rede administrativa isolada, air-gap) as três telas de login caem para a fonte de fallback do SO, sem quebrar funcionalmente | vendorizar as fontes localmente (mesmo padrão já usado pra SheetJS) em release futura, se a operação exigir ambiente sem saída de internet |
| **Redesenho de autenticação não cobre `reset_password.html`** (novo na v0.6.0) | tela de troca de senha do admin ainda usa o Bootstrap/`form-signin` herdado, com visual inconsistente em relação às três telas já migradas | estender `ethphish-auth.css` a essa tela em release futura — decisão de corte, não limitação técnica |
| **Sem certificado de conclusão de treinamento** | assignments/quiz_attempts gravam nota e aprovação, mas não há PDF/documento emitido ao concluir | gerar certificado (reaproveitar o gerador de relatórios Word/Excel já existente) em release futura |
| **Sem dashboard de indicadores de treinamento** | taxa de início/conclusão, nota média, evolução entre tentativas, departamentos com menor adesão, reincidência pós-treinamento e impacto nas campanhas seguintes (Sprint08 §14.4) não têm tela — os dados brutos já existem em `training_assignments`/`quiz_attempts`/`training_lesson_views` | construir a camada de agregação e visualização em cima do que já é gravado |
| **Treinamento não aparece no portal do cliente** | `/portal/*` (Sprint 7.5) reserva o conceito na navegação mas não lista treinamentos do tenant nem progresso agregado | wiring simples, dado já existe; adiar foi decisão de escopo, não limitação técnica |
| **E-mail de atribuição de treinamento depende de perfil SMTP configurado no tenant** | sem perfil SMTP, a atribuição é criada no banco (link válido) mas ninguém recebe o e-mail; mesma limitação já documentada para aprovações na v0.4.0 | mesmo aviso preventivo pendente de implementar na tela de Contratos vale aqui |
| **Login self-service do portal depende de já existir um `ContractApprover` cadastrado** | não há cadastro próprio de cliente; o portal só reconhece quem já foi nomeado aprovador em algum contrato pelo admin | portal self-service completo (onboarding próprio) é item de release futura, já listado desde a v0.4.0 |
| **Sem conteúdo de vídeo/SCORM em treinamento** | lições são só HTML estático | avaliar necessidade real antes de acrescentar suporte a mídia rica |
| **Conta de emergência não implementada** (herdado da v0.4.0) | escopo do Sprint 5 previa break-glass administrativo se o IdP OIDC cair; ainda não existe, só o login local do `admin` cobre esse caso | definir escopo e implementar |
| **Auditoria de login do portal do cliente não confirmada** (herdado da v0.4.0) | tentativas de login em `/approvals/*`/`/portal/*` não têm log estruturado dedicado | revisar e acrescentar, sem logar token em texto claro |
| **Workers ainda são internos ao servidor** (herdado da v0.3.0) | RabbitMQ está no caminho crítico de e-mail (campanha, aprovação, treinamento), mas o consumidor é um pool de goroutines no processo `server` | extrair worker node externo, AMQP TLS 5671, sem acesso a banco/painel |
| **SMS e relatórios não usam a fila durável** (herdado da v0.3.0) | permanecem no polling de banco já existente | estender o padrão `mail.send` |
| **Sem transactional outbox** (herdado da v0.3.0) | publicação na fila não é atômica com o commit do registro correspondente | implementar outbox transacional |
| **`ETHPHISH_PHISH_CSRF_KEY` sem valor fixo em produção** (herdado da v0.4.0) | sem ela, um restart do `server` invalida formulários de decisão de aprovação e de quiz em andamento | definir a variável com um valor fixo e gerenciado antes de publicar externamente |
| **RLS não cobre as tabelas de contrato/aprovação/portal/treinamento** | as tabelas do Sprint 5–7 (`contracts`, `client_users`, `portal_login_tokens`, `trainings`, `training_assignments` etc.) usam `tenant_id` + escopo em nível de aplicação (`withTenantTransaction`), mas não têm a policy PostgreSQL RLS aplicada na migration original — mesmo padrão desde a v0.4.0, não é regressão desta release, mas segue sem cobertura de RLS | avaliar estender a policy `tenant_isolation` pra essas tabelas |
| **Admin UI publicada em porta própria (9444)** (herdado da v0.3.0) | 9443 agora carrega também portal e treinamento do cliente | restringir ambas as portas conforme o caso de uso em produção |
| **Certificados autoassinados** (herdado da v0.3.0) | adequados só para desenvolvimento | usar CA corporativa ou certificados públicos gerenciados em produção |
| **Origens confiáveis customizadas no CSRF** (herdado da v0.3.0) | não suportadas nesta release por vulnerabilidade sem correção do componente upstream | usar um único origin administrativo HTTPS |
| **Restore automatizado não existe** (herdado da v0.3.0) | recuperação requer runbook e operador autorizado | automatizar retenção, restore periódico e evidência de teste |
| **Dependências legadas fora do escopo do scanner** (herdado da v0.3.0) | GORM v1 e afins exigem acompanhamento contínuo | CI bloqueia CVEs HIGH/CRITICAL corrigíveis em cada alteração de imagem |

## Restrições de segurança e produto

- Não usar com destinatários, domínios ou provedores sem autorização formal.
- Não registrar credenciais reais; a adequação completa dos fluxos legados será
  verificada antes de piloto.
- Não expor o painel administrativo (porta 9444) na internet; a porta 9443
  carrega o portal de aprovação, o portal completo do cliente e a entrega
  de treinamento — restringir conforme o público real de cada campanha.
- Não habilitar ignorar erros de certificado SMTP/IMAP fora de teste autorizado.
- Não tratar o workflow de contrato/aprovação como substituto de um processo
  jurídico formal de contratação.
- Não conectar o runtime do `server` com o role privilegiado de migrations
  (`ethphish`); use sempre o role restrito (`ethphish_app`).
- Não publicar o portal do cliente ou a entrega de treinamento sem
  `ETHPHISH_PHISH_CSRF_KEY` fixa e sem ao menos um perfil SMTP por tenant.
- Não tratar a nota/aprovação de quiz como avaliação de desempenho
  individual formal sem revisão jurídica/RH prévia — é evidência de
  conscientização, não instrumento disciplinar por padrão do produto.

## Suporte

Reporte defeitos conforme [SECURITY.md](SECURITY.md). Não inclua dados pessoais,
credenciais, conteúdo de campanhas ou detalhes de alvos em issues públicas.
