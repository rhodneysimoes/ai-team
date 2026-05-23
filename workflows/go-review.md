# Workflow: Revisao Go

Use este fluxo para revisar alteracoes existentes.

## Etapas

1. Orquestrador identifica diff, arquivos alterados e intencao da mudanca.
2. Reviewer lista achados por severidade.
3. Security revisa se houver entrada externa, credenciais, rede ou autorizacao.
4. Tester aponta lacunas de cobertura.
5. Orquestrador consolida achados e perguntas abertas.

## Formato Da Revisao

- Achados primeiro, ordenados por severidade.
- Cada achado deve ter arquivo, linha, impacto e sugestao.
- Depois liste perguntas abertas.
- Resumo vem por ultimo e deve ser breve.

## Verificacoes Recomendadas

- `go test ./...`
- `go test -race ./...` para concorrencia.
- `go vet ./...`
- `govulncheck ./...` quando dependencias ou superficie de seguranca mudarem.
