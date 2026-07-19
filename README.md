# SMS Gateway Service

## Overview

SMS Gateway Service is a distributed backend service responsible for receiving SMS requests, routing them for delivery, tracking their delivery status, and dynamically classifying users based on traffic patterns.

The service is designed to support high-throughput workloads by separating responsibilities into independent applications communicating through RabbitMQ while using PostgreSQL for persistence and Redis for caching and traffic analysis.

A distributed SMS gateway built with Go, RabbitMQ, Redis, and PostgreSQL that supports asynchronous message processing, traffic classification, and delivery tracking.

---

## Features

- Send single SMS
- Send batch SMS
- User balance management
- Delivery status tracking
- Dynamic traffic classification
- RabbitMQ-based asynchronous processing
- Redis-backed caching
- PostgreSQL persistence
- Daily partitioned SMS table
- Swagger / OpenAPI documentation
- Structured logging with Zap
- Concurrent batch processing with bounded worker pool
- Idempotent SMS submission using Client ID

---

# Architecture

The system is composed of three independent applications.

| Application | Responsibility |
|--------------|---------------|
| API | Receives HTTP requests, validates them, stores SMS records, publishes messages to RabbitMQ |
| Traffic Classifier | Periodically analyzes request history and classifies users as Standard or Bulk |
| Message Consumer | Consumes SMS delivery acknowledgements and updates message status |

Each application shares the same dependency wiring while containing only the logic required for its responsibility.

---

# High Level Flow

```text
                    Client
                       │
                 HTTP REST API
                       │
               Validate Request
                       │
              Decrease User Balance
                       │
              Store SMS (Pending)
                       │
         Increment Redis Hour Counter
                       │
          Resolve User Traffic Class
                       │
              Publish to RabbitMQ
                       │
         ┌─────────────┴─────────────┐
         │                           │
 standard / express          bulk / express
         │                           │
         └─────────────┬─────────────┘
                       │
                  SMS Provider
                       │
                  Delivery ACK
                       │
                  sms.ack Queue
                       │
                 Message Consumer
                       │
             Update SMS Delivery Status
```

---

---

# Dependency Injection & Application Bootstrap

The project uses **manual dependency injection** with a shared application bootstrap.

The `App` container is responsible for creating and wiring shared dependencies such as repositories, services, Redis, PostgreSQL, and RabbitMQ components. This provides a single composition root for the entire system.

Different executables are created through an **Application Factory**, where each application implements a common interface:

```go
type Application interface {
    Build() error
    Run() error
}
```

Registered applications:

- API
- Traffic Classifier
- Message Consumer

Each application only builds the components specific to its responsibility, while sharing the common infrastructure and business services.

```
App
 │
 ├── Shared Dependencies
 │     ├── Repositories
 │     ├── Services
 │     ├── Redis
 │     └── RabbitMQ
 │
 └── Application Factory
       ├── API
       ├── Traffic Classifier
       └── Message Consumer
```

This design keeps dependency wiring centralized, reduces duplication, and makes adding new applications straightforward.

---

## Traffic Classification Algorithm

The system automatically classifies users as **Standard** or **Bulk** senders based on their recent SMS traffic.

- Every SMS request increments an **hourly counter** in Redis.
- A background **Traffic Classifier** periodically reads each user's request counts for the **last 4 hours** using Redis.
- If **any** hourly count reaches or exceeds the configured threshold (**500 requests/hour**), the user is classified as **Bulk**; otherwise, the user remains **Standard**.
- Only users whose traffic class has changed are updated. The classifier performs a **bulk database update** and refreshes the Redis cache using a pipeline to minimize database writes and network round trips.

The user's traffic class is then used to determine the RabbitMQ routing key:

| Traffic Class | Ordinary | Express |
|---------------|----------|----------|
| Standard | `standard` | `standard.express` |
| Bulk | `bulk` | `bulk.express` |

This approach enables high-volume users to be routed through dedicated queues while keeping routing decisions fast through Redis caching.


---

# Delivery Tracking

Every SMS is persisted with an initial `Pending` status before being published to RabbitMQ.

Once the SMS provider publishes a delivery acknowledgement, the Message Consumer updates the status in PostgreSQL.

Example:   

```json
{
  "sms_id": 123,
  "status": "delivered"
}
```

The Message Consumer processes this event and updates the SMS status in PostgreSQL.

---

## Idempotency

Each SMS request may include a `client_id`.

The combination of `(user_id, client_id)` uniquely identifies a request, preventing duplicate SMS creation when clients retry requests.

---

# Design Decisions

### RabbitMQ

RabbitMQ decouples the HTTP API from SMS delivery. The API publishes SMS requests asynchronously, allowing message processing to continue independently while improving responsiveness and scalability.

### Redis

Redis is used to cache user traffic classifications and maintain hourly request counters. This enables fast routing decisions and efficient traffic analysis without querying PostgreSQL on every request.

### PostgreSQL

PostgreSQL is the system of record for users and SMS messages. It ensures reliable persistence for user balances, SMS records, and delivery status updates.

### Daily Partitioning

The `sms` table is partitioned daily by `created_at` to improve write performance, optimize time-based queries, and simplify data retention as the dataset grows.

### Background Workers

Traffic classification and delivery acknowledgement processing run as separate applications. Keeping these tasks outside the HTTP API allows each workload to scale and evolve independently.

### Manual Dependency Injection

The project uses a centralized application container to create and inject shared dependencies. This keeps dependency wiring explicit, avoids global state, and simplifies maintenance.

### Bulk Operations

High-volume updates, such as traffic classification changes, are performed using bulk database updates and Redis pipelining to reduce network overhead and improve performance.
---

# RabbitMQ Topology

## Exchanges

### SMS Dispatch

```
sms.dispatch
```

### SMS Acknowledgements

```
sms.ack
```

---

## Dispatch Queues

```
sms.standard.ordinary
sms.standard.express
sms.bulk.ordinary
sms.bulk.express
```

---

## Acknowledgement Queue

```
sms.ack
```

---

# Project Structure

```
cmd/
├── api/                      # HTTP API application
├── traffic_classifier/       # Background service for user traffic classification
└── message_consumer/         # RabbitMQ ACK consumer

internals/
├── app/                      # Application bootstrap, dependency injection and application factory
├── cache/                    # Redis client, cache helpers and key definitions
├── config/                   # Environment configuration loading
├── database/                 # PostgreSQL connection and migrations
├── dtos/                     # Internal DTOs used between layers
├── http/
│   ├── handlers/             # HTTP handlers
│   ├── middleware/           # HTTP middleware
│   ├── requests/             # Request models
│   ├── responses/            # Response models
│   ├── routes/               # Route registration
│   └── errors/               # Domain and HTTP error definitions
├── logger/                   # Zap logger configuration
├── message_broker/           # RabbitMQ publishers, consumers and topology
├── models/                   # Database models (GORM)
├── repositories/             # Data access layer
├── services/                 # Business logic
└── value_objects/            # Domain value objects and enums

deployments/
└── docker-compose.yml        # Local development infrastructure

migrations/                   # SQL database migrations
```

---

# Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.26 |
| HTTP | Gin |
| ORM | GORM |
| Database | PostgreSQL 17 |
| Cache | Redis 7 |
| Messaging | RabbitMQ 4 |
| Logging | Zap |
| API Documentation | Swagger / OpenAPI |
| Configuration | godotenv |

---

# Running the Project

## Start Infrastructure

```bash
docker compose -f deployments/docker-compose.yml up -d
```

---

## Run API

```bash
go run ./cmd/api
```

---

## Run Traffic Classifier

```bash
go run ./cmd/traffic_classifier
```

---

## Run Message Consumer

```bash
go run ./cmd/message_consumer
```

---

# API Endpoints

| Method | Endpoint | Description |
|---------|----------|-------------|
| GET | `/user/:id` | Get user information |
| PUT | `/user/:id/balance` | Increase user balance |
| POST | `/sms` | Send a single SMS |
| POST | `/sms/batch` | Send multiple SMS messages |
| GET | `/sms/report?user_id={id}` | Get SMS report for a user |

Swagger UI is available at:

```
http://localhost:8080/swagger/index.html
```

---

# Design Principles

- Clean Architecture
- Repository Pattern
- Dependency Injection
- Event-driven communication
- Asynchronous processing
- Background workers
- Separation of concerns
- Redis-as-cache
- PostgreSQL partitioning
- Versioned SQL migrations

---

# Future Improvements

- Transactional Outbox Pattern for guaranteed RabbitMQ delivery
- Automatic partition creation and cleanup
- Prometheus metrics
- Distributed tracing (OpenTelemetry)
- Authentication and authorization
- Rate limiting
- Retry and dead-letter queues
- Integration and end-to-end tests