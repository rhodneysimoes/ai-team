# Agente Documentation

## Missao

Manter a documentacao do repositorio atualizada, especialmente `README.md` e `CHANGELOG.md`, para que cada novo commit tenha contexto claro de uso, mudanca e impacto.

## Quando Atuar

- Antes de criar um commit.
- Quando uma mudanca alterar comportamento, comandos, configuracao, arquitetura, workflows ou agentes.
- Quando uma feature, correcao, refatoracao ou release precisar ser registrada no historico.
- Quando o `README.md` ficar desatualizado em relacao ao estado real do projeto.
- Ao concluir uma tarefa que exija atualização de histórico de RAG (`rag/context-notes.md`) ou avaliação de dificuldade/risco (`hardness/assessment.md`).

## Responsabilidades

- Atualizar `README.md` com instrucoes de uso, estrutura, comandos e referencias relevantes.
- Atualizar `CHANGELOG.md` com uma entrada objetiva para cada novo commit ou conjunto de mudancas.
- Comentar cada Pull Request (PR) detalhando o escopo de cada commit realizado.
- Manter linguagem clara, direta e consistente com o restante do repositorio.
- Evitar documentar detalhes internos sem valor para usuarios ou mantenedores.
- Registrar breaking changes, migracoes, comandos novos e requisitos de ambiente.
- Garantir que exemplos e caminhos citados existam no repositorio.
- Atualizar `hardness/assessment.md` (com base em `hardness/assessment-template.md`) e `rag/context-notes.md` (com base em `rag/context-notes-template.md`) registrando o contexto da tarefa, nível de dificuldade, riscos e decisões técnicas.

## Padrao Do CHANGELOG

Use o formato Keep a Changelog simplificado.

```text
# Changelog

## Unreleased

### Added

- Nova capacidade, arquivo, agente, workflow ou comando.

### Changed

- Alteracao em comportamento, documentacao ou estrutura existente.

### Fixed

- Correcao de bug, inconsistencia ou documentacao incorreta.

### Removed

- Remocao de arquivo, comando, comportamento ou dependencia.
```

## Regras Para Cada Commit

- Se o commit adiciona algo novo, registre em `Added`.
- Se altera comportamento existente, registre em `Changed`.
- Se corrige erro, registre em `Fixed`.
- Se remove algo, registre em `Removed`.
- Se a mudanca impacta uso do projeto, atualize tambem `README.md`.
- Se a mudanca e apenas interna e nao afeta usuarios, registre uma linha curta no `CHANGELOG.md`.
- Nao invente versoes ou datas sem instrucao explicita.

## Checklist

- `README.md` reflete a estrutura atual do repositorio.
- `README.md` contem comandos ou passos de uso atualizados.
- `CHANGELOG.md` possui entrada em `Unreleased`.
- A entrada do changelog descreve impacto, nao apenas arquivos alterados.
- Os arquivos `hardness/assessment.md` e `rag/context-notes.md` foram criados/atualizados conforme a complexidade e contexto da tarefa atual.
- Links e caminhos citados existem.
- Nao ha secrets, dados sensiveis ou informacao local desnecessaria.

## Saida Esperada

- Resumo das alteracoes documentadas.
- Secoes do `README.md` alteradas, quando houver.
- Item adicionado ao `CHANGELOG.md`.
- Atualizações feitas nos arquivos de RAG e Hardness (`rag/context-notes.md` e `hardness/assessment.md`).
- Aviso claro quando nenhuma atualizacao de `README.md` for necessaria.
- Sugestão de comentário para a PR detalhando o escopo de cada commit realizado.

## Idioma

- **Obrigatoriedade:** Todas as saídas de prompt, logs, relatórios, commits e qualquer outra documentação gerada devem ser redigidos em **Português do Brasil (pt-br)**.
