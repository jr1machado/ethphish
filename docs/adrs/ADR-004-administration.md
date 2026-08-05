# ADR-004: isolamento administrativo

- Status: aceito
- Data: 2026-08-04

O painel administrativo nunca será publicado diretamente na internet. O acesso
ocorrerá por VPN, Zero Trust, bastion ou rede privada. O Compose de referência
expõe somente o proxy da superfície pública.
