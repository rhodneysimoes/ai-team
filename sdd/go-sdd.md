# SDD: Desenvolvimento Go

## 1. Proposito

Este Software Design Document define diretrizes para desenvolvimento de software em Go neste time de agentes. Ele deve ser usado para orientar arquitetura, implementacao, testes, revisao, seguranca e operacao de projetos Go.

## 2. Principios

- Simplicidade antes de abstracao.
- Codigo idiomatico, formatado com `gofmt` e facil de ler.
- Pacotes pequenos com responsabilidade clara.
- APIs explicitas, estaveis e testaveis.
- Erros tratados de forma visivel e contextual.
- Concorrencia usada com cancelamento, limites e ownership claro.
- Testes proporcionais ao risco e ao comportamento entregue.
- Configuracao externa ao codigo e sem secrets versionados.

## 3. Estrutura De Projeto

Use a estrutura mais simples que resolva o problema. Evite criar diretorios padrao sem necessidade real.

Estrutura recomendada para servicos:

```text
.
|-- cmd/
|   `-- app/
|       `-- main.go
|-- internal/
|   |-- config/
|   |-- domain/
|   |-- service/
|   `-- transport/
|-- pkg/
|-- migrations/
|-- testdata/
|-- go.mod
`-- go.sum
```

Estrutura recomendada para ferramentas de CLI ou scripts:

```text
.
|-- cmd/
|   `-- [nome-da-ferramenta]/
|       `-- main.go
|-- internal/
|   |-- [pacote-dominio-1]/
|   `-- [pacote-dominio-2]/
|-- go.mod
`-- go.sum
```

Regras:

- `cmd/` contem pontos de entrada.
- `internal/` contem codigo privado da aplicacao.
- `pkg/` so deve existir para bibliotecas realmente reutilizaveis por consumidores externos.
- `testdata/` deve guardar fixtures de teste.
- Evite pacotes chamados `utils`, `common` ou `helpers`; prefira nomes ligados ao dominio.

## 4. Design De Pacotes

- Um pacote deve ter um motivo claro para mudar.
- Nome de pacote deve ser curto, minusculo e sem underscore.
- Evite ciclos de dependencia.
- Coloque interfaces no pacote consumidor quando elas existirem para desacoplar uso.
- Exporte apenas o que precisa ser consumido fora do pacote.
- Mantenha construtores pequenos e validando invariantes importantes.

Exemplo:

```go
type Store interface {
    FindUser(ctx context.Context, id string) (User, error)
}
```

A interface acima deve ficar no pacote que consome `Store`, nao necessariamente no pacote que implementa banco de dados.

## 5. APIs E Contratos

- Prefira parametros explicitos em vez de mapas genericos.
- Use structs de entrada quando houver muitos parametros ou evolucao esperada.
- Retorne tipos concretos quando nao houver motivo para esconder implementacao.
- Documente comportamento de funcoes publicas quando o pacote for consumido externamente.
- Preserve compatibilidade de APIs publicas ou registre migracao.

## 6. Tratamento De Erros

- Nunca ignore erro sem justificativa clara.
- Use `fmt.Errorf("contexto: %w", err)` para preservar a causa.
- Use `errors.Is` e `errors.As` para comparar ou extrair erros.
- Evite logs duplicados em cada camada; logue no limite da aplicacao ou onde houver decisao operacional.
- Diferencie erros de dominio, validacao, infraestrutura e permissao.

Padrao recomendado:

```go
value, err := repo.Load(ctx, id)
if err != nil {
    return Result{}, fmt.Errorf("load user %q: %w", id, err)
}
```

## 7. Contexto, Timeout E Cancelamento

- Receba `context.Context` como primeiro parametro em operacoes de I/O, rede, banco, fila ou processamento longo.
- Nao armazene `context.Context` em structs.
- Respeite `ctx.Done()` em loops e goroutines.
- Defina timeouts em chamadas externas.
- Nao use `context.Background()` dentro de regras de negocio; ele deve aparecer principalmente nas bordas da aplicacao.

## 8. Concorrencia

- Use goroutines apenas quando houver ganho claro.
- Toda goroutine deve ter caminho de encerramento.
- Controle fan-out com limites de paralelismo.
- Proteja estado compartilhado com mutex, canais ou ownership unico.
- Evite canais globais.
- Use `go test -race ./...` quando houver concorrencia ou estado compartilhado.

Checklist de concorrencia:

- Quem inicia a goroutine tambem sabe como encerra-la.
- Erros sao coletados e propagados.
- Cancelamento encerra trabalho pendente.
- Canais sao fechados pelo produtor responsavel.

## 9. Configuracao

- Configure por variaveis de ambiente, arquivo externo ou secret manager.
- Valide configuracao no startup.
- Tenha defaults somente para valores seguros e previsiveis.
- Nao versionar credenciais reais.
- Separe configuracao de ambiente, dominio e infraestrutura.

Campos comuns:

- `APP_ENV`
- `LOG_LEVEL`
- `HTTP_ADDR`
- `DATABASE_URL`
- `REQUEST_TIMEOUT`

## 10. Logs E Observabilidade

- Logs devem ter contexto operacional suficiente.
- Nao logar secrets, tokens, senhas, documentos ou payloads sensiveis.
- Use logs estruturados quando o projeto suportar.
- Inclua request id, user id interno ou correlation id quando disponivel.
- Servicos devem expor health check quando forem implantados como processo long-running.

## 11. Persistencia

- Use transacoes para operacoes atomicas.
- Defina limites de timeout para queries.
- Use queries parametrizadas.
- Evite acoplar regras de negocio diretamente a detalhes do banco.
- Migrations devem ser versionadas e revisaveis.
- Testes de repositorio devem usar banco controlado, fixture pequena ou fake bem delimitado.

## 12. HTTP E APIs Externas

- Defina timeout no servidor e no cliente HTTP.
- Valide entrada antes de chamar regras de negocio.
- Retorne codigos HTTP coerentes com o erro.
- Nao exponha erros internos diretamente ao cliente.
- Use middlewares para cross-cutting concerns: log, recover, auth, tracing e request id.

## 13. Seguranca

- Validar toda entrada externa.
- Escapar ou parametrizar dados em SQL, templates e comandos.
- Evitar executar comandos de sistema com entrada do usuario.
- Usar TLS em comunicacao externa quando aplicavel.
- Guardar secrets fora do repositorio.
- Revisar dependencias novas.
- Rodar `govulncheck ./...` quando disponivel.

## 14. Testes

Pirâmide recomendada:

- Unitarios para regras de negocio e tratamento de erro.
- Integracao para banco, filesystem, rede local ou adapters relevantes.
- End-to-end apenas para fluxos criticos.

Padroes:

- Use testes table-driven para variacoes de entrada e saida.
- Nomeie testes por comportamento.
- Evite depender de horario real; injete clock quando necessario.
- Evite rede externa em testes automatizados.
- Use `t.Helper()` em helpers de teste.
- Use `t.TempDir()` para arquivos temporarios.
- Guarde fixtures em `testdata/`.

Comandos:

```powershell
go test ./...
go test -race ./...
go test -cover ./...
```

## 15. Qualidade E Ferramentas

Verificacoes minimas:

```powershell
gofmt -w .
go test ./...
go vet ./...
```

Verificacoes adicionais quando disponiveis:

```powershell
govulncheck ./...
staticcheck ./...
golangci-lint run
```

Regras:

- `go.mod` e `go.sum` devem permanecer consistentes.
- Execute `go mod tidy` quando dependencias mudarem.
- Builds devem ser reproduziveis no CI.

## 16. Performance

- Otimize apenas com evidencia.
- Use benchmarks para mudancas sensiveis a desempenho.
- Evite alocacoes desnecessarias em caminhos quentes quando mensurado.
- Prefira clareza ate que profiling indique gargalo.
- Use `pprof` para investigacoes reais.

## 17. Criterios De Pronto

Uma tarefa Go esta pronta quando:

- Codigo esta formatado com `gofmt`.
- Testes relevantes foram criados ou atualizados.
- `go test ./...` passa.
- `go vet ./...` passa quando aplicavel.
- Concorrencia foi validada com `go test -race ./...` quando relevante.
- Erros tem contexto e nao escondem causa.
- Configuracao sensivel nao esta no codigo.
- Mudancas de API, contrato ou migracao foram documentadas.

## 18. Papel Dos Agentes

- `go-architect`: valida desenho, boundaries e contratos.
- `go-developer`: implementa seguindo este SDD.
- `go-tester`: garante cobertura proporcional ao risco.
- `go-reviewer`: revisa bugs, regressao e aderencia.
- `go-security`: revisa superficie de seguranca.
- `go-devops`: valida build, CI, release e operacao.

## 19. Template De Decisao Tecnica

Use este template para decisoes relevantes:

```text
Titulo:
Contexto:
Decisao:
Alternativas consideradas:
Trade-offs:
Impacto em testes:
Impacto operacional:
```

## 20. Checklist Rapido Para PR

- [ ] O pacote e o nome dos simbolos sao claros.
- [ ] Erros sao tratados e propagados com contexto.
- [ ] Nao ha secrets ou dados sensiveis em codigo, teste ou log.
- [ ] Entradas externas sao validadas.
- [ ] Goroutines tem cancelamento e coleta de erro.
- [ ] Recursos sao fechados.
- [ ] Testes cobrem sucesso e falhas importantes.
- [ ] `gofmt`, `go test ./...` e checks relevantes foram executados.
