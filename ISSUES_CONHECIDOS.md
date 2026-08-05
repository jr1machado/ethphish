# Issues conhecidos — v0.1.0

## Limitações técnicas

| Item | Impacto | Mitigação/planejamento |
| --- | --- | --- |
| Workers são internos ao servidor | não há escala independente nem isolamento por node | transactional outbox e workers AMQP nas próximas releases |
| RabbitMQ ainda não processa jobs | broker está provisionado, mas não integra entrega | implementar filas, idempotência, retry e DLQ |
| Multitenancy ainda não existe | não é adequado para portal de clientes ou dados compartilhados | tenant obrigatório, RLS e testes negativos antes de uso multicliente |
| SQLite segue no código/testes | CGO ainda é requerido em parte da suíte | remover do runtime somente após paridade PostgreSQL |
| GORM v1 e frontend legado | dependências com manutenção limitada | modernização incremental após estabilização funcional |
| Certificados autoassinados | adequados apenas para desenvolvimento | usar CA corporativa ou certificados públicos gerenciados em produção |
| Restore automatizado não existe | recuperação requer runbook e operador autorizado | automatizar retenção, restore periódico e evidência de teste |
| CI ainda não está publicado | token GitHub atual não possui escopo `workflow` | executar `gh auth refresh -h github.com -s workflow` e publicar a branch |

## Restrições de segurança e produto

- Não usar com destinatários, domínios ou provedores sem autorização formal.
- Não registrar credenciais reais; a adequação completa dos fluxos legados será
  verificada antes de piloto.
- Não expor o painel administrativo na internet.
- Não habilitar ignorar erros de certificado SMTP/IMAP fora de teste autorizado.
- Não tratar a presença de RabbitMQ como arquitetura distribuída já concluída.

## Suporte

Reporte defeitos conforme [SECURITY.md](SECURITY.md). Não inclua dados pessoais,
credenciais, conteúdo de campanhas ou detalhes de alvos em issues públicas.
