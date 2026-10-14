package dialect

import (
	"context"
	"database/sql"
	"fmt"
)

// MySQLDialect implementa o contrato Dialect para o motor MySQL / MariaDB
type MySQLDialect struct{}

// NewMySQL cria uma nova instância de MySQLDialect
func NewMySQL() *MySQLDialect {
	return &MySQLDialect{}
}

func (m *MySQLDialect) Engine() EngineType {
	return EngineMySQL
}

func (m *MySQLDialect) ListDatabases(ctx context.Context, db *sql.DB) ([]string, error) {
	query := `SELECT schema_name FROM information_schema.schemata WHERE schema_name NOT IN ('information_schema', 'mysql', 'performance_schema', 'sys') ORDER BY schema_name`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar bancos mysql: %w", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		databases = append(databases, name)
	}
	return databases, rows.Err()
}

func (m *MySQLDialect) ListTables(ctx context.Context, db *sql.DB, schema string, includeViews bool) ([]TableInfo, error) {
	query := `
		SELECT 
			TABLE_NAME, 
			TABLE_SCHEMA, 
			TABLE_TYPE, 
			COALESCE(TABLE_ROWS, 0)
		FROM information_schema.tables 
		WHERE TABLE_SCHEMA = IF(? = '', DATABASE(), ?)
		ORDER BY TABLE_NAME;
	`

	rows, err := db.QueryContext(ctx, query, schema, schema)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar tabelas mysql: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var t TableInfo
		if err := rows.Scan(&t.Name, &t.Schema, &t.Type, &t.EstimatedRows); err != nil {
			return nil, err
		}
		if !includeViews && t.Type == "VIEW" {
			continue
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (m *MySQLDialect) DescribeTable(ctx context.Context, db *sql.DB, schema, table string) (*TableDetail, error) {
	query := `
		SELECT 
			COLUMN_NAME, 
			COLUMN_TYPE, 
			IS_NULLABLE = 'YES', 
			COLUMN_KEY = 'PRI', 
			COLUMN_DEFAULT, 
			COLUMN_COMMENT 
		FROM information_schema.columns 
		WHERE TABLE_SCHEMA = IF(? = '', DATABASE(), ?) AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION;
	`

	rows, err := db.QueryContext(ctx, query, schema, schema, table)
	if err != nil {
		return nil, fmt.Errorf("falha ao descrever colunas mysql: %w", err)
	}
	defer rows.Close()

	detail := &TableDetail{
		TableInfo: TableInfo{
			Schema: schema,
			Name:   table,
			Type:   "BASE TABLE",
		},
		Columns: []ColumnInfo{},
		Indices: []IndexInfo{},
	}

	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.Name, &col.DataType, &col.IsNullable, &col.IsPrimaryKey, &col.DefaultValue, &col.Comment); err != nil {
			return nil, err
		}
		detail.Columns = append(detail.Columns, col)
	}

	return detail, rows.Err()
}

func (m *MySQLDialect) GetRelations(ctx context.Context, db *sql.DB, schema, table string) ([]RelationInfo, error) {
	query := `
		SELECT 
			kcu.CONSTRAINT_NAME, 
			kcu.TABLE_SCHEMA AS source_schema, 
			kcu.TABLE_NAME AS source_table, 
			kcu.COLUMN_NAME AS source_column, 
			kcu.REFERENCED_TABLE_SCHEMA AS target_schema, 
			kcu.REFERENCED_TABLE_NAME AS target_table, 
			kcu.REFERENCED_COLUMN_NAME AS target_column,
			rc.DELETE_RULE AS on_delete_action
		FROM information_schema.KEY_COLUMN_USAGE kcu
		JOIN information_schema.REFERENTIAL_CONSTRAINTS rc
			ON kcu.CONSTRAINT_NAME = rc.CONSTRAINT_NAME
			AND kcu.CONSTRAINT_SCHEMA = rc.CONSTRAINT_SCHEMA
		WHERE kcu.TABLE_SCHEMA = IF(? = '', DATABASE(), ?)
		  AND kcu.REFERENCED_TABLE_NAME IS NOT NULL
		  AND (? = '' OR kcu.TABLE_NAME = ? OR kcu.REFERENCED_TABLE_NAME = ?)
		ORDER BY kcu.TABLE_NAME, kcu.COLUMN_NAME;
	`

	rows, err := db.QueryContext(ctx, query, schema, schema, table, table, table)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar relacionamentos mysql: %w", err)
	}
	defer rows.Close()

	var relations []RelationInfo
	for rows.Next() {
		var r RelationInfo
		if err := rows.Scan(
			&r.ConstraintName,
			&r.SourceSchema,
			&r.SourceTable,
			&r.SourceColumn,
			&r.TargetSchema,
			&r.TargetTable,
			&r.TargetColumn,
			&r.OnDeleteAction,
		); err != nil {
			return nil, err
		}
		relations = append(relations, r)
	}
	return relations, rows.Err()
}

func (m *MySQLDialect) BuildExplainQuery(query string) string {
	return fmt.Sprintf("EXPLAIN FORMAT=JSON %s", query)
}

func (m *MySQLDialect) BuildSampleQuery(schema, table string, limit int) string {
	if limit <= 0 {
		limit = 5
	}
	if schema == "" {
		return fmt.Sprintf("SELECT * FROM `%s` LIMIT %d;", table, limit)
	}
	return fmt.Sprintf("SELECT * FROM `%s`.`%s` LIMIT %d;", schema, table, limit)
}
