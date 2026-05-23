# Notas RAG: Estratégia de Evasão de Bloqueio, Parsing de Preços e Gestão de Contexto dos Agentes

## Pergunta

Como contornar bloqueios do Cloudflare de forma robusta e obter informações de preço sem comprometer a simplicidade e a performance padrão do scraper Go?

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

## Decisao Influenciada

- Adotou-se o uso do `chromedp` de forma condicional como um fallback de auto-recuperação (self-healing), preservando a requisição HTTP comum como padrão rápido e de baixo consumo de recursos.
- Configuração de instruções de agentes para manter `hardness/assessment.md` e `rag/context-notes.md` sincronizados no fim do processo de documentação.
- Padronização de requisitos idiomáticos no Orchestrator e na Documentation abrangendo títulos e descrições de PRs.

## Confianca

Alta.

## Lacunas

Nenhuma.
