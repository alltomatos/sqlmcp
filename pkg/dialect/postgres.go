package dialect

import (
	"context"
	"database/sql"
	"fmt"
)

// PostgresDialect implementa o contrato Dialect para o motor PostgreSQL
type PostgresDialect struct{}

// NewPostgres cria uma nova instância de PostgresDialect
func NewPostgres() *PostgresDialect {
	return &PostgresDialect{}
}

func (p *PostgresDialect) Engine() EngineType {
	return EnginePostgres
}

func (p *PostgresDialect) ListDatabases(ctx context.Context, db *sql.DB) ([]string, error) {
	query := `SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar bancos postgres: %w", err)
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

func (p *PostgresDialect) ListTables(ctx context.Context, db *sql.DB, schema string, includeViews bool) ([]TableInfo, error) {
	if schema == "" {
		schema = "public"
	}

	query := `
		SELECT 
			c.relname AS table_name,
			n.nspname AS table_schema,
			CASE c.relkind 
				WHEN 'r' THEN 'BASE TABLE'
				WHEN 'v' THEN 'VIEW'
				WHEN 'm' THEN 'MATERIALIZED VIEW'
				ELSE 'OTHER'
			END AS table_type,
			COALESCE(c.reltuples, 0)::bigint AS estimated_rows
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1
		  AND c.relkind IN ('r', 'v', 'm')
		ORDER BY c.relname;
	`

	rows, err := db.QueryContext(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar tabelas postgres: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var t TableInfo
		if err := rows.Scan(&t.Name, &t.Schema, &t.Type, &t.EstimatedRows); err != nil {
			return nil, err
		}
		if !includeViews && (t.Type == "VIEW" || t.Type == "MATERIALIZED VIEW") {
			continue
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (p *PostgresDialect) DescribeTable(ctx context.Context, db *sql.DB, schema, table string) (*TableDetail, error) {
	if schema == "" {
		schema = "public"
	}

	colQuery := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable = 'YES' AS is_nullable,
			COALESCE(tc.constraint_type = 'PRIMARY KEY', false) AS is_pk,
			c.column_default,
			pg_catalog.col_description(format('%I.%I', c.table_schema, c.table_name)::regclass::oid, c.ordinal_position) AS comment
		FROM information_schema.columns c
		LEFT JOIN information_schema.key_column_usage kcu 
			ON c.table_schema = kcu.table_schema 
			AND c.table_name = kcu.table_name 
			AND c.column_name = kcu.column_name
		LEFT JOIN information_schema.table_constraints tc 
			ON kcu.table_schema = tc.table_schema 
			AND kcu.table_name = tc.table_name 
			AND kcu.constraint_name = tc.constraint_name 
			AND tc.constraint_type = 'PRIMARY KEY'
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position;
	`

	rows, err := db.QueryContext(ctx, colQuery, schema, table)
	if err != nil {
		return nil, fmt.Errorf("falha ao descrever colunas postgres: %w", err)
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

func (p *PostgresDialect) GetRelations(ctx context.Context, db *sql.DB, schema, table string) ([]RelationInfo, error) {
	if schema == "" {
		schema = "public"
	}

	query := `
		SELECT
			tc.constraint_name,
			tc.table_schema AS source_schema,
			tc.table_name AS source_table,
			kcu.column_name AS source_column,
			ccu.table_schema AS target_schema,
			ccu.table_name AS target_table,
			ccu.column_name AS target_column,
			rc.delete_rule AS on_delete_action
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		JOIN information_schema.referential_constraints AS rc
			ON rc.constraint_name = tc.constraint_name
			AND rc.constraint_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND tc.table_schema = $1
		  AND ($2 = '' OR tc.table_name = $2 OR ccu.table_name = $2)
		ORDER BY tc.table_name, kcu.column_name;
	`

	rows, err := db.QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar relacionamentos postgres: %w", err)
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

func (p *PostgresDialect) BuildExplainQuery(query string) string {
	return fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, VERBOSE, FORMAT JSON) %s", query)
}

func (p *PostgresDialect) BuildSampleQuery(schema, table string, limit int) string {
	if limit <= 0 {
		limit = 5
	}
	if schema == "" {
		schema = "public"
	}
	return fmt.Sprintf("SELECT * FROM \"%s\".\"%s\" LIMIT %d;", schema, table, limit)
}
