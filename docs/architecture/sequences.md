# Diagramas de Sequência e Fluxos Operacionais - SQLMCP

## 1. Fluxo de Execução Segura Read-Only (`execute_query`)

Este fluxo ilustra as barreiras de segurança atômicas pelas quais uma query passa antes de ser enviada ao banco de dados:

```mermaid
sequenceDiagram
    autonumber
    actor Agent as Agente de IA / MCP Client
    participant Server as MCP Transport Handler (mcp-go)
    participant Guard as Security Guard (AST & Limiters)
    participant Pool as Connection Pool Manager
    participant Dialect as Dialect Adapter (PG/MySQL/MSSQL)
    participant DB as Banco de Dados Relacional

    Agent->>Server: CallTool("execute_query", {query, db_alias, max_rows})
    Server->>Guard: ValidateQuery(query, max_rows)
    
    alt Query contém DML/DDL (INSERT/DROP/UPDATE/etc) ou múltiplos statements
        Guard-->>Server: Error: Query bloqueada (Políticas Read-Only violadas)
        Server-->>Agent: Retorno com erro explicativo amigável
    else Query é puramente de leitura (SELECT/WITH)
        Guard->>Guard: Injeta contexto com Timeout (15s) e Max Rows Cap (500)
        Guard->>Pool: GetConnection(db_alias)
        Pool-->>Guard: *sql.DB
        Guard->>Dialect: ExecuteSafe(ctx, db, sanitizedQuery, maxRows)
        Dialect->>DB: QueryContext(ctx, query)
        DB-->>Dialect: Rows stream
        Dialect->>Dialect: Lê até maxRows (+1 para has_more) e trunca células > 4KB
        Dialect-->>Guard: QueryResult{Columns, Rows, RowCount, HasMore, IsTruncated}
        Guard-->>Server: Result payload
        Server-->>Agent: JSON-RPC Response com dados estruturados
    end
```

---

## 2. Fluxo de Aceleração Analítica Local (`stage_to_parquet` + `query_duckdb`)

Este fluxo é acionado quando o agente identifica que a análise sobrecarregaria o banco OLTP ou excederia os limites de contexto:

```mermaid
sequenceDiagram
    autonumber
    actor Agent as Agente de IA
    participant Server as MCP Server
    participant Extractor as Parquet Extractor Engine
    participant DB as Banco de Produção (PG/MySQL/MSSQL)
    participant LocalDisk as Armazenamento Temporário (Cache)
    participant DuckDB as DuckDB In-Process Engine

    Agent->>Server: CallTool("stage_to_parquet", {query, stage_name, db_alias})
    Server->>Extractor: Extrair fatia para Parquet
    Extractor->>DB: Executa query de extração (SELECT com stream)
    DB-->>Extractor: Chunk de dados
    Extractor->>LocalDisk: Grava arquivo stage_name.parquet (comprimido em colunas)
    LocalDisk-->>Extractor: Arquivo pronto (~5 MB vs 100 MB raw)
    Extractor-->>Server: Retorna metadados: {stage_name, row_count, file_size_bytes}
    Server-->>Agent: "Fatia estagiada com sucesso no stage: 'vendas_2025'"

    opt Consulta Analítica sobre a fatia
        Agent->>Server: CallTool("query_duckdb", {query: "SELECT status, SUM(total) FROM vendas_2025 GROUP BY 1"})
        Server->>DuckDB: Executa SQL colunar vetorizado sobre stage_name.parquet
        DuckDB->>LocalDisk: Scan ultra-rápido do arquivo colunar
        DuckDB-->>Server: Resultados agregados em milissegundos
        Server-->>Agent: Retorna agregação pronta sem tocar mais no banco de produção
    end
```
