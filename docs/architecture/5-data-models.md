# 5. Data Models

## 5.1 Entidades Principais

* **Workflow:** Define o nome e o JSON da estrutura de tarefas (`definition`).
* **Execution:** Instância de execução com `status` e `context_snapshot` (JSONB).
* **TaskLog:** Registos granulares de sucesso/falha/retry de cada passo.
