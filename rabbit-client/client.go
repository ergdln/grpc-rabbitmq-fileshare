package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"grpc-rabbitmq-fileshare/common"

	"github.com/streadway/amqp"
)

const (
	requestQueue = "rpc-file-requests"
	timeout      = 30 * time.Second
)

// Client representa o cliente RabbitMQ
type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	replyQueue amqp.Queue
}

// NewClient cria uma nova instância do cliente RabbitMQ
func NewClient(amqpURL string) (*Client, error) {
	// Conecta ao RabbitMQ
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao RabbitMQ: %w", err)
	}

	// Cria um canal
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("falha ao abrir canal: %w", err)
	}

	// Declara a fila de requisições (caso não exista)
	_, err = channel.QueueDeclare(
		requestQueue, // nome
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("falha ao declarar fila de requisições: %w", err)
	}

	// Cria uma fila de resposta exclusiva para este cliente
	replyQueue, err := channel.QueueDeclare(
		"",    // nome (vazio = RabbitMQ gera automaticamente)
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("falha ao criar fila de resposta: %w", err)
	}

	return &Client{
		conn:       conn,
		channel:    channel,
		replyQueue: replyQueue,
	}, nil
}

// Close fecha as conexões
func (c *Client) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// sendRequest envia uma requisição e aguarda a resposta
func (c *Client) sendRequest(req common.RequestMessage) (*common.ResponseMessage, error) {
	// Gera um correlation_id único
	correlationID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Serializa a requisição
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %w", err)
	}

	// Consome mensagens da fila de resposta
	msgs, err := c.channel.Consume(
		c.replyQueue.Name, // queue
		"",                 // consumer
		true,               // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao registrar consumidor: %w", err)
	}

	// Publica a requisição
	err = c.channel.Publish(
		"",           // exchange
		requestQueue, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       c.replyQueue.Name,
			Body:          body,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao publicar mensagem: %w", err)
	}

	// Aguarda a resposta com timeout
	timeoutChan := time.After(timeout)
	for {
		select {
		case msg := <-msgs:
			// Verifica se é a resposta correta
			if msg.CorrelationId == correlationID {
				var resp common.ResponseMessage
				if err := json.Unmarshal(msg.Body, &resp); err != nil {
					return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
				}
				return &resp, nil
			}
		case <-timeoutChan:
			return nil, fmt.Errorf("timeout aguardando resposta")
		}
	}
}

// ListFiles lista todos os arquivos disponíveis no servidor
func (c *Client) ListFiles() error {
	req := common.RequestMessage{
		Operation: "list",
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("erro: %s", resp.Message)
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

	// Obtém apenas o nome do arquivo
	fileName := filepath.Base(filePath)

	// Codifica os dados em base64 para JSON
	encodedData := base64.StdEncoding.EncodeToString(data)

	req := common.RequestMessage{
		Operation: "upload",
		FileName:  fileName,
		FileData:  []byte(encodedData),
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	if !resp.Success {
		fmt.Printf("❌ Falha no upload!\n")
		fmt.Printf("   Mensagem: %s\n", resp.Message)
		return fmt.Errorf("upload falhou: %s", resp.Message)
	}

	fmt.Printf("✅ Upload realizado com sucesso!\n")
	fmt.Printf("   Arquivo: %s\n", fileName)
	fmt.Printf("   Tamanho: %d bytes\n", len(data))
	fmt.Printf("   Mensagem: %s\n", resp.Message)

	return nil
}

// DownloadFile faz download de um arquivo do servidor
func (c *Client) DownloadFile(fileName string, outputPath string) error {
	req := common.RequestMessage{
		Operation: "download",
		FileName:  fileName,
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("erro: %s", resp.Message)
	}

	// Decodifica os dados (se vierem em base64)
	data := resp.FileData
	if len(data) > 0 {
		decoded, err := base64.StdEncoding.DecodeString(string(data))
		if err == nil {
			data = decoded
		}
	}

	// Se outputPath não foi especificado, usa o nome do arquivo
	if outputPath == "" {
		outputPath = fileName
	}

	// Escreve o arquivo
	err = os.WriteFile(outputPath, data, 0644)
	if err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %w", err)
	}

	fmt.Printf("✅ Download realizado com sucesso!\n")
	fmt.Printf("   Arquivo: %s\n", fileName)
	fmt.Printf("   Tamanho: %d bytes\n", len(data))
	fmt.Printf("   Salvo em: %s\n", outputPath)

	return nil
}

