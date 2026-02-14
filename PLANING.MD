Excellent! Based on your answers, I now have a complete picture. Let me create a comprehensive implementation plan:
---
📋 COMPREHENSIVE IMPLEMENTATION PLAN
SUMMARY OF YOUR DECISIONS
| Component | Decision |
|-----------|----------|
| Isolation Profile | Special profile in bandwidth_profiles (category='isolated') |
| Payment Portal | Full-featured portal with Xendit integration |
| Payment Methods | All: VA, E-Wallet (GoPay/OVO/Dana), Card, QRIS |
| Invoice Schedule | System-wide billing day (configurable in settings) |
| WhatsApp Groups | Hardcoded in system_settings |
---
ARCHITECTURE OVERVIEW
┌─────────────────────────────────────────────────────────────────┐
│                        Application Layer                      │
│  (Gin HTTP + WebSocket + Temporal Workflows + RabbitMQ)    │
└─────────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  Domain      │    │  Domain      │    │  Domain      │
│  Layer       │    │  Layer       │    │  Layer       │
└──────────────┘    └──────────────┘    └──────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  PostgreSQL  │    │   Redis      │    │   Mikrotik   │
│  (Database)  │    │   (Cache)    │    │   (API)      │
└──────────────┘    └──────────────┘    └──────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Xendit     │    │  Gowa       │    │  Temporal    │
│   (Payment)   │    │ (WhatsApp)   │    │  (Scheduler) │
└──────────────┘    └──────────────┘    └──────────────┘
---
PHASE 1: CORE BUSINESS DATA (Foundations)
1.1 Database Migrations
Migration 6: Bandwidth Profiles
File: internal/migration/postgres/6_bandwidth_profile.go
Table: bandwidth_profiles
Fields:
- id (UUID, PK)
- profile_code (VARCHAR, UNIQUE)
- name (VARCHAR)
- description (TEXT)
- category (ENUM: 'pppoe', 'ip-static', 'hotspot', 'isolated')
- ppp_profile_name (VARCHAR) - profile name on Mikrotik
- download_speed (BIGINT) - in kbps
- upload_speed (BIGINT) - in kbps
- burst_download (BIGINT)
- burst_upload (BIGINT)
- burst_threshold (INTEGER) - percentage 80%
- burst_time (INTEGER) - seconds
- priority (INTEGER 1-8)
- queue_type (VARCHAR)
- shared_users (INTEGER)
- queue_name (VARCHAR)
- price_monthly (DECIMAL 12,2)
- price_installation (DECIMAL 12,2)
- tax_rate (DECIMAL 5,4) - default 0.11
- is_active (BOOLEAN)
- is_visible (BOOLEAN)
- sort_order (INTEGER)
- created_at, updated_at, deleted_at
Indexes:
- idx_bandwidth_profiles_category
- idx_bandwidth_profiles_is_active
- idx_bandwidth_profiles_is_visible
- idx_bandwidth_profiles_code
Migration 7: Customers
File: internal/migration/postgres/7_customer.go
Table: customers
Fields:
- id (UUID, PK)
- customer_code (VARCHAR, UNIQUE)
- full_name (VARCHAR)
- email (VARCHAR, UNIQUE)
- phone (VARCHAR)
- address (TEXT)
- coordinates (GEOGRAPHY) - lat/lng
- status (ENUM: 'pending', 'active', 'suspended', 'isolated', 'terminated')
- activation_date (DATE)
- installation_date (DATE)
- termination_date (DATE)
- expiry_date (DATE) - billing due date
- router_id (UUID, FK to mikrotik_routers)
- ppp_secret_name (VARCHAR, UNIQUE) - username on Mikrotik
- ppp_secret_password (TEXT, encrypted) - password on Mikrotik
- ppp_service (VARCHAR) - pppoe, pptp, l2tp, ovpn
- static_ip (VARCHAR)
- mac_address (VARCHAR)
- profile_id (UUID, FK to bandwidth_profiles)
- billing_cycle (VARCHAR) - monthly, quarterly, yearly
- billing_day (INTEGER 1-31)
- payment_method_preference (VARCHAR)
- auto_isolate (BOOLEAN) - default true
- grace_period_days (INTEGER) - default 3
- notes (TEXT)
- tags (JSONB)
- created_by, updated_by (FK to users)
- created_at, updated_at, deleted_at
Indexes:
- idx_customers_status
- idx_customers_router
- idx_customers_profile
- idx_customers_code
- idx_customers_phone
- idx_customers_email
- idx_customers_expiry
- idx_customers_ppp_secret
- idx_customers_coordinates (GIST)
Migration 8: System Settings
File: internal/migration/postgres/8_system_setting.go
Table: system_settings
Fields:
- id (UUID, PK)
- key (VARCHAR, UNIQUE)
- value (TEXT)
- value_type (ENUM: 'string', 'number', 'boolean', 'json')
- category (VARCHAR) - 'billing', 'payment', 'notification', 'system', 'company', 'whatsapp', 'xendit'
- description (TEXT)
- is_public (BOOLEAN)
- updated_by (FK to users)
- updated_at
Default settings (INSERT):
- invoice.due_days = 7
- invoice.grace_period_days = 3
- invoice.late_fee_enabled = true
- invoice.late_fee_amount = 50000
- invoice.auto_generate_day = 25
- company.name = "PT Internet Provider"
- company.address = "Jl. Raya No. 123"
- company.phone = "021-12345678"
- company.email = "info@isp.com"
- company.tax_id = "01.234.567.8-901.000"
- whatsapp.enabled = true
- whatsapp.group_billing = "group-id"
- whatsapp.group_support = "group-id"
- whatsapp.group_notifications = "group-id"
- xendit.secret_key = ""
- xendit.api_key = ""
- payment_portal_url = "https://portal.example.com"
1.2 Domain Layers
Bandwidth Profile Domain
internal/domain/bandwidth_profile/
├── domain.go
└── domain_test.go
Interface Methods:
- CreateProfile(input BandwidthProfileInput) (*BandwidthProfile, error)
- GetProfile(id string) (*BandwidthProfile, error)
- ListProfiles(filter BandwidthProfileFilter) ([]BandwidthProfile, error)
- UpdateProfile(id string, input BandwidthProfileInput) (*BandwidthProfile, error)
- DeleteProfile(id string) error
- SyncToMikrotik(profile *BandwidthProfile) error
- GetIsolatedProfile() (*BandwidthProfile, error)
1.3 Models
Model: Bandwidth Profile
internal/model/bandwidth_profile.go
Types:
- BandwidthProfile
- BandwidthProfileInput
- BandwidthProfileFilter
Model: Customer
internal/model/customer.go
Types:
- Customer
- CustomerInput
- CustomerFilter
Model: System Setting
internal/model/system_setting.go
Types:
- SystemSetting
- SystemSettingInput
- SystemSettingFilter
1.4 Database Adapters
Postgres Adapter: Bandwidth Profile
internal/adapter/outbound/postgres/bandwidth_profile.go
Methods:
- CRUD operations
- Find by category (for isolation profile)
- Find by profile_code
Postgres Adapter: Customer
internal/adapter/outbound/postgres/customer.go
Methods:
- CRUD operations
- Find by router_id
- Find by profile_id
- Find expired customers (for isolation)
- Find customers expiring soon (for notifications)
Postgres Adapter: System Setting
internal/adapter/outbound/postgres/system_setting.go
Methods:
- GetByKey(key string) (*SystemSetting, error)
- UpdateByKey(key string, value string) error
- GetAll(category string) ([]SystemSetting, error)
1.5 HTTP Handlers (Gin)
Bandwidth Profile Handler
internal/adapter/inbound/gin/bandwidth_profile_handler.go
Endpoints:
- POST /v1/bandwidth-profiles
- GET /v1/bandwidth-profiles
- GET /v1/bandwidth-profiles/:id
- PUT /v1/bandwidth-profiles/:id
- DELETE /v1/bandwidth-profiles/:id
Customer Handler
internal/adapter/inbound/gin/customer_handler.go
Endpoints:
- POST /v1/customers
- GET /v1/customers
- GET /v1/customers/:id
- PUT /v1/customers/:id
- DELETE /v1/customers/:id
- GET /v1/customers/:id/billing-info
- POST /v1/customers/:id/isolate
- POST /v1/customers/:id/activate
System Settings Handler
internal/adapter/inbound/gin/system_setting_handler.go
Endpoints:
- GET /v1/settings
- GET /v1/settings/:key
- PUT /v1/settings/:key
---
PHASE 2: BILLING SYSTEM
2.1 Database Migrations
Migration 9: Invoices
File: internal/migration/postgres/9_invoice.go
Table: invoices
Fields:
- id (UUID, PK)
- invoice_number (VARCHAR, UNIQUE)
- customer_id (UUID, FK to customers)
- billing_period_start (DATE)
- billing_period_end (DATE)
- billing_month (INTEGER 1-12)
- billing_year (INTEGER)
- issue_date (DATE)
- due_date (DATE)
- payment_deadline (TIMESTAMP)
- subtotal (DECIMAL 12,2)
- tax_amount (DECIMAL 12,2)
- discount_amount (DECIMAL 12,2)
- late_fee (DECIMAL 12,2)
- total_amount (DECIMAL 12,2)
- paid_amount (DECIMAL 12,2)
- status (ENUM: 'draft', 'sent', 'partial', 'paid', 'overdue', 'cancelled', 'refunded')
- payment_status (ENUM: 'unpaid', 'partial', 'paid', 'overpaid')
- payment_date (TIMESTAMP)
- payment_method (VARCHAR)
- invoice_type (ENUM: 'recurring', 'installation', 'additional', 'refund')
- is_auto_generated (BOOLEAN)
- reminder_sent_count (INTEGER)
- last_reminder_sent (TIMESTAMP)
- notes (TEXT)
- internal_notes (TEXT)
- created_by, updated_by (FK to users)
- created_at, updated_at, deleted_at
Indexes:
- idx_invoices_customer
- idx_invoices_status
- idx_invoices_due_date
- idx_invoices_number
- idx_invoices_billing_period
- idx_invoices_payment_status
Migration 10: Invoice Items
File: internal/migration/postgres/10_invoice_item.go
Table: invoice_items
Fields:
- id (UUID, PK)
- invoice_id (UUID, FK to invoices, CASCADE)
- item_type (ENUM: 'subscription', 'installation', 'equipment', 'other')
- description (VARCHAR)
- profile_id (UUID, FK to bandwidth_profiles)
- quantity (INTEGER)
- unit_price (DECIMAL 12,2)
- subtotal (DECIMAL 12,2)
- tax_rate (DECIMAL 5,4)
- tax_amount (DECIMAL 12,2)
- total (DECIMAL 12,2)
- is_prorated (BOOLEAN)
- proration_days (INTEGER)
- proration_percentage (DECIMAL 5,2)
- sort_order (INTEGER)
- created_at
Indexes:
- idx_invoice_items_invoice
- idx_invoice_items_profile
2.2 Temporal Workflows
Invoice Generation Workflow
internal/adapter/inbound/temporal/billing/worker.go
internal/adapter/inbound/temporal/billing/workflow.go
Workflow:
GenerateMonthlyInvoicesWorkflow(ctx workflow.Context, input GenerateInvoicesInput) (result string, err error)
Activities:
- GenerateInvoicesActivity - Create invoices for all customers
- SendInvoiceNotificationsActivity - Send WhatsApp notifications
- CheckOverdueInvoicesActivity - Update status for overdue invoices
Schedule:
- Run daily at configured invoice.auto_generate_day (e.g., 25th of month)
- Check if invoices already generated for current month
- Generate only if not generated
Invoice Payment Workflow
internal/adapter/inbound/temporal/billing/payment_workflow.go
Workflow:
ProcessPaymentWorkflow(ctx workflow.Context, input PaymentWebhookInput) (result string, err error)
Activities:
- ValidatePaymentActivity - Validate Xendit webhook
- UpdateInvoiceStatusActivity - Mark invoice as paid
- RecordCashTransactionActivity - Auto-create cash transaction
- SendPaymentNotificationActivity - Notify customer via WhatsApp
- ReactivateCustomerActivity - If customer was isolated
2.3 Domain: Billing
internal/domain/billing/
├── domain.go
└── domain_test.go
Interface Methods:
- CreateInvoice(input InvoiceInput) (*Invoice, error)
- GetInvoice(id string) (*Invoice, error)
- ListInvoices(filter InvoiceFilter) ([]Invoice, error)
- UpdateInvoiceStatus(id string, status string) error
- GenerateMonthlyInvoices(year int, month int) ([]Invoice, error)
- ApplyPayment(invoiceID string, amount decimal.Decimal) error
- CheckOverdueInvoices() ([]Invoice, error)
- CalculateLateFee(invoice *Invoice) decimal.Decimal
---
PHASE 3: PAYMENT SYSTEM (XENDIT)
3.1 Database Migrations
Migration 11: Payments
File: internal/migration/postgres/11_payment.go
Table: payments
Fields:
- id (UUID, PK)
- payment_number (VARCHAR, UNIQUE)
- customer_id (UUID, FK to customers)
- invoice_id (UUID, FK to invoices, nullable)
- amount (DECIMAL 12,2)
- allocated_amount (DECIMAL 12,2)
- payment_method (ENUM: 'va', 'e-wallet', 'credit_card', 'debit_card', 'check')
- payment_date (TIMESTAMP)
- bank_name (VARCHAR)
- bank_account_number (VARCHAR)
- bank_account_name (VARCHAR)
- transaction_reference (VARCHAR) - Xendit payment ID
- ewallet_provider (VARCHAR) - gopay, ovo, dana
- ewallet_number (VARCHAR)
- proof_image (TEXT)
- receipt_number (VARCHAR)
- status (ENUM: 'pending', 'confirmed', 'rejected', 'refunded')
- processed_by (FK to users)
- processed_at (TIMESTAMP)
- rejection_reason (TEXT)
- refund_amount (DECIMAL 12,2)
- refund_date (TIMESTAMP)
- refund_reason (TEXT)
- refunded_by (FK to users)
- notes (TEXT)
- created_by (FK to users)
- created_at, updated_at, deleted_at
Indexes:
- idx_payments_customer
- idx_payments_invoice
- idx_payments_status
- idx_payments_date
- idx_payments_method
- idx_payments_number
Migration 12: Payment Allocations
File: internal/migration/postgres/12_payment_allocation.go
Table: payment_allocations
Fields:
- id (UUID, PK)
- payment_id (UUID, FK to payments, CASCADE)
- invoice_id (UUID, FK to invoices, CASCADE)
- allocated_amount (DECIMAL 12,2)
- created_at
Indexes:
- idx_payment_allocations_payment
- idx_payment_allocations_invoice
UNIQUE: (payment_id, invoice_id)
3.2 Xendit Integration Utility
utils/xendit/
├── client.go          // Xendit HTTP client
├── invoice.go         // Xendit Invoice API
├── payment.go         // Xendit Virtual Account/E-Wallet API
├── webhook.go        // Webhook verification
└── types.go         // Request/Response types
Features:
- Create payment link (VA, E-Wallet, QRIS)
- Get payment status
- Verify webhook signature
- Handle payment callbacks
3.3 Payment Portal (Frontend)
web/payment-portal/
├── index.html
├── static/
│   ├── css/
│   └── js/
└── templates/
    └── payment.html
Features:
- Display unpaid invoices
- Show payment history
- Select payment method
- Redirect to Xendit checkout
- Payment success/failure pages
- Download invoice PDF
API Endpoints:
- GET /portal/:customer_code - Customer portal
- POST /api/v1/payments/create - Create Xendit payment
- POST /api/v1/webhooks/xendit - Xendit webhook handler
- GET /api/v1/payments/:invoice_id/status - Payment status
3.4 Domain: Payment
internal/domain/payment/
├── domain.go
└── domain_test.go
Interface Methods:
- CreatePayment(input PaymentInput) (*Payment, error)
- GetPayment(id string) (*Payment, error)
- ListPayments(filter PaymentFilter) ([]Payment, error)
- ProcessWebhook(webhookData XenditWebhookData) error
- AllocatePayment(paymentID string, invoiceID string, amount decimal.Decimal) error
- GetUnpaidInvoices(customerID string) ([]Invoice, error)
---
PHASE 4: CASH MANAGEMENT
4.1 Database Migrations
Migration 13: Cash Categories
File: internal/migration/postgres/13_cash_category.go
Table: cash_categories
Fields:
- id (UUID, PK)
- code (VARCHAR, UNIQUE)
- name (VARCHAR)
- type (ENUM: 'income', 'expense')
- parent_category_id (UUID, FK to cash_categories, CASCADE)
- description (TEXT)
- is_system (BOOLEAN) - cannot be deleted
- is_active (BOOLEAN)
- sort_order (INTEGER)
- created_at, updated_at
Indexes:
- idx_cash_categories_type
- idx_cash_categories_parent
Default categories:
- INC-SUB: Subscription Revenue
- INC-INS: Installation Revenue
- INC-OTH: Other Income
- EXP-OPR: Operational Expense
- EXP-EQP: Equipment Purchase
- EXP-SAL: Salary
- EXP-UTL: Utilities
- EXP-MNT: Maintenance
- EXP-OTH: Other Expense
Migration 14: Cash Transactions
File: internal/migration/postgres/14_cash_transaction.go
Table: cash_transactions
Fields:
- id (UUID, PK)
- transaction_number (VARCHAR, UNIQUE)
- transaction_date (TIMESTAMP)
- type (ENUM: 'income', 'expense')
- category_id (UUID, FK to cash_categories)
- amount (DECIMAL 12,2)
- payment_method (VARCHAR)
- description (TEXT)
- reference_type (VARCHAR) - 'payment', 'invoice', 'expense', 'other'
- reference_id (UUID)
- customer_id (UUID, FK to customers)
- account_name (VARCHAR)
- account_number (VARCHAR)
- proof_image (TEXT)
- receipt_number (VARCHAR)
- requires_approval (BOOLEAN)
- approval_status (ENUM: 'pending', 'approved', 'rejected')
- approved_by (FK to users)
- approved_at (TIMESTAMP)
- notes (TEXT)
- processed_by (FK to users)
- created_at, updated_at, deleted_at
Indexes:
- idx_cash_transactions_date
- idx_cash_transactions_type
- idx_cash_transactions_category
- idx_cash_transactions_customer
- idx_cash_transactions_reference
- idx_cash_transactions_number
- idx_cash_transactions_approval
4.2 Domain: Cash
internal/domain/cash/
├── domain.go
└── domain_test.go
Interface Methods:
- CreateTransaction(input CashTransactionInput) (*CashTransaction, error)
- GetTransaction(id string) (*CashTransaction, error)
- ListTransactions(filter CashTransactionFilter) ([]CashTransaction, error)
- CreateCategory(input CashCategoryInput) (*CashCategory, error)
- ListCategories(type string) ([]CashCategory, error)
- AutoRecordPayment(payment *Payment) error
---
PHASE 5: ISOLATION SYSTEM
5.1 Temporal Workflows
Customer Isolation Workflow
internal/adapter/inbound/temporal/isolation/
├── worker.go
└── workflow.go
Workflow:
IsolateExpiredCustomersWorkflow(ctx workflow.Context, input IsolateInput) (result string, err error)
Activities:
- FindExpiredCustomersActivity - Get customers past expiry_date
- CheckAutoIsolateFlagActivity - Filter customers with auto_isolate=true
- IsolateCustomerActivity - Change profile to limited speed on Mikrotik
- UpdateCustomerStatusActivity - Change status to 'isolated'
- SendIsolationNotificationActivity - Notify via WhatsApp
- SendPortalLinkActivity - Send payment portal link
Schedule:
- Run daily at 00:00
- Check all customers with expiry_date < today
- Apply grace period (default 3 days)
- Isolate if past grace period
Customer Reactivation Workflow
internal/adapter/inbound/temporal/isolation/reactivation_workflow.go
Workflow:
ReactivateCustomerWorkflow(ctx workflow.Context, customerID string) (result string, err error)
Activities:
- GetCustomerActivity - Fetch customer details
- GetOriginalProfileActivity - Get the profile before isolation
- ReactivateOnMikrotikActivity - Restore original profile
- UpdateCustomerStatusActivity - Change status back to 'active'
- UpdateExpiryDateActivity - Set new expiry_date
- SendReactivationNotificationActivity - Notify via WhatsApp
5.2 Mikrotik Integration
Isolation Logic:
// In customer domain
func (d *domain) IsolateCustomer(customerID string) error {
    customer, err := d.dbPort.Customer().GetByID(customerID)
    if err != nil {
        return err
    }
    
    // Get isolation profile (limited speed)
    isolatedProfile, err := d.bandwidthProfile.GetIsolatedProfile()
    if err != nil {
        return err
    }
    
    // Update customer status
    customer.Status = "isolated"
    
    // Update PPP secret on Mikrotik
    secret := &model.PppoeSecret{
        Name:    customer.PppSecretName,
        Profile:  isolatedProfile.PppProfileName,
        Disabled: false, // Allow login but with limited speed
    }
    
    router, err := d.dbPort.Mikrotik().FindByID(customer.RouterID.String())
    if err != nil {
        return err
    }
    
    err = d.mikrotikPort.UpdateSecret(router, secret)
    if err != nil {
        return err
    }
    
    return d.dbPort.Customer().Update(customer)
}
Portal Redirect Script on Mikrotik:
// This script will be added to the isolated profile
func (d *domain) GenerateRedirectScript(customer *model.Customer) string {
    portalURL := d.getPaymentPortalURL()
    customerCode := customer.CustomerCode
    
    // RouterOS script to redirect to portal
    script := fmt.Sprintf(`
# Customer: %s
:local portalURL "%s"
:local customerCode "%s"
# Redirect HTTP to portal
/ip firewall nat add chain=dstnat protocol=tcp dst-port=80 src-address=%s action=redirect to-ports=8080
# Walled garden (allow payment portal)
/ip firewall nat add chain=dstnat protocol=tcp dst-port=80,443 dst-address-list=portal-allowed action=accept
    `, customerCode, portalURL, customerCode, customer.IPAddress)
    
    return script
}
---
PHASE 6: NOTIFICATIONS (WHATSAPP - GOWA)
6.1 Database Migration
Migration 15: Notifications
File: internal/migration/postgres/15_notification.go
Table: notifications
Fields:
- id (UUID, PK)
- customer_id (UUID, FK to customers, CASCADE)
- type (VARCHAR) - 'invoice_created', 'payment_received', 'overdue_warning', 'customer_isolated', 'customer_reactivated'
- channel (ENUM: 'email', 'whatsapp', 'sms', 'in-app')
- title (VARCHAR)
- message (TEXT)
- template_name (VARCHAR)
- template_data (JSONB)
- status (ENUM: 'pending', 'sent', 'failed', 'read')
- sent_at (TIMESTAMP)
- read_at (TIMESTAMP)
- reference_type (VARCHAR)
- reference_id (UUID)
- error_message (TEXT)
- retry_count (INTEGER)
- max_retries (INTEGER)
- created_at
Indexes:
- idx_notifications_customer
- idx_notifications_status
- idx_notifications_created
- idx_notifications_type
6.2 Gowa Integration Utility
utils/gowa/
├── client.go         // Gowa HTTP client
├── message.go        // Send message methods
├── group.go          // Send to group methods
└── types.go         // Request/Response types
Features:
- Send text message
- Send to group
- Check delivery status
- Handle rate limits
6.3 Domain: Notification
internal/domain/notification/
├── domain.go
└── domain_test.go
Interface Methods:
- CreateNotification(input NotificationInput) (*Notification, error)
- SendNotification(notificationID string) error
- SendInvoiceNotification(invoice *Invoice) error
- SendPaymentNotification(payment *Payment) error
- SendOverdueReminder(invoice *Invoice) error
- SendIsolationNotification(customer *Customer) error
- SendReactivationNotification(customer *Customer) error
- SendToGroup(groupID string, message string) error
- ProcessPendingNotifications() error
6.4 Notification Templates
Templates (stored in system_settings or separate table):
type NotificationTemplate struct {
    Name         string
    Subject      string
    Message      string
    Type         string
}
const (
    TemplateInvoiceCreated = "invoice_created"
    TemplatePaymentReceived = "payment_received"
    TemplateOverdueWarning = "overdue_warning"
    TemplateCustomerIsolated = "customer_isolated"
    TemplateCustomerReactivated = "customer_reactivated"
)
// Example templates
var templates = map[string]NotificationTemplate{
    TemplateInvoiceCreated: {
        Subject: "Invoice #[InvoiceNumber] - [CustomerName]",
        Message: "Halo [CustomerName], invoice bulanan Anda telah dibuat. Total: Rp[TotalAmount]. Jatuh tempo: [DueDate]. Silakan bayar melalui portal: [PortalURL]",
    },
    TemplateOverdueWarning: {
        Subject: "Pengingat Pembayaran - Invoice #[InvoiceNumber]",
        Message: "Halo [CustomerName], invoice Anda sudah overdue. Mohon segera lakukan pembayaran untuk menghindari isolasi. Portal: [PortalURL]",
    },
    TemplateCustomerIsolated: {
        Subject: "Layanan Terisolasi - [CustomerName]",
        Message: "Halo [CustomerName], layanan Anda telah diisolasi karena pembayaran terlambat. Silakan bayar melalui portal: [PortalURL]",
    },
}
---
PHASE 7: ACTIVITY LOGGING
7.1 Database Migration
Migration 16: Activity Logs
File: internal/migration/postgres/16_activity_log.go
Table: activity_logs
Fields:
- id (UUID, PK)
- user_id (UUID, FK to users, nullable)
- action (VARCHAR) - 'create', 'update', 'delete', 'login', 'logout', 'isolate', 'reactivate'
- entity_type (VARCHAR) - 'customer', 'invoice', 'payment', 'bandwidth_profile'
- entity_id (UUID)
- description (TEXT)
- old_values (JSONB)
- new_values (JSONB)
- ip_address (VARCHAR)
- user_agent (TEXT)
- created_at
Indexes:
- idx_activity_logs_user
- idx_activity_logs_entity
- idx_activity_logs_created
- idx_activity_logs_action
7.2 Domain: Activity
internal/domain/activity/
├── domain.go
└── domain_test.go
Interface Methods:
- LogActivity(input ActivityLogInput) error
- ListLogs(filter ActivityLogFilter) ([]ActivityLog, error)
- GetEntityHistory(entityType string, entityID string) ([]ActivityLog, error)
7.3 Middleware Integration
// In gin/middleware.go
func ActivityLoggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Capture request details
        start := time.Now()
        
        c.Next()
        
        // Log after request completes
        duration := time.Since(start)
        
        activityLog := &model.ActivityLog{
            Action:      c.Request.Method + " " + c.Request.URL.Path,
            EntityType:  "api_request",
            Description: fmt.Sprintf("API call completed in %v", duration),
            IPAddress:   c.ClientIP(),
            UserAgent:   c.Request.UserAgent(),
            UserID:      getUserIDFromContext(c),
        }
        
        // Log asynchronously
        go logActivity(activityLog)
    }
}
---
PHASE 8: PORT REGISTRY & INTEGRATION
8.1 Update Registry Files
All new domains need to be registered:
// internal/domain/registry.go
type Domain interface {
    // Existing domains...
    
    // New domains
    BandwidthProfile() bandwidth_profile.BandwidthProfileDomain
    Customer() customer.CustomerDomain
    Billing() billing.BillingDomain
    Payment() payment.PaymentDomain
    Cash() cash.CashDomain
    Notification() notification.NotificationDomain
    Activity() activity.ActivityDomain
}
// internal/adapter/inbound/gin/route.go
// Add new route groups:
bandwidthProfile := app.Group("/bandwidth-profiles")
customer := app.Group("/customers")
billing := app.Group("/billing")
payment := app.Group("/payments")
cash := app.Group("/cash")
notification := app.Group("/notifications")
portal := app.Group("/portal") // Public access
8.2 Environment Variables
Add to .env.example:
# Invoice Settings
INVOICE_DUE_DAYS=7
INVOICE_GRACE_PERIOD_DAYS=3
INVOICE_LATE_FEE_ENABLED=true
INVOICE_LATE_FEE_AMOUNT=50000
INVOICE_AUTO_GENERATE_DAY=25
# Xendit Settings
XENDIT_SECRET_KEY=
XENDIT_API_KEY=
XENDIT_WEBHOOK_TOKEN=
XENDIT_ENVIRONMENT=development # development, production
# Gowa WhatsApp Settings
GOWA_API_URL=http://localhost:3000
GOWA_API_KEY=
GOWA_TIMEOUT=30
# WhatsApp Group IDs
WHATSAPP_GROUP_BILLING=
WHATSAPP_GROUP_SUPPORT=
WHATSAPP_GROUP_NOTIFICATIONS=
# Payment Portal
PAYMENT_PORTAL_URL=https://portal.example.com
PAYMENT_PORTAL_ENABLED=true
---
IMPLEMENTATION PRIORITY & DEPENDENCIES
Phase 1 (FOUNDATIONS)
├── Migration 6, 7, 8
├── Domain: BandwidthProfile, Customer, SystemSetting
├── Adapters: Database, HTTP
└── [BLOCKS] Phase 2, 3, 4, 5, 6
Phase 2 (BILLING)
├── [DEPENDS ON] Phase 1
├── Migration 9, 10
├── Domain: Billing
├── Workflows: InvoiceGeneration
├── [ENABLES] Phase 3, Phase 5
Phase 3 (PAYMENTS)
├── [DEPENDS ON] Phase 1, 2
├── Migration 11, 12
├── Utils: Xendit
├── Domain: Payment
├── Web: Payment Portal
└── [ENABLES] Phase 4
Phase 4 (CASH)
├── [DEPENDS ON] Phase 1, 3
├── Migration 13, 14
├── Domain: Cash
└── [OPTIONAL - Can start with Phase 1-3]
Phase 5 (ISOLATION)
├── [DEPENDS ON] Phase 1, 2
├── Domain: Isolation (part of Customer)
├── Workflows: Isolation, Reactivation
├── Mikrotik: Profile switching
└── [REQUIRES] Phase 6 for notifications
Phase 6 (NOTIFICATIONS)
├── [DEPENDS ON] Phase 1
├── Migration 15
├── Utils: Gowa
├── Domain: Notification
└── [INTEGRATES] All phases
Phase 7 (ACTIVITY LOGS)
├── [DEPENDS ON] Phase 1
├── Migration 16
├── Domain: Activity
├── Middleware: Activity logging
└── [OPTIONAL - Can implement anytime]
---
CRITICAL SUCCESS FACTORS
1. Transaction Safety
- All database operations must be wrapped in transactions
- Use dbPort.DoInTransaction() for complex operations
- Example: Creating invoice + items + notifications
2. Mikrotik Sync Reliability
- Implement retry logic for Mikrotik API calls
- Log all sync operations
- Handle connection failures gracefully
- Queue failed sync operations for retry
3. Payment Security
- Verify Xendit webhook signatures
- Store sensitive keys securely (encrypted in DB)
- Use HTTPS for all payment endpoints
- Validate payment amounts before applying
4. Notification Delivery
- Implement retry mechanism with exponential backoff
- Mark failed notifications for manual review
- Support both individual and group notifications
- Rate limit to avoid spam
5. Isolation Safety
- Never delete customer data
- Always preserve original profile
- Send notifications before and after isolation
- Provide one-click reactivation
---
ESTIMATED EFFORT
| Phase | Tasks | Estimated Time |
|--------|--------|----------------|
| Phase 1: Core Business Data | 3 migrations + 3 domains + 3 adapters + 3 handlers | 3-5 days |
| Phase 2: Billing System | 2 migrations + 1 domain + 2 workflows | 4-6 days |
| Phase 3: Payment System | 2 migrations + 1 utility + 1 domain + portal | 5-7 days |
| Phase 4: Cash Management | 2 migrations + 1 domain | 2-3 days |
| Phase 5: Isolation System | 2 workflows + Mikrotik integration | 3-4 days |
| Phase 6: Notifications | 1 migration + 1 utility + 1 domain + templates | 3-4 days |
| Phase 7: Activity Logs | 1 migration + 1 domain + middleware | 1-2 days |
| Phase 8: Integration | Registry updates + routing + env vars | 1-2 days |
| TOTAL | 16 migrations + 8 domains + 3 utilities | 22-33 days |

  Jika ingin melanjutkan:
  1. Implement Temporal Activity Functions (untuk workflow execution)
  2. Add Refund HTTP Endpoints (CRUD refund)
  3. Write Tests (unit & integration tests)
  4. Configure Temporal Cron (scheduled isolation check)
  5. Add Reconciliation Reports (PDF/CSV export)