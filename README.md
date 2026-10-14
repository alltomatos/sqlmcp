# SQLMCP - Multi-Dialect Database MCP Bridge

[![Go Version](https://img.shields.io/badge/go-1.24%2B-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Protocol](https://img.shields.io/badge/MCP-Model%20Context%20Protocol-purple.svg)](https://modelcontextprotocol.io)

O **SQLMCP** é um servidor e biblioteca de alto desempenho escrito em **Go**, projetado para conectar agentes de Inteligência Artificial (LLMs) a bancos de dados relacionais corporativos (**PostgreSQL**, **MySQL** e **Microsoft SQL Server**) com **salvaguardas rigorosas de segurança**, **proteção de janela de contexto**, **visualização gráfica em Mermaid ERD** e **aceleração analítica local via DuckDB + Parquet**.

---

## 🚀 Principais Recursos

1. **Segurança 100% Read-Only:** Bloqueio preventivo na camada de AST de comandos destrutivos ou de mutação (`INSERT`, `UPDATE`, `DELETE`, `DROP`, `ALTER`, `TRUNCATE`, `EXEC`) e de múltiplos statements (`;`).
2. **Proteção da Janela de Contexto (Tokens):** Timeouts padrão de 15 segundos, cap de linhas (padrão 500, máx 2.000) e truncamento automático de colunas extensas (>4KB).
3. **Multi-Dialeto Unificado:** Suporte nativo para PostgreSQL, MySQL/MariaDB e Microsoft SQL Server sob a mesma interface.
4. **Visualização Arquitetural Automática:** Geração instantânea de diagramas de entidade-relacionamento em formato visual **Mermaid ERD**.
5. **Aceleração Analítica DuckDB + Parquet:** Capacidade de estagiar fatias analíticas de tabelas volumosas (>1M de linhas) em arquivos Parquet compactos e consultá-las com DuckDB em memória sem derrubar o banco de produção.
6. **Dual Mode (CLI + Library):** Pode ser executado como binário autônomo (transporte Stdio/SSE) ou ser importado diretamente como pacote Go (`pkg/sqlmcp`) por outras aplicações.
7. **Skill de Agente Especialista Integrada:** Acompanha a skill executiva `dba-fullstack` orientada a resultados com integração com `/grill-me`.

---

## 📦 Instalação

### Pré-requisitos
- **Go 1.23+** instalado na máquina.
- Git.

### 1. Clonar o Repositório
```bash
git clone https://github.com/alltomatos/sqlmcp.git
cd sqlmcp
```

### 2. Compilar o Binário
```bash
# Via Makefile
make build

# Ou diretamente via Go CLI
go build -o bin/sqlmcp.exe ./cmd/sqlmcp
```

---

## 🛠️ Como Usar

### Modo 1: Configuração em Clientes MCP (Claude Desktop, OpenCode, Cursor)

Adicione o SQLMCP ao arquivo de configuração de servidores MCP do seu cliente:

#### Claude Desktop (`claude_desktop_config.json`) / OpenCode
```json
{
  "mcpServers": {
    "sqlmcp": {
      "command": "D:/dev/sqlmcp/bin/sqlmcp.exe",
      "args": [],
      "env": {
        "SQLMCP_TIMEOUT": "15s",
        "SQLMCP_MAX_ROWS": "500"
      }
    }
  }
}
```

---

### Modo 2: Uso como Biblioteca Go (Vendorização)

Você pode importar e utilizar o motor semântico diretamente no seu backend Go sem overhead de subprocessos:

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/alltomatos/sqlmcp/pkg/dialect"
	"github.com/alltomatos/sqlmcp/pkg/sqlmcp"
)

func main() {
	client := sqlmcp.New()
	defer client.Close()

	// 1. Conecta ao banco de dados
	db, err := sql.Open("pgx", "postgres://user:password@localhost:5432/mydb?sslmode=disable")
	if err != nil {
		panic(err)
	}

	// 2. Registra no pool sob um alias
	_ = client.RegisterConnection("production_pg", dialect.EnginePostgres, db)

	// 3. Executa uma consulta segura com timeout e limites
	ctx := context.Background()
	result, err := client.ExecuteQuery(ctx, "production_pg", "SELECT id, name FROM users", 100, 10*time.Second)
	if err != nil {
		fmt.Printf("Erro: %v\n", err)
		return
	}

	fmt.Printf("Linhas retornadas: %d em %d ms\n", result.RowCount, result.ExecutionMs)
	for _, row := range result.Rows {
		fmt.Printf("Usuário: %v\n", row["name"])
	}
}
```

---

## 🧰 Catálogo de Ferramentas (MCP Tools)

| Ferramenta | Descrição | Parâmetros Principais |
|---|---|---|
| `list_databases` | Lista bancos de dados acessíveis na conexão. | `{ db_alias?: string }` |
| `list_tables` | Lista tabelas e views com contagem aproximada de linhas. | `{ db_alias?: string, schema?: string, include_views?: boolean }` |
| `describe_table` | Detalha colunas, tipos primitivos, nulidade, PKs e comentários. | `{ table_name: string, db_alias?: string, schema?: string }` |
| `get_relations` | Mapeia o grafo de Foreign Keys (relacionamentos) de entrada e saída. | `{ table_name?: string, db_alias?: string, schema?: string }` |
| `generate_mermaid_erd` | Gera diagrama visual Mermaid ERD renderizável no Markdown. | `{ tables?: string[], db_alias?: string, schema?: string }` |
| `execute_query` | Executa consultas `SELECT` com validação de AST, timeout e cap de linhas. | `{ query: string, db_alias?: string, max_rows?: number }` |
| `explain_query` | Traz o plano de execução no dialeto nativo (`EXPLAIN`, `SHOWPLAN`). | `{ query: string, db_alias?: string }` |
| `sample_table` | Traz amostra representativa de dados (ex: 5 linhas). | `{ table_name: string, limit?: number, db_alias?: string }` |
| `stage_to_parquet` | Exporta fatia analítica pesada para arquivo Parquet local. | `{ query: string, stage_name: string, db_alias?: string }` |
| `query_duckdb` | Consulta arquivos Parquet localmente com DuckDB vetorizado. | `{ query: string, stage_name?: string, max_rows?: number }` |
| `connect_database` | Registra dinamicamente uma nova conexão em runtime. | `{ db_alias: string, engine: string, connection_string: string }` |

---

## 🧪 Testes Automatizados

O projeto foi construído sob metodologia rigorosa de TDD e conta com 100% de cobertura nos componentes de core e segurança:

```bash
# Executar todos os testes com saída detalhada
make test
# Ou
go test -v ./...
```

---

## 📚 Documentação Técnica de Arquitetura

- **[ADD.md](docs/architecture/ADD.md):** Architecture Design Document completo com C4 Context & Container.
- **[Sequences](docs/architecture/sequences.md):** Diagramas de sequência dos fluxos de execução segura e aceleração DuckDB.
- **[ADR 0001](docs/adr/0001-stack-go-mcp-vendorizavel.md):** Seleção da Stack Go e empacotamento vendorizável.
- **[ADR 0002](docs/adr/0002-seguranca-read-only-guardrails.md):** Postura de segurança Read-Only estrita.
- **[ADR 0003](docs/adr/0003-aceleracao-analitica-duckdb-parquet.md):** Aceleração analítica com DuckDB e Parquet.
- **[Skill dba-fullstack](.opencode/skills/dba-fullstack/SKILL.md):** Diretrizes para agentes de IA operarem o SQLMCP com foco executivo.

---

## 📄 Licença

Distribuído sob licença MIT. Consulte `LICENSE` para mais informações.
