# 📦 Inventory Movement Processing System

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=for-the-badge&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker)](https://www.docker.com/)

> **A high-throughput, concurrent warehouse inventory service designed to eliminate ghost inventory through strict data consistency and a reliable audit trail.**

---

## 📖 Introduction: The "Ghost Inventory" Problem
In fast-paced warehouse environments, barcode scanners continuously log transactions (IN, OUT, ADJUST) simultaneously. If two scanners write to the same item's stock balance without synchronization:
- **Lost Updates** occur, causing discrepancies between digital balances and physical stock.
- **Negative Stock Balances** happen when an OUT scan executes while another is pending, bypassing stock validation checks.
- **Duplicate Transactions** slip through if scanner retry-logs aren't verified.

This system is built specifically to address these issues, guaranteeing **serializable-like correctness at the application level** with maximum parallel performance.

---

## 🗺️ Table of Contents
- [✨ Key Features](#-key-features)
- [🏗️ Architectural Architecture](#%EF%B8%8F-architectural-architecture)
- [⚡ High-Throughput Concurrency Model](#-high-throughput-concurrency-model)
- [📁 Project Directory Structure](#-project-directory-structure)
- [🛡️ Security & Role-Based Authorization](#%EF%B8%8F-security--role-based-authorization)
- [🚀 Getting Started](#-getting-started)
- [📡 API Reference & cURL Examples](#-api-reference--curl-examples)
- [📝 Key Technical Assumptions](#-key-technical-assumptions)

---

## ✨ Key Features
- **High-Concurrency Worker Engine** — Groups incoming scanner rows by `item_id` so that mutations to the same item execute sequentially while separate items run concurrently in parallel.
- **Strict Stock Constraint Enforcement** — Guarantees balances never drop below zero.
- **Full Audit Trail** — Each inventory mutation tracks a scanner-generated `external_id` for absolute traceability.
- **Daily ETL Aggregations** — Automated summaries featuring total active items, daily turnover, and proactive low-stock alerts.
- **Smart Redis Caching** — Optimizes reporting performance; reports are stored with a 24-hour TTL, while active today reports refresh with a 1-minute TTL.
- **Centralized ServiceContext Container** — Elegant registry-component pattern controlling the boot and shutdown lifecycle of all system resources.

---

## 🏗️ Architectural Architecture

This application adopts a clean, layered architectural layout, wrapped in a centralized dependency injection registry known as **ServiceContext**. 

```mermaid
graph TD
    Client[HTTP Client / Warehouse Scanner] -->|REST API Requests| GinRouter[Gin HTTP Router / Middleware Layer]
    GinRouter -->|1. AuthByRole Middleware| AuthCheck{Validate API Key}
    AuthCheck -->|Failed| Err401[401 Unauthorized / 403 Forbidden]
    AuthCheck -->|Passed| RouteGroup[V1 API Routes]
    
    subgraph Composer Layer
        composer[Composer Wires Repositories, Services, and Handlers]
    end
    
    RouteGroup -->|Invokes Handlers| Handlers[HTTP Handlers: Item, Movement, Report]
    composer -.-> Handlers

    subgraph Service Context Container (Infrastructure Registry)
        sctx[ServiceContext]
        sctx --> ConfigComponent[Config Component]
        sctx --> GinComponent[Gin Component]
        sctx --> GORMComponent[GORM Postgres Component]
        sctx --> RedisComponent[Redis Cache Component]
        sctx --> WorkerComponent[WorkerPool Component]
    end

    Handlers -->|Delegates to| Service[Service Layer]
    Service -->|Interacts with| Repo[Repository Layer]
    Repo -->|Read / Write SQL| GORMComponent
    Service -.->|Cache Read / Write| RedisComponent

    subgraph Movement Concurrency Engine
        Service -->|Group CSV Rows by ItemID| WorkerComponent
        WorkerComponent -->|Goroutine Tasks| ProcessOne[ProcessOne Transaction]
        ProcessOne -->|SQL Transactions| GORMComponent
    end
```

### Infrastructure Lifecycle Container (`ServiceContext`)
Instead of global variables or complex DI frameworks, the system employs the `Component` interface under `pkg/service_context` which manages:
1. **InitFlags**: Registration of CLI flags and parameters.
2. **Activate**: Initializing connections (DB connection pool, Redis cache client, worker pool buffers).
3. **Stop**: Gracefully closing network pools and flushing tasks on SIGTERM/SIGINT.

---

## ⚡ High-Throughput Concurrency Model
To process large batch CSV files efficiently without running into database lock-contention, the system employs a grouped worker pool engine:

```
[Uploaded CSV File] 
        │
        ▼
 [Parse & Validate] ─── (Fail fast if header or row format is corrupted)
        │
        ▼
[Group by Item ID]
   ├── Item 101: [Movement A, Movement B]  ──► Job 1 (Submitted to Pool)
   ├── Item 102: [Movement C]              ──► Job 2 (Submitted to Pool)
   └── Item 103: [Movement D, Movement E]  ──► Job 3 (Submitted to Pool)
        │
        ▼
   [Worker Pool] (Max Workers: Configurable)
     Goroutines pick up Jobs:
     - Worker 1 processes Item 101 sequentially (A -> B)
     - Worker 2 processes Item 102 concurrently
     - Worker 3 processes Item 103 sequentially (D -> E)
```

> [!TIP]
> By processing movements for the same `item_id` sequentially, we avoid database deadlocks and race conditions on the stock balance of that item, while maintaining maximum performance by processing distinct items concurrently across multiple workers.

---

## 📁 Project Directory Structure

The codebase is strictly organized for separation of concerns:

```
inventory-movement-processing/
├── backend/
│   ├── cmd/
│   │   ├── migrations/      # Auto-migration entrypoint binary
│   │   └── server/          # Main HTTP server entrypoint
│   ├── common/              # Global constants, interface helpers, and centralized AppError
│   ├── composer/            # Structural wiring layer (creates repos -> services -> handlers)
│   ├── docs/                # Auto-generated Swagger documentation
│   ├── internal/            # Core business domains (Clean Architecture)
│   │   ├── item/            # Item registration, lookup, and search
│   │   ├── movement/        # High-throughput batch imports and audits
│   │   └── report/          # Daily aggregations, top active ranking, and low-stock alerts
│   └── pkg/                 # Reusable utility infrastructure
│       ├── components/      # Lifecycle-managed components (Redis, GORM, Gin, WorkerPool)
│       ├── core/            # Centralized api response envelope & standard WriteError helper
│       ├── logger/          # Zap-based structured logger
│       └── service_context/ # Centralized infrastructure registry container
├── docker-compose.yml       # Production-ready local stack configurations
└── README.md
```

---

## 🛡️ Security & Role-Based Authorization
API Endpoints are guarded by static API tokens via a GIn HTTP Middleware (`AuthByRole`). Tokens must be passed in the `Authorization` header as a Bearer token:

```http
Authorization: Bearer <api-key>
```

Two distinct roles are mapped in the system configuration:

| Role | Environment Key Name | Target Context |
| :--- | :--- | :--- |
| `STOREKEEPER` | `STOREKEEPER_API_KEY` | Restricted to Item registrations and CSV Batch imports. |
| `MANAGER` | `MANAGER_API_KEY` | Authorized to view inventory summaries, low-stock items, and ranking dashboards. |

---

## 🚀 Getting Started

### 📦 Prerequisites
- [Docker & Docker Compose](https://www.docker.com/) (highly recommended)
- Alternatively: [Go 1.26+](https://go.dev/) with running PostgreSQL and Redis instances.

### 🛠️ Launching the Application
Launch the complete stack (Postgres 16, Redis 7, Adminer DB UI, and Go server) using a single command:

```bash
# 1. Spin up all containers in detached mode
docker compose up -d

# 2. Apply database schemas and seed initial data
docker compose exec backend /migrations
```

After startup:
* **HTTP Server** is listening on: `http://localhost:3000`
* **Swagger Documentation** is available at: `http://localhost:3000/swagger/index.html`
* **Adminer Database Dashboard** is open at: `http://localhost:8080`

---

## 📡 API Reference & cURL Examples

All endpoints return a uniform standard response format (`core.APIResponse`):
```json
{
  "code": 200,
  "message": "success message (optional)",
  "result": { ... }
}
```

Here are interactive cURL examples to test each flow:

### 1. Register a New Item (Role: `STOREKEEPER`)
```bash
curl -X POST http://localhost:3000/api/v1/items \
  -H "Authorization: Bearer 82b81a84c19b8d17e6c0a4b9a3553a4e" \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "SKU-PRO-001",
    "name": "Mechanical Keyboard",
    "description": "RGB Backlit tactile blue switches",
    "price": 89.99,
    "quantity": 100,
    "safety_threshold": 10
  }'
```

### 2. Batch Import Movements (Role: `STOREKEEPER`)
CSV imports process `IN`, `OUT`, and `ADJUST` mutations. 
First, construct a `movements.csv` file:
```csv
external_id,item_id,movement_type,quantity,movement_time,note
EXT-MOV-001,1,IN,50,2026-05-18T10:00:00Z,Regular morning replenishment
EXT-MOV-002,1,OUT,10,2026-05-18T11:30:00Z,Online customer purchase
```

Submit the CSV via multipart form upload:
```bash
curl -X POST http://localhost:3000/api/v1/inventory-movements/import \
  -H "Authorization: Bearer 82b81a84c19b8d17e6c0a4b9a3553a4e" \
  -F "file=@movements.csv"
```

### 3. Fetch Top Active & Low-Stock Alerts Dashboard (Role: `MANAGER`)
Retrieves the most active items of a given date. If requested for today, it also attaches items currently running below their `safety_threshold`.
```bash
curl -X GET "http://localhost:3000/api/v1/reports/daily?date=2026-05-18&limit=5" \
  -H "Authorization: Bearer 0a04630fdcfebe605c3b3c9ece58b824"
```

---

## 📝 Key Technical Assumptions

- **Immutable External ID** — Scanner logs are considered authoritative source records. If a scanner uploads an `external_id` that already exists in the system, it is recognized as a duplicate transaction, skipped safely, and listed under `duplicate` in the response payload without corrupting balances or failing the rest of the batch.
- **Fail-Fast File Schema** — If the uploaded CSV file is physically corrupt or misses header declarations, it fails immediately to prevent partial imports of broken records.
- **Zero Balance Floors** — No scanner is allowed to decrease stock levels below zero. Any action that attempts to do so will be logged as a row-level validation failure, returned in the `failed_rows` report, while other valid rows in the CSV are fully committed.
- **Idempotent ETL Re-runs** — The report aggregation can be manual or automated. Re-running the aggregation for a specific date overwrites previous data for that date, ensuring correct data synchronization after correction imports.
