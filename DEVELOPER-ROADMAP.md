# Roadmap de Desenvolvimento - SQLMCP

Este roadmap documenta o plano de fatiamento técnico e as Epics do projeto SQLMCP.

## Epics

- [x] [**[E01] Fundação do Core e Abstração Multi-Dialeto (PG, MySQL, MSSQL)**](https://github.com/alltomatos/sqlmcp/issues/1) - `done`
- [**[E02] Camada de Segurança Read-Only e Guardrails de Contexto**](https://github.com/alltomatos/sqlmcp/issues/2) - `in_progress`
- [**[E03] Ferramentas MCP de Catálogo, Grafo de Relacionamentos e Mermaid ERD**](https://github.com/alltomatos/sqlmcp/issues/3) - `todo`
- [**[E04] Camada Analítica Acelerada com DuckDB e Parquet**](https://github.com/alltomatos/sqlmcp/issues/4) - `todo`
- [**[E05] Servidor MCP (Stdio/SSE) e Pacote Go Vendorizável**](https://github.com/alltomatos/sqlmcp/issues/5) - `todo`
- **[E06] Governança de Agente: Skill dba-fullstack** - `done`

---

### Slices Planejadas

#### Epic E01: Fundação do Core e Abstração Multi-Dialeto
- [x] [**#6 - S01.1: Estruturar connection pool híbrido com sync.Map e suporte dinâmico**](https://github.com/alltomatos/sqlmcp/issues/6) - `done`
- [x] [**#7 - S01.2: Implementar adaptador PostgreSQL (pgx/v5) para Dialect**](https://github.com/alltomatos/sqlmcp/issues/7) - `done`
- [x] [**#8 - S01.3: Implementar adaptador MySQL para Dialect**](https://github.com/alltomatos/sqlmcp/issues/8) - `done`
- [x] [**#9 - S01.4: Implementar adaptador Microsoft SQL Server para Dialect**](https://github.com/alltomatos/sqlmcp/issues/9) - `done`
- [**#8 - S01.3: Implementar adaptador MySQL para Dialect**](https://github.com/alltomatos/sqlmcp/issues/8) - `todo`
- [**#9 - S01.4: Implementar adaptador Microsoft SQL Server para Dialect**](https://github.com/alltomatos/sqlmcp/issues/9) - `todo`

#### Epic E02: Camada de Segurança Read-Only e Guardrails de Contexto
- **S02.1:** Validador AST léxico de queries para bloqueio de DML/DDL/múltiplos statements.
- **S02.2:** Middleware de limites: timeout estrito (15s default), cap de linhas (padrão 500, máx 2.000) e truncamento de texto longo (>4KB).

#### Epic E03: Ferramentas MCP de Catálogo, Grafo e Mermaid ERD
- **S03.1:** Implementação das ferramentas `list_databases`, `list_tables` e `describe_table`.
- **S03.2:** Implementação da ferramenta `get_relations` (Foreign Keys de entrada e saída).
- **S03.3:** Gerador de diagrama `generate_mermaid_erd` a partir dos metadados.
- **S03.4:** Ferramentas `execute_query`, `explain_query` e `sample_table`.

#### Epic E04: Camada Analítica Acelerada com DuckDB e Parquet
- **S04.1:** Mecanismo de exportação de stream de consulta para Parquet (`stage_to_parquet`).
- **S04.2:** Integração com engine in-process DuckDB (`query_duckdb`) e limpeza/TTL de arquivos temporários.

#### Epic E05: Servidor MCP e Pacote Go Vendorizável
- **S05.1:** Implementação dos handlers `mark3labs/mcp-go` (transporte Stdio e SSE).
- **S05.2:** Exposição do cliente público `pkg/sqlmcp` para importação e vendorização limpa por outros projetos Go.
- **S05.3:** Testes integrados e documentação de uso com Claude Desktop / OpenCode.

#### Epic E06: Governança de Agente: Skill dba-fullstack
- **S06.1:** Implementação da especificação de agente em `.opencode/skills/dba-fullstack/SKILL.md` com workflows de modelagem, tuning de queries, matriz de dialetos e integração com ferramentas do SQLMCP.
