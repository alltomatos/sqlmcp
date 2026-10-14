# ADR 0003: Aceleração Analítica via DuckDB + Parquet para Grandes Volumes

- **Status:** Aceito
- **Data:** 2026-10-14
- **Autor:** Arquiteto de Software e Soluções

## Contexto
Quando um agente de IA se depara com tabelas de alta volumetria (milhões ou bilhões de registros) em bancos relacionais (Postgres, MySQL, SQL Server), queries analíticas pesadas (GROUP BY em múltiplas colunas, janelas temporais extensas, agregações e joins complexos) causam:
1. Saturação de CPU e I/O no banco transacional (OLTP).
2. Timeouts frequentes do agente (> 15s-30s).
3. Ineficiência na exploração de hipóteses analíticas pelo agente de IA.

## Decisão
Introduzir no MCP uma **Camada Aceleradora Analítica (DuckDB + Parquet)** embarcada:
1. **Ferramenta de Amostragem/Extração (`stage_to_parquet`):**
   - O agente pode exportar uma fatia filtrada, particionada ou amostragem representativa do banco relacional diretamente para um arquivo Parquet compacto local em disco temporário/cache.
2. **Ferramenta de Análise Vetorizada Local (`query_duckdb`):**
   - O agente executa SQL analítico completo sobre o arquivo Parquet usando a engine in-process DuckDB.
   - DuckDB utiliza execução colunar vetorizada, processando milhões de linhas por segundo com uso eficiente de memória e sem gerar concorrência nem lock no banco de produção.
3. **Ciclo de Vida do Cache Local:**
   - Arquivos Parquet temporários vinculados a um `stage_id` ou hash de consulta com TTL automático de expiração e limpeza segura.

## Consequências
### Positivas
- Desonera completamente o banco transacional de queries pesadas e repetitivas.
- Oferece velocidade analítica em ordens de grandeza superior para o agente.
- O formato Parquet reduz drasticamente o tráfego de dados e o uso de disco local.

### Negativas / Mitigações
- Dependência do runtime DuckDB CGO no binário Go.
- Requer espaço em disco local temporário (mitigado com limites rígidos de cota de tamanho, ex.: máximo 500 MB por exportação).
