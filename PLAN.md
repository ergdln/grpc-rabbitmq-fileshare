# Plano do Projeto: gRPC vs RabbitMQ File Sharing System

## Visão Geral

Sistema cliente-servidor de compartilhamento de arquivos implementado em Go, comparando duas abordagens de comunicação:
- **gRPC**: Protocolo RPC baseado em HTTP/2
- **RabbitMQ**: Message broker usando AMQP

O projeto permite testes de performance controlados, medindo RTT (Round-Trip Time) variando:
- Número de clientes simultâneos (1, 10, 20)
- Tamanhos de arquivo (10KB, 1MB, 10MB)
- Operações remotas (list, upload, download)

## Estrutura do Projeto

```
grpc-rabbitmq-fileshare/
├── grpc-server/          # Servidor gRPC
├── grpc-client/          # Cliente gRPC
├── rabbit-server/         # Servidor RabbitMQ
├── rabbit-client/         # Cliente RabbitMQ
├── common/               # Código compartilhado
├── tests/                # Testes de benchmark
├── scripts/              # Scripts de teste e análise
├── docker/               # Dockerfiles
└── results/              # Resultados dos testes
```

## Funcionalidades Implementadas

### Servidores

#### gRPC Server (`grpc-server/`)
- ✅ Servidor gRPC na porta 50051
- ✅ Operações: ListFiles, UploadFile, DownloadFile
- ✅ **Limite de mensagem aumentado para 50MB** (correção aplicada)
- ✅ Armazenamento local de arquivos
- ✅ Logs detalhados de operações

#### RabbitMQ Server (`rabbit-server/`)
- ✅ Consome mensagens da fila `rpc-file-requests`
- ✅ Processa requisições RPC via RabbitMQ
- ✅ Operações: ListFiles, UploadFile, DownloadFile
- ✅ Armazenamento local de arquivos
- ✅ Base64 encoding para transferência de dados

### Clientes

#### gRPC Client (`grpc-client/`)
- ✅ CLI com argumentos: `list`, `upload <arquivo>`, `download <arquivo>`
- ✅ Conexão reutilizável
- ✅ **Limite de mensagem aumentado para 50MB** (correção aplicada)
- ✅ Timeout configurável (30s)

#### RabbitMQ Client (`rabbit-client/`)
- ✅ CLI com argumentos: `list`, `upload <arquivo>`, `download <arquivo>`
- ✅ Padrão RPC via RabbitMQ
- ✅ Timeout configurável (30s)

### Testes de Benchmark (`tests/`)

#### Runner (`runner.go`)
- ✅ Executa testes de performance
- ✅ Suporta múltiplos sistemas (gRPC, RabbitMQ)
- ✅ Múltiplas operações (list, upload, download)
- ✅ Variação de tamanhos de arquivo e clientes
- ✅ Gera CSVs com resultados detalhados
- ✅ **Pool de conexões reutilizáveis** (otimização)
- ✅ **Warm-up de conexões** (otimização)

#### Métricas
- ✅ RTT (Round-Trip Time) em milissegundos
- ✅ Taxa de sucesso/falha
- ✅ Timestamp de cada operação
- ✅ Suporte a 10.000+ operações por teste

## Correções e Melhorias Aplicadas

### 1. Limite de Tamanho de Mensagem gRPC (50MB)

**Problema Identificado:**
- gRPC tem limite padrão de 4MB por mensagem
- Arquivos de 10MB estavam falhando com `success=false`
- Erro: mensagem excedia o tamanho máximo permitido

**Solução Implementada:**
- ✅ **Servidor gRPC** (`grpc-server/server.go`):
  - `MaxRecvMsgSize(50MB)` e `MaxSendMsgSize(50MB)`
- ✅ **Cliente gRPC** (`grpc-client/client.go`):
  - `MaxCallRecvMsgSize(50MB)` e `MaxCallSendMsgSize(50MB)`
- ✅ **Testes** (`tests/benchmark.go`):
  - Mesmas configurações aplicadas

**Arquivos Modificados:**
- `grpc-server/server.go` (linha 113-116)
- `grpc-client/client.go` (linha 27-33)
- `tests/benchmark.go` (linha 122-128)

### 2. Otimizações de Performance

#### Pool de Conexões
- ✅ Conexões gRPC reutilizadas entre operações
- ✅ Conexões RabbitMQ reutilizadas
- ✅ Redução significativa de overhead de conexão

#### Warm-up de Conexões
- ✅ Operações de warm-up antes dos testes
- ✅ Evita incluir tempo de conexão inicial nos resultados
- ✅ Funções: `WarmUpGRPCConnection`, `WarmUpRabbitConnection`

### 3. Correção de Deadlock

**Problema:**
- Deadlock em `common/localstorage.go`
- `ensureDir()` tentava travar mutex já travado

**Solução:**
- ✅ Removido `ls.mu.Lock()` de `ensureDir()`
- ✅ Mutex já é gerenciado pelas funções chamadoras

## Scripts de Teste e Análise

### Scripts de Execução

#### `scripts/run_all_experiments.sh`
- Executa todos os experimentos (gRPC + RabbitMQ)
- 42 experimentos totais
- Combinações: 2 sistemas × 3 operações × 3 tamanhos × 3 clientes

#### `scripts/run_grpc_experiments.sh` ⭐ **NOVO**
- Executa apenas experimentos gRPC
- 21 experimentos totais:
  - 3 experimentos de `list` (1, 10, 20 clientes)
  - 9 experimentos de `upload` (3 tamanhos × 3 clientes)
  - 9 experimentos de `download` (3 tamanhos × 3 clientes)
- Mostra progresso (X/21)
- Exibe tempo de execução por experimento
- Compila runner automaticamente se necessário

**Uso:**
```bash
./scripts/run_grpc_experiments.sh
# Ou com endereço customizado
GRPC_ADDR=grpc-server:50051 ./scripts/run_grpc_experiments.sh
```

#### `scripts/run_random_concurrency.sh`
- Executa experimentos aleatórios
- Útil para testes exploratórios

#### Scripts de Teste Individual
- `scripts/test_grpc_list.sh`
- `scripts/test_grpc_upload.sh`
- `scripts/test_grpc_download.sh`

### Scripts de Análise

#### `scripts/analyze_results.py` / `scripts/analyze_results.ipynb` ⭐ **NOVO**
- Analisa todos os CSVs de resultados
- Calcula estatísticas:
  - Média (mean_ms)
  - Desvio padrão (stddev_ms)
  - Mínimo (min_ms)
  - Máximo (max_ms)
- Agrupa por: sistema, operação, tamanho, clientes
- Gera `results/results_summary.csv`

**Uso:**
```bash
# Script Python
python3 scripts/analyze_results.py

# Ou notebook Jupyter
jupyter notebook scripts/analyze_results.ipynb
```

#### `scripts/generate_plots.py` / `scripts/generate_plots.py` ⭐ **NOVO**
- Gera gráficos comparativos gRPC vs RabbitMQ
- Tipos de gráficos:
  1. **RTT vs Número de Clientes** (7 gráficos)
  2. **RTT vs Tamanho de Arquivo** (6 gráficos)
- Alta resolução (300 DPI)
- Cores: gRPC (azul), RabbitMQ (vermelho)
- Salva em `results/plots/`

**Uso:**
```bash
# Script Python
python3 scripts/generate_plots.py

# Ou notebook Jupyter
jupyter notebook scripts/generate_plots.ipynb
```

## Docker e Containerização

### Serviços Docker Compose

#### Serviços de Infraestrutura
- **rabbitmq**: Message broker (porta 5672, management 15672)

#### Serviços de Aplicação
- **grpc-server**: Servidor gRPC (porta 50051)
- **rabbit-server**: Servidor RabbitMQ
- **grpc-client**: Cliente gRPC (escalável)
- **rabbit-client**: Cliente RabbitMQ (escalável)

### Volumes
- `file-storage`: Armazenamento persistente de arquivos
- `rabbitmq-data`: Dados do RabbitMQ

### Networks
- `fileshare-network`: Rede isolada para comunicação entre serviços

## Fluxo de Trabalho Completo

### 1. Preparação do Ambiente

```bash
# Iniciar servidores
docker-compose up -d

# Verificar serviços
docker-compose ps
```

### 2. Executar Testes

```bash
# Todos os experimentos (gRPC + RabbitMQ)
./scripts/run_all_experiments.sh

# Apenas gRPC
./scripts/run_grpc_experiments.sh
```

### 3. Analisar Resultados

```bash
# Opção 1: Notebook Jupyter (recomendado)
jupyter notebook scripts/analyze_results.ipynb

# Opção 2: Script Python
python3 scripts/analyze_results.py
```

### 4. Gerar Gráficos

```bash
# Opção 1: Notebook Jupyter (recomendado)
jupyter notebook scripts/generate_plots.ipynb

# Opção 2: Script Python
python3 scripts/generate_plots.py
```

### 5. Visualizar Resultados

- CSVs individuais: `results/benchmark_*.csv`
- CSV consolidado: `results/results_summary.csv`
- Gráficos: `results/plots/*.png`

## Configurações Importantes

### Variáveis de Ambiente

```bash
# gRPC
GRPC_SERVER_ADDR=localhost:50051

# RabbitMQ
AMQP_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_USER=guest
RABBITMQ_PASS=guest

# Storage
DATA_DIR=/data
```

### Parâmetros de Teste

- **Número de operações**: 10.000 por padrão
- **Clientes**: 1, 10, 20
- **Tamanhos**: 10KB, 1MB, 10MB
- **Timeout**: 30 segundos

## Estrutura de Dados

### CSV de Resultados Individuais

```csv
timestamp,system,operation,file_size_kb,clients,rtt_ms,success
2025-11-23T12:00:00.000Z,grpc,upload,1024,10,80.696,true
```

### CSV de Resumo (`results_summary.csv`)

```csv
system,operation,file_size_kb,clients,mean_ms,stddev_ms,min_ms,max_ms
grpc,upload,1024,10,80.696,22.925,9.871,307.397
```

## Problemas Conhecidos e Soluções

### 1. ✅ Resolvido: Limite de Mensagem gRPC
- **Status**: Corrigido
- **Solução**: Limite aumentado para 50MB

### 2. ✅ Resolvido: Deadlock em LocalStorage
- **Status**: Corrigido
- **Solução**: Removido lock duplo em `ensureDir()`

### 3. ✅ Resolvido: Alto RTT Inicial
- **Status**: Otimizado
- **Solução**: Pool de conexões e warm-up

### 4. ⚠️ Conhecido: RabbitMQ Bottleneck
- **Status**: Identificado
- **Descrição**: Prefetch count baixo (1) pode limitar throughput
- **Nota**: Usuário reverteu mudanças propostas

## Próximos Passos Sugeridos

### Melhorias de Performance
- [ ] Implementar streaming para arquivos grandes (>50MB)
- [ ] Aumentar prefetch count do RabbitMQ para melhor throughput
- [ ] Implementar compressão de dados
- [ ] Adicionar métricas de throughput (MB/s)

### Funcionalidades
- [ ] Suporte a múltiplos arquivos simultâneos
- [ ] Progresso de upload/download
- [ ] Validação de checksum
- [ ] Suporte a metadados de arquivo

### Análise
- [ ] Análise estatística mais avançada (percentis, quartis)
- [ ] Gráficos de distribuição de RTT
- [ ] Análise de correlação entre variáveis
- [ ] Relatórios em PDF

## Dependências

### Go
- Versão: 1.24.0+
- Pacotes principais:
  - `google.golang.org/grpc` (v1.77.0)
  - `github.com/streadway/amqp`

### Python (para análise)
- pandas
- matplotlib
- jupyter

### Docker
- Docker Engine
- Docker Compose

## Referências

- [gRPC Documentation](https://grpc.io/docs/)
- [RabbitMQ Documentation](https://www.rabbitmq.com/documentation.html)
- [Go gRPC Package](https://pkg.go.dev/google.golang.org/grpc)

## Changelog

### 2025-11-23
- ✅ Aumentado limite de mensagem gRPC para 50MB
- ✅ Criado script `run_grpc_experiments.sh`
- ✅ Criados notebooks Jupyter para análise
- ✅ Otimizações de pool de conexões
- ✅ Correção de deadlock em LocalStorage

### 2025-11-22
- ✅ Implementação inicial do sistema
- ✅ Testes de benchmark funcionais
- ✅ Scripts de execução de experimentos
- ✅ Docker Compose configurado

---

**Última atualização**: 2025-11-23

