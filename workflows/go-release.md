# Workflow: Release Go

Use este fluxo para preparar build, CI, empacotamento ou entrega de uma aplicacao Go.

## Etapas

1. DevOps identifica tipo de artefato: binario, container, biblioteca ou servico.
2. Architect valida impactos de configuracao e compatibilidade quando necessario.
3. Developer ajusta codigo ou arquivos de build.
4. Tester executa testes e checks de regressao.
5. Security revisa secrets, dependencias e superficie exposta.
6. Reviewer faz revisao final.

## Checklist

- Versao Go definida.
- `go mod tidy` executado quando dependencias mudarem.
- `go test ./...` passa.
- `go build ./...` passa.
- Configuracao sensivel vem de ambiente ou secret manager.
- Health check e graceful shutdown existem para servicos.
- Dockerfile usa multi-stage build quando houver container.

## Saida Final

- Comandos de build.
- Comandos de teste.
- Artefatos gerados.
- Variaveis de configuracao.
- Riscos operacionais restantes.
