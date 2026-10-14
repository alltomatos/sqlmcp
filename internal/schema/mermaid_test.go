package schema_test

import (
	"strings"
	"testing"

	"github.com/alltomatos/sqlmcp/internal/schema"
	"github.com/alltomatos/sqlmcp/pkg/dialect"
)

func TestMermaid_GenerateERD(t *testing.T) {
	tables := []dialect.TableDetail{
		{
			TableInfo: dialect.TableInfo{Schema: "public", Name: "users"},
			Columns: []dialect.ColumnInfo{
				{Name: "id", DataType: "uuid", IsPrimaryKey: true},
				{Name: "name", DataType: "text"},
			},
		},
		{
			TableInfo: dialect.TableInfo{Schema: "public", Name: "orders"},
			Columns: []dialect.ColumnInfo{
				{Name: "id", DataType: "uuid", IsPrimaryKey: true},
				{Name: "user_id", DataType: "uuid"},
				{Name: "amount", DataType: "numeric"},
			},
		},
	}

	relations := []dialect.RelationInfo{
		{
			ConstraintName: "fk_orders_users",
			SourceSchema:   "public",
			SourceTable:    "orders",
			SourceColumn:   "user_id",
			TargetSchema:   "public",
			TargetTable:    "users",
			TargetColumn:   "id",
		},
	}

	erd := schema.GenerateMermaidERD(tables, relations)

	if !strings.HasPrefix(erd, "erDiagram") {
		t.Errorf("esperava que o diagrama começasse com 'erDiagram', obteve: %s", erd)
	}

	if !strings.Contains(erd, "users ||--o{ orders : \"fk_orders_users\"") {
		t.Errorf("esperava relacionamento users -> orders, obteve:\n%s", erd)
	}

	if !strings.Contains(erd, "uuid id PK") {
		t.Errorf("esperava anotação de chave primária 'uuid id PK', obteve:\n%s", erd)
	}
}
