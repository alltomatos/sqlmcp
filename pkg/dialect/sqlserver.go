package dialect

import (
	"context"
	"database/sql"
	"fmt"
)

// SQLServerDialect implementa o contrato Dialect para Microsoft SQL Server
type SQLServerDialect struct{}

// NewSQLServer cria uma nova instância de SQLServerDialect
func NewSQLServer() *SQLServerDialect {
	return &SQLServerDialect{}
}

func (s *SQLServerDialect) Engine() EngineType {
	return EngineSQLServer
}

func (s *SQLServerDialect) ListDatabases(ctx context.Context, db *sql.DB) ([]string, error) {
	query := `SELECT name FROM sys.databases WHERE state_desc = 'ONLINE' AND name NOT IN ('master', 'tempdb', 'model', 'msdb') ORDER BY name`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar bancos sqlserver: %w", err)
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

func (s *SQLServerDialect) ListTables(ctx context.Context, db *sql.DB, schema string, includeViews bool) ([]TableInfo, error) {
	query := `
		SELECT 
			t.name AS table_name,
			s.name AS schema_name,
			CASE t.type 
				WHEN 'U' THEN 'BASE TABLE'
				WHEN 'V' THEN 'VIEW'
				ELSE 'OTHER'
			END AS table_type,
			ISNULL(p.rows, 0) AS estimated_rows
		FROM sys.tables t
		INNER JOIN sys.schemas s ON t.schema_id = s.schema_id
		LEFT JOIN sys.partitions p ON t.object_id = p.object_id AND p.index_id IN (0, 1)
		WHERE (@p1 = '' OR s.name = @p1)
		ORDER BY t.name;
	`

	rows, err := db.QueryContext(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar tabelas sqlserver: %w", err)
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

func (s *SQLServerDialect) DescribeTable(ctx context.Context, db *sql.DB, schema, table string) (*TableDetail, error) {
	if schema == "" {
		schema = "dbo"
	}

	query := `
		SELECT 
			c.name AS column_name,
			tp.name AS data_type,
			c.is_nullable,
			ISNULL(ic.is_primary_key, 0) AS is_pk,
			dc.definition AS default_value,
			CAST(ep.value AS NVARCHAR(MAX)) AS comment
		FROM sys.columns c
		INNER JOIN sys.tables t ON c.object_id = t.object_id
		INNER JOIN sys.schemas s ON t.schema_id = s.schema_id
		INNER JOIN sys.types tp ON c.user_type_id = tp.user_type_id
		LEFT JOIN (
			SELECT i.object_id, ic.column_id, 1 AS is_primary_key
			FROM sys.indexes i
			INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
			WHERE i.is_primary_key = 1
		) ic ON c.object_id = ic.object_id AND c.column_id = ic.column_id
		LEFT JOIN sys.default_constraints dc ON c.default_object_id = dc.object_id
		LEFT JOIN sys.extended_properties ep ON ep.major_id = c.object_id AND ep.minor_id = c.column_id AND ep.name = 'MS_Description'
		WHERE s.name = @p1 AND t.name = @p2
		ORDER BY c.column_id;
	`

	rows, err := db.QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("falha ao descrever colunas sqlserver: %w", err)
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

func (s *SQLServerDialect) GetRelations(ctx context.Context, db *sql.DB, schema, table string) ([]RelationInfo, error) {
	if schema == "" {
		schema = "dbo"
	}

	query := `
		SELECT 
			fk.name AS constraint_name,
			s1.name AS source_schema,
			t1.name AS source_table,
			c1.name AS source_column,
			s2.name AS target_schema,
			t2.name AS target_table,
			c2.name AS target_column,
			fk.delete_referential_action_desc AS on_delete_action
		FROM sys.foreign_keys fk
		INNER JOIN sys.foreign_key_columns fkc ON fk.object_id = fkc.constraint_object_id
		INNER JOIN sys.tables t1 ON fkc.parent_object_id = t1.object_id
		INNER JOIN sys.schemas s1 ON t1.schema_id = s1.schema_id
		INNER JOIN sys.columns c1 ON fkc.parent_object_id = c1.object_id AND fkc.parent_column_id = c1.column_id
		INNER JOIN sys.tables t2 ON fkc.referenced_object_id = t2.object_id
		INNER JOIN sys.schemas s2 ON t2.schema_id = s2.schema_id
		INNER JOIN sys.columns c2 ON fkc.referenced_object_id = c2.object_id AND fkc.referenced_column_id = c2.column_id
		WHERE (@p1 = '' OR s1.name = @p1)
		  AND (@p2 = '' OR t1.name = @p2 OR t2.name = @p2)
		ORDER BY t1.name, c1.name;
	`

	rows, err := db.QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar relacionamentos sqlserver: %w", err)
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

func (s *SQLServerDialect) BuildExplainQuery(query string) string {
	return fmt.Sprintf("SET SHOWPLAN_ALL ON;\nGO\n%s;\nGO\nSET SHOWPLAN_ALL OFF;", query)
}

func (s *SQLServerDialect) BuildSampleQuery(schema, table string, limit int) string {
	if limit <= 0 {
		limit = 5
	}
	if schema == "" {
		schema = "dbo"
	}
	return fmt.Sprintf("SELECT TOP (%d) * FROM [%s].[%s];", limit, schema, table)
}
