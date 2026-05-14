# 📦 Inventory Movement Processing System
 
> A warehouse inventory management system — built to eliminate ghost inventory through reliable, concurrent stock movement processing.
 
---
 
## Introduction
Warehouses often report stock that isn't actually there — a problem known as **ghost inventory**. It happens when scanners write data at the same time without proper conflict handling, when duplicate transactions go undetected, or when there's simply no reliable audit trail to trace what went wrong.

## Table of Contents
- [About](#about)
- [Features](#features)
- [Assumptions](#assumptions)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Security](#security)
---

## About
Inventory Movement Processing System is a service that processes stock movements (IN, OUT, ADJUST) submitted in bulk from warehouse scanners via CSV files. It uses a concurrent worker pool backed by PostgreSQL transactions to guarantee that stock quantities are always consistent — no negative balances, no lost updates, no duplicate records.

Each movement is recorded with a unique external ID from the scanner, giving warehouse staff a complete audit trail per item. At the end of each day, an ETL job aggregates the raw movement data into a summary table, which powers the reporting API — surfacing totals, top active items, and low-stock alerts.

---

## Features
- **Item Management** — create and query inventory items with filtering.
- **Batch Import** — upload CSV movement files with size validation; returns accepted, rejected, and duplicate counts along with a downloadable error file.
- **Concurrent Processing** — worker pool with DB transactions ensures no negative stock or lost updates; duplicates are detected via `external_id` and skipped gracefully.
- **Audit Trail** — full movement history per item, traceable back to the source scanner.
- **Daily ETL Reports** — automated end-of-day aggregation into summary reports (totals, top K active items, low-stock alerts); re-runnable via CLI for any date.

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

## Tech Stack
| Layer | Technology |
| :--- | :--- |
| **Language** | Go (Golang) |
| **API Style** | REST API |
| **Database** | PostgreSQL |
| **Local Dev** | Docker Compose |
| **Authentication** | API Key |
| **Concurrency** | Goroutines, Channels, DB Transactions |
| **CI** | GitHub Actions |

---

## Architecture

---

## Project Structure
```
inventory-movement-processing
├── backend
│   ├── cmd
│   │   ├── import-sample
│   │   ├── migrations
│   │   └── server
│   │       └── routes
│   │           └── v1
│   ├── common
│   ├── composer
│   ├── docs
│   ├── internal
│   │   ├── item
│   │   │   ├── entity
│   │   │   ├── repository
│   │   │   │   └── postgres
│   │   │   ├── service
│   │   │   └── transport
│   │   │       └── http
│   │   ├── movement
│   │   │   ├── entity
│   │   │   ├── repository
│   │   │   │   └── postgres
│   │   │   ├── service
│   │   │   └── transport
│   │   │       └── http
│   │   └── report
│   │       ├── entity
│   │       ├── repository
│   │       │   └── postgres
│   │       ├── service
│   │       └── transport
│   │           └── http
│   └── pkg
│       ├── components
│       │   ├── configc
│       │   ├── ginc
│       │   │   └── middleware
│       │   ├── gormc
│       │   │   └── dialets
│       │   ├── jwtc
│       │   ├── redisc
│       │   └── workerc
│       ├── core
│       ├── logger
│       │   └── zap
│       ├── migrations
│       └── service_context
├── docker-compose.yml
└── README.md
```
---

## Getting Started
### How to run
```bash
# Start containers
docker compose up -d
 
# Run migration
docker compose exec backend /migrations
```
---

## Security
All API endpoints are protected by static API keys configured in the environment. Keys are passed via the `Authorization` header as a Bearer token:
 
```
Authorization: Bearer <api-key>
```
There are two hardcoded roles, both with access to all endpoints:
 
| Role | Key | Access |
|---|---|---|
| `STOREKEEPER` | `API_KEY_STOREKEEPER` | Item management and batch import |
| `MANAGER` | `API_KEY_MANAGER` | Reporting and inventory viewing |
 
> There is no login flow or dynamic key issuance. Keys are set once in `.env` and shared within the team.

---
