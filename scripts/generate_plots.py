#!/usr/bin/env python3
"""
Script para gerar gráficos comparativos de performance gRPC vs RabbitMQ.
Gera gráficos de RTT médio vs número de clientes e vs tamanho de arquivo.
"""

import csv
import matplotlib.pyplot as plt
import matplotlib.patches as mpatches
from pathlib import Path
import sys

def load_summary_data(summary_file):
    """Carrega dados do CSV de resumo."""
    data = []
    
    if not Path(summary_file).exists():
        print(f"❌ Arquivo {summary_file} não encontrado!")
        print("   Execute primeiro: python scripts/analyze_results.py")
        return None
    
    with open(summary_file, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            data.append({
                'system': row['system'],
                'operation': row['operation'],
                'file_size_kb': int(row['file_size_kb']),
                'clients': int(row['clients']),
                'mean_ms': float(row['mean_ms']),
                'stddev_ms': float(row['stddev_ms']),
                'min_ms': float(row['min_ms']),
                'max_ms': float(row['max_ms'])
            })
    
    return data

def plot_rtt_vs_clients(data, output_dir):
    """Gera gráfico de RTT médio vs número de clientes."""
    # Agrupa por sistema, operação e tamanho de arquivo
    plots_data = {}
    
    for row in data:
        key = (row['operation'], row['file_size_kb'])
        if key not in plots_data:
            plots_data[key] = {'grpc': {'clients': [], 'rtt': []}, 
                               'rabbit': {'clients': [], 'rtt': []}}
        
        system = row['system']
        if system in ['grpc', 'rabbit']:
            plots_data[key][system]['clients'].append(row['clients'])
            plots_data[key][system]['rtt'].append(row['mean_ms'])
    
    # Cria um gráfico para cada combinação de operação e tamanho
    for (operation, file_size_kb), systems_data in plots_data.items():
        fig, ax = plt.subplots(figsize=(10, 6))
        
        # Dados gRPC
        if systems_data['grpc']['clients']:
            grpc_clients, grpc_rtt = zip(*sorted(zip(systems_data['grpc']['clients'], 
                                                      systems_data['grpc']['rtt'])))
            ax.plot(grpc_clients, grpc_rtt, 'o-', label='gRPC', linewidth=2, 
                   markersize=8, color='#4285F4')
        
        # Dados RabbitMQ
        if systems_data['rabbit']['clients']:
            rabbit_clients, rabbit_rtt = zip(*sorted(zip(systems_data['rabbit']['clients'], 
                                                         systems_data['rabbit']['rtt'])))
            ax.plot(rabbit_clients, rabbit_rtt, 's-', label='RabbitMQ', linewidth=2, 
                   markersize=8, color='#EA4335')
        
        ax.set_xlabel('Número de Clientes', fontsize=12, fontweight='bold')
        ax.set_ylabel('RTT Médio (ms)', fontsize=12, fontweight='bold')
        
        size_label = f"{file_size_kb} KB" if file_size_kb > 0 else "N/A"
        title = f'RTT Médio vs Número de Clientes\n{operation.upper()} - {size_label}'
        ax.set_title(title, fontsize=14, fontweight='bold')
        
        ax.grid(True, alpha=0.3, linestyle='--')
        ax.legend(loc='best', fontsize=11)
        
        # Melhora a formatação dos eixos
        ax.set_xlim(left=0)
        ax.set_ylim(bottom=0)
        
        plt.tight_layout()
        
        # Salva o gráfico
        filename = f'rtt_vs_clients_{operation}_{file_size_kb}kb.png'
        output_path = Path(output_dir) / filename
        plt.savefig(output_path, dpi=300, bbox_inches='tight')
        print(f"✅ Gráfico salvo: {output_path}")
        plt.close()

def plot_rtt_vs_file_size(data, output_dir):
    """Gera gráfico de RTT médio vs tamanho de arquivo."""
    # Agrupa por sistema, operação e número de clientes
    plots_data = {}
    
    for row in data:
        # Ignora operação 'list' que não tem tamanho de arquivo
        if row['operation'] == 'list':
            continue
            
        key = (row['operation'], row['clients'])
        if key not in plots_data:
            plots_data[key] = {'grpc': {'size': [], 'rtt': []}, 
                              'rabbit': {'size': [], 'rtt': []}}
        
        system = row['system']
        if system in ['grpc', 'rabbit']:
            plots_data[key][system]['size'].append(row['file_size_kb'])
            plots_data[key][system]['rtt'].append(row['mean_ms'])
    
    # Cria um gráfico para cada combinação de operação e clientes
    for (operation, clients), systems_data in plots_data.items():
        fig, ax = plt.subplots(figsize=(10, 6))
        
        # Dados gRPC
        if systems_data['grpc']['size']:
            grpc_size, grpc_rtt = zip(*sorted(zip(systems_data['grpc']['size'], 
                                                   systems_data['grpc']['rtt'])))
            ax.plot(grpc_size, grpc_rtt, 'o-', label='gRPC', linewidth=2, 
                   markersize=8, color='#4285F4')
        
        # Dados RabbitMQ
        if systems_data['rabbit']['size']:
            rabbit_size, rabbit_rtt = zip(*sorted(zip(systems_data['rabbit']['size'], 
                                                       systems_data['rabbit']['rtt'])))
            ax.plot(rabbit_size, rabbit_rtt, 's-', label='RabbitMQ', linewidth=2, 
                   markersize=8, color='#EA4335')
        
        ax.set_xlabel('Tamanho do Arquivo (KB)', fontsize=12, fontweight='bold')
        ax.set_ylabel('RTT Médio (ms)', fontsize=12, fontweight='bold')
        
        title = f'RTT Médio vs Tamanho do Arquivo\n{operation.upper()} - {clients} Cliente(s)'
        ax.set_title(title, fontsize=14, fontweight='bold')
        
        ax.grid(True, alpha=0.3, linestyle='--')
        ax.legend(loc='best', fontsize=11)
        
        # Melhora a formatação dos eixos
        ax.set_xlim(left=0)
        ax.set_ylim(bottom=0)
        
        # Formata o eixo X para mostrar tamanhos de arquivo de forma legível
        ax.set_xscale('log', base=10)
        
        plt.tight_layout()
        
        # Salva o gráfico
        filename = f'rtt_vs_file_size_{operation}_{clients}clients.png'
        output_path = Path(output_dir) / filename
        plt.savefig(output_path, dpi=300, bbox_inches='tight')
        print(f"✅ Gráfico salvo: {output_path}")
        plt.close()

def main():
    """Função principal."""
    script_dir = Path(__file__).parent
    project_root = script_dir.parent
    summary_file = project_root / 'results' / 'results_summary.csv'
    output_dir = project_root / 'results' / 'plots'
    
    # Cria diretório de saída
    output_dir.mkdir(parents=True, exist_ok=True)
    
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print("  Geração de Gráficos Comparativos")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print()
    
    # Carrega dados
    print("📊 Carregando dados do resumo...")
    data = load_summary_data(summary_file)
    
    if not data:
        return 1
    
    print(f"✅ {len(data)} registros carregados")
    print()
    
    # Gera gráficos
    print("📈 Gerando gráficos...")
    print()
    
    print("  📊 RTT vs Número de Clientes...")
    plot_rtt_vs_clients(data, output_dir)
    print()
    
    print("  📊 RTT vs Tamanho de Arquivo...")
    plot_rtt_vs_file_size(data, output_dir)
    print()
    
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print(f"✅ Todos os gráficos salvos em: {output_dir}")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    
    return 0

if __name__ == '__main__':
    exit(main())

