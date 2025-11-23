#!/bin/bash

# Script para executar todos os experimentos de benchmark apenas para gRPC

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TESTS_DIR="$PROJECT_ROOT/tests"
RESULTS_DIR="$PROJECT_ROOT/results"

# Cria diretório de resultados
mkdir -p "$RESULTS_DIR"

cd "$PROJECT_ROOT"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Executando Experimentos de Benchmark - gRPC"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Parâmetros
NUM_OPS=10000
CLIENTS=(1 10 20)
FILE_SIZES_KB=(10 1024 10240)  # 10KB, 1MB, 10MB
OPERATIONS=("list" "upload" "download")
SYSTEM="grpc"  # Apenas gRPC

# Configurações padrão
GRPC_ADDR="${GRPC_ADDR:-localhost:50051}"

# Compila o runner se necessário
if [ ! -f "$TESTS_DIR/runner" ]; then
    echo "🔨 Compilando runner..."
    cd "$TESTS_DIR"
    go build -o runner runner.go benchmark.go
    echo "✅ Runner compilado"
    echo ""
fi

cd "$PROJECT_ROOT"

# Contador de experimentos
TOTAL_EXPERIMENTS=0
COMPLETED_EXPERIMENTS=0

# Calcula total de experimentos
for operation in "${OPERATIONS[@]}"; do
    if [ "$operation" == "list" ]; then
        # list tem apenas 1 tamanho (0)
        for numClients in "${CLIENTS[@]}"; do
            TOTAL_EXPERIMENTS=$((TOTAL_EXPERIMENTS + 1))
        done
    else
        # upload e download têm 3 tamanhos
        for fileSizeKB in "${FILE_SIZES_KB[@]}"; do
            for numClients in "${CLIENTS[@]}"; do
                TOTAL_EXPERIMENTS=$((TOTAL_EXPERIMENTS + 1))
            done
        done
    fi
done

echo "📊 Total de experimentos: $TOTAL_EXPERIMENTS"
echo ""

# Executa experimentos
for operation in "${OPERATIONS[@]}"; do
    if [ "$operation" == "list" ]; then
        # list não precisa de tamanho de arquivo
        fileSizeKB=0
        for numClients in "${CLIENTS[@]}"; do
            COMPLETED_EXPERIMENTS=$((COMPLETED_EXPERIMENTS + 1))
            outputFile="$RESULTS_DIR/benchmark_${SYSTEM}_${operation}_${fileSizeKB}kb_${numClients}clients.csv"
            
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo "  [$COMPLETED_EXPERIMENTS/$TOTAL_EXPERIMENTS] Experimento gRPC"
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo "  Sistema: $SYSTEM"
            echo "  Operação: $operation"
            echo "  Tamanho: ${fileSizeKB} KB"
            echo "  Clientes: $numClients"
            echo "  Operações: $NUM_OPS"
            echo "  Saída: $(basename "$outputFile")"
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo ""

            START_TIME=$(date +%s)

            "$TESTS_DIR/runner" \
                --system "$SYSTEM" \
                --operation "$operation" \
                --file-size-kb "$fileSizeKB" \
                --clients "$numClients" \
                --ops "$NUM_OPS" \
                --grpc-addr "$GRPC_ADDR" \
                --output "$outputFile" \
                --temp-dir "/tmp/benchmark"

            END_TIME=$(date +%s)
            DURATION=$((END_TIME - START_TIME))
            MINUTES=$((DURATION / 60))
            SECONDS=$((DURATION % 60))

            echo ""
            echo "✅ Experimento concluído em ${MINUTES}m ${SECONDS}s: $(basename "$outputFile")"
            echo ""
            
            # Pequeno delay entre experimentos
            sleep 2
        done
    else
        # upload e download precisam de tamanhos de arquivo
        for fileSizeKB in "${FILE_SIZES_KB[@]}"; do
            for numClients in "${CLIENTS[@]}"; do
                COMPLETED_EXPERIMENTS=$((COMPLETED_EXPERIMENTS + 1))
                outputFile="$RESULTS_DIR/benchmark_${SYSTEM}_${operation}_${fileSizeKB}kb_${numClients}clients.csv"
                
                echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                echo "  [$COMPLETED_EXPERIMENTS/$TOTAL_EXPERIMENTS] Experimento gRPC"
                echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                echo "  Sistema: $SYSTEM"
                echo "  Operação: $operation"
                echo "  Tamanho: ${fileSizeKB} KB"
                echo "  Clientes: $numClients"
                echo "  Operações: $NUM_OPS"
                echo "  Saída: $(basename "$outputFile")"
                echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                echo ""

                START_TIME=$(date +%s)

                "$TESTS_DIR/runner" \
                    --system "$SYSTEM" \
                    --operation "$operation" \
                    --file-size-kb "$fileSizeKB" \
                    --clients "$numClients" \
                    --ops "$NUM_OPS" \
                    --grpc-addr "$GRPC_ADDR" \
                    --output "$outputFile" \
                    --temp-dir "/tmp/benchmark"

                END_TIME=$(date +%s)
                DURATION=$((END_TIME - START_TIME))
                MINUTES=$((DURATION / 60))
                SECONDS=$((DURATION % 60))

                echo ""
                echo "✅ Experimento concluído em ${MINUTES}m ${SECONDS}s: $(basename "$outputFile")"
                echo ""
                
                # Pequeno delay entre experimentos
                sleep 2
            done
        done
    fi
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Todos os Experimentos gRPC Concluídos!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Resultados salvos em: $RESULTS_DIR"
echo ""
echo "📈 Estatísticas:"
echo "   Total de experimentos: $COMPLETED_EXPERIMENTS"
echo ""
echo "Para analisar os resultados:"
echo "  1. Execute o notebook de análise:"
echo "     jupyter notebook scripts/analyze_results.ipynb"
echo ""
echo "  2. Ou use o script Python:"
echo "     python3 scripts/analyze_results.py"
echo ""

