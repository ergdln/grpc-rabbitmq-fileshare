#!/bin/bash

# Script para testar o comando download do cliente gRPC

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Teste: Download de arquivo (gRPC)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Verifica se um arquivo foi fornecido como argumento
if [ -z "$1" ]; then
    echo "❌ Erro: especifique o arquivo para download"
    echo "   Uso: $0 <arquivo> [arquivo_saida]"
    exit 1
fi

FILE_NAME="$1"
OUTPUT_FILE="${2:-downloaded_${FILE_NAME}}"

# Navega para o diretório do cliente
cd "$(dirname "$0")/../grpc-client" || exit 1

# Executa o comando download
echo "Arquivo solicitado: $FILE_NAME"
if [ -n "$2" ]; then
    echo "Arquivo de saída: $OUTPUT_FILE"
    echo ""
    echo "Executando: go run main.go client.go download \"$FILE_NAME\" \"$OUTPUT_FILE\""
    echo ""
    go run main.go client.go download "$FILE_NAME" "$OUTPUT_FILE"
else
    echo ""
    echo "Executando: go run main.go client.go download \"$FILE_NAME\""
    echo ""
    go run main.go client.go download "$FILE_NAME"
    OUTPUT_FILE="$FILE_NAME"
fi

echo ""
if [ -f "$OUTPUT_FILE" ]; then
    echo "✅ Arquivo baixado com sucesso!"
    echo "   Localização: $(pwd)/$OUTPUT_FILE"
    echo "   Tamanho: $(du -h "$OUTPUT_FILE" | cut -f1)"
else
    echo "⚠️  Arquivo não foi encontrado após o download"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Teste concluído"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

