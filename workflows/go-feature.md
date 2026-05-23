# Workflow: Funcionalidade Go

Use este fluxo para criar ou alterar funcionalidades em projetos Go.

## Etapas

1. Orquestrador identifica objetivo, arquivos relevantes e criterio de pronto.
2. Architect define desenho somente se houver impacto estrutural.
3. Developer implementa a menor mudanca correta.
4. Tester adiciona ou atualiza testes.
5. Reviewer revisa comportamento, riscos e cobertura.
6. Developer corrige achados relevantes.
7. Orquestrador executa verificacoes e resume a entrega.

## Verificacoes

- `gofmt` nos arquivos alterados.
- `go test ./...`
- `go test -race ./...` quando houver concorrencia.
- `go vet ./...` quando a mudanca tocar APIs publicas, build ou comportamento delicado.

## Saida Final

- Arquivos alterados.
- Comportamento entregue.
- Testes executados.
- Riscos ou limitacoes restantes.
