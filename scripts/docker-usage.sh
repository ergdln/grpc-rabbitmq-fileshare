#!/bin/bash

# Script de exemplo para usar os clientes Docker

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Docker Compose - File Sharing System"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Verifica se o .env existe
if [ ! -f .env ]; then
    echo "⚠️  Arquivo .env não encontrado. Copiando env.example..."
    cp env.example .env
    echo "✅ Arquivo .env criado. Ajuste as variáveis se necessário."
    echo ""
fi

case "${1:-help}" in
    start)
        echo "🚀 Iniciando serviços..."
        docker-compose up -d
        echo ""
        echo "✅ Serviços iniciados!"
        echo ""
        echo "📊 Acesse a interface do RabbitMQ em: http://localhost:15672"
        echo "   Usuário: guest (ou conforme .env)"
        echo "   Senha: guest (ou conforme .env)"
        ;;

    stop)
        echo "🛑 Parando serviços..."
        docker-compose down
        echo "✅ Serviços parados!"
        ;;

    logs)
        SERVICE="${2:-}"
        if [ -z "$SERVICE" ]; then
            docker-compose logs -f
        else
            docker-compose logs -f "$SERVICE"
        fi
        ;;

    grpc-list)
        echo "📋 Listando arquivos via gRPC..."
        docker-compose run --rm grpc-client list
        ;;

    grpc-upload)
        if [ -z "$2" ]; then
            echo "❌ Erro: especifique o arquivo para upload"
            echo "   Uso: $0 grpc-upload <arquivo>"
            exit 1
        fi
        FILE_PATH="$2"
        if [ ! -f "$FILE_PATH" ]; then
            echo "❌ Erro: arquivo não encontrado: $FILE_PATH"
            exit 1
        fi
        echo "📤 Fazendo upload via gRPC..."
        docker-compose run --rm -v "$(pwd):/workspace" grpc-client upload "/workspace/$FILE_PATH"
        ;;

    grpc-download)
        if [ -z "$2" ]; then
            echo "❌ Erro: especifique o arquivo para download"
            echo "   Uso: $0 grpc-download <arquivo> [saida]"
            exit 1
        fi
        FILE_NAME="$2"
        OUTPUT="${3:-}"
        echo "📥 Fazendo download via gRPC..."
        if [ -n "$OUTPUT" ]; then
            docker-compose run --rm -v "$(pwd):/workspace" grpc-client download "$FILE_NAME" "/workspace/$OUTPUT"
        else
            docker-compose run --rm -v "$(pwd):/workspace" grpc-client download "$FILE_NAME"
        fi
        ;;

    rabbit-list)
        echo "📋 Listando arquivos via RabbitMQ..."
        docker-compose run --rm rabbit-client list
        ;;

    rabbit-upload)
        if [ -z "$2" ]; then
            echo "❌ Erro: especifique o arquivo para upload"
            echo "   Uso: $0 rabbit-upload <arquivo>"
            exit 1
        fi
        FILE_PATH="$2"
        if [ ! -f "$FILE_PATH" ]; then
            echo "❌ Erro: arquivo não encontrado: $FILE_PATH"
            exit 1
        fi
        echo "📤 Fazendo upload via RabbitMQ..."
        docker-compose run --rm -v "$(pwd):/workspace" rabbit-client upload "/workspace/$FILE_PATH"
        ;;

    rabbit-download)
        if [ -z "$2" ]; then
            echo "❌ Erro: especifique o arquivo para download"
            echo "   Uso: $0 rabbit-download <arquivo> [saida]"
            exit 1
        fi
        FILE_NAME="$2"
        OUTPUT="${3:-}"
        echo "📥 Fazendo download via RabbitMQ..."
        if [ -n "$OUTPUT" ]; then
            docker-compose run --rm -v "$(pwd):/workspace" rabbit-client download "$FILE_NAME" "/workspace/$OUTPUT"
        else
            docker-compose run --rm -v "$(pwd):/workspace" rabbit-client download "$FILE_NAME"
        fi
        ;;

    scale-grpc)
        SCALE="${2:-2}"
        echo "📈 Escalando cliente gRPC para $SCALE instâncias..."
        docker-compose up -d --scale grpc-client="$SCALE"
        ;;

    scale-rabbit)
        SCALE="${2:-2}"
        echo "📈 Escalando cliente RabbitMQ para $SCALE instâncias..."
        docker-compose up -d --scale rabbit-client="$SCALE"
        ;;

    help|*)
        echo "Uso: $0 <comando> [argumentos]"
        echo ""
        echo "Comandos de gerenciamento:"
        echo "  start              Inicia todos os serviços"
        echo "  stop               Para todos os serviços"
        echo "  logs [servico]     Mostra logs (de todos ou de um serviço específico)"
        echo ""
        echo "Comandos gRPC:"
        echo "  grpc-list                    Lista arquivos"
        echo "  grpc-upload <arquivo>        Faz upload de arquivo"
        echo "  grpc-download <arquivo> [saida]  Faz download de arquivo"
        echo ""
        echo "Comandos RabbitMQ:"
        echo "  rabbit-list                 Lista arquivos"
        echo "  rabbit-upload <arquivo>      Faz upload de arquivo"
        echo "  rabbit-download <arquivo> [saida]  Faz download de arquivo"
        echo ""
        echo "Escalamento:"
        echo "  scale-grpc <numero>         Escala clientes gRPC"
        echo "  scale-rabbit <numero>       Escala clientes RabbitMQ"
        echo ""
        echo "Exemplos:"
        echo "  $0 start"
        echo "  $0 grpc-list"
        echo "  $0 grpc-upload arquivo.txt"
        echo "  $0 rabbit-download arquivo.txt copia.txt"
        ;;
esac

