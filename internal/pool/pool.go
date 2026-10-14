package pool

import (
	"database/sql"
	"errors"
	"sync"

	"github.com/alltomatos/sqlmcp/pkg/dialect"
)

// Entry representa uma conexão gerenciada no pool
type Entry struct {
	Alias  string
	Engine dialect.EngineType
	DB     *sql.DB
}

// Pool gerencia instâncias de *sql.DB de forma thread-safe
type Pool struct {
	connections sync.Map // map[string]*Entry
}

// New cria uma nova instância vazia de Pool
func New() *Pool {
	return &Pool{}
}

// Register adiciona ou substitui uma conexão ativa por alias
func (p *Pool) Register(alias string, engine dialect.EngineType, db *sql.DB) error {
	if alias == "" {
		return errors.New("alias de conexão não pode ser vazio")
	}
	if db == nil {
		return errors.New("instância sql.DB não pode ser nula")
	}

	entry := &Entry{
		Alias:  alias,
		Engine: engine,
		DB:     db,
	}

	p.connections.Store(alias, entry)
	return nil
}

// Get recupera uma conexão pelo alias informado
func (p *Pool) Get(alias string) (*Entry, bool) {
	val, ok := p.connections.Load(alias)
	if !ok {
		return nil, false
	}
	return val.(*Entry), true
}

// Remove desconecta e remove uma entrada do pool pelo alias
func (p *Pool) Remove(alias string) bool {
	val, ok := p.connections.LoadAndDelete(alias)
	if !ok {
		return false
	}
	entry := val.(*Entry)
	if entry.DB != nil {
		_ = entry.DB.Close()
	}
	return true
}

// ListAliases retorna a lista de todos os aliases registrados
func (p *Pool) ListAliases() []string {
	var aliases []string
	p.connections.Range(func(key, _ interface{}) bool {
		aliases = append(aliases, key.(string))
		return true
	})
	return aliases
}

// CloseAll encerra todas as conexões do pool
func (p *Pool) CloseAll() {
	p.connections.Range(func(key, value interface{}) bool {
		entry := value.(*Entry)
		if entry.DB != nil {
			_ = entry.DB.Close()
		}
		p.connections.Delete(key)
		return true
	})
}
