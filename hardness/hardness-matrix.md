# Matriz De Hardness

## Objetivo

Classificar uma tarefa antes da execucao para definir profundidade de analise, agentes necessarios e verificacoes.

## Escala

### H1 - Simples

Mudanca pequena, localizada e de baixo risco.

Exemplos:

- Ajuste de texto.
- Pequena correcao em documentacao.
- Alteracao em arquivo isolado sem impacto de comportamento.

Conduta:

- Pouco RAG.
- Um agente principal.
- Verificacao simples.

### H2 - Moderada

Mudanca funcional pequena ou media, com impacto limitado.

Exemplos:

- Nova funcao em pacote existente.
- Ajuste de teste.
- Pequena mudanca de configuracao.

Conduta:

- RAG focado nos arquivos relacionados.
- Developer e Tester.
- Reviewer quando houver comportamento novo.

### H3 - Complexa

Mudanca com impacto em multiplos arquivos, contratos ou fluxo de execucao.

Exemplos:

- Nova feature com testes.
- Mudanca em API interna.
- Integracao com banco, fila, HTTP ou filesystem.

Conduta:

- RAG com notas.
- Architect, Developer, Tester e Reviewer.
- Security ou DevOps quando houver superficie correspondente.

### H4 - Critica

Mudanca de alto risco, ampla ou dificil de reverter.

Exemplos:

- Alteracao de API publica.
- Migracao de dados.
- Concorrencia complexa.
- Autenticacao, autorizacao ou criptografia.
- Deploy, CI ou release de producao.

Conduta:

- RAG obrigatorio com fontes e confianca.
- Architect obrigatorio.
- Tester e Reviewer obrigatorios.
- Security e DevOps conforme impacto.
- Verificacoes ampliadas.

## Fatores De Classificacao

Avalie:

- Escopo de arquivos.
- Impacto em comportamento.
- Impacto em usuarios ou integracoes.
- Reversibilidade.
- Necessidade de migracao.
- Concorrencia.
- Seguranca.
- Dependencias externas.
- Cobertura de testes existente.

## Verificacoes Por Nivel

- H1: leitura final e validacao local quando aplicavel.
- H2: testes relevantes e revisao leve.
- H3: `go test ./...`, testes novos ou ajustados e revisao.
- H4: `go test ./...`, `go test -race ./...` quando relevante, `go vet ./...`, revisao de seguranca ou operacao.
