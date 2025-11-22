.PHONY: proto generate clean

# Diretório onde estão os arquivos .proto
PROTO_DIR := grpc-server/proto
# Diretório de saída para os arquivos gerados
PROTO_OUT := grpc-server/proto

# Comando protoc
PROTOC := protoc
PROTOC_GEN_GO := protoc-gen-go
PROTOC_GEN_GO_GRPC := protoc-gen-go-grpc

# Verifica se os plugins necessários estão instalados
check-protoc:
	@which $(PROTOC) > /dev/null || (echo "Erro: protoc não está instalado. Instale com: brew install protobuf" && exit 1)
	@which $(PROTOC_GEN_GO) > /dev/null || (echo "Erro: protoc-gen-go não está instalado. Instale com: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest" && exit 1)
	@which $(PROTOC_GEN_GO_GRPC) > /dev/null || (echo "Erro: protoc-gen-go-grpc não está instalado. Instale com: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest" && exit 1)

# Gera os arquivos Go a partir dos arquivos .proto
proto: check-protoc
	@echo "Gerando arquivos Go a partir dos arquivos .proto..."
	$(PROTOC) \
		--go_out=$(PROTO_OUT) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto
	@echo "Arquivos gerados com sucesso!"

# Alias para generate
generate: proto

# Limpa os arquivos gerados
clean:
	@echo "Limpando arquivos gerados..."
	@find $(PROTO_OUT) -name "*.pb.go" -delete
	@find $(PROTO_OUT) -name "*_grpc.pb.go" -delete
	@echo "Limpeza concluída!"

# Instala as dependências necessárias
install-deps:
	@echo "Instalando dependências do protobuf..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Dependências instaladas!"

