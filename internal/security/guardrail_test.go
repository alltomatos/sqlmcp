package security_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alltomatos/sqlmcp/internal/security"
)

func TestGuardrail_ResolveMaxRows(t *testing.T) {
	// Padrão quando 0 ou negativo deve ser 500
	if r := security.ResolveMaxRows(0); r != 500 {
		t.Errorf("esperava 500, obteve %d", r)
	}
	if r := security.ResolveMaxRows(-10); r != 500 {
		t.Errorf("esperava 500, obteve %d", r)
	}

	// Valor intermediário válido
	if r := security.ResolveMaxRows(100); r != 100 {
		t.Errorf("esperava 100, obteve %d", r)
	}

	// Teto máximo deve ser limitado a 2000
	if r := security.ResolveMaxRows(5000); r != 2000 {
		t.Errorf("esperava teto de 2000, obteve %d", r)
	}
}

func TestGuardrail_ResolveTimeout(t *testing.T) {
	// Timeout padrão de 15 segundos
	if d := security.ResolveTimeout(0); d != 15*time.Second {
		t.Errorf("esperava 15s, obteve %v", d)
	}

	// Timeout informado respeitado se menor que 60s
	if d := security.ResolveTimeout(5 * time.Second); d != 5*time.Second {
		t.Errorf("esperava 5s, obteve %v", d)
	}

	// Teto máximo de timeout 60s
	if d := security.ResolveTimeout(120 * time.Second); d != 60*time.Second {
		t.Errorf("esperava teto de 60s, obteve %v", d)
	}
}

func TestGuardrail_TruncateCell(t *testing.T) {
	shortStr := "texto normal de cliente"
	if res, truncated := security.TruncateValue(shortStr); truncated || res != shortStr {
		t.Errorf("não deveria truncar texto curto")
	}

	// Texto com mais de 4096 bytes (4KB)
	longStr := strings.Repeat("A", 5000)
	res, truncated := security.TruncateValue(longStr)
	if !truncated {
		t.Errorf("esperava que texto de 5000 chars fosse truncado")
	}
	if !strings.HasSuffix(res.(string), "[truncated...]") {
		t.Errorf("esperava sufixo [truncated...]")
	}
}
