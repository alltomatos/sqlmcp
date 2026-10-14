package security_test

import (
	"testing"

	"github.com/alltomatos/sqlmcp/internal/security"
)

func TestValidator_ValidReadQueries(t *testing.T) {
	validQueries := []string{
		"SELECT * FROM users",
		"SELECT id, name FROM clients WHERE active = true",
		"WITH sales AS (SELECT * FROM orders) SELECT * FROM sales",
		"EXPLAIN ANALYZE SELECT 1",
		"  select 1+1 ; ",
	}

	for _, q := range validQueries {
		clean, err := security.ValidateReadOnlyQuery(q)
		if err != nil {
			t.Errorf("esperava query válida para '%s', mas obteve erro: %v", q, err)
		}
		if clean == "" {
			t.Errorf("query limpa não deveria ser vazia para '%s'", q)
		}
	}
}

func TestValidator_BlockedMutations(t *testing.T) {
	blockedQueries := []string{
		"INSERT INTO users (name) VALUES ('hacker')",
		"UPDATE clients SET active = false",
		"DELETE FROM orders WHERE id = 1",
		"DROP TABLE products",
		"ALTER TABLE accounts ADD COLUMN balance int",
		"TRUNCATE TABLE logs",
		"EXEC sp_executesql N'SELECT 1'",
		"GRANT ALL PRIVILEGES ON *.* TO 'user'@'%'",
	}

	for _, q := range blockedQueries {
		_, err := security.ValidateReadOnlyQuery(q)
		if err == nil {
			t.Errorf("esperava erro de segurança para query mutadora '%s', mas passou", q)
		}
	}
}

func TestValidator_BlockMultipleStatements(t *testing.T) {
	multiStatements := []string{
		"SELECT 1; DROP TABLE users;",
		"SELECT * FROM a; SELECT * FROM b",
	}

	for _, q := range multiStatements {
		_, err := security.ValidateReadOnlyQuery(q)
		if err == nil {
			t.Errorf("esperava erro ao tentar múltiplos statements: '%s'", q)
		}
	}
}
