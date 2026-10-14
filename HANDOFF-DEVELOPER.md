# Guia Oficial de Handoff de Arquitetura -> /developer

- **Projeto:** SQLMCP (Go Multi-Dialect Database MCP Bridge)
- **Data:** 2026-10-14
- **Autor:** Arquiteto de Software e Soluções
- **Destinatário:** `/developer` (Orquestrador de Implementação)

---

## 1. Contexto e Visão Geral da Solução
O **SQLMCP** é um servidor e biblioteca de alto desempenho escrito em **Go 1.24+**, projetado para conectar agentes de IA a instâncias corporativas de **PostgreSQL**, **MySQL** e **SQL Server**.

A arquitetura foi inteiramente alinhada sob três pilares inegociáveis:
1. **Segurança 100% Read-Only:** Bloqueio atômico de qualquer comando de escrita (DML/DDL) na camada AST antes de tocar nos bancos de produção.
2. **Proteção da Janela de Contexto de IA:** Timeouts atômicos de 15s, teto máximo de 500 linhas por query (máx 2.000) e truncamento de texto longo (>4KB).
3. **Aceleração Analítica DuckDB + Parquet:** Capacidade de estagiar fatias volumosas de dados para arquivos locais Parquet e rodar consultas agregadas no DuckDB sem onerar o banco de produção.

---

## 2. Mapa dos Artefatos de Arquitetura Disponíveis

| Artefato | Caminho do Arquivo | Descrição |
|---|---|---|
| **ADD Geral** | `docs/architecture/ADD.md` | Documento de design, C4 Context & Container, contratos das 11 tools e estrutura de pastas. |
| **Diagramas de Sequência** | `docs/architecture/sequences.md` | Fluxo detalhado de execução `execute_query` e aceleração `stage_to_parquet + query_duckdb`. |
| **ADR 0001** | `docs/adr/0001-stack-go-mcp-vendorizavel.md` | Justificativa técnica da escolha de Go + `mcp-go` e modelo vendorizável. |
| **ADR 0002** | `docs/adr/0002-seguranca-read-only-guardrails.md` | Regras estritas de validação léxica e limites de contexto. |
| **ADR 0003** | `docs/adr/0003-aceleracao-analitica-duckdb-parquet.md` | Estratégia da engine DuckDB in-process e cache Parquet. |
| **Skill de Domínio** | `.opencode/skills/dba-fullstack/SKILL.md` | Skill do agente com tom executivo para gestores e integração com `/grill-me`. |
| **Contrato Core** | `pkg/dialect/dialect.go` | Interface `Dialect` e structs padronizadas para PG, MySQL, MSSQL e DuckDB. |
| **Roadmap de Epics** | `DEVELOPER-ROADMAP.md` | Fatiamento ordenado das Epics `E01` a `E06`. |

---

## 3. Ordem de Execução Recomendada para o `/developer`

### Fase 1: Fundação do Core e Dialetos (`Epic E01`)
1. Implementar `internal/pool/pool.go` usando `sync.Map` para guardar instâncias `*sql.DB` por alias.
2. Implementar os três adaptadores que satisfazem `pkg/dialect/dialect.go`:
   - `pkg/dialect/postgres.go` (usando `github.com/jackc/pgx/v5/stdlib`)
   - `pkg/dialect/mysql.go` (usando `github.com/go-sql-driver/mysql`)
   - `pkg/dialect/sqlserver.go` (usando `github.com/microsoft/go-mssqldb`)

### Fase 2: Segurança e Guardrails de Leitura (`Epic E02`)
1. Criar `internal/security/validator.go` com verificação de AST de SQL.
2. Rejeitar encadeamento de múltiplos statements (`;`) e bloquear palavras de mutação (`INSERT`, `UPDATE`, `DELETE`, `DROP`, `ALTER`, `TRUNCATE`, `EXEC`).
3. Implementar sanitização e truncamento de linhas/colunas para preservar tokens do LLM.

### Fase 3: Ferramentas Semânticas do Catálogo e Mermaid (`Epic E03`)
1. Implementar inspetor de tabelas, colunas e chaves estrangeiras (`get_relations`).
2. Implementar gerador de texto Mermaid ERD (`generate_mermaid_erd`).
3. Implementar `execute_query`, `explain_query` e `sample_table`.

### Fase 4: Aceleração Analítica DuckDB + Parquet (`Epic E04`)
1. Implementar `stage_to_parquet` gravando arquivos colunares em cache temporário.
2. Implementar `query_duckdb` utilizando `go-duckdb` para consultas vetorizadas locais.

### Fase 5: Servidor MCP e Pacote Público (`Epic E05`)
1. Integrar `github.com/mark3labs/mcp-go` no `cmd/sqlmcp/main.go` expondo as ferramentas registradas via Stdio.
2. Exportar o pacote `pkg/sqlmcp` para permitir importação limpa como biblioteca Go.

---

## 4. Comandos de Verificação & Build
- Compilar binário: `go build -o bin/sqlmcp.exe ./cmd/sqlmcp`
- Executar testes unitários: `go test -v ./...`
- Rodar servidor MCP: `./bin/sqlmcp.exe`
