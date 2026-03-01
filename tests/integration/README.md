# MikroTik Integration Testing

Dokumen ini menjelaskan cara melakukan integration testing dengan MikroTik router nyata.

## Prerequisites

1. **MikroTik Router** yang accessible dari network tempat test dijalankan
2. **PostgreSQL** running (via Docker atau local)
3. **RabbitMQ** running (via Docker atau local)
4. Environment variables yang sudah di-set

## Setup Environment Variables

Buat file `.env` di root project dengan konten:

```bash
# Database (untuk test container akan otomatis dibuat)
DB_DSN=postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable

# MikroTik Test Credentials
MIKROTIK_TEST_IP=192.168.233.1
MIKROTIK_TEST_USER=admin
MIKROTIK_TEST_PASS=r00t

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

## Pastikan Profile di MikroTik Tersedia

Sebelum menjalankan test, pastikan profile berikut ada di MikroTik Anda:

```bash
# Login ke MikroTik via Winbox/Terminal
# Cek PPP profiles
/ppp profile print

# Jika "default" tidak ada, buat terlebih dahulu:
/ppp profile add name=default local-address=10.10.10.1 remote-address=10.10.10.0/24

# Cek Hotspot profiles
/ip hotspot user profile print

# Jika "default" tidak ada:
/ip hotspot user profile add name=default
```

## Menjalankan Test

### 1. Start PostgreSQL dan RabbitMQ (jika belum running)

```bash
# Via Docker Compose
docker-compose up -d postgres rabbitmq

# Atau manual
docker run -d --name postgres-test -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:15
docker run -d --name rabbitmq-test -p 5672:5672 -p 15672:15672 rabbitmq:3-management
```

### 2. Run Integration Test

```bash
# Dari root project
cd tests/integration

# Run semua integration test
go test -v ./... -run TestMikrotikSyncIntegration

# Run specific test only (PPPoE only)
go test -v ./... -run TestMikrotikSyncIntegration/PPPoE

# Run E2E registration flow
go test -v ./... -run TestEndToEndRegistrationFlow
```

### 3. Run dengan timeout lebih lama (jika network lambat)

```bash
go test -v -timeout 120s ./... -run TestMikrotikSyncIntegration
```

## Verifikasi Manual (Tanpa Test Framework)

Jika Anda ingin test manual tanpa menjalankan test suite:

### 1. Build aplikasi

```bash
cd /path/to/project
go build -o mikrotik-billing main.go
```

### 2. Test koneksi MikroTik

```bash
# Set env vars
export MIKROTIK_TEST_IP=192.168.233.1
export MIKROTIK_TEST_USER=admin
export MIKROTIK_TEST_PASS=r00t

# Jalankan test connection
./mikrotik-billing test mikrotik-connection
```

### 3. Test RabbitMQ Publish/Subscribe

Terminal 1 - Start consumer:
```bash
export RABBITMQ_URL=amqp://guest:guest@localhost:5672/
./mikrotik-billing message mikrotik_sync
```

Terminal 2 - Publish test message:
```bash
# Buat test message via API atau direct ke RabbitMQ Management UI
# http://localhost:15672 (guest/guest)
# Publish ke exchange: mikrotik.sync dengan routing key: mikrotik.sync.pppoe
```

## Troubleshooting

### 1. Connection Refused ke MikroTik

**Error:** `dial tcp 192.168.233.1:8728: connect: connection refused`

**Solusi:**
```bash
# Pastikan API service aktif di MikroTik
/ip service print
/ip service enable api
/ip service set api port=8728

# Pastikan firewall tidak block
/ip firewall filter print
# Tambahkan rule jika perlu:
/ip firewall filter add chain=input protocol=tcp dst-port=8728 action=accept comment="Allow API"
```

### 2. Authentication Failed

**Error:** `cannot log in`

**Solusi:**
- Cek username/password
- Pastikan user memiliki permission full
- Cek IP services allowed address

### 3. Profile Not Found

**Error:** `profile not found`

**Solusi:**
Buat profile di MikroTik sebelum test:
```bash
/ppp profile add name=default
```

### 4. RabbitMQ Connection Failed

**Error:** `dial tcp localhost:5672: connect: connection refused`

**Solusi:**
```bash
# Start RabbitMQ
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

# Tunggu sampai ready
sleep 10
```

## Hasil Test yang Diharapkan

Jika test berhasil, Anda akan melihat output seperti:

```
=== RUN   TestMikrotikSyncIntegration
=== RUN   TestMikrotikSyncIntegration/PPPoE_Secret_Sync
=== RUN   TestMikrotikSyncIntegration/PPPoE_Secret_Sync/Create_Secret
    mikrotik_sync_test.go:85: PPPoE Secret created: test_pppoe_a1b2c3d4
=== RUN   TestMikrotikSyncIntegration/PPPoE_Secret_Sync/Update_Secret
    mikrotik_sync_test.go:96: PPPoE Secret updated: test_pppoe_a1b2c3d4
=== RUN   TestMikrotikSyncIntegration/PPPoE_Secret_Sync/Delete_Secret
    mikrotik_sync_test.go:107: PPPoE Secret deleted: test_pppoe_a1b2c3d4
=== RUN   TestMikrotikSyncIntegration/Hotspot_User_Sync
...
--- PASS: TestMikrotikSyncIntegration (5.23s)
```

## Catatan Penting

1. **Test akan membuat dan menghapus data** di MikroTik Anda. Pastikan menggunakan test username yang unik.
2. **Jalankan test hanya di lab environment**, jangan di production router.
3. **Test akan cleanup** data yang dibuat setelah selesai, tapi jika test gagal di tengah, mungkin ada sisa data yang perlu dihapus manual.

## Manual Cleanup

Jika test gagal dan data tertinggal di MikroTik:

```bash
# Hapus semua test users (yang diawali dengan "test_" atau "e2e_test_")
/ppp secret remove [find name~"^test_"]
/ip hotspot user remove [find name~"^test_"]
```
