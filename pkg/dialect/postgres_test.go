package dialect_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alltomatos/sqlmcp/pkg/dialect"
)

func TestPostgresDialect_ListDatabases(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro ao criar sqlmock: %v", err)
	}
	defer db.Close()

	pg := dialect.NewPostgres()

	rows := sqlmock.NewRows([]string{"datname"}).
		AddRow("postgres").
		AddRow("erp_db").
		AddRow("analytics")

	mock.ExpectQuery(`SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname`).
		WillReturnRows(rows)

	dbs, err := pg.ListDatabases(context.Background(), db)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if len(dbs) != 3 {
		t.Errorf("esperava 3 bancos, obteve %d", len(dbs))
	}
	if dbs[1] != "erp_db" {
		t.Errorf("esperava 'erp_db', obteve '%s'", dbs[1])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectativas do mock não foram cumpridas: %v", err)
	}
}

func TestPostgresDialect_BuildExplainQuery(t *testing.T) {
	pg := dialect.NewPostgres()
	query := "SELECT * FROM users WHERE active = true"
	explain := pg.BuildExplainQuery(query)

	expected := "EXPLAIN (ANALYZE, BUFFERS, VERBOSE, FORMAT JSON) SELECT * FROM users WHERE active = true"
	if explain != expected {
		t.Errorf("esperava '%s', obteve '%s'", expected, explain)
	}
}
