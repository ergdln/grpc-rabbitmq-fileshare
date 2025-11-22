package common

// RequestMessage representa uma mensagem de requisição do cliente
type RequestMessage struct {
	Operation string `json:"operation"` // "list", "upload", "download"
	FileName  string `json:"file_name,omitempty"`
	FileData  []byte `json:"file_data,omitempty"` // Base64 encoded para JSON
}

// ResponseMessage representa uma mensagem de resposta do servidor
type ResponseMessage struct {
	Success   bool     `json:"success"`
	Message   string   `json:"message,omitempty"`
	Files     []string `json:"files,omitempty"`     // Para operação "list"
	FileData  []byte   `json:"file_data,omitempty"` // Base64 encoded para JSON
	FileName  string   `json:"file_name,omitempty"` // Para operação "download"
}

