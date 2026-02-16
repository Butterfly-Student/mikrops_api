# MikrOps Development Planning

Dokumen ini berisi analisis status implementasi project MikrOps berdasarkan planning awal,
serta planning ulang untuk task-task yang belum selesai.

Tanggal analisis: 2026-02-16

---

## Ringkasan Status Per Phase

| Phase | Nama                          | Status         | Progress |
|-------|-------------------------------|----------------|----------|
| 0     | GORM Foundation               | SELESAI        | 100%     |
| 1     | Shared Utilities              | SELESAI        | 100%     |
| 2     | Models                        | SELESAI        | 100%     |
| 3     | Migrations                    | SELESAI        | 100%     |
| 4     | Outbound Port Interfaces      | SELESAI        | 100%     |
| 5     | Outbound Database Adapters    | SELESAI        | 100%     |
| 6     | Cache + MikroTik Adapters     | SELESAI        | 100%     |
| 7     | Domain Logic                  | SELESAI        | 100%     |
| 8     | Inbound Port Interfaces       | SELESAI        | 100%     |
| 9     | Inbound Gin HTTP Adapters     | SELESAI        | 100%     |
| 10    | Tests                         | SELESAI        | 100%     |
| 11    | Configuration                 | SELESAI        | 100%     |

**Overall Progress: 100%**

---

## Detail Analisis Per Phase

### Phase 0: GORM Foundation - SELESAI

Semua task sudah diimplementasi:

- [x] 0.1 Add GORM dependencies (`gorm.io/gorm v1.31.1`, `gorm.io/driver/postgres v1.6.0`)
- [x] 0.2 Modify `utils/database/database.go` - InitGormDatabase dengan connection pooling
- [x] 0.3 Modify `internal/port/outbound/registry_database.go` - WithTenantScope support
- [x] 0.4 Modify `internal/adapter/outbound/postgres/registry.go` - GORM-based adapter
- [x] 0.5 Modify `internal/adapter/outbound/postgres/client.go` - GORM CRUD operations
- [x] 0.6 Modify `internal/app.go` - InitGormDatabase wiring
- [x] 0.7 Update tests - client_test.go dan integration test

### Phase 1: Shared Utilities - SELESAI

Semua utilities sudah dibuat:

- [x] 1.1 `utils/gorm/scope.go` - TenantScope function
- [x] 1.2 `utils/crypto/encryption.go` - AES-GCM encryption/decryption
- [x] 1.3 `utils/jwt/signer.go` - GenerateStaffToken, GenerateCustomerToken, ParseStaffToken, ParseCustomerToken, refresh tokens
- [x] 1.4 `utils/jwt/validator.go` - HMAC validation untuk self-issued tokens
- [x] 1.5 `utils/activity/activity.go` - Context keys: TenantID, StaffID, CustomerID, UserRole, UserPermissions
- [x] 1.6 `.env.example` - JWT_SECRET, ENCRYPTION_KEY, UPLOAD_PATH

### Phase 2: Models - SELESAI

Semua 13 model files telah dibuat di `internal/model/`:

- [x] 2.1 `tenant.go` - Tenant, TenantInput, TenantFilter, TenantPrepare
- [x] 2.2 `role.go` - Role, RoleInput, RoleFilter
- [x] 2.3 `permission.go` - Permission, PermissionInput, PermissionFilter, RolePermission, constants
- [x] 2.4 `staff.go` - Staff, StaffInput, StaffFilter
- [x] 2.5 `nas.go` - Nas, NasInput, NasFilter
- [x] 2.6 `internet_package.go` - InternetPackage, InternetPackageInput, InternetPackageFilter
- [x] 2.7 `customer.go` - Customer, CustomerInput, CustomerFilter
- [x] 2.8 `subscription.go` - Subscription, SubscriptionInput, SubscriptionFilter
- [x] 2.9 `payment_method.go` - PaymentMethod, PaymentMethodInput, PaymentMethodFilter
- [x] 2.10 `invoice.go` - Invoice, InvoiceInput, InvoiceFilter
- [x] 2.11 `payment.go` - Payment, PaymentInput, PaymentFilter
- [x] 2.12 `mikrotik.go` - Non-persisted structs (Connection, Traffic, PPPoE, Hotspot, Queue, Profile)
- [x] 2.13 `auth.go` - StaffAuthRequest, CustomerAuthRequest, AuthTokenResponse, JWT Claims

### Phase 3: Migrations - SELESAI

Semua 14 migration files telah dibuat di `internal/migration/postgres/`:

- [x] `1_client.go` - clients table
- [x] `2_tenant.go` - tenants table
- [x] `3_role.go` - roles table
- [x] `4_permission.go` - permissions table
- [x] `5_role_permission.go` - role_permissions join table
- [x] `6_staff.go` - staffs table (FK: tenants, roles)
- [x] `7_nas.go` - nas table (FK: tenants)
- [x] `8_internet_package.go` - internet_packages table (FK: tenants)
- [x] `9_customer.go` - customers table (FK: tenants, nas)
- [x] `10_subscription.go` - subscriptions table (FK: tenants, customers, internet_packages, nas)
- [x] `11_payment_method.go` - payment_methods table (FK: tenants)
- [x] `12_invoice.go` - invoices table (FK: tenants, customers, subscriptions)
- [x] `13_payment.go` - payments table (FK: tenants, invoices, payment_methods, staffs)
- [x] `14_seed_roles_permissions.go` - Seed data untuk roles, permissions, role_permissions

### Phase 4: Outbound Port Interfaces - SELESAI

Semua port interfaces telah dibuat di `internal/port/outbound/`:

- [x] `tenant.go`, `role.go`, `permission.go`, `staff.go`, `nas.go`
- [x] `internet_package.go`, `customer.go`, `subscription.go`
- [x] `payment_method.go`, `invoice.go`, `payment.go`
- [x] `mikrotik.go` - MikrotikPort dengan REST + RouterOS operations
- [x] `registry_database.go` - Semua entity accessors + WithTenantScope
- [x] `registry_cache.go` - Staff(), Customer()
- [x] `registry_http.go` - Mikrotik()

### Phase 5: Outbound Database Adapters - SELESAI

Semua adapter files telah dibuat di `internal/adapter/outbound/postgres/`:

- [x] `tenant.go`, `role.go`, `permission.go`, `staff.go`, `nas.go`
- [x] `internet_package.go`, `customer.go`, `subscription.go`
- [x] `payment_method.go`, `invoice.go`, `payment.go`
- [x] `registry.go` - Updated dengan semua entity adapters

### Phase 6: Cache + MikroTik Adapters - SELESAI

Cache Adapters (`internal/adapter/outbound/redis/`):
- [x] `staff.go` - Session caching untuk staff
- [x] `customer.go` - Session caching untuk customer
- [x] `registry.go` - Updated

MikroTik Adapters (`internal/adapter/outbound/mikrotik/`):
- [x] `rest_client.go` - HTTP client untuk MikroTik REST API
- [x] `routeros_client.go` - RouterOS API client (go-routeros/routeros/v3)
- [x] `adapter.go` - Main adapter implementing MikrotikPort
- [x] `registry.go` - Registry

Wiring:
- [x] `internal/app.go` - httpOutbound() function, inject ke domain

### Phase 7: Domain Logic - SELESAI

Semua 12 domain packages telah dibuat dengan business logic lengkap:

- [x] 7.1 `tenant` - CRUD, slug unique validation
- [x] 7.2 `auth` - Staff+Customer login, bcrypt verify, JWT generation, refresh
- [x] 7.3 `staff` - CRUD, password hashing
- [x] 7.4 `nas` - CRUD, max 3 per tenant enforcement, password encryption, test connection
- [x] 7.5 `internet_package` - CRUD
- [x] 7.6 `customer` - CRUD, PPPoE credential management
- [x] 7.7 `subscription` - Lifecycle management, auto-sync ke MikroTik
- [x] 7.8 `payment_method` - CRUD
- [x] 7.9 `invoice` - Generation, bulk generate, status tracking
- [x] 7.10 `payment` - Processing, verify/reject, update invoice status
- [x] 7.11 `mikrotik` - Orchestration sync subscription↔MikroTik, bandwidth monitoring
- [x] `registry.go` - Domain interface + constructor dengan httpPort

### Phase 8: Inbound Port Interfaces - SELESAI

Yang sudah selesai:
- [x] `middleware.go` - InternalAuth, ClientAuth, StaffAuth, CustomerAuth, RequirePermission
- [x] `auth.go` - StaffLogin, StaffRefresh, CustomerLogin, CustomerRefresh
- [x] `tenant.go` - Get, Update
- [x] `staff.go` - List, Create, Get, Update, Delete
- [x] `internet_package.go` - List, Create, Get, Update, Delete
- [x] `subscription.go` - List, Create, Get, Update, Suspend, Activate, Cancel, SetVacation
- [x] `payment_method.go` - List, Create, Get, Update, Delete
- [x] `invoice.go` - List, Create, Get, Update, GenerateBulk
- [x] `payment.go` - List, Get, Verify, Reject
- [x] `portal.go` - Dashboard, GetProfile, UpdateProfile, UpdatePassword, GetSubscription, GetConnectionStatus, GetBandwidth, ListInvoices, GetInvoice, ListPayments, CreatePayment, GetPayment, UploadProof, ListPackages, ListPaymentMethods
- [x] `mikrotik.go` - Semua MikroTik operations
- [x] `registry_http.go` - Semua entity accessors

Selesai (ditambahkan pada 2026-02-16):
- [x] **8.1** `nas.go` - Ditambah method `GetIdentity`
- [x] **8.2** `customer.go` - Ditambah method `GetConnectionStatus`
- [x] **8.3** `portal.go` - Ditambah methods: `GetSubscriptionHistory`, `UpgradeSubscription`, `DowngradeSubscription`, `RequestVacation`
- [x] **8.4** `tenant.go` - Ditambah method `Upsert`

### Phase 9: Inbound Gin HTTP Adapters - SELESAI

Yang sudah selesai:
- [x] `auth.go`, `tenant.go`, `staff.go`, `internet_package.go`
- [x] `customer.go`, `subscription.go`, `payment_method.go`
- [x] `invoice.go`, `payment.go`, `portal.go`, `mikrotik.go`
- [x] `middleware.go` - Semua auth middleware + RBAC
- [x] `registry.go` - Semua adapter registrations
- [x] `route.go` - Auth, Internal, Client, Staff Admin API, MikroTik, Portal (sebagian besar)

Selesai (ditambahkan pada 2026-02-16):
- [x] **9.1** Route `GET /api/v1/nas/:id/identity` - Ditambahkan
- [x] **9.2** Route `GET /api/v1/customers/:id/connection` - Ditambahkan
- [x] **9.3** Route `POST /internal/tenant-upsert` - Ditambahkan
- [x] **9.4** Portal routes ditambahkan:
  - `GET /portal/v1/subscription/history`
  - `POST /portal/v1/subscription/upgrade`
  - `POST /portal/v1/subscription/downgrade`
  - `POST /portal/v1/subscription/vacation`
- [x] **9.5** Gin adapter `nas.go` - Handler `GetIdentity` diimplementasi
- [x] **9.6** Gin adapter `customer.go` - Handler `GetConnectionStatus` diimplementasi
- [x] **9.7** Gin adapter `portal.go` - Handlers subscription history/upgrade/downgrade/vacation diimplementasi
- [x] **9.8** Gin adapter `tenant.go` - Handler `Upsert` diimplementasi
- [x] **9.9** Domain `nas/domain.go` - Method `GetIdentity` ditambahkan ke NasDomain interface + implementasi

### Phase 10: Tests - SELESAI

Semua test files yang di-plan sudah ada:

Mock Generation:
- [x] 34 mock files di `tests/mocks/` untuk semua ports

Domain Unit Tests (12 packages):
- [x] `auth/domain_test.go`, `client/domain_test.go`, `customer/domain_test.go`
- [x] `internet_package/domain_test.go`, `invoice/domain_test.go`, `mikrotik/domain_test.go`
- [x] `nas/domain_test.go`, `payment/domain_test.go`, `payment_method/domain_test.go`
- [x] `subscription/domain_test.go`, `staff/domain_test.go`, `tenant/domain_test.go`

HTTP Tests (12 files):
- [x] `tests/http/auth_test.go`, `client_test.go`, `customer_test.go`
- [x] `internet_package_test.go`, `invoice_test.go`, `mikrotik_test.go`
- [x] `nas_test.go`, `payment_test.go`, `payment_method_test.go`
- [x] `staff_test.go`, `subscription_test.go`, `tenant_test.go`

Integration Tests:
- [x] `tests/integration/client_integration_test.go`
- [x] `tests/integration/tenant_integration_test.go`
- [x] `tests/integration/billing_integration_test.go`

Test Fixtures:
- [x] `tests/fixtures/testdata.go`

Adapter Tests:
- [x] `internal/adapter/inbound/gin/client_test.go`
- [x] `internal/adapter/inbound/gin/middleware_test.go`
- [x] `internal/adapter/outbound/postgres/client_test.go`

### Phase 11: Configuration - SELESAI

- [x] `.env.example` - Lengkap dengan JWT_SECRET, ENCRYPTION_KEY, UPLOAD_PATH
- [x] `Makefile` - generate-mocks target updated
- [x] `Dockerfile` - Multi-stage build
- [x] `docker-compose.yml` - PostgreSQL, RabbitMQ, Redis

---

## PLANNING ULANG: Task Yang Belum Selesai

### Phase 8A: Melengkapi Inbound Port Interfaces

#### Task 8A.1: Tambah GetIdentity ke NasHttpPort
- File: `internal/port/inbound/nas.go`
- Tambah method `GetIdentity(a any) error` ke interface NasHttpPort
- Fungsi: Mendapatkan identity/hostname dari router MikroTik via NAS ID

#### Task 8A.2: Tambah GetConnectionStatus ke CustomerHttpPort
- File: `internal/port/inbound/customer.go`
- Tambah method `GetConnectionStatus(a any) error` ke interface CustomerHttpPort
- Fungsi: Melihat status koneksi customer (PPPoE/Hotspot aktif atau tidak)

#### Task 8A.3: Tambah Portal Subscription Methods ke PortalHttpPort
- File: `internal/port/inbound/portal.go`
- Tambah methods:
  - `GetSubscriptionHistory(a any) error` - Riwayat langganan customer
  - `UpgradeSubscription(a any) error` - Upgrade paket customer
  - `DowngradeSubscription(a any) error` - Downgrade paket customer
  - `RequestVacation(a any) error` - Request vacation mode

### Phase 9A: Melengkapi Gin HTTP Adapters + Routes

#### Task 9A.1: Implement NAS GetIdentity Handler
- File: `internal/adapter/inbound/gin/nas.go`
- Tambah handler `GetIdentity` yang:
  1. Ambil NAS by ID dari database
  2. Decrypt password NAS
  3. Call MikroTik GetIdentity via outbound port
  4. Return identity string

#### Task 9A.2: Implement Customer GetConnectionStatus Handler
- File: `internal/adapter/inbound/gin/customer.go`
- Tambah handler `GetConnectionStatus` yang:
  1. Ambil customer by ID
  2. Ambil NAS terkait
  3. Call MikroTik GetActiveConnections
  4. Filter connections berdasarkan customer PPPoE username
  5. Return connection status

#### Task 9A.3: Implement Portal Subscription Handlers
- File: `internal/adapter/inbound/gin/portal.go`
- Tambah handlers:
  - `GetSubscriptionHistory` - Query subscriptions by customer ID, ordered by date
  - `UpgradeSubscription` - Upgrade ke paket lebih tinggi, sync MikroTik
  - `DowngradeSubscription` - Downgrade ke paket lebih rendah, sync MikroTik
  - `RequestVacation` - Set vacation start/end, disable MikroTik secret

#### Task 9A.4: Tambah Missing Routes ke route.go
- File: `internal/adapter/inbound/gin/route.go`
- Tambah routes:
  ```
  nas.GET("/:id/identity", func(c *gin.Context) { port.Nas().GetIdentity(c) })

  customers.GET("/:id/connection", func(c *gin.Context) { port.Customer().GetConnectionStatus(c) })

  portal.GET("/subscription/history", func(c *gin.Context) { port.Portal().GetSubscriptionHistory(c) })
  portal.POST("/subscription/upgrade", func(c *gin.Context) { port.Portal().UpgradeSubscription(c) })
  portal.POST("/subscription/downgrade", func(c *gin.Context) { port.Portal().DowngradeSubscription(c) })
  portal.POST("/subscription/vacation", func(c *gin.Context) { port.Portal().RequestVacation(c) })
  ```

#### Task 9A.5: Tambah Internal Tenant Upsert Route
- File: `internal/adapter/inbound/gin/route.go`
- Tambah route di internal group:
  ```
  internal.POST("/tenant-upsert", func(c *gin.Context) { port.Tenant().Upsert(c) })
  ```
- Requires: Tambah `Upsert(a any) error` ke TenantHttpPort jika diperlukan

### Phase 10A: Update Tests untuk Fitur Baru

#### Task 10A.1: Update Mocks
- Jalankan `make generate-mocks` setelah port interface berubah
- Verifikasi mock files ter-generate untuk methods baru

#### Task 10A.2: Tambah HTTP Tests
- File: `tests/http/nas_test.go` - Tambah test untuk GET /api/v1/nas/:id/identity
- File: `tests/http/customer_test.go` - Tambah test untuk GET /api/v1/customers/:id/connection
- File: `tests/http/portal_test.go` (BARU) - Test portal subscription endpoints

#### Task 10A.3: Update Domain Tests jika ada logic baru
- Jika ada domain logic baru untuk subscription upgrade/downgrade/vacation, tambah test cases

---

## Urutan Eksekusi (Recommended)

### Batch 1: Port Interfaces (Phase 8A)
1. Task 8A.1 - NasHttpPort.GetIdentity
2. Task 8A.2 - CustomerHttpPort.GetConnectionStatus
3. Task 8A.3 - PortalHttpPort subscription methods

### Batch 2: Gin Adapters + Routes (Phase 9A)
4. Task 9A.1 - NAS GetIdentity handler
5. Task 9A.2 - Customer GetConnectionStatus handler
6. Task 9A.3 - Portal subscription handlers
7. Task 9A.4 - Missing routes di route.go
8. Task 9A.5 - Internal tenant-upsert route

### Batch 3: Tests (Phase 10A)
9. Task 10A.1 - Regenerate mocks
10. Task 10A.2 - HTTP tests baru
11. Task 10A.3 - Update domain tests

### Batch 4: Verification
12. `go build ./...` - Compile check
13. `make test` - Semua unit tests pass
14. `make http` - Server start tanpa error
15. Manual test endpoints baru via curl/Postman

---

## Database Schema Overview (Tidak Berubah)

```
tenants (root)
├── staffs (tenant_id FK, role_id FK → roles)
├── nas (tenant_id FK)
│   └── [MikroTik: PPPoE secrets, hotspot users, queues - external, not in DB]
├── internet_packages (tenant_id FK)
├── customers (tenant_id FK, nas_id FK → nas)
│   └── subscriptions (tenant_id FK, customer_id FK, package_id FK, nas_id FK)
│       └── invoices (tenant_id FK, customer_id FK, subscription_id FK)
│           └── payments (tenant_id FK, invoice_id FK, payment_method_id FK, verified_by FK → staffs)
├── payment_methods (tenant_id FK)
└── clients (existing)

roles (global, not tenant-scoped)
├── permissions (global)
└── role_permissions (join table)
```

---

## RBAC Permission Mapping (Tidak Berubah)

| Resource         | Admin  | Technician | Sales  | Finance |
|------------------|--------|------------|--------|---------|
| tenant           | manage | -          | -      | -       |
| staff            | manage | -          | -      | -       |
| nas              | manage | manage     | read   | -       |
| internet_package | manage | read       | read   | read    |
| customer         | manage | read       | manage | read    |
| subscription     | manage | read       | manage | read    |
| payment_method   | manage | -          | -      | manage  |
| invoice          | manage | -          | read   | manage  |
| payment          | manage | -          | read   | manage  |
| report           | manage | -          | -      | manage  |

---

## Catatan Tambahan

### Fitur yang Sudah Ada Tapi Tidak di Plan Awal
- `internal/adapter/inbound/rabbitmq/` - RabbitMQ consumer untuk client upsert (skeleton)
- `internal/adapter/inbound/command/` - CLI command handlers untuk client (skeleton)
- `internal/model/variant.go` - Variant model
- Temporal workflow integration structure

### Dependency Versions (Current)
- Go >= 1.24.0
- `gorm.io/gorm v1.31.1`
- `gorm.io/driver/postgres v1.6.0`
- `github.com/gin-gonic/gin v1.11.0`
- `github.com/golang-jwt/jwt/v5 v5.3.0`
- `github.com/go-routeros/routeros/v3 v3.0.1`
- `github.com/redis/go-redis/v9 v9.8.0`
- `go.temporal.io/sdk v1.39.0`
