# Issues conhecidos — EthPhish v0.3.0

## Limitações técnicas

| Item | Impacto | Mitigação/planejamento |
| --- | --- | --- |
| Workers ainda são internos ao servidor | RabbitMQ já está no caminho crítico de e-mail, mas o consumidor é um pool de goroutines no processo `server`; não há escala independente nem isolamento por node | extrair worker node externo, AMQP TLS 5671, sem acesso a banco/painel |
| SMS e relatórios não usam a fila durável | permanecem no polling de banco já existente; não têm o retry/DLQ que o e-mail ganhou nesta release | estender o padrão `mail.send` para SMS e geração de relatórios |
| Sem transactional outbox | a publicação na fila não é atômica com o commit do `MailLog`; uma falha entre os dois pode duplicar ou perder a publicação (o retry SMTP por backoff mitiga perda, não duplicação) | implementar outbox transacional antes de tratar o e-mail como exactly-once |
| Multitenancy ainda não é auto-serviço | RLS e escopo por tenant existem e são testados, mas não há portal de cliente, fluxo de onboarding ou aprovação de campanha por tenant fora do admin interno | portal de clientes e fluxo auditável de aprovação em release futura |
| RLS depende de disciplina de configuração | qualquer DSN de runtime apontando para o role privilegiado `ethphish` (em vez de `ethphish_app`) desativa a proteção de RLS silenciosamente, pois superusuário ignora `FORCE ROW LEVEL SECURITY` | validar a DSN de runtime em cada deploy; considerar um check de partida que rejeite conexão como role privilegiado |
| Sessões sem tenant enxergam todas as linhas | intencional — preserva workers legados (monitor IMAP externo, drenagem da fila de relatórios, limpeza agendada) que ainda não migraram para `TenantScope` | migrar os processos remanescentes para sessões com `ethphish.tenant_id` definido, ou documentar formalmente cada exceção |
| Admin UI publicada em porta própria (9444) | reduz risco de path-based leakage, mas ainda é publicada no host pelo Compose de desenvolvimento | em produção, restringir 9444 a rede administrativa/VPN via firewall/security group, nunca expor na internet aberta |
| Reconciliação ampliada de importação | o importador aprovado compara contagens e preservação de IDs; hash de conteúdo, órfãos e equivalência de timestamps ainda não são gate automático | concluir ETH-307 antes de importar dados reais |
| SQLite em ferramenta/testes legados | o runtime e a imagem não incluem SQLite; o driver permanece somente no importador e na caracterização | remover após aposentadoria das bases legadas |
| GORM v1 e frontend legado | dependências com manutenção limitada | modernização incremental após estabilização funcional |
| Certificados autoassinados | adequados apenas para desenvolvimento | usar CA corporativa ou certificados públicos gerenciados em produção |
| Origens confiáveis customizadas no CSRF | não são suportadas nesta release por vulnerabilidade sem correção do componente upstream | usar um único origin administrativo HTTPS; reavaliar após correção upstream |
| Restore automatizado não existe | recuperação requer runbook e operador autorizado | automatizar retenção, restore periódico e evidência de teste |
| Dependências legadas fora do escopo do scanner | componentes herdados como GORM v1 exigem acompanhamento contínuo | CI bloqueia CVEs HIGH/CRITICAL corrigíveis em cada alteração de imagem |

## Restrições de segurança e produto

- Não usar com destinatários, domínios ou provedores sem autorização formal.
- Não registrar credenciais reais; a adequação completa dos fluxos legados será
  verificada antes de piloto.
- Não expor o painel administrativo (porta 9444) na internet; restringir a
  rede administrativa/VPN mesmo estando disponível via proxy reverso.
- Não habilitar ignorar erros de certificado SMTP/IMAP fora de teste autorizado.
- Não tratar a presença de RabbitMQ e RLS como arquitetura multitenant
  distribuída e self-service já concluída — permanece uso interno,
  administrado pelo time de segurança.
- Não conectar o runtime do `server` com o role privilegiado de migrations
  (`ethphish`); use sempre o role restrito (`ethphish_app`) para preservar o
  isolamento por RLS.

## Suporte

Reporte defeitos conforme [SECURITY.md](SECURITY.md). Não inclua dados pessoais,
credenciais, conteúdo de campanhas ou detalhes de alvos em issues públicas.
