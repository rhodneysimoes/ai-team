# Avaliacao De Hardness: Implementação de Preços, Evasão de Bloqueio e Configuração de Agentes

## Tarefa

Adicionar extração de preços ao scraper, implementar fallback headless via chromedp para evitar bloqueios do Cloudflare, configurar o token gh para pull requests, e atualizar as instruções dos agentes para consultarem e manterem os arquivos RAG e Hardness atualizados.

## Nivel

H2 (Médio)

## Justificativa

- **Escopo:** Alterações na struct core de promoções, adição de regex de busca de preços, integração com biblioteca externa (chromedp), modificação de regras de prompts de agentes.
- **Impacto:** Alta confiabilidade no bypass do Cloudflare e no enriquecimento de dados de promoções.
- **Riscos:** Baixo, pois a chamada do navegador é configurada como um fallback (self-healing) quando a requisição HTTP comum falha, não impactando a performance dos fluxos normais que já funcionam.
- **Reversibilidade:** Totalmente reversível, bastando desativar a chamada do helper de browser em `collectSite`.

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
