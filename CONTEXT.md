# Contexto de Domínio e Engenharia - SQLMCP

## 1. Visão Geral
O **SQLMCP** é um servidor e biblioteca de alto desempenho escrito em **Go**, projetado para expor recursos e ferramentas do protocolo MCP (*Model Context Protocol*) para agentes de IA interagirem com bancos de dados relacionais corporativos (**PostgreSQL**, **MySQL**, **SQL Server**), com salvaguardas rigorosas de segurança, contextualização de esquema, visualização gráfica em Mermaid ERD e aceleração analítica local via **DuckDB + Parquet**.

## 2. Princípios de Engenharia e Guardrails
- **Read-Only Rigoroso:** Agentes de IA nunca devem mutar dados nos bancos de produção. Apenas comandos de leitura validados sintaticamente (`SELECT`, `EXPLAIN`) são executados.
- **Proteção de Janela de Contexto:** Max rows cap padrão de 500 linhas (teto 2.000). Truncamento de campos de texto extensos (>4KB).
- **Timeouts Atômicos:** `15s` padrão via `context.Context` cancelável.
- **Dual Mode (CLI + Library):** Pode rodar como binário MCP autônomo (Stdio / SSE) ou ser importado diretamente como pacote Go (`pkg/sqlmcp`) por outras aplicações.
- **Camada Analítica Acelerada:** Para tabelas de alta volumetria ou cálculos analíticos pesados, o MCP oferece a exportação para Parquet e consulta local em DuckDB in-process.

## 3. Dialetos Suportados
1. **PostgreSQL** (via `pgx/v5`)
2. **MySQL / MariaDB** (via `go-sql-driver/mysql`)
3. **Microsoft SQL Server** (via `go-mssqldb`)
4. **DuckDB Local** (via `go-duckdb`)
