package executor_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alltomatos/sqlmcp/internal/executor"
	"github.com/alltomatos/sqlmcp/internal/pool"
	"github.com/alltomatos/sqlmcp/pkg/dialect"
)

func TestExecutor_ExecuteSafeQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro ao criar sqlmock: %v", err)
	}
	defer db.Close()

	p := pool.New()
	_ = p.Register("test_db", dialect.EnginePostgres, db)

	exec := executor.New(p)

	rows := sqlmock.NewRows([]string{"id", "username"}).
		AddRow(1, "alice").
		AddRow(2, "bob")

	mock.ExpectQuery(`SELECT id, username FROM users`).WillReturnRows(rows)

	res, err := exec.ExecuteQuery(context.Background(), "test_db", "SELECT id, username FROM users", 10, 5*time.Second)
	if err != nil {
		t.Fatalf("erro ao executar query segura: %v", err)
	}

	if res.RowCount != 2 {
		t.Errorf("esperava 2 linhas, obteve %d", res.RowCount)
	}
	if len(res.Columns) != 2 || res.Columns[0] != "id" {
		t.Errorf("colunas incorretas: %v", res.Columns)
	}
}

func TestExecutor_BlockMutations(t *testing.T) {
	p := pool.New()
	exec := executor.New(p)

	_, err := exec.ExecuteQuery(context.Background(), "test_db", "DELETE FROM users", 10, 5*time.Second)
	if err == nil {
		t.Errorf("esperava que mutação fosse bloqueada pelo executor")
	}
}
