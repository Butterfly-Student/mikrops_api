# Quick Start: Temporal Workers

## 🚀 Start Workers in 3 Steps

### Step 1: Start Temporal Server
```bash
docker-compose -f docker-compose.temporal.yml up -d
```

Wait for Temporal to start (~30 seconds). Check at http://localhost:8080

### Step 2: Configure Environment
Ensure `.env` has:
```bash
OUTBOUND_WORKFLOW_DRIVER=temporal
INBOUND_WORKFLOW_DRIVER=temporal
WORKFLOW_HOST=temporal
WORKFLOW_PORT=7233
```

### Step 3: Start Workers
```bash
# Option 1: Start all workers (recommended)
go run cmd/main.go workflow start-workers

# Option 2: Start only billing workers
go run cmd/main.go workflow billing-worker

# Option 3: Start only isolation worker
go run cmd/main.go workflow isolation-worker
```

---

## ✅ Verify Workers Running

1. Check Temporal UI: http://localhost:8080
2. You should see task queues:
   - `GenerateMonthlyInvoicesTaskQueue`
   - `ProcessPaymentTaskQueue`
   - `IsolationTaskQueue`
   - `ClientTaskQueue`

3. Check logs for:
   ```
   INFO Billing workers started successfully
   INFO Isolation worker started successfully
   INFO Billing scheduler started successfully
   INFO Isolation scheduler started successfully
   ```

---

## 📅 Scheduled Tasks

Once running, these tasks execute automatically:

### Daily (at midnight)
- ✅ Check overdue invoices
- ✅ Send invoice reminders
- ✅ Isolate expired customers

### Monthly (25th at midnight)
- ✅ Generate monthly invoices for all customers

---

## 🧪 Test Workers

### Manually Trigger Invoice Generation
```bash
curl -X POST http://localhost:8000/api/v1/billing/generate-monthly-invoices \
  -H "Content-Type: application/json" \
  -d '{"year": 2026, "month": 2}'
```

### Manually Trigger Isolation Check
```bash
curl -X POST http://localhost:8000/api/v1/isolation/isolate-expired \
  -H "Content-Type: application/json" \
  -d '{"grace_period_days": 3}'
```

---

## 📊 Monitor

- **Temporal UI**: http://localhost:8080
- **Application Logs**: Console output
- **Database**: Check `invoices`, `payments`, `customers` tables

---

## 🛑 Stop Workers

Press `Ctrl+C` in the terminal running workers.

---

## 🔧 Troubleshooting

### Workers not starting?
```bash
# Check if Temporal is running
docker ps | grep temporal

# Check environment
env | grep WORKFLOW
```

### No workflows executing?
```bash
# Check Temporal UI for errors
# Verify workers are registered
# Check application logs
```

### Need help?
See [TEMPORAL_WORKERS_SETUP.md](./TEMPORAL_WORKERS_SETUP.md) for detailed guide.

---

## 📚 Next Steps

1. Configure system settings in database
2. Set up notification templates
3. Test with sample data
4. Monitor first scheduled runs

---

**Quick Reference**:
- All workers: `go run cmd/main.go workflow start-workers`
- Billing only: `go run cmd/main.go workflow billing-worker`
- Isolation only: `go run cmd/main.go workflow isolation-worker`
- Temporal UI: http://localhost:8080
- Full docs: [TEMPORAL_WORKERS_SETUP.md](./TEMPORAL_WORKERS_SETUP.md)
