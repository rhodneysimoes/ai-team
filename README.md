# AI Team Go

Definicoes de agentes especializados para desenvolvimento em Go.

## Estrutura

- `AGENT.md`: ponto de entrada e regras gerais do time.
- `agents/`: perfis dos agentes especializados.
- `cmd/promoscraper/`: CLI para coletar promocoes a partir dos sites definidos em `sites.md`.
- `internal/`: pacotes Go internos do scraper.
- `workflows/`: fluxos de execucao para funcionalidades, revisoes e releases.
- `sdd/`: documentos de desenho de software e padroes tecnicos.
- `rag/`: estrutura para recuperar contexto tecnico antes de decidir ou implementar.
- `hardness/`: criterios para classificar dificuldade, risco e profundidade de execucao.
- `sites.md`: lista de sites, regex de promocao e flag de ativacao.
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

## Promo Scraper

O projeto inclui um CLI em Go para coletar promocoes de uma lista de sites definida em `sites.md`.
Quando a pagina contem tags de thumbnail, imagens `og:image` ou `twitter:image`, o JSON inclui `thumbnail_url`.

Edite `sites.md` usando o formato de tabela:

```markdown
| name | url | pattern | enabled |
| --- | --- | --- | --- |
| Loja Exemplo | https://example.com/promocoes | (?i)(promocao\|oferta\|desconto).{0,120} | true |
```

Em tabelas Markdown, use `\|` para alternancia de regex dentro da coluna `pattern`.
Opcionalmente, adicione uma coluna `headers` com pares `Nome=Valor` separados por `;` para ajustar requisicoes por site. Use `\;` para ponto e virgula literal dentro de um valor.
O scraper ja envia headers padrao parecidos com navegador e marca respostas bloqueadas com `blocked` e `block_reason`.

Execute imprimindo JSON no terminal:

```powershell
go run ./cmd/promoscraper -sites sites.md
```

Execute gravando em arquivo:

```powershell
go run ./cmd/promoscraper -sites sites.md -output promotions.json
```

Flags disponiveis:

- `-sites`: caminho para o arquivo Markdown de sites.
- `-output`: caminho opcional para salvar o JSON.
- `-timeout`: timeout total das requisicoes.
- `-concurrency`: quantidade maxima de sites coletados em paralelo.
