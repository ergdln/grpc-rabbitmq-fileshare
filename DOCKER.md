# Docker Compose - File Sharing System

Este documento descreve como usar o sistema de compartilhamento de arquivos com Docker Compose.

## Pré-requisitos

- Docker
- Docker Compose

## Configuração Inicial

1. Copie o arquivo de exemplo de variáveis de ambiente:
```bash
cp env.example .env
```

2. Edite o arquivo `.env` conforme necessário:
```bash
# RabbitMQ Configuration
RABBITMQ_USER=guest
RABBITMQ_PASS=guest
RABBITMQ_PORT=5672
RABBITMQ_MANAGEMENT_PORT=15672

# AMQP URL
AMQP_URL=amqp://guest:guest@rabbitmq:5672/

# gRPC Server Configuration
GRPC_SERVER_PORT=50051
GRPC_SERVER_ADDR=grpc-server:50051

# Storage Configuration
DATA_DIR=/data
```

## Serviços

### RabbitMQ
- **Porta AMQP**: 5672
- **Interface de Gerenciamento**: http://localhost:15672
- **Usuário padrão**: guest
- **Senha padrão**: guest

### Servidores
- **grpc-server**: Servidor gRPC na porta 50051
- **rabbit-server**: Servidor RabbitMQ que consome da fila `rpc-file-requests`

### Clientes (escaláveis)
- **grpc-client**: Cliente gRPC
- **rabbit-client**: Cliente RabbitMQ

## Comandos Básicos

### Iniciar todos os serviços
```bash
docker-compose up -d
```

### Parar todos os serviços
```bash
docker-compose down
```

### Ver logs
```bash
# Todos os serviços
docker-compose logs -f

# Serviço específico
docker-compose logs -f grpc-server
docker-compose logs -f rabbit-server
```

### Reconstruir imagens
```bash
docker-compose build
```

## Usando os Clientes

### Cliente gRPC

#### Listar arquivos
```bash
docker-compose run --rm grpc-client list
```

#### Upload de arquivo
```bash
docker-compose run --rm -v "$(pwd):/workspace" grpc-client upload /workspace/arquivo.txt
```

#### Download de arquivo
```bash
docker-compose run --rm -v "$(pwd):/workspace" grpc-client download arquivo.txt /workspace/copia.txt
```

### Cliente RabbitMQ

#### Listar arquivos
```bash
docker-compose run --rm rabbit-client list
```

#### Upload de arquivo
```bash
docker-compose run --rm -v "$(pwd):/workspace" rabbit-client upload /workspace/arquivo.txt
```

#### Download de arquivo
```bash
docker-compose run --rm -v "$(pwd):/workspace" rabbit-client download arquivo.txt /workspace/copia.txt
```

## Script Auxiliar

Use o script `scripts/docker-usage.sh` para facilitar o uso:

```bash
# Iniciar serviços
./scripts/docker-usage.sh start

# Listar arquivos via gRPC
./scripts/docker-usage.sh grpc-list

# Upload via gRPC
./scripts/docker-usage.sh grpc-upload arquivo.txt

# Download via RabbitMQ
./scripts/docker-usage.sh rabbit-download arquivo.txt copia.txt
```

## Volumes

- **file-storage**: Volume compartilhado entre os servidores para armazenar arquivos
- **rabbitmq-data**: Volume para dados persistentes do RabbitMQ

## Rede

Todos os serviços estão na rede `fileshare-network`, permitindo comunicação entre eles usando os nomes dos serviços como hostnames.

## Escalamento

Os clientes podem ser escalados conforme necessário. No entanto, como são executados via `docker-compose run`, o escalamento é feito executando múltiplas instâncias manualmente.

## Troubleshooting

### Verificar se os serviços estão rodando
```bash
docker-compose ps
```

### Verificar logs de erro
```bash
docker-compose logs --tail=50 grpc-server
docker-compose logs --tail=50 rabbit-server
```

### Acessar interface do RabbitMQ
Abra http://localhost:15672 no navegador e faça login com as credenciais do `.env`.

### Verificar conectividade
```bash
# Testar conexão gRPC
docker-compose exec grpc-server ping grpc-server

# Testar conexão RabbitMQ
docker-compose exec rabbit-server ping rabbitmq
```

