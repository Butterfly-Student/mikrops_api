# Auto-Start Temporal Schedulers - IMPLEMENTATION COMPLETE

## ✅ **Status: COMPLETE**

**Date**: 2026-02-16

---

## 📋 **OVERVIEW**

Temporal schedulers sekarang otomatis berjalan saat HTTP server di-start, tidak perlu menjalankan worker terpisah untuk menjalankan schedulers.

---

## ✅ **IMPLEMENTATION**

### **1. App.go Modifications**

**File**: `internal/app.go`

#### Changes Made:

1. **Convert all outbound functions to App methods**
   - `databaseOutbound(ctx)` → `a.databaseOutbound(ctx)`
   - `messageOutbound(ctx)` → `a.messageOutbound(ctx)`
   - `cacheOutbound(ctx)` → `a.cacheOutbound(ctx)`
   - `mikrotikOutbound()` → `a.mikrotikOutbound()`
   - `workflowOutbound(ctx)` → `a.workflowOutbound(ctx)`

2. **Refactor NewApp()** to create app instance first, then initialize dependencies

3. **Auto-start schedulers in httpInbound()**

```go
// Auto-start temporal schedulers if workflow is configured
workflowPort := a.domain.Workflow()
if workflowPort != nil {
    log.WithContext(ctx).Info("Starting temporal schedulers")

    // Start billing scheduler
    workflowPort.Billing().StartScheduler(ctx)

    // Start isolation scheduler
    workflowPort.Isolation().StartScheduler(ctx)

    log.WithContext(ctx).Info("Temporal schedulers started successfully")
}
```

---

### **2. Scheduler Simplifications**

#### **Billing Scheduler** (`internal/adapter/inbound/temporal/billing/scheduler.go`)

**Changes**:
- Remove `workflowPort` parameter from `Scheduler` struct
- Remove `workflowPort` parameter from `NewScheduler()`
- Update all calls from `s.workflowPort.Billing()` to `s.domain.Workflow().Billing()`

```go
type Scheduler struct {
    domain domain.Domain  // Removed workflowPort
}

func NewScheduler(domain domain.Domain) *Scheduler {  // Removed workflowPort parameter
    return &Scheduler{
        domain: domain,
    }
}

// Usage in scheduler functions:
err := s.domain.Workflow().Billing().TriggerGenerateMonthlyInvoices(ctx, year, month)
```

#### **Isolation Scheduler** (`internal/adapter/inbound/temporal/isolation/scheduler.go`)

**Changes**:
- Remove `workflowPort` parameter from `IsolationScheduler` struct
- Remove `workflowPort` parameter from `NewIsolationScheduler()`
- Update calls from `s.workflowPort.Isolation()` to `s.domain.Workflow().Isolation()`

```go
type IsolationScheduler struct {
    domain domain.Domain  // Removed workflowPort
}

func NewIsolationScheduler(domain domain.Domain) *IsolationScheduler {  // Removed workflowPort parameter
    return &IsolationScheduler{
        domain: domain,
    }
}

// Usage in triggerIsolationCheck():
err := s.domain.Workflow().Isolation().TriggerIsolateExpiredCustomers(ctx, gracePeriodDays)
```

---

### **3. Adapter Updates**

#### **Billing Adapter** (`internal/adapter/inbound/temporal/billing/adapter.go`)

**Updated StartScheduler method**:
```go
func (a *billingAdapter) StartScheduler(ctx context.Context) {
    scheduler := NewScheduler(a.domain)  // Removed workflowPort parameter
    scheduler.Start(ctx)
}
```

#### **Isolation Adapter** (`internal/adapter/inbound/temporal/isolation/adapter.go`)

**Updated StartScheduler method**:
```go
func (a *isolationAdapter) StartScheduler(ctx context.Context) {
    scheduler := NewIsolationScheduler(a.domain)  // Removed workflowPort parameter
    scheduler.Start(ctx)
}
```

#### **Temporal Outbound Adapter** (`internal/adapter/outbound/temporal/registry.go`)

**Updated StartScheduler methods**:
```go
func (a *adapter) StartBillingScheduler(ctx context.Context) {
    a.billingAdapter.StartScheduler(ctx)  // Removed wrapper parameter
}

func (a *adapter) StartIsolationScheduler(ctx context.Context) {
    a.isolationAdapter.StartScheduler(ctx)  // Removed wrapper parameter
}
```

**Removed unused**:
- `workflowWrapper` struct
- Wrapper methods

---

### **4. Port Interface Updates**

#### **Billing Workflow Port** (`internal/port/inbound/billing_workflow.go`)

**Updated interface**:
```go
type BillingWorkflowPort interface {
    // Workers
    StartInvoiceGenerationWorker(ctx context.Context)
    StartPaymentProcessingWorker(ctx context.Context)

    // Scheduler
    StartScheduler(ctx context.Context)  // Removed workflowPort parameter

    // Workflow Triggers
    TriggerGenerateMonthlyInvoices(ctx context.Context, year, month int) error
    TriggerProcessPayment(ctx context.Context, input ProcessPaymentWorkflowInput) error
    TriggerCheckOverdue(ctx context.Context, applyLateFees bool) error
    TriggerSendReminders(ctx context.Context, daysBefore, daysAfter []int) error
}
```

#### **Isolation Workflow Port** (`internal/port/inbound/isolation_workflow.go`)

**Updated interface**:
```go
type IsolationWorkflowPort interface {
    // Worker
    StartIsolationWorker(ctx context.Context)

    // Scheduler
    StartScheduler(ctx context.Context)  // Removed workflowPort parameter

    // Workflow Triggers
    TriggerIsolateExpiredCustomers(ctx context.Context, gracePeriodDays int) error
    TriggerReactivateCustomer(ctx context.Context, customerID string, invoiceID string) error
}
```

---

## 🎯 **HOW IT WORKS**

### **Startup Flow**

```
1. User runs: go run cmd/main.go http

2. NewApp() creates App instance:
   - Initializes all dependencies
   - Creates domain with workflowPort

3. httpInbound() is called:
   - Initializes Gin HTTP server
   - Registers routes
   - Checks if workflowPort exists

4. If workflowPort exists:
   - Logs "Starting temporal schedulers"
   - Calls workflowPort.Billing().StartScheduler(ctx)
   - Calls workflowPort.Isolation().StartScheduler(ctx)
   - Logs "Temporal schedulers started successfully"

5. Schedulers start in background goroutines:
   - Billing scheduler: 3 goroutines (monthly, overdue, reminders)
   - Isolation scheduler: 1 goroutine (daily check)

6. HTTP server starts
   - Server runs normally
   - Schedulers run in background

7. On server shutdown:
   - HTTP server stops
   - Schedulers stop gracefully
```

---

## 📊 **SCHEDULERS RUNNING**

### **Billing Scheduler**

| Schedule | Frequency | Function |
|----------|-----------|---------|
| **Monthly Invoice Generation** | 25th of each month at 00:00 | Generate invoices for all active customers |
| **Daily Overdue Check** | Every 24 hours | Check and mark overdue invoices |
| **Daily Reminders** | Every 24 hours | Send payment reminders |

**Configuration**:
```go
// System settings (from database or .env)
invoice.auto_generate_day = 25
reminder.days_before_due = "3,1"
reminder.days_after_due = "1,3,7"
```

### **Isolation Scheduler**

| Schedule | Frequency | Function |
|----------|-----------|---------|
| **Daily Isolation Check** | Every day at 00:00 (midnight) | Find and isolate expired customers |

**Configuration**:
```go
// System settings (from database or .env)
invoice.grace_period_days = 3
```

---

## 🧪 **TESTING**

### **1. Start HTTP Server**

```bash
# Set environment variables
export OUTBOUND_WORKFLOW_DRIVER=temporal
export INBOUND_WORKFLOW_DRIVER=temporal
export WORKFLOW_HOST=temporal
export WORKFLOW_PORT=7233

# Start HTTP server
go run cmd/main.go http
```

### **2. Verify Logs**

You should see:
```
INFO Starting temporal schedulers
INFO Starting monthly invoice generation scheduler
INFO Starting daily overdue check scheduler
INFO Starting daily reminder scheduler
INFO Starting daily isolation check scheduler
INFO Temporal schedulers started successfully
INFO Next monthly invoice generation scheduled at 2026-02-25 00:00:00 +0700 WIB (in 8d23h45m12s)
INFO First isolation check scheduled at 2026-02-17 00:00:00 +0700 WIB (in 23h45m10s)
INFO [GIN-debug] Listening and serving HTTP on :8000
```

### **3. Verify Temporal UI**

Open http://localhost:8080

You should see:
- Task queues registered
- Scheduled workflows visible
- Executions visible when schedulers trigger

---

## ✅ **VERIFICATION CHECKLIST**

- [x] App.go compiles without errors
- [x] workflowOutbound() can access a.domain
- [x] Schedulers auto-start when HTTP server starts
- [x] No manual worker start needed for schedulers
- [x] Billing scheduler runs 3 schedules
- [x] Isolation scheduler runs 1 schedule
- [x] All schedulers use domain.Workflow() to trigger workflows
- [x] No interface mismatch errors

---

## 📝 **IMPORTANT NOTES**

### **Schedulers vs Workers**

**Schedulers** (auto-start):
- **DO NOT** process workflows
- **ONLY** trigger workflows at scheduled times
- Run inside HTTP server process
- Lightweight, use timers/tickers

**Workers** (manual start):
- Process and execute workflows
- Need to be started separately
- Run as independent processes
- Heavy, require Temporal connection

### **When to Use Each**

**Schedulers (auto-start)**:
- ✅ Production with HTTP server
- ✅ Automated invoice generation
- ✅ Automated isolation checks
- ✅ Automated reminders

**Workers (manual start)**:
- ✅ Manual workflow execution
- ✅ Testing workflows
- ✅ One-time workflow runs
- ✅ Processing payment webhooks

### **Recommended Setup for Production**

```bash
# Terminal 1: HTTP Server (includes schedulers)
go run cmd/main.go http

# Terminal 2: Invoice Generation Worker (optional - for manual triggers)
go run cmd/main.go workflow billing-worker

# Terminal 3: Payment Processing Worker (required - processes payment webhooks)
go run cmd/main.go workflow billing-worker

# Terminal 4: Isolation Worker (optional - for manual triggers)
go run cmd/main.go workflow isolation-worker
```

**Minimal Production Setup**:
- Terminal 1: HTTP server (schedulers auto-start)
- Terminal 2: Billing workers (invoice gen + payment processing)

---

## 🚀 **BENEFITS**

### **Before**
- ❌ Had to run separate command to start schedulers
- ❌ Schedulers tied to worker lifecycle
- ❌ Complex deployment orchestration

### **After**
- ✅ Schedulers auto-start with HTTP server
- ✅ Schedulers independent of workers
- ✅ Simple deployment (just HTTP server)
- ✅ Schedulers always run when HTTP server runs

---

## 📁 **FILES MODIFIED**

1. ✅ `internal/app.go` - Auto-start schedulers in httpInbound()
2. ✅ `internal/adapter/inbound/temporal/billing/scheduler.go` - Simplified to use domain.Workflow()
3. ✅ `internal/adapter/inbound/temporal/isolation/scheduler.go` - Simplified to use domain.Workflow()
4. ✅ `internal/adapter/inbound/temporal/billing/adapter.go` - Updated StartScheduler() signature
5. ✅ `internal/adapter/inbound/temporal/isolation/adapter.go` - Updated StartScheduler() signature
6. ✅ `internal/adapter/outbound/temporal/registry.go` - Removed workflowWrapper, updated StartScheduler methods
7. ✅ `internal/port/inbound/billing_workflow.go` - Removed workflowPort from interface
8. ✅ `internal/port/inbound/isolation_workflow.go` - Removed workflowPort from interface

---

## 🎉 **SUCCESS CRITERIA**

✅ **All criteria met:**
- [x] Mendapatkan workflowPort dari a.domain.Workflow()
- [x] Memanggil workflowPort.Billing() dan workflowPort.Isolation()
- [x] Memanggil StartScheduler(ctx) tanpa parameter workflowPort
- [x] Log "Starting temporal schedulers" saat startup
- [x] Log "Temporal schedulers started successfully" selesai
- [x] No compilation errors
- [x] Clean architecture maintained

---

**Status**: ✅ **COMPLETE & PRODUCTION READY**

Temporal schedulers sekarang otomatis berjalan saat HTTP server di-start!

---

**Last Updated**: 2026-02-16
**Implementation**: Claude Code Agent
