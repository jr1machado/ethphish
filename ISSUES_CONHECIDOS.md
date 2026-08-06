# Issues conhecidos — EthPhish v0.4.0

## Limitações técnicas

| Item | Impacto | Mitigação/planejamento |
| --- | --- | --- |
| **Conta de emergência não implementada** | O escopo do Sprint 5 previa uma conta de emergência (break-glass) para acesso administrativo caso o IdP OIDC fique indisponível; não existe no código nem foi encontrada em nenhum commit — apenas o login local do usuário `admin` cobre esse caso hoje, sem processo formal de custódia/rotação | definir escopo (quem, como é provisionada, como fica fora do fluxo OIDC/RBAC normal) e implementar em release futura |
| **Entrega de e-mail de aprovação depende de perfil SMTP configurado no tenant** | Sem um perfil SMTP, "Request Approval"/"Resend" criam o registro no banco mas ninguém é notificado; a UI agora reporta `approvers_notified`/`approvers_total` (corrigido nesta release), mas ainda não há um alerta proativo na tela de Contratos quando o tenant não tem SMTP configurado | adicionar aviso preventivo na criação do contrato/aprovador se o tenant não tiver perfil SMTP |
| **Auditoria de login do portal do cliente não confirmada** | Tentativas de login (sucesso/falha) do portal `/approvals/login` não foram localizadas em log estruturado dedicado durante a validação manual — diferente do login administrativo, que aparece no log do `server` | revisar `controllers/approval_portal.go` e acrescentar log estruturado de tentativa/sucesso/falha, sem registrar o token em texto claro |
| **Workers ainda são internos ao servidor** (herdado da v0.3.0) | RabbitMQ está no caminho crítico de e-mail, mas o consumidor é um pool de goroutines no processo `server`; não há escala independente nem isolamento por node | extrair worker node externo, AMQP TLS 5671, sem acesso a banco/painel |
| **SMS e relatórios não usam a fila durável** (herdado da v0.3.0) | permanecem no polling de banco já existente; não têm o retry/DLQ que o e-mail (de campanha e de aprovação) já tem | estender o padrão `mail.send` para SMS e geração de relatórios |
| **Sem transactional outbox** (herdado da v0.3.0) | a publicação na fila não é atômica com o commit do `MailLog`; uma falha entre os dois pode duplicar ou perder a publicação | implementar outbox transacional antes de tratar o e-mail como exactly-once |
| **Portal do cliente ainda não é onboarding self-service completo** | o cliente só entra via magic link emitido pelo admin a partir de um aprovador já cadastrado no contrato; não há cadastro próprio, recuperação de acesso alternativa ou gestão de múltiplos contratos numa única sessão além dos que a sessão do aprovador já autoriza | portal de clientes com onboarding próprio em release futura |
| **RLS depende de disciplina de configuração** (herdado da v0.3.0) | qualquer DSN de runtime apontando para o role privilegiado `ethphish` (em vez de `ethphish_app`) desativa a proteção de RLS silenciosamente | validar a DSN de runtime em cada deploy; considerar um check de partida que rejeite conexão como role privilegiado |
| **`ETHPHISH_PHISH_CSRF_KEY` sem valor fixo em produção** | se não definida, uma chave é gerada por processo — um restart do `server` invalida qualquer formulário de decisão de aprovação que o cliente tenha aberto (a sessão de login continua válida, só o POST de decisão falha por CSRF) | definir a variável com um valor fixo e gerenciado (secret) antes de publicar o portal externamente |
| **Import XLSX no cliente sem limite de tamanho aplicado no frontend** | a importação de grupos via XLSX roda inteiramente no navegador (SheetJS); arquivos muito grandes podem travar a aba antes de a validação/preview rodar | validar tamanho do arquivo antes do parse; considerar mover para o backend se o volume justificar |
| **Admin UI publicada em porta própria (9444)** (herdado da v0.3.0) | reduz risco de path-based leakage, mas ainda é publicada no host pelo Compose de desenvolvimento | em produção, restringir 9444 (e agora também 9443, que passou a servir o portal do cliente) a redes autorizadas via firewall/security group/VPN |
| **Reconciliação ampliada de importação** (herdado da v0.3.0) | o importador aprovado compara contagens e preservação de IDs; hash de conteúdo, órfãos e equivalência de timestamps ainda não são gate automático | concluir antes de importar dados reais |
| **SQLite em ferramenta/testes legados** (herdado da v0.3.0) | o runtime e a imagem não incluem SQLite; o driver permanece somente no importador e na caracterização | remover após aposentadoria das bases legadas |
| **GORM v1 e frontend legado** (herdado da v0.3.0) | dependências com manutenção limitada | modernização incremental após estabilização funcional |
| **Certificados autoassinados** (herdado da v0.3.0) | adequados apenas para desenvolvimento | usar CA corporativa ou certificados públicos gerenciados em produção |
| **Origens confiáveis customizadas no CSRF** (herdado da v0.3.0) | não são suportadas nesta release por vulnerabilidade sem correção do componente upstream | usar um único origin administrativo HTTPS; reavaliar após correção upstream |
| **Restore automatizado não existe** (herdado da v0.3.0) | recuperação requer runbook e operador autorizado | automatizar retenção, restore periódico e evidência de teste |
| **Dependências legadas fora do escopo do scanner** (herdado da v0.3.0) | componentes herdados como GORM v1 exigem acompanhamento contínuo | CI bloqueia CVEs HIGH/CRITICAL corrigíveis em cada alteração de imagem |

## Restrições de segurança e produto

- Não usar com destinatários, domínios ou provedores sem autorização formal.
- Não registrar credenciais reais; a adequação completa dos fluxos legados será
  verificada antes de piloto.
- Não expor o painel administrativo (porta 9444) na internet; a porta 9443
  agora também carrega o portal do cliente aprovador — restringir ambas à
  rede administrativa/VPN ou a uma lista de origens autorizadas conforme o
  caso de uso.
- Não habilitar ignorar erros de certificado SMTP/IMAP fora de teste autorizado.
- Não tratar o workflow de contrato/aprovação como substituto de um processo
  jurídico formal de contratação — ele é evidência operacional de que um
  responsável nomeado aprovou um escopo específico, não um instrumento
  contratual em si.
- Não conectar o runtime do `server` com o role privilegiado de migrations
  (`ethphish`); use sempre o role restrito (`ethphish_app`) para preservar o
  isolamento por RLS.
- Não publicar o portal do cliente sem `ETHPHISH_PHISH_CSRF_KEY` fixa e sem
  ao menos um perfil SMTP por tenant — sem isso, a experiência de aprovação
  fica quebrada mesmo com o backend funcionando corretamente.

## Suporte

Reporte defeitos conforme [SECURITY.md](SECURITY.md). Não inclua dados pessoais,
credenciais, conteúdo de campanhas ou detalhes de alvos em issues públicas.
