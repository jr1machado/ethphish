# EthPhish — Especificação Consolidada de Produto, Negócio e Arquitetura

# 1. Identidade do projeto

**Nome do produto:** EthPhish
**Base tecnológica:** fork corporativo do Anglerphish 1.3.0
**Categoria:** plataforma multitenant de conscientização, simulação autorizada de phishing e mensuração de risco humano
**Modelo operacional:** servidor central conteinerizado, PostgreSQL, broker de mensagens e workers distribuídos
**Público-alvo:** empresas, consultorias de segurança, equipes de GRC, SOC, segurança da informação, auditoria e conscientização

## 1.1 Controle do documento

| Campo | Definição |
| --- | --- |
| Documento | Especificação consolidada do EthPhish |
| Versão | 1.0 — baseline de planejamento |
| Data-base | 4 de agosto de 2026 |
| Status | Documento mestre para arquitetura, produto, segurança e roadmap |
| Base tecnológica | Fork corporativo do Anglerphish 1.3.0 |
| Público do documento | Produto, arquitetura, desenvolvimento, DevSecOps, segurança, operação, jurídico e clientes aprovadores |

Este documento consolida as premissas, objetivos, regras de negócio, requisitos funcionais e não funcionais, arquitetura-alvo, controles de segurança, critérios de aceite e planejamento por sprints do EthPhish.

## 1.2 Premissas consolidadas de baseline

- O EthPhish será derivado de um fork corporativo controlado do Anglerphish 1.3.0.
- O sistema será utilizado exclusivamente em campanhas éticas, autorizadas e contratualmente delimitadas.
- O servidor central será conteinerizado e atuará como control plane, fonte de verdade e interface de gestão.
- O PostgreSQL substituirá o SQLite como banco principal de produção.
- O RabbitMQ será utilizado para distribuição de jobs, retries e dead-letter queue, com transactional outbox no PostgreSQL.
- A entrega será executada por um ou mais workers distribuídos, sem acesso administrativo e sem acesso direto ao banco.
- O MVP poderá iniciar com servidor central de 2 vCPU, 4 GB de RAM e 100 GB de disco, e dois workers de 1 vCPU, 1 GB de RAM e 50 GB de disco cada.
- O primeiro upgrade esperado será a memória do servidor central, de 4 GB para 8 GB, conforme o crescimento de banco, filas, relatórios e observabilidade.
- O painel administrativo não será publicado diretamente na internet.
- Landing pages, treinamentos e demais conteúdos públicos operarão externamente por HTTPS na porta 443.
- O portal do cliente será multitenant e terá segregação aplicada no banco, backend, cache, storage, API e interface.
- Credenciais reais ou dados sensíveis digitados por participantes não serão armazenados.
- Contratos, escopos, templates, landing pages e treinamentos sujeitos à aprovação serão versionados e imutáveis.
- Nenhuma campanha será liberada sem a conclusão das aprovações obrigatórias e dos controles de escopo.
- Alterações destinadas a contornar ferramentas de proteção de e-mail não fazem parte dos objetivos do produto.

---

# 2. Visão do produto

O EthPhish será uma plataforma corporativa para planejar, executar, acompanhar e avaliar campanhas autorizadas de conscientização contra phishing.

A solução deverá combinar:

* campanhas por e-mail, SMS, QR Code e canais genéricos;
* entrega distribuída por workers;
* segmentação de participantes;
* portal multitenant para clientes;
* treinamentos de conscientização;
* indicadores operacionais e executivos;
* redução de falsos positivos;
* segregação segura de dados;
* governança, auditoria e privacidade;
* integração com provedores de e-mail;
* operação conteinerizada e observável.

O EthPhish não será apenas uma ferramenta de envio de campanhas. Ele deverá funcionar como uma plataforma de gestão do risco humano, permitindo acompanhar a evolução dos participantes, departamentos, empresas e tenants ao longo do tempo.

---

# 3. Missão

Ajudar organizações a reduzir riscos relacionados à engenharia social por meio de campanhas éticas, autorizadas, mensuráveis e integradas a programas contínuos de conscientização.

---

# 4. Objetivos estratégicos

## 4.1 Objetivos de produto

1. Transformar o Anglerphish em uma plataforma corporativa multitenant.
2. Oferecer campanhas, treinamentos, relatórios e dashboards em uma única solução.
3. Permitir que cada cliente acompanhe seus próprios resultados com segregação segura.
4. Disponibilizar entrega distribuída por múltiplos workers.
5. Integrar relays SMTP e provedores de envio por API.
6. Melhorar a qualidade das métricas e reduzir falsos positivos.
7. Permitir evolução gradual para alta disponibilidade.
8. Proteger o painel administrativo contra exposição externa.
9. Construir uma experiência moderna para administradores, operadores e clientes.
10. Tornar o EthPhish uma referência em conscientização contínua e mensuração de risco humano.

11. Permitir que clientes revisem e aprovem templates de e-mail e landing pages por meio do portal multitenant.

12. Implementar workflow formal de aprovação com confirmação por token enviado ao e-mail cadastrado do cliente.

13. Permitir que o cliente aprove eletronicamente o contrato de prestação de serviço e o escopo da campanha antes de sua execução.

14. Disponibilizar um fluxo para solicitação de novos templates de e-mail, landing pages e conteúdos de treinamento.

15. Manter evidências auditáveis de todas as aprovações, rejeições, solicitações e alterações de escopo.

## 4.2 Objetivos de segurança

1. Nunca expor diretamente o painel administrativo à internet.
2. Segregar completamente dados, usuários e campanhas entre tenants.
3. Aplicar criptografia em trânsito e em repouso.
4. Registrar ações administrativas em trilha de auditoria.
5. Impedir que workers executem campanhas ou destinos não autorizados.
6. Não armazenar senhas reais submetidas por participantes.
7. Permitir autenticação corporativa por OIDC.
8. Preparar a arquitetura para MFA com Microsoft Authenticator, Google Authenticator e demais aplicativos TOTP.
9. Implementar gestão segura de segredos.
10. Aplicar controles contra abuso interno ou comprometimento da plataforma.

## 4.3 Objetivos operacionais

1. Permitir expansão horizontal dos workers.
2. Monitorar central, banco, broker, aplicações e workers.
3. Possibilitar pause, cancelamento, reenvio e limitação de campanhas.
4. Manter histórico auditável dos jobs.
5. Evitar entregas duplicadas.
6. Implementar retries controlados e dead-letter queue.
7. Permitir backup e restauração documentados.
8. Disponibilizar health checks e indicadores de capacidade.
9. Implantar releases reproduzíveis e assinadas.
10. Manter o sistema operacional mesmo durante indisponibilidades temporárias dos workers.

---

# 5. Princípios do EthPhish

## 5.1 Uso ético e autorizado

Toda campanha deve estar associada a:

* tenant responsável;
* empresa-alvo autorizada;
* escopo aprovado;
* período autorizado;
* grupos de participantes;
* domínios permitidos;
* operador responsável;
* registro de aprovação.

A plataforma deverá impedir campanhas destinadas a domínios não autorizados pelo tenant.

## 5.2 Privacidade por padrão

O EthPhish deverá registrar comportamentos necessários para conscientização, sem armazenar credenciais reais.

Quando existir formulário de simulação, a plataforma poderá registrar:

* que houve preenchimento;
* quais tipos de campos foram utilizados;
* horário da interação;
* etapa alcançada;
* duração aproximada da interação.

O conteúdo real digitado em campos de senha, código MFA, cartão, documentos ou informações sensíveis deverá ser descartado ou substituído imediatamente por indicadores não reversíveis.

## 5.3 Separação entre planos

A arquitetura deverá possuir separação lógica e de rede entre:

* plano administrativo;
* plano de clientes;
* plano de campanhas e landing pages;
* plano de execução dos workers;
* plano de dados;
* plano de observabilidade.

## 5.4 Segurança antes de conveniência

Nenhuma funcionalidade deverá justificar:

* exposição pública do painel administrativo;
* compartilhamento de credenciais;
* ausência de auditoria;
* acesso direto de workers ao banco;
* armazenamento desnecessário de dados sensíveis;
* mistura de dados entre tenants.

## 5.5 Evolução modular

O EthPhish poderá permanecer inicialmente como um monólito modular, mas seus domínios deverão possuir fronteiras claras para permitir futura extração de serviços.

---

## 5.6 Aprovação explícita do cliente

Conforme a política do tenant, deverão ser aprovados:

* contrato de prestação de serviço;
* escopo da campanha;
* empresas e domínios autorizados;
* período de execução;
* quantidade ou grupos de participantes;
* canais de comunicação;
* template de e-mail;
* landing page;
* treinamento associado;
* remetente apresentado;
* domínios utilizados;
* eventuais exceções acordadas.

A aprovação deverá ser explícita, individual, auditável e vinculada à versão exata do conteúdo aprovado.

Qualquer alteração material posterior deverá invalidar a aprovação anterior e iniciar um novo ciclo de aprovação.

## 5.7 Não repúdio e rastreabilidade

Toda aprovação deverá gerar evidências suficientes para demonstrar:

* quem aprovou;
* qual tenant e empresa estavam envolvidos;
* qual item foi aprovado;
* qual versão foi aprovada;
* quando ocorreu a aprovação;
* qual endereço de e-mail recebeu o token;
* quando o token foi validado;
* qual endereço IP realizou a ação;
* qual user-agent foi utilizado;
* qual contrato e escopo estavam vigentes;
* qual hash representa o conteúdo aprovado.

O registro de aprovação não deverá depender apenas de uma marcação visual na interface.

---

# 6. Arquitetura-alvo

```text
                              REDE ADMINISTRATIVA
                                      │
                           VPN / Zero Trust / Bastion
                                      │
                       ┌──────────────▼───────────────┐
                       │ EthPhish Control Plane       │
                       │                              │
                       │ Administração                │
                       │ Campanhas                    │
                       │ Tenants e clientes           │
                       │ Treinamentos                 │
                       │ Analytics                    │
                       │ Scheduler                    │
                       │ Relatórios                   │
                       └──────────────┬───────────────┘
                                      │
            ┌─────────────────────────┼─────────────────────────┐
            │                         │                         │
     ┌──────▼───────┐         ┌──────▼───────┐         ┌──────▼───────┐
     │ PostgreSQL   │         │ RabbitMQ     │         │ Object Store │
     │              │         │              │         │              │
     │ Dados        │         │ Jobs         │         │ Assets       │
     │ Auditoria    │         │ Retries      │         │ Relatórios   │
     │ Eventos      │         │ DLQ          │         │ Exportações  │
     └──────────────┘         └──────┬───────┘         └──────────────┘
                                     │
                         ┌───────────┴───────────┐
                         │                       │
                  ┌──────▼──────┐         ┌──────▼──────┐
                  │ Worker A    │         │ Worker B    │
                  │ Rede A      │         │ Rede B      │
                  │ SMTP/API    │         │ SMTP/API    │
                  └─────────────┘         └─────────────┘
```

---

# 7. Componentes principais

## 7.1 EthPhish Control Server

Responsável por:

* gerenciamento de tenants;
* autenticação e autorização;
* cadastro de usuários;
* campanhas e templates;
* landing pages;
* treinamentos;
* agendamentos;
* distribuição de jobs;
* dashboards;
* auditoria;
* relatórios;
* configuração dos workers;
* configuração dos provedores de entrega.

## 7.2 PostgreSQL

Será a fonte primária de verdade da plataforma.

Deverá armazenar:

* tenants;
* usuários;
* empresas;
* campanhas;
* destinatários;
* eventos;
* treinamentos;
* progresso;
* auditoria;
* workers;
* provedores;
* políticas;
* configurações.

O SQLite não deverá ser utilizado em produção.

## 7.3 RabbitMQ

Responsável por:

* filas de entrega;
* filas por worker pool;
* retries;
* jobs expirados;
* dead-letter queue;
* confirmação de processamento;
* distribuição de carga.

O PostgreSQL, por meio de transactional outbox, continuará sendo a fonte durável dos jobs.

## 7.4 Workers

Os workers deverão ser pequenos, stateless e substituíveis.

Cada worker poderá possuir uma ou mais capacidades:

* SMTP relay;
* SMTP autenticado;
* Brevo API;
* SendGrid API;
* SMS;
* canal genérico.

Os workers não terão acesso administrativo e não deverão consultar diretamente o PostgreSQL.

## 7.5 Portal do cliente

Área autenticada destinada aos clientes dos tenants.

Deverá permitir:

* consultar campanhas próprias;
* acompanhar indicadores;
* visualizar treinamentos;
* exportar relatórios autorizados;
* acompanhar evolução histórica;
* analisar departamentos e unidades;
* consultar status de campanhas em andamento.

## 7.6 Landing pages e treinamentos

A camada pública de campanhas deverá operar externamente exclusivamente por HTTPS na porta 443.

Ela será responsável por:

* landing pages;
* páginas educativas;
* treinamentos;
* quizzes;
* certificados, quando aplicável;
* tracking de progresso;
* coleta de eventos autorizados.

---

# 8. Modelo multitenant

## 8.1 Estrutura

```text
Tenant
├── Empresas
│   ├── Departamentos
│   ├── Usuários
│   ├── Campanhas
│   ├── Grupos
│   └── Treinamentos
├── Operadores
├── Clientes
├── Workers autorizados
├── Provedores de envio
└── Políticas
```

Um tenant poderá administrar uma ou mais empresas, conforme seu plano e permissões.

## 8.2 Estratégia inicial

A primeira versão deverá utilizar:

* banco PostgreSQL compartilhado;
* tabelas compartilhadas;
* coluna `tenant_id` obrigatória;
* Row-Level Security do PostgreSQL;
* validação de tenant na aplicação;
* testes automatizados contra vazamento entre tenants;
* cache sempre particionado por tenant;
* object storage particionado por tenant;
* logs com identificadores de tenant, sem dados sensíveis.

## 8.3 Regras obrigatórias

Toda entidade de negócio deverá possuir `tenant_id`, direta ou indiretamente.

Nenhuma consulta deverá depender apenas de um identificador global:

```text
Incorreto:
GET /campaigns/123

Correto:
tenant autenticado + campanha 123 pertencente ao tenant
```

A aplicação deverá negar por padrão qualquer operação em que o tenant não possa ser determinado.

## 8.4 Perfis de acesso

### Superadministrador da plataforma

* gerencia infraestrutura;
* cadastra tenants;
* gerencia planos;
* acompanha saúde da plataforma;
* não acessa automaticamente conteúdo sensível de campanhas.

### Administrador do tenant

* gerencia empresas;
* gerencia operadores;
* configura integrações;
* administra campanhas e treinamentos.

### Operador de campanhas

* cria e executa campanhas autorizadas;
* não altera configurações críticas do tenant.

### Gestor de conscientização

* gerencia treinamentos;
* acompanha evolução;
* acessa relatórios educativos.

### Cliente executivo

* visualiza indicadores agregados;
* não visualiza dados sensíveis individuais.

### Auditor

* consulta trilhas de auditoria;
* possui acesso somente leitura.

### Usuário de suporte

* possui acesso temporário;
* acesso concedido mediante aprovação;
* toda ação registrada;
* sem acesso padrão a dados capturados.

---

## 8.5 Responsáveis por aprovação

Os responsáveis poderão possuir escopos distintos:

```text
Aprovador contratual
Aprovador da campanha
Aprovador de comunicação
Aprovador de landing page
Aprovador de treinamento
Aprovador executivo
```

O tenant poderá exigir:

* uma aprovação simples;
* aprovações sequenciais;
* aprovações paralelas;
* aprovação de mais de uma pessoa;
* aprovação obrigatória de áreas distintas;
* substituição temporária de aprovador;
* prazo máximo para aprovação.

Exemplo:

```text
Contrato e escopo
        ↓
Responsável contratual
        ↓
Template de e-mail
        ↓
Comunicação ou marketing
        ↓
Landing page e treinamento
        ↓
Segurança da informação
        ↓
Liberação da campanha
```

## 8.6 Segregação das aprovações

Um aprovador deverá acessar somente:

* o tenant ao qual pertence;
* as empresas para as quais possui autorização;
* as campanhas incluídas em seu escopo;
* os itens que exigem sua decisão;
* as versões disponibilizadas para revisão.

Tokens, links e identificadores de aprovação deverão estar vinculados ao tenant, ao usuário, ao item e à versão correspondente.

Um token emitido para um tenant nunca poderá ser utilizado para aprovar recursos de outro tenant.

---

# 9. Cadastro de participantes e contatos

O cadastro atual de nome e e-mail deverá ser ampliado.

## 9.1 Campos obrigatórios ou suportados

```text
Nome
E-mail
Departamento
Empresa
Telefone
Cidade
Estado
País
```

## 9.2 Campos adicionais recomendados

```text
Cargo
Unidade
Centro de custo
Gestor
Idioma preferencial
Fuso horário
Tipo de vínculo
Data de admissão
Status do participante
Tags
Grupo de risco
Código externo
Origem da sincronização
```

O telefone deverá ser normalizado no padrão E.164 quando utilizado em campanhas SMS.

## 9.3 Importação

O sistema deverá suportar:

* CSV;
* XLSX;
* API;
* sincronização futura por SCIM;
* integração futura com Microsoft Entra ID;
* integração futura com Google Workspace;
* validação antes da importação;
* preview;
* identificação de duplicidades;
* atualização incremental;
* relatório de erros.

## 9.4 Segmentação

Campanhas poderão ser direcionadas por:

* empresa;
* departamento;
* cidade;
* estado;
* país;
* cargo;
* unidade;
* tags;
* nível de risco;
* histórico de treinamento;
* reincidência;
* idioma.

---

## 9.5 Contatos responsáveis pelo cliente

Campos recomendados:

```text
Nome
E-mail
Telefone
Cargo
Departamento
Empresa
Papel no contrato
Papel na aprovação
Status
Idioma
Fuso horário
Data de início da autorização
Data de término da autorização
```

Esses contatos não deverão ser automaticamente incluídos como participantes das campanhas.

Os endereços utilizados para aprovação deverão ser previamente validados e associados ao tenant correto.

---

# 10. Autenticação e acesso administrativo

## 10.1 Painel administrativo

O painel administrativo não deverá ser exposto diretamente à internet.

O acesso deverá ocorrer por uma das opções autorizadas:

* VPN corporativa;
* Zero Trust Access;
* rede administrativa privada;
* bastion;
* túnel autenticado;
* proxy corporativo com identidade.

A porta administrativa do container não deverá ser publicada diretamente no host.

## 10.2 OIDC

O sistema deverá suportar:

* Microsoft Entra ID;
* Keycloak;
* outros provedores compatíveis com OIDC.

A conta administrativa local será destinada apenas à recuperação de emergência.

## 10.3 MFA

A arquitetura deverá estar preparada desde o início para MFA.

A evolução prevista inclui:

* TOTP;
* Microsoft Authenticator;
* Google Authenticator;
* códigos de recuperação;
* obrigatoriedade de MFA por perfil;
* MFA obrigatório para administradores;
* revogação de dispositivos;
* registro de eventos de autenticação.

O mecanismo TOTP deverá seguir padrões interoperáveis, evitando dependência exclusiva de um fornecedor.

## 10.4 Evoluções recomendadas

* WebAuthn e passkeys;
* chaves físicas FIDO2;
* autenticação adaptativa;
* bloqueio por risco;
* sessões administrativas curtas;
* reautenticação para operações críticas.

---

# 11. Entrega de e-mails

## 11.1 Camada de abstração

O EthPhish deverá possuir uma interface única para provedores de entrega.

```text
DeliveryProvider
├── SMTP Relay
├── SMTP autenticado
├── Brevo API
├── SendGrid API
└── provedores futuros
```

## 11.2 SMTP

Deverá suportar:

* SMTP relay;
* SMTP com autenticação;
* STARTTLS;
* TLS direto;
* autenticação por usuário e senha;
* certificados de cliente, quando aplicável;
* limites por minuto;
* limites por domínio;
* timeout configurável;
* validação de conexão.

## 11.3 Brevo

O provider Brevo deverá suportar:

* autenticação por API key;
* seleção de remetente aprovado;
* tags de campanha;
* consulta de status;
* classificação de erros;
* métricas de entrega;
* rate limiting.

## 11.4 SendGrid

O provider SendGrid deverá suportar:

* autenticação por API key;
* remetentes verificados;
* categorias;
* templates quando autorizados;
* eventos de entrega;
* tratamento de bounces;
* rate limiting.

## 11.5 Segredos

Credenciais de SMTP e API deverão ser:

* criptografadas;
* mascaradas na interface;
* nunca registradas em logs;
* rotacionáveis;
* associadas a tenants;
* acessíveis somente aos workers autorizados;
* preferencialmente armazenadas em cofre de segredos.

## 11.6 Políticas de envio

Cada campanha deverá definir:

* provider;
* worker pool;
* taxa máxima;
* janela de envio;
* domínios permitidos;
* quantidade máxima de tentativas;
* política de retry;
* data de expiração;
* aprovação necessária.

---

## 11.7 E-mails transacionais de aprovação

Eles deverão ser enviados por um provider transacional aprovado, com:

* remetente institucional;
* domínio corporativo da plataforma;
* assunto padronizado;
* identidade visual do EthPhish;
* prazo de validade;
* identificação do tenant;
* descrição do item pendente;
* link seguro para o portal;
* token de uso único ou código de confirmação;
* canal para comunicar divergências.

Esses e-mails não deverão compartilhar tracking, remetentes ou domínios utilizados nas simulações.

## 11.8 Segurança dos tokens

Os tokens de aprovação deverão:

* ser criptograficamente aleatórios;
* possuir uso único;
* ter validade limitada;
* ser armazenados de forma não reversível;
* ser vinculados ao usuário;
* ser vinculados ao tenant;
* ser vinculados ao recurso;
* ser vinculados à versão;
* ser invalidados após o uso;
* ser invalidados após alteração do conteúdo;
* ser invalidados quando um novo token for emitido;
* possuir proteção contra tentativas repetidas.

O link recebido por e-mail não deverá aprovar automaticamente o item.

Fluxo recomendado:

```text
Cliente recebe o e-mail
        ↓
Acessa o portal seguro
        ↓
Realiza autenticação
        ↓
Visualiza o conteúdo completo
        ↓
Informa ou confirma o token
        ↓
Aceita ou rejeita explicitamente
        ↓
Sistema registra a evidência
```

Para ações contratuais ou de maior impacto, o tenant poderá exigir autenticação e token no mesmo fluxo.

---

# 12. Workers distribuídos

## 12.1 Objetivo

Permitir que entregas sejam executadas por workers em redes diferentes, mantendo controle centralizado.

## 12.2 Características

Cada worker deverá possuir:

* identidade única;
* certificado ou credencial própria;
* pool;
* tenant ou tenants autorizados;
* capacidades;
* limites de concorrência;
* limites de envio;
* provedores disponíveis;
* heartbeat;
* versão do software;
* informações de hardware;
* informações do sistema operacional;
* status de saúde.

## 12.3 Segurança

Os workers deverão:

* iniciar conexões de saída;
* utilizar TLS ou mTLS;
* não expor portas administrativas;
* validar assinatura e expiração dos jobs;
* rejeitar domínios não autorizados;
* rejeitar jobs de tenants não autorizados;
* descartar conteúdo após processamento;
* utilizar spool local cifrado;
* limitar volume do spool;
* operar como usuário não root.

## 12.4 Idempotência

Cada entrega deverá possuir uma chave idempotente.

O mesmo job não poderá gerar duas entregas sem uma ação manual auditada.

## 12.5 Hardware inicial

### Servidor central

```text
2 vCPU
4 GB RAM
100 GB SSD
```

Essa configuração será aceita para MVP de baixo volume, com monitoramento leve e limites explícitos.

### Cada worker

```text
1 vCPU
1 GB RAM
50 GB SSD
```

Concorrência inicial recomendada:

```text
2 a 5 entregas simultâneas
```

---

# 13. Landing pages em HTTPS

## 13.1 Exposição

Toda landing page deverá ser publicada externamente na porta 443.

A porta 80 poderá ser utilizada apenas para:

* redirecionamento imediato para HTTPS;
* validação ACME, quando necessária.

Nenhum conteúdo de campanha deverá ser servido em HTTP sem criptografia.

## 13.2 TLS

A plataforma deverá suportar:

* TLS 1.2 e TLS 1.3;
* certificados automatizados;
* certificados corporativos;
* múltiplos domínios;
* SNI;
* renovação automática;
* alertas de expiração;
* HSTS quando compatível com o domínio utilizado.

## 13.3 Separação

Os domínios públicos de campanhas não deverão expor:

* painel administrativo;
* API administrativa;
* métricas internas;
* documentação técnica;
* endpoints de gerenciamento.

## 13.4 Domínios personalizados

Cada tenant poderá utilizar domínios de campanha previamente autorizados.

O processo deverá incluir:

* cadastro;
* validação DNS;
* emissão de certificado;
* aprovação;
* vínculo ao tenant;
* data de expiração;
* revogação.

---

# 14. Módulo de treinamento

## 14.1 Objetivo

Converter uma interação de risco em uma experiência educativa.

Após uma interação definida pela campanha, o participante poderá ser direcionado para:

* explicação do cenário;
* sinais que deveriam ter sido observados;
* conteúdo educativo;
* vídeo;
* quiz;
* confirmação de leitura;
* treinamento completo.

## 14.2 Estrutura

```text
Programa
├── Módulos
│   ├── Aulas
│   ├── Conteúdos
│   ├── Vídeos
│   └── Quizzes
└── Certificação
```

## 14.3 Métricas

O módulo deverá registrar:

* treinamento atribuído;
* treinamento iniciado;
* aulas visualizadas;
* percentual concluído;
* tempo de permanência;
* quiz iniciado;
* quiz concluído;
* nota;
* aprovação;
* número de tentativas;
* conclusão;
* data de expiração.

## 14.4 Indicadores

* taxa de início;
* taxa de conclusão;
* tempo médio de conclusão;
* nota média;
* evolução entre tentativas;
* departamentos com menor adesão;
* reincidência após treinamento;
* impacto do treinamento nas campanhas seguintes.

## 14.5 Treinamento adaptativo

Evolução recomendada:

* conteúdos baseados no tipo de erro;
* trilhas por departamento;
* trilhas por função;
* treinamento mais avançado para reincidentes;
* reforço periódico;
* recomendações automáticas.

---

# 15. Dashboards

## 15.1 Dashboard operacional

Voltado para administradores e operadores.

Deverá apresentar:

* campanhas ativas;
* campanhas agendadas;
* jobs em fila;
* entregas por minuto;
* falhas;
* retries;
* dead-letter queue;
* saúde dos workers;
* saúde dos provedores;
* latência de envio;
* status de landing pages;
* certificados próximos da expiração;
* uso de PostgreSQL;
* uso de RabbitMQ;
* utilização do servidor central;
* eventos recentes;
* incidentes operacionais.

## 15.2 Dashboard executivo

Voltado para clientes e liderança.

Deverá apresentar:

* taxa de interação;
* taxa de reporte;
* taxa de conclusão de treinamento;
* evolução histórica;
* risco por departamento;
* risco por empresa;
* risco por localização;
* reincidência;
* tempo médio para reporte;
* impacto dos treinamentos;
* cobertura das campanhas;
* índice de resiliência humana;
* tendências.

## 15.3 Dashboard do cliente

Cada cliente visualizará apenas dados pertencentes ao seu tenant ou às empresas autorizadas.

O portal deverá aplicar:

* filtros por empresa;
* filtros por campanha;
* filtros por período;
* filtros por departamento;
* filtros geográficos;
* anonimização opcional;
* exportação conforme permissão.

---

## 15.4 Central de aprovações

Ela deverá apresentar:

* itens aguardando aprovação;
* itens aprovados;
* itens rejeitados;
* itens expirados;
* itens substituídos por novas versões;
* solicitações de alteração;
* prazos;
* responsáveis;
* histórico de decisões;
* campanhas bloqueadas por pendências.

A central deverá permitir filtrar por:

* empresa;
* campanha;
* tipo de aprovação;
* responsável;
* status;
* período;
* prazo;
* prioridade.

## 15.5 Aprovação de templates de e-mail

O cliente deverá conseguir revisar o template exatamente como será utilizado.

A interface deverá apresentar:

* assunto;
* nome do remetente;
* endereço do remetente;
* resposta configurada;
* corpo HTML;
* corpo em texto;
* imagens;
* links;
* variáveis dinâmicas;
* anexos autorizados;
* visualização desktop;
* visualização mobile;
* versão;
* data da última alteração;
* autor da alteração.

O cliente poderá:

```text
Aprovar
Rejeitar
Solicitar alteração
Adicionar comentário
Solicitar novo template
```

O sistema deverá impedir a execução da campanha quando o template obrigatório ainda não estiver aprovado.

## 15.6 Aprovação de landing pages

O cliente deverá visualizar a landing page em ambiente seguro de preview.

A interface deverá apresentar:

* layout completo;
* textos;
* imagens;
* campos simulados;
* mensagens posteriores à interação;
* redirecionamento previsto;
* treinamento associado;
* domínio planejado;
* versão;
* comportamento em dispositivos móveis.

O preview deverá ser isolado, sem registrar o cliente como participante e sem produzir métricas reais da campanha.

O cliente poderá:

```text
Aprovar
Rejeitar
Solicitar alteração
Adicionar comentário
Solicitar nova landing page
```

## 15.7 Aprovação do contrato e do escopo

Antes da execução da campanha, o cliente deverá visualizar e aprovar:

* contrato de prestação de serviço;
* termo ou ordem de serviço;
* escopo da campanha;
* empresas envolvidas;
* domínios autorizados;
* canais utilizados;
* volume aproximado;
* período de execução;
* objetivos;
* limitações;
* responsabilidades;
* regras de privacidade;
* dados coletados;
* política de retenção;
* contatos de emergência;
* condições de interrupção.

A interface deverá exigir uma ação afirmativa, como:

```text
Declaro que li e concordo com o contrato e com o escopo apresentado.
```

A aceitação deverá estar vinculada ao hash do documento e da versão do escopo.

O sistema deverá permitir baixar uma cópia do contrato aceito e do comprovante da aprovação.

## 15.8 Solicitação de novos templates

O portal deverá permitir que o cliente solicite:

* novo template de e-mail;
* nova landing page;
* novo treinamento;
* adaptação de um conteúdo existente;
* tradução;
* alteração de identidade visual;
* conteúdo específico para determinado departamento;
* cenário específico de conscientização.

A solicitação deverá incluir:

```text
Título
Objetivo
Empresa
Campanha relacionada
Público-alvo
Idioma
Prazo desejado
Referências
Arquivos anexos
Observações
Nível de prioridade
```

Workflow sugerido:

```text
Solicitação criada
        ↓
Triagem interna
        ↓
Em elaboração
        ↓
Disponível para revisão
        ↓
Alteração solicitada ou aprovação
        ↓
Aprovado
        ↓
Disponível para campanha
```

## 15.9 Comentários e colaboração

Cada item sujeito à aprovação deverá possuir uma linha de discussão.

Os comentários deverão registrar:

* autor;
* data e hora;
* tenant;
* item;
* versão;
* anexos;
* menções;
* resolução;
* histórico de edição.

Comentários internos da equipe EthPhish não deverão ser apresentados ao cliente, salvo quando explicitamente marcados como públicos.

---

## 15.10 Estados do workflow de aprovação

Os estados padronizados deverão ser:

```text
Rascunho
Em elaboração
Em revisão interna
Aguardando aprovação do cliente
Alteração solicitada
Rejeitado
Aprovado
Aprovação expirada
Aprovação invalidada
Substituído
Arquivado
```

Uma campanha somente poderá avançar para `Aprovada` ou `Agendada` quando todas as aprovações obrigatórias estiverem válidas.

### 15.10.1 Matriz de bloqueios

| Pendência                        |   Criação |       Agendamento |                     Execução |
| -------------------------------- | --------: | ----------------: | ---------------------------: |
| Contrato não aprovado            | Permitida |         Bloqueado |                    Bloqueado |
| Escopo não aprovado              | Permitida |         Bloqueado |                    Bloqueado |
| E-mail não aprovado              | Permitida |         Bloqueado |                    Bloqueado |
| Landing page não aprovada        | Permitida |         Bloqueado |                    Bloqueado |
| Treinamento não aprovado         | Permitida | Conforme política | Bloqueado quando obrigatório |
| Aprovação expirada               | Permitida |         Bloqueado |                    Bloqueado |
| Conteúdo alterado após aprovação | Permitida |         Bloqueado |                    Bloqueado |

---

# 16. Estatísticas de hardware e software

## 16.1 Central

A interface administrativa deverá mostrar:

* CPU total e utilizada;
* memória total e utilizada;
* disco total e utilizado;
* load average;
* uptime;
* versão do sistema operacional;
* versão do Docker;
* versão do EthPhish;
* versão do PostgreSQL;
* versão do RabbitMQ;
* quantidade de containers;
* reinícios;
* status de backups;
* tamanho do banco;
* conexões ativas;
* filas;
* latência interna.

## 16.2 Workers

Para cada worker:

* status online ou offline;
* último heartbeat;
* CPU;
* memória;
* disco;
* uptime;
* versão do worker;
* sistema operacional;
* arquitetura;
* IP ou rede lógica;
* pool;
* capacidades;
* jobs ativos;
* jobs concluídos;
* falhas;
* latência;
* spool local;
* providers disponíveis.

## 16.3 Alertas

A plataforma deverá alertar para:

* memória acima de 85%;
* disco acima de 80%;
* worker offline;
* fila crescendo continuamente;
* certificados expirando;
* backup atrasado;
* excesso de retries;
* falhas de autenticação;
* provider indisponível;
* versão de worker incompatível.

---

# 17. Classificação de falsos positivos

O EthPhish deverá separar eventos em:

```text
Interação humana provável
Interação humana confirmada
Scanner provável
Gateway de segurança
Evento inconclusivo
Evento reclassificado manualmente
```

O cálculo deverá considerar:

* sequência dos eventos;
* intervalo entre entrega e acesso;
* origem previamente conhecida;
* comportamento de carregamento;
* informações recebidas dos gateways;
* repetição;
* confirmação posterior;
* revisão manual.

O sistema deverá mostrar a pontuação de confiança e os motivos da classificação.

Um clique isolado não deverá ser automaticamente interpretado como falha definitiva.

---

# 18. Melhorias recomendadas

## 18.1 Workflow de aprovação

Campanhas de maior risco deverão exigir aprovação de duas pessoas.

Estados:

```text
Rascunho
Em revisão
Aprovada
Agendada
Em execução
Pausada
Concluída
Cancelada
```

## 18.2 Limites por tenant

Cada tenant deverá possuir:

* quantidade máxima de participantes;
* campanhas mensais;
* mensagens por período;
* workers;
* armazenamento;
* domínios;
* operadores.

## 18.3 Score de risco humano

Criar um índice baseado em:

* interações recentes;
* reincidência;
* reporte;
* conclusão de treinamento;
* tempo de resposta;
* dificuldade da campanha.

O score deverá ser explicável, revisável e não utilizado como instrumento disciplinar automático.

## 18.4 Biblioteca de campanhas

Criar biblioteca versionada com:

* templates;
* landing pages;
* treinamentos;
* níveis de dificuldade;
* categorias de risco;
* idiomas;
* tags.

## 18.5 Internacionalização

Suportar:

* português;
* inglês;
* espanhol;
* formatos regionais;
* fusos horários;
* conteúdos por idioma.

## 18.6 Acessibilidade

Interface e treinamentos deverão buscar conformidade WCAG AA.

## 18.7 Integração com SIEM

Disponibilizar eventos para:

* Microsoft Sentinel;
* Splunk;
* Elastic;
* QRadar;
* syslog;
* webhooks.

## 18.8 API corporativa

Criar API para:

* usuários;
* grupos;
* campanhas;
* treinamentos;
* relatórios;
* métricas;
* sincronização.

A API deverá utilizar escopos, rate limiting e credenciais por tenant.

## 18.9 Auditoria inviolável

A trilha de auditoria deverá registrar:

* login;
* falhas de login;
* criação e aprovação de campanhas;
* alteração de templates;
* exportações;
* visualização de dados sensíveis;
* mudanças de permissões;
* configuração de providers;
* configuração de workers.

Como evolução, os registros poderão utilizar encadeamento de hashes ou armazenamento imutável.

## 18.10 Segurança da cadeia de software

A pipeline deverá incluir:

* testes;
* análise estática;
* `govulncheck`;
* CodeQL;
* Gitleaks;
* Trivy;
* análise de dependências;
* scan de imagens;
* SBOM;
* assinatura Cosign;
* artefatos versionados;
* Actions fixadas por SHA;
* análise de licenças.

---

## 18.11 Assinatura eletrônica e integração contratual

Exemplos de capacidades:

* envio do contrato para assinatura;
* consulta de status;
* múltiplos signatários;
* ordem de assinatura;
* download do documento final;
* armazenamento do comprovante;
* webhooks de conclusão;
* expiração;
* cancelamento;
* reenvio.

A aceitação por token poderá ser utilizada inicialmente como aprovação registrada do workflow. Quando o contexto jurídico exigir assinatura eletrônica qualificada ou avançada, deverá ser utilizado um provedor apropriado e homologado.

## 18.12 SLA de aprovação

Cada tenant poderá definir:

* prazo padrão para revisão;
* prazo para contrato;
* prazo para conteúdo;
* lembretes automáticos;
* escalonamento;
* substituto do aprovador;
* cancelamento automático após expiração.

## 18.13 Versionamento obrigatório

Templates, landing pages, treinamentos, contratos e escopos deverão possuir versões imutáveis.

Exemplo:

```text
Template de e-mail
v1 — rejeitado
v2 — alteração solicitada
v3 — aprovado
v4 — criado após nova alteração, aguardando aprovação
```

A criação da versão 4 não deverá apagar ou alterar a evidência da aprovação da versão 3.

## 18.14 Comparação entre versões

O portal deverá oferecer comparação visual entre versões, destacando:

* textos adicionados;
* textos removidos;
* links alterados;
* imagens alteradas;
* campos adicionados;
* alterações de remetente;
* mudanças de domínio;
* mudanças no treinamento;
* mudanças no escopo.

## 18.15 Aprovação emergencial

Campanhas urgentes poderão possuir um processo de aprovação acelerado, desde que:

* seja habilitado pelo tenant;
* exija usuário com permissão especial;
* registre a justificativa;
* mantenha as aprovações contratuais obrigatórias;
* gere alerta para auditoria;
* não elimine controles de domínio e escopo.

---

# 19. Roadmap por sprints

Premissa: sprints de duas semanas.

## Sprint 0 — Descoberta, governança e baseline

### Objetivo

Formalizar escopo, arquitetura e critérios de segurança.

### Entregas

* repositório do EthPhish;
* definição do upstream Anglerphish 1.3.0;
* inventário funcional;
* threat model inicial;
* política de uso aceitável;
* modelo inicial de tenants;
* ADRs arquiteturais;
* backlog priorizado;
* critérios de pronto.

### Critérios de aceite

* arquitetura aprovada;
* riscos principais documentados;
* responsabilidades definidas;
* ambientes identificados.

---

## Sprint 1 — Fundação do fork e CI/CD

### Objetivo

Criar uma base reproduzível e segura.

### Entregas

* fork corporativo;
* atualização do Dockerfile;
* build multi-stage;
* execução sem root;
* configurações por variáveis de ambiente;
* pipeline de build e testes;
* scans básicos;
* versionamento próprio;
* imagem do servidor;
* imagem futura do worker.

### Critérios de aceite

* build reproduzível;
* container inicia sem root;
* painel administrativo não é publicado;
* pipeline bloqueia falhas críticas.

---

## Sprint 2 — PostgreSQL e migrations

### Objetivo

Substituir SQLite por PostgreSQL.

### Entregas

* driver PostgreSQL;
* schema;
* migrations;
* pool de conexões;
* testes de integração;
* ferramenta inicial de migração;
* backup e restore;
* health check.

### Critérios de aceite

* funções principais operam no PostgreSQL;
* migrations são reversíveis ou possuem rollback documentado;
* restauração testada;
* SQLite não é necessário em produção.

---

## Sprint 3 — Fundação multitenant

### Objetivo

Adicionar segregação de tenants ao domínio.

### Entregas

* tabelas de tenants;
* empresas;
* vínculos de usuários;
* `tenant_id`;
* Row-Level Security;
* middleware de tenant;
* testes de isolamento;
* storage particionado.

* responsáveis por aprovação;
* papéis de aprovação;
* segregação de workflows por tenant;
* políticas de aprovação por empresa.

### Critérios de aceite

* tenant A não acessa dados do tenant B;
* testes automatizados comprovam isolamento;
* consultas sem tenant são negadas.
* um aprovador acessa apenas itens de empresas autorizadas;
* tokens não funcionam fora do tenant de origem.

---

## Sprint 4 — Cadastro ampliado e segmentação

### Objetivo

Expandir o cadastro de participantes.

### Entregas

* departamento;
* empresa;
* telefone;
* cidade;
* estado;
* país;
* cargo;
* unidade;
* tags;
* importação CSV/XLSX;
* validação e preview;
* filtros e grupos dinâmicos.

### Critérios de aceite

* importação informa erros sem corromper dados;
* telefone é normalizado;
* participantes podem ser segmentados por todos os campos principais.

---

## Sprint 5 — Autenticação, OIDC e proteção administrativa

### Objetivo

Proteger o plano administrativo.

### Entregas

* integração OIDC;
* Microsoft Entra ID;
* Keycloak;
* RBAC;
* sessões seguras;
* conta de emergência;
* rede administrativa;
* reverse proxy privado;
* logs de autenticação.

* autenticação de clientes aprovadores;
* validação de endereço de e-mail;
* sessão específica para fluxo de aprovação;
* emissão e validação de tokens de uso único.

### Critérios de aceite

* painel não acessível diretamente pela internet;
* acesso administrativo exige canal autorizado;
* permissões são aplicadas no backend.
* token sozinho não concede acesso a outros recursos;
* tokens expirados, utilizados ou revogados são rejeitados;
* todas as tentativas são auditadas.

---

## Sprint 6 — Contratos, escopo e workflow de aprovação

### Objetivo

Implementar o processo formal de revisão e autorização das campanhas.

### Entregas

* cadastro de contratos;
* upload de documentos;
* versionamento;
* definição do escopo;
* responsáveis por aprovação;
* Central de Aprovações;
* emissão de token;
* aprovação e rejeição;
* comentários;
* solicitação de alterações;
* expiração;
* lembretes;
* evidências de aceite;
* bloqueio de campanhas sem autorização.

### Critérios de aceite

* contrato e escopo podem ser vinculados a uma campanha;
* cliente recebe token no e-mail validado;
* aprovação fica associada à versão exata;
* alteração material invalida aprovação;
* campanha não pode ser executada sem aprovações obrigatórias;
* comprovante de aprovação pode ser consultado e exportado.

---

## Sprint 7 — Aprovação de e-mails, landing pages e solicitações

Permitir colaboração segura entre a equipe EthPhish e o cliente.

### Entregas

* preview de e-mail;
* preview de landing page;
* preview mobile e desktop;
* comentários;
* aprovação;
* rejeição;
* solicitação de alteração;
* solicitação de novo template;
* anexos;
* comparação entre versões;
* notificações;
* histórico completo.

### Critérios de aceite

* preview não gera eventos reais;
* aprovação identifica o hash da versão;
* templates alterados retornam ao estado pendente;
* cliente pode solicitar novo conteúdo;
* campanha permanece bloqueada enquanto houver pendência obrigatória.

---

## Sprint 8 — Abstração de provedores de entrega

### Objetivo

Criar camada única de entrega.

### Entregas

* contrato `DeliveryProvider`;
* SMTP relay;
* SMTP autenticado;
* Brevo API;
* SendGrid API;
* teste de conexão;
* gestão de credenciais;
* classificação de erros;
* limites de envio.

### Critérios de aceite

* uma campanha pode selecionar um provider;
* credenciais não aparecem em logs;
* erros temporários e permanentes são diferenciados.

---

## Sprint 9 — Broker, outbox e worker SMTP

### Objetivo

Implantar execução distribuída.

### Entregas

* RabbitMQ;
* transactional outbox;
* worker SMTP;
* heartbeat;
* registro de worker;
* worker pools;
* idempotência;
* retries;
* dead-letter queue;
* pause e cancelamento.

### Critérios de aceite

* central distribui jobs para dois workers;
* job duplicado não gera entrega duplicada;
* worker offline não causa perda permanente;
* campanha pode ser pausada.

---

## Sprint 10 — Workers Brevo e SendGrid

### Objetivo

Distribuir também as entregas por API.

### Entregas

* executor Brevo;
* executor SendGrid;
* rate limiting;
* tratamento de respostas;
* métricas por provider;
* seleção de provider por worker pool.

### Critérios de aceite

* providers por API funcionam sem execução no servidor central;
* falhas são auditáveis;
* limites são respeitados.

---

## Sprint 11 — Landing pages na porta 443

### Objetivo

Criar uma camada pública segura e isolada.

### Entregas

* listener público HTTPS;
* automação de certificados;
* domínios por tenant;
* validação DNS;
* SNI;
* separação de rotas;
* hardening HTTP;
* isolamento do painel administrativo.

### Critérios de aceite

* landing pages funcionam externamente em 443;
* painel e APIs internas não aparecem no domínio público;
* certificados são renovados automaticamente.

---

## Sprint 12 — Módulo de treinamento

### Objetivo

Integrar conscientização após a campanha.

### Entregas

* programas;
* módulos;
* aulas;
* quizzes;
* atribuições;
* progresso;
* conclusão;
* pontuação;
* vínculo campanha-treinamento;
* página educativa pós-interação.

### Critérios de aceite

* campanha pode atribuir treinamento;
* progresso é registrado;
* conteúdo sensível digitado não é persistido.

---

## Sprint 13 — Portal multitenant do cliente

### Objetivo

Permitir acesso seguro aos clientes.

### Entregas

* área autenticada;
* dashboard por tenant;
* campanhas;
* treinamentos;
* filtros;
* relatórios;
* perfis executivos e analíticos;
* anonimização.

* Central de Aprovações;
* contratos;
* escopos;
* previews;
* solicitações de templates;
* comentários;
* notificações;
* comprovantes;
* histórico de versões.

### Critérios de aceite

* cliente acessa somente dados autorizados;
* dados de outros tenants não são expostos;
* permissões são aplicadas no backend e no frontend.
* cliente aprova somente itens autorizados;
* aprovação exige autenticação e token conforme a política;
* contrato aceito fica disponível para consulta;
* dados e workflows permanecem isolados entre tenants.

---

## Sprint 14 — Observabilidade de hardware e software

### Objetivo

Disponibilizar visão operacional da plataforma.

### Entregas

* métricas do central;
* métricas dos workers;
* versões;
* CPU;
* memória;
* disco;
* filas;
* banco;
* uptime;
* painel de saúde;
* alertas.

### Critérios de aceite

* administrador identifica worker indisponível;
* consumo de recursos é visível;
* alertas são gerados para limites críticos.

---

## Sprint 15 — Analytics e falsos positivos

### Objetivo

Melhorar a qualidade das métricas.

### Entregas

* classificação de eventos;
* pontuação de confiança;
* identificação de scanners;
* revisão manual;
* funis corrigidos;
* métricas de comportamento;
* histórico por participante.

### Critérios de aceite

* eventos automatizados são separados;
* classificação apresenta justificativa;
* relatórios diferenciam clique bruto e interação provável.

---

## Sprint 16 — Dashboard executivo

### Objetivo

Transformar eventos em indicadores de negócio.

### Entregas

* evolução histórica;
* risco por departamento;
* risco geográfico;
* reincidência;
* treinamento;
* tempo para reporte;
* índice de resiliência;
* relatórios executivos.

### Critérios de aceite

* indicadores podem ser filtrados;
* métricas agregadas preservam privacidade;
* relatórios são consistentes com os eventos classificados.

---

## Sprint 17 — Exportação, importação e biblioteca

### Objetivo

Permitir portabilidade e reutilização.

### Entregas

* bundle versionado;
* exportação de campanhas;
* templates;
* landing pages;
* treinamentos;
* assets;
* preview de importação;
* resolução de conflitos;
* biblioteca compartilhada.

### Critérios de aceite

* segredos não são exportados;
* importação valida schema;
* conteúdo possui checksums;
* conflitos são apresentados antes da gravação.

---

## Sprint 18 — MFA e segurança avançada

### Objetivo

Elevar a segurança das identidades.

### Entregas

* TOTP;
* Microsoft Authenticator;
* Google Authenticator;
* códigos de recuperação;
* MFA obrigatório por perfil;
* reautenticação;
* gestão de sessões;
* preparação para WebAuthn.

### Critérios de aceite

* administradores podem ser obrigados a usar MFA;
* dispositivo pode ser revogado;
* ações críticas exigem autenticação recente.

---

## Sprint 19 — Hardening e supply chain

### Objetivo

Preparar a solução para uso corporativo.

### Entregas

* SAST;
* DAST;
* scans de imagem;
* SBOM;
* assinatura;
* secret scanning;
* análise de dependências;
* testes de isolamento;
* pentest;
* correções.

* teste de adulteração de tokens;
* teste de reutilização;
* teste de expiração;
* teste de enumeração de aprovações;
* teste de acesso cruzado entre tenants;
* teste de alteração posterior à aprovação;
* teste de replay;
* proteção CSRF;
* rate limiting;
* logs de não repúdio.

### Critérios de aceite

* nenhuma vulnerabilidade crítica conhecida permanece sem plano aprovado;
* imagens são assinadas;
* SBOM acompanha as releases;
* pentest concluído.

---

## Sprint 20 — Piloto controlado

### Objetivo

Validar o EthPhish com usuários reais autorizados.

### Entregas

* tenant piloto;
* duas redes de workers;
* SMTP e API;
* campanha controlada;
* treinamento;
* dashboard;
* relatório de desempenho;
* coleta de feedback.

### Critérios de aceite

* campanha concluída sem perda de jobs;
* isolamento validado;
* métricas reconciliadas;
* restauração testada;
* feedback registrado.

---

## Sprint 21 — Disponibilidade geral

### Objetivo

Preparar a primeira versão estável.

### Entregas

* documentação;
* runbooks;
* termos de uso;
* política de privacidade;
* política de retenção;
* suporte;
* SLA inicial;
* plano de continuidade;
* roadmap seguinte.

### Critérios de aceite

* documentação operacional completa;
* monitoramento ativo;
* backups testados;
* processo de incidentes definido;
* release assinada e aprovada.

---

# 20. Indicadores de sucesso

## Produto

* número de tenants ativos;
* campanhas concluídas;
* participantes treinados;
* taxa de adoção do portal;
* taxa de conclusão de treinamentos;
* retenção de clientes.

## Segurança

* zero vazamento entre tenants;
* percentual de administradores com MFA;
* vulnerabilidades críticas abertas;
* tempo de correção;
* número de acessos administrativos bloqueados.

## Operação

* disponibilidade;
* jobs processados;
* taxa de falha;
* entregas duplicadas impedidas;
* tempo médio de processamento;
* uptime dos workers;
* sucesso dos backups.

## Conscientização

* redução de reincidência;
* aumento da taxa de reporte;
* redução do tempo de reporte;
* evolução da nota de treinamento;
* melhoria do índice de resiliência.

---

## Workflow comercial e de aprovação

* tempo médio para aprovação de contratos;
* tempo médio para aprovação de templates;
* quantidade de alterações solicitadas;
* percentual de campanhas aprovadas no prazo;
* número de tokens expirados;
* número de campanhas bloqueadas por pendência;
* quantidade de novos templates solicitados;
* tempo médio de produção de novo conteúdo;
* percentual de aprovações concluídas sem suporte manual.

---

# 21. Definition of Done

Uma funcionalidade será considerada pronta somente quando possuir:

* código revisado;
* testes unitários;
* testes de integração;
* validação multitenant;
* logs estruturados;
* métricas;
* documentação;
* tratamento de erros;
* controle de autorização;
* análise de segurança;
* migration quando aplicável;
* rollback documentado;
* interface acessível;
* critérios de aceite aprovados.

---

Para recursos sujeitos à aprovação, também serão obrigatórios:

* versão imutável;
* hash do conteúdo;
* workflow configurado;
* notificação ao responsável;
* token seguro;
* registro da decisão;
* teste de isolamento multitenant;
* invalidação após alteração;
* trilha de auditoria;
* comprovante exportável;
* bloqueio de execução quando houver pendência.

---

# 22. Gate de liberação de campanha

Antes do agendamento ou execução, o EthPhish deverá avaliar automaticamente:

```text
Contrato vigente e aprovado?
Escopo aprovado?
Domínios autorizados?
Público-alvo dentro do escopo?
Template de e-mail aprovado?
Landing page aprovada?
Treinamento aprovado quando obrigatório?
Provider autorizado?
Worker pool autorizado?
Período da campanha válido?
Todas as aprovações ainda estão vigentes?
```

Caso qualquer resposta obrigatória seja negativa, a campanha deverá permanecer bloqueada.

O sistema deverá apresentar claramente:

* motivo do bloqueio;
* responsável pela pendência;
* prazo;
* ação necessária;
* histórico relacionado.

---

# 23. Fluxo consolidado de contratação, aprovação e execução

```text
Equipe cria a oportunidade ou campanha
        ↓
Contrato e escopo são cadastrados
        ↓
Cliente recebe notificação e token
        ↓
Cliente autentica e aprova o contrato
        ↓
Equipe prepara e-mail, landing page e treinamento
        ↓
Cliente revisa os previews
        ↓
Cliente aprova ou solicita alterações
        ↓
Todas as aprovações são validadas
        ↓
Sistema libera o agendamento
        ↓
Campanha é executada
        ↓
Cliente acompanha indicadores e treinamentos
        ↓
Evidências permanecem disponíveis para auditoria
```

Esse processo transforma o portal do cliente em uma área de governança e colaboração, não apenas de consulta de indicadores. A campanha passa a possuir autorização técnica, comercial e operacional comprovável antes de qualquer execução.

---

# 24. Fora do escopo

O EthPhish não deverá ser desenvolvido para:

* campanhas não autorizadas;
* evasão de ferramentas de segurança;
* captura de credenciais reais;
* coleta invasiva de fingerprints;
* ocultação da origem de atividades;
* execução de payloads;
* exploração de vulnerabilidades;
* comprometimento de contas;
* interceptação de sessões;
* contorno de MFA.

---

# 25. Visão de longo prazo

O EthPhish deverá evoluir para uma plataforma de Human Risk Management, conectando campanhas, treinamentos e comportamento organizacional.

Possíveis evoluções:

* SCIM;
* Entra ID e Google Workspace;
* passkeys;
* aplicativo móvel;
* assistente para criação de conteúdos;
* trilhas adaptativas;
* detecção de anomalias;
* benchmark entre períodos;
* integrações com LMS;
* integrações com GRC;
* gestão de políticas;
* certificados de treinamento;
* alta disponibilidade;
* PostgreSQL gerenciado;
* múltiplas regiões;
* worker autoscaling;
* marketplace privado de conteúdos.

---

# 26. Declaração final

O EthPhish deverá ser construído como uma plataforma de conscientização corporativa segura, auditável e orientada por dados.

Seu diferencial não será apenas executar simulações, mas conectar:

```text
Campanha
    ↓
Comportamento
    ↓
Classificação
    ↓
Treinamento
    ↓
Evolução
    ↓
Redução de risco
```

A plataforma deverá manter o Anglerphish como base funcional, mas assumir identidade, arquitetura, governança e ciclo de desenvolvimento próprios.

O objetivo final é entregar uma ferramenta confiável para programas contínuos de conscientização, com segurança multitenant, execução distribuída, privacidade por padrão e indicadores capazes de orientar decisões técnicas e executivas.

