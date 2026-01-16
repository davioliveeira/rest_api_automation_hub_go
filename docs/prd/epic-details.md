# Epic Details 

Épico 1: Fundação e Core Engine 

* 
**História 1.1: Estrutura Monorepo e API de Saúde** 


* **AC1:** Estrutura seguindo padrões Go (cmd, internal, pkg).
* **AC2:** Servidor Gin responde 200 OK em `/health`.


* 
**História 1.2: Motor de Execução e Contexto** 


* **AC1:** Struct `ExecutionContext` implementada com métodos Thread-safe.
* **AC2:** Loop sequencial que percorre a lista de tarefas JSON.


* 
**História 1.3: Contrato de TaskResult e Registry** 


* **AC1:** Interface `TaskExecutor` define o retorno `TaskResult` (Status, Output, Error).
* **AC2:** Sistema de registo dinâmico de nós por tipo.



Épico 2: Nós de Execução Essenciais (MVP) 

* 
**História 2.1: Nó HTTP Request** 


* **AC1:** Suporte a métodos HTTP, headers dinâmicos e captura de resposta para o contexto.


* 
**História 2.2: Nó Transform (Go Templates)** 


* **AC1:** Interpolação de dados usando `text/template` nativo do Go.


* 
**História 2.3: Nó HTML Parser (Goquery)** 


* **AC1:** Extração via seletores CSS devolvendo `[]map[string]any`.



Épico 3: Persistência e Ciclo de Vida 

* 
**História 3.1: Persistência de Workflows (Postgres)** 


* **AC1:** Endpoints para salvar/carregar definições de workflow.


* 
**História 3.2: Logs de Execução e Auditoria** 


* **AC1:** Registo persistente de inputs, outputs e status de cada tarefa.



Épico 4: Validação de MVP - Scraper Airbnb 

* 
**História 4.1: Execução do Workflow Airbnb** 


* **AC1:** Sucesso na execução: Fetch -> Parse -> DB via trigger de API.


