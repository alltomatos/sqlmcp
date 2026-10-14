# ADR 0001: Seleção de Stack e Modelo de Empacotamento (Go + mcp-go)

- **Status:** Aceito
- **Data:** 2026-10-14
- **Autor:** Arquiteto de Software e Soluções

## Contexto
O servidor MCP precisa dar suporte unificado a três bancos relacionais dominantes no mercado corporativo (**PostgreSQL**, **MySQL** e **Microsoft SQL Server**) e atender a dois modos operacionais:
1. Execução como servidor MCP autônomo (via transporte Stdio ou SSE/HTTP) configurado em clientes de IA (OpenCode, Claude Desktop, Cursor, etc.).
2. Capacidade de ser **vendorizado/embutido** diretamente como biblioteca interna (pacote Go) dentro de outras aplicações e ferramentas de engenharia do ecossistema do cliente.

## Decisão
Adotar **Go (Golang 1.23+)** como linguagem base, utilizando a biblioteca **`github.com/mark3labs/mcp-go`** para a implementação do protocolo MCP.

A estrutura interna será organizada em camadas limpas (Clean Architecture / Ports & Adapters):
- Camada de Transporte/Servidor MCP (`internal/transport/mcp`).
- Camada de Aplicação / Casos de Uso (`internal/app`).
- Camada de Domínio e Contratos (`pkg/domain`, `pkg/dialect`).
- Drivers e Adaptadores de Bancos:
  - PostgreSQL: `github.com/jackc/pgx/v5`
  - MySQL: `github.com/go-sql-driver/mysql`
  - SQL Server: `github.com/microsoft/go-mssqldb`
  - Motor Analítico Local: `github.com/marcboeker/go-duckdb` (com suporte Parquet nativo via DuckDB).

## Consequências
### Positivas
- Compilação para binário nativo único, estático, sem dependências de runtime (Node.js/Python).
- Baixo consumo de memória (< 30 MB baseline) e inicialização instantânea (< 100 ms).
- Facilidade imediata para ser importado como módulo Go (`import "github.com/.../sqlmcp/pkg/core"`) por outras ferramentas.
- Concorrência de goroutines eficiente para controle de pools de conexões e streams de resultados.

### Negativas / Mitigações
- DuckDB em Go utiliza bindings CGO (`go-duckdb`), o que exige suporte a compilador C (gcc/clang) durante o build binário ou compilação com tags opcionais (`-tags no_duckdb` caso o host queira vendorização puramente pura em Go).
