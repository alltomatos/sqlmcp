# ADR 0002: Postura de Segurança Read-Only Estrita e Proteção de Contexto

- **Status:** Aceito
- **Data:** 2026-10-14
- **Autor:** Arquiteto de Software e Soluções

## Contexto
Agentes de IA operando sobre bancos de produção e analíticos representam riscos de:
1. Execução acidental de comandos destrutivos ou de mutação (`DELETE`, `DROP`, `UPDATE`, `TRUNCATE`, `ALTER`).
2. Extração inadvertida de volumes massivos de dados (`SELECT * FROM table_com_10M_linhas`), gerando estouro de memória, saturação de I/O no banco e consumo excessivo de janela de contexto de tokens no LLM.
3. SQL Injection ou encadeamento de múltiplos statements via injeção (`SELECT 1; DROP TABLE users;`).

## Decisão
Implementar uma política inegociável de **Read-Only Estrito por Padrão** diretamente na camada de aplicação do MCP, antes de qualquer envio ao driver do banco:
1. **Parser AST / Validação Léxica de Query:**
   - As queries submetidas à tool `execute_query` e `explain_query` são analisadas sintaticamente.
   - Apenas comandos que comecem e contenham exclusivamente operações de leitura (`SELECT`, `WITH ... SELECT`, `EXPLAIN`) são aceitos.
   - Encadeamento de múltiplos comandos (separados por ponto e vírgula) é bloqueado.
   - Comandos com palavras-chave de mutação ou execução de procedures opacas (`EXEC`, `sp_executesql`, `INSERT`, `UPDATE`, `DELETE`, `CREATE`, `DROP`, `ALTER`, `GRANT`, `REVOKE`, etc.) são sumariamente rejeitados com erro explicativo amigável para o agente.
2. **Defesa em Profundidade (Nível de Conexão):**
   - Transações abertas com flags explícitas de somente leitura (`READ ONLY`) nos bancos que suportam.
3. **Proteção de Janela de Contexto (Context Window Guardrails):**
   - **Max Rows Cap:** Limite padrão de 500 linhas por execução (parametrizável com teto máximo inegociável de 2.000 linhas).
   - **Timeout Estrito:** Cancelamento via `context.WithTimeout` padrão de 15 segundos por consulta.
   - **Truncamento Seguro de Colunas de Texto/Blob:** Colunas contendo textos muito longos (> 4 KB por célula) são truncadas com indicador `[truncated...]` para economizar tokens, a menos que solicitado explicitamente.

## Consequências
### Positivas
- Risco zero de modificação ou destruição de dados acidental por agentes autônomos.
- Proteção garantida contra saturação de contexto do LLM e contenção de custos de tokens.
- Comportamento previsível para times de DBA e segurança corporativa.

### Negativas / Mitigações
- Não atende cenários onde o agente precisa rodar migrações DDL ou gravar dados diretamente no banco fonte (para isso, a decisão é transferir a análise para a sandbox analítica DuckDB local, sem mutar o banco relacional).
