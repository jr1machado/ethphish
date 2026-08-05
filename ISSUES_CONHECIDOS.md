# Issues conhecidos — EthPhish v0.2.0

## Limitações técnicas

| Item | Impacto | Mitigação/planejamento |
| --- | --- | --- |
| Workers são internos ao servidor | não há escala independente nem isolamento por node | transactional outbox e workers AMQP nas próximas releases |
| RabbitMQ ainda não processa jobs | broker está provisionado, mas não integra entrega | implementar filas, idempotência, retry e DLQ |
| Multitenancy ainda não existe | não é adequado para portal de clientes ou dados compartilhados | tenant obrigatório, RLS e testes negativos antes de uso multicliente |
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
- Não expor o painel administrativo na internet.
- Não habilitar ignorar erros de certificado SMTP/IMAP fora de teste autorizado.
- Não tratar a presença de RabbitMQ como arquitetura distribuída já concluída.

## Suporte

Reporte defeitos conforme [SECURITY.md](SECURITY.md). Não inclua dados pessoais,
credenciais, conteúdo de campanhas ou detalhes de alvos em issues públicas.
