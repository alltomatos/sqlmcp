package security

import (
	"fmt"
	"time"
)

const (
	DefaultMaxRows = 500
	CeilingMaxRows = 2000

	DefaultTimeout = 15 * time.Second
	CeilingTimeout = 60 * time.Second

	MaxCellBytes = 4096 // 4 KB
)

// ResolveMaxRows aplica o limite padrão e teto máximo de linhas retornadas
func ResolveMaxRows(requested int) int {
	if requested <= 0 {
		return DefaultMaxRows
	}
	if requested > CeilingMaxRows {
		return CeilingMaxRows
	}
	return requested
}

// ResolveTimeout define a duração máxima da query respeitando o teto
func ResolveTimeout(requested time.Duration) time.Duration {
	if requested <= 0 {
		return DefaultTimeout
	}
	if requested > CeilingTimeout {
		return CeilingTimeout
	}
	return requested
}

// TruncateValue trunca strings ou arrays de bytes que ultrapassem 4KB para evitar flood de tokens
func TruncateValue(val interface{}) (interface{}, bool) {
	if val == nil {
		return nil, false
	}

	switch v := val.(type) {
	case string:
		if len(v) > MaxCellBytes {
			return fmt.Sprintf("%s... [truncated...]", v[:MaxCellBytes]), true
		}
		return v, false
	case []byte:
		if len(v) > MaxCellBytes {
			return fmt.Sprintf("%s... [truncated...]", string(v[:MaxCellBytes])), true
		}
		return v, false
	default:
		return v, false
	}
}
