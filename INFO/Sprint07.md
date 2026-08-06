
Sprint 7

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

Toda entidade de negócio deverá possuir `tenant_id`, diretamente

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


