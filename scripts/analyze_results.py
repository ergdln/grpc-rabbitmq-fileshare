#!/usr/bin/env python3
"""
Script para analisar resultados de benchmark e gerar estatísticas resumidas.
Lê todos os CSVs em results/ e gera um CSV consolidado com estatísticas.
"""

import csv
import os
import statistics
from pathlib import Path
from collections import defaultdict

def read_csv_files(results_dir):
    """Lê todos os arquivos CSV de resultados."""
    all_results = []
    results_path = Path(results_dir)
    
    if not results_path.exists():
        print(f"❌ Diretório {results_dir} não encontrado!")
        return []
    
    csv_files = list(results_path.glob("*.csv"))
    
    if not csv_files:
        print(f"⚠️  Nenhum arquivo CSV encontrado em {results_dir}")
        return []
    
    print(f"📊 Encontrados {len(csv_files)} arquivo(s) CSV")
    
    for csv_file in csv_files:
        try:
            with open(csv_file, 'r', encoding='utf-8') as f:
                reader = csv.DictReader(f)
                for row in reader:
                    # Ignora linhas vazias ou inválidas
                    if not row.get('timestamp') or not row.get('success'):
                        continue
                    
                    # Apenas operações bem-sucedidas
                    if row.get('success', '').lower() != 'true':
                        continue
                    
                    all_results.append({
                        'system': row.get('system', ''),
                        'operation': row.get('operation', ''),
                        'file_size_kb': int(row.get('file_size_kb', 0)),
                        'clients': int(row.get('clients', 0)),
                        'rtt_ms': float(row.get('rtt_ms', 0))
                    })
        except Exception as e:
            print(f"⚠️  Erro ao ler {csv_file.name}: {e}")
            continue
    
    print(f"✅ Total de {len(all_results)} resultados válidos processados")
    return all_results

def calculate_statistics(results):
    """Calcula estatísticas agrupadas por sistema, operação, tamanho e clientes."""
    # Agrupa resultados por chave única
    grouped = defaultdict(list)
    
    for result in results:
        key = (
            result['system'],
            result['operation'],
            result['file_size_kb'],
            result['clients']
        )
        grouped[key].append(result['rtt_ms'])
    
    # Calcula estatísticas para cada grupo
    summary = []
    
    for (system, operation, file_size_kb, clients), rtt_values in grouped.items():
        if len(rtt_values) < 2:
            # Precisa de pelo menos 2 valores para desvio padrão
            stddev = 0.0
        else:
            stddev = statistics.stdev(rtt_values)
        
        summary.append({
            'system': system,
            'operation': operation,
            'file_size_kb': file_size_kb,
            'clients': clients,
            'mean_ms': statistics.mean(rtt_values),
            'stddev_ms': stddev,
            'min_ms': min(rtt_values),
            'max_ms': max(rtt_values),
            'count': len(rtt_values)
        })
    
    # Ordena por sistema, operação, tamanho, clientes
    summary.sort(key=lambda x: (x['system'], x['operation'], x['file_size_kb'], x['clients']))
    
    return summary

def write_summary_csv(summary, output_file):
    """Escreve o CSV de resumo."""
    fieldnames = ['system', 'operation', 'file_size_kb', 'clients', 
                  'mean_ms', 'stddev_ms', 'min_ms', 'max_ms']
    
    with open(output_file, 'w', newline='', encoding='utf-8') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        
        for row in summary:
            writer.writerow({
                'system': row['system'],
                'operation': row['operation'],
                'file_size_kb': row['file_size_kb'],
                'clients': row['clients'],
                'mean_ms': f"{row['mean_ms']:.3f}",
                'stddev_ms': f"{row['stddev_ms']:.3f}",
                'min_ms': f"{row['min_ms']:.3f}",
                'max_ms': f"{row['max_ms']:.3f}"
            })
    
    print(f"✅ Resumo salvo em: {output_file}")
    print(f"📊 Total de {len(summary)} combinações únicas")

def main():
    """Função principal."""
    script_dir = Path(__file__).parent
    project_root = script_dir.parent
    results_dir = project_root / 'results'
    output_file = project_root / 'results' / 'results_summary.csv'
    
    # Garante que o diretório de resultados existe
    results_dir.mkdir(exist_ok=True)
    
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print("  Análise de Resultados de Benchmark")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print()
    
    # Lê todos os CSVs
    results = read_csv_files(results_dir)
    
    if not results:
        print("❌ Nenhum resultado encontrado para processar")
        return 1
    
    # Calcula estatísticas
    print("\n📈 Calculando estatísticas...")
    summary = calculate_statistics(results)
    
    # Escreve CSV de resumo
    print(f"\n💾 Gerando arquivo de resumo...")
    write_summary_csv(summary, output_file)
    
    # Mostra estatísticas gerais
    print("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print("  Estatísticas Gerais")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    
    systems = set(s['system'] for s in summary)
    operations = set(s['operation'] for s in summary)
    
    print(f"Sistemas: {', '.join(sorted(systems))}")
    print(f"Operações: {', '.join(sorted(operations))}")
    print(f"Total de combinações: {len(summary)}")
    
    return 0

if __name__ == '__main__':
    exit(main())

