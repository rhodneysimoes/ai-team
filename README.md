# AI Team Go

Definicoes de agentes especializados para desenvolvimento em Go.

## Estrutura

- `AGENT.md`: ponto de entrada e regras gerais do time.
- `agents/`: perfis dos agentes especializados.
- `workflows/`: fluxos de execucao para funcionalidades, revisoes e releases.
- `sdd/`: documentos de desenho de software e padroes tecnicos.
- `rag/`: estrutura para recuperar contexto tecnico antes de decidir ou implementar.
- `hardness/`: criterios para classificar dificuldade, risco e profundidade de execucao.
- `CHANGELOG.md`: historico de mudancas mantido pelo agente de documentacao.

## Uso Rapido

1. Leia `AGENT.md`.
2. Escolha um workflow em `workflows/`.
3. Consulte `sdd/go-sdd.md` para alinhar decisoes tecnicas em Go.
4. Classifique a tarefa com `hardness/hardness-matrix.md`.
5. Recupere contexto com `rag/rag-playbook.md` quando a tarefa depender de conhecimento do projeto.
6. Acione os agentes em `agents/` conforme a necessidade da tarefa.
7. Antes de cada commit, acione `agents/documentation.md` para revisar `README.md` e `CHANGELOG.md`.

## Verificacoes Go Recomendadas

```powershell
gofmt -w .
go test ./...
go vet ./...
```
