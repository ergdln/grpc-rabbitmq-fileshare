package main

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"grpc-rabbitmq-fileshare/common"
	"grpc-rabbitmq-fileshare/grpc-server/proto"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Result representa um resultado de operação
type Result struct {
	Timestamp  time.Time
	System     string
	Operation  string
	FileSizeKB int
	Clients    int
	RTTMs      float64
	Success    bool
	Error      error
}

// BenchmarkRunner executa benchmarks
type BenchmarkRunner struct {
	results   []Result
	resultsMu sync.Mutex
	csvWriter *csv.Writer
	csvFile   *os.File
}

// NewBenchmarkRunner cria um novo runner de benchmark
func NewBenchmarkRunner(csvPath string) (*BenchmarkRunner, error) {
	file, err := os.Create(csvPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar arquivo CSV: %w", err)
	}

	writer := csv.NewWriter(file)

	// Escreve cabeçalho
	header := []string{"timestamp", "system", "operation", "file_size_kb", "clients", "rtt_ms", "success"}
	if err := writer.Write(header); err != nil {
		file.Close()
		return nil, fmt.Errorf("erro ao escrever cabeçalho: %w", err)
	}
	writer.Flush()

	return &BenchmarkRunner{
		results:   make([]Result, 0),
		csvWriter: writer,
		csvFile:   file,
	}, nil
}

// Close fecha o arquivo CSV
func (br *BenchmarkRunner) Close() error {
	br.csvWriter.Flush()
	return br.csvFile.Close()
}

// RecordResult registra um resultado
func (br *BenchmarkRunner) RecordResult(result Result) {
	br.resultsMu.Lock()
	defer br.resultsMu.Unlock()

	br.results = append(br.results, result)

	// Escreve no CSV
	success := "true"
	if !result.Success {
		success = "false"
	}

	record := []string{
		result.Timestamp.Format(time.RFC3339Nano),
		result.System,
		result.Operation,
		fmt.Sprintf("%d", result.FileSizeKB),
		fmt.Sprintf("%d", result.Clients),
		fmt.Sprintf("%.3f", result.RTTMs),
		success,
	}

	if err := br.csvWriter.Write(record); err != nil {
		fmt.Printf("⚠️  Erro ao escrever no CSV: %v\n", err)
	} else {
		br.csvWriter.Flush()
	}
}

// GetOrCreateGRPCConnection obtém ou cria uma conexão gRPC (reutiliza conexões)
var grpcConnections = make(map[string]*grpc.ClientConn)
var grpcConnMu sync.Mutex

func getOrCreateGRPCConnection(serverAddr string) (*grpc.ClientConn, error) {
	grpcConnMu.Lock()
	defer grpcConnMu.Unlock()

	if conn, exists := grpcConnections[serverAddr]; exists {
		// Verifica se a conexão ainda está válida
		state := conn.GetState()
		if state.String() == "READY" || state.String() == "IDLE" {
			return conn, nil
		}
		// Conexão inválida, remove e cria nova
		conn.Close()
		delete(grpcConnections, serverAddr)
	}

	// Cria nova conexão com tamanho máximo de mensagem de 50MB
	// Isso permite upload/download de arquivos de até 50MB
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(50*1024*1024), // 50MB
			grpc.MaxCallSendMsgSize(50*1024*1024), // 50MB
		),
	)
	if err != nil {
		return nil, err
	}

	grpcConnections[serverAddr] = conn
	return conn, nil
}

// WarmUpGRPCConnection faz warm-up de uma conexão gRPC sem registrar no CSV
func WarmUpGRPCConnection(serverAddr string, operation string) error {
	conn, err := getOrCreateGRPCConnection(serverAddr)
	if err != nil {
		return err
	}

	client := proto.NewFileServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Executa uma operação simples para aquecer a conexão
	switch operation {
	case "list":
		_, err = client.ListFiles(ctx, &proto.Empty{})
	default:
		// Para outras operações, faz list como warm-up
		_, err = client.ListFiles(ctx, &proto.Empty{})
	}

	return err
}

// RunGRPCOperation executa uma operação gRPC e mede o RTT
func (br *BenchmarkRunner) RunGRPCOperation(
	serverAddr string,
	operation string,
	filePath string,
	fileName string,
	fileSizeKB int,
	clientNum int,
) {
	start := time.Now()
	var err error
	success := true

	// Obtém ou cria conexão gRPC (reutiliza conexões)
	conn, connErr := getOrCreateGRPCConnection(serverAddr)
	if connErr != nil {
		br.RecordResult(Result{
			Timestamp:  start,
			System:     "grpc",
			Operation:  operation,
			FileSizeKB: fileSizeKB,
			Clients:    clientNum,
			RTTMs:      0,
			Success:    false,
			Error:      connErr,
		})
		return
	}
	// NÃO fecha a conexão aqui - ela será reutilizada

	client := proto.NewFileServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch operation {
	case "list":
		_, err = client.ListFiles(ctx, &proto.Empty{})
	case "upload":
		if filePath != "" {
			data, readErr := os.ReadFile(filePath)
			if readErr != nil {
				err = readErr
			} else {
				req := &proto.UploadRequest{
					Name: fileName,
					Data: data,
				}
				resp, callErr := client.UploadFile(ctx, req)
				if callErr != nil {
					err = callErr
				} else if !resp.Success {
					err = fmt.Errorf(resp.Message)
				}
			}
		}
	case "download":
		req := &proto.DownloadRequest{Name: fileName}
		_, err = client.DownloadFile(ctx, req)
	default:
		err = fmt.Errorf("operação desconhecida: %s", operation)
	}

	if err != nil {
		success = false
	}

	rtt := time.Since(start)
	rttMs := float64(rtt.Nanoseconds()) / 1e6

	br.RecordResult(Result{
		Timestamp:  start,
		System:     "grpc",
		Operation:  operation,
		FileSizeKB: fileSizeKB,
		Clients:    clientNum,
		RTTMs:      rttMs,
		Success:    success,
		Error:      err,
	})
}

// RunRabbitOperation executa uma operação RabbitMQ e mede o RTT
func (br *BenchmarkRunner) RunRabbitOperation(
	amqpURL string,
	operation string,
	filePath string,
	fileName string,
	fileSizeKB int,
	clientNum int,
) {
	start := time.Now()
	var err error
	success := true

	// Conecta ao RabbitMQ
	conn, connErr := amqp.Dial(amqpURL)
	if connErr != nil {
		br.RecordResult(Result{
			Timestamp:  start,
			System:     "rabbit",
			Operation:  operation,
			FileSizeKB: fileSizeKB,
			Clients:    clientNum,
			RTTMs:      0,
			Success:    false,
			Error:      connErr,
		})
		return
	}
	defer conn.Close()

	channel, channelErr := conn.Channel()
	if channelErr != nil {
		br.RecordResult(Result{
			Timestamp:  start,
			System:     "rabbit",
			Operation:  operation,
			FileSizeKB: fileSizeKB,
			Clients:    clientNum,
			RTTMs:      0,
			Success:    false,
			Error:      channelErr,
		})
		return
	}
	defer channel.Close()

	// Cria fila de resposta
	replyQueue, queueErr := channel.QueueDeclare("", false, true, true, false, nil)
	if queueErr != nil {
		br.RecordResult(Result{
			Timestamp:  start,
			System:     "rabbit",
			Operation:  operation,
			FileSizeKB: fileSizeKB,
			Clients:    clientNum,
			RTTMs:      0,
			Success:    false,
			Error:      queueErr,
		})
		return
	}

	// Prepara requisição
	req := common.RequestMessage{Operation: operation}
	if operation == "upload" && filePath != "" {
		data, readErr := os.ReadFile(filePath)
		if readErr != nil {
			err = readErr
		} else {
			req.FileName = fileName
			req.FileData = []byte(base64.StdEncoding.EncodeToString(data))
		}
	} else if operation == "download" {
		req.FileName = fileName
	}

	if err == nil {
		// Envia requisição e aguarda resposta
		correlationID := fmt.Sprintf("%d", time.Now().UnixNano())
		body, _ := json.Marshal(req)

		msgs, consumeErr := channel.Consume(replyQueue.Name, "", true, false, false, false, nil)
		if consumeErr != nil {
			err = consumeErr
		} else {
			pubErr := channel.Publish("", "rpc-file-requests", false, false, amqp.Publishing{
				ContentType:   "application/json",
				CorrelationId: correlationID,
				ReplyTo:       replyQueue.Name,
				Body:          body,
			})

			if pubErr != nil {
				err = pubErr
			} else {
				timeout := time.After(30 * time.Second)
				select {
				case msg := <-msgs:
					if msg.CorrelationId == correlationID {
						var resp common.ResponseMessage
						if unmarshalErr := json.Unmarshal(msg.Body, &resp); unmarshalErr != nil {
							err = unmarshalErr
						} else if !resp.Success {
							err = fmt.Errorf(resp.Message)
						}
					}
				case <-timeout:
					err = fmt.Errorf("timeout aguardando resposta")
				}
			}
		}
	}

	if err != nil {
		success = false
	}

	rtt := time.Since(start)
	rttMs := float64(rtt.Nanoseconds()) / 1e6

	br.RecordResult(Result{
		Timestamp:  start,
		System:     "rabbit",
		Operation:  operation,
		FileSizeKB: fileSizeKB,
		Clients:    clientNum,
		RTTMs:      rttMs,
		Success:    success,
		Error:      err,
	})
}

// GetStats retorna estatísticas dos resultados
func (br *BenchmarkRunner) GetStats() map[string]interface{} {
	br.resultsMu.Lock()
	defer br.resultsMu.Unlock()

	if len(br.results) == 0 {
		return nil
	}

	var totalRTT float64
	var successCount int
	var minRTT, maxRTT float64

	for i, result := range br.results {
		if result.Success {
			successCount++
			totalRTT += result.RTTMs

			if i == 0 || result.RTTMs < minRTT {
				minRTT = result.RTTMs
			}
			if result.RTTMs > maxRTT {
				maxRTT = result.RTTMs
			}
		}
	}

	avgRTT := totalRTT / float64(successCount)
	successRate := float64(successCount) / float64(len(br.results)) * 100

	return map[string]interface{}{
		"total_operations": len(br.results),
		"successful":       successCount,
		"failed":           len(br.results) - successCount,
		"success_rate":     fmt.Sprintf("%.2f%%", successRate),
		"avg_rtt_ms":       fmt.Sprintf("%.3f", avgRTT),
		"min_rtt_ms":       fmt.Sprintf("%.3f", minRTT),
		"max_rtt_ms":       fmt.Sprintf("%.3f", maxRTT),
	}
}
