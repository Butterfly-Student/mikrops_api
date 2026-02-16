# Temporal Workers & Schedulers Setup Guide

## Overview

This document explains how to configure and run Temporal workers and schedulers for the MikroTik ISP Management System.

---

## Prerequisites

1. **Temporal Server** must be running:
   ```bash
   docker-compose -f docker-compose.temporal.yml up -d
   ```

2. **Environment Variables** must be configured in `.env`:
   ```bash
   OUTBOUND_WORKFLOW_DRIVER=temporal
   INBOUND_WORKFLOW_DRIVER=temporal
   WORKFLOW_HOST=temporal
   WORKFLOW_PORT=7233
   WORKFLOW_NAMESPACE=default
   ```

---

## Available Workers

### 1. **Billing Workers**

#### Invoice Generation Worker
- **Purpose**: Automatically generates monthly invoices for all active customers
- **Task Queue**: `GenerateMonthlyInvoicesTaskQueue`
- **Workflows**:
  - `GenerateMonthlyInvoicesWorkflow` - Main invoice generation
  - `CheckOverdueInvoicesWorkflow` - Checks and updates overdue invoices
  - `SendInvoiceRemindersWorkflow` - Sends reminder notifications
- **Activities**:
  - `CheckInvoicesGeneratedActivity` - Checks if invoices already generated
  - `GenerateInvoicesActivity` - Creates invoice records
  - `SendInvoiceNotificationsActivity` - Sends notifications
  - `FindOverdueInvoicesActivity` - Finds overdue invoices
  - `ApplyLateFeesActivity` - Applies late fees
  - `FindInvoicesDueSoonActivity` - Finds invoices due soon

#### Payment Processing Worker
- **Purpose**: Processes payment webhooks from Xendit
- **Task Queue**: `ProcessPaymentTaskQueue`
- **Workflows**:
  - `ProcessPaymentWorkflow` - Processes payment notifications
- **Activities**:
  - `ValidatePaymentWebhookActivity` - Validates webhook signature
  - `UpdateInvoiceStatusActivity` - Updates invoice to paid
  - `CreatePaymentRecordActivity` - Creates payment record
  - `RecordCashTransactionActivity` - Records cash transaction
  - `SendPaymentNotificationActivity` - Sends payment confirmation
  - `CheckCustomerReactivationActivity` - Checks if reactivation needed
  - `TriggerReactivationWorkflowActivity` - Triggers reactivation

### 2. **Isolation Worker**

#### Customer Isolation Worker
- **Purpose**: Isolates customers with expired payments and reactivates after payment
- **Task Queue**: `IsolationTaskQueue`
- **Workflows**:
  - `IsolateExpiredCustomersWorkflow` - Isolates expired customers
  - `ReactivateCustomerWorkflow` - Reactivates customers after payment
- **Activities**:
  - `FindExpiredCustomersActivity` - Finds expired customers
  - `CheckAutoIsolateActivity` - Checks auto-isolate flag
  - `IsolateCustomerActivity` - Changes profile to isolated on MikroTik
  - `SendIsolationNotificationActivity` - Sends isolation notification
  - `GetCustomerDetailsActivity` - Gets customer details
  - `GetOriginalProfileActivity` - Gets original profile
  - `ReactivateOnMikrotikActivity` - Restores original profile
  - `UpdateCustomerStatusActivity` - Updates customer status
  - `UpdateExpiryDateActivity` - Updates expiry date
  - `SendReactivationNotificationActivity` - Sends reactivation notification

### 3. **Client Worker**

#### Client Worker
- **Purpose**: Handles client-related operations
- **Task Queue**: `ClientTaskQueue`
- **Workflows**:
  - `UpsertClientWorkflow` - Upserts client data

---

## Schedulers

### 1. **Billing Scheduler**

Automatically runs the following scheduled tasks:

#### Monthly Invoice Generation
- **Schedule**: 25th of every month at 00:00 (configurable)
- **System Setting**: `invoice.auto_generate_day` (default: 25)
- **What it does**:
  - Checks if invoices already generated for current month
  - Generates invoices for all active customers
  - Sends invoice notifications

#### Daily Overdue Check
- **Schedule**: Every 24 hours
- **What it does**:
  - Finds all overdue invoices
  - Applies late fees if enabled
  - Updates invoice status to overdue

#### Daily Reminders
- **Schedule**: Every 24 hours
- **System Settings**:
  - `reminder.days_before_due` (default: "3,1")
  - `reminder.days_after_due` (default: "1,3,7")
- **What it does**:
  - Sends reminders before due date (e.g., 3 days and 1 day before)
  - Sends reminders after due date (e.g., 1, 3, 7 days after)

### 2. **Isolation Scheduler**

#### Daily Isolation Check
- **Schedule**: Every day at 00:00 (midnight)
- **System Setting**: `invoice.grace_period_days` (default: 3)
- **What it does**:
  - Finds all customers past expiry date + grace period
  - Checks if `auto_isolate` flag is true
  - Isolates customer on MikroTik (changes profile)
  - Sends isolation notification
  - Updates customer status to 'isolated'

---

## Running Workers

### Start All Workers (Recommended for Production)

```bash
go run cmd/main.go workflow start-workers
```

This starts:
- Billing Invoice Generation Worker
- Billing Payment Processing Worker
- Isolation Worker
- Client Worker
- Billing Scheduler (all cron jobs)
- Isolation Scheduler (daily check)

### Start Billing Workers Only

```bash
go run cmd/main.go workflow billing-worker
```

This starts:
- Billing Invoice Generation Worker
- Billing Payment Processing Worker
- Billing Scheduler

### Start Isolation Worker Only

```bash
go run cmd/main.go workflow isolation-worker
```

This starts:
- Isolation Worker
- Isolation Scheduler

### Start Client Worker Only

```bash
go run cmd/main.go workflow client-worker
```

---

## System Settings Configuration

### Invoice Settings

| Setting Key | Default | Description |
|-------------|---------|-------------|
| `invoice.auto_generate_day` | 25 | Day of month to generate invoices (1-28) |
| `invoice.due_days` | 7 | Number of days before invoice is due |
| `invoice.grace_period_days` | 3 | Grace period before isolation |
| `invoice.late_fee_enabled` | true | Whether to apply late fees |
| `invoice.late_fee_amount` | 50000 | Late fee amount in IDR |

### Reminder Settings

| Setting Key | Default | Description |
|-------------|---------|-------------|
| `reminder.days_before_due` | "3,1" | Days before due to send reminders |
| `reminder.days_after_due` | "1,3,7" | Days after due to send reminders |

### Payment Portal Settings

| Setting Key | Default | Description |
|-------------|---------|-------------|
| `payment_portal_url` | https://portal.example.com | Payment portal URL |

### Company Settings (for Invoices)

| Setting Key | Default | Description |
|-------------|---------|-------------|
| `company.name` | "PT Internet Provider" | Company name |
| `company.address` | "Jl. Raya No. 123" | Company address |
| `company.phone` | "021-12345678" | Company phone |
| `company.email` | "info@isp.com" | Company email |
| `company.tax_id` | "01.234.567.8-901.000" | Tax ID number |

---

## Monitoring

### Temporal Web UI

Access Temporal Web UI at: http://localhost:8080

- View running workflows
- Check workflow history
- Monitor worker status
- Debug failed workflows

### Logs

Workers and schedulers log important events:

```log
INFO Starting all temporal workers...
INFO Billing workers started successfully
INFO Isolation worker started successfully
INFO Billing scheduler started successfully
INFO   - Monthly invoice generation: Scheduled
INFO   - Daily overdue check: Scheduled
INFO   - Daily reminders: Scheduled
INFO Isolation scheduler started successfully
INFO   - Daily isolation check: Scheduled
INFO Next monthly invoice generation scheduled at 2026-02-25 00:00:00 +0700 WIB (in 23h45m12s)
INFO First isolation check scheduled at 2026-02-17 00:00:00 +0700 WIB (in 5h23m10s)
```

---

## Production Deployment

### Docker Compose

Add to your `docker-compose.yml`:

```yaml
services:
  temporal-workers:
    build: .
    command: ["./app", "workflow", "start-workers"]
    environment:
      - OUTBOUND_WORKFLOW_DRIVER=temporal
      - INBOUND_WORKFLOW_DRIVER=temporal
      - WORKFLOW_HOST=temporal
      - WORKFLOW_PORT=7233
    depends_on:
      - temporal
      - postgres
      - redis
    restart: unless-stopped
```

### Manual Process Management

Using PM2:

```bash
# Install PM2
npm install -g pm2

# Start workers
pm2 start "go run cmd/main.go workflow start-workers" --name mikrotik-workers

# Check status
pm2 status

# View logs
pm2 logs mikrotik-workers

# Restart
pm2 restart mikrotik-workers
```

Using systemd:

Create `/etc/systemd/system/mikrotik-workers.service`:

```ini
[Unit]
Description=MikroTik ISP Temporal Workers
After=network.target postgres.service temporal.service

[Service]
Type=simple
User=mikrotik
WorkingDirectory=/opt/mikrotik-mikrops
Environment="OUTBOUND_WORKFLOW_DRIVER=temporal"
Environment="INBOUND_WORKFLOW_DRIVER=temporal"
ExecStart=/opt/mikrotik-mikrops/bin/app workflow start-workers
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start
sudo systemctl enable mikrotik-workers
sudo systemctl start mikrotik-workers

# Check status
sudo systemctl status mikrotik-workers

# View logs
sudo journalctl -u mikrotik-workers -f
```

---

## Troubleshooting

### Workers Not Starting

1. Check Temporal server is running:
   ```bash
   docker ps | grep temporal
   ```

2. Check environment variables:
   ```bash
   env | grep WORKFLOW
   ```

3. Check worker logs for errors

### Workflows Not Executing

1. Check Temporal Web UI at http://localhost:8080
2. Verify task queues match between worker and workflow
3. Check if workers are registered in Temporal

### Scheduler Not Running

1. Verify workers started successfully
2. Check logs for scheduler initialization
3. Verify system settings exist in database

### Isolation Not Working

1. Check customer has `auto_isolate = true`
2. Verify `expiry_date` is past grace period
3. Check MikroTik connection
4. Verify isolation profile exists in `bandwidth_profiles` with `category = 'isolated'`

---

## Best Practices

1. **Run workers separately from HTTP server**
   - HTTP server: `go run cmd/main.go http`
   - Workers: `go run cmd/main.go workflow start-workers`

2. **Monitor Temporal UI regularly**
   - Check for failed workflows
   - Monitor workflow execution time
   - Review worker activity

3. **Configure proper grace periods**
   - Give customers enough time to pay
   - Default: 3 days after expiry date

4. **Test in development first**
   - Manually trigger workflows before relying on schedulers
   - Verify email/SMS notifications are sent
   - Check MikroTik integration works

5. **Backup system settings**
   - Document your configured values
   - Keep default settings as reference

---

## API Endpoints for Manual Triggers

You can manually trigger workflows via HTTP API:

### Generate Monthly Invoices
```http
POST /api/v1/billing/generate-monthly-invoices
Content-Type: application/json

{
  "year": 2026,
  "month": 2
}
```

### Check Overdue Invoices
```http
POST /api/v1/billing/check-overdue
Content-Type: application/json

{
  "apply_late_fees": true
}
```

### Send Reminders
```http
POST /api/v1/billing/send-reminders
Content-Type: application/json

{
  "days_before_due": [3, 1],
  "days_after_due": [1, 3, 7]
}
```

### Isolate Expired Customers
```http
POST /api/v1/isolation/isolate-expired
Content-Type: application/json

{
  "grace_period_days": 3
}
```

### Reactivate Customer
```http
POST /api/v1/isolation/reactivate/:customer_id
Content-Type: application/json

{
  "invoice_id": "uuid"
}
```

---

## Related Documentation

- [PLANING.MD](../PLANING.MD) - Original implementation plan
- [FINAL_IMPLEMENTATION_SUMMARY.md](./FINAL_IMPLEMENTATION_SUMMARY.md) - Overall project status
- [PHASE_5_COMPLETION_SUMMARY.md](./PHASE_5_COMPLETION_SUMMARY.md) - Phase 5 details
- [Temporal Documentation](https://docs.temporal.io/) - Official Temporal docs

---

## Support

For issues or questions:
1. Check Temporal Web UI for workflow history
2. Review application logs
3. Verify system settings in database
4. Check MikroTik connection

---

**Last Updated**: 2026-02-16
**Version**: 1.0.0
