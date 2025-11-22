package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"grpc-rabbitmq-fileshare/common"
)

const (
	defaultAMQPURL = "amqp://guest:guest@localhost:5672/"
	defaultDataDir = "./data"
)

func main() {
	// Define flags
	amqpURL := flag.String("amqp-url", defaultAMQPURL, "URL de conexão do RabbitMQ")
	dataDir := flag.String("data-dir", defaultDataDir, "Diretório para armazenar arquivos")
	flag.Parse()

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("  RabbitMQ Server - File Sharing System")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("Conectando ao RabbitMQ: %s", *amqpURL)
	log.Printf("Diretório de dados: %s", *dataDir)

	// Cria o serviço de armazenamento local
	storage, err := common.NewLocalStorage(*dataDir)
	if err != nil {
		log.Fatalf("Erro ao criar serviço de armazenamento: %v", err)
	}

	log.Println("Serviço de armazenamento inicializado com sucesso")

	// Cria o servidor RabbitMQ
	server, err := NewServer(*amqpURL, storage)
	if err != nil {
		log.Fatalf("Erro ao criar servidor: %v", err)
	}
	defer server.Close()

	// Inicia o servidor
	if err := server.Start(); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}

	// Aguarda sinal de interrupção
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("Pressione Ctrl+C para encerrar o servidor")
	<-sigChan

	log.Println("\nEncerrando servidor...")
}
