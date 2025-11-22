package common

// FileService define a interface para operações de sistema de arquivos remoto
type FileService interface {
	// ListFiles retorna uma lista de nomes de arquivos disponíveis
	ListFiles() ([]string, error)

	// UploadFile faz upload de um arquivo com o nome e dados especificados
	UploadFile(name string, data []byte) error

	// DownloadFile faz download de um arquivo pelo nome e retorna seus dados
	DownloadFile(name string) ([]byte, error)
}

