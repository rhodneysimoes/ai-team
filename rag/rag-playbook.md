# RAG Playbook

## Objetivo

Recuperar contexto confiavel do repositorio antes de planejar, implementar ou revisar uma tarefa.

## Principios

- Priorize contexto local do repositorio antes de conhecimento generico.
- Prefira fontes primarias: codigo, testes, SDD, workflows e documentacao mantida.
- Registre apenas o contexto que influencia uma decisao.
- Nao use RAG como desculpa para leitura infinita; busque o suficiente para agir com seguranca.

## Etapas

1. Defina a pergunta de contexto.
2. Liste palavras-chave, pacotes, comandos, tipos ou arquivos provaveis.
3. Busque arquivos com `rg --files`.
4. Busque usos e referencias com `rg`.
5. Leia os arquivos mais relevantes.
6. Extraia regras, contratos, exemplos e excecoes.
7. Registre notas usando `context-notes-template.md` quando a tarefa for complexa.
8. Use o contexto recuperado para escolher agentes, plano e verificacoes.

## Consultas Recomendadas

```powershell
rg --files
rg -n "NomeDoTipo|Funcao|Comando|Erro|Config"
rg -n "TODO|FIXME|Deprecated|BREAKING|go test|gofmt"
```

## Saida Esperada

- Pergunta respondida.
- Fontes consultadas.
- Decisoes influenciadas pelo contexto.
- Lacunas ou incertezas restantes.

## Regras De Parada

Pare a recuperacao quando:

- O padrao local estiver claro.
- Os arquivos de entrada, saida e teste estiverem identificados.
- Os riscos principais estiverem mapeados.
- Continuar lendo nao mudar a decisao provavel.
