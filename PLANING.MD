MikrOps Enhancement Plan: ISP Billing & Auto-Cutoff System

Context

MikrOps sudah memiliki fondasi yang solid (13 model, 14 migrasi, 12 domain service, routes lengkap admin+portal). Plan ini menambahkan fitur ISP billing yang belum ada: tenant settings, customer self-registration, PPPoE
account management terpisah, auto-cutoff/isolir, payment-triggered restore, dan audit logging. Hotspot voucher dan WebSocket monitoring dikeluarkan ke fase berikutnya.

---
BAGIAN 1: DATABASE DESIGN

1.1 Tabel Baru

A. tenant_settings (pengaturan billing per tenant)

- id              UUID (PK, auto)
- tenant_id       UUID (FK -> tenants, unique)
- cutoff_day      INT (1-28, default: 1) -- tanggal cutoff bulanan
- grace_period    INT (default: 3) -- hari toleransi sebelum isolir
- isolir_profile_name VARCHAR(100, default: 'ISOLIR_MIKROPS')
- wa_gateway_api  TEXT (nullable) -- URL WhatsApp gateway
- auto_approve_registration BOOL (default: false)
- default_ppp_password_type VARCHAR(20, default: 'random') -- random/custom
- default_ppp_password_length INT (default: 8)
- created_at      TIMESTAMP
- updated_at      TIMESTAMP

B. customer_registrations (self-registration workflow)

- id                       UUID (PK, auto)
- tenant_id                UUID (FK -> tenants, NOT NULL)
- customer_id              UUID (FK -> customers, nullable) -- filled saat approved
- full_name                VARCHAR(255, NOT NULL)
- email                    VARCHAR(255)
- phone                    VARCHAR(50)
- address                  TEXT
- requested_package_id     UUID (FK -> internet_packages, NOT NULL)
- requested_nas_id         UUID (FK -> nas, nullable)
- status                   VARCHAR(50, default: 'pending') -- pending/approved/rejected
- rejection_reason         TEXT (nullable)
- approved_by              UUID (FK -> staffs, nullable)
- approved_at              TIMESTAMP (nullable)
- created_at               TIMESTAMP
- updated_at               TIMESTAMP

C. pppoe_accounts (akun PPPoE terpisah dari customer)

- id                UUID (PK, auto)
- tenant_id         UUID (FK -> tenants, NOT NULL)
- customer_id       UUID (FK -> customers, NOT NULL)
- subscription_id   UUID (FK -> subscriptions, nullable) -- link ke subscription
- nas_id            UUID (FK -> nas, NOT NULL)
- package_id        UUID (FK -> internet_packages, NOT NULL)
- username          VARCHAR(100, NOT NULL) -- unique per tenant
- password_encrypted TEXT (NOT NULL) -- encrypted
- profile_name      VARCHAR(100) -- dari internet_packages, untuk MikroTik
- original_profile  VARCHAR(100) -- tracking saat isolir (profile sebelum diubah)
- status            VARCHAR(50, default: 'active') -- active/isolated/disabled
- sync_status       VARCHAR(50, default: 'pending') -- synced/pending/error
- last_sync_at      TIMESTAMP (nullable)
- created_at        TIMESTAMP
- updated_at        TIMESTAMP
INDEX: (tenant_id, username) UNIQUE

D. activity_logs (audit trail)

- id              UUID (PK, auto)
- tenant_id       UUID (FK -> tenants, NOT NULL)
- user_id         UUID (nullable) -- staff_id or customer_id
- user_type       VARCHAR(50) -- 'staff' / 'customer' / 'system'
- action          VARCHAR(100, NOT NULL) -- contoh: 'approve_registration', 'isolate_pppoe'
- resource_type   VARCHAR(100) -- contoh: 'customer_registrations', 'pppoe_accounts'
- resource_id     VARCHAR(100) -- UUID as string
- details         JSONB (nullable) -- detail tambahan
- ip_address      VARCHAR(45)
- created_at      TIMESTAMP
INDEX: (tenant_id, action)
INDEX: (tenant_id, resource_type, resource_id)

E. mikrotik_sync_logs (sync tracking)

- id                  UUID (PK, auto)
- tenant_id           UUID (FK -> tenants, NOT NULL)
- nas_id              UUID (FK -> nas, NOT NULL)
- action              VARCHAR(50, NOT NULL) -- create/update/delete/isolate/restore
- resource_type       VARCHAR(50, NOT NULL) -- ppp_secret/hotspot_user/profile/queue
- resource_identifier VARCHAR(255) -- username atau profile name
- status              VARCHAR(50, NOT NULL) -- success/failed
- error_message       TEXT (nullable)
- request_payload     JSONB (nullable) -- data yang dikirim
- created_at          TIMESTAMP
INDEX: (nas_id, resource_type)
INDEX: (tenant_id, created_at)

1.2 Modifikasi Tabel Existing

A. customers - Tambah field

+ auto_cutoff  BOOL (default: true) -- apakah customer otomatis dicutoff
Hapus field PPPoE dari customer (pppoe_username, pppoe_password) karena dipindah ke pppoe_accounts.
Catatan: Migrasi ini harus dijalankan SETELAH data PPPoE dimigrasikan ke tabel baru.

B. internet_packages - Tambah field

+ profile_name  VARCHAR(100) -- nama profile di MikroTik (contoh: '10M', '20M')
Field ini digunakan untuk mapping langsung ke MikroTik PPP profile name.

C. subscriptions - Tambah field

+ pppoe_account_id UUID (FK -> pppoe_accounts, nullable) -- link ke PPPoE account
Ini menghubungkan subscription dengan PPPoE account yang spesifik.

1.3 ERD Relationships (Updated)

tenants ─────────── tenant_settings (1:1)
    **│**

    **├── staffs**

    **├── nas**

    **├── internet\_packages**

    **├── payment\_methods**

    **│**

    **├── customers ──────── customer\_registrations (1:many)**

    **│      │**

    **│      ├── pppoe\_accounts (1:many)**

    **│      │      └── subscriptions (1:1, via pppoe\_account\_id)**

    **│      │**

    **│      ├── subscriptions (1:many)**

    **│      │      └── invoices (1:many)**

    **│      │             └── payments (1:many)**

    **│      │**

    **│      └── activity\_logs**

    **│**

    **├── activity\_logs**

    **└── mikrotik\_sync\_logs**




---
BAGIAN 2: MIGRATION FILES

Urutan migrasi baru (lanjutan dari migration 14):
┌─────┬───────────────────────────────────────────────┬─────────────────────────────────────────────────────┐
│ No  │                     File                      │                      Deskripsi                      │
├─────┼───────────────────────────────────────────────┼─────────────────────────────────────────────────────┤
│ 15  │ 15_tenant_setting.go                          │ Buat tabel tenant_settings                          │
├─────┼───────────────────────────────────────────────┼─────────────────────────────────────────────────────┤
│ 16  │ 16_customer_registration.go                   │ Buat tabel customer_registrations                   │
├─────┼───────────────────────────────────────────────┼─────────────────────────────────────────────────────┤
│ 17  │ 17_pppoe_account.go                           │ Buat tabel pppoe_accounts                           │
├─────┼───────────────────────────────────────────────┼─────────────────────────────────────────────────────┤
│ 18  │ 18_activity_log.go                            │ Buat tabel activity_logs                            │
├─────┼───────────────────────────────────────────────┼─────────────────────────────────────────────────────┤
│ 19  │ 19_mikrotik_sync_log.go                       │ Buat tabel mikrotik_sync_logs                       │
├─────┼───────────────────────────────────────────────┼─────────────────────────────────────────────────────┤
│ 20  │ 20_alter_customer_add_auto_cutoff.go          │ Tambah auto_cutoff ke customers                     │
├─────┼───────────────────────────────────────────────┼─────────────────────────────────────────────────────┤
│ 21  │ 21_alter_internet_package_add_profile_name.go │ Tambah profile_name ke internet_packages            │
├─────┼───────────────────────────────────────────────┼─────────────────────────────────────────────────────┤
│ 22  │ 22_alter_subscription_add_pppoe_account_id.go │ Tambah pppoe_account_id ke subscriptions            │
├─────┼───────────────────────────────────────────────┼─────────────────────────────────────────────────────┤
│ 23  │ 23_migrate_pppoe_data.go                      │ Migrasi data PPPoE dari customers ke pppoe_accounts │
├─────┼───────────────────────────────────────────────┼─────────────────────────────────────────────────────┤
│ 24  │ 24_alter_customer_remove_pppoe_fields.go      │ Hapus pppoe_username, pppoe_password dari customers │
└─────┴───────────────────────────────────────────────┴─────────────────────────────────────────────────────┘
Path: internal/migration/postgres/
Pattern: Mengikuti pola Goose migration yang sudah ada (lihat migration 1-14)

---
BAGIAN 3: MODELS (Go Structs)

3.1 Model Baru

internal/model/tenant_setting.go

type TenantSetting struct {
     **ID        string `json:"id" gorm:"primaryKey;type:uuid;default:gen\_random\_uuid()"`**

     **TenantSettingInput**

}

type TenantSettingInput struct {
     **TenantID                 string `json:"tenant\_id" gorm:"uniqueIndex;type:uuid;not null"`**

     **CutoffDay                int    `json:"cutoff\_day" gorm:"default:1"`**

     **GracePeriod              int    `json:"grace\_period" gorm:"default:3"`**

     **IsolirProfileName        string `json:"isolir\_profile\_name" gorm:"default:'ISOLIR\_MIKROPS'"`**

     **WaGatewayApi             string `json:"wa\_gateway\_api" gorm:"type:text"`**

     **AutoApproveRegistration  bool   `json:"auto\_approve\_registration" gorm:"default:false"`**

     **DefaultPppPasswordType   string `json:"default\_ppp\_password\_type" gorm:"default:'random'"`**

     **DefaultPppPasswordLength int    `json:"default\_ppp\_password\_length" gorm:"default:8"`**

     **CreatedAt                time.Time**

     **UpdatedAt                time.Time**

}

internal/model/customer_registration.go

type CustomerRegistration struct {
     **ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen\_random\_uuid()"`**

     **CustomerRegistrationInput**

     **Customer        \*Customer        `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`**

     **InternetPackage \*InternetPackage `json:"internet\_package,omitempty" gorm:"foreignKey:RequestedPackageID"`**

     **Nas             \*Nas             `json:"nas,omitempty" gorm:"foreignKey:RequestedNasID"`**

     **ApprovedByStaff \*Staff           `json:"approved\_by\_staff,omitempty" gorm:"foreignKey:ApprovedBy"`**

}

// Constants: RegistrationStatusPending, RegistrationStatusApproved, RegistrationStatusRejected

internal/model/pppoe_account.go

type PppoeAccount struct {
     **ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen\_random\_uuid()"`**

     **PppoeAccountInput**

     **Customer        \*Customer        `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`**

     **Nas             \*Nas             `json:"nas,omitempty" gorm:"foreignKey:NasID"`**

     **InternetPackage \*InternetPackage `json:"internet\_package,omitempty" gorm:"foreignKey:PackageID"`**

}

// Constants: PppoeStatusActive, PppoeStatusIsolated, PppoeStatusDisabled
// Constants: SyncStatusSynced, SyncStatusPending, SyncStatusError

internal/model/activity_log.go

type ActivityLog struct {
     **ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen\_random\_uuid()"`**

     **ActivityLogInput**

}

internal/model/mikrotik_sync_log.go

type MikrotikSyncLog struct {
     **ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen\_random\_uuid()"`**

     **MikrotikSyncLogInput**

}

3.2 Modifikasi Model Existing

internal/model/customer.go

- Tambah field AutoCutoff bool ke CustomerInput
- Hapus field PppoeUsername, PppoePassword (setelah migrasi data)
- Tambah relasi: PppoeAccounts []PppoeAccount (optional preload)

internal/model/internet_package.go

- Tambah field ProfileName string ke InternetPackageInput

internal/model/subscription.go

- Tambah field PppoeAccountID *string ke SubscriptionInput
- Tambah relasi: PppoeAccount *PppoeAccount (optional preload)

---
BAGIAN 4: BUSINESS FLOWS

Flow 1: Customer Self-Registration

Customer                    System                      Admin
    **│                          │                           │**

    **├─ Submit form ──────────> │                           │**

    **│                          ├─ Validate input           │**

    **│                          ├─ Insert customer\_         │**

    **│                          │  registrations (pending)  │**

    **│                          │                           │**

    **│                          ├─ Check auto\_approve ──────┤**

    **│                          │  setting                  │**

    **│                          │                           │**

    **│  \[auto\_approve=true]     │                           │**

    **│  <── Credentials ────────┤                           │**

    **│                          ├─ Create customer          │**

    **│                          ├─ Create pppoe\_account     │**

    **│                          ├─ Create subscription      │**

    **│                          ├─ Enqueue MikroTik sync    │**

    **│                          │                           │**

    **│  \[auto\_approve=false]    │                           │**

    **│                          ├── Notifikasi ────────────>│**

    **│                          │                           ├─ Review**

    **│                          │                           ├─ Approve/Reject**

    **│                          │<── Approval action ───────┤**

    **│                          │                           │**

    **│  \[approved]              │                           │**

    **│  <── Credentials ────────┤                           │**

    **│                          ├─ Same as auto\_approve     │**

    **│                          │                           │**

    **│  \[rejected]              │                           │**

    **│  <── Rejection email ────┤                           │**




Domain Methods Needed:
- CustomerRegistration.Submit(ctx, input) (registration, error) - public endpoint
- CustomerRegistration.List(ctx, filter) ([]registration, error) - admin
- CustomerRegistration.Approve(ctx, id, staffID) error - admin approval
- CustomerRegistration.Reject(ctx, id, staffID, reason) error - admin rejection

Approval Sub-steps (dalam domain):
1. Create customers record
2. Generate PPPoE credentials (sesuai tenant_settings)
3. Insert pppoe_accounts (sync_status: pending)
4. Create subscriptions linked to pppoe_account
5. Create initial invoices (first month)
6. Enqueue RabbitMQ job: mikrotik.ppp.create
7. Log ke activity_logs

Flow 2: Auto-Cutoff (Isolir)

Cron (daily 00:01)
    **│**

    **├─ Query tenants WHERE cutoff\_day = today**

    **│**

    **├─ For each tenant:**

    **│     ├─ Query customers WHERE:**

    **│     │    - auto\_cutoff = true**

    **│     │    - is\_active = true**

    **│     │    - Has unpaid invoice past due\_date + grace\_period**

    **│     │**

    **│     ├─ For each customer:**

    **│     │    ├─ Get pppoe\_accounts (status: active)**

    **│     │    ├─ Enqueue RabbitMQ: mikrotik.ppp.isolate**

    **│     │    │    payload: {pppoe\_account\_id, nas\_id, username, isolir\_profile}**

    **│     │    ├─ Update pppoe\_account:**

    **│     │    │    - original\_profile = current profile\_name**

    **│     │    │    - status = 'isolated'**

    **│     │    ├─ Update subscription: status = 'suspended'**

    **│     │    └─ Log ke activity\_logs + mikrotik\_sync\_logs**

    **│**

    **└─ Worker:**

          **├─ Connect ke MikroTik via API**

          **├─ Change profile → ISOLIR\_MIKROPS**

          **├─ Force disconnect active session**

          **├─ Update sync\_status = 'synced'**

          **└─ Log result**




Domain Methods Needed:
- Cutoff.RunDailyCutoff(ctx) error - called by cron/scheduler
- Cutoff.IsolateCustomer(ctx, customerID) error - manual isolate
- Cutoff.IsolatePppoeAccount(ctx, pppoeAccountID) error - isolate specific account

MikroTik Commands:
# Change profile to isolir
/ppp secret set [find name={username}] profile=ISOLIR_MIKROPS
# Force disconnect
/ppp active remove [find name={username}]

Flow 3: Payment → Restore

Customer/Admin
    **│**

    **├─ Record payment (existing flow)**

    **│**

    **├─ Admin verifies payment (existing)**

    **│     │**

    **│     ├─ Update invoice status → 'paid'**

    **│     ├─ Check: customer has isolated pppoe\_accounts?**

    **│     │**

    **│     ├─ \[has isolated accounts]**

    **│     │    ├─ For each isolated pppoe\_account:**

    **│     │    │    ├─ Enqueue RabbitMQ: mikrotik.ppp.restore**

    **│     │    │    │    payload: {pppoe\_account\_id, nas\_id, username, restore\_profile}**

    **│     │    │    ├─ Update pppoe\_account:**

    **│     │    │    │    - profile\_name = original\_profile**

    **│     │    │    │    - original\_profile = NULL**

    **│     │    │    │    - status = 'active'**

    **│     │    │    ├─ Update subscription: status = 'active'**

    **│     │    │    └─ Log ke activity\_logs**

    **│     │    │**

    **│     │    └─ Worker:**

    **│     │         ├─ Connect ke MikroTik**

    **│     │         ├─ Restore profile → original profile**

    **│     │         ├─ Force disconnect (reconnect with new profile)**

    **│     │         └─ Update sync\_status**

    **│     │**

    **│     └─ \[no isolated accounts]**

    **│          └─ Done (just mark invoice as paid)**




Domain Method Modification:
- Extend existing Payment.Verify() to trigger restore flow
- Add Cutoff.RestorePppoeAccount(ctx, pppoeAccountID) error

Flow 4: NAS Provisioning (Isolir Profile Setup)

Admin adds NAS
    **│**

    **├─ Create NAS record (existing)**

    **│**

    **├─ \[new] Provision isolir infrastructure:**

    **│    ├─ Check if ISOLIR profile exists on MikroTik**

    **│    ├─ Create if not exists:**

    **│    │    - /ppp profile add name=ISOLIR\_MIKROPS rate-limit=128k/128k**

    **│    └─ Log ke mikrotik\_sync\_logs**




Domain Method:
- Extend Nas.Create() to trigger isolir profile provisioning

---
BAGIAN 5: PORTS (Interfaces)

5.1 Outbound Ports Baru

internal/port/outbound/tenant_setting.go

type TenantSettingDatabasePort interface {
     **Create(data model.TenantSettingInput) (model.TenantSetting, error)**

     **FindByTenantID(tenantID string) (model.TenantSetting, error)**

     **Update(id string, data model.TenantSettingInput) error**

}

internal/port/outbound/customer_registration.go

type CustomerRegistrationDatabasePort interface {
     **Create(data model.CustomerRegistrationInput) (model.CustomerRegistration, error)**

     **FindByFilter(filter model.CustomerRegistrationFilter) (\[]model.CustomerRegistration, error)**

     **FindByID(id string) (model.CustomerRegistration, error)**

     **Update(id string, data model.CustomerRegistrationInput) error**

}

internal/port/outbound/pppoe_account.go

type PppoeAccountDatabasePort interface {
     **Create(data model.PppoeAccountInput) (model.PppoeAccount, error)**

     **FindByFilter(filter model.PppoeAccountFilter) (\[]model.PppoeAccount, error)**

     **FindByID(id string) (model.PppoeAccount, error)**

     **Update(id string, data model.PppoeAccountInput) error**

     **Delete(id string) error**

}

internal/port/outbound/activity_log.go

type ActivityLogDatabasePort interface {
     **Create(data model.ActivityLogInput) (model.ActivityLog, error)**

     **FindByFilter(filter model.ActivityLogFilter) (\[]model.ActivityLog, error)**

}

internal/port/outbound/mikrotik_sync_log.go

type MikrotikSyncLogDatabasePort interface {
     **Create(data model.MikrotikSyncLogInput) (model.MikrotikSyncLog, error)**

     **FindByFilter(filter model.MikrotikSyncLogFilter) (\[]model.MikrotikSyncLog, error)**

}

Update internal/port/outbound/registry_database.go

Tambah methods:
TenantSetting() TenantSettingDatabasePort
CustomerRegistration() CustomerRegistrationDatabasePort
PppoeAccount() PppoeAccountDatabasePort
ActivityLog() ActivityLogDatabasePort
MikrotikSyncLog() MikrotikSyncLogDatabasePort

5.2 Inbound Ports Baru

internal/port/inbound/tenant_setting.go

type TenantSettingHttpPort interface {
     **Get(a any) error**

     **Update(a any) error**

}

internal/port/inbound/customer_registration.go

type CustomerRegistrationHttpPort interface {
     **Submit(a any) error    // public endpoint (no auth)**

     **List(a any) error      // admin**

     **Get(a any) error       // admin**

     **Approve(a any) error   // admin**

     **Reject(a any) error    // admin**

}

internal/port/inbound/pppoe_account.go

type PppoeAccountHttpPort interface {
     **List(a any) error**

     **Get(a any) error**

     **Create(a any) error**

     **Update(a any) error**

     **Delete(a any) error**

     **Isolate(a any) error   // manual isolate**

     **Restore(a any) error   // manual restore**

}

Update internal/port/inbound/registry_http.go

Tambah methods:
TenantSetting() TenantSettingHttpPort
CustomerRegistration() CustomerRegistrationHttpPort
PppoeAccount() PppoeAccountHttpPort

---
BAGIAN 6: DOMAIN SERVICES

6.1 Domain Baru

internal/domain/tenant_setting/ (TenantSettingDomain)

- Get(ctx, tenantID) (TenantSetting, error)
- Update(ctx, tenantID, input) error
- GetOrCreateDefault(ctx, tenantID) (TenantSetting, error) -- auto-create default settings

internal/domain/customer_registration/ (CustomerRegistrationDomain)

- Submit(ctx, input) (CustomerRegistration, error) -- public registration
- List(ctx, filter) ([]CustomerRegistration, error)
- Get(ctx, id) (CustomerRegistration, error)
- Approve(ctx, id, staffID) error -- triggers full provisioning flow
- Reject(ctx, id, staffID, reason) error

Approve Flow Internal:
1. Validate registration exists & status=pending
2. Get tenant_settings
3. Create customer record (via existing Customer domain)
4. Generate PPPoE credentials based on tenant_settings
5. Create pppoe_account
6. Create subscription
7. Generate first invoice
8. Enqueue MikroTik sync via RabbitMQ
9. Send credentials notification
10. Update registration status=approved
11. Log activity

internal/domain/pppoe_account/ (PppoeAccountDomain)

- Create(ctx, input) (PppoeAccount, error)
- List(ctx, filter) ([]PppoeAccount, error)
- Get(ctx, id) (PppoeAccount, error)
- Update(ctx, id, input) error
- Delete(ctx, id) error
- Isolate(ctx, id) error -- manual isolate, triggers MikroTik
- Restore(ctx, id) error -- manual restore, triggers MikroTik

internal/domain/cutoff/ (CutoffDomain)

- RunDailyCutoff(ctx) error -- main cron job
- ScanTenantsForCutoff(ctx) ([]TenantCutoffInfo, error)
- IsolateCustomer(ctx, customerID) error -- isolate all PPPoE accounts
- RestoreCustomer(ctx, customerID) error -- restore all PPPoE accounts
- OnInvoicePaid(ctx, invoiceID) error -- triggered by payment verification

internal/domain/activity_log/ (ActivityLogDomain)

- Log(ctx, input) error -- simple logging
- List(ctx, filter) ([]ActivityLog, error) -- admin view

6.2 Modifikasi Domain Existing

internal/domain/payment/

- Extend Verify() method: setelah verify → check isolated pppoe_accounts → trigger restore

internal/domain/nas/

- Extend Create(): setelah create NAS → provision isolir profile di MikroTik

internal/domain/invoice/

- Extend GenerateBulk(): gunakan tenant_settings.cutoff_day untuk due_date calculation

6.3 Update Domain Registry

internal/domain/registry.go tambah:
TenantSetting() tenant_setting.TenantSettingDomain
CustomerRegistration() customer_registration.CustomerRegistrationDomain
PppoeAccount() pppoe_account.PppoeAccountDomain
Cutoff() cutoff.CutoffDomain
ActivityLog() activity_log.ActivityLogDomain

---
BAGIAN 7: ADAPTERS

7.1 Outbound Adapters (Postgres)

Files baru di internal/adapter/outbound/postgres/:
- tenant_setting.go -- CRUD tenant_settings
- customer_registration.go -- CRUD customer_registrations
- pppoe_account.go -- CRUD pppoe_accounts
- activity_log.go -- Insert + query activity_logs
- mikrotik_sync_log.go -- Insert + query mikrotik_sync_logs

Update internal/adapter/outbound/postgres/registry.go:
- Tambah 5 adapter baru ke registry

7.2 Inbound Adapters (Gin HTTP)

Files baru di internal/adapter/inbound/gin/:
- tenant_setting.go -- GET/PUT tenant settings
- customer_registration.go -- Submit (public) + List/Approve/Reject (admin)
- pppoe_account.go -- CRUD + Isolate/Restore

Update internal/adapter/inbound/gin/adapter.go:
- Tambah 3 adapter baru

7.3 Message Consumer (RabbitMQ)

Files baru/update di internal/adapter/inbound/rabbitmq/:
- pppoe_sync.go -- Consumer untuk mikrotik.ppp.create, mikrotik.ppp.isolate, mikrotik.ppp.restore
- cutoff_worker.go -- Consumer untuk cutoff.process

7.4 Command (Cron)

File baru di internal/adapter/inbound/command/:
- cutoff.go -- Command handler untuk daily cutoff job

---
BAGIAN 8: ROUTES

8.1 Routes Baru

Update internal/adapter/inbound/gin/route.go:

# Public (no auth) - Customer Self-Registration
POST /register/:tenant_slug                    → CustomerRegistration.Submit

# Admin API (StaffAuth)
GET    /api/v1/tenant/settings                 → TenantSetting.Get
PUT    /api/v1/tenant/settings                 → TenantSetting.Update

GET    /api/v1/registrations                   → CustomerRegistration.List
GET    /api/v1/registrations/:id               → CustomerRegistration.Get
POST   /api/v1/registrations/:id/approve       → CustomerRegistration.Approve
POST   /api/v1/registrations/:id/reject        → CustomerRegistration.Reject

GET    /api/v1/pppoe-accounts                  → PppoeAccount.List
GET    /api/v1/pppoe-accounts/:id              → PppoeAccount.Get
POST   /api/v1/pppoe-accounts                  → PppoeAccount.Create
PUT    /api/v1/pppoe-accounts/:id              → PppoeAccount.Update
DELETE /api/v1/pppoe-accounts/:id              → PppoeAccount.Delete
POST   /api/v1/pppoe-accounts/:id/isolate      → PppoeAccount.Isolate
POST   /api/v1/pppoe-accounts/:id/restore      → PppoeAccount.Restore

GET    /api/v1/activity-logs                   → ActivityLog.List (read-only)
GET    /api/v1/sync-logs                       → MikrotikSyncLog.List (read-only)

# Customer Portal (existing group, tambahan)
GET    /portal/v1/pppoe-accounts               → Portal.ListPppoeAccounts

8.2 Permission Seeds (Migration Update)

Tambah permissions baru ke seed data:
- tenant_setting: manage
- customer_registration: read, create, update
- pppoe_account: read, create, update, delete, manage
- activity_log: read
- mikrotik_sync_log: read

---
BAGIAN 9: IMPLEMENTATION ORDER

Phase A: Database Foundation (Migrasi + Model)

1. Buat 5 model baru (tenant_setting, customer_registration, pppoe_account, activity_log, mikrotik_sync_log)
2. Update 3 model existing (customer, internet_package, subscription)
3. Buat 10 migration files (15-24)
4. Run migrations

Phase B: Outbound Ports + Adapters

5. Buat 5 outbound port interfaces
6. Update registry_database.go
7. Buat 5 postgres adapters
8. Update postgres registry
9. Generate mocks

Phase C: Domain Logic

10. Buat domain tenant_setting
11. Buat domain customer_registration (dengan approval flow)
12. Buat domain pppoe_account (dengan isolate/restore)
13. Buat domain cutoff (auto-cutoff logic)
14. Buat domain activity_log
15. Update domain payment (verify → trigger restore)
16. Update domain nas (create → provision isolir)
17. Update domain registry

Phase D: Inbound Ports + Adapters + Routes

18. Buat 3 inbound port interfaces
19. Update registry_http.go
20. Buat 3 gin adapters
21. Update route.go
22. Buat RabbitMQ consumers
23. Buat command handler cutoff
24. Update permission seeds

Phase E: Testing

25. Unit tests untuk semua domain baru
26. Unit tests untuk domain yang dimodifikasi
27. Integration tests untuk flows (registration, cutoff, restore)

---
BAGIAN 10: FILE MANIFEST

Files Baru (31 files)

Models (5):
- internal/model/tenant_setting.go
- internal/model/customer_registration.go
- internal/model/pppoe_account.go
- internal/model/activity_log.go
- internal/model/mikrotik_sync_log.go

Migrations (10):
- internal/migration/postgres/15_tenant_setting.go
- internal/migration/postgres/16_customer_registration.go
- internal/migration/postgres/17_pppoe_account.go
- internal/migration/postgres/18_activity_log.go
- internal/migration/postgres/19_mikrotik_sync_log.go
- internal/migration/postgres/20_alter_customer_add_auto_cutoff.go
- internal/migration/postgres/21_alter_internet_package_add_profile_name.go
- internal/migration/postgres/22_alter_subscription_add_pppoe_account_id.go
- internal/migration/postgres/23_migrate_pppoe_data.go
- internal/migration/postgres/24_alter_customer_remove_pppoe_fields.go

Outbound Ports (5):
- internal/port/outbound/tenant_setting.go
- internal/port/outbound/customer_registration.go
- internal/port/outbound/pppoe_account.go
- internal/port/outbound/activity_log.go
- internal/port/outbound/mikrotik_sync_log.go

Outbound Adapters - Postgres (5):
- internal/adapter/outbound/postgres/tenant_setting.go
- internal/adapter/outbound/postgres/customer_registration.go
- internal/adapter/outbound/postgres/pppoe_account.go
- internal/adapter/outbound/postgres/activity_log.go
- internal/adapter/outbound/postgres/mikrotik_sync_log.go

Inbound Ports (3):
- internal/port/inbound/tenant_setting.go
- internal/port/inbound/customer_registration.go
- internal/port/inbound/pppoe_account.go

Inbound Adapters - Gin (3):
- internal/adapter/inbound/gin/tenant_setting.go
- internal/adapter/inbound/gin/customer_registration.go
- internal/adapter/inbound/gin/pppoe_account.go

Domain Logic (5):
- internal/domain/tenant_setting/domain.go
- internal/domain/customer_registration/domain.go
- internal/domain/pppoe_account/domain.go
- internal/domain/cutoff/domain.go
- internal/domain/activity_log/domain.go

Files Dimodifikasi (12 files)

- internal/model/customer.go -- tambah AutoCutoff, hapus PPPoE fields
- internal/model/internet_package.go -- tambah ProfileName
- internal/model/subscription.go -- tambah PppoeAccountID
- internal/port/outbound/registry_database.go -- tambah 5 port baru
- internal/port/inbound/registry_http.go -- tambah 3 port baru
- internal/adapter/outbound/postgres/registry.go -- tambah 5 adapter baru
- internal/adapter/inbound/gin/adapter.go -- tambah 3 adapter baru
- internal/adapter/inbound/gin/route.go -- tambah routes baru
- internal/domain/registry.go -- tambah 5 domain baru
- internal/domain/payment/domain.go -- extend Verify → trigger restore
- internal/domain/nas/domain.go -- extend Create → provision isolir
- internal/migration/postgres/14_seed_roles_permissions.go -- tambah permissions baru

---
BAGIAN 11: VERIFICATION

Database

1. make migration-up -- semua migrasi berhasil
2. make migration-status -- semua migrasi applied
3. Verify tabel baru via psql:
- \dt -- list all tables
- \d tenant_settings -- verify columns
- \d pppoe_accounts -- verify columns + indexes

Unit Tests

4. make test -- semua existing tests tetap pass
5. Test domain tenant_setting (CRUD)
6. Test domain customer_registration (submit, approve, reject)
7. Test domain pppoe_account (create, isolate, restore)
8. Test domain cutoff (scan tenants, isolate flow)
9. Test payment verify → restore flow

Integration Tests (manual)

10. Submit registration → verify record di DB
11. Approve registration → verify customer + pppoe_account + subscription created
12. Create overdue invoice → run cutoff → verify pppoe isolated
13. Verify payment → verify pppoe restored
14. Test MikroTik sync (jika router tersedia)

API Testing

15. Test semua new routes via curl/Postman
16. Verify permission checks on all endpoints
17. Test public registration endpoint (no auth)