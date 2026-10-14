package dialect_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alltomatos/sqlmcp/pkg/dialect"
)

func TestSQLServerDialect_ListDatabases(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro ao criar sqlmock: %v", err)
	}
	defer db.Close()

	mssql := dialect.NewSQLServer()

	rows := sqlmock.NewRows([]string{"name"}).
		AddRow("master").
		AddRow("CorporateDW").
		AddRow("Billing")

	mock.ExpectQuery(`SELECT name FROM sys.databases WHERE state_desc = 'ONLINE' AND name NOT IN \('master', 'tempdb', 'model', 'msdb'\) ORDER BY name`).
		WillReturnRows(rows)

	dbs, err := mssql.ListDatabases(context.Background(), db)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if len(dbs) != 3 {
		t.Errorf("esperava 3 bancos, obteve %d", len(dbs))
	}
	if dbs[1] != "CorporateDW" {
		t.Errorf("esperava 'CorporateDW', obteve '%s'", dbs[1])
	}
}

func TestSQLServerDialect_BuildExplainQuery(t *testing.T) {
	mssql := dialect.NewSQLServer()
	query := "SELECT * FROM Employees WHERE DepartmentID = 10"
	explain := mssql.BuildExplainQuery(query)

	expected := "SET SHOWPLAN_ALL ON;\nGO\nSELECT * FROM Employees WHERE DepartmentID = 10;\nGO\nSET SHOWPLAN_ALL OFF;"
	if explain != expected {
		t.Errorf("esperava '%s', obteve '%s'", expected, explain)
	}
}
