# Changelog

Todas as mudancas relevantes deste repositorio devem ser documentadas neste arquivo.

## Unreleased

### Added

- Adicionados headers padrao de navegador, headers customizaveis por site e diagnostico de bloqueio HTTP no scraper.
- Adicionado suporte a `thumbnail_url` no resultado do scraper a partir de tags thumbnail, `og:image` ou `twitter:image`.
- Adicionado CLI `cmd/promoscraper` para coletar promocoes de sites definidos em `sites.md`.
- Adicionados pacotes internos `internal/sites` e `internal/scraper` com testes unitarios.
- Adicionado `sites.md` como arquivo de configuracao dos sites monitorados.
- Adicionado `go.mod` para o projeto Go.
- Adicionada estrutura `rag/` com playbook de recuperacao de contexto, fontes e template de notas.
- Adicionada estrutura `hardness/` com matriz de dificuldade, sinais de risco e template de avaliacao.
- Adicionado agente `documentation` para manter `README.md` e `CHANGELOG.md` atualizados a cada novo commit.
- Adicionado `CHANGELOG.md` como historico central de mudancas do repositorio.

### Changed

- Atualizado parser de headers em `sites.md` para aceitar `\;` em valores como `Accept-Language`.
- Atualizado o padrao da Terabyte em `sites.md` com sinais reais da pagina: percentual OFF, Tera Maio, Termina em, Frete gratis, Mais vendido e preco De/por.
- Atualizado parser de `sites.md` para aceitar `\|` em regex dentro de tabelas Markdown.
- Atualizado `sites.md` com Kabum, Pichau e Terabyte Shop como sites habilitados para coleta de promocoes.
- Atualizado `README.md` com instrucoes de uso do scraper de promocoes.
- Atualizado `AGENT.md` e `README.md` para incluir RAG e Hardness no fluxo de execucao.
- Atualizado `AGENT.md`, `README.md` e `agents/orchestrator.md` para incluir o fluxo de documentacao antes de commits.

### Fixed

- Corrigido parser de `sites.md` para ignorar tabelas dentro de blocos de codigo Markdown.

### Removed
