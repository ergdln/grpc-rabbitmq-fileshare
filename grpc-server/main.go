package main

import (
	"flag"
	"log"
	"os"

	"grpc-rabbitmq-fileshare/common"
)

func main() {
	// Define flags para configuração
	port := flag.String("port", "50051", "Porta para o servidor gRPC escutar")
	dataDir := flag.String("data-dir", "./data", "Diretório para armazenar arquivos")
	flag.Parse()

	log.Println("=== gRPC Server - File Sharing System ===")
	log.Printf("Iniciando servidor na porta %s", *port)
	log.Printf("Diretório de dados: %s", *dataDir)

	// Cria o serviço de armazenamento local
	storage, err := common.NewLocalStorage(*dataDir)
	if err != nil {
		log.Fatalf("Erro ao criar serviço de armazenamento: %v", err)
		os.Exit(1)
	}

	log.Println("Serviço de armazenamento inicializado com sucesso")

	// Inicia o servidor gRPC
	if err := StartServer(*port, storage); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
		os.Exit(1)
	}
}

