package sqlmcp

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/alltomatos/sqlmcp/internal/analytics"
	"github.com/alltomatos/sqlmcp/internal/executor"
	"github.com/alltomatos/sqlmcp/internal/pool"
	"github.com/alltomatos/sqlmcp/pkg/dialect"
)

// Client é o ponto de entrada público e unificado do SQLMCP para ser usado
// tanto pelo executável MCP quanto por outras ferramentas Go via importação de pacote.
type Client struct {
	pool      *pool.Pool
	executor  *executor.Executor
	analytics *analytics.Manager
	dialects  map[dialect.EngineType]dialect.Dialect
}

// New cria uma nova instância de Client
func New() *Client {
	p := pool.New()
	return &Client{
		pool:      p,
		executor:  executor.New(p),
		analytics: analytics.New("./tmp/analytics"),
		dialects: map[dialect.EngineType]dialect.Dialect{
			dialect.EnginePostgres:  dialect.NewPostgres(),
			dialect.EngineMySQL:     dialect.NewMySQL(),
			dialect.EngineSQLServer: dialect.NewSQLServer(),
		},
	}
}

// RegisterConnection cadastra uma instância de *sql.DB sob um alias
func (c *Client) RegisterConnection(alias string, engine dialect.EngineType, db *sql.DB) error {
	return c.pool.Register(alias, engine, db)
}

// ExecuteQuery executa com segurança uma consulta de leitura com limites e truncamento
func (c *Client) ExecuteQuery(ctx context.Context, dbAlias, query string, maxRows int, timeout time.Duration) (*dialect.QueryResult, error) {
	return c.executor.ExecuteQuery(ctx, dbAlias, query, maxRows, timeout)
}

// ListTables lista as tabelas do banco associado ao alias
func (c *Client) ListTables(ctx context.Context, dbAlias, schema string, includeViews bool) ([]dialect.TableInfo, error) {
	entry, found := c.pool.Get(dbAlias)
	if !found {
		return nil, fmt.Errorf("alias '%s' não encontrado", dbAlias)
	}

	d, ok := c.dialects[entry.Engine]
	if !ok {
		return nil, fmt.Errorf("dialeto não suportado: %s", entry.Engine)
	}

	return d.ListTables(ctx, entry.DB, schema, includeViews)
}

// DescribeTable descreve colunas, tipos e constraints de uma tabela
func (c *Client) DescribeTable(ctx context.Context, dbAlias, schema, table string) (*dialect.TableDetail, error) {
	entry, found := c.pool.Get(dbAlias)
	if !found {
		return nil, fmt.Errorf("alias '%s' não encontrado", dbAlias)
	}

	d, ok := c.dialects[entry.Engine]
	if !ok {
		return nil, fmt.Errorf("dialeto não suportado: %s", entry.Engine)
	}

	return d.DescribeTable(ctx, entry.DB, schema, table)
}

// Close encerra os pools de conexões e limpa recursos
func (c *Client) Close() {
	c.pool.CloseAll()
	c.analytics.CleanAll()
}
