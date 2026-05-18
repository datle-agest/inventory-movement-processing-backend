
# Inventory Movement Processing System

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=for-the-badge&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker)](https://www.docker.com/)

> **A high-throughput, concurrent warehouse inventory service designed to eliminate ghost inventory through strict data consistency and a reliable audit trail.**

---

## Introduction: The "Ghost Inventory" Problem
In fast-paced warehouse environments, barcode scanners continuously log transactions (IN, OUT, ADJUST) simultaneously. If two scanners write to the same item's stock balance without synchronization:
- **Lost Updates** occur, causing discrepancies between digital balances and physical stock.
- **Negative Stock Balances** happen when an OUT scan executes while another is pending, bypassing stock validation checks.
- **Duplicate Transactions** slip through if scanner retry-logs aren't verified.

This system is built specifically to address these issues, guaranteeing **serializable-like correctness at the application level** with maximum parallel performance.

---

##  Table of Contents
- [Key Features](#-key-features)
- [Assumptions](#-assumptions)
- [ Architectural Architecture](#%EF%B8%8F-architectural-architecture)
- [Main Flow](#-high-throughput-concurrency-model)
- [ Project Directory Structure](#-project-directory-structure)
- [ Security & Role-Based Authorization](#%EF%B8%8F-security--role-based-authorization)
- [ Getting Started](#-getting-started)
- [ API Reference & cURL Examples](#-api-reference--curl-examples)

---

## Key Features
- **High-Concurrency Worker Engine** — Groups incoming scanner rows by `item_id` so that mutations to the same item execute sequentially while separate items run concurrently in parallel.
- **Strict Stock Constraint Enforcement** — Guarantees balances never drop below zero.
- **Full Audit Trail** — Each inventory mutation tracks a scanner-generated `external_id` for absolute traceability.
- **Daily ETL Aggregations** — Automated summaries featuring total active items, daily turnover, and proactive low-stock alerts.
- **Smart Redis Caching** — Optimizes reporting performance; reports are stored with a 24-hour TTL, while active today reports refresh with a 1-minute TTL.
- **Centralized ServiceContext Container** — Elegant registry-component pattern controlling the boot and shutdown lifecycle of all system resources.

---

## Assumptions
- **Single warehouse** — the system manages one warehouse only.
- **No login flow** — two static API keys (`STOREKEEPER`, `MANAGER`) are configured in `.env`; no registration or dynamic key issuance.
- **Scanner flow** — scanners don't call the API per scan. Instead, staff collects scanned items into a CSV and submits it via `POST /inventory-movements/import`.
- **CSV format** — files must have a valid header row and use comma as delimiter. Max size: 5MB or 10,000 rows. Any structural mismatch rejects the entire batch (fail-fast).
- **Synchronous import** — clients wait for a direct response (< 30s) with accepted, rejected, and duplicate counts. Failed rows are returned as a downloadable error file.
- **Movement values** — `IN` and `OUT` are always positive; `ADJUST` can be positive or negative.
- **Idempotency via `external_id`** — each CSV row must carry a unique scanner-generated ID. Duplicates are skipped and flagged, not rejected as errors.
- **Master data required** — all `item_id` values must exist before import. Unknown IDs are treated as row-level errors; no auto-creation.
- **ETL re-runnable** — the daily report job can be triggered manually for any date via CLI; re-running overwrites the existing summary for that day.
- ---


## Architecture
<p align="center">
  <img width="4321" height="2184" alt="image" src="https://github.com/user-attachments/assets/589df82b-3622-4d87-b401-add5c8548d2b" />
</p>

## Main Flow
### 📥 1. Batch Import Flow (CSV Import & Worker Concurrency)
High-level workflow of the concurrent batch import engine, showing CSV parsing, item-based worker pool parallelization, and GORM database transaction locking with scanner ID idempotency:

<p align="center">
  <img width="1740" height="1373" alt="image" src="https://github.com/user-attachments/assets/46d08d61-f7e1-4b37-9362-0ada63c5d553" />
</p>

### 📊 2. Report ETL Flow (ETL & Redis Caching)
High-level workflow of the daily aggregation service, highlighting the Redis caching strategy, database fallback query, and real-time low-stock alert merges:

<p align="center">
  <img width="626" height="1244" alt="image" src="https://github.com/user-attachments/assets/6c38892c-0759-448f-ad3a-7a2bc135e246" />
</p>

## Project Directory Structure

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

## Security & Role-Based Authorization
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

## Getting Started

### Prerequisites
- [Docker & Docker Compose](https://www.docker.com/) (highly recommended)
- Alternatively: [Go 1.26+](https://go.dev/) with running PostgreSQL and Redis instances.

### Launching the Application
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

## API Reference & cURL Examples

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
