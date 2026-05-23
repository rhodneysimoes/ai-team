# Sinais De Risco

Use estes sinais para elevar a classificacao de Hardness.

## Elevar Para H2 Ou Mais

- Mudanca altera comportamento, nao apenas texto.
- Arquivo alterado e usado por mais de um pacote.
- Falta teste cobrindo o comportamento tocado.

## Elevar Para H3 Ou Mais

- Mudanca toca multiplos pacotes.
- Envolve persistencia, rede, filesystem ou processos externos.
- Altera configuracao de runtime.
- Introduz dependencia nova.
- Requer compatibilidade com comportamento antigo.

## Elevar Para H4

- Altera autenticacao, autorizacao, criptografia ou secrets.
- Envolve migracao de dados.
- Altera API publica ou contrato externo.
- Usa concorrencia com estado compartilhado.
- Afeta deploy, CI, release ou rollback.
- Pode causar perda de dados, indisponibilidade ou vazamento de informacao.

## Sinais De Baixa Confianca

- Nao ha testes existentes.
- Documentacao contradiz codigo.
- O comportamento esperado depende de regra implicita.
- O impacto nao pode ser confirmado localmente.

Quando houver baixa confianca, use RAG com notas e registre lacunas.
