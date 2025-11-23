# Análise de Resultados de Benchmark

Este diretório contém notebooks Jupyter para analisar e visualizar os resultados dos testes de benchmark.

## Notebooks Disponíveis

### 1. `analyze_results.ipynb`

Analisa todos os CSVs de resultados e gera um resumo estatístico.

**Uso:**
```bash
jupyter notebook scripts/analyze_results.ipynb
# ou
jupyter lab scripts/analyze_results.ipynb
```

**Funcionalidades:**
- Lê todos os arquivos CSV em `results/`
- Calcula estatísticas (média, desvio padrão, min, max)
- Visualiza dados com pandas
- Gera `results/results_summary.csv`

**Campos do CSV de resumo:**
- `system`: Sistema usado (grpc ou rabbit)
- `operation`: Operação (list, upload, download)
- `file_size_kb`: Tamanho do arquivo em KB
- `clients`: Número de clientes concorrentes
- `mean_ms`: RTT médio em milissegundos
- `stddev_ms`: Desvio padrão em milissegundos
- `min_ms`: RTT mínimo em milissegundos
- `max_ms`: RTT máximo em milissegundos

### 2. `generate_plots.ipynb`

Gera gráficos comparativos de performance gRPC vs RabbitMQ.

**Uso:**
```bash
jupyter notebook scripts/generate_plots.ipynb
# ou
jupyter lab scripts/generate_plots.ipynb
```

**Requisitos:**
- Python 3
- pandas (`pip install pandas`)
- matplotlib (`pip install matplotlib`)
- Arquivo `results_summary.csv` gerado pelo notebook de análise

**Saída:**
- Gráficos salvos em `results/plots/`

**Tipos de gráficos gerados:**

1. **RTT vs Número de Clientes**
   - Um gráfico para cada combinação de operação e tamanho de arquivo
   - Compara gRPC vs RabbitMQ
   - Arquivo: `rtt_vs_clients_{operation}_{size}kb.png`

2. **RTT vs Tamanho de Arquivo**
   - Um gráfico para cada combinação de operação e número de clientes
   - Compara gRPC vs RabbitMQ
   - Arquivo: `rtt_vs_file_size_{operation}_{clients}clients.png`

## Fluxo de Trabalho Recomendado

1. Execute os testes de benchmark:
   ```bash
   ./scripts/run_all_experiments.sh
   ```

2. Abra e execute o notebook de análise:
   ```bash
   jupyter notebook scripts/analyze_results.ipynb
   ```
   - Execute todas as células (Cell → Run All)
   - Verifique o CSV gerado em `results/results_summary.csv`

3. Abra e execute o notebook de gráficos:
   ```bash
   jupyter notebook scripts/generate_plots.ipynb
   ```
   - Execute todas as células (Cell → Run All)
   - Visualize os gráficos gerados em `results/plots/`

4. Visualize os gráficos em `results/plots/`

## Exemplo de Saída

### CSV de Resumo
```csv
system,operation,file_size_kb,clients,mean_ms,stddev_ms,min_ms,max_ms
grpc,list,0,1,0.749,0.532,0.378,30.254
grpc,list,0,10,0.863,0.751,0.337,21.929
rabbit,list,0,1,7.234,2.145,3.456,45.123
...
```

### Gráficos
- 7 gráficos de RTT vs Clientes
- 6 gráficos de RTT vs Tamanho de Arquivo
- Todos em alta resolução (300 DPI)

## Instalação de Dependências

```bash
pip install pandas matplotlib jupyter
```

## Vantagens dos Notebooks

- **Interatividade**: Execute células individualmente
- **Visualização**: Veja os dados e gráficos diretamente no notebook
- **Exploração**: Fácil de modificar e experimentar
- **Documentação**: Código e explicações juntos

## Notas

- Os notebooks processam apenas operações bem-sucedidas (`success=true`)
- O notebook de análise agrupa resultados por sistema, operação, tamanho e clientes
- Os gráficos usam cores diferentes para gRPC (azul) e RabbitMQ (vermelho)
- Gráficos de tamanho de arquivo usam escala logarítmica no eixo X
- Os notebooks podem ser executados célula por célula para melhor controle

