# MikroTik Sync System Guide

## Overview

Sistem ini secara otomatis menyinkronkan subscription (PPPoE dan Hotspot) ke MikroTik router saat:
1. Customer registration di-approve
2. Subscription dibuat atau di-update
3. Status subscription berubah (enable/disable)

## Architecture

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────────┐
│   Registration  │────▶│  PostgreSQL  │────▶│  Subscription   │
│     Approval    │     │  (Database)  │     │   (Pending)     │
└─────────────────┘     └──────────────┘     └─────────────────┘
                                                       │
                                                       ▼
                                              ┌─────────────────┐
                                              │  RabbitMQ       │
                                              │  (mikrotik.sync)│
                                              └─────────────────┘
                                                       │
                                    ┌──────────────────┼──────────────────┐
                                    ▼                  ▼                  ▼
                            ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
                            │ PPPoE Worker │  │Hotspot Worker│  │ Other Worker │
                            └──────────────┘  └──────────────┘  └──────────────┘
                                    │                  │
                                    ▼                  ▼
                            ┌──────────────────────────────────────┐
                            │         MikroTik Router              │
                            │   (PPPoE Secrets / Hotspot Users)    │
                            └──────────────────────────────────────┘
```

## Flow Detail

### 1. Registration Approval

```go
// 1. Admin approve registration
domain.Registration().Approve(ctx, registrationID, approverID, input)

// 2. System creates:
//    - Customer (identity only)
//    - Subscription (with mt_synced=false)
//    - Publishes sync message to RabbitMQ

// 3. RabbitMQ message contains:
{
  "subscription_id": "uuid",
  "router_id": "uuid", 
  "action": "create",
  "service_type": "pppoe",
  "pppoe_data": {
    "username": "johndoe",
    "password": "randompass",
    "profile": "10Mbps",
    "rate_limit": "10M/10M"
  }
}
```

### 2. Message Processing

```go
// Worker consumes message and:
// 1. Connects to MikroTik via API
// 2. Executes RouterOS command:
//    For PPPoE: /ppp/secret/add
//    For Hotspot: /ip/hotspot/user/add
// 3. Updates subscription mt_synced=true
```

## Konfigurasi

### Environment Variables

```bash
# Database
DB_DSN=postgres://user:pass@localhost/dbname?sslmode=disable

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# MikroTik (opsional untuk test)
MIKROTIK_TEST_IP=192.168.233.1
MIKROTIK_TEST_USER=admin
MIKROTIK_TEST_PASS=r00t
```

### MikroTik Preparation

1. **Enable API Service:**
```bash
/ip service enable api
/ip service set api port=8728
```

2. **Create PPP Profile (jika belum ada):**
```bash
/ppp profile add name="10Mbps" rate-limit=10M/10M local-address=10.10.10.1
```

3. **Create Hotspot Profile (jika belum ada):**
```bash
/ip hotspot user profile add name="10Mbps" shared-users=1 rate-limit=10M/10M
```

4. **Firewall Rule (jika perlu):**
```bash
/ip firewall filter add chain=input protocol=tcp dst-port=8728 action=accept comment="Allow API"
```

## Menjalankan Sistem

### 1. Development Mode

Terminal 1 - Jalankan HTTP API:
```bash
go run main.go http
# atau
make run-http
```

Terminal 2 - Jalankan RabbitMQ Consumer:
```bash
go run main.go message mikrotik_sync
# atau  
make run-message MESSAGE_TYPE=mikrotik_sync
```

### 2. Docker Compose

```bash
# Start semua services
docker-compose up -d

# Scale worker jika perlu
docker-compose up -d --scale mikrotik-sync-worker=3
```

## Testing

### Unit Test

```bash
go test ./internal/adapter/outbound/postgres/... -v
go test ./internal/domain/registration/... -v
```

### Integration Test dengan MikroTik Nyata

```bash
# Set credentials
export MIKROTIK_TEST_IP=192.168.233.1
export MIKROTIK_TEST_USER=admin
export MIKROTIK_TEST_PASS=r00t

# Run test
./scripts/test-mikrotik.sh

# Atau langsung dengan Go
cd tests/integration
go test -v -run TestMikrotikSyncIntegration
```

### Manual Test via API

1. **Create Registration:**
```bash
curl -X POST http://localhost:8080/api/registrations \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Test Customer",
    "phone": "08123456789",
    "bandwidth_profile_id": "uuid-profile-10m"
  }'
```

2. **Approve Registration:**
```bash
curl -X POST http://localhost:8080/api/registrations/{id}/approve \
  -H "Content-Type: application/json" \
  -d '{
    "router_id": "uuid-router-1"
  }'
```

3. **Check MikroTik:**
```bash
# Di MikroTik Terminal
/ppp secret print where name="testcustomer"
```

## Monitoring

### Check RabbitMQ Queue

```bash
# Via CLI
curl -u guest:guest http://localhost:15672/api/queues

# Via Web UI
open http://localhost:15672
```

### Check Logs

```bash
# Worker logs
docker logs -f mikrotik-sync-worker

# Application logs
tail -f logs/app.log
```

### Database Queries

```sql
-- Check pending sync
SELECT id, username, service_type, mt_synced, mt_error 
FROM subscriptions 
WHERE mt_synced = false;

-- Check sync queue status
SELECT * FROM mikrotik_sync_queue WHERE status = 'pending';

-- Check failed syncs
SELECT * FROM mikrotik_sync_queue WHERE status = 'failed';
```

## Troubleshooting

### Common Issues

#### 1. Sync tidak terjadi setelah approval

**Cek:**
- RabbitMQ running?
- Worker running?
- Message masuk queue?

```bash
# Check queue
curl -u guest:guest http://localhost:15672/api/queues/%2f/mikrotik.sync

# Check worker logs
docker logs mikrotik-sync-worker
```

#### 2. Koneksi ke MikroTik gagal

**Error:** `connection refused` atau `authentication failed`

**Solusi:**
```bash
# Test koneksi manual
nc -zv 192.168.233.1 8728

# Cek API service di MikroTik
/ip service print
```

#### 3. Secret/User gagal dibuat

**Error:** `profile not found` atau `invalid configuration`

**Solusi:**
- Pastikan profile name sama persis antara database dan MikroTik
- Cek rate_limit format (harus seperti `10M/10M`)

#### 4. Message processing tapi tidak terupdate di MikroTik

**Cek:**
```bash
# Di MikroTik
/ppp secret print where comment~"registration"
/log print where topics~"pppoe"
```

### Retry Failed Syncs

Jika ada sync yang failed, sistem akan otomatis retry (max 3 attempts). Tapi jika perlu manual retry:

```bash
# Update status jadi pending lagi
UPDATE mikrotik_sync_queue SET status='pending', attempts=0 WHERE id='uuid';
```

Atau via API:
```bash
curl -X POST http://localhost:8080/api/subscriptions/{id}/resync
```

## Security Considerations

1. **Jangan expose MikroTik API ke internet**
   - Gunakan VPN atau private network
   - Restrict IP access di MikroTik firewall

2. **Gunakan user dedicated untuk API**
   - Buat user khusus: `/user add name=billing-api group=write`
   - Jangan gunakan admin untuk API

3. **Enable SSL jika possible**
   ```bash
   /ip service set api-ssl disabled=no
   ```

## Performance Tuning

### Scale Workers

```bash
# Multiple workers untuk parallel processing
docker-compose up -d --scale mikrotik-sync-worker=5
```

### Queue Priorities

Message dengan `priority=1` akan diproses lebih dulu:
```go
message := model.MikrotikSyncMessage{
    Priority: 1, // Urgent
    // ...
}
```

### Batch Processing

Untuk bulk operations (misal: migrasi banyak customer), gunakan batch:
```go
// Publish multiple messages dalam satu batch
for _, sub := range subscriptions {
    go func(s Subscription) {
        syncPort.PublishSyncMessage(message)
    }(sub)
}
```

## Backup & Recovery

### Backup MikroTik Config

```bash
/system backup save name=before-billing-sync
/export file=mikrotik-config.rsc
```

### Backup Database

```bash
pg_dump -h localhost -U postgres billing_db > backup.sql
```

## Development Tips

### Menambahkan Service Type Baru

1. Tambah const di `model/mikrotik_sync.go`
2. Update switch case di `domain/registration/domain.go`
3. Update consumer di `adapter/inbound/rabbitmq/mikrotik_sync.go`

### Testing Changes

```bash
# Run specific test
go test -v ./internal/domain/registration/... -run TestApprove

# Run dengan race detection
go test -race ./...
```
