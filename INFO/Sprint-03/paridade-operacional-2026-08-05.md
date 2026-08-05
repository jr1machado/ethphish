# Paridade operacional PostgreSQL

- Commit validado: `bd38f4b`
- Workflow aprovado: GitHub Actions `31048677388`
- Dados e destinos externos: nenhum; todos os registros são sintéticos.

## Cobertura aprovada

- campanhas, resultados, templates, landing pages, SMTP e SMS;
- IMAP persistido e lido sem abrir conexão de rede;
- segredo cifrado persistido, legível somente com a chave correta e recusado
  com chave incorreta;
- webhooks persistidos e consultados sem entrega HTTP;
- relatórios persistidos com transição `queued → processing → completed`.

O workflow também aprovou migrations PostgreSQL, qualidade, vulnerabilidades,
segredos, build/scan de container e SBOM.
