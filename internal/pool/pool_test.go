package pool_test

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/alltomatos/sqlmcp/internal/pool"
	"github.com/alltomatos/sqlmcp/pkg/dialect"
)

type mockDriver struct{}

func (m *mockDriver) Open(name string) (driver.Conn, error) {
	return &mockConn{}, nil
}

type mockConn struct{}

func (m *mockConn) Prepare(query string) (driver.Stmt, error) { return nil, nil }
func (m *mockConn) Close() error                              { return nil }
func (m *mockConn) Begin() (driver.Tx, error                  ) { return nil, nil }

func init() {
	sql.Register("mock_driver", &mockDriver{})
}

func getMockDB() *sql.DB {
	db, _ := sql.Open("mock_driver", "test")
	return db
}

func TestPool_RegisterAndGetConnection(t *testing.T) {
	p := pool.New()
	defer p.CloseAll()

	mockDB := getMockDB()
	err := p.Register("primary_pg", dialect.EnginePostgres, mockDB)
	if err != nil {
		t.Fatalf("esperava sucesso ao registrar conexão, obteve: %v", err)
	}

	entry, found := p.Get("primary_pg")
	if !found {
		t.Fatalf("esperava encontrar conexão 'primary_pg'")
	}

	if entry.Engine != dialect.EnginePostgres {
		t.Errorf("esperava engine '%s', obteve '%s'", dialect.EnginePostgres, entry.Engine)
	}

	if entry.DB != mockDB {
		t.Errorf("instância do DB retornada é diferente da registrada")
	}
}

func TestPool_RemoveConnection(t *testing.T) {
	p := pool.New()
	defer p.CloseAll()

	mockDB := getMockDB()
	_ = p.Register("test_mysql", dialect.EngineMySQL, mockDB)

	removed := p.Remove("test_mysql")
	if !removed {
		t.Errorf("esperava true ao remover conexão existente")
	}

	_, found := p.Get("test_mysql")
	if found {
		t.Errorf("não deveria encontrar conexão após remoção")
	}
}
