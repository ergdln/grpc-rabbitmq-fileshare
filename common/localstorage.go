package common

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// LocalStorage implementa FileService usando armazenamento local em disco
type LocalStorage struct {
	baseDir string
	mu      sync.RWMutex
}

// NewLocalStorage cria uma nova instância de LocalStorage
// Cria o diretório base se ele não existir
func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	ls := &LocalStorage{
		baseDir: baseDir,
	}

	// Garante que o diretório existe
	if err := ls.ensureDir(); err != nil {
		return nil, fmt.Errorf("falha ao criar diretório base: %w", err)
	}

	return ls, nil
}

// ensureDir cria o diretório base se ele não existir
func (ls *LocalStorage) ensureDir() error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if err := os.MkdirAll(ls.baseDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório %s: %w", ls.baseDir, err)
	}

	return nil
}

// ListFiles retorna uma lista de nomes de arquivos disponíveis
func (ls *LocalStorage) ListFiles() ([]string, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	entries, err := os.ReadDir(ls.baseDir)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler diretório %s: %w", ls.baseDir, err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// UploadFile faz upload de um arquivo com o nome e dados especificados
func (ls *LocalStorage) UploadFile(name string, data []byte) error {
	// Validação básica do nome do arquivo
	if name == "" {
		return fmt.Errorf("nome do arquivo não pode ser vazio")
	}

	// Previne path traversal attacks
	if filepath.Base(name) != name {
		return fmt.Errorf("nome do arquivo inválido: não pode conter caminhos relativos")
	}

	ls.mu.Lock()
	defer ls.mu.Unlock()

	// Garante que o diretório existe antes de escrever
	if err := ls.ensureDir(); err != nil {
		return err
	}

	filePath := filepath.Join(ls.baseDir, name)

	// Escreve o arquivo
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("erro ao escrever arquivo %s: %w", filePath, err)
	}

	return nil
}

// DownloadFile faz download de um arquivo pelo nome e retorna seus dados
func (ls *LocalStorage) DownloadFile(name string) ([]byte, error) {
	// Validação básica do nome do arquivo
	if name == "" {
		return nil, fmt.Errorf("nome do arquivo não pode ser vazio")
	}

	// Previne path traversal attacks
	if filepath.Base(name) != name {
		return nil, fmt.Errorf("nome do arquivo inválido: não pode conter caminhos relativos")
	}

	ls.mu.RLock()
	defer ls.mu.RUnlock()

	filePath := filepath.Join(ls.baseDir, name)

	// Verifica se o arquivo existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("arquivo não encontrado: %s", name)
	}

	// Lê o arquivo
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo %s: %w", filePath, err)
	}

	return data, nil
}

