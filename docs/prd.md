# GoAutomation Hub Product Requirements Document (PRD)

## Goals and Background Context

### Goals 

* **Proficiência Técnica:** Demonstrar domínio sénior da linguagem Go, tratamento de erros idiomático e padrões de design modular.
* **Qualidade de Código:** Garantir uma cobertura de testes unitários superior a 80% nos componentes core e nos nós de execução.
* **Validação de MVP:** Executar com sucesso o fluxo completo do "Monitor de Preços Airbnb" (Fetch -> Parse -> DB -> Notify).
* **Extensibilidade:** Permitir a adição de um novo tipo de nó em menos de 2 horas por um desenvolvedor familiarizado com o sistema.

### Background Context 

O mercado de automação é dominado por ferramentas *No-Code* visuais (n8n, Zapier) que abstraem a lógica de engenharia, ou por motores industriais (Temporal) com elevada complexidade. O **GoAutomation Hub** preenche esta lacuna com um motor *headless* e *API-first* focado em desenvolvedores que precisam de controlo total, performance e uma base de código sólida para portfólio.

### Change Log

| Data | Versão | Descrição | Autor |
| --- | --- | --- | --- |
| 15/01/2026 | 1.0 | Rascunho Inicial baseado no Project Brief | John (PM) |

## Requirements

### Functional Requirements (FR) 

* **FR1:** O sistema deve expor uma API REST (usando Gin) para receber definições de workflow via JSON.
* **FR2:** O motor deve processar tarefas de forma sequencial conforme a ordem definida no array do JSON.
* **FR3:** O sistema deve manter um `ExecutionContext` isolado por execução para partilha de dados entre nós.
* **FR4:** Implementar nó `http_request` com suporte a GET/POST, headers e interpolação de corpo.
* **FR5:** Implementar nó `html_parser` usando seletores CSS (via Goquery) para extrair dados estruturados.
* **FR6:** Implementar nó `database_query` para operações CRUD em PostgreSQL com parâmetros dinâmicos.
* **FR7:** Implementar nó `transform` usando a biblioteca nativa `text/template` para mapeamento de dados.
* **FR8:** O motor deve suportar uma política de *retry* configurável (tentativas e backoff) por tarefa.

### Non-Functional Requirements (NFR)

* **NFR1:** O motor deve ser escrito exclusivamente em Golang (versão 1.21+).
* **NFR2:** Cobertura de testes unitários mínima de 80% no core e nos nós.
* **NFR3:** Persistência do histórico de execuções e logs de erro em base de dados relacional.
* **NFR4:** Overhead de orquestração interno inferior a 50ms por transição de nó.

## Technical Assumptions 

* 
**Repository Structure:** Monorepo para simplificar a gestão de dependências e visibilidade do projeto.


* 
**Service Architecture:** Monolítico Modular, separando API, Engine e Task Executors.


* 
**Testing:** Foco em Testes Unitários e Testes de Integração com mocks para serviços externos.


*
**Stack Principal:** Golang (1.21+), Gin, GORM/pgx e Goquery.



## Epic List 

* 
**Épico 1: Fundação e Core Engine:** Estrutura base, API Gin e ciclo básico do executor com `ExecutionContext`.


* 
**Épico 2: Nós de Execução Essenciais (MVP):** Implementação dos nós `http_request`, `html_parser`, `database_query` e `transform`.


* 
**Épico 3: Persistência e Ciclo de Vida:** Integração com PostgreSQL para armazenamento de workflows e logs.


* 
**Épico 4: Validação de MVP - Scraper Airbnb:** Workflow real fim-a-fim para validação do sistema.



## Epic Details 

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



## Checklist Results Report

*(A ser preenchido após a execução da tarefa `pm-checklist`)*

## Next Steps 

1. Executar a validação de qualidade via `pm-checklist`.
2. Handoff para o **Architect (Winston)** para criação do documento de arquitetura técnica.



---

*Documento gerado seguindo o framework BMAD-METHOD™*

---
