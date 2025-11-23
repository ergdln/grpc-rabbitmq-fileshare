package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func main() {
	// Parâmetros
	system := flag.String("system", "", "Sistema a testar: grpc, rabbit, ou ambos (all)")
	operation := flag.String("operation", "", "Operação: list, upload, download, ou todas (all)")
	fileSizeKB := flag.Int("file-size-kb", 0, "Tamanho do arquivo em KB (10, 1024, 10240)")
	numClients := flag.Int("clients", 1, "Número de clientes concorrentes")
	numOperations := flag.Int("ops", 10000, "Número total de operações")
	grpcAddr := flag.String("grpc-addr", "localhost:50051", "Endereço do servidor gRPC")
	amqpURL := flag.String("amqp-url", "amqp://guest:guest@localhost:5672/", "URL do RabbitMQ")
	outputCSV := flag.String("output", "benchmark_results.csv", "Arquivo CSV de saída")
	tempDir := flag.String("temp-dir", "/tmp/benchmark", "Diretório temporário para arquivos de teste")
	flag.Parse()

	// Validações
	if *system == "" {
		log.Fatal("❌ Erro: --system é obrigatório (grpc, rabbit, ou all)")
	}
	if *operation == "" {
		log.Fatal("❌ Erro: --operation é obrigatório (list, upload, download, ou all)")
	}

	// Cria diretório temporário
	if err := os.MkdirAll(*tempDir, 0755); err != nil {
		log.Fatalf("❌ Erro ao criar diretório temporário: %v", err)
	}

	// Cria runner
	runner, err := NewBenchmarkRunner(*outputCSV)
	if err != nil {
		log.Fatalf("❌ Erro ao criar runner: %v", err)
	}
	defer runner.Close()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Benchmark Runner - File Sharing System")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Sistema: %s\n", *system)
	fmt.Printf("Operação: %s\n", *operation)
	fmt.Printf("Tamanho do arquivo: %d KB\n", *fileSizeKB)
	fmt.Printf("Número de clientes: %d\n", *numClients)
	fmt.Printf("Número de operações: %d\n", *numOperations)
	fmt.Printf("Arquivo de saída: %s\n", *outputCSV)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Gera arquivos de teste se necessário
	testFiles := make(map[int]string)
	if *operation == "upload" || *operation == "download" || *operation == "all" {
		if *fileSizeKB > 0 {
			filePath := filepath.Join(*tempDir, fmt.Sprintf("test_%dkb.dat", *fileSizeKB))
			if err := generateTestFile(filePath, *fileSizeKB*1024); err != nil {
				log.Fatalf("❌ Erro ao gerar arquivo de teste: %v", err)
			}
			testFiles[*fileSizeKB] = filePath
			fmt.Printf("✅ Arquivo de teste gerado: %s (%d KB)\n", filePath, *fileSizeKB)
		} else {
			// Gera arquivos para todos os tamanhos padrão
			sizes := []int{10, 1024, 10240}
			for _, size := range sizes {
				filePath := filepath.Join(*tempDir, fmt.Sprintf("test_%dkb.dat", size))
				if err := generateTestFile(filePath, size*1024); err != nil {
					log.Fatalf("❌ Erro ao gerar arquivo de teste: %v", err)
				}
				testFiles[size] = filePath
			}
			fmt.Printf("✅ Arquivos de teste gerados para tamanhos: 10KB, 1MB, 10MB\n")
		}
	}
	fmt.Println()

	// Determina sistemas e operações a testar
	systems := []string{}
	if *system == "all" {
		systems = []string{"grpc", "rabbit"}
	} else {
		systems = []string{*system}
	}

	operations := []string{}
	if *operation == "all" {
		operations = []string{"list", "upload", "download"}
	} else {
		operations = []string{*operation}
	}

	// Executa benchmarks
	startTime := time.Now()

	for _, sys := range systems {
		for _, op := range operations {
			// Determina tamanhos de arquivo
			fileSizes := []int{}
			if op == "list" {
				fileSizes = []int{0} // list não usa tamanho
			} else if *fileSizeKB > 0 {
				fileSizes = []int{*fileSizeKB}
			} else {
				fileSizes = []int{10, 1024, 10240} // 10KB, 1MB, 10MB
			}

			for _, sizeKB := range fileSizes {
				fmt.Printf("🚀 Executando: %s/%s (arquivo: %d KB, clientes: %d)\n", sys, op, sizeKB, *numClients)

				// Warm-up: cria conexões antes de começar a medir
				if sys == "grpc" {
					fmt.Printf("   🔥 Aquecendo conexões gRPC...\n")
					for i := 0; i < *numClients; i++ {
						// Cria conexão de warm-up (não registra no CSV)
						if err := WarmUpGRPCConnection(*grpcAddr, op); err != nil {
							fmt.Printf("   ⚠️  Aviso: erro no warm-up: %v\n", err)
						}
					}
					fmt.Printf("   ✅ Conexões aquecidas\n")
				}

				// Calcula operações por cliente
				opsPerClient := *numOperations / *numClients
				remainingOps := *numOperations % *numClients

				var wg sync.WaitGroup
				startBarrier := sync.WaitGroup{}
				startBarrier.Add(*numClients)

				// Dispara clientes concorrentes
				for i := 0; i < *numClients; i++ {
					wg.Add(1)
					clientOps := opsPerClient
					if i < remainingOps {
						clientOps++
					}

					go func(clientID, ops int) {
						defer wg.Done()

						// Espera todos os clientes estarem prontos
						startBarrier.Done()
						startBarrier.Wait()

						// Executa operações
						for j := 0; j < ops; j++ {
							var filePath, fileName string
							if op != "list" && sizeKB > 0 {
								filePath = testFiles[sizeKB]
								fileName = filepath.Base(filePath)
							}

							if sys == "grpc" {
								runner.RunGRPCOperation(*grpcAddr, op, filePath, fileName, sizeKB, *numClients)
							} else {
								runner.RunRabbitOperation(*amqpURL, op, filePath, fileName, sizeKB, *numClients)
							}

							// Pequeno delay para não sobrecarregar
							time.Sleep(1 * time.Millisecond)
						}
					}(i, clientOps)
				}

				wg.Wait()
				fmt.Printf("✅ Concluído: %s/%s (%d KB)\n\n", sys, op, sizeKB)
			}
		}
	}

	totalTime := time.Since(startTime)

	// Mostra estatísticas
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Estatísticas")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	stats := runner.GetStats()
	if stats != nil {
		for key, value := range stats {
			fmt.Printf("%s: %v\n", key, value)
		}
	}
	fmt.Printf("Tempo total: %v\n", totalTime)
	fmt.Printf("Operações/segundo: %.2f\n", float64(*numOperations)/totalTime.Seconds())
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("\n✅ Resultados salvos em: %s\n", *outputCSV)
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
