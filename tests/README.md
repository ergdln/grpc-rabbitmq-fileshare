# Sistema de Benchmark - File Sharing

Este diretório contém o sistema de benchmark para medir o desempenho dos sistemas gRPC e RabbitMQ.

## Estrutura

- `runner.go` - Executor principal que aceita parâmetros via CLI
- `benchmark.go` - Funções de benchmark e gravação de resultados

## Compilação

```bash
cd tests
go build -o runner runner.go benchmark.go
```

## Uso do Runner

### Parâmetros

- `--system`: Sistema a testar (`grpc`, `rabbit`, ou `all`)
- `--operation`: Operação a executar (`list`, `upload`, `download`, ou `all`)
- `--file-size-kb`: Tamanho do arquivo em KB (10, 1024, 10240)
- `--clients`: Número de clientes concorrentes (1, 10, 20)
- `--ops`: Número total de operações (padrão: 10000)
- `--grpc-addr`: Endereço do servidor gRPC (padrão: localhost:50051)
- `--amqp-url`: URL do RabbitMQ (padrão: amqp://guest:guest@localhost:5672/)
- `--output`: Arquivo CSV de saída (padrão: benchmark_results.csv)
- `--temp-dir`: Diretório temporário para arquivos de teste (padrão: /tmp/benchmark)

### Exemplos

```bash
# Testar list com gRPC, 1 cliente, 10000 operações
./runner --system grpc --operation list --clients 1 --ops 10000

# Testar upload com RabbitMQ, 10 clientes, arquivo de 1MB
./runner --system rabbit --operation upload --file-size-kb 1024 --clients 10 --ops 10000

# Testar download com gRPC, 20 clientes, arquivo de 10MB
./runner --system grpc --operation download --file-size-kb 10240 --clients 20 --ops 10000

# Testar todos os sistemas e operações
./runner --system all --operation all --clients 10 --ops 10000
```

## Scripts de Execução

### run_all_experiments.sh

Executa todos os experimentos com todas as combinações de parâmetros:

- Sistemas: grpc, rabbit
- Operações: list, upload, download
- Tamanhos de arquivo: 10KB, 1MB, 10MB
- Número de clientes: 1, 10, 20
- Operações por experimento: 10000

```bash
./scripts/run_all_experiments.sh
```

Os resultados serão salvos em `results/benchmark_<system>_<operation>_<size>kb_<clients>clients.csv`

### run_random_concurrency.sh

Executa experimentos aleatórios com concorrência variável:

```bash
# Executa 50 experimentos aleatórios
./scripts/run_random_concurrency.sh 50
```

Os resultados serão salvos em `results/random_concurrency_<timestamp>.csv`

## Formato do CSV

O arquivo CSV gerado tem o seguinte formato:

```csv
timestamp,system,operation,file_size_kb,clients,rtt_ms,success
2024-01-01T12:00:00.000Z,grpc,list,0,1,5.234,true
2024-01-01T12:00:01.000Z,rabbit,upload,1024,10,125.456,true
```

### Campos

- `timestamp`: Timestamp da operação (RFC3339Nano)
- `system`: Sistema usado (`grpc` ou `rabbit`)
- `operation`: Operação executada (`list`, `upload`, `download`)
- `file_size_kb`: Tamanho do arquivo em KB (0 para `list`)
- `clients`: Número de clientes concorrentes
- `rtt_ms`: Round-trip time em milissegundos
- `success`: Se a operação foi bem-sucedida (`true` ou `false`)

## Análise dos Resultados

### Consolidar todos os resultados

```bash
cat results/*.csv | grep -v "^timestamp" | sort > results/all_results.csv
```

### Estatísticas básicas

```bash
# Média de RTT por sistema e operação
awk -F',' 'NR>1 {sum[$2","$3]+=$6; count[$2","$3]++} END {for (i in sum) print i, sum[i]/count[i]}' results/all_results.csv
```

### Filtrar por parâmetros

```bash
# Apenas operações bem-sucedidas
awk -F',' '$7=="true"' results/all_results.csv

# Apenas gRPC
awk -F',' '$2=="grpc"' results/all_results.csv

# Apenas upload de 1MB
awk -F',' '$3=="upload" && $4=="1024"' results/all_results.csv
```

## Requisitos

- Servidores gRPC e RabbitMQ rodando
- Go 1.21 ou superior
- Acesso aos servidores configurados

## Notas

- Para operações de `download`, certifique-se de que os arquivos já existam no servidor (faça upload primeiro)
- O runner gera arquivos de teste automaticamente para operações de upload
- Os arquivos temporários são criados em `/tmp/benchmark` por padrão
- Cada operação é medida individualmente (RTT por operação)

