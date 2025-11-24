#!/bin/bash

# Script para executar testes de concorrência mista com operações aleatórias

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TESTS_DIR="$PROJECT_ROOT/tests"
RESULTS_DIR="$PROJECT_ROOT/results"

# Cria diretório de resultados
mkdir -p "$RESULTS_DIR"

cd "$PROJECT_ROOT"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Teste de Concorrência Mista - Operações Aleatórias"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Parâmetros padrão
NUM_CLIENTS="${1:-100}"
OPS_PER_CLIENT="${2:-1500}"  # 100 clientes × 1500 ops = 150k operações
DURATION="${3:-}"
DISTRIBUTION="${4:-list:30,upload:35,download:35}"
SYSTEM="${5:-all}"

# Configurações padrão
GRPC_ADDR="${GRPC_ADDR:-localhost:50051}"
AMQP_URL="${AMQP_URL:-amqp://guest:guest@localhost:5672/}"

# Compila o mixed_runner se necessário
if [ ! -f "$TESTS_DIR/mixed_runner" ]; then
    echo "🔨 Compilando mixed_runner..."
    cd "$TESTS_DIR"
    go build -o mixed_runner mixed_runner.go mixed_concurrency.go benchmark.go
    echo "✅ mixed_runner compilado"
    echo ""
fi

cd "$PROJECT_ROOT"

# Constrói comando
CMD="$TESTS_DIR/mixed_runner"
CMD="$CMD --system $SYSTEM"
CMD="$CMD --clients $NUM_CLIENTS"
CMD="$CMD --ops-per-client $OPS_PER_CLIENT"
CMD="$CMD --distribution $DISTRIBUTION"
CMD="$CMD --grpc-addr $GRPC_ADDR"
CMD="$CMD --amqp-url $AMQP_URL"
CMD="$CMD --output-dir $RESULTS_DIR"

if [ -n "$DURATION" ]; then
    CMD="$CMD --duration $DURATION"
fi

echo "📋 Parâmetros:"
echo "   Sistema: $SYSTEM"
echo "   Clientes: $NUM_CLIENTS"
echo "   Operações por cliente: $OPS_PER_CLIENT"
if [ -n "$DURATION" ]; then
    echo "   Duração máxima: $DURATION"
fi
echo "   Distribuição: $DISTRIBUTION"
echo ""

# Executa
eval $CMD

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅ Teste concluído!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Resultados salvos em: $RESULTS_DIR"
echo ""
echo "Para analisar os resultados:"
echo "  python3 scripts/analyze_results.py"
echo "  ou"
echo "  jupyter notebook scripts/analyze_results.ipynb"
echo ""

