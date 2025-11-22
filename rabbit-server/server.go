package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"grpc-rabbitmq-fileshare/common"

	"github.com/streadway/amqp"
)

const (
	requestQueue = "rpc-file-requests"
)

// Server representa o servidor RabbitMQ
type Server struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	storage common.FileService
}

// NewServer cria uma nova instância do servidor RabbitMQ
func NewServer(amqpURL string, storage common.FileService) (*Server, error) {
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

	// Declara a fila de requisições
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
		return nil, fmt.Errorf("falha ao declarar fila: %w", err)
	}

	// Configura QoS para processar uma mensagem por vez
	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("falha ao configurar QoS: %w", err)
	}

	return &Server{
		conn:    conn,
		channel: channel,
		storage: storage,
	}, nil
}

// Start inicia o servidor e começa a consumir mensagens
func (s *Server) Start() error {
	log.Println("📡 Servidor RabbitMQ iniciado")
	log.Printf("📥 Consumindo da fila: %s", requestQueue)

	// Consome mensagens da fila
	msgs, err := s.channel.Consume(
		requestQueue, // queue
		"",           // consumer
		false,        // auto-ack (manual ack)
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return fmt.Errorf("falha ao registrar consumidor: %w", err)
	}

	// Processa mensagens
	go func() {
		for msg := range msgs {
			s.handleMessage(msg)
		}
	}()

	log.Println("✅ Servidor pronto para processar requisições")
	return nil
}

// handleMessage processa uma mensagem recebida
func (s *Server) handleMessage(msg amqp.Delivery) {
	log.Printf("[%s] Nova requisição recebida", msg.CorrelationId)

	// Decodifica a mensagem JSON
	var req common.RequestMessage
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		log.Printf("[%s] ❌ Erro ao decodificar mensagem: %v", msg.CorrelationId, err)
		s.sendErrorResponse(msg, fmt.Sprintf("erro ao decodificar mensagem: %v", err))
		msg.Nack(false, false) // Rejeita e não reenvia
		return
	}

	log.Printf("[%s] Operação: %s", msg.CorrelationId, req.Operation)

	// Processa a operação
	var resp common.ResponseMessage
	var err error

	switch req.Operation {
	case "list":
		resp, err = s.handleList()
	case "upload":
		resp, err = s.handleUpload(req)
	case "download":
		resp, err = s.handleDownload(req)
	default:
		err = fmt.Errorf("operação desconhecida: %s", req.Operation)
	}

	if err != nil {
		log.Printf("[%s] ❌ Erro ao processar operação: %v", msg.CorrelationId, err)
		s.sendErrorResponse(msg, err.Error())
		msg.Nack(false, false)
		return
	}

	// Envia resposta
	if err := s.sendResponse(msg, resp); err != nil {
		log.Printf("[%s] ❌ Erro ao enviar resposta: %v", msg.CorrelationId, err)
		msg.Nack(false, true) // Rejeita mas reenvia
		return
	}

	// Confirma processamento
	msg.Ack(false)
	log.Printf("[%s] ✅ Requisição processada com sucesso", msg.CorrelationId)
}

// handleList processa a operação de listar arquivos
func (s *Server) handleList() (common.ResponseMessage, error) {
	files, err := s.storage.ListFiles()
	if err != nil {
		return common.ResponseMessage{
			Success: false,
			Message: fmt.Sprintf("erro ao listar arquivos: %v", err),
		}, nil
	}

	log.Printf("📋 Listados %d arquivo(s)", len(files))
	return common.ResponseMessage{
		Success: true,
		Files:   files,
		Message: fmt.Sprintf("%d arquivo(s) encontrado(s)", len(files)),
	}, nil
}

// handleUpload processa a operação de upload
func (s *Server) handleUpload(req common.RequestMessage) (common.ResponseMessage, error) {
	if req.FileName == "" {
		return common.ResponseMessage{
			Success: false,
			Message: "nome do arquivo não pode ser vazio",
		}, nil
	}

	if len(req.FileData) == 0 {
		return common.ResponseMessage{
			Success: false,
			Message: "dados do arquivo não podem ser vazios",
		}, nil
	}

	// Decodifica os dados do arquivo (vêm como string base64 em FileData)
	// FileData é []byte mas contém uma string base64
	var data []byte
	if len(req.FileData) > 0 {
		decoded, err := base64.StdEncoding.DecodeString(string(req.FileData))
		if err != nil {
			return common.ResponseMessage{
				Success: false,
				Message: fmt.Sprintf("erro ao decodificar dados do arquivo: %v", err),
			}, nil
		}
		data = decoded
	}

	err := s.storage.UploadFile(req.FileName, data)
	if err != nil {
		return common.ResponseMessage{
			Success: false,
			Message: fmt.Sprintf("erro ao fazer upload: %v", err),
		}, nil
	}

	log.Printf("📤 Upload realizado: %s (%d bytes)", req.FileName, len(data))
	return common.ResponseMessage{
		Success: true,
		Message: fmt.Sprintf("arquivo %s enviado com sucesso", req.FileName),
	}, nil
}

// handleDownload processa a operação de download
func (s *Server) handleDownload(req common.RequestMessage) (common.ResponseMessage, error) {
	if req.FileName == "" {
		return common.ResponseMessage{
			Success: false,
			Message: "nome do arquivo não pode ser vazio",
		}, nil
	}

	data, err := s.storage.DownloadFile(req.FileName)
	if err != nil {
		return common.ResponseMessage{
			Success: false,
			Message: fmt.Sprintf("erro ao fazer download: %v", err),
		}, nil
	}

	log.Printf("📥 Download realizado: %s (%d bytes)", req.FileName, len(data))
	
	// Codifica os dados em base64 para JSON
	encodedData := base64.StdEncoding.EncodeToString(data)
	
	return common.ResponseMessage{
		Success:  true,
		FileName: req.FileName,
		FileData: []byte(encodedData),
		Message:  fmt.Sprintf("arquivo %s baixado com sucesso", req.FileName),
	}, nil
}

// sendResponse envia uma resposta para a fila de retorno
func (s *Server) sendResponse(msg amqp.Delivery, resp common.ResponseMessage) error {
	// Serializa a resposta
	body, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("erro ao serializar resposta: %w", err)
	}

	// Publica na fila de resposta (usando ReplyTo da mensagem original)
	err = s.channel.Publish(
		"",              // exchange
		msg.ReplyTo,     // routing key (fila de resposta)
		false,           // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: msg.CorrelationId,
			Body:          body,
		},
	)

	return err
}

// sendErrorResponse envia uma resposta de erro
func (s *Server) sendErrorResponse(msg amqp.Delivery, errorMsg string) {
	resp := common.ResponseMessage{
		Success: false,
		Message: errorMsg,
	}
	s.sendResponse(msg, resp)
}

// Close fecha as conexões
func (s *Server) Close() error {
	if s.channel != nil {
		s.channel.Close()
	}
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

