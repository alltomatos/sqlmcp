package sqlmcp_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alltomatos/sqlmcp/pkg/dialect"
	"github.com/alltomatos/sqlmcp/pkg/sqlmcp"
)

func TestClient_Lifecycle(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro ao criar sqlmock: %v", err)
	}
	defer db.Close()

	client := sqlmcp.New()
	defer client.Close()

	err = client.RegisterConnection("main_pg", dialect.EnginePostgres, db)
	if err != nil {
		t.Fatalf("erro ao registrar conexao: %v", err)
	}

	rows := sqlmock.NewRows([]string{"id", "title"}).
		AddRow(1, "Post 1").
		AddRow(2, "Post 2")

	mock.ExpectQuery(`SELECT id, title FROM posts`).WillReturnRows(rows)

	res, err := client.ExecuteQuery(context.Background(), "main_pg", "SELECT id, title FROM posts", 10, 5*time.Second)
	if err != nil {
		t.Fatalf("falha ao rodar query via client: %v", err)
	}

	if res.RowCount != 2 {
		t.Errorf("esperava 2 linhas, obteve %d", res.RowCount)
	}
}
