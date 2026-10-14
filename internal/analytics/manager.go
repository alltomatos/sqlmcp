package analytics

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

// StageInfo armazena os metadados de uma fatia de dados em cache Parquet
type StageInfo struct {
	Name      string    `json:"name"`
	FilePath  string    `json:"file_path"`
	RowCount  int64     `json:"row_count"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
}

// Manager gerencia o ciclo de vida dos arquivos Parquet e instâncias analíticas
type Manager struct {
	baseDir string
	stages  sync.Map // map[string]*StageInfo
}

// New cria um novo gerenciador de analytics no diretório base fornecido
func New(baseDir string) *Manager {
	_ = os.MkdirAll(baseDir, 0755)
	return &Manager{
		baseDir: baseDir,
	}
}

// RegisterStage cataloga um arquivo Parquet já gerado no diretório
func (m *Manager) RegisterStage(name, filePath string, rowCount int64) (*StageInfo, error) {
	if name == "" {
		return nil, errors.New("o nome do stage não pode ser vazio")
	}

	fi, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("arquivo do stage não encontrado: %w", err)
	}

	info := &StageInfo{
		Name:      name,
		FilePath:  filePath,
		RowCount:  rowCount,
		SizeBytes: fi.Size(),
		CreatedAt: time.Now(),
	}

	m.stages.Store(name, info)
	return info, nil
}

// GetStage busca metadados de um stage ativo
func (m *Manager) GetStage(name string) (*StageInfo, bool) {
	val, ok := m.stages.Load(name)
	if !ok {
		return nil, false
	}
	return val.(*StageInfo), true
}

// ListStages retorna todos os stages cadastrados
func (m *Manager) ListStages() []StageInfo {
	var list []StageInfo
	m.stages.Range(func(_, val interface{}) bool {
		list = append(list, *val.(*StageInfo))
		return true
	})
	return list
}

// RemoveStage apaga o arquivo do disco e descadastra o stage
func (m *Manager) RemoveStage(name string) bool {
	val, ok := m.stages.LoadAndDelete(name)
	if !ok {
		return false
	}
	info := val.(*StageInfo)
	_ = os.Remove(info.FilePath)
	return true
}

// CleanAll remove todos os arquivos e zera os stages
func (m *Manager) CleanAll() {
	m.stages.Range(func(key, val interface{}) bool {
		info := val.(*StageInfo)
		_ = os.Remove(info.FilePath)
		m.stages.Delete(key)
		return true
	})
}
