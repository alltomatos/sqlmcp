package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/alltomatos/sqlmcp/internal/pool"
	"github.com/alltomatos/sqlmcp/internal/security"
	"github.com/alltomatos/sqlmcp/pkg/dialect"
)

// Executor orquestra validação de segurança, limites e execução com drivers
type Executor struct {
	pool *pool.Pool
}

// New cria uma nova instância de Executor
func New(p *pool.Pool) *Executor {
	return &Executor{pool: p}
}

// ExecuteQuery valida, limita e executa uma consulta SELECT no banco de dados referenciado
func (e *Executor) ExecuteQuery(ctx context.Context, dbAlias, rawQuery string, requestedMaxRows int, requestedTimeout time.Duration) (*dialect.QueryResult, error) {
	// 1. Validação léxica Read-Only
	cleanQuery, err := security.ValidateReadOnlyQuery(rawQuery)
	if err != nil {
		return nil, err
	}

	// 2. Resolução de limites e guardrails
	maxRows := security.ResolveMaxRows(requestedMaxRows)
	timeout := security.ResolveTimeout(requestedTimeout)

	// 3. Obtenção da conexão
	entry, found := e.pool.Get(dbAlias)
	if !found {
		return nil, fmt.Errorf("banco de dados com alias '%s' não encontrado no pool", dbAlias)
	}

	// 4. Execução com timeout estrito
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	rows, err := entry.DB.QueryContext(execCtx, cleanQuery)
	if err != nil {
		return nil, fmt.Errorf("falha na execução da query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("falha ao ler colunas: %w", err)
	}

	result := &dialect.QueryResult{
		Columns: cols,
		Rows:    []map[string]interface{}{},
	}

	hasTruncatedCell := false

	// Lê até maxRows (+1 tentativa para sinalizar has_more)
	for rows.Next() {
		if len(result.Rows) >= maxRows {
			result.HasMore = true
			break
		}

		scanArgs := make([]interface{}, len(cols))
		values := make([]interface{}, len(cols))
		for i := range scanArgs {
			scanArgs[i] = &values[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("falha ao escanear linha: %w", err)
		}

		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			val := values[i]
			// Converte []byte para string legível
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			truncatedVal, wasTruncated := security.TruncateValue(val)
			if wasTruncated {
				hasTruncatedCell = true
			}
			rowMap[colName] = truncatedVal
		}

		result.Rows = append(result.Rows, rowMap)
	}

	result.RowCount = len(result.Rows)
	result.ExecutionMs = time.Since(start).Milliseconds()
	result.IsTruncated = hasTruncatedCell

	return result, rows.Err()
}
