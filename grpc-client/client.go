package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"grpc-rabbitmq-fileshare/grpc-server/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client representa o cliente gRPC
type Client struct {
	conn   *grpc.ClientConn
	client proto.FileServiceClient
}

// NewClient cria uma nova instância do cliente gRPC
func NewClient(serverAddr string) (*Client, error) {
	// Conecta ao servidor
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao servidor: %w", err)
	}

	client := proto.NewFileServiceClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

// Close fecha a conexão com o servidor
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ListFiles lista todos os arquivos disponíveis no servidor
func (c *Client) ListFiles() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.ListFiles(ctx, &proto.Empty{})
	if err != nil {
		return fmt.Errorf("erro ao listar arquivos: %w", err)
	}

	if len(resp.Files) == 0 {
		fmt.Println("📁 Nenhum arquivo encontrado no servidor")
		return nil
	}

	fmt.Println("📁 Arquivos disponíveis no servidor:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for i, file := range resp.Files {
		fmt.Printf("  %d. %s\n", i+1, file)
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Total: %d arquivo(s)\n", len(resp.Files))

	return nil
}

// UploadFile faz upload de um arquivo para o servidor
func (c *Client) UploadFile(filePath string) error {
	// Abre o arquivo
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo %s: %w", filePath, err)
	}
	defer file.Close()

	// Lê o conteúdo do arquivo
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo %s: %w", filePath, err)
	}

	// Obtém apenas o nome do arquivo (sem o caminho)
	fileName := filepath.Base(filePath)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Faz o upload
	req := &proto.UploadRequest{
		Name: fileName,
		Data: data,
	}

	resp, err := c.client.UploadFile(ctx, req)
	if err != nil {
		return fmt.Errorf("erro ao fazer upload: %w", err)
	}

	if resp.Success {
		fmt.Printf("✅ Upload realizado com sucesso!\n")
		fmt.Printf("   Arquivo: %s\n", fileName)
		fmt.Printf("   Tamanho: %d bytes\n", len(data))
		fmt.Printf("   Mensagem: %s\n", resp.Message)
	} else {
		fmt.Printf("❌ Falha no upload!\n")
		fmt.Printf("   Mensagem: %s\n", resp.Message)
		return fmt.Errorf("upload falhou: %s", resp.Message)
	}

	return nil
}

// DownloadFile faz download de um arquivo do servidor
func (c *Client) DownloadFile(fileName string, outputPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Faz o download
	req := &proto.DownloadRequest{
		Name: fileName,
	}

	resp, err := c.client.DownloadFile(ctx, req)
	if err != nil {
		return fmt.Errorf("erro ao fazer download: %w", err)
	}

	// Se outputPath não foi especificado, usa o nome do arquivo
	if outputPath == "" {
		outputPath = fileName
	}

	// Escreve o arquivo
	err = os.WriteFile(outputPath, resp.Data, 0644)
	if err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %w", err)
	}

	fmt.Printf("✅ Download realizado com sucesso!\n")
	fmt.Printf("   Arquivo: %s\n", fileName)
	fmt.Printf("   Tamanho: %d bytes\n", len(resp.Data))
	fmt.Printf("   Salvo em: %s\n", outputPath)

	return nil
}

