# Notas RAG: Estratégia de Evasão de Bloqueio, Parsing de Preços e Gestão de Contexto dos Agentes

## Pergunta

Como contornar bloqueios do Cloudflare de forma robusta, obter informações de preço sem comprometer a simplicidade e a performance padrão do scraper Go, evitar registros sem preço (incompletos) no arquivo promotions.json, estender a coleta recursiva para todas as subpáginas internas (Nível 2), garantir que o projeto e o SDD estejam alinhados, e criar um agrupador e comparador de menor/maior preço por produto?

## Fontes Consultadas

- `internal/scraper/scraper.go`: Estrutura do loop de coleta, cabeçalhos HTTP e payload Next.js.
- `internal/scraper/scraper_test.go`: Testes unitários para Next.js, cabeçalhos e tratamento de bloqueios.
- `go.mod`: Módulos de terceiros necessários (adição do chromedp).
- `agents/orchestrator.md` e `agents/documentation.md`: Diretrizes operacionais e ciclo de vida de documentação da equipe.

## Evidencias

- **Assinatura TLS (JA3):** O Cloudflare detecta a assinatura TLS padrão do Go `net/http` e retorna erro `403 Forbidden` mesmo fornecendo cabeçalhos reais.
- **Navegador Headless:** A navegação real usando a biblioteca Go `chromedp` com opções apropriadas (`--headless`, `--no-sandbox`, etc.) gerou respostas corretas com status 200 no site Pichau.
- **Regex para Preços:** As expressões regulares `[0-9]+(?:[.,][0-9]+)*` mostraram-se seguras para extrair preços com separadores tanto no formato brasileiro (`R$ 1.899,90`) quanto no americano (`1499.99`).
- **Redução de Análise:** Ao consultarem as notas estruturadas de RAG e avaliações de Hardness existentes, os agentes reduzem o escopo de análise em tarefas subsequentes.
- **Idioma dos PRs:** A imposição expressa nos prompts dos agentes garante que as contribuições e revisões de código de toda a equipe de agentes gerem PRs com títulos e descrições uniformemente localizados em Português do Brasil (pt-br).
- **Ativação da Kabum:** O site da Kabum foi reabilitado no fluxo de scraping através da modificação da coluna `enabled` em `sites.md`.
- **Filtro de Preço no Output:** A inserção de promoções sem preço (Price <= 0) no JSON de saída causa inconsistências na base de dados. Um filtro adicionado no CLI principal (`cmd/promoscraper/main.go`) remove os itens que não tiveram preço detectado ou que possuem valor inválido.
- **Scraping de Nível 1 e 2:** A coleta recursiva aumenta significativamente a quantidade de promoções encontradas. As subpáginas internas são identificadas pelas tags `<a href="...">` no HTML da página inicial (Nível 1), filtrando-se arquivos estáticos e apenas URLs com o mesmo host/domínio do site de origem (links internos).
- **Concurrency Worker Pool:** O fetch simultâneo das subpáginas do Nível 2 é executado de forma concorrente em cada site por meio de um pool com limite de 5 workers para otimizar velocidade e gerenciar o tempo limite geral de forma limpa.
- **Alinhamento SDD vs Código:** Identificamos que erros de parsing de URL no pacote `sites` não usavam o operador `%w` recomendado no SDD, e o pool de workers concorrente do Nível 2 continuava consumindo a fila mesmo após o cancelamento do contexto.
- **Cancelamento Imediato Concorrente:** A verificação de `ctx.Done()` na fila concorrente de Nível 2 evita consumo desnecessário de CPU após timeouts.
- **Normalização de Textos:** Para agrupar produtos descritos com ligeiras variações, desenvolvemos um algoritmo de limpeza (removendo tags de marketing, cupons, preços antigos) e geramos chaves de comparação a partir das palavras ordenadas alfabeticamente. Isso provou-se altamente eficaz para unificar itens semelhantes de fontes diferentes.

## Decisao Influenciada

- Adotou-se o uso do `chromedp` de forma condicional como um fallback de auto-recuperação (self-healing), preservando a requisição HTTP comum como padrão rápido e de baixo consumo de recursos.
- Configuração de instruções de agentes para manter `hardness/assessment.md` e `rag/context-notes.md` sincronizados no fim do processo de documentação.
- Padronização de requisitos idiomáticos no Orchestrator e na Documentation abrangendo títulos e descrições de PRs.
- Habilitação da coleta concorrente de todos os três grandes e-commerces (Kabum, Pichau e Terabyte Shop) simultaneamente no fluxo principal.
- Implementação da filtragem de preços na camada de persistência em arquivo do CLI, mantendo o parseador resiliente para testes unitários com mocks sem preço, mas preservando apenas itens precificados válidos na saída final.
- Refatoração de `collectSite` separando a lógica básica de fetch em `fetchAndExtract`. Implementação de extrator de links, pool de concorrência com WaitGroup e deduplicação unificada de ofertas encontradas nas fases de Nível 1 e Nível 2.
- Inclusão da estrutura recomendada de projetos CLI e utilitários na Seção 3 do SDD para refletir o layout do repositório.
- Refatoração da função `parseDefinition` para encapsular erros de URL com `%w`.
- Inclusão de verificação `select { case <-ctx.Done(): return; default: }` na goroutine dos workers de nível 2.
- Criação do utilitário `cmd/pricecomparator/main.go` para agrupar promoções de forma order-independent e gerar deterministamente os arquivos `menor_preco.json` e `maior_preco.json`.

## Confianca

Alta.

## Lacunas

Nenhuma.
