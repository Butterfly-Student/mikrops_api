# MikroPS API — Complete API Documentation

> **Base URL:** `http://localhost:8000`  
> **Content-Type:** `application/json`

---

## Table of Contents

1. [Authentication & Authorization](#1-authentication--authorization)
2. [Auth Endpoints](#2-auth-endpoints)
3. [User Endpoints](#3-user-endpoints)
4. [MikroTik Router Management](#4-mikrotik-router-management)
5. [PPPoE Management](#5-pppoe-management)
6. [Queue Management](#6-queue-management)
7. [Interface Monitoring](#7-interface-monitoring)
8. [IP Pool Management](#8-ip-pool-management)
9. [Bandwidth Profile Management](#9-bandwidth-profile-management)
10. [Customer Management](#10-customer-management)
11. [Invoice Management](#11-invoice-management)
12. [Payment Management](#12-payment-management)
13. [Hotspot Management](#13-hotspot-management)
14. [Ping Management](#14-ping-management)
15. [Internal Endpoints](#15-internal-endpoints)
16. [WebSocket Endpoints](#16-websocket-endpoints)
17. [Webhook Endpoints](#17-webhook-endpoints)
18. [Standard Response Format](#18-standard-response-format)
19. [Error Codes](#19-error-codes)

---

## 1. Authentication & Authorization

### Middleware Types

| Middleware | Description | Header Required |
|---|---|---|
| `UserAuth` | JWT-based user authentication | `Authorization: Bearer <access_token>` |
| `ClientAuth` | API client key or JWT (JWKS) authentication | `Authorization: Bearer <bearer_key>` |
| `InternalAuth` | Internal service key authentication | `Authorization: Bearer <INTERNAL_KEY>` |
| `RBAC` | Role-Based Access Control (Casbin) | Requires `UserAuth` first |
| `RouterAuth` | Validates MikroTik router from path param | Requires `UserAuth` first |

### JWT Token Structure

```json
{
  "sub": 1,
  "role": "admin",
  "exp": 1700000000
}
```

---

## 2. Auth Endpoints

### POST /auth/login

Login with email and password.

**Request Body:**
```json
{
  "email": "admin@mikrotik.local",
  "password": "Admin@123"
}
```

**Response 200:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response 401:**
```json
{
  "error": "invalid credentials"
}
```

---

### POST /auth/register

Register a new user account.

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "secret123",
  "role": "admin"
}
```

> `role` is optional. Defaults to `user` if not provided.

**Response 201:**
```json
{
  "message": "User registered successfully"
}
```

---

### POST /auth/refresh

Refresh access token using a refresh token.

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response 200:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

## 3. User Endpoints

> **Auth:** `UserAuth` required

### POST /user/change-password

Change the authenticated user's password.

**Request Body:**
```json
{
  "old_password": "oldSecret123",
  "new_password": "newSecret456"
}
```

**Response 200:**
```json
{
  "message": "Password changed successfully"
}
```

---

### POST /user/logout

Logout the current user session.

**Response 200:**
```json
{
  "message": "Logged out successfully"
}
```

---

### GET /user/profile

> **Auth:** `UserAuth` + `RBAC` required

Get the authenticated user's profile.

**Response 200:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "role": "admin"
  }
}
```

---

### PUT /user/profile

> **Auth:** `UserAuth` + `RBAC` required

Update the authenticated user's profile.

**Request Body:**
```json
{
  "name": "John Updated",
  "email": "john.updated@example.com"
}
```

**Response 200:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "John Updated",
    "email": "john.updated@example.com",
    "role": "admin"
  }
}
```

---

## 4. MikroTik Router Management

> **Auth:** `UserAuth` required  
> **Base Path:** `/mikrotik`

### POST /mikrotik

Register a new MikroTik router.

**Request Body:**
```json
{
  "name": "Router Utama",
  "address": "192.168.1.1",
  "api_port": 8728,
  "username": "admin",
  "password": "admin123",
  "use_ssl": false,
  "is_active": true
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | ✅ | Router display name |
| `address` | string | ✅ | IP address of the router |
| `api_port` | int | ❌ | RouterOS API port (default: 8728) |
| `username` | string | ✅ | RouterOS username |
| `password` | string | ✅ | RouterOS password |
| `use_ssl` | bool | ❌ | Use SSL for API connection (default: false) |
| `is_active` | bool | ❌ | Whether router is active (default: true) |

**Response 201:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Router Utama",
    "address": "192.168.1.1",
    "api_port": 8728,
    "rest_port": 80,
    "username": "admin",
    "use_ssl": false,
    "is_active": true,
    "router_os_version": null,
    "identity": null,
    "last_seen_at": null,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

---

### GET /mikrotik

List all registered MikroTik routers.

**Response 200:**
```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Router Utama",
      "address": "192.168.1.1",
      "is_active": true,
      "created_at": "2025-01-01T00:00:00Z"
    }
  ]
}
```

---

### GET /mikrotik/:router_id

Get a specific MikroTik router by ID.

**Path Parameters:**
- `router_id` (UUID) — Router ID

**Response 200:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Router Utama",
    "address": "192.168.1.1",
    "api_port": 8728,
    "username": "admin",
    "is_active": true,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

---

### PUT /mikrotik/:router_id

Update a MikroTik router.

**Path Parameters:**
- `router_id` (UUID) — Router ID

**Request Body:** Same as POST /mikrotik

**Response 200:**
```json
{
  "success": true,
  "data": { ... }
}
```

---

### DELETE /mikrotik/:router_id

Delete a MikroTik router.

**Response 200:**
```json
{
  "success": true,
  "data": {
    "message": "MikroTik router deleted successfully"
  }
}
```

---

### POST /mikrotik/:router_id/test

Test connection to a MikroTik router.

**Response 200:**
```json
{
  "success": true,
  "data": {
    "success": true,
    "message": "Connection successful",
    "router_os_version": "7.12",
    "identity": "MikroTik"
  }
}
```

---

### POST /mikrotik/:router_id/isolation/setup

> **Auth:** `UserAuth` + `RouterAuth` required

Setup customer isolation on a MikroTik router. Creates a PPP profile, firewall rules, and NAT rules for isolated customers.

**Request Body (optional overrides):**
```json
{
  "profile_name": "isolir",
  "address_list": "isolated-users",
  "portal_ip": "192.168.100.1",
  "portal_port": "80",
  "dns_server": "8.8.8.8",
  "rate_limit": "256k/256k"
}
```

> `portal_ip` is **required**.

**Response 200:**
```json
{
  "success": true,
  "data": {
    "message": "Isolation setup completed successfully",
    "config": {
      "profile_name": "isolir",
      "address_list": "isolated-users",
      "portal_ip": "192.168.100.1",
      "portal_port": "80",
      "dns_server": "8.8.8.8",
      "rate_limit": "256k/256k"
    }
  }
}
```

---

### GET /mikrotik/:router_id/isolation/status

> **Auth:** `UserAuth` + `RouterAuth` required

Check if isolation is configured on a MikroTik router.

**Response 200:**
```json
{
  "success": true,
  "data": {
    "isolation_configured": true
  }
}
```

---

## 5. PPPoE Management

> **Auth:** `UserAuth` required  
> **Note:** Routes can be accessed via `/pppoe/*` (requires `?router_id=<uuid>` query param) or via `/mikrotik/:router_id/pppoe/*` (router resolved from path).

### PPPoE Secrets

#### POST /mikrotik/:router_id/pppoe/secrets

Create a PPPoE secret (user account) on MikroTik.

**Request Body:**
```json
{
  "name": "customer001",
  "password": "secret123",
  "service": "pppoe",
  "profile": "10Mbps",
  "local_address": "10.0.0.1",
  "remote_address": "10.0.0.100",
  "caller_id": "",
  "limit_bytes_in": 0,
  "limit_bytes_out": 0,
  "routes": "",
  "comment": "Customer 001",
  "disabled": false
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | ✅ | PPPoE username |
| `password` | string | ✅ | PPPoE password |
| `service` | string | ❌ | Service type: `pppoe`, `pptp`, `l2tp` |
| `profile` | string | ❌ | PPP profile name |
| `local_address` | string | ❌ | Server-side IP address |
| `remote_address` | string | ❌ | Client-side IP address |
| `limit_bytes_in` | int64 | ❌ | Download byte limit (0 = unlimited) |
| `limit_bytes_out` | int64 | ❌ | Upload byte limit (0 = unlimited) |
| `disabled` | bool | ❌ | Whether secret is disabled |

**Response 201:**
```json
{
  "message": "Secret created successfully"
}
```

---

#### GET /mikrotik/:router_id/pppoe/secrets

List all PPPoE secrets on MikroTik.

**Response 200:**
```json
[
  {
    "id": "*1",
    "name": "customer001",
    "service": "pppoe",
    "profile": "10Mbps",
    "disabled": false
  }
]
```

---

#### GET /mikrotik/:router_id/pppoe/secrets/:id

Get a specific PPPoE secret by MikroTik ID.

**Response 200:**
```json
{
  "id": "*1",
  "name": "customer001",
  "password": "secret123",
  "service": "pppoe",
  "profile": "10Mbps",
  "local_address": "10.0.0.1",
  "remote_address": "10.0.0.100",
  "disabled": false
}
```

---

#### PUT /mikrotik/:router_id/pppoe/secrets/:id

Update a PPPoE secret.

**Request Body:** Same as POST, with `id` field optional (taken from path).

**Response 200:**
```json
{
  "message": "Secret updated successfully"
}
```

---

#### DELETE /mikrotik/:router_id/pppoe/secrets/:id

Delete a PPPoE secret.

**Response 200:**
```json
{
  "message": "Secret deleted successfully"
}
```

---

### PPPoE Profiles

#### POST /mikrotik/:router_id/pppoe/profiles

Create a PPPoE profile on MikroTik.

**Request Body:**
```json
{
  "name": "10Mbps",
  "local_address": "10.0.0.1",
  "remote_address": "pppoe-pool",
  "rate_limit": "10M/10M",
  "parent_queue": "",
  "queue_type": "",
  "only_one": "default",
  "dns_server": "8.8.8.8",
  "comment": "10 Mbps Package"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | ✅ | Profile name |
| `local_address` | string | ❌ | Server-side IP or pool name |
| `remote_address` | string | ❌ | Client-side IP or pool name |
| `rate_limit` | string | ❌ | Format: `upload/download` e.g. `5M/10M` |
| `dns_server` | string | ❌ | DNS server for clients |
| `only_one` | string | ❌ | `default`, `yes`, or `no` |

**Response 201:**
```json
{
  "message": "Profile created successfully"
}
```

---

#### GET /mikrotik/:router_id/pppoe/profiles

List all PPPoE profiles.

**Response 200:**
```json
[
  {
    "id": "*1",
    "name": "10Mbps",
    "rate_limit": "10M/10M",
    "dns_server": "8.8.8.8"
  }
]
```

---

#### GET /mikrotik/:router_id/pppoe/profiles/:id

Get a specific PPPoE profile.

---

#### PUT /mikrotik/:router_id/pppoe/profiles/:id

Update a PPPoE profile.

---

#### DELETE /mikrotik/:router_id/pppoe/profiles/:id

Delete a PPPoE profile.

---

### PPPoE Sessions

#### GET /mikrotik/:router_id/pppoe/sessions/active

List active PPPoE sessions.

**Response 200:**
```json
[
  {
    "id": "*1",
    "name": "customer001",
    "service": "pppoe",
    "caller_id": "AA:BB:CC:DD:EE:FF",
    "address": "10.0.0.100",
    "uptime": "1d2h3m",
    "encoding": "",
    "session_id": "0x1234",
    "limit_bytes_in": 0,
    "limit_bytes_out": 0,
    "radius": false
  }
]
```

---

#### GET /mikrotik/:router_id/pppoe/sessions/inactive

List inactive PPPoE sessions.

---

#### GET /pppoe/sessions/history

> ⚠️ **Not implemented** — Returns 501.

---

## 6. Queue Management

> **Auth:** `UserAuth` required  
> **Note:** Routes can be accessed via `/queues/*` (requires `?router_id=<uuid>`) or via `/mikrotik/:router_id/queues/*`.

### POST /mikrotik/:router_id/queues

Create a Simple Queue on MikroTik.

**Request Body:**
```json
{
  "name": "customer001-queue",
  "target": "10.0.0.100/32",
  "max_limit": "10M/10M",
  "burst_limit": "20M/20M",
  "burst_threshold": "8M/8M",
  "burst_time": "8/8",
  "priority": "8",
  "parent": "none",
  "comment": "Customer 001 Queue"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | ✅ | Queue name |
| `target` | string | ✅ | Target IP/subnet |
| `max_limit` | string | ❌ | Format: `upload/download` e.g. `10M/10M` |
| `burst_limit` | string | ❌ | Burst limit |
| `priority` | string | ❌ | Queue priority (1-8) |
| `parent` | string | ❌ | Parent queue name |

**Response 201:**
```json
{
  "message": "Queue created successfully"
}
```

---

### GET /mikrotik/:router_id/queues

List all queues.

**Response 200:**
```json
[
  {
    "id": "*1",
    "name": "customer001-queue",
    "target": "10.0.0.100/32",
    "max_limit": "10M/10M"
  }
]
```

---

### GET /mikrotik/:router_id/queues/:id

Get a specific queue.

---

### PUT /mikrotik/:router_id/queues/:id

Update a queue.

---

### DELETE /mikrotik/:router_id/queues/:id

Delete a queue.

---

### Queue Monitoring (Streaming)

#### POST /mikrotik/:router_id/queues/monitor

Start streaming stats for all queues.

**Response 200:**
```json
{
  "message": "Streaming started for all queues"
}
```

---

#### POST /mikrotik/:router_id/queues/monitor/:name

Start streaming stats for a specific queue by name.

---

#### DELETE /mikrotik/:router_id/queues/monitor

Stop streaming all queues.

---

#### DELETE /mikrotik/:router_id/queues/monitor/:name

Stop streaming a specific queue.

---

## 7. Interface Monitoring

> **Auth:** `UserAuth` required  
> **Note:** Routes can be accessed via `/interfaces/*` or via `/mikrotik/:router_id/interfaces/*`.

### POST /mikrotik/:router_id/interfaces/monitor

Start monitoring all interfaces.

**Response 200:**
```json
{
  "message": "Interface monitoring started"
}
```

---

### POST /mikrotik/:router_id/interfaces/monitor/:name

Start monitoring a specific interface by name.

---

### DELETE /mikrotik/:router_id/interfaces/monitor

Stop monitoring all interfaces.

---

### DELETE /mikrotik/:router_id/interfaces/monitor/:name

Stop monitoring a specific interface.

---

## 8. IP Pool Management

> **Auth:** `UserAuth` required  
> **Note:** Routes can be accessed via `/ip-pools/*` or via `/mikrotik/:router_id/ip-pools/*`.

### POST /mikrotik/:router_id/ip-pools

Create an IP Pool on MikroTik.

**Request Body:**
```json
{
  "name": "pppoe-pool",
  "ranges": "10.0.0.100-10.0.0.200",
  "next_pool": "none",
  "comment": "PPPoE IP Pool"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | ✅ | Pool name |
| `ranges` | string | ✅ | IP range e.g. `192.168.1.100-192.168.1.200` |
| `next_pool` | string | ❌ | Next pool name when this pool is exhausted |
| `comment` | string | ❌ | Comment |

**Response 201:**
```json
{
  "message": "IP Pool created successfully"
}
```

---

### GET /mikrotik/:router_id/ip-pools

List all IP pools.

**Response 200:**
```json
[
  {
    "id": "*1",
    "name": "pppoe-pool",
    "ranges": "10.0.0.100-10.0.0.200",
    "next_pool": "none"
  }
]
```

---

### GET /mikrotik/:router_id/ip-pools/:id

Get a specific IP pool.

---

### PUT /mikrotik/:router_id/ip-pools/:id

Update an IP pool.

---

### DELETE /mikrotik/:router_id/ip-pools/:id

Delete an IP pool.

---

## 9. Bandwidth Profile Management

> **Auth:** `UserAuth` required  
> **Base Path:** `/bandwidth-profiles` (DB only) or `/mikrotik/:router_id/bandwidth-profiles` (MikroTik-first)

### POST /bandwidth-profiles

Create a bandwidth profile (database only, no MikroTik sync).

**Request Body:**
```json
{
  "profile_code": "PKG-10M",
  "name": "Paket 10 Mbps",
  "description": "Paket internet 10 Mbps untuk rumahan",
  "category": "residential",
  "ppp_profile_name": "10Mbps",
  "local_address": "10.0.0.1",
  "remote_address": "pppoe-pool",
  "parent_queue": null,
  "dns_server": "8.8.8.8",
  "download_speed": 10240,
  "upload_speed": 5120,
  "burst_download": 20480,
  "burst_upload": 10240,
  "burst_threshold": 80,
  "burst_time": 8,
  "priority": 8,
  "queue_type": "default",
  "shared_users": 1,
  "price_monthly": 150000,
  "price_installation": 200000,
  "tax_rate": 0.11,
  "is_active": true,
  "is_visible": true,
  "sort_order": 1
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `profile_code` | string | ✅ | Unique profile code (max 50 chars) |
| `name` | string | ✅ | Profile display name (max 100 chars) |
| `category` | string | ✅ | `residential`, `business`, `corporate`, `promo` |
| `ppp_profile_name` | string | ✅ | MikroTik PPP profile name |
| `download_speed` | int64 | ✅ | Download speed in kbps (min 1) |
| `upload_speed` | int64 | ✅ | Upload speed in kbps (min 1) |
| `price_monthly` | float64 | ✅ | Monthly price |
| `burst_download` | int64 | ❌ | Burst download speed in kbps |
| `burst_upload` | int64 | ❌ | Burst upload speed in kbps |
| `burst_threshold` | int | ❌ | Burst threshold percentage (1-100, default 80) |
| `burst_time` | int | ❌ | Burst time in seconds (default 8) |
| `priority` | int | ❌ | Queue priority (1-8, default 8) |
| `tax_rate` | float64 | ❌ | Tax rate (0-1, default 0.11 = 11% PPN) |

**Response 201:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "profile_code": "PKG-10M",
    "name": "Paket 10 Mbps",
    "category": "residential",
    "ppp_profile_name": "10Mbps",
    "download_speed": 10240,
    "upload_speed": 5120,
    "price_monthly": 150000,
    "is_active": true,
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

---

### POST /mikrotik/:router_id/bandwidth-profiles

Create a bandwidth profile AND sync to MikroTik PPP profile.

**Request Body:** Same as POST /bandwidth-profiles

**Response 201:** Same as above

---

### GET /bandwidth-profiles

List all bandwidth profiles.

**Query Parameters:**
| Param | Type | Description |
|---|---|---|
| `category` | string | Filter by category |
| `is_active` | bool | Filter by active status |
| `is_visible` | bool | Filter by visibility |
| `min_price` | float64 | Minimum monthly price |
| `max_price` | float64 | Maximum monthly price |
| `search` | string | Search by name, code, or description |

**Response 200:**
```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "profile_code": "PKG-10M",
      "name": "Paket 10 Mbps",
      "category": "residential",
      "download_speed": 10240,
      "upload_speed": 5120,
      "price_monthly": 150000,
      "is_active": true
    }
  ]
}
```

---

### GET /bandwidth-profiles/:id

Get a bandwidth profile by UUID.

---

### GET /bandwidth-profiles/code/:code

Get a bandwidth profile by profile code.

---

### PUT /bandwidth-profiles/:id

Update a bandwidth profile (database only).

**Request Body:** Same as POST /bandwidth-profiles

---

### PUT /mikrotik/:router_id/bandwidth-profiles/:id

Update a bandwidth profile AND sync to MikroTik.

---

### DELETE /bandwidth-profiles/:id

Delete a bandwidth profile (database only).

**Response 200:**
```json
{
  "success": true,
  "data": {
    "message": "Bandwidth profile deleted successfully"
  }
}
```

---

### DELETE /mikrotik/:router_id/bandwidth-profiles/:id

Delete a bandwidth profile AND remove from MikroTik.

---

### POST /bandwidth-profiles/:id/sync

Sync a bandwidth profile to a specific MikroTik router.

**Query Parameters:**
- `router_id` (UUID, required) — Target router ID

**Response 200:**
```json
{
  "success": true,
  "data": {
    "message": "Bandwidth profile synced to MikroTik successfully"
  }
}
```

---

## 10. Customer Management

> **Auth:** `UserAuth` required  
> **Note:** Read-only routes available at `/customers/*`. Full CRUD requires `/mikrotik/:router_id/customers/*`.

### POST /mikrotik/:router_id/customers

Create a new customer. Automatically creates a PPPoE secret on the MikroTik router.

**Request Body:**
```json
{
  "customer_code": "CUST-001",
  "full_name": "Budi Santoso",
  "email": "budi@example.com",
  "phone": "08123456789",
  "address": "Jl. Merdeka No. 1, Jakarta",
  "latitude": -6.2088,
  "longitude": 106.8456,
  "status": "pending",
  "activation_date": "2025-01-01",
  "installation_date": "2025-01-01",
  "expiry_date": "2025-02-01",
  "ppp_secret_name": "budi001",
  "ppp_secret_password": "secret123",
  "ppp_service": "pppoe",
  "static_ip": null,
  "mac_address": null,
  "profile_id": "550e8400-e29b-41d4-a716-446655440000",
  "billing_cycle": "monthly",
  "billing_day": 1,
  "payment_method_preference": "transfer",
  "auto_isolate": true,
  "grace_period_days": 3,
  "notes": "Pelanggan baru",
  "tags": {}
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `customer_code` | string | ✅ | Unique customer code (max 50 chars) |
| `full_name` | string | ✅ | Customer full name (max 100 chars) |
| `phone` | string | ✅ | Phone number (max 20 chars) |
| `email` | string | ❌ | Email address |
| `status` | string | ❌ | `pending`, `active`, `suspended`, `isolated`, `terminated` (default: `pending`) |
| `ppp_service` | string | ❌ | `pppoe`, `pptp`, `l2tp`, `ovpn` (default: `pppoe`) |
| `billing_cycle` | string | ❌ | `monthly`, `quarterly`, `yearly` (default: `monthly`) |
| `billing_day` | int | ❌ | Day of month for billing (1-31, default: 1) |
| `auto_isolate` | bool | ❌ | Auto-isolate on overdue (default: true) |
| `grace_period_days` | int | ❌ | Grace period before isolation (default: 3) |
| `payment_method_preference` | string | ❌ | `cash`, `transfer`, `e-wallet`, `auto-debit` |

**Response 201:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "customer_code": "CUST-001",
    "full_name": "Budi Santoso",
    "email": "budi@example.com",
    "phone": "08123456789",
    "status": "pending",
    "ppp_service": "pppoe",
    "billing_cycle": "monthly",
    "auto_isolate": true,
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

---

### GET /customers

List all customers (no router filter).

**Query Parameters:**
| Param | Type | Description |
|---|---|---|
| `status` | string | Filter by status |
| `router_id` | UUID | Filter by router |
| `profile_id` | UUID | Filter by bandwidth profile |
| `billing_cycle` | string | Filter by billing cycle |
| `auto_isolate` | bool | Filter by auto-isolate setting |
| `expiry_date_from` | date | Filter by expiry date range start |
| `expiry_date_to` | date | Filter by expiry date range end |
| `search` | string | Search by name, code, phone, or email |

**Response 200:**
```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "customer_code": "CUST-001",
      "full_name": "Budi Santoso",
      "status": "active",
      "phone": "08123456789"
    }
  ]
}
```

---

### GET /mikrotik/:router_id/customers

List customers filtered by router.

---

### GET /customers/:id

Get a customer by UUID.

---

### GET /customers/code/:code

Get a customer by customer code.

---

### PUT /mikrotik/:router_id/customers/:id

Update a customer and sync to MikroTik.

**Request Body:** Same as POST /mikrotik/:router_id/customers

---

### DELETE /mikrotik/:router_id/customers/:id

Delete a customer and remove PPPoE secret from MikroTik.

**Response 200:**
```json
{
  "success": true,
  "data": {
    "message": "Customer deleted successfully"
  }
}
```

---

### POST /mikrotik/:router_id/customers/:id/status

Change customer status.

**Request Body:**
```json
{
  "status": "active"
}
```

> Valid values: `pending`, `active`, `suspended`, `isolated`, `terminated`

**Response 200:**
```json
{
  "success": true,
  "data": {
    "message": "Customer status changed successfully",
    "status": "active"
  }
}
```

---

### POST /mikrotik/:router_id/customers/:id/isolate

Isolate a customer (change PPP profile to isolation profile on MikroTik).

**Response 200:**
```json
{
  "success": true,
  "data": {
    "message": "Customer isolated successfully"
  }
}
```

---

### POST /mikrotik/:router_id/customers/:id/unisolate

Remove customer isolation (restore original PPP profile).

**Response 200:**
```json
{
  "success": true,
  "data": {
    "message": "Customer un-isolated successfully"
  }
}
```

---

### POST /mikrotik/:router_id/customers/:id/sync

Sync customer data to MikroTik (update PPPoE secret).

**Response 200:**
```json
{
  "success": true,
  "data": {
    "message": "Customer synced to MikroTik successfully"
  }
}
```

---

## 11. Invoice Management

> **Auth:** `UserAuth` required  
> **Base Path:** `/invoices`

### POST /invoices

Create a new invoice.

**Request Body:**
```json
{
  "invoice_number": "INV-2025-001",
  "customer_id": "550e8400-e29b-41d4-a716-446655440001",
  "billing_period_start": "2025-01-01",
  "billing_period_end": "2025-01-31",
  "billing_month": 1,
  "billing_year": 2025,
  "issue_date": "2025-01-01",
  "due_date": "2025-01-15",
  "subtotal": 150000,
  "tax_amount": 16500,
  "discount_amount": 0,
  "late_fee": 0,
  "total_amount": 166500,
  "status": "draft",
  "invoice_type": "recurring",
  "notes": "Tagihan bulan Januari 2025",
  "items": [
    {
      "item_type": "subscription",
      "description": "Paket Internet 10 Mbps - Januari 2025",
      "profile_id": "550e8400-e29b-41d4-a716-446655440000",
      "quantity": 1,
      "unit_price": 150000,
      "subtotal": 150000,
      "tax_rate": 0.11,
      "tax_amount": 16500,
      "total": 166500
    }
  ]
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `invoice_number` | string | ✅ | Unique invoice number (max 50 chars) |
| `customer_id` | UUID | ✅ | Customer UUID |
| `billing_period_start` | date | ✅ | Billing period start date |
| `billing_period_end` | date | ✅ | Billing period end date |
| `due_date` | date | ✅ | Payment due date |
| `total_amount` | float64 | ✅ | Total invoice amount |
| `status` | string | ❌ | `draft`, `sent`, `partial`, `paid`, `overdue`, `cancelled`, `refunded` |
| `invoice_type` | string | ❌ | `recurring`, `installation`, `additional`, `refund` |

**Response 201:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "invoice_number": "INV-2025-001",
    "customer_id": "550e8400-e29b-41d4-a716-446655440001",
    "status": "draft",
    "total_amount": 166500,
    "paid_amount": 0,
    "balance": 166500,
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

---

### GET /invoices

List all invoices.

**Query Parameters:**
| Param | Type | Description |
|---|---|---|
| `customer_id` | UUID | Filter by customer |
| `status` | string | Filter by invoice status |
| `payment_status` | string | Filter by payment status (`unpaid`, `partial`, `paid`, `overpaid`) |
| `invoice_type` | string | Filter by invoice type |
| `billing_month` | int | Filter by billing month (1-12) |
| `billing_year` | int | Filter by billing year |
| `due_date_from` | date | Filter by due date range start |
| `due_date_to` | date | Filter by due date range end |
| `search` | string | Search by invoice number |

---

### GET /invoices/:id

Get an invoice by UUID.

---

### GET /invoices/number/:number

Get an invoice by invoice number.

---

### PUT /invoices/:id

Update an invoice.

**Request Body:** Same as POST /invoices

---

### DELETE /invoices/:id

Delete an invoice (soft delete).

---

### POST /invoices/generate/:customer_id

Auto-generate a monthly invoice for a customer.

**Query Parameters:**
- `month` (int, optional) — Month (1-12), defaults to current month
- `year` (int, optional) — Year, defaults to current year

**Response 201:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "invoice_number": "INV-2025-01-CUST001",
    "customer_id": "550e8400-e29b-41d4-a716-446655440001",
    "billing_month": 1,
    "billing_year": 2025,
    "total_amount": 166500,
    "status": "draft"
  }
}
```

---

### POST /invoices/:id/late-fee

Calculate and apply late fee to an overdue invoice.

**Response 200:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "late_fee": 15000,
    "total_amount": 181500
  }
}
```

---

## 12. Payment Management

> **Auth:** `UserAuth` required  
> **Base Path:** `/payments`

### POST /payments

Create a new payment record.

**Request Body:**
```json
{
  "payment_number": "PAY-2025-001",
  "customer_id": "550e8400-e29b-41d4-a716-446655440001",
  "invoice_id": "550e8400-e29b-41d4-a716-446655440002",
  "amount": 166500,
  "payment_method": "bank_transfer",
  "payment_date": "2025-01-10T10:00:00Z",
  "bank_name": "BCA",
  "bank_account_number": "1234567890",
  "bank_account_name": "Budi Santoso",
  "transaction_reference": "TRF20250110001",
  "proof_image": "https://storage.example.com/proof/pay001.jpg",
  "receipt_number": "RCP-001",
  "status": "pending",
  "notes": "Transfer via ATM"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `payment_number` | string | ✅ | Unique payment number (max 50 chars) |
| `customer_id` | UUID | ✅ | Customer UUID |
| `amount` | float64 | ✅ | Payment amount |
| `payment_method` | string | ✅ | `cash`, `bank_transfer`, `e-wallet`, `credit_card`, `debit_card`, `check`, `xendit` |
| `payment_date` | datetime | ✅ | Payment date and time |
| `invoice_id` | UUID | ❌ | Invoice UUID (null for advance payment) |
| `status` | string | ❌ | `pending`, `confirmed`, `rejected`, `refunded` (default: `pending`) |

**Response 201:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440003",
    "payment_number": "PAY-2025-001",
    "customer_id": "550e8400-e29b-41d4-a716-446655440001",
    "amount": 166500,
    "payment_method": "bank_transfer",
    "status": "pending",
    "created_at": "2025-01-10T10:00:00Z"
  }
}
```

---

### GET /payments

List all payments.

**Query Parameters:**
| Param | Type | Description |
|---|---|---|
| `customer_id` | UUID | Filter by customer |
| `invoice_id` | UUID | Filter by invoice |
| `status` | string | Filter by payment status |
| `payment_method` | string | Filter by payment method |
| `payment_date_from` | datetime | Filter by payment date range start |
| `payment_date_to` | datetime | Filter by payment date range end |
| `search` | string | Search by payment number or reference |

---

### GET /payments/:id

Get a payment by UUID.

---

### GET /payments/number/:number

Get a payment by payment number.

---

### PUT /payments/:id

Update a payment record.

---

### DELETE /payments/:id

Delete a payment (soft delete).

---

### POST /payments/:id/confirm

Confirm a pending payment. Requires authenticated user.

**Response 200:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440003",
    "status": "confirmed",
    "processed_at": "2025-01-10T11:00:00Z"
  }
}
```

---

### POST /payments/:id/reject

Reject a pending payment.

**Request Body:**
```json
{
  "reason": "Bukti transfer tidak valid"
}
```

**Response 200:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440003",
    "status": "rejected",
    "rejection_reason": "Bukti transfer tidak valid"
  }
}
```

---

### POST /payments/:id/allocate

Allocate a payment to a specific invoice.

**Request Body:**
```json
{
  "invoice_id": "550e8400-e29b-41d4-a716-446655440002",
  "amount": 166500
}
```

**Response 201:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440004",
    "payment_id": "550e8400-e29b-41d4-a716-446655440003",
    "invoice_id": "550e8400-e29b-41d4-a716-446655440002",
    "allocated_amount": 166500,
    "created_at": "2025-01-10T11:00:00Z"
  }
}
```

---

## 13. Hotspot Management

> **Auth:** `UserAuth` + `RouterAuth` required  
> **Base Path:** `/mikrotik/:router_id/hotspot`

### Hotspot Profiles

#### POST /mikrotik/:router_id/hotspot/profiles

Create a hotspot profile on MikroTik.

**Request Body:**
```json
{
  "name": "1hour",
  "shared_users": 1,
  "rate_limit": "2M/2M",
  "session_timeout": "1h",
  "idle_timeout": "10m",
  "keepalive_timeout": "2m",
  "status_autorefresh": "1m",
  "add_mac_cookie": true,
  "mac_cookie_timeout": "3d"
}
```

**Response 201:**
```json
{
  "message": "Profile created successfully"
}
```

---

#### GET /mikrotik/:router_id/hotspot/profiles

List all hotspot profiles.

---

#### GET /mikrotik/:router_id/hotspot/profiles/:name

Get a hotspot profile by name.

---

#### PUT /mikrotik/:router_id/hotspot/profiles/:name

Update a hotspot profile.

---

#### DELETE /mikrotik/:router_id/hotspot/profiles/:name

Delete a hotspot profile.

---

### Hotspot Users

#### POST /mikrotik/:router_id/hotspot/users

Create a hotspot user.

**Request Body:**
```json
{
  "name": "user001",
  "password": "pass123",
  "profile": "1hour",
  "limit_uptime": "1h",
  "limit_bytes_in": 0,
  "limit_bytes_out": 0,
  "comment": "Voucher user"
}
```

**Response 201:**
```json
{
  "message": "User created successfully"
}
```

---

#### GET /mikrotik/:router_id/hotspot/users

List all hotspot users.

**Query Parameters:**
- `profile` (string, optional) — Filter by profile name

---

#### GET /mikrotik/:router_id/hotspot/users/:username

Get a hotspot user by username.

---

#### PUT /mikrotik/:router_id/hotspot/users/:username

Update a hotspot user.

---

#### DELETE /mikrotik/:router_id/hotspot/users/:username

Delete a hotspot user.

---

### Voucher Generation

#### POST /mikrotik/:router_id/hotspot/vouchers

Generate hotspot vouchers.

**Request Body:**
```json
{
  "mode": "vc",
  "quantity": 10,
  "profile": "1hour",
  "prefix": "VCH",
  "length": 8,
  "comment": "Batch voucher Januari 2025"
}
```

| Field | Type | Description |
|---|---|---|
| `mode` | string | `vc` (voucher code, username=password) or `up` (username/password separate) |
| `quantity` | int | Number of vouchers to generate |
| `profile` | string | Hotspot profile name |
| `prefix` | string | Voucher code prefix |
| `length` | int | Length of generated code |

**Response 201:**
```json
{
  "vouchers": [
    {
      "username": "VCH12345678",
      "password": "VCH12345678",
      "profile": "1hour"
    }
  ],
  "count": 10
}
```

---

### Hotspot Sessions

#### GET /mikrotik/:router_id/hotspot/sessions

Get active hotspot sessions.

**Response 200:**
```json
[
  {
    "id": "*1",
    "user": "user001",
    "address": "192.168.88.100",
    "mac_address": "AA:BB:CC:DD:EE:FF",
    "uptime": "1h30m",
    "bytes_in": 1048576,
    "bytes_out": 2097152
  }
]
```

---

#### GET /mikrotik/:router_id/hotspot/sessions/stats

Get hotspot session statistics.

**Response 200:**
```json
{
  "total_sessions": 25,
  "active_sessions": 10,
  "total_bytes_in": 104857600,
  "total_bytes_out": 209715200
}
```

---

#### DELETE /mikrotik/:router_id/hotspot/sessions/:username

Disconnect a hotspot user session.

**Response 200:**
```json
{
  "message": "User disconnected successfully"
}
```

---

### Hotspot Sales

#### POST /mikrotik/:router_id/hotspot/sales

Record a hotspot voucher sale.

**Request Body:**
```json
{
  "voucher_code": "VCH12345678",
  "profile": "1hour",
  "price": 5000,
  "sold_by": "kasir001",
  "notes": "Penjualan voucher"
}
```

**Response 201:**
```json
{
  "message": "Sale recorded successfully"
}
```

---

#### GET /mikrotik/:router_id/hotspot/sales

Get hotspot sales records.

**Query Parameters:**
- `start_date` (string, optional) — Start date filter
- `end_date` (string, optional) — End date filter
- `prefix` (string, optional) — Filter by voucher prefix

---

#### GET /mikrotik/:router_id/hotspot/sales/revenue

Get total revenue from hotspot sales.

**Query Parameters:**
- `start_date` (string, optional) — Start date
- `end_date` (string, optional) — End date

**Response 200:**
```json
{
  "revenue": 500000,
  "start_date": "2025-01-01",
  "end_date": "2025-01-31"
}
```

---

### Expiry Schedulers

#### POST /mikrotik/:router_id/hotspot/schedulers/:profile

Create an expiry scheduler for a hotspot profile. Automatically removes expired users.

**Response 201:**
```json
{
  "message": "Expiry scheduler created successfully"
}
```

---

#### DELETE /mikrotik/:router_id/hotspot/schedulers/:profile

Remove an expiry scheduler for a hotspot profile.

**Response 200:**
```json
{
  "message": "Expiry scheduler removed successfully"
}
```

---

## 14. Ping Management

> **Auth:** `UserAuth` required  
> **Note:** Routes can be accessed via `/ping/*` or via `/mikrotik/:router_id/ping/*`.

### POST /ping

Start a ping session to a target address.

**Request Body:**
```json
{
  "address": "8.8.8.8",
  "count": 0,
  "size": 56,
  "interval": "1s"
}
```

| Field | Type | Description |
|---|---|---|
| `address` | string | Target IP or hostname |
| `count` | int | Number of pings (0 = continuous) |
| `size` | int | Packet size in bytes (default: 56) |
| `interval` | string | Ping interval e.g. `1s`, `500ms` |

**Query Parameters:**
- `router_id` (UUID, required for `/ping` route) — Router ID

**Response 200:**
```json
{
  "message": "Ping started",
  "address": "8.8.8.8"
}
```

---

### DELETE /ping/:address

Stop a ping session to a specific address.

**Response 200:**
```json
{
  "message": "Ping stopped",
  "address": "8.8.8.8"
}
```

---

### GET /v1/ping

> **Auth:** `ClientAuth` required

Health check endpoint for API clients.

**Response 200:**
```json
{
  "success": true,
  "data": "pong"
}
```

---

## 15. Internal Endpoints

> **Auth:** `InternalAuth` required (`Authorization: Bearer <INTERNAL_KEY>`)  
> **Base Path:** `/internal`

These endpoints are for internal service-to-service communication.

### POST /internal/client-upsert

Create or update an API client.

**Request Body:**
```json
{
  "name": "service-name",
  "bearer_key": "optional-custom-key"
}
```

**Response 200:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "service-name",
    "bearer_key": "generated-or-provided-key",
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

---

### POST /internal/client-find

Find API clients by filter.

**Request Body:**
```json
{
  "ids": [1, 2],
  "names": ["service-name"],
  "bearer_keys": ["key1", "key2"]
}
```

---

### DELETE /internal/client-delete

Delete an API client.

**Request Body:**
```json
{
  "ids": [1],
  "names": [],
  "bearer_keys": []
}
```

---

## 16. WebSocket Endpoints

WebSocket connections for real-time data streaming.

### GET /ws/ping

Real-time ping results stream.

**Query Parameters:**
- `address` (string, required) — Target address to filter ping results

**Message Format:**
```json
{
  "host": "8.8.8.8",
  "seq": 1,
  "size": 56,
  "ttl": 64,
  "time": "5ms",
  "status": "success",
  "sent_time": "2025-01-01T00:00:00Z",
  "received_time": "2025-01-01T00:00:00.005Z"
}
```

---

### GET /ws/pppoe

Real-time PPPoE session events stream.

**Message Format:**
```json
{
  "event": "session_up",
  "data": {
    "user": "customer001",
    "ip-address": "10.0.0.100",
    "caller-id": "AA:BB:CC:DD:EE:FF",
    "session-id": "0x1234",
    "interface": "ether1",
    "uptime": "0s"
  }
}
```

> Events: `session_up`, `session_down`

---

### GET /ws/queues

Real-time queue statistics stream.

**Query Parameters:**
- `name` (string, optional) — Filter by queue name

**Message Format:**
```json
{
  "name": "customer001-queue",
  "bytes-in": 1048576,
  "bytes-out": 2097152,
  "packets-in": 1000,
  "packets-out": 2000,
  "rate-in": 10240,
  "rate-out": 20480,
  "packet-rate-in": 100,
  "packet-rate-out": 200
}
```

---

### GET /ws/interfaces

Real-time network interface statistics stream.

**Query Parameters:**
- `name` (string, optional) — Filter by interface name

**Message Format:**
```json
{
  "name": "ether1",
  "rx-bits-per-second": 10485760,
  "tx-bits-per-second": 5242880,
  "rx-packets-per-second": 1000,
  "tx-packets-per-second": 500,
  "rx-drops-per-second": 0,
  "tx-drops-per-second": 0,
  "rx-errors-per-second": 0,
  "tx-errors-per-second": 0
}
```

---

## 17. Webhook Endpoints

### POST /webhooks/pppoe/on-up

Callback from MikroTik when a PPPoE session comes up.

**Request Body:**
```json
{
  "user": "customer001",
  "ip-address": "10.0.0.100",
  "caller-id": "AA:BB:CC:DD:EE:FF",
  "session-id": "0x1234",
  "interface": "ether1",
  "uptime": "0s",
  "bytes-in": 0,
  "bytes-out": 0,
  "packets-in": 0,
  "packets-out": 0
}
```

**Response 200:**
```json
{
  "message": "OK"
}
```

---

### POST /webhooks/pppoe/on-down

Callback from MikroTik when a PPPoE session goes down.

**Request Body:** Same as `/webhooks/pppoe/on-up`

---

## 18. Standard Response Format

Most endpoints return a standard response envelope:

```json
{
  "success": true,
  "data": { ... },
  "error": ""
}
```

| Field | Type | Description |
|---|---|---|
| `success` | bool | Whether the request was successful |
| `data` | any | Response payload (null on error) |
| `error` | string | Error message (empty on success) |

---

## 19. Error Codes

| HTTP Status | Description |
|---|---|
| `200 OK` | Request successful |
| `201 Created` | Resource created successfully |
| `400 Bad Request` | Invalid request body or parameters |
| `401 Unauthorized` | Missing or invalid authentication token |
| `403 Forbidden` | Insufficient permissions (RBAC) |
| `404 Not Found` | Resource not found |
| `500 Internal Server Error` | Server-side error |
| `502 Bad Gateway` | MikroTik connection error |

### Error Response Example

```json
{
  "success": false,
  "error": "record not found"
}
```

Or for auth errors:
```json
{
  "error": "Invalid token: token is expired"
}
```
