# Avaliacao De Hardness: Implementação de Preços, Evasão de Bloqueio e Configuração de Agentes

## Tarefa

Adicionar extração de preços ao scraper, implementar fallback headless via chromedp para evitar bloqueios do Cloudflare, configurar o token gh para pull requests, atualizar as instruções dos agentes para consultarem e manterem os arquivos RAG/Hardness atualizados, forçar a obrigatoriedade de títulos e descrições de PRs em Português do Brasil (pt-br), ativar a coleta do site da Kabum em sites.md e filtrar promoções sem preço no arquivo de saída promotions.json.

## Nivel

H2 (Médio)

## Justificativa

- **Escopo:** Alterações na struct core de promoções, adição de regex de busca de preços, integração com biblioteca externa (chromedp), modificação de regras de prompts de agentes para RAG/Hardness, restrições de idioma nos PRs e ativação de site em sites.md.
- **Impacto:** Alta confiabilidade no bypass do Cloudflare, enriquecimento de dados de promoções, conformidade idiomática dos PRs, maior contextualização contínua nos agentes e expansão da cobertura de lojas ativas.
- **Riscos:** Baixo.
- **Reversibilidade:** Totalmente reversível.

## Agentes Necessarios

- Orchestrator: Sim
- Architect: Não
- Developer: Sim
- Tester: Sim
- Reviewer: Sim
- Security: Não
- DevOps: Não
- Documentation: Sim

## RAG Necessario

Sim.

Fontes provaveis:

- `internal/scraper/scraper.go`
- `internal/scraper/scraper_test.go`
- `agents/orchestrator.md`
- `agents/documentation.md`

## Verificacoes

- `go test ./...` (Sucesso na compilação e teste unitário simulando bloqueio)
- `go run cmd/promoscraper/main.go -sites sites.md -output promotions.json` (Sucesso na coleta contornando Cloudflare da Pichau)

## Lacunas

Nenhuma incerteza pendente.
