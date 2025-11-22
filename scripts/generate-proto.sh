#!/bin/bash

# Script para gerar arquivos Go a partir dos arquivos .proto

set -e

PROTO_DIR="grpc-server/proto"
PROTO_OUT="grpc-server/proto"

echo "Verificando dependências..."

# Verifica se protoc está instalado
if ! command -v protoc &> /dev/null; then
    echo "Erro: protoc não está instalado."
    echo "Instale com: brew install protobuf (macOS) ou apt-get install protobuf-compiler (Linux)"
    exit 1
fi

# Verifica se protoc-gen-go está instalado
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Erro: protoc-gen-go não está instalado."
    echo "Instale com: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
fi

# Verifica se protoc-gen-go-grpc está instalado
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Erro: protoc-gen-go-grpc não está instalado."
    echo "Instale com: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
fi

echo "Gerando arquivos Go a partir dos arquivos .proto..."

protoc \
    --go_out="$PROTO_OUT" \
    --go_opt=paths=source_relative \
    --go-grpc_out="$PROTO_OUT" \
    --go-grpc_opt=paths=source_relative \
    "$PROTO_DIR"/*.proto

echo "Arquivos gerados com sucesso em $PROTO_OUT!"

