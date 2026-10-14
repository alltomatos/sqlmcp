package security

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var blockedKeywords = []string{
	"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "TRUNCATE",
	"CREATE", "RENAME", "REPLACE", "EXEC", "EXECUTE", "GRANT",
	"REVOKE", "MERGE", "CALL", "UPSERT", "INTO",
}

// ValidateReadOnlyQuery valida se a query SQL informada é estritamente de leitura.
// Rejeita múltiplos statements e comandos de mutação/DDL.
func ValidateReadOnlyQuery(rawQuery string) (string, error) {
	trimmed := strings.TrimSpace(rawQuery)
	if trimmed == "" {
		return "", errors.New("a query não pode ser vazia")
	}

	// Remove eventual ponto e vírgula no final antes da checagem de múltiplos statements
	cleanEnd := strings.TrimRightFunc(trimmed, func(r rune) bool {
		return r == ';' || unicode.IsSpace(r)
	})

	if strings.Contains(cleanEnd, ";") {
		return "", errors.New("segurança: execução de múltiplos statements separados por ponto e vírgula é estritamente proibida")
	}

	upper := strings.ToUpper(cleanEnd)
	tokens := strings.Fields(upper)
	if len(tokens) == 0 {
		return "", errors.New("a query não possui tokens válidos")
	}

	firstToken := tokens[0]
	// Permite apenas comandos que iniciam com SELECT, WITH ou EXPLAIN
	if firstToken != "SELECT" && firstToken != "WITH" && firstToken != "EXPLAIN" {
		return "", fmt.Errorf("segurança: comando '%s' não permitido. Apenas consultas de leitura (SELECT, WITH, EXPLAIN) são aceitas", firstToken)
	}

	// Varre palavras-chave bloqueadas no corpo da query para evitar injeções ou subqueries destrutivas
	for _, token := range tokens {
		// Remove caracteres de pontuação colados ao token
		cleanedToken := strings.Trim(token, "(),;\"'`")
		for _, blocked := range blockedKeywords {
			if cleanedToken == blocked {
				// Exceção segura: a palavra INTO em expressões que não sejam INSERT INTO (embora SELECT INTO crie tabelas e deva ser bloqueado)
				return "", fmt.Errorf("segurança: palavra-chave mutadora ou perigosa '%s' detectada e bloqueada", blocked)
			}
		}
	}

	return cleanEnd, nil
}
