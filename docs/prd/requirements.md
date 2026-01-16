# Requirements

## Functional Requirements (FR) 

* **FR1:** O sistema deve expor uma API REST (usando Gin) para receber definições de workflow via JSON.
* **FR2:** O motor deve processar tarefas de forma sequencial conforme a ordem definida no array do JSON.
* **FR3:** O sistema deve manter um `ExecutionContext` isolado por execução para partilha de dados entre nós.
* **FR4:** Implementar nó `http_request` com suporte a GET/POST, headers e interpolação de corpo.
* **FR5:** Implementar nó `html_parser` usando seletores CSS (via Goquery) para extrair dados estruturados.
* **FR6:** Implementar nó `database_query` para operações CRUD em PostgreSQL com parâmetros dinâmicos.
* **FR7:** Implementar nó `transform` usando a biblioteca nativa `text/template` para mapeamento de dados.
* **FR8:** O motor deve suportar uma política de *retry* configurável (tentativas e backoff) por tarefa.

## Non-Functional Requirements (NFR)

* **NFR1:** O motor deve ser escrito exclusivamente em Golang (versão 1.21+).
* **NFR2:** Cobertura de testes unitários mínima de 80% no core e nos nós.
* **NFR3:** Persistência do histórico de execuções e logs de erro em base de dados relacional.
* **NFR4:** Overhead de orquestração interno inferior a 50ms por transição de nó.
