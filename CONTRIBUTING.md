# Contribuindo com o EthPhish

O EthPhish é um fork corporativo do Anglerphish 1.3.0 para simulações éticas e
autorizadas. Mudanças voltadas a evasão, captura de credenciais reais ou uso não
autorizado não fazem parte do produto.

## Fluxo de desenvolvimento

1. Crie uma branch `feature/ETH-nnn-descricao`, `fix/ETH-nnn-descricao` ou
   `docs/ETH-nnn-descricao` a partir de `develop`.
2. Mantenha commits pequenos, rastreáveis e sem segredos.
3. Execute formatação, análise estática e testes.
4. Abra pull request com risco, testes, impacto multitenant e rollback.
5. Obtenha as aprovações exigidas por `CODEOWNERS`.

## Definition of Done

Uma alteração precisa de revisão, testes proporcionais ao risco, autorização no
backend, logs sem dados sensíveis, documentação, análise de segurança e plano de
reversão. Mudanças de banco precisam de migration e procedimento de rollback.

Mudanças que manipulam entidades de negócio também devem declarar como o tenant
é determinado e incluir teste que impeça acesso cruzado.

## Verificação local

```sh
gofmt -w caminho/do/arquivo.go
go vet ./...
go test -race ./...
docker compose config
docker build -t ethphish:dev .
```

Consulte `SECURITY.md` antes de relatar vulnerabilidades.
