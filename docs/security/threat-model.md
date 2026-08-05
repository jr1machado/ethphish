# Threat model inicial

## Ativos

Dados de tenants, escopos autorizados, participantes, eventos, templates,
credenciais de providers, chaves, relatórios e trilhas de auditoria.

## Fronteiras de confiança

Painel administrativo, portal do cliente, superfície pública de campanhas,
control plane, banco, broker, object storage, workers e providers externos são
fronteiras distintas.

## Ameaças prioritárias e controles

| Ameaça | Controle inicial |
| --- | --- |
| Painel administrativo público | rede privada e ausência de porta publicada |
| Vazamento entre tenants | tenant obrigatório, RLS e testes negativos |
| Campanha fora do escopo | gate de autorização e allowlist de domínios |
| Credencial real persistida | descarte na entrada e testes de regressão |
| Job duplicado | chave idempotente, outbox e confirmação |
| Worker comprometido | identidade própria, escopo mínimo e jobs assinados |
| SSRF por webhook/importação | dialer restrito e validação de destino |
| Segredo em log ou repositório | masking, secret scanning e secret store |
| Supply-chain comprometida | versões fixadas, scans, SBOM e assinatura futura |

Este documento é inicial e deve evoluir para diagramas de fluxo de dados e uma
análise STRIDE antes do piloto.
