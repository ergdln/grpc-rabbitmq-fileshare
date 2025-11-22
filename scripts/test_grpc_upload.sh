#!/bin/bash

# Script para testar o comando upload do cliente gRPC

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Teste: Upload de arquivo (gRPC)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Verifica se um arquivo foi fornecido como argumento
if [ -z "$1" ]; then
    echo "❌ Erro: especifique o arquivo para upload"
    echo "   Uso: $0 <arquivo>"
    exit 1
fi

FILE_PATH="$1"

# Verifica se o arquivo existe
if [ ! -f "$FILE_PATH" ]; then
    echo "❌ Erro: arquivo não encontrado: $FILE_PATH"
    exit 1
fi

# Navega para o diretório do cliente
cd "$(dirname "$0")/../grpc-client" || exit 1

# Executa o comando upload
echo "Arquivo: $FILE_PATH"
echo "Tamanho: $(du -h "$FILE_PATH" | cut -f1)"
echo ""
echo "Executando: go run main.go client.go upload \"$FILE_PATH\""
echo ""
go run main.go client.go upload "$FILE_PATH"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Teste concluído"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

