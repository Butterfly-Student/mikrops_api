# MikroPS API — Project Documentation

> **Author:** Moch Dieqy Dzulqaidar  
> **License:** MIT  
> **Go Version:** >= go1.24.0

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Architecture](#2-architecture)
3. [Project Structure](#3-project-structure)
4. [Technology Stack](#4-technology-stack)
5. [Environment Configuration](#5-environment-configuration)
6. [Getting Started](#6-getting-started)
7. [Database Schema](#7-database-schema)
8. [Domain Models](#8-domain-models)
9. [Authentication & Authorization](#9-authentication--authorization)
10. [Seed Data](#10-seed-data)
11. [Testing](#11-testing)
12. [Makefile Commands](#12-makefile-commands)
13. [Docker Deployment](#13-docker-deployment)

---

## 1. Project Overview

**MikroPS API** is a comprehensive ISP (Internet Service Provider) management backend built with Go. It provides a RESTful API for managing MikroTik routers, PPPoE customers, bandwidth profiles, invoices, payments, and hotspot services.

### Key Features

- 🔐 **JWT Authentication** with role-based access control (Casbin RBAC)
- 🌐 **MikroTik Router Management** — Register and manage multiple MikroTik routers
- 📡 **PPPoE Management** — Manage PPPoE secrets, profiles, and active sessions
- 👥 **Customer Management** — Full customer lifecycle with MikroTik integration
- 📦 **Bandwidth Profiles** — Manage internet packages with MikroTik PPP profile sync
- 🔒 **Customer Isolation** — Auto-isolate overdue customers via MikroTik firewall rules
- 🧾 **Invoice Management** — Auto-generate monthly invoices with late fee calculation
- 💳 **Payment Management** — Multi-method payment tracking with allocation to invoices
- 🔥 **Hotspot Management** — Manage hotspot profiles, users, vouchers, and sessions
- 📊 **Real-time Monitoring** — WebSocket streams for queues, interfaces, and ping
- 🐇 **RabbitMQ Integration** — Async message processing
- ⚡ **Redis Pub/Sub** — Real-time PPPoE session events
- ⏱️ **Temporal Workflows** — Optional workflow engine integration

---

## 2. Architecture

The project follows **Hexagonal Architecture** (Ports and Adapters pattern), ensuring clean separation between business logic and infrastructure.

```
┌─────────────────────────────────────────────────────────────┐
│                        Inbound Adapters                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  HTTP (Gin)  │  │  RabbitMQ    │  │  Temporal Worker │  │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘  │
│         │                 │                    │             │
│  ┌──────▼─────────────────▼────────────────────▼──────────┐ │
│  │                    Inbound Ports                        │ │
│  └──────────────────────────┬──────────────────────────────┘ │
│                             │                                │
│  ┌──────────────────────────▼──────────────────────────────┐ │
│  │                      Domain Layer                        │ │
│  │  Auth │ Customer │ Invoice │ Payment │ Bandwidth │ ...   │ │
│  └──────────────────────────┬──────────────────────────────┘ │
│                             │                                │
│  ┌──────────────────────────▼──────────────────────────────┐ │
│  │                    Outbound Ports                        │ │
│  └──────┬──────────────┬──────────────┬────────────────────┘ │
│         │              │              │                       │
│  ┌──────▼───┐  ┌───────▼──┐  ┌───────▼──┐  ┌────────────┐  │
│  │PostgreSQL│  │ RabbitMQ │  │  Redis   │  │  MikroTik  │  │
│  └──────────┘  └──────────┘  └──────────┘  └────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Layer Descriptions

| Layer | Description |
|---|---|
| **Inbound Adapters** | HTTP handlers (Gin), message consumers (RabbitMQ), workflow workers (Temporal) |
| **Inbound Ports** | Interfaces defining how external systems interact with the domain |
| **Domain Layer** | Core business logic, independent of external systems |
| **Outbound Ports** | Interfaces defining how the domain interacts with external systems |
| **Outbound Adapters** | PostgreSQL, Redis, RabbitMQ, MikroTik API clients |

---

## 3. Project Structure

```
mikrops_api/
├── cmd/
│   ├── main.go                    # Application entry point
│   └── seed/
│       └── main.go                # Seed data command
├── internal/
│   ├── app.go                     # Application initialization & DI
│   ├── adapter/
│   │   ├── inbound/
│   │   │   ├── gin/               # HTTP handlers (Gin framework)
│   │   │   │   ├── route.go       # Route definitions
│   │   │   │   ├── middleware.go  # Auth, RBAC, RouterAuth middleware
│   │   │   │   ├── auth.go        # Auth handlers
│   │   │   │   ├── customer.go    # Customer handlers
│   │   │   │   ├── invoice.go     # Invoice handlers
│   │   │   │   ├── payment.go     # Payment handlers
│   │   │   │   ├── bandwidth_profile.go
│   │   │   │   ├── mikrotik_router.go
│   │   │   │   ├── pppoe_handler.go
│   │   │   │   ├── hotspot_handler.go
│   │   │   │   ├── queue_handler.go
│   │   │   │   ├── interface_handler.go
│   │   │   │   ├── ippool_handler.go
│   │   │   │   └── ping.go
│   │   │   ├── command/           # CLI command handlers
│   │   │   ├── rabbitmq/          # RabbitMQ consumers
│   │   │   └── temporal/          # Temporal workflow workers
│   │   └── outbound/
│   │       ├── postgres/          # PostgreSQL adapters
│   │       ├── mikrotik/          # MikroTik API adapters
│   │       ├── redis/             # Redis adapters
│   │       ├── rabbitmq/          # RabbitMQ publishers
│   │       └── temporal/          # Temporal workflow starters
│   ├── domain/                    # Business logic
│   │   ├── auth/                  # Authentication domain
│   │   ├── customer/              # Customer domain
│   │   ├── invoice/               # Invoice domain
│   │   ├── payment/               # Payment domain
│   │   ├── bandwidth_profile/     # Bandwidth profile domain
│   │   ├── mikrotik_router/       # MikroTik router domain
│   │   ├── pppoe/                 # PPPoE domain
│   │   ├── hotspot/               # Hotspot domain
│   │   ├── queue/                 # Queue domain
│   │   ├── iface/                 # Interface monitoring domain
│   │   ├── ippool/                # IP Pool domain
│   │   ├── ping/                  # Ping domain
│   │   ├── user/                  # User domain
│   │   └── client/                # API client domain
│   ├── migration/
│   │   └── postgres/              # Database migration files
│   ├── model/                     # Data models & DTOs
│   ├── port/
│   │   ├── inbound/               # Inbound port interfaces
│   │   └── outbound/              # Outbound port interfaces
│   └── seeds/                     # Seed data
│       ├── runner/                # Seed execution engine
│       ├── production/            # Production seed data
│       └── testing/               # Test seed data
├── tests/
│   ├── helpers/                   # Test utilities (Testcontainers)
│   └── mocks/                     # Mock implementations
├── utils/                         # Utility functions
│   ├── activity/                  # Activity tracking
│   ├── database/                  # Database utilities
│   ├── jwt/                       # JWT utilities
│   ├── token/                     # Token utilities
│   ├── log/                       # Logging utilities
│   ├── rabbitmq/                  # RabbitMQ utilities
│   └── redis/                     # Redis utilities
├── design-docs/                   # Architecture documentation
├── docker-compose.yml             # Main Docker Compose
├── docker-compose.authentik.yml   # Authentik auth services
├── docker-compose.temporal.yml    # Temporal workflow services
├── Dockerfile                     # Docker image definition
├── Makefile                       # Build & development automation
├── go.mod                         # Go module definition
└── .env.example                   # Environment variables template
```

---

## 4. Technology Stack

| Component | Technology |
|---|---|
| **Language** | Go >= 1.24.0 |
| **HTTP Framework** | Gin |
| **Database** | PostgreSQL |
| **ORM** | GORM |
| **Cache** | Redis |
| **Message Queue** | RabbitMQ |
| **Workflow Engine** | Temporal (optional) |
| **Authentication** | JWT (HS256) + Casbin RBAC |
| **External Auth** | Authentik (JWKS, optional) |
| **MikroTik API** | RouterOS API (port 8728) |
| **Real-time** | WebSocket + Redis Pub/Sub |
| **Testing** | Testcontainers (PostgreSQL) |
| **Containerization** | Docker + Docker Compose |

---

## 5. Environment Configuration

Copy the example file and configure:

```sh
cp .env.example .env
```

### Environment Variables

```env
# Application
APP_MODE=release                    # debug or release
SERVER_PORT=8000                    # HTTP server port

# Security Keys (generate with: openssl rand -hex 32)
INTERNAL_KEY=<32+ char random key>  # Internal service auth key
JWT_SECRET=<32+ char random key>    # JWT access token secret
JWT_REFRESH_SECRET=<32+ char key>   # JWT refresh token secret

# Driver Configuration
OUTBOUND_DATABASE_DRIVER=postgres   # Database driver
OUTBOUND_MESSAGE_DRIVER=rabbitmq    # Message queue driver
OUTBOUND_CACHE_DRIVER=redis         # Cache driver
OUTBOUND_WORKFLOW_DRIVER=           # Workflow driver (temporal or empty)
INBOUND_HTTP_DRIVER=gin             # HTTP server driver
INBOUND_MESSAGE_DRIVER=rabbitmq     # Message consumer driver
AUTH_DRIVER=                        # Auth driver (jwt for JWKS, or empty for bearer key)

# Database
DATABASE_USERNAME=root
DATABASE_PASSWORD=changeme
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=mikrops_db
DATABASE_SSLMODE=disable            # disable, require, verify-ca, verify-full
DATABASE_MIGRATION_DIR=./internal/migration

# RabbitMQ
MESSAGE_HOST=localhost
MESSAGE_PORT=5672
MESSAGE_USER=root
MESSAGE_PASSWORD=changeme
MESSAGE_VHOST=

# Redis
CACHE_HOST=localhost
CACHE_PORT=6379
CACHE_PASSWORD=changeme

# Temporal (optional)
WORKFLOW_HOST=temporal
WORKFLOW_PORT=7233
WORKFLOW_NAMESPACE=default

# External Auth (optional, for Authentik JWKS)
AUTH_JWKS_URL=http://authentik.example.com/application/o/go-template/jwks/

# Message Subscriptions
UPSERT_CLIENT_MESSAGE_SUBSCRIBE=client.upsert.subscribe
```

---

## 6. Getting Started

### Prerequisites

- Go >= 1.24.0
- Docker & Docker Compose
- PostgreSQL, Redis, RabbitMQ (via Docker Compose)

### Quick Start

```sh
# 1. Clone and setup
cp .env.example .env
# Edit .env with your configuration

# 2. Start external services
docker-compose up -d

# 3. Run the application
go run cmd/main.go http

# 4. Seed initial data
go run cmd/seed/main.go
# or
make seed-prod
```

### Running with Makefile

```sh
# Start HTTP server
make http

# Run with Docker (builds image first)
make http BUILD=true

# Run seeds
make seed-prod    # Production seeds (admin user, default clients)
make seed-dev     # Development seeds
make seed-test    # Test seeds
```

---

## 7. Database Schema

### Tables

#### `clients`
API client credentials for service-to-service authentication.

| Column | Type | Description |
|---|---|---|
| `id` | int | Primary key |
| `name` | varchar | Client name |
| `bearer_key` | varchar | Unique API key |
| `created_at` | timestamp | Creation time |
| `updated_at` | timestamp | Last update time |

---

#### `users`
System user accounts for admin/operator access.

| Column | Type | Description |
|---|---|---|
| `id` | bigint | Primary key |
| `name` | varchar | Full name |
| `email` | varchar | Unique email address |
| `password` | varchar | Bcrypt hashed password |
| `role` | varchar | User role (admin, user) |
| `status` | varchar | Account status |
| `created_at` | timestamp | Creation time |
| `updated_at` | timestamp | Last update time |

---

#### `casbin_rules`
RBAC authorization rules (Casbin).

| Column | Type | Description |
|---|---|---|
| `id` | bigint | Primary key |
| `ptype` | varchar | Policy type (p = policy, g = group) |
| `v0` | varchar | Subject (user ID or role) |
| `v1` | varchar | Object (URL path) |
| `v2` | varchar | Action (GET, POST, PUT, DELETE) |

---

#### `mikrotik_routers`
Registered MikroTik router configurations.

| Column | Type | Description |
|---|---|---|
| `id` | uuid | Primary key |
| `name` | varchar | Router display name |
| `address` | varchar | IP address |
| `api_port` | int | RouterOS API port (default: 8728) |
| `rest_port` | int | REST API port (default: 80) |
| `username` | varchar | RouterOS username |
| `password` | varchar | RouterOS password |
| `password_encrypted` | text | Encrypted password |
| `use_ssl` | bool | Use SSL for API connection |
| `router_os_version` | varchar | RouterOS version |
| `identity` | varchar | Router identity name |
| `is_active` | bool | Whether router is active |
| `last_seen_at` | timestamp | Last successful connection |
| `created_at` | timestamp | Creation time |
| `updated_at` | timestamp | Last update time |
| `deleted_at` | timestamp | Soft delete timestamp |

---

#### `bandwidth_profiles`
Internet package/bandwidth profile definitions.

| Column | Type | Description |
|---|---|---|
| `id` | uuid | Primary key |
| `profile_code` | varchar(50) | Unique profile code |
| `name` | varchar(100) | Display name |
| `description` | text | Description |
| `category` | varchar(20) | `residential`, `business`, `corporate`, `promo` |
| `ppp_profile_name` | varchar(100) | MikroTik PPP profile name |
| `local_address` | varchar(45) | Server-side IP or pool |
| `remote_address` | varchar(45) | Client-side IP or pool |
| `parent_queue` | varchar(100) | Parent queue for hierarchical QoS |
| `dns_server` | varchar(100) | DNS server for clients |
| `download_speed` | bigint | Download speed in kbps |
| `upload_speed` | bigint | Upload speed in kbps |
| `burst_download` | bigint | Burst download speed in kbps |
| `burst_upload` | bigint | Burst upload speed in kbps |
| `burst_threshold` | int | Burst threshold % (default: 80) |
| `burst_time` | int | Burst time in seconds (default: 8) |
| `priority` | int | Queue priority 1-8 (default: 8) |
| `queue_type` | varchar(20) | Queue type |
| `shared_users` | int | Shared users count (default: 1) |
| `price_monthly` | decimal(12,2) | Monthly price |
| `price_installation` | decimal(12,2) | Installation fee |
| `tax_rate` | decimal(5,4) | Tax rate (default: 0.11 = 11%) |
| `is_active` | bool | Active status |
| `is_visible` | bool | Visible for new customers |
| `sort_order` | int | Display sort order |
| `created_at` | timestamp | Creation time |
| `updated_at` | timestamp | Last update time |
| `deleted_at` | timestamp | Soft delete timestamp |

---

#### `customers`
Internet service customers.

| Column | Type | Description |
|---|---|---|
| `id` | uuid | Primary key |
| `customer_code` | varchar(50) | Unique customer code |
| `full_name` | varchar(100) | Customer full name |
| `email` | varchar(100) | Email address (unique) |
| `phone` | varchar(20) | Phone number |
| `address` | text | Physical address |
| `latitude` | decimal(10,8) | GPS latitude |
| `longitude` | decimal(11,8) | GPS longitude |
| `status` | varchar(20) | `pending`, `active`, `suspended`, `isolated`, `terminated` |
| `activation_date` | date | Service activation date |
| `installation_date` | date | Installation date |
| `termination_date` | date | Service termination date |
| `expiry_date` | date | Current billing period expiry |
| `router_id` | uuid | FK to mikrotik_routers |
| `ppp_secret_name` | varchar(100) | PPPoE username (unique) |
| `ppp_secret_password` | varchar(255) | PPPoE password (encrypted) |
| `ppp_service` | varchar(20) | `pppoe`, `pptp`, `l2tp`, `ovpn` |
| `static_ip` | varchar(45) | Static IP address |
| `mac_address` | varchar(17) | MAC address |
| `profile_id` | uuid | FK to bandwidth_profiles |
| `previous_profile_id` | uuid | Previous bandwidth profile |
| `billing_cycle` | varchar(20) | `monthly`, `quarterly`, `yearly` |
| `billing_day` | int | Day of month for billing (1-31) |
| `payment_method_preference` | varchar(20) | Preferred payment method |
| `auto_isolate` | bool | Auto-isolate on overdue (default: true) |
| `grace_period_days` | int | Grace period before isolation (default: 3) |
| `notes` | text | Internal notes |
| `tags` | jsonb | Custom tags |
| `created_by` | uuid | Creator user ID |
| `updated_by` | uuid | Last updater user ID |
| `created_at` | timestamp | Creation time |
| `updated_at` | timestamp | Last update time |
| `deleted_at` | timestamp | Soft delete timestamp |

---

#### `invoices`
Customer billing invoices.

| Column | Type | Description |
|---|---|---|
| `id` | uuid | Primary key |
| `invoice_number` | varchar(50) | Unique invoice number |
| `customer_id` | uuid | FK to customers |
| `billing_period_start` | date | Billing period start |
| `billing_period_end` | date | Billing period end |
| `billing_month` | int | Billing month (1-12) |
| `billing_year` | int | Billing year |
| `issue_date` | date | Invoice issue date |
| `due_date` | date | Payment due date |
| `payment_deadline` | timestamp | Payment deadline |
| `subtotal` | decimal(12,2) | Subtotal before tax |
| `tax_amount` | decimal(12,2) | Tax amount |
| `discount_amount` | decimal(12,2) | Discount amount |
| `late_fee` | decimal(12,2) | Late payment fee |
| `total_amount` | decimal(12,2) | Total amount due |
| `paid_amount` | decimal(12,2) | Amount paid so far |
| `balance` | decimal(12,2) | Remaining balance (computed) |
| `status` | varchar(20) | `draft`, `sent`, `partial`, `paid`, `overdue`, `cancelled`, `refunded` |
| `payment_status` | varchar(20) | `unpaid`, `partial`, `paid`, `overpaid` |
| `invoice_type` | varchar(20) | `recurring`, `installation`, `additional`, `refund` |
| `is_auto_generated` | bool | Auto-generated flag |
| `notes` | text | Customer-visible notes |
| `internal_notes` | text | Internal notes (not in PDF) |
| `created_at` | timestamp | Creation time |
| `updated_at` | timestamp | Last update time |
| `deleted_at` | timestamp | Soft delete timestamp |

---

#### `invoice_items`
Line items within an invoice.

| Column | Type | Description |
|---|---|---|
| `id` | uuid | Primary key |
| `invoice_id` | uuid | FK to invoices |
| `item_type` | varchar(20) | `subscription`, `installation`, `equipment`, `other` |
| `description` | varchar(255) | Item description |
| `profile_id` | uuid | FK to bandwidth_profiles (optional) |
| `quantity` | int | Quantity |
| `unit_price` | decimal(12,2) | Unit price |
| `subtotal` | decimal(12,2) | Subtotal |
| `tax_rate` | decimal(5,4) | Tax rate |
| `tax_amount` | decimal(12,2) | Tax amount |
| `total` | decimal(12,2) | Total including tax |
| `is_prorated` | bool | Prorated item flag |
| `proration_days` | int | Number of prorated days |
| `proration_percentage` | decimal(5,2) | Proration percentage |
| `sort_order` | int | Display order |
| `created_at` | timestamp | Creation time |

---

#### `payments`
Customer payment records.

| Column | Type | Description |
|---|---|---|
| `id` | uuid | Primary key |
| `payment_number` | varchar(50) | Unique payment number |
| `customer_id` | uuid | FK to customers |
| `invoice_id` | uuid | FK to invoices (nullable for advance payment) |
| `amount` | decimal(12,2) | Payment amount |
| `allocated_amount` | decimal(12,2) | Amount allocated to invoices |
| `remaining_amount` | decimal(12,2) | Unallocated amount (computed) |
| `payment_method` | varchar(20) | Payment method |
| `payment_date` | timestamp | Payment date |
| `bank_name` | varchar(100) | Bank name (for bank transfer) |
| `bank_account_number` | varchar(50) | Bank account number |
| `bank_account_name` | varchar(100) | Bank account holder name |
| `transaction_reference` | varchar(100) | Transaction reference number |
| `ewallet_provider` | varchar(50) | E-wallet provider (GoPay, OVO, DANA) |
| `ewallet_number` | varchar(50) | E-wallet number |
| `xendit_invoice_id` | varchar(100) | Xendit invoice ID |
| `xendit_external_id` | varchar(100) | Xendit external ID |
| `xendit_payment_channel` | varchar(50) | Xendit payment channel |
| `proof_image` | text | Payment proof image URL |
| `receipt_number` | varchar(50) | Receipt number |
| `status` | varchar(20) | `pending`, `confirmed`, `rejected`, `refunded` |
| `processed_by` | uuid | User who processed the payment |
| `processed_at` | timestamp | Processing timestamp |
| `rejection_reason` | text | Rejection reason |
| `refund_amount` | decimal(12,2) | Refund amount |
| `refund_date` | timestamp | Refund date |
| `refund_reason` | text | Refund reason |
| `notes` | text | Notes |
| `created_at` | timestamp | Creation time |
| `updated_at` | timestamp | Last update time |
| `deleted_at` | timestamp | Soft delete timestamp |

---

#### `payment_allocations`
Allocation of payments to specific invoices.

| Column | Type | Description |
|---|---|---|
| `id` | uuid | Primary key |
| `payment_id` | uuid | FK to payments |
| `invoice_id` | uuid | FK to invoices |
| `allocated_amount` | decimal(12,2) | Allocated amount |
| `created_at` | timestamp | Creation time |

---

## 8. Domain Models

### Customer Status Flow

```
pending → active → suspended → terminated
                ↓
            isolated (auto on overdue)
                ↓
            active (on payment)
```

### Invoice Status Flow

```
draft → sent → partial → paid
              ↓
           overdue (past due date)
              ↓
         cancelled / refunded
```

### Payment Status Flow

```
pending → confirmed
        ↓
      rejected
        ↓
      refunded
```

---

## 9. Authentication & Authorization

### JWT Authentication

The API uses JWT tokens for user authentication:

1. **Login** → Receive `access_token` (short-lived) + `refresh_token` (long-lived)
2. **API Requests** → Include `Authorization: Bearer <access_token>` header
3. **Token Refresh** → Use `refresh_token` to get new `access_token`

### RBAC (Role-Based Access Control)

Authorization is managed by **Casbin** with the following default roles:

| Role | Description |
|---|---|
| `admin` | Full access to all resources |
| `user` | Limited access (read-only for most resources) |

Casbin rules are stored in the `casbin_rules` table and can be managed via seed data.

### API Client Authentication

For service-to-service communication, use bearer key authentication:

```
Authorization: Bearer <bearer_key>
```

Bearer keys are managed via the `/internal/client-*` endpoints.

### RouterAuth Middleware

Routes under `/mikrotik/:router_id/*` automatically validate the router:
1. Checks if `router_id` exists in the database
2. Verifies the router is active (`is_active = true`)
3. Injects the router object into the request context

---

## 10. Seed Data

### Production Seeds

Run with `make seed-prod` or `SEED_ENV=production make seed`:

| Entity | Data |
|---|---|
| Admin User | `admin@mikrotik.local` / `Admin@123` |
| API Client | Default system client |
| Casbin Rules | Admin and user RBAC rules |

### Environment Variables for Seeds

```env
SEED_ENV=production          # production, development, testing
SEED_ENTITIES=all            # all, users, clients, routers, casbin
ADMIN_EMAIL=admin@mikrotik.local
ADMIN_PASSWORD=Admin@123
SEED_ROUTERS=false           # Whether to seed routers
SEED_ROUTER_NAME=            # Router name
SEED_ROUTER_ADDRESS=         # Router IP address
SEED_ROUTER_USERNAME=        # Router username
SEED_ROUTER_PASSWORD=        # Router password
```

### Makefile Seed Commands

```sh
make seed-prod      # Production seeds
make seed-dev       # Development seeds
make seed-test      # Test seeds
make seed-clean     # Clean all seed data
make seed-refresh   # Clean and re-seed
```

---

## 11. Testing

### Unit Tests

```sh
# Run all domain unit tests
go test -cover ./internal/domain/...

# Generate coverage report
go test -coverprofile=coverage.profile -cover ./internal/domain/...
go tool cover -html coverage.profile -o coverage.html
```

### Integration Tests

Integration tests use **Testcontainers** to spin up a real PostgreSQL instance:

```sh
go test ./tests/...
```

Test seeds are automatically loaded when using Testcontainers. When using `TEST_DB_DSN` (external DB), seeds are NOT auto-loaded.

---

## 12. Makefile Commands

### Code Generation

```sh
make model VAL=name                    # Create model/entity
make migration-postgres VAL=name       # Create PostgreSQL migration
make inbound-http-gin VAL=name         # Create HTTP handler
make inbound-message-rabbitmq VAL=name # Create RabbitMQ consumer
make outbound-database-postgres VAL=name # Create PostgreSQL adapter
make generate-mocks                    # Generate mock implementations
```

### Runtime

```sh
make http                              # Run HTTP server
make http BUILD=true                   # Rebuild and run
make message SUB=upsert_client         # Run message consumer
make command CMD=publish_upsert_client VAL=name # Run command
make workflow WFL=upsert_client        # Run workflow worker
```

### Seeds

```sh
make seed                              # Run seeds (current SEED_ENV)
make seed-prod                         # Production seeds
make seed-dev                          # Development seeds
make seed-test                         # Test seeds
make seed-clean                        # Clean seed data
make seed-refresh                      # Clean and re-seed
```

---

## 13. Docker Deployment

### Main Services (PostgreSQL, Redis, RabbitMQ)

```sh
# Start
docker-compose up -d

# Stop
docker-compose down
```

### Authentik (External Auth Provider)

```sh
# Start
docker-compose -f docker-compose.authentik.yml up -d

# Stop
docker-compose -f docker-compose.authentik.yml down
```

### Temporal (Workflow Engine)

```sh
# Start (includes PostgreSQL, Elasticsearch, Temporal Server, Web UI)
docker-compose -f docker-compose.temporal.yml up -d

# Temporal UI: http://localhost:8080
# Temporal Server: localhost:7233

# Stop
docker-compose -f docker-compose.temporal.yml down
```

### Build Application Docker Image

```sh
make build           # Build image
make build BUILD=true # Force rebuild
```

---

## API Documentation

For complete API documentation including all endpoints, request/response examples, see:

📄 **[API Documentation](./API.md)**
