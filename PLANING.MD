 Phase 0: GORM Foundation

 Migrasi database layer dari Goqu + raw SQL ke GORM.

 0.1 Add GORM dependencies

 go get gorm.io/gorm gorm.io/driver/postgres

 0.2 Modify utils/database/database.go

 - Tambah fungsi InitGormDatabase(ctx, driver) *gorm.DB
 - GORM dibuka via gorm.Open(postgres.Open(dsn), &gorm.Config{})
 - Extract *sql.DB dari GORM untuk tetap jalankan Goose migrations
 - Set connection pool: MaxIdleConns=10, MaxOpenConns=100, ConnMaxLifetime=1h

 0.3 Modify internal/port/outbound/registry_database.go

 - Hapus DatabaseExecutor interface (raw SQL)
 - DatabasePort tetap sama pattern-nya, tapi internal adapter terima *gorm.DB
 - Tambah method WithTenantScope(tenantID string) DatabasePort untuk tenant filtering

 0.4 Modify internal/adapter/outbound/postgres/registry.go

 - Ganti *sql.DB + DatabaseExecutor → *gorm.DB
 - Transaction pakai db.Begin() / tx.Commit() / tx.Rollback() dari GORM
 - Method getDB() return tx jika dalam transaksi, atau db jika tidak
 - Implement WithTenantScope() yang return adapter baru dengan scoped DB

 0.5 Modify internal/adapter/outbound/postgres/client.go

 - Ganti semua Goqu query → GORM operations
 - Upsert: db.Clauses(clause.OnConflict{...}).Create(&clients)
 - FindByFilter: db.Where(...).Find(&clients) dengan helper applyClientFilter()
 - DeleteByFilter: db.Where(...).Delete(&model.Client{})
 - IsExists: db.Where("bearer_key = ?", key).First(&client)

 0.6 Modify internal/app.go

 - databaseOutbound(): panggil database.InitGormDatabase() bukan database.InitDatabase()
 - postgres_outbound_adapter.NewAdapter(db) terima *gorm.DB

 0.7 Update tests

 - internal/adapter/outbound/postgres/client_test.go: Gunakan go-sqlmock via GORM adapter atau in-memory SQLite
 - tests/integration/client_integration_test.go: Init GORM dengan testcontainers PostgreSQL

 Files modified:
 - go.mod / go.sum
 - utils/database/database.go
 - internal/port/outbound/registry_database.go
 - internal/adapter/outbound/postgres/registry.go
 - internal/adapter/outbound/postgres/client.go
 - internal/adapter/outbound/postgres/client_test.go
 - internal/app.go
 - tests/integration/client_integration_test.go

 ---
 Phase 1: Shared Utilities

 1.1 Create utils/gorm/scope.go

 func TenantScope(tenantID string) func(db *gorm.DB) *gorm.DB
 Menambahkan WHERE tenant_id = ? ke semua query secara otomatis.

 1.2 Create utils/crypto/encryption.go

 - AES-GCM encryption/decryption untuk password NAS MikroTik
 - Key dari env var ENCRYPTION_KEY (32 bytes)
 - Encrypt(plaintext) (ciphertext, error) dan Decrypt(ciphertext) (plaintext, error)

 1.3 Create utils/jwt/signer.go

 - Menggunakan github.com/golang-jwt/jwt/v5
 - GenerateStaffToken(staffID, tenantID, role, permissions, secret) (string, error)
 - GenerateCustomerToken(customerID, tenantID, secret) (string, error)
 - ParseStaffToken(tokenString, secret) (*StaffClaims, error)
 - ParseCustomerToken(tokenString, secret) (*CustomerClaims, error)
 - Claims struct: StaffClaims{StaffID, TenantID, Role, Permissions}, CustomerClaims{CustomerID, TenantID}

 1.4 Modify utils/jwt/validator.go

 - Tambah HMAC validation untuk self-issued tokens (selain JWKS yang sudah ada)

 1.5 Modify utils/activity/activity.go

 - Tambah context keys: TenantID, StaffID, CustomerID, UserRole, UserPermissions
 - Helper functions: WithTenantID(), GetTenantID(), WithStaffID(), GetStaffID(), dll.

 1.6 Modify .env.example

 Tambah:
 JWT_SECRET=your-secret-key-min-32-chars
 ENCRYPTION_KEY=your-32-byte-encryption-key
 UPLOAD_PATH=./uploads

 Files created: utils/gorm/scope.go, utils/crypto/encryption.go, utils/jwt/signer.go
 Files modified: utils/jwt/validator.go, utils/activity/activity.go, .env.example

 ---
 Phase 2: Models

 Buat model files menggunakan make model VAL=name lalu customize sesuai kebutuhan.

 2.1 internal/model/tenant.go

 Tenant{ID(uuid), TenantInput}
 TenantInput{Name, Slug(unique), Email, Phone, Address, LogoURL, MaxNas(default:3),
             SubscriptionPlan(default:'basic'), SubscriptionExpiresAt, IsActive, CreatedAt, UpdatedAt}
 TenantFilter{IDs, Slugs, Emails}

 2.2 internal/model/role.go

 Role{ID(serial), RoleInput}
 RoleInput{Name(unique), Description, CreatedAt, UpdatedAt}
 RoleFilter{IDs, Names}

 2.3 internal/model/permission.go

 Permission{ID(serial), PermissionInput}
 PermissionInput{Name(unique), Resource, Action, Description, CreatedAt, UpdatedAt}
 PermissionFilter{IDs, Resources, Actions}
 RolePermission{RoleID, PermissionID}  // join table

 // Constants:
 const (RoleAdmin="admin", RoleTechnician="technician", RoleSales="sales", RoleFinance="finance")
 const (ResourceNas, ResourceCustomer, ResourceSubscription, ResourceInvoice, ...)
 const (ActionCreate, ActionRead, ActionUpdate, ActionDelete, ActionManage)

 2.4 internal/model/staff.go

 Staff{ID(uuid), StaffInput}
 StaffInput{TenantID, RoleID, Email, PasswordHash, FullName, Phone, IsActive, LastLoginAt, CreatedAt, UpdatedAt}
 // Password field (gorm:"-") untuk input only
 StaffFilter{IDs, TenantIDs, Emails, RoleIDs}

 2.5 internal/model/nas.go

 Nas{ID(uuid), NasInput}
 NasInput{TenantID, Name, Host, ApiPort(8728), RestPort(80), Username, PasswordEncrypted,
          UseSSL, RouterOsVersion, Identity, IsActive, LastSeenAt, CreatedAt, UpdatedAt}
 NasFilter{IDs, TenantIDs, Hosts}

 2.6 internal/model/internet_package.go

 InternetPackage{ID(uuid), InternetPackageInput}
 InternetPackageInput{TenantID, Name, Description, Type(pppoe/hotspot/static),
                      UploadRate, DownloadRate, UploadBurst, DownloadBurst,
                      Price(int64), BillingCycle(daily/weekly/monthly/yearly),
                      ValidityDays, IsActive, CreatedAt, UpdatedAt}
 InternetPackageFilter{IDs, TenantIDs, Types, Names}

 2.7 internal/model/customer.go

 Customer{ID(uuid), CustomerInput}
 CustomerInput{TenantID, FullName, Email, Phone, Address, IdentityNumber,
               Username, PasswordHash, PppoeUsername, PppoePassword,
               StaticIP, NasID, IsActive, RegisteredAt, CreatedAt, UpdatedAt}
 CustomerFilter{IDs, TenantIDs, Emails, Usernames, PppoeUsernames, NasIDs}

 2.8 internal/model/subscription.go

 Subscription{ID(uuid), SubscriptionInput}
 SubscriptionInput{TenantID, CustomerID, PackageID, NasID, Status(active/suspended/expired/cancelled/vacation),
                   StartDate, EndDate, AutoRenew, VacationStart, VacationEnd,
                   MikrotikQueueName, MikrotikSecretName, CreatedAt, UpdatedAt}
 SubscriptionFilter{IDs, TenantIDs, CustomerIDs, PackageIDs, Statuses}

 2.9 internal/model/payment_method.go

 PaymentMethod{ID(uuid), PaymentMethodInput}
 PaymentMethodInput{TenantID, Name, Type(bank_transfer/e_wallet/cash/qris/other),
                    AccountName, AccountNumber, BankName, Instructions, IsActive, CreatedAt, UpdatedAt}
 PaymentMethodFilter{IDs, TenantIDs, Types}

 2.10 internal/model/invoice.go

 Invoice{ID(uuid), InvoiceInput}
 InvoiceInput{TenantID, CustomerID, SubscriptionID, InvoiceNumber, Amount, TaxAmount, TotalAmount,
              Status(unpaid/paid/overdue/cancelled/partial), DueDate, PeriodStart, PeriodEnd,
              Notes, PaidAt, CreatedAt, UpdatedAt}
 InvoiceFilter{IDs, TenantIDs, CustomerIDs, SubscriptionIDs, Statuses, InvoiceNumbers}

 2.11 internal/model/payment.go

 Payment{ID(uuid), PaymentInput}
 PaymentInput{TenantID, InvoiceID, PaymentMethodID, Amount, PaymentDate, ProofURL,
              Status(pending/verified/rejected), VerifiedBy, VerifiedAt, Notes, CreatedAt, UpdatedAt}
 PaymentFilter{IDs, TenantIDs, InvoiceIDs, PaymentMethodIDs, Statuses}

 2.12 internal/model/mikrotik.go

 Non-persisted structs untuk MikroTik API responses:
 - MikrotikConnection{ID, Name, Service, CallerID, Address, Uptime}
 - MikrotikTraffic{Interface, RxBytes, TxBytes, RxPackets, TxPackets}
 - MikrotikBandwidth{Target, Upload, Download}
 - MikrotikPPPoESecret{ID, Name, Password, Service, Profile, Disabled, Comment}
 - MikrotikPPPoESecretInput{Name, Password, Service, Profile, Comment}
 - MikrotikHotspotUser, MikrotikHotspotUserInput
 - MikrotikSimpleQueue, MikrotikSimpleQueueInput
 - MikrotikProfile, MikrotikProfileInput

 2.13 internal/model/auth.go

 - StaffAuthRequest{Email, Password}
 - CustomerAuthRequest{Username, Password, TenantSlug}
 - AuthTokenResponse{AccessToken, RefreshToken, ExpiresIn, TokenType}
 - StaffJWTClaims{StaffID, TenantID, Role, Permissions}
 - CustomerJWTClaims{CustomerID, TenantID}

 ---
 Phase 3: Migrations

 Menggunakan make migration-postgres VAL=name lalu customize SQL. Urutan berdasarkan foreign key dependencies.
 ┌─────┬──────────────────────────────┬───────────────────┬────────────────────────────────────────────┐
 │  #  │             File             │       Table       │              FK Dependencies               │
 ├─────┼──────────────────────────────┼───────────────────┼────────────────────────────────────────────┤
 │ 2   │ 2_tenant.go                  │ tenants           │ -                                          │
 ├─────┼──────────────────────────────┼───────────────────┼────────────────────────────────────────────┤
 │ 3   │ 3_role.go                    │ roles             │ -                                          │
 ├─────┼──────────────────────────────┼───────────────────┼────────────────────────────────────────────┤
 │ 4   │ 4_permission.go              │ permissions       │ -                                          │
 ├─────┼──────────────────────────────┼───────────────────┼────────────────────────────────────────────┤
 │ 5   │ 5_role_permission.go         │ role_permissions  │ roles, permissions                         │
 ├─────┼──────────────────────────────┼───────────────────┼────────────────────────────────────────────┤
 │ 6   │ 6_staff.go                   │ staffs            │ tenants, roles                             │
 ├─────┼──────────────────────────────┼───────────────────┼────────────────────────────────────────────┤
 │ 7   │ 7_nas.go                     │ nas               │ tenants                                    │
 ├─────┼──────────────────────────────┼───────────────────┼────────────────────────────────────────────┤
 │ 8   │ 8_internet_package.go        │ internet_packages │ tenants                                    │
 ├─────┼──────────────────────────────┼───────────────────┼────────────────────────────────────────────┤
 │ 9   │ 9_customer.go                │ customers         │ tenants, nas                               │
 ├─────┼──────────────────────────────┼───────────────────┼────────────────────────────────────────────┤
 │ 10  │ 10_subscription.go           │ subscriptions     │ tenants, customers, internet_packages, nas │
 ├─────┼──────────────────────────────┼───────────────────┼────────────────────────────────────────────┤
 │ 11  │ 11_payment_method.go         │ payment_methods   │ tenants                                    │
 ├─────┼──────────────────────────────┼───────────────────┼────────────────────────────────────────────┤
 │ 12  │ 12_invoice.go                │ invoices          │ tenants, customers, subscriptions          │
 ├─────┼──────────────────────────────┼───────────────────┼────────────────────────────────────────────┤
 │ 13  │ 13_payment.go                │ payments          │ tenants, invoices, payment_methods, staffs │
 ├─────┼──────────────────────────────┼───────────────────┼────────────────────────────────────────────┤
 │ 14  │ 14_seed_roles_permissions.go │ Seed data         │ roles, permissions, role_permissions       │
 └─────┴──────────────────────────────┴───────────────────┴────────────────────────────────────────────┘
 Semua tenant-scoped tables memiliki tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE dan CREATE INDEX idx_{table}_tenant_id ON {table}(tenant_id).

 ---
 Phase 4: Outbound Port Interfaces

 Menggunakan make outbound-database-postgres VAL=name untuk generate port+adapter skeleton.

 Database Ports (per entity)

 Buat file di internal/port/outbound/ untuk setiap entity:
 - tenant.go: TenantDatabasePort{Create, FindByFilter, FindByID, Update, Delete}
 - role.go: RoleDatabasePort{FindAll, FindByID, FindByName}
 - permission.go: PermissionDatabasePort{FindByRoleID, FindAll}
 - staff.go: StaffDatabasePort{Create, FindByFilter, FindByID, FindByEmail, Update, Delete}
 - nas.go: NasDatabasePort{Create, FindByFilter, FindByID, Update, Delete, CountByTenantID}
 - internet_package.go: InternetPackageDatabasePort{Create, FindByFilter, FindByID, Update, Delete}
 - customer.go: CustomerDatabasePort{Create, FindByFilter, FindByID, FindByUsername, Update, Delete}
 - subscription.go: SubscriptionDatabasePort{Create, FindByFilter, FindByID, Update, FindExpiring}
 - payment_method.go: PaymentMethodDatabasePort{Create, FindByFilter, FindByID, Update, Delete}
 - invoice.go: InvoiceDatabasePort{Create, FindByFilter, FindByID, Update, BulkCreate}
 - payment.go: PaymentDatabasePort{Create, FindByFilter, FindByID, Update}

 Update Registry Files

 - registry_database.go: Tambah semua entity accessors ke DatabasePort interface + WithTenantScope()
 - registry_cache.go: Tambah Staff(), Customer()
 - registry_http.go: Tambah Mikrotik() MikrotikPort

 MikroTik Port

 - mikrotik.go: Interface MikrotikPort dengan semua operasi RouterOS API + REST API

 ---
 Phase 5: Outbound Database Adapters

 Menggunakan make outbound-database-postgres VAL=name lalu implement GORM queries.

 Buat file di internal/adapter/outbound/postgres/ untuk setiap entity:
 tenant.go, role.go, permission.go, staff.go, nas.go, internet_package.go,
 customer.go, subscription.go, payment_method.go, invoice.go, payment.go

 Setiap adapter menerima *gorm.DB (yang sudah di-scope tenant_id jika perlu) dan implement GORM CRUD operations.

 Update registry.go untuk return semua adapter baru.

 ---
 Phase 6: Outbound Cache + MikroTik Adapters

 Cache Adapters

 - internal/adapter/outbound/redis/staff.go: Session caching untuk staff
 - internal/adapter/outbound/redis/customer.go: Session caching untuk customer
 - Update internal/adapter/outbound/redis/registry.go

 MikroTik Adapter (NEW directory)

 - internal/adapter/outbound/mikrotik/rest_client.go: HTTP client untuk MikroTik REST API
   - Base URL: http(s)://{host}:{rest_port}/rest
   - Basic Auth: username + decrypted password
   - Methods: CRUD PPPoE secrets, hotspot users, simple queues, profiles
 - internal/adapter/outbound/mikrotik/routeros_client.go: RouterOS API client
   - Library: github.com/go-routeros/routeros/v3
   - Connect via routeros.DialTimeout(addr, user, pass, timeout)
   - Methods: GetActiveConnections, GetInterfaceTraffic, MonitorBandwidth (ListenArgs)
 - internal/adapter/outbound/mikrotik/adapter.go: Main adapter implementing MikrotikPort
   - Delegates ke rest_client atau routeros_client sesuai operasi
 - internal/adapter/outbound/mikrotik/registry.go: Registry

 Wiring di internal/app.go

 - Tambah httpOutbound() function yang return MikroTik adapter
 - Inject ke domain sebagai parameter baru atau lewat existing httpPort

 ---
 Phase 7: Domain Logic

 Menggunakan make domain VAL=name untuk generate skeleton, lalu implement business logic.
 ┌──────┬──────────────────┬────────────────────────────────────────────┬────────────────────────────────────────────────────────────────────────────────┐
 │  #   │      Domain      │                    File                    │                                   Key Logic                                    │
 ├──────┼──────────────────┼────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ 7.1  │ tenant           │ internal/domain/tenant/domain.go           │ CRUD tenant, validasi slug unique                                              │
 ├──────┼──────────────────┼────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ 7.2  │ auth             │ internal/domain/auth/domain.go             │ Staff+Customer login, bcrypt verify, JWT generation, refresh token             │
 ├──────┼──────────────────┼────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ 7.3  │ staff            │ internal/domain/staff/domain.go            │ Staff CRUD, password hashing (bcrypt)                                          │
 ├──────┼──────────────────┼────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ 7.4  │ nas              │ internal/domain/nas/domain.go              │ NAS CRUD, enforce max 3 per tenant, encrypt password, test connection          │
 ├──────┼──────────────────┼────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ 7.5  │ internet_package │ internal/domain/internet_package/domain.go │ Package CRUD                                                                   │
 ├──────┼──────────────────┼────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ 7.6  │ customer         │ internal/domain/customer/domain.go         │ Customer CRUD, PPPoE credential management                                     │
 ├──────┼──────────────────┼────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ 7.7  │ subscription     │ internal/domain/subscription/domain.go     │ Lifecycle: create→active→suspend/vacation→expire→cancel, auto-sync ke MikroTik │
 ├──────┼──────────────────┼────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ 7.8  │ payment_method   │ internal/domain/payment_method/domain.go   │ Payment method CRUD                                                            │
 ├──────┼──────────────────┼────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ 7.9  │ invoice          │ internal/domain/invoice/domain.go          │ Invoice generation, bulk generate per period, status tracking                  │
 ├──────┼──────────────────┼────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ 7.10 │ payment          │ internal/domain/payment/domain.go          │ Payment processing, verify/reject, update invoice status                       │
 ├──────┼──────────────────┼────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ 7.11 │ mikrotik         │ internal/domain/mikrotik/domain.go         │ Orchestration: sync subscription↔MikroTik, bandwidth monitoring                │
 └──────┴──────────────────┴────────────────────────────────────────────┴────────────────────────────────────────────────────────────────────────────────┘
 Update internal/domain/registry.go:
 type Domain interface {
     Client() client.ClientDomain
     Tenant() tenant.TenantDomain
     Auth() auth.AuthDomain
     Staff() staff.StaffDomain
     Nas() nas.NasDomain
     InternetPackage() internet_package.InternetPackageDomain
     Customer() customer.CustomerDomain
     Subscription() subscription.SubscriptionDomain
     PaymentMethod() payment_method.PaymentMethodDomain
     Invoice() invoice.InvoiceDomain
     Payment() payment.PaymentDomain
     Mikrotik() mikrotik.MikrotikDomain
 }
 Domain constructor receives tambahan httpPort outbound_port.HttpPort untuk MikroTik access.

 ---
 Phase 8: Inbound Port Interfaces

 Menggunakan make inbound-http-gin VAL=name untuk generate skeleton.

 Middleware (modify internal/port/inbound/middleware.go)

 type MiddlewareHttpPort interface {
     InternalAuth(a any) error      // existing
     ClientAuth(a any) error        // existing
     StaffAuth(a any) error         // NEW - JWT → extract tenant_id, role, permissions
     CustomerAuth(a any) error      // NEW - JWT → extract customer_id, tenant_id
     RequirePermission(resource, action string) gin.HandlerFunc  // NEW - RBAC check
 }

 New Port Files di internal/port/inbound/

 - auth.go: AuthHttpPort{StaffLogin, StaffRefresh, CustomerLogin, CustomerRefresh}
 - tenant.go: TenantHttpPort{Get, Update}
 - staff.go: StaffHttpPort{List, Create, Get, Update, Delete}
 - nas.go: NasHttpPort{List, Create, Get, Update, Delete, TestConnection, GetIdentity}
 - internet_package.go: InternetPackageHttpPort{List, Create, Get, Update, Delete}
 - customer.go: CustomerHttpPort{List, Create, Get, Update, Delete, GetConnectionStatus}
 - subscription.go: SubscriptionHttpPort{List, Create, Get, Update, Suspend, Activate, Cancel}
 - payment_method.go: PaymentMethodHttpPort{List, Create, Get, Update, Delete}
 - invoice.go: InvoiceHttpPort{List, Create, Get, Update, GenerateBulk}
 - payment.go: PaymentHttpPort{List, Get, Verify, Reject}
 - portal.go: PortalHttpPort{Dashboard, GetProfile, UpdateProfile, ...} (customer self-service)
 - MikroTik operations: MikrotikHttpPort{ListPPPoESecrets, CreatePPPoESecret, ...}

 Update internal/port/inbound/registry_http.go

 type HttpPort interface {
     Middleware() MiddlewareHttpPort
     Ping() PingHttpPort
     Client() ClientHttpPort
     Auth() AuthHttpPort
     Tenant() TenantHttpPort
     Staff() StaffHttpPort
     Nas() NasHttpPort
     InternetPackage() InternetPackageHttpPort
     Customer() CustomerHttpPort
     Subscription() SubscriptionHttpPort
     PaymentMethod() PaymentMethodHttpPort
     Invoice() InvoiceHttpPort
     Payment() PaymentHttpPort
     Portal() PortalHttpPort
     Mikrotik() MikrotikHttpPort
 }

 ---
 Phase 9: Inbound Gin HTTP Adapters + Routes

 Adapter Files di internal/adapter/inbound/gin/

 Buat handler files: auth.go, tenant.go, staff.go, nas.go, internet_package.go,
 customer.go, subscription.go, payment_method.go, invoice.go, payment.go,
 portal.go, mikrotik.go

 Modify: middleware.go, registry.go, route.go

 Route Structure (route.go)

 # Auth (no middleware)
 POST   /auth/staff/login
 POST   /auth/staff/refresh
 POST   /auth/customer/login
 POST   /auth/customer/refresh

 # Internal (InternalAuth - existing)
 POST   /internal/client-upsert
 POST   /internal/client-find
 DELETE /internal/client-delete
 POST   /internal/tenant-upsert

 # Staff Admin API (StaffAuth + RequirePermission)
 GET/PUT          /api/v1/tenant

 GET/POST         /api/v1/staffs
 GET/PUT/DELETE   /api/v1/staffs/:id

 GET/POST         /api/v1/nas
 GET/PUT/DELETE   /api/v1/nas/:id
 POST             /api/v1/nas/:id/test
 GET              /api/v1/nas/:id/identity

 # MikroTik Operations via NAS
 GET/POST         /api/v1/nas/:id/pppoe/secrets
 PUT/DELETE       /api/v1/nas/:id/pppoe/secrets/:sid
 POST             /api/v1/nas/:id/pppoe/secrets/:sid/disable
 POST             /api/v1/nas/:id/pppoe/secrets/:sid/enable
 GET/POST         /api/v1/nas/:id/pppoe/profiles
 GET/POST         /api/v1/nas/:id/hotspot/users
 PUT/DELETE       /api/v1/nas/:id/hotspot/users/:uid
 GET/POST         /api/v1/nas/:id/queues
 PUT/DELETE       /api/v1/nas/:id/queues/:qid
 GET              /api/v1/nas/:id/connections
 GET              /api/v1/nas/:id/traffic/:interface

 GET/POST         /api/v1/packages
 GET/PUT/DELETE   /api/v1/packages/:id

 GET/POST         /api/v1/customers
 GET/PUT/DELETE   /api/v1/customers/:id
 GET              /api/v1/customers/:id/connection

 GET/POST         /api/v1/subscriptions
 GET/PUT          /api/v1/subscriptions/:id
 POST             /api/v1/subscriptions/:id/suspend
 POST             /api/v1/subscriptions/:id/activate
 POST             /api/v1/subscriptions/:id/cancel

 GET/POST         /api/v1/payment-methods
 GET/PUT/DELETE   /api/v1/payment-methods/:id

 GET/POST         /api/v1/invoices
 GET/PUT          /api/v1/invoices/:id
 POST             /api/v1/invoices/generate

 GET              /api/v1/payments
 GET              /api/v1/payments/:id
 POST             /api/v1/payments/:id/verify
 POST             /api/v1/payments/:id/reject

 # Customer Portal (CustomerAuth)
 GET    /portal/v1/dashboard
 GET    /portal/v1/profile
 PUT    /portal/v1/profile
 PUT    /portal/v1/profile/password
 GET    /portal/v1/subscription
 GET    /portal/v1/subscription/history
 POST   /portal/v1/subscription/upgrade
 POST   /portal/v1/subscription/downgrade
 POST   /portal/v1/subscription/vacation
 GET    /portal/v1/connection/status
 GET    /portal/v1/connection/bandwidth
 GET    /portal/v1/invoices
 GET    /portal/v1/invoices/:id
 GET    /portal/v1/payments
 POST   /portal/v1/payments
 GET    /portal/v1/payments/:id
 POST   /portal/v1/payments/:id/upload-proof
 GET    /portal/v1/packages
 GET    /portal/v1/payment-methods

 RBAC Permission Mapping per Role
 ┌──────────────────┬────────┬────────────┬────────┬─────────┐
 │     Resource     │ Admin  │ Technician │ Sales  │ Finance │
 ├──────────────────┼────────┼────────────┼────────┼─────────┤
 │ tenant           │ manage │ -          │ -      │ -       │
 ├──────────────────┼────────┼────────────┼────────┼─────────┤
 │ staff            │ manage │ -          │ -      │ -       │
 ├──────────────────┼────────┼────────────┼────────┼─────────┤
 │ nas              │ manage │ manage     │ read   │ -       │
 ├──────────────────┼────────┼────────────┼────────┼─────────┤
 │ internet_package │ manage │ read       │ read   │ read    │
 ├──────────────────┼────────┼────────────┼────────┼─────────┤
 │ customer         │ manage │ read       │ manage │ read    │
 ├──────────────────┼────────┼────────────┼────────┼─────────┤
 │ subscription     │ manage │ read       │ manage │ read    │
 ├──────────────────┼────────┼────────────┼────────┼─────────┤
 │ payment_method   │ manage │ -          │ -      │ manage  │
 ├──────────────────┼────────┼────────────┼────────┼─────────┤
 │ invoice          │ manage │ -          │ read   │ manage  │
 ├──────────────────┼────────┼────────────┼────────┼─────────┤
 │ payment          │ manage │ -          │ read   │ manage  │
 ├──────────────────┼────────┼────────────┼────────┼─────────┤
 │ report           │ manage │ -          │ -      │ manage  │
 └──────────────────┴────────┴────────────┴────────┴─────────┘
 ---
 Phase 10: Tests

 Mock Generation

 make generate-mocks

 Domain Unit Tests

 Buat domain_test.go di setiap domain package mengikuti pattern GoConvey + gomock yang ada di internal/domain/client/domain_test.go.

 Test Fixtures

 Update tests/fixtures/testdata.go dengan fixtures untuk semua entity baru.

 Integration Tests

 - tests/integration/tenant_integration_test.go
 - tests/integration/billing_integration_test.go (end-to-end billing flow)

 ---
 Phase 11: Configuration

 Update .env.example

 # Existing
 OUTBOUND_DATABASE_DRIVER=postgres
 OUTBOUND_MESSAGE_DRIVER=rabbitmq
 OUTBOUND_CACHE_DRIVER=redis
 INBOUND_HTTP_DRIVER=gin
 DATABASE_HOST=localhost
 ...

 # New
 JWT_SECRET=your-secret-key-minimum-32-characters
 ENCRYPTION_KEY=your-32-byte-aes-encryption-key
 UPLOAD_PATH=./uploads

 Update Makefile

 - Update generate-mocks target untuk include semua port baru

 ---
 Database Schema Overview

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
 └── clients (existing, will add tenant_id later)

 roles (global, not tenant-scoped)
 ├── permissions (global)
 └── role_permissions (join table)

 ---
 Verification Plan

 1. Phase 0 verification: make test - semua existing tests masih pass setelah GORM migration
 2. Phase 1-2: go build ./... - compile check
 3. Phase 3: make migration-up - semua migrations berhasil
 4. Phase 4-6: make generate-mocks && go build ./... - compile check
 5. Phase 7: make test - semua domain unit tests pass
 6. Phase 8-9: make http - server start tanpa error, test endpoints via curl/Postman
 7. Phase 10: make test && make test-integration - semua tests pass
 8. End-to-end: Test flow lengkap:
   - Create tenant → Create staff → Staff login → Create NAS → Test MikroTik connection
   - Create package → Create customer → Create subscription → Auto-sync ke MikroTik
   - Generate invoice → Customer login portal → View invoice → Submit payment → Verify payment
   - Customer check bandwidth → Connection status

 ---
 Implementation Summary
 ┌───────┬───────────────┬────────────────┬──────────────────────┐
 │ Phase │ Files Created │ Files Modified │  Estimated Entities  │
 ├───────┼───────────────┼────────────────┼──────────────────────┤
 │ 0     │ 0             │ 8              │ 0 (GORM migration)   │
 ├───────┼───────────────┼────────────────┼──────────────────────┤
 │ 1     │ 3             │ 3              │ 0 (utilities)        │
 ├───────┼───────────────┼────────────────┼──────────────────────┤
 │ 2     │ 13            │ 0              │ 13 models            │
 ├───────┼───────────────┼────────────────┼──────────────────────┤
 │ 3     │ 13            │ 0              │ 13 migrations        │
 ├───────┼───────────────┼────────────────┼──────────────────────┤
 │ 4     │ 12            │ 3              │ 12 port interfaces   │
 ├───────┼───────────────┼────────────────┼──────────────────────┤
 │ 5     │ 11            │ 1              │ 11 DB adapters       │
 ├───────┼───────────────┼────────────────┼──────────────────────┤
 │ 6     │ 6             │ 2              │ 2 cache + 4 mikrotik │
 ├───────┼───────────────┼────────────────┼──────────────────────┤
 │ 7     │ 11            │ 1              │ 11 domain services   │
 ├───────┼───────────────┼────────────────┼──────────────────────┤
 │ 8     │ 12            │ 2              │ 12 inbound ports     │
 ├───────┼───────────────┼────────────────┼──────────────────────┤
 │ 9     │ 12            │ 3              │ 12 HTTP handlers     │
 ├───────┼───────────────┼────────────────┼──────────────────────┤
 │ 10    │ 15+           │ 1              │ tests                │
 ├───────┼───────────────┼────────────────┼──────────────────────┤
 │ 11    │ 0             │ 2              │ config               │
 ├───────┼───────────────┼────────────────┼──────────────────────┤
 │ Total │ ~108          │ ~26            │                      │
 └───────┴───────────────┴────────────────┴──────────────────────┘