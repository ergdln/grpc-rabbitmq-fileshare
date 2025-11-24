# gRPC vs RabbitMQ - Sistema de Compartilhamento de Arquivos

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-Compose-blue.svg)](https://docs.docker.com/compose/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Sistema cliente-servidor de compartilhamento de arquivos implementado em Go, comparando o desempenho de duas abordagens de comunicação: **gRPC** e **RabbitMQ**. O projeto permite testes de performance controlados medindo RTT (Round-Trip Time) sob diferentes condições de carga.

## 📋 Índice

- [Visão Geral](#-visão-geral)
- [Características](#-características)
- [Estrutura do Projeto](#-estrutura-do-projeto)
- [Metodologia](#-metodologia)
- [Pré-requisitos](#-pré-requisitos)
- [Instalação](#-instalação)
- [Como Usar](#-como-usar)
- [Resultados](#-resultados)
- [Análise de Dados](#-análise-de-dados)
- [Arquitetura](#-arquitetura)
- [Contribuindo](#-contribuindo)

## 🎯 Visão Geral

Este projeto implementa um sistema de compartilhamento de arquivos que permite comparar o desempenho de duas tecnologias de comunicação:

- **gRPC**: Protocolo RPC baseado em HTTP/2 com serialização Protocol Buffers
- **RabbitMQ**: Message broker usando AMQP para comunicação assíncrona

Ambos os sistemas implementam as mesmas operações:
- `list`: Lista arquivos disponíveis
- `upload`: Faz upload de arquivo
- `download`: Faz download de arquivo

## ✨ Características

- ✅ **Containerização completa** com Docker Compose
- ✅ **Testes sistemáticos** variando parâmetros de forma controlada
- ✅ **Testes de concorrência mista** simulando cenários realistas
- ✅ **Análise estatística** com cálculo de percentis (p50, p95, p99)
- ✅ **Visualizações** com gráficos comparativos
- ✅ **Reprodutibilidade** com sequências determinísticas
- ✅ **Suporte a arquivos grandes** (até 50MB)

## 📁 Estrutura do Projeto

```
grpc-rabbitmq-fileshare/
├── grpc-server/              # Servidor gRPC
│   ├── proto/                # Definições Protocol Buffers
│   ├── main.go
│   └── server.go
├── grpc-client/              # Cliente gRPC
│   ├── main.go
│   └── client.go
├── rabbit-server/             # Servidor RabbitMQ
│   ├── main.go
│   └── server.go
├── rabbit-client/             # Cliente RabbitMQ
│   ├── main.go
│   └── client.go
├── common/                    # Código compartilhado
│   ├── fileservice.go         # Interface comum
│   ├── localstorage.go        # Implementação de armazenamento
│   └── messages.go           # Estruturas de mensagem
├── tests/                     # Testes de benchmark
│   ├── runner.go              # Runner para testes sistemáticos
│   ├── benchmark.go           # Lógica de benchmark
│   ├── mixed_runner.go        # Runner para concorrência mista
│   └── mixed_concurrency.go   # Lógica de concorrência mista
├── scripts/                   # Scripts auxiliares
│   ├── run_all_experiments.sh      # Executa todos os testes sistemáticos
│   ├── run_grpc_experiments.sh     # Executa apenas testes gRPC
│   ├── run_mixed_concurrency.sh    # Executa testes de concorrência mista
│   ├── analyze_results.ipynb      # Análise de resultados sistemáticos
│   └── analyze_mixed_concurrency.ipynb  # Análise de concorrência mista
├── docker/                    # Dockerfiles
│   ├── Dockerfile.grpc-server
│   ├── Dockerfile.grpc-client
│   ├── Dockerfile.rabbit-server
│   └── Dockerfile.rabbit-client
├── results/                   # Resultados dos testes
│   ├── benchmark_*.csv        # CSVs de testes sistemáticos
│   ├── mixed_concurrency_*.csv    # CSVs de concorrência mista
│   ├── results_summary.csv    # Resumo estatístico
│   └── plots/                # Gráficos gerados
├── docker-compose.yml         # Configuração Docker Compose
├── go.mod                     # Dependências Go
└── README.md                  # Este arquivo
```

## 🔬 Metodologia

### Testes Sistemáticos

Executa experimentos controlados variando parâmetros de forma isolada:

- **Sistemas**: gRPC, RabbitMQ
- **Operações**: list, upload, download
- **Tamanhos de arquivo**: 10KB, 1MB, 10MB
- **Níveis de concorrência**: 1, 10, 20 clientes
- **Operações por teste**: 10.000

**Total**: 42 experimentos (2 sistemas × 3 operações × 3 tamanhos × 3 níveis de concorrência)

### Testes de Concorrência Mista

Simula cenário realista com alta concorrência e carga mista:

- **Clientes simultâneos**: 100
- **Operações totais**: 150.000 (100 clientes × 1.500 ops/cliente)
- **Distribuição de operações**: list (30%), upload (35%), download (35%)
- **Tamanhos variados**: 10KB, 1MB, 10MB (aleatório)
- **Sequência determinística**: Mesma sequência para ambos os sistemas (garantindo comparação justa)

### Métricas Coletadas

- **RTT (Round-Trip Time)**: Tempo de ida e volta em milissegundos
- **Taxa de sucesso**: Percentual de operações bem-sucedidas
- **Throughput**: Operações por segundo
- **Estatísticas**: Média, desvio padrão, mínimo, máximo
- **Percentis**: p50 (mediana), p95, p99

### Warm-up de Conexões

Antes de iniciar as medições, executa operações de warm-up para:
- Estabelecer conexões
- Evitar incluir overhead de setup nas medições
- Garantir RTTs consistentes

## 📦 Pré-requisitos

- **Go**: 1.24 ou superior
- **Docker**: 20.10 ou superior
- **Docker Compose**: 2.0 ou superior
- **Python 3** (opcional, para análise): pandas, matplotlib, jupyter
- **Jupyter Notebook** (opcional, para análise interativa)

## 🚀 Instalação

### 1. Clone o repositório

```bash
git clone https://github.com/seu-usuario/grpc-rabbitmq-fileshare.git
cd grpc-rabbitmq-fileshare
```

### 2. Configure variáveis de ambiente (opcional)

```bash
cp env.example .env
# Edite .env conforme necessário
```

### 3. Inicie os servidores

```bash
docker-compose up -d
```

Isso iniciará:
- RabbitMQ (porta 5672, management 15672)
- Servidor gRPC (porta 50051)
- Servidor RabbitMQ

### 4. Verifique se os serviços estão rodando

```bash
docker-compose ps
```

## 💻 Como Usar

### Testes Sistemáticos

#### Executar todos os experimentos (gRPC + RabbitMQ)

```bash
./scripts/run_all_experiments.sh
```

#### Executar apenas experimentos gRPC

```bash
./scripts/run_grpc_experiments.sh
```

**Resultados**: Salvos em `results/benchmark_<system>_<operation>_<size>kb_<clients>clients.csv`

### Testes de Concorrência Mista

```bash
# Padrão: 100 clientes, 1500 ops/cliente = 150k operações
./scripts/run_mixed_concurrency.sh

# Customizar parâmetros
./scripts/run_mixed_concurrency.sh <clientes> <ops_por_cliente> <duração> <distribuição> <sistema>
```

**Exemplos**:
```bash
# 50 clientes, 200 ops cada
./scripts/run_mixed_concurrency.sh 50 200

# 100 clientes, duração de 60 segundos
./scripts/run_mixed_concurrency.sh 100 0 60s

# Distribuição customizada
./scripts/run_mixed_concurrency.sh 100 1500 "" "list:50,upload:25,download:25"
```

**Resultados**: 
- `results/mixed_concurrency_grpc_*.csv`
- `results/mixed_concurrency_rabbit_*.csv`
- `results/mixed_concurrency_summary_*.csv`

### Uso Manual dos Clientes

#### Cliente gRPC

```bash
# Listar arquivos
docker-compose run --rm grpc-client list

# Upload
docker-compose run --rm -v "$(pwd):/workspace" grpc-client upload /workspace/arquivo.txt

# Download
docker-compose run --rm -v "$(pwd):/workspace" grpc-client download arquivo.txt /workspace/copia.txt
```

#### Cliente RabbitMQ

```bash
# Listar arquivos
docker-compose run --rm rabbit-client list

# Upload
docker-compose run --rm -v "$(pwd):/workspace" rabbit-client upload /workspace/arquivo.txt

# Download
docker-compose run --rm -v "$(pwd):/workspace" rabbit-client download arquivo.txt /workspace/copia.txt
```

## 📊 Resultados

> **Nota**: Os resultados apresentados são exemplos baseados em execuções reais. Valores podem variar dependendo do hardware e condições do sistema.

### Testes Sistemáticos

Os testes sistemáticos geram gráficos comparativos mostrando:

- **RTT vs. Número de Clientes**: Como o RTT varia com a concorrência
- **RTT vs. Tamanho de Arquivo**: Impacto do tamanho no desempenho

**Exemplo de resultados** (valores típicos):

| Sistema | Operação | Tamanho | Clientes | RTT Médio | RTT p95 |
|---------|----------|---------|----------|-----------|---------|
| gRPC | list | 0 KB | 1 | 0.75 ms | 1.2 ms |
| gRPC | upload | 1 MB | 10 | 80.70 ms | 120.5 ms |
| gRPC | upload | 10 MB | 20 | 748.72 ms | 850.0 ms |
| RabbitMQ | upload | 1 MB | 10 | 144.93 ms | 200.3 ms |
| RabbitMQ | download | 10 MB | 20 | 1024.92 ms | 1200.0 ms |

### Insights Principais

- **gRPC** geralmente apresenta menor latência para operações pequenas
- **RabbitMQ** tem overhead adicional devido ao message broker
- Ambos escalam bem com concorrência, mas gRPC mantém latência mais baixa
- Para arquivos grandes (10MB), a diferença de performance é mais pronunciada

### Testes de Concorrência Mista

Os testes de concorrência mista geram análises detalhadas:

- **RTT médio por sistema**: Comparação geral de performance sob alta concorrência
- **RTT por operação**: Análise por tipo de operação (list, upload, download)
- **RTT por tamanho**: Impacto do tamanho de arquivo no desempenho
- **Distribuição de RTT**: Histogramas e box plots mostrando variabilidade
- **Taxa de sucesso**: Confiabilidade dos sistemas sob carga
- **Evolução temporal**: RTT ao longo do tempo (150k operações)

**Exemplo de resultados** (150k operações, 100 clientes):

| Sistema | RTT Médio | RTT p50 | RTT p95 | RTT p99 | Taxa Sucesso | Throughput |
|---------|-----------|---------|---------|---------|--------------|------------|
| gRPC | ~15 ms | ~12 ms | ~45 ms | ~120 ms | >99% | ~800 ops/s |
| RabbitMQ | ~25 ms | ~20 ms | ~60 ms | ~150 ms | >99% | ~600 ops/s |

### Gráficos Gerados

Os notebooks de análise geram os seguintes gráficos (salvos em `results/plots/`):

**Testes Sistemáticos** (`generate_plots.ipynb`):
- `rtt_vs_clients_list_0kb.png` - RTT de listagem vs. número de clientes
- `rtt_vs_clients_upload_10kb.png` - Upload 10KB vs. clientes
- `rtt_vs_clients_upload_1024kb.png` - Upload 1MB vs. clientes
- `rtt_vs_clients_upload_10240kb.png` - Upload 10MB vs. clientes
- `rtt_vs_clients_download_10kb.png` - Download 10KB vs. clientes
- `rtt_vs_clients_download_1024kb.png` - Download 1MB vs. clientes
- `rtt_vs_clients_download_10240kb.png` - Download 10MB vs. clientes
- `rtt_vs_file_size_*.png` - RTT vs. tamanho de arquivo

**Concorrência Mista** (`analyze_mixed_concurrency.ipynb`):
- `mixed_concurrency_rtt_by_system.png` - Comparação geral gRPC vs RabbitMQ
- `mixed_concurrency_rtt_by_operation.png` - RTT por tipo de operação
- `mixed_concurrency_rtt_by_file_size.png` - RTT por tamanho de arquivo
- `mixed_concurrency_rtt_distribution.png` - Distribuição de RTT (histograma)
- `mixed_concurrency_rtt_boxplot.png` - Box plot com percentis
- `mixed_concurrency_success_rate.png` - Taxa de sucesso por sistema
- `mixed_concurrency_rtt_over_time.png` - Evolução temporal do RTT

> 💡 **Dica**: Execute os notebooks Jupyter para gerar os gráficos interativamente e explorar os dados em detalhes.

## 📈 Análise de Dados

### Análise de Resultados Sistemáticos

```bash
# Opção 1: Notebook Jupyter (recomendado)
jupyter notebook scripts/analyze_results.ipynb

# Opção 2: Script Python (se disponível)
python3 scripts/analyze_results.py
```

**Saída**: 
- `results/results_summary.csv` - Estatísticas agregadas (média, stddev, min, max)
- Estatísticas calculadas por combinação de parâmetros

### Análise de Concorrência Mista

```bash
# Notebook Jupyter (recomendado)
jupyter notebook scripts/analyze_mixed_concurrency.ipynb
```

**Saída**:
- Estatísticas detalhadas por sistema, operação e tamanho
- 7 gráficos comparativos salvos em `results/plots/`
- Análise de percentis (p50, p95, p99)
- Cálculo de throughput e taxa de sucesso

### Geração de Gráficos

```bash
# Gráficos de testes sistemáticos
jupyter notebook scripts/generate_plots.ipynb

# Gráficos de concorrência mista (incluído no notebook de análise)
jupyter notebook scripts/analyze_mixed_concurrency.ipynb
```

**Gráficos gerados**:
- Comparação gRPC vs RabbitMQ
- Análise de escalabilidade (RTT vs. clientes)
- Impacto do tamanho de arquivo
- Distribuições e percentis

## 🏗️ Arquitetura

### Componentes

```
┌─────────────┐         ┌──────────────┐
│   Cliente   │────────▶│  Servidor    │
│   (gRPC)    │  gRPC   │   (gRPC)     │
└─────────────┘         └──────────────┘

┌─────────────┐         ┌──────────────┐         ┌──────────────┐
│   Cliente   │────────▶│   RabbitMQ   │────────▶│  Servidor    │
│  (RabbitMQ) │  AMQP   │   Broker     │  AMQP  │  (RabbitMQ)  │
└─────────────┘         └──────────────┘         └──────────────┘
```

### Fluxo de Dados

**gRPC**:
1. Cliente estabelece conexão HTTP/2
2. Requisição serializada em Protocol Buffers
3. Servidor processa e retorna resposta
4. Medição de RTT (ida e volta)

**RabbitMQ**:
1. Cliente publica mensagem na fila `rpc-file-requests`
2. Servidor consome mensagem
3. Servidor processa e publica resposta na fila de resposta
4. Cliente consome resposta
5. Medição de RTT (ida e volta)

### Armazenamento

- **Volume Docker**: `file-storage` montado em `/data`
- **Persistência**: Arquivos mantidos entre reinicializações
- **Acesso concorrente**: Protegido com mutex

## ⚙️ Configurações

### Parâmetros de Aplicação

| Parâmetro | Valor | Descrição |
|-----------|-------|-----------|
| **Tamanho máximo de mensagem gRPC** | 50 MB | Limite configurado no servidor e cliente |
| **Timeout de operação** | 30s | Timeout para upload/download |
| **Timeout de list** | 10s | Timeout para listagem |
| **Prefetch RabbitMQ** | 1 | Mensagens pré-buscar por consumidor |

### Variáveis de Ambiente

Consulte `env.example` para todas as variáveis configuráveis.

## 🔧 Desenvolvimento

### Compilar manualmente

```bash
# Servidor gRPC
cd grpc-server
go build -o grpc-server .

# Cliente gRPC
cd grpc-client
go build -o grpc-client .

# Runner de testes
cd tests
go build -o runner runner.go benchmark.go

# Mixed runner
cd tests
go build -o mixed_runner mixed_runner.go mixed_concurrency.go benchmark.go
```

### Executar localmente (sem Docker)

```bash
# Terminal 1: Servidor gRPC
cd grpc-server
go run main.go server.go

# Terminal 2: Servidor RabbitMQ (requer RabbitMQ rodando)
cd rabbit-server
go run main.go server.go

# Terminal 3: Cliente
cd grpc-client
go run main.go list
```

## 📝 Formato dos Dados

### CSV de Resultados Individuais

```csv
timestamp,system,operation,file_size_kb,clients,rtt_ms,success
2025-11-23T12:00:00.000Z,grpc,upload,1024,10,80.696,true
```

### CSV de Resumo

```csv
system,operation,file_size_kb,clients,mean_ms,stddev_ms,min_ms,max_ms
grpc,upload,1024,10,80.696,22.925,9.871,307.397
```

### CSV de Concorrência Mista

```csv
timestamp,system,operation,file_size_kb,client_id,rtt_ms,success
2025-11-23T12:00:00.000Z,grpc,upload,1024,45,80.696,true
```

## 🐛 Troubleshooting

### Servidor não responde

```bash
# Verificar logs
docker-compose logs grpc-server
docker-compose logs rabbit-server

# Verificar se está rodando
docker-compose ps

# Reiniciar serviços
docker-compose restart grpc-server rabbit-server
```

### Erro de timeout

- Verifique se os servidores estão acessíveis
- Aumente o timeout no código se necessário (padrão: 30s)
- Verifique recursos do sistema (CPU/memória)
- Verifique se há muitos clientes simultâneos sobrecarregando o sistema

### Arquivos grandes falhando (gRPC)

- Verifique se o limite de 50MB está configurado no servidor e cliente
- Reconstrua o Docker: `docker-compose build --no-cache grpc-server`
- Reinicie o container: `docker-compose up -d grpc-server`

### RabbitMQ não processa mensagens

- Verifique a interface de gerenciamento: http://localhost:15672 (guest/guest)
- Verifique logs: `docker-compose logs rabbit-server`
- Verifique se a fila `rpc-file-requests` está sendo consumida
- Verifique se o servidor RabbitMQ está rodando: `docker-compose ps rabbit-server`

### Erro "no such file or directory" ao executar testes

- Crie o diretório de resultados: `mkdir -p results`
- Verifique permissões do diretório
- Use caminhos absolutos ou relativos corretos

### Performance degradada

- Verifique recursos do sistema: `docker stats`
- Reduza número de clientes simultâneos
- Verifique se há outros processos consumindo recursos
- Considere aumentar limites de recursos no Docker Compose

## 📚 Documentação Adicional

- [DOCKER.md](DOCKER.md) - Guia completo de uso com Docker
- [PLAN.md](PLAN.md) - Plano detalhado do projeto
- [tests/README.md](tests/README.md) - Documentação dos testes
- [scripts/README_ANALYSIS.md](scripts/README_ANALYSIS.md) - Guia de análise

## 🤝 Contribuindo

Contribuições são bem-vindas! Por favor:

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## 👥 Autores

- Seu Nome - [@seu-usuario](https://github.com/seu-usuario)

## 🙏 Agradecimentos

- Google gRPC team
- RabbitMQ team
- Comunidade Go

---

**Nota**: Este projeto foi desenvolvido para fins de pesquisa e comparação de performance entre gRPC e RabbitMQ em cenários de compartilhamento de arquivos.
