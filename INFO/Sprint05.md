Fundação multitenant

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

