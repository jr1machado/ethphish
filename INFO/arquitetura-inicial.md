Perfeito. A decisão-base passa a ser:

> **Anglerphish 1.3.0 como upstream funcional, nosso fork corporativo como produto, PostgreSQL como banco principal e workers distribuídos para execução das entregas.**

A escolha reduz bastante o desenvolvimento inicial porque o Anglerphish já traz OIDC, criptografia de campos sensíveis, SMS, campanhas genéricas, QR Code, campanhas em lote, múltiplos perfis IMAP, relatórios e melhorias de interface. A versão `1.3.0` aparece atualmente marcada como **Latest**, com publicação em 27 de julho. ([GitHub][1])

## Pequena correção na avaliação

A maior parte da análise apresentada está confirmada pelo repositório. Entretanto, no estado consultado agora, a página de segurança informa que **não existe um `SECURITY.md` detectado** e não há advisories publicados. Portanto, o problema atual é ausência de uma política de segurança, e não somente uma política apontando para o mantenedor incorreto. ([GitHub][2])

## Arquitetura atualizada

```text
                         REDE ADMINISTRATIVA
                                  │
                         VPN / Zero Trust / mTLS
                                  │
                    ┌─────────────▼──────────────┐
                    │ Anglerphish Control Plane  │
                    │                            │
                    │ Administração e OIDC       │
                    │ Campanhas                  │
                    │ Templates                  │
                    │ Landing pages              │
                    │ SMS / QR / IMAP            │
                    │ Treinamento                │
                    │ Analytics e relatórios     │
                    │ Scheduler de entregas      │
                    └─────────────┬──────────────┘
                                  │
                   ┌──────────────┼───────────────┐
                   │              │               │
             PostgreSQL       RabbitMQ       Object Storage
                   │              │
                   │       Jobs de entrega
                   │              │
                   │      ┌───────┴────────┐
                   │      │                │
               ┌───▼──────▼──┐       ┌────▼─────────┐
               │ Worker 1    │       │ Worker 2     │
               │ Rede A      │       │ Rede B       │
               │ SMTP Relay A│       │ SMTP Relay B │
               └─────────────┘       └──────────────┘
```

## O que muda em relação ao plano anterior

### Recursos que deixam de precisar ser construídos do zero

O Anglerphish já oferece:

* OIDC para o painel administrativo;
* criptografia AES-256-GCM de campos sensíveis;
* campanhas por e-mail, SMS e canais genéricos;
* Twilio e Vonage;
* múltiplas configurações IMAP;
* QR Codes;
* conjuntos de campanhas;
* reenvio de falhas;
* variáveis globais;
* relatórios Word e Excel;
* anonimização de relatórios;
* preview de templates e landing pages;
* documentação da API na interface. ([GitHub][3])

### Recursos que continuam sendo nossos

Ainda teremos de desenvolver:

1. PostgreSQL como banco oficialmente suportado.
2. Arquitetura de workers distribuídos.
3. Broker e transactional outbox.
4. Exportação e importação de bundles completos.
5. Módulo estruturado de treinamento.
6. Classificação de scanners e falsos positivos.
7. Dashboard operacional.
8. Dashboard executivo.
9. Auditoria ampliada e RBAC.
10. Hardening e nova cadeia DevSecOps.

## Repositório e estratégia de fork

Eu estruturaria os branches assim:

```text
upstream/master
    Código original do Anglerphish

vendor/anglerphish-1.3
    Espelho controlado da versão upstream

main
    Nosso produto estável

develop
    Integração das funcionalidades

feature/postgresql
feature/distributed-workers
feature/training
feature/export-import
feature/analytics
feature/ui-modernization
```

E usaria um nome interno diferente para evitar confundir o produto corporativo com o upstream:

```text
anglerphish upstream
        ↓
corporate-angler control server
corporate-angler worker
corporate-angler migration tool
```

Isso facilita rastrear:

* código herdado;
* alterações próprias;
* correções trazidas do upstream;
* patches de segurança;
* conflitos de merge;
* divergência arquitetural.

## PostgreSQL continua sendo a primeira grande alteração

O Anglerphish ainda carrega GORM 1.9.12, Goose antigo, SQLite e drivers legados. O `go.mod` declara Go 1.24 e inclui várias dependências herdadas, incluindo Gorilla Mux 1.7.3, Logrus 1.4.2 e GORM 1.9.12. O driver `lib/pq` aparece como dependência indireta, mas isso, isoladamente, não significa que o PostgreSQL esteja implementado e testado como banco suportado. ([GitHub][4])

A primeira release própria deveria conter:

```text
Release 0.1 — Foundation
├── PostgreSQL
├── migrations PostgreSQL
├── configuração por ambiente
├── encryption key obrigatória
├── OIDC
├── Dockerfile endurecido
├── CI/CD DevSecOps
├── backup e restore
└── testes de regressão
```

Eu não atualizaria simultaneamente GORM, banco, frontend e arquitetura de workers. A sequência mais segura seria:

```text
1. Congelar Anglerphish 1.3.0
2. Criar testes de caracterização
3. Implementar PostgreSQL
4. Validar paridade funcional
5. Criar transactional outbox
6. Extrair o worker
7. Atualizar dependências gradualmente
```

## Dockerfile precisa ser substituído

O Dockerfile atual:

* utiliza `node:latest`;
* executa `npm install`;
* utilizava `golang:1.24` sem digest; a imagem atual fixa Go 1.25.12.
* executa `go get` durante o build;
* troca todos os endereços `127.0.0.1` por `0.0.0.0`;
* expõe `3333`, `8080`, `8443` e `80`;
* roda como usuário não root, o que é positivo. 

O trecho que altera os endereços é especialmente relevante:

```dockerfile
RUN sed -i 's/127.0.0.1/0.0.0.0/g' config.json
EXPOSE 3333 8080 8443 80
```

No nosso fork, isso deve ser substituído por configuração explícita:

```text
ANGLER_ADMIN_LISTEN=127.0.0.1:3333
ANGLER_CAMPAIGN_LISTEN=0.0.0.0:8080
ANGLER_TRAINING_LISTEN=0.0.0.0:8081
ANGLER_METRICS_LISTEN=127.0.0.1:9090
```

Em Docker, como `127.0.0.1` dentro do container não é acessível diretamente por outro container, uma alternativa operacional é ouvir em `0.0.0.0` internamente, mas **não publicar a porta no host**, permitindo acesso apenas pelo reverse proxy em uma rede administrativa interna:

```yaml
angler-control:
  expose:
    - "3333"
  # sem ports para 3333
  networks:
    - admin_internal
    - application_internal
```

Somente o proxy administrativo participaria de `admin_internal`.

## Workers com Anglerphish

Os workers não deveriam executar uma cópia completa do Anglerphish. Vamos extrair somente o mecanismo necessário para as entregas:

```text
cmd/
├── angler-control/
├── angler-worker/
└── angler-migrate/

internal/
├── delivery/
├── workers/
├── scheduler/
├── messaging/
├── smtp/
├── sms/
├── campaigns/
├── training/
├── analytics/
└── audit/
```

O worker poderá futuramente suportar mais de um executor:

```text
angler-worker --capability smtp
angler-worker --capability sms
angler-worker --capability generic
```

Inicialmente, eu habilitaria apenas SMTP:

```yaml
worker:
  id: worker-rede-a-01
  pool: rede-a
  capabilities:
    - smtp
  concurrency: 3
  max_messages_per_minute: 20
  smtp_profile: relay-a
```

O recurso de reenvio já existente no Anglerphish precisa ser adaptado para não disputar com o mecanismo de retry do broker. A regra deve ser:

```text
Retry automático:
controlado pelo worker e broker

Reenvio manual:
gera um novo delivery job auditado

Nunca:
reenviar diretamente pelo servidor central
```

## Impacto no hardware proposto

A adoção do Anglerphish não inviabiliza a configuração econômica:

```text
Central:
2 vCPU
4 GB RAM
100 GB SSD

Worker 1:
1 vCPU
1 GB RAM
50 GB SSD

Worker 2:
1 vCPU
1 GB RAM
50 GB SSD
```

Porém, o central ficará mais pressionado porque o Anglerphish adiciona:

* criptografia e descriptografia;
* processamento de relatórios Word e Excel;
* monitoramento IMAP;
* OIDC;
* SMS;
* QR Code;
* mais estruturas de campanha;
* novas telas e consultas.

A configuração ainda serve para um **MVP controlado**, desde que:

* relatórios sejam gerados em background, um por vez;
* IMAP tenha intervalos conservadores;
* Prometheus, Loki e Grafana não estejam todos no mesmo host;
* o PostgreSQL tenha memória limitada;
* RabbitMQ tenha limite de memória;
* anexos e exportações fiquem fora do banco;
* campanhas sejam escalonadas;
* não existam muitas operações administrativas simultâneas.

### Distribuição de memória no central de 4 GB

```text
Sistema operacional e Docker:  600–750 MB
PostgreSQL:                    900–1.100 MB
RabbitMQ:                      450–600 MB
Anglerphish:                   500–900 MB
Reverse proxy/exporters:       150–250 MB
Margem:                        400–800 MB
```

O primeiro upgrade continuará sendo:

```text
Central: 4 GB → 8 GB RAM
```

Os workers de 1 GB podem permanecer dessa forma por muito mais tempo.

## CI/CD própria é obrigatória

O workflow atual executa somente:

* instalação do Go;
* checkout;
* download de dependências;
* build;
* verificação de formatação;
* testes. ([GitHub][5])

Nossa pipeline deve adicionar:

```text
Build e qualidade
├── gofmt
├── go vet
├── staticcheck
├── testes unitários
├── testes de integração PostgreSQL
└── testes de migrations

Segurança
├── govulncheck
├── CodeQL
├── Gitleaks
├── Trivy filesystem
├── Trivy image
├── npm audit
├── osv-scanner
└── análise de licenças

Supply chain
├── dependências travadas
├── imagens por digest
├── Actions por SHA
├── SBOM CycloneDX/SPDX
├── assinatura Cosign
├── provenance/SLSA
└── hashes dos artefatos
```

## Roadmap revisado

### Etapa 1 — Fundação segura

* congelar Anglerphish 1.3.0;
* criar nosso repositório;
* substituir Dockerfile;
* ativar obrigatoriamente a chave de criptografia;
* configurar OIDC;
* implantar CI/CD;
* corrigir política de segurança;
* testes de caracterização.

### Etapa 2 — PostgreSQL

* implementar driver;
* migrations;
* pool;
* ferramenta SQLite → PostgreSQL;
* backup e restore;
* testes de integridade.

### Etapa 3 — Workers

* transactional outbox;
* RabbitMQ;
* registro e pools de workers;
* worker SMTP;
* heartbeat;
* idempotência;
* retry e dead-letter queue.

### Etapa 4 — Conteúdo e treinamento

* exportação/importação;
* bundles versionados;
* treinamento pós-interação;
* quiz;
* progresso;
* relatórios.

### Etapa 5 — Analytics

* classificação de eventos;
* separação entre scanners e interações prováveis;
* dashboard operacional;
* dashboard executivo;
* retenção e anonimização.

## Decisão consolidada

A arquitetura final fica:

```text
Anglerphish 1.3.0
        ↓
Nosso fork corporativo endurecido
        ↓
PostgreSQL + RabbitMQ
        ↓
Control plane central
        ↓
Workers SMTP distribuídos
        ↓
Treinamento + analytics + dashboards
```

É uma base melhor que o GoPhish original porque já incorpora diversas funcionalidades que estavam no nosso roadmap. Ao mesmo tempo, o Anglerphish deve ser tratado como **código-fonte upstream**, não como uma imagem pronta para produção. O principal cuidado agora é preservar seus recursos enquanto modernizamos banco, pipeline, containerização e execução distribuída.

[1]: https://github.com/geopetro/anglerphish/releases/tag/v1.3.0 "Release Anglerphish v1.3.0 · geopetro/anglerphish · GitHub"
[2]: https://github.com/geopetro/anglerphish/security "Overview · geopetro/anglerphish · GitHub"
[3]: https://github.com/geopetro/anglerphish/blob/master/FEATURES.md "anglerphish/FEATURES.md at master · geopetro/anglerphish · GitHub"
[4]: https://github.com/geopetro/anglerphish/blob/master/go.mod "anglerphish/go.mod at master · geopetro/anglerphish · GitHub"
[5]: https://github.com/geopetro/anglerphish/blob/master/.github/workflows/ci.yml "anglerphish/.github/workflows/ci.yml at master · geopetro/anglerphish · GitHub"
