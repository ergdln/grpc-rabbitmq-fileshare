package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	defaultServerAddr = "localhost:50051"
)

func main() {
	// Obtém o endereço do servidor da variável de ambiente ou usa o padrão
	serverAddrEnv := os.Getenv("GRPC_SERVER_ADDR")
	if serverAddrEnv == "" {
		serverAddrEnv = defaultServerAddr
	}

	// Define flags
	serverAddr := flag.String("server", serverAddrEnv, "Endereço do servidor gRPC (ex: localhost:50051)")
	flag.Parse()

	// Verifica se há argumentos suficientes
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	// Cria o cliente
	client, err := NewClient(*serverAddr)
	if err != nil {
		log.Fatalf("Erro ao criar cliente: %v", err)
	}
	defer client.Close()

	// Processa o comando
	command := args[0]

	switch command {
	case "list":
		if err := client.ListFiles(); err != nil {
			log.Fatalf("Erro ao listar arquivos: %v", err)
		}

	case "upload":
		if len(args) < 2 {
			fmt.Println("❌ Erro: especifique o arquivo para upload")
			fmt.Println("   Uso: upload <arquivo>")
			os.Exit(1)
		}
		filePath := args[1]
		if err := client.UploadFile(filePath); err != nil {
			log.Fatalf("Erro ao fazer upload: %v", err)
		}

	case "download":
		if len(args) < 2 {
			fmt.Println("❌ Erro: especifique o arquivo para download")
			fmt.Println("   Uso: download <arquivo> [arquivo_saida]")
			os.Exit(1)
		}
		fileName := args[1]
		outputPath := ""
		if len(args) >= 3 {
			outputPath = args[2]
		}
		if err := client.DownloadFile(fileName, outputPath); err != nil {
			log.Fatalf("Erro ao fazer download: %v", err)
		}

	default:
		fmt.Printf("❌ Comando desconhecido: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  gRPC Client - File Sharing System")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Uso:")
	fmt.Println("  go run main.go client.go [flags] <comando> [argumentos]")
	fmt.Println()
	fmt.Println("Comandos:")
	fmt.Println("  list                          Lista todos os arquivos no servidor")
	fmt.Println("  upload <arquivo>              Faz upload de um arquivo")
	fmt.Println("  download <arquivo> [saida]    Faz download de um arquivo")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -server <endereço>            Endereço do servidor (padrão: localhost:50051)")
	fmt.Println()
	fmt.Println("Exemplos:")
	fmt.Println("  go run main.go client.go list")
	fmt.Println("  go run main.go client.go upload arquivo.txt")
	fmt.Println("  go run main.go client.go download arquivo.txt")
	fmt.Println("  go run main.go client.go download arquivo.txt copia.txt")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
