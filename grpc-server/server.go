package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"grpc-rabbitmq-fileshare/common"
	"grpc-rabbitmq-fileshare/grpc-server/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fileServiceServer implementa o servidor gRPC para FileService
type fileServiceServer struct {
	proto.UnimplementedFileServiceServer
	storage common.FileService
}

// NewFileServiceServer cria uma nova instância do servidor gRPC
func NewFileServiceServer(storage common.FileService) *fileServiceServer {
	return &fileServiceServer{
		storage: storage,
	}
}

// ListFiles lista todos os arquivos disponíveis
func (s *fileServiceServer) ListFiles(ctx context.Context, req *proto.Empty) (*proto.FileListResponse, error) {
	log.Printf("[ListFiles] Requisição recebida")

	files, err := s.storage.ListFiles()
	if err != nil {
		log.Printf("[ListFiles] Erro ao listar arquivos: %v", err)
		return nil, status.Errorf(codes.Internal, "erro ao listar arquivos: %v", err)
	}

	log.Printf("[ListFiles] %d arquivo(s) encontrado(s)", len(files))
	return &proto.FileListResponse{
		Files: files,
	}, nil
}

// UploadFile faz upload de um arquivo
func (s *fileServiceServer) UploadFile(ctx context.Context, req *proto.UploadRequest) (*proto.OperationResult, error) {
	log.Printf("[UploadFile] Requisição recebida para arquivo: %s (tamanho: %d bytes)", req.Name, len(req.Data))

	if req.Name == "" {
		log.Printf("[UploadFile] Erro: nome do arquivo vazio")
		return &proto.OperationResult{
			Success: false,
			Message: "nome do arquivo não pode ser vazio",
		}, nil
	}

	if len(req.Data) == 0 {
		log.Printf("[UploadFile] Erro: dados do arquivo vazios")
		return &proto.OperationResult{
			Success: false,
			Message: "dados do arquivo não podem ser vazios",
		}, nil
	}

	err := s.storage.UploadFile(req.Name, req.Data)
	if err != nil {
		log.Printf("[UploadFile] Erro ao fazer upload do arquivo %s: %v", req.Name, err)
		return &proto.OperationResult{
			Success: false,
			Message: fmt.Sprintf("erro ao fazer upload: %v", err),
		}, nil
	}

	log.Printf("[UploadFile] Arquivo %s enviado com sucesso (%d bytes)", req.Name, len(req.Data))
	return &proto.OperationResult{
		Success: true,
		Message: fmt.Sprintf("arquivo %s enviado com sucesso", req.Name),
	}, nil
}

// DownloadFile faz download de um arquivo
func (s *fileServiceServer) DownloadFile(ctx context.Context, req *proto.DownloadRequest) (*proto.DownloadResponse, error) {
	log.Printf("[DownloadFile] Requisição recebida para arquivo: %s", req.Name)

	if req.Name == "" {
		log.Printf("[DownloadFile] Erro: nome do arquivo vazio")
		return nil, status.Errorf(codes.InvalidArgument, "nome do arquivo não pode ser vazio")
	}

	data, err := s.storage.DownloadFile(req.Name)
	if err != nil {
		log.Printf("[DownloadFile] Erro ao fazer download do arquivo %s: %v", req.Name, err)
		return nil, status.Errorf(codes.NotFound, "erro ao fazer download: %v", err)
	}

	log.Printf("[DownloadFile] Arquivo %s baixado com sucesso (%d bytes)", req.Name, len(data))
	return &proto.DownloadResponse{
		Data: data,
	}, nil
}

// StartServer inicia o servidor gRPC na porta especificada
func StartServer(port string, storage common.FileService) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("falha ao escutar na porta %s: %w", port, err)
	}

	// Cria o servidor gRPC
	grpcServer := grpc.NewServer()

	// Registra o serviço
	fileServiceServer := NewFileServiceServer(storage)
	proto.RegisterFileServiceServer(grpcServer, fileServiceServer)

	log.Printf("Servidor gRPC iniciado e escutando na porta %s", port)
	log.Printf("Diretório de armazenamento: %s", getStorageDir(storage))

	// Inicia o servidor
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("falha ao iniciar servidor: %w", err)
	}

	return nil
}

// getStorageDir tenta obter o diretório de armazenamento para logs
func getStorageDir(storage common.FileService) string {
	// Tenta fazer type assertion para LocalStorage
	if _, ok := storage.(*common.LocalStorage); ok {
		// Por enquanto, retornamos uma string genérica
		// Poderia adicionar um método GetBaseDir() na interface se necessário
		return "configurado"
	}
	return "configurado"
}

