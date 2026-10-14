package schema

import (
	"fmt"
	"strings"

	"github.com/alltomatos/sqlmcp/pkg/dialect"
)

// GenerateMermaidERD gera uma representação visual Mermaid ERD a partir de detalhes de tabelas e relacionamentos
func GenerateMermaidERD(tables []dialect.TableDetail, relations []dialect.RelationInfo) string {
	var sb strings.Builder
	sb.WriteString("erDiagram\n")

	// Renderiza os relacionamentos
	for _, rel := range relations {
		sb.WriteString(fmt.Sprintf("    %s ||--o{ %s : \"%s\"\n", rel.TargetTable, rel.SourceTable, rel.ConstraintName))
	}

	// Renderiza as tabelas e suas colunas
	for _, t := range tables {
		sb.WriteString(fmt.Sprintf("    %s {\n", t.Name))
		for _, col := range t.Columns {
			dataType := col.DataType
			if dataType == "" {
				dataType = "string"
			}
			// Normaliza tipos compostos para formato amigável ao Mermaid
			dataType = strings.ReplaceAll(dataType, " ", "_")

			pkAnnotation := ""
			if col.IsPrimaryKey {
				pkAnnotation = " PK"
			}

			sb.WriteString(fmt.Sprintf("        %s %s%s\n", dataType, col.Name, pkAnnotation))
		}
		sb.WriteString("    }\n")
	}

	return sb.String()
}
