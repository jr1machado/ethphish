# ADR-003: monólito modular e workers distribuídos

- Status: aceito
- Data: 2026-08-04

O control plane começará como monólito modular. Entregas serão gradualmente
extraídas para workers distribuídos via transactional outbox e RabbitMQ. Workers
não terão acesso administrativo nem acesso direto ao PostgreSQL.
