# Time de Agentes para Desenvolvimento Go

Este repositorio contem definicoes para executar um time de agentes especializados em desenvolvimento Go. Use este arquivo como ponto de entrada do orquestrador e os arquivos em `agents/` como perfis de atuacao.

## Objetivo

Entregar software Go simples, testavel, idiomatico e pronto para manutencao, com atencao a arquitetura, implementacao, testes, revisao, seguranca e operacao.

## Agentes

- `agents/orchestrator.md`: coordena o trabalho, quebra tarefas e decide quais especialistas acionar.
- `agents/go-architect.md`: define arquitetura, modulos, contratos, boundaries e decisoes tecnicas.
- `agents/go-developer.md`: implementa codigo Go idiomatico e integrado ao projeto.
- `agents/go-tester.md`: cria e executa testes unitarios, integracao, benchmarks e validacoes.
- `agents/go-reviewer.md`: revisa codigo, riscos, regressoes, legibilidade e aderencia a padroes Go.
- `agents/go-security.md`: avalia seguranca, validacao de entrada, dependencias e dados sensiveis.
- `agents/go-devops.md`: cuida de build, CI, Docker, releases e observabilidade.
- `agents/documentation.md`: atualiza `README.md` e `CHANGELOG.md` a cada novo commit.
- `sdd/go-sdd.md`: define o Software Design Document base para projetos Go.
- `rag/`: estrutura de Retrieval Augmented Generation para organizar fontes, consultas e contexto recuperado.
- `hardness/`: matriz para classificar dificuldade, risco e estrategia de execucao das tarefas.

## Fluxo Padrao

1. O orquestrador le a solicitacao e identifica o tipo de tarefa.
2. Classifica a tarefa usando `hardness/hardness-matrix.md`.
3. Recupera contexto relevante usando `rag/rag-playbook.md` quando houver decisao dependente de conhecimento do projeto.
4. O arquiteto atua quando houver mudanca estrutural, API publica, persistencia, concorrencia ou integracao externa.
5. O desenvolvedor implementa a solucao seguindo os padroes existentes do repositorio.
6. O tester adiciona ou ajusta testes proporcionais ao risco da alteracao.
7. O reviewer faz revisao final antes da entrega.
8. O agente de documentacao atualiza `README.md` e `CHANGELOG.md` quando a mudanca alterar uso, comportamento, arquitetura, comandos ou artefatos.
9. Seguranca e DevOps entram quando a tarefa tocar autenticacao, autorizacao, secrets, rede, build, deploy ou infraestrutura.

## Padroes Go

- Use `gofmt` e `go test ./...` como verificacao minima.
- Prefira APIs pequenas, explicitas e faceis de testar.
- Propague erros com contexto usando `fmt.Errorf("contexto: %w", err)`.
- Mantenha interfaces no pacote consumidor quando isso simplificar testes e boundaries.
- Evite abstracoes prematuras e estado global mutavel.
- Use `context.Context` para operacoes de I/O, rede, banco e chamadas potencialmente longas.
- Prefira testes table-driven para cenarios com variacao clara.

## Como Usar

Para uma tarefa comum, comece pelo `agents/orchestrator.md` e siga o fluxo em `workflows/go-feature.md`.

Para revisoes, use `workflows/go-review.md`.

Para preparar build, CI ou entrega, use `workflows/go-release.md`.

Para orientar decisoes de desenho e qualidade em Go, use `sdd/go-sdd.md`.

Para tarefas com contexto espalhado no repositorio, use `rag/rag-playbook.md`.
 
Para calibrar dificuldade, risco e profundidade de verificacao, use `hardness/hardness-matrix.md`.

## Idioma

- **Obrigatoriedade:** Todas as saídas de prompts, explicações, planos de execução, relatórios, commits e qualquer outra comunicação gerada pelos agentes devem ser obrigatoriamente escritas em **Português do Brasil (pt-br)**.
