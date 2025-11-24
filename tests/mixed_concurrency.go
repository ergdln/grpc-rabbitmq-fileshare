package main

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"grpc-rabbitmq-fileshare/common"
	"grpc-rabbitmq-fileshare/grpc-server/proto"

	"github.com/streadway/amqp"
)

// MixedResult representa um resultado de operação no teste de concorrência mista
type MixedResult struct {
	Timestamp  time.Time
	System     string
	Operation  string
	FileSizeKB int
	ClientID   int
	RTTMs      float64
	Success    bool
}

// MixedConcurrencyRunner executa testes de concorrência mista
type MixedConcurrencyRunner struct {
	results   []MixedResult
	resultsMu sync.Mutex
	csvWriter *csv.Writer
	csvFile   *os.File
	startTime time.Time
	endTime   time.Time
}

// OperationDistribution define a distribuição de operações
type OperationDistribution struct {
	List     int // percentual
	Upload   int // percentual
	Download int // percentual
}

// OperationSequenceItem representa uma operação na sequência
type OperationSequenceItem struct {
	Operation  string
	FileSizeKB int
	ClientID   int
	OpIndex    int // índice da operação para este cliente
}

// GenerateOperationSequence gera uma sequência de operações determinística
// que será reutilizada para ambos os sistemas (gRPC e RabbitMQ)
func GenerateOperationSequence(
	numClients int,
	opsPerClient int,
	distribution OperationDistribution,
	fileSizes []int,
	seed int64,
) [][]OperationSequenceItem {
	// Cria RNG com seed fixo para garantir reprodutibilidade
	rng := rand.New(rand.NewSource(seed))

	// Sequência: [cliente][operação]
	sequence := make([][]OperationSequenceItem, numClients)

	for clientID := 0; clientID < numClients; clientID++ {
		clientOps := make([]OperationSequenceItem, opsPerClient)
		for opIndex := 0; opIndex < opsPerClient; opIndex++ {
			// Seleciona operação baseada na distribuição
			operation := selectOperationFromDistribution(rng, distribution)
			fileSizeKB := 0

			if operation != "list" {
				// Seleciona tamanho aleatório
				fileSizeKB = fileSizes[rng.Intn(len(fileSizes))]
			}

			clientOps[opIndex] = OperationSequenceItem{
				Operation:  operation,
				FileSizeKB: fileSizeKB,
				ClientID:   clientID,
				OpIndex:    opIndex,
			}
		}
		sequence[clientID] = clientOps
	}

	return sequence
}

// selectOperationFromDistribution seleciona uma operação baseada na distribuição
func selectOperationFromDistribution(rng *rand.Rand, dist OperationDistribution) string {
	total := dist.List + dist.Upload + dist.Download
	if total == 0 {
		// Distribuição padrão se não especificada
		dist = OperationDistribution{List: 30, Upload: 35, Download: 35}
		total = 100
	}

	roll := rng.Intn(total)
	if roll < dist.List {
		return "list"
	} else if roll < dist.List+dist.Upload {
		return "upload"
	}
	return "download"
}

// NewMixedConcurrencyRunner cria um novo runner para concorrência mista
func NewMixedConcurrencyRunner(csvPath string) (*MixedConcurrencyRunner, error) {
	file, err := os.Create(csvPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar arquivo CSV: %w", err)
	}

	writer := csv.NewWriter(file)

	// Escreve cabeçalho com client_id
	header := []string{"timestamp", "system", "operation", "file_size_kb", "client_id", "rtt_ms", "success"}
	if err := writer.Write(header); err != nil {
		file.Close()
		return nil, fmt.Errorf("erro ao escrever cabeçalho: %w", err)
	}
	writer.Flush()

	return &MixedConcurrencyRunner{
		results:   make([]MixedResult, 0),
		csvWriter: writer,
		csvFile:   file,
		startTime: time.Now(),
	}, nil
}

// Close fecha o arquivo CSV
func (mcr *MixedConcurrencyRunner) Close() error {
	mcr.endTime = time.Now()
	mcr.csvWriter.Flush()
	return mcr.csvFile.Close()
}

// RecordResult registra um resultado
func (mcr *MixedConcurrencyRunner) RecordResult(result MixedResult) {
	mcr.resultsMu.Lock()
	defer mcr.resultsMu.Unlock()

	mcr.results = append(mcr.results, result)

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
		fmt.Sprintf("%d", result.ClientID),
		fmt.Sprintf("%.3f", result.RTTMs),
		success,
	}

	if err := mcr.csvWriter.Write(record); err != nil {
		fmt.Printf("⚠️  Erro ao escrever no CSV: %v\n", err)
	} else {
		mcr.csvWriter.Flush()
	}
}

// GetResults retorna todos os resultados
func (mcr *MixedConcurrencyRunner) GetResults() []MixedResult {
	mcr.resultsMu.Lock()
	defer mcr.resultsMu.Unlock()
	return mcr.results
}

// RunMixedConcurrency executa o teste de concorrência mista
// Se sequence for nil, gera uma nova sequência. Caso contrário, usa a sequência fornecida.
func (mcr *MixedConcurrencyRunner) RunMixedConcurrency(
	system string,
	numClients int,
	opsPerClient int,
	duration time.Duration,
	distribution OperationDistribution,
	fileSizes []int,
	testFiles map[int]string,
	grpcAddr string,
	amqpURL string,
	sequence [][]OperationSequenceItem, // Sequência de operações (nil para gerar nova)
) {
	fmt.Printf("🚀 Iniciando teste de concorrência mista: %s\n", system)
	fmt.Printf("   Clientes: %d\n", numClients)
	fmt.Printf("   Operações por cliente: %d\n", opsPerClient)
	fmt.Printf("   Duração máxima: %v\n", duration)
	fmt.Printf("   Distribuição: list=%d%%, upload=%d%%, download=%d%%\n",
		distribution.List, distribution.Upload, distribution.Download)
	if sequence != nil {
		fmt.Printf("   ✅ Usando sequência pré-definida (garantindo igualdade entre sistemas)\n")
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// Canal para sinalizar quando todos devem começar
	startBarrier := make(chan struct{})
	close(startBarrier) // Fecha imediatamente para que todos comecem juntos

	// Dispara clientes concorrentes
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			<-startBarrier // Espera sinal para começar

			// RNG apenas para delays (não para operações)
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(clientID)))

			opsCompleted := 0
			for opsCompleted < opsPerClient {
				select {
				case <-ctx.Done():
					return
				default:
					var operation string
					var fileSizeKB int
					var filePath string
					var fileName string

					// Usa sequência pré-definida se disponível
					if sequence != nil && clientID < len(sequence) && opsCompleted < len(sequence[clientID]) {
						seqItem := sequence[clientID][opsCompleted]
						operation = seqItem.Operation
						fileSizeKB = seqItem.FileSizeKB
					} else {
						// Fallback: gera aleatoriamente (não deveria acontecer se sequence foi gerada corretamente)
						operation = mcr.selectRandomOperation(rng, distribution)
						if operation != "list" {
							fileSizeKB = fileSizes[rng.Intn(len(fileSizes))]
						}
					}

					// Prepara arquivo se necessário
					if operation != "list" && fileSizeKB > 0 {
						filePath = testFiles[fileSizeKB]
						fileName = fmt.Sprintf("test_%dkb_%d.dat", fileSizeKB, clientID)
					}

					// Executa operação
					var result MixedResult
					if system == "grpc" {
						result = mcr.runGRPCOperation(grpcAddr, operation, filePath, fileName, fileSizeKB, clientID)
					} else {
						result = mcr.runRabbitOperation(amqpURL, operation, filePath, fileName, fileSizeKB, clientID)
					}

					mcr.RecordResult(result)
					opsCompleted++

					// Pequeno delay aleatório para não sobrecarregar
					time.Sleep(time.Duration(rng.Intn(10)) * time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()
}

// selectRandomOperation seleciona uma operação aleatória baseada na distribuição
// (mantido para compatibilidade, mas preferir usar GenerateOperationSequence)
func (mcr *MixedConcurrencyRunner) selectRandomOperation(rng *rand.Rand, dist OperationDistribution) string {
	return selectOperationFromDistribution(rng, dist)
}

// runGRPCOperation executa uma operação gRPC
func (mcr *MixedConcurrencyRunner) runGRPCOperation(
	serverAddr string,
	operation string,
	filePath string,
	fileName string,
	fileSizeKB int,
	clientID int,
) MixedResult {
	start := time.Now()
	var err error
	success := true

	// Obtém ou cria conexão gRPC
	conn, connErr := getOrCreateGRPCConnection(serverAddr)
	if connErr != nil {
		return MixedResult{
			Timestamp:  start,
			System:     "grpc",
			Operation:  operation,
			FileSizeKB: fileSizeKB,
			ClientID:   clientID,
			RTTMs:      0,
			Success:    false,
		}
	}

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
		// Para download, precisa de um arquivo que existe
		// Usa o primeiro arquivo disponível ou gera um nome baseado no tamanho
		if fileSizeKB > 0 {
			req := &proto.DownloadRequest{Name: fmt.Sprintf("test_%dkb.dat", fileSizeKB)}
			_, err = client.DownloadFile(ctx, req)
		}
	default:
		err = fmt.Errorf("operação desconhecida: %s", operation)
	}

	if err != nil {
		success = false
	}

	rtt := time.Since(start)
	rttMs := float64(rtt.Nanoseconds()) / 1e6

	return MixedResult{
		Timestamp:  start,
		System:     "grpc",
		Operation:  operation,
		FileSizeKB: fileSizeKB,
		ClientID:   clientID,
		RTTMs:      rttMs,
		Success:    success,
	}
}

// runRabbitOperation executa uma operação RabbitMQ
func (mcr *MixedConcurrencyRunner) runRabbitOperation(
	amqpURL string,
	operation string,
	filePath string,
	fileName string,
	fileSizeKB int,
	clientID int,
) MixedResult {
	start := time.Now()
	var err error
	success := true

	// Conecta ao RabbitMQ
	conn, connErr := amqp.Dial(amqpURL)
	if connErr != nil {
		return MixedResult{
			Timestamp:  start,
			System:     "rabbit",
			Operation:  operation,
			FileSizeKB: fileSizeKB,
			ClientID:   clientID,
			RTTMs:      0,
			Success:    false,
		}
	}
	defer conn.Close()

	channel, channelErr := conn.Channel()
	if channelErr != nil {
		return MixedResult{
			Timestamp:  start,
			System:     "rabbit",
			Operation:  operation,
			FileSizeKB: fileSizeKB,
			ClientID:   clientID,
			RTTMs:      0,
			Success:    false,
		}
	}
	defer channel.Close()

	// Declara fila de resposta
	responseQueue, queueErr := channel.QueueDeclare(
		"",    // nome gerado automaticamente
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if queueErr != nil {
		return MixedResult{
			Timestamp:  start,
			System:     "rabbit",
			Operation:  operation,
			FileSizeKB: fileSizeKB,
			ClientID:   clientID,
			RTTMs:      0,
			Success:    false,
		}
	}

	// Prepara requisição
	requestMsg := common.RequestMessage{
		Operation: operation,
		FileName:  fileName,
	}

	if operation == "upload" && filePath != "" {
		data, readErr := os.ReadFile(filePath)
		if readErr != nil {
			err = readErr
		} else {
			requestMsg.FileData = []byte(base64.StdEncoding.EncodeToString(data))
		}
	} else if operation == "download" && fileSizeKB > 0 {
		requestMsg.FileName = fmt.Sprintf("test_%dkb.dat", fileSizeKB)
	}

	requestBody, _ := json.Marshal(requestMsg)
	correlationID := fmt.Sprintf("%d_%d", clientID, time.Now().UnixNano())

	// Publica requisição
	err = channel.Publish(
		"",                  // exchange
		"rpc-file-requests", // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       responseQueue.Name,
			Body:          requestBody,
		},
	)

	if err != nil {
		return MixedResult{
			Timestamp:  start,
			System:     "rabbit",
			Operation:  operation,
			FileSizeKB: fileSizeKB,
			ClientID:   clientID,
			RTTMs:      0,
			Success:    false,
		}
	}

	// Consome resposta
	msgs, consumeErr := channel.Consume(
		responseQueue.Name,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	if consumeErr != nil {
		return MixedResult{
			Timestamp:  start,
			System:     "rabbit",
			Operation:  operation,
			FileSizeKB: fileSizeKB,
			ClientID:   clientID,
			RTTMs:      0,
			Success:    false,
		}
	}

	// Aguarda resposta com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	select {
	case msg := <-msgs:
		if msg.CorrelationId == correlationID {
			var resp common.ResponseMessage
			if jsonErr := json.Unmarshal(msg.Body, &resp); jsonErr == nil {
				success = resp.Success
			} else {
				success = false
			}
		} else {
			success = false
		}
	case <-ctx.Done():
		err = fmt.Errorf("timeout aguardando resposta")
		success = false
	}

	rtt := time.Since(start)
	rttMs := float64(rtt.Nanoseconds()) / 1e6

	return MixedResult{
		Timestamp:  start,
		System:     "rabbit",
		Operation:  operation,
		FileSizeKB: fileSizeKB,
		ClientID:   clientID,
		RTTMs:      rttMs,
		Success:    success,
	}
}

// CalculateStatistics calcula estatísticas agregadas
func (mcr *MixedConcurrencyRunner) CalculateStatistics() map[string]interface{} {
	results := mcr.GetResults()
	if len(results) == 0 {
		return nil
	}

	stats := make(map[string]interface{})

	// Separa por sucesso
	successful := make([]MixedResult, 0)
	failed := make([]MixedResult, 0)
	for _, r := range results {
		if r.Success {
			successful = append(successful, r)
		} else {
			failed = append(failed, r)
		}
	}

	// Estatísticas gerais
	stats["total_operations"] = len(results)
	stats["successful"] = len(successful)
	stats["failed"] = len(failed)
	stats["success_rate"] = float64(len(successful)) / float64(len(results)) * 100

	if len(successful) > 0 {
		// RTTs dos sucessos
		rtts := make([]float64, len(successful))
		for i, r := range successful {
			rtts[i] = r.RTTMs
		}
		sort.Float64s(rtts)

		// Estatísticas de RTT
		var sum float64
		for _, rtt := range rtts {
			sum += rtt
		}
		stats["mean_rtt_ms"] = sum / float64(len(rtts))
		stats["min_rtt_ms"] = rtts[0]
		stats["max_rtt_ms"] = rtts[len(rtts)-1]

		// Percentis
		stats["p50_rtt_ms"] = percentile(rtts, 50)
		stats["p95_rtt_ms"] = percentile(rtts, 95)
		stats["p99_rtt_ms"] = percentile(rtts, 99)

		// Desvio padrão
		var variance float64
		mean := stats["mean_rtt_ms"].(float64)
		for _, rtt := range rtts {
			variance += (rtt - mean) * (rtt - mean)
		}
		stats["stddev_rtt_ms"] = variance / float64(len(rtts))
	}

	// Throughput
	duration := mcr.endTime.Sub(mcr.startTime).Seconds()
	if duration > 0 {
		stats["throughput_ops_per_sec"] = float64(len(results)) / duration
	}

	// Estatísticas por operação
	byOperation := make(map[string][]MixedResult)
	for _, r := range successful {
		byOperation[r.Operation] = append(byOperation[r.Operation], r)
	}

	stats["by_operation"] = make(map[string]map[string]float64)
	for op, ops := range byOperation {
		if len(ops) > 0 {
			rtts := make([]float64, len(ops))
			for i, o := range ops {
				rtts[i] = o.RTTMs
			}
			sort.Float64s(rtts)

			var sum float64
			for _, rtt := range rtts {
				sum += rtt
			}

			opStats := make(map[string]float64)
			opStats["count"] = float64(len(ops))
			opStats["mean_rtt_ms"] = sum / float64(len(ops))
			opStats["p50_rtt_ms"] = percentile(rtts, 50)
			opStats["p95_rtt_ms"] = percentile(rtts, 95)
			opStats["p99_rtt_ms"] = percentile(rtts, 99)

			stats["by_operation"].(map[string]map[string]float64)[op] = opStats
		}
	}

	// Estatísticas por tamanho de arquivo
	bySize := make(map[int][]MixedResult)
	for _, r := range successful {
		if r.FileSizeKB > 0 {
			bySize[r.FileSizeKB] = append(bySize[r.FileSizeKB], r)
		}
	}

	stats["by_file_size"] = make(map[int]map[string]float64)
	for size, ops := range bySize {
		if len(ops) > 0 {
			rtts := make([]float64, len(ops))
			for i, o := range ops {
				rtts[i] = o.RTTMs
			}
			sort.Float64s(rtts)

			var sum float64
			for _, rtt := range rtts {
				sum += rtt
			}

			sizeStats := make(map[string]float64)
			sizeStats["count"] = float64(len(ops))
			sizeStats["mean_rtt_ms"] = sum / float64(len(ops))
			sizeStats["p50_rtt_ms"] = percentile(rtts, 50)
			sizeStats["p95_rtt_ms"] = percentile(rtts, 95)
			sizeStats["p99_rtt_ms"] = percentile(rtts, 99)

			stats["by_file_size"].(map[int]map[string]float64)[size] = sizeStats
		}
	}

	return stats
}

// percentile calcula o percentil de um slice ordenado
func percentile(sorted []float64, p int) float64 {
	if len(sorted) == 0 {
		return 0
	}
	index := float64(len(sorted)-1) * float64(p) / 100.0
	lower := int(index)
	upper := lower + 1
	if upper >= len(sorted) {
		return sorted[lower]
	}
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// parseDistribution parseia string de distribuição "list:30,upload:35,download:35"
func parseDistribution(distStr string) (OperationDistribution, error) {
	dist := OperationDistribution{List: 30, Upload: 35, Download: 35} // padrão

	if distStr == "" {
		return dist, nil
	}

	parts := strings.Split(distStr, ",")
	for _, part := range parts {
		kv := strings.Split(strings.TrimSpace(part), ":")
		if len(kv) != 2 {
			continue
		}

		op := strings.TrimSpace(kv[0])
		val, err := strconv.Atoi(strings.TrimSpace(kv[1]))
		if err != nil {
			continue
		}

		switch op {
		case "list":
			dist.List = val
		case "upload":
			dist.Upload = val
		case "download":
			dist.Download = val
		}
	}

	return dist, nil
}
