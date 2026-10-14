package dialect_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alltomatos/sqlmcp/pkg/dialect"
)

func TestMySQLDialect_ListDatabases(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro ao criar sqlmock: %v", err)
	}
	defer db.Close()

	my := dialect.NewMySQL()

	rows := sqlmock.NewRows([]string{"schema_name"}).
		AddRow("mysql_app").
		AddRow("finance_db")

	mock.ExpectQuery(`SELECT schema_name FROM information_schema.schemata WHERE schema_name NOT IN \('information_schema', 'mysql', 'performance_schema', 'sys'\) ORDER BY schema_name`).
		WillReturnRows(rows)

	dbs, err := my.ListDatabases(context.Background(), db)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if len(dbs) != 2 {
		t.Errorf("esperava 2 bancos, obteve %d", len(dbs))
	}
	if dbs[0] != "mysql_app" {
		t.Errorf("esperava 'mysql_app', obteve '%s'", dbs[0])
	}
}

func TestMySQLDialect_BuildExplainQuery(t *testing.T) {
	my := dialect.NewMySQL()
	query := "SELECT * FROM orders WHERE status = 'PAID'"
	explain := my.BuildExplainQuery(query)

	expected := "EXPLAIN FORMAT=JSON SELECT * FROM orders WHERE status = 'PAID'"
	if explain != expected {
		t.Errorf("esperava '%s', obteve '%s'", expected, explain)
	}
}
