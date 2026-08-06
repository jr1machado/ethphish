Autenticação, OIDC e proteção administrativa

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

