#!/bin/bash

# Script para executar experimentos com concorrência aleatória

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TESTS_DIR="$PROJECT_ROOT/tests"
RESULTS_DIR="$PROJECT_ROOT/results"

# Cria diretório de resultados
mkdir -p "$RESULTS_DIR"

cd "$PROJECT_ROOT"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Executando Experimentos com Concorrência Aleatória"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Parâmetros
NUM_OPS=10000
CLIENTS=(1 10 20)
FILE_SIZES_KB=(10 1024 10240)
OPERATIONS=("list" "upload" "download")
SYSTEMS=("grpc" "rabbit")

# Configurações padrão
GRPC_ADDR="${GRPC_ADDR:-localhost:50051}"
AMQP_URL="${AMQP_URL:-amqp://guest:guest@localhost:5672/}"

# Compila o runner se necessário
if [ ! -f "$TESTS_DIR/runner" ]; then
    echo "🔨 Compilando runner..."
    cd "$TESTS_DIR"
    go build -o runner runner.go benchmark.go
    echo "✅ Runner compilado"
    echo ""
fi

cd "$PROJECT_ROOT"

# Gera timestamp para o arquivo de saída
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
OUTPUT_FILE="$RESULTS_DIR/random_concurrency_${TIMESTAMP}.csv"

echo "📝 Arquivo de saída: $OUTPUT_FILE"
echo ""

# Número de experimentos aleatórios
NUM_EXPERIMENTS="${1:-50}"

echo "🎲 Executando $NUM_EXPERIMENTS experimentos aleatórios..."
echo ""

for i in $(seq 1 $NUM_EXPERIMENTS); do
    # Seleciona parâmetros aleatórios
    system=${SYSTEMS[$RANDOM % ${#SYSTEMS[@]}]}
    operation=${OPERATIONS[$RANDOM % ${#OPERATIONS[@]}]}
    numClients=${CLIENTS[$RANDOM % ${#CLIENTS[@]}]}
    
    # Para list, não precisa de tamanho de arquivo
    if [ "$operation" == "list" ]; then
        fileSizeKB=0
    else
        fileSizeKB=${FILE_SIZES_KB[$RANDOM % ${#FILE_SIZES_KB[@]}]}
    fi

    echo "[$i/$NUM_EXPERIMENTS] $system/$operation (${fileSizeKB}KB, $numClients clientes)"

    # Anexa ao arquivo CSV (remove cabeçalho após primeira execução)
    if [ $i -eq 1 ]; then
        "$TESTS_DIR/runner" \
            --system "$system" \
            --operation "$operation" \
            --file-size-kb "$fileSizeKB" \
            --clients "$numClients" \
            --ops "$NUM_OPS" \
            --grpc-addr "$GRPC_ADDR" \
            --amqp-url "$AMQP_URL" \
            --output "$OUTPUT_FILE" \
            --temp-dir "/tmp/benchmark"
    else
        # Para anexar, precisamos modificar o runner ou usar tail para remover cabeçalho
        TEMP_FILE=$(mktemp)
        "$TESTS_DIR/runner" \
            --system "$system" \
            --operation "$operation" \
            --file-size-kb "$fileSizeKB" \
            --clients "$numClients" \
            --ops "$NUM_OPS" \
            --grpc-addr "$GRPC_ADDR" \
            --amqp-url "$AMQP_URL" \
            --output "$TEMP_FILE" \
            --temp-dir "/tmp/benchmark"
        # Anexa sem cabeçalho
        tail -n +2 "$TEMP_FILE" >> "$OUTPUT_FILE"
        rm -f "$TEMP_FILE"
    fi

    # Pequeno delay entre experimentos
    sleep 1
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Experimentos Aleatórios Concluídos!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Resultados salvos em: $OUTPUT_FILE"
echo ""

