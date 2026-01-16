# 3. High Level Architecture

## 3.1 Technical Summary

O sistema utiliza um **Monólito Modular** com o padrão **Registry** para descoberta de executores. O motor é agnóstico ao tipo de tarefa, operando sobre uma interface comum `TaskExecutor` e um mapa de memória `ExecutionContext`.

## 3.2 High Level Project Diagram

```mermaid
graph TD
    Client[Client/Webhook] --> API[Gin REST API]
    API --> Service[Workflow Service]
    Service --> Engine[Execution Engine]
    Engine --> Registry[Task Registry]
    Registry --> T1[HTTP Task]
    Registry --> T2[HTML Parser Task]
    Registry --> T3[DB Task]
    Registry --> T4[Transform Task]
    T1 & T2 & T3 & T4 <--> Context[ExecutionContext - In Memory]
    Engine --> DB[(PostgreSQL - Persistence & Logs)]

```

## 3.3 Design Patterns

* **Registry Pattern:** Desacoplamento de tipos de tarefas do motor central.
* **Repository Pattern:** Abstração da camada de persistência (PostgreSQL).
* **Strategy Pattern:** Implementação polimórfica dos executores de nós.
