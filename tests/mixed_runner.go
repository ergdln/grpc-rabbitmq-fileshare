package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Parâmetros
	system := flag.String("system", "", "Sistema a testar: grpc, rabbit, ou ambos (all)")
	numClients := flag.Int("clients", 100, "Número de clientes concorrentes")
	opsPerClient := flag.Int("ops-per-client", 1500, "Número de operações por cliente (padrão: 1500 = 150k total com 100 clientes)")
	durationStr := flag.String("duration", "", "Duração máxima do teste (ex: 60s, 5m). Se vazio, usa ops-per-client")
	distributionStr := flag.String("distribution", "list:30,upload:35,download:35", "Distribuição de operações (ex: list:30,upload:35,download:35)")
	grpcAddr := flag.String("grpc-addr", "localhost:50051", "Endereço do servidor gRPC")
	amqpURL := flag.String("amqp-url", "amqp://guest:guest@localhost:5672/", "URL do RabbitMQ")
	outputDir := flag.String("output-dir", "../results", "Diretório para salvar resultados")
	tempDir := flag.String("temp-dir", "/tmp/benchmark", "Diretório temporário para arquivos de teste")
	flag.Parse()

	// Validações
	if *system == "" {
		log.Fatal("❌ Erro: --system é obrigatório (grpc, rabbit, ou all)")
	}

	// Parse distribuição
	distribution, err := parseDistribution(*distributionStr)
	if err != nil {
		log.Fatalf("❌ Erro ao parsear distribuição: %v", err)
	}

	// Parse duração
	var duration time.Duration
	if *durationStr != "" {
		duration, err = time.ParseDuration(*durationStr)
		if err != nil {
			log.Fatalf("❌ Erro ao parsear duração: %v", err)
		}
	} else {
		// Duração muito longa se não especificada (será limitada por ops-per-client)
		duration = 24 * time.Hour
	}

	// Cria diretório temporário
	if err := os.MkdirAll(*tempDir, 0755); err != nil {
		log.Fatalf("❌ Erro ao criar diretório temporário: %v", err)
	}

	// Cria diretório de saída
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("❌ Erro ao criar diretório de saída: %v", err)
	}

	// Gera arquivos de teste
	testFiles := make(map[int]string)
	fileSizes := []int{10, 1024, 10240} // 10KB, 1MB, 10MB
	for _, sizeKB := range fileSizes {
		filePath := filepath.Join(*tempDir, fmt.Sprintf("test_%dkb.dat", sizeKB))
		if err := generateTestFile(filePath, sizeKB*1024); err != nil {
			log.Fatalf("❌ Erro ao gerar arquivo de teste: %v", err)
		}
		testFiles[sizeKB] = filePath
		fmt.Printf("✅ Arquivo de teste gerado: %s (%d KB)\n", filePath, sizeKB)
	}

	// Determina sistemas a testar
	systems := []string{}
	if *system == "all" {
		systems = []string{"grpc", "rabbit"}
	} else {
		systems = []string{*system}
	}

	// Timestamp para arquivos
	timestamp := time.Now().Format("20060102_150405")

	// Gera sequência de operações UMA VEZ para garantir que ambos os sistemas
	// executem exatamente as mesmas operações na mesma ordem
	fmt.Println("\n🎲 Gerando sequência de operações determinística...")
	sequenceSeed := time.Now().UnixNano() // Seed baseado no timestamp
	operationSequence := GenerateOperationSequence(
		*numClients,
		*opsPerClient,
		distribution,
		fileSizes,
		sequenceSeed,
	)
	fmt.Printf("✅ Sequência gerada: %d clientes × %d operações = %d operações totais\n",
		*numClients, *opsPerClient, *numClients**opsPerClient)
	fmt.Printf("   Seed usado: %d (garante reprodutibilidade)\n", sequenceSeed)

	// Executa testes para cada sistema
	var summaryData []map[string]string

	for _, sys := range systems {
		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("  Teste de Concorrência Mista: %s\n", strings.ToUpper(sys))
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Cria runner
		outputFile := filepath.Join(*outputDir, fmt.Sprintf("mixed_concurrency_%s_%s.csv", sys, timestamp))
		runner, err := NewMixedConcurrencyRunner(outputFile)
		if err != nil {
			log.Fatalf("❌ Erro ao criar runner: %v", err)
		}

		// Executa teste com a mesma sequência para ambos os sistemas
		runner.RunMixedConcurrency(
			sys,
			*numClients,
			*opsPerClient,
			duration,
			distribution,
			fileSizes,
			testFiles,
			*grpcAddr,
			*amqpURL,
			operationSequence, // Passa a mesma sequência para ambos
		)

		// Fecha runner
		runner.Close()

		// Calcula estatísticas
		stats := runner.CalculateStatistics()
		if stats != nil {
			fmt.Println("\n📊 Estatísticas:")
			fmt.Printf("   Total de operações: %v\n", stats["total_operations"])
			fmt.Printf("   Sucessos: %v\n", stats["successful"])
			fmt.Printf("   Falhas: %v\n", stats["failed"])
			fmt.Printf("   Taxa de sucesso: %.2f%%\n", stats["success_rate"])

			if meanRTT, ok := stats["mean_rtt_ms"].(float64); ok {
				fmt.Printf("   RTT médio: %.3f ms\n", meanRTT)
				fmt.Printf("   RTT mínimo: %.3f ms\n", stats["min_rtt_ms"])
				fmt.Printf("   RTT máximo: %.3f ms\n", stats["max_rtt_ms"])
				fmt.Printf("   RTT p50: %.3f ms\n", stats["p50_rtt_ms"])
				fmt.Printf("   RTT p95: %.3f ms\n", stats["p95_rtt_ms"])
				fmt.Printf("   RTT p99: %.3f ms\n", stats["p99_rtt_ms"])
				fmt.Printf("   Desvio padrão: %.3f ms\n", stats["stddev_rtt_ms"])
			}

			if throughput, ok := stats["throughput_ops_per_sec"].(float64); ok {
				fmt.Printf("   Throughput: %.2f ops/segundo\n", throughput)
			}

			// Estatísticas por operação
			if byOp, ok := stats["by_operation"].(map[string]map[string]float64); ok {
				fmt.Println("\n   Por operação:")
				for op, opStats := range byOp {
					fmt.Printf("     %s: média=%.3f ms, p50=%.3f ms, p95=%.3f ms, p99=%.3f ms (count=%.0f)\n",
						op, opStats["mean_rtt_ms"], opStats["p50_rtt_ms"], opStats["p95_rtt_ms"], opStats["p99_rtt_ms"], opStats["count"])
				}
			}

			// Estatísticas por tamanho
			if bySize, ok := stats["by_file_size"].(map[int]map[string]float64); ok {
				fmt.Println("\n   Por tamanho de arquivo:")
				for size, sizeStats := range bySize {
					fmt.Printf("     %d KB: média=%.3f ms, p50=%.3f ms, p95=%.3f ms, p99=%.3f ms (count=%.0f)\n",
						size, sizeStats["mean_rtt_ms"], sizeStats["p50_rtt_ms"], sizeStats["p95_rtt_ms"], sizeStats["p99_rtt_ms"], sizeStats["count"])
				}
			}

			// Adiciona ao resumo
			summaryData = append(summaryData, map[string]string{
				"system":           sys,
				"total_operations": fmt.Sprintf("%v", stats["total_operations"]),
				"successful":       fmt.Sprintf("%v", stats["successful"]),
				"failed":           fmt.Sprintf("%v", stats["failed"]),
				"success_rate":     fmt.Sprintf("%.2f", stats["success_rate"]),
				"mean_rtt_ms":      fmt.Sprintf("%.3f", stats["mean_rtt_ms"]),
				"p50_rtt_ms":       fmt.Sprintf("%.3f", stats["p50_rtt_ms"]),
				"p95_rtt_ms":       fmt.Sprintf("%.3f", stats["p95_rtt_ms"]),
				"p99_rtt_ms":       fmt.Sprintf("%.3f", stats["p99_rtt_ms"]),
				"throughput":       fmt.Sprintf("%.2f", stats["throughput_ops_per_sec"]),
			})
		}

		fmt.Printf("\n✅ Resultados salvos em: %s\n", outputFile)
	}

	// Gera CSV de resumo
	if len(summaryData) > 0 {
		summaryFile := filepath.Join(*outputDir, fmt.Sprintf("mixed_concurrency_summary_%s.csv", timestamp))
		file, err := os.Create(summaryFile)
		if err == nil {
			writer := csv.NewWriter(file)
			header := []string{"system", "total_operations", "successful", "failed", "success_rate", "mean_rtt_ms", "p50_rtt_ms", "p95_rtt_ms", "p99_rtt_ms", "throughput_ops_per_sec"}
			writer.Write(header)

			for _, row := range summaryData {
				record := []string{
					row["system"],
					row["total_operations"],
					row["successful"],
					row["failed"],
					row["success_rate"],
					row["mean_rtt_ms"],
					row["p50_rtt_ms"],
					row["p95_rtt_ms"],
					row["p99_rtt_ms"],
					row["throughput"],
				}
				writer.Write(record)
			}

			writer.Flush()
			file.Close()
			fmt.Printf("\n📊 Resumo salvo em: %s\n", summaryFile)
		}
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  ✅ Todos os testes concluídos!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// generateTestFile gera um arquivo de teste com o tamanho especificado
func generateTestFile(filePath string, sizeBytes int) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Gera dados aleatórios
	data := make([]byte, sizeBytes)
	for i := range data {
		data[i] = byte(i % 256)
	}

	_, err = file.Write(data)
	return err
}
