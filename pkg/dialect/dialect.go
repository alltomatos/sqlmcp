package dialect

import (
	"context"
	"database/sql"
)

// EngineType identifica o dialeto suportado
type EngineType string

const (
	EnginePostgres  EngineType = "postgres"
	EngineMySQL     EngineType = "mysql"
	EngineSQLServer EngineType = "sqlserver"
	EngineDuckDB    EngineType = "duckdb"
)

// ColumnInfo mapeia as propriedades de uma coluna no catálogo
type ColumnInfo struct {
	Name         string  `json:"name"`
	DataType     string  `json:"data_type"`
	IsNullable   bool    `json:"is_nullable"`
	IsPrimaryKey bool    `json:"is_primary_key"`
	DefaultValue *string `json:"default_value,omitempty"`
	Comment      *string `json:"comment,omitempty"`
}

// TableInfo mapeia a visão simplificada da tabela no catálogo
type TableInfo struct {
	Schema       string `json:"schema"`
	Name         string `json:"name"`
	Type         string `json:"type"` // "BASE TABLE" ou "VIEW"
	EstimatedRows int64 `json:"estimated_rows"`
}

// TableDetail contém o esquema completo de uma tabela
type TableDetail struct {
	TableInfo
	Columns []ColumnInfo `json:"columns"`
	Indices []IndexInfo  `json:"indices"`
}

// IndexInfo descreve um índice existente na tabela
type IndexInfo struct {
	Name      string   `json:"name"`
	IsUnique  bool     `json:"is_unique"`
	IsPrimary bool     `json:"is_primary"`
	Columns   []string `json:"columns"`
}

// RelationInfo descreve um relacionamento entre duas tabelas (Foreign Key)
type RelationInfo struct {
	ConstraintName    string `json:"constraint_name"`
	SourceSchema      string `json:"source_schema"`
	SourceTable       string `json:"source_table"`
	SourceColumn      string `json:"source_column"`
	TargetSchema      string `json:"target_schema"`
	TargetTable       string `json:"target_table"`
	TargetColumn      string `json:"target_column"`
	OnDeleteAction    string `json:"on_delete_action"`
}

// QueryResult padroniza o retorno seguro para consumo pelo LLM
type QueryResult struct {
	Columns      []string                 `json:"columns"`
	Rows         []map[string]interface{} `json:"rows"`
	RowCount     int                      `json:"row_count"`
	HasMore      bool                     `json:"has_more"`
	ExecutionMs  int64                    `json:"execution_ms"`
	IsTruncated  bool                     `json:"is_truncated"`
}

// Dialect define o contrato que cada banco relacional deve implementar
type Dialect interface {
	Engine() EngineType
	ListDatabases(ctx context.Context, db *sql.DB) ([]string, error)
	ListTables(ctx context.Context, db *sql.DB, schema string, includeViews bool) ([]TableInfo, error)
	DescribeTable(ctx context.Context, db *sql.DB, schema, table string) (*TableDetail, error)
	GetRelations(ctx context.Context, db *sql.DB, schema, table string) ([]RelationInfo, error)
	BuildExplainQuery(query string) string
	BuildSampleQuery(schema, table string, limit int) string
}
