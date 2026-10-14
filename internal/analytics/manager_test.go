package analytics_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alltomatos/sqlmcp/internal/analytics"
)

func TestAnalytics_StageLifecycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sqlmcp_analytics_test_*")
	if err != nil {
		t.Fatalf("erro ao criar dir temporario: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mgr := analytics.New(tempDir)

	// Registra um stage dummy
	stageFile := filepath.Join(tempDir, "vendas_2025.parquet")
	err = os.WriteFile(stageFile, []byte("PAR1_mock_content"), 0644)
	if err != nil {
		t.Fatalf("erro ao criar arquivo dummy: %v", err)
	}

	info, err := mgr.RegisterStage("vendas_2025", stageFile, 1500)
	if err != nil {
		t.Fatalf("falha ao registrar stage: %v", err)
	}

	if info.Name != "vendas_2025" || info.RowCount != 1500 {
		t.Errorf("dados incorretos do stage: %+v", info)
	}

	stages := mgr.ListStages()
	if len(stages) != 1 {
		t.Errorf("esperava 1 stage registrado, obteve %d", len(stages))
	}

	// Remove stage
	removed := mgr.RemoveStage("vendas_2025")
	if !removed {
		t.Errorf("esperava sucesso ao remover stage")
	}

	if len(mgr.ListStages()) != 0 {
		t.Errorf("lista de stages deveria estar vazia após remoção")
	}
}
