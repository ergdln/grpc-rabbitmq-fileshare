#!/bin/bash

# Script para executar todos os experimentos de benchmark

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TESTS_DIR="$PROJECT_ROOT/tests"
RESULTS_DIR="$PROJECT_ROOT/results"

# Cria diretório de resultados
mkdir -p "$RESULTS_DIR"

cd "$PROJECT_ROOT"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Executando Todos os Experimentos de Benchmark"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Parâmetros
NUM_OPS=10000
CLIENTS=(1 10 20)
FILE_SIZES_KB=(10 1024 10240)  # 10KB, 1MB, 10MB
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

# Executa experimentos
for system in "${SYSTEMS[@]}"; do
    for operation in "${OPERATIONS[@]}"; do
        for fileSizeKB in "${FILE_SIZES_KB[@]}"; do
            # list não precisa de tamanho de arquivo
            if [ "$operation" == "list" ]; then
                fileSizeKB=0
            fi

            for numClients in "${CLIENTS[@]}"; do
                outputFile="$RESULTS_DIR/benchmark_${system}_${operation}_${fileSizeKB}kb_${numClients}clients.csv"
                
                echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                echo "  Sistema: $system"
                echo "  Operação: $operation"
                echo "  Tamanho: ${fileSizeKB} KB"
                echo "  Clientes: $numClients"
                echo "  Operações: $NUM_OPS"
                echo "  Saída: $outputFile"
                echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                echo ""

                "$TESTS_DIR/runner" \
                    --system "$system" \
                    --operation "$operation" \
                    --file-size-kb "$fileSizeKB" \
                    --clients "$numClients" \
                    --ops "$NUM_OPS" \
                    --grpc-addr "$GRPC_ADDR" \
                    --amqp-url "$AMQP_URL" \
                    --output "$outputFile" \
                    --temp-dir "/tmp/benchmark"

                echo ""
                echo "✅ Experimento concluído: $outputFile"
                echo ""
                
                # Pequeno delay entre experimentos
                sleep 2
            done
        done
    done
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Todos os Experimentos Concluídos!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Resultados salvos em: $RESULTS_DIR"
echo ""
echo "Para consolidar os resultados:"
echo "  cat $RESULTS_DIR/*.csv > $RESULTS_DIR/all_results.csv"

