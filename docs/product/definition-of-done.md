# Definition of Done

Uma história somente é concluída quando contém:

- código revisado;
- testes unitários e de integração proporcionais ao risco;
- autorização verificada no backend;
- logs estruturados sem segredos ou dados pessoais desnecessários;
- métricas e health checks quando aplicável;
- documentação e tratamento de erro;
- migration e rollback documentado, quando persistir dados;
- avaliação de segurança e de isolamento multitenant;
- interface acessível, quando houver interface.

Itens de aprovação também exigem versão imutável, hash de conteúdo, token de uso
único, evidência de decisão, invalidação depois de alteração, comprovante
exportável e bloqueio de execução enquanto houver pendência.

Nenhuma mudança de produto pode ampliar o escopo para evasão, captura de
credenciais reais ou campanha não autorizada.
