# Fontes RAG

## Prioridade De Fontes

1. Codigo existente no repositorio.
2. Testes existentes.
3. `sdd/` e documentos de design.
4. `workflows/` e perfis em `agents/`.
5. `README.md` e `CHANGELOG.md`.
6. Issues, notas externas ou documentacao de terceiros, quando fornecidas pelo usuario.

## Fontes Fortes

- Implementacoes atuais.
- Testes que codificam comportamento esperado.
- Interfaces publicas.
- Migrations e contratos de API.
- Configuracoes de build e CI.

## Fontes Fracas

- Comentarios antigos sem cobertura de teste.
- TODOs sem contexto.
- Exemplos desatualizados.
- Suposicoes nao verificadas.

## Como Registrar Confianca

Use tres niveis:

- Alta: confirmada por codigo e teste.
- Media: confirmada por codigo ou documentacao atual.
- Baixa: inferencia sem confirmacao direta.
