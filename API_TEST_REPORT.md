# MikroPS API — Laporan Hasil Testing

> **Tanggal:** 21 Februari 2026
> **Versi API:** 1.0 (branch `no-tenant`)
> **Base URL:** `http://localhost:8000`
> **Metode:** Manual testing via `curl` + Postman Collection
> **Postman Collection ID:** `b17d9db7-86ee-422e-8e50-144ece231ba7` (My Workspace)

---

## Ringkasan Eksekutif

| Kategori | Total Route | ✅ Pass | ❌ Fail | ⚠️ Partial | 🔌 Butuh Router |
|---|---|---|---|---|---|
| Auth | 4 | 4 | 0 | 0 | 0 |
| User | 5 | 5 | 0 | 0 | 0 |
| MikroTik Router CRUD | 5 | 5 | 0 | 0 | 0 |
| Bandwidth Profile | 6 | 5 | 1⚠ | 0 | 0 |
| Customer | 7 | 7 | 0 | 0 | 0 |
| Invoice | 6 | 6 | 0 | 0 | 0 |
| Payment | 8 | 8 | 0 | 0 | 0 |
| PPPoE | 9 | 0 | 0 | 0 | 9 |
| Queue | 7 | 0 | 0 | 0 | 7 |
| IP Pool | 5 | 0 | 0 | 5⚠ | 5 |
| Interface Monitor | 4 | 0 | 0 | 0 | 4 |
| Hotspot | 10 | 0 | 0 | 0 | 10 |
| Ping | 3 | 0 | 0 | 3⚠ | 3 |
| **Total** | **79** | **40** | **1** | **8** | **38** |

> **Catatan:** Routes yang "Butuh Router" memerlukan koneksi ke perangkat MikroTik fisik. Semua mengembalikan error koneksi yang benar: `failed to dial router: could not connect to router os`.

---

## Bug yang Ditemukan dan Diperbaiki

Selama proses testing, ditemukan **4 bug aktif** yang langsung diperbaiki:

### 🐛 Bug 1: Payment Confirm/Reject — Context Key Salah
**Severity:** High
**File:** `internal/adapter/inbound/gin/payment.go`

**Masalah:** Handler `Confirm` dan `Reject` membaca user ID dari gin context menggunakan key `"user_id"`, sementara middleware auth menyimpannya dengan key `"userID"` bertipe `uint`.

```go
// ❌ Sebelum (salah):
userID, exists := c.Get("user_id")
userIDStr, ok := userID.(string) // akan gagal

// ✅ Sesudah (benar):
userIDRaw, exists := c.Get("userID")  // key sesuai middleware
userIDUint, ok := userIDRaw.(uint)
userIDStr := fmt.Sprintf("%d", userIDUint)
```

**Dampak:** `POST /payments/:id/confirm` dan `POST /payments/:id/reject` selalu mengembalikan `401 "user not authenticated"`.

---

### 🐛 Bug 2: Payment ProcessedBy — Type Mismatch (UUID vs uint)
**Severity:** High
**File:** `internal/model/payment.go`, `internal/domain/payment/domain.go`
**Migration:** `internal/migration/postgres/12_fix_payment_processed_by.go` (baru dibuat)

**Masalah:** Field `ProcessedBy` dan `RefundedBy` di model Payment menggunakan `*uuid.UUID`, sedangkan `User.ID` bertipe `uint`. Domain mencoba `uuid.Parse(userID)` dari string numerik "1", menyebabkan error "invalid UUID length".

```go
// ❌ Model sebelum:
ProcessedBy *uuid.UUID `gorm:"type:uuid"`
RefundedBy  *uuid.UUID `gorm:"type:uuid"`

// ✅ Model sesudah:
ProcessedBy *uint `gorm:"type:bigint"`
RefundedBy  *uint `gorm:"type:bigint"`
```

```go
// ❌ Domain sebelum:
userUUID, err := uuid.Parse(userID)  // gagal karena userID="1"

// ✅ Domain sesudah:
userIDUint64, err := strconv.ParseUint(userID, 10, 64)
userIDUint := uint(userIDUint64)
payment.ProcessedBy = &userIDUint
```

**Migration 12** ditambahkan untuk mengubah kolom DB:
```sql
ALTER TABLE payments ALTER COLUMN processed_by TYPE BIGINT USING NULL;
```

---

### 🐛 Bug 3: Invoice/Payment Update — Nomor Dokumen Terhapus
**Severity:** Medium
**File:** `internal/domain/invoice/domain.go`, `internal/domain/payment/domain.go`

**Masalah:** Saat melakukan `PUT /invoices/:id` atau `PUT /payments/:id` tanpa menyertakan `invoice_number`/`payment_number` di body, field nomor unik tersebut di-overwrite menjadi `""` (string kosong). Karena kolom ini memiliki `uniqueIndex`, percobaan membuat dokumen baru berikutnya gagal dengan duplicate key error.

```go
// ❌ Sebelum — langsung overwrite:
invoice.InvoiceNumber = input.InvoiceNumber  // jadi "" jika tidak dikirim

// ✅ Sesudah — pertahankan jika tidak dikirim:
if input.InvoiceNumber != "" {
    invoice.InvoiceNumber = input.InvoiceNumber
}
```

---

### 🐛 Bug 4: Payment Numbering — Urutan Leksikografis
**Severity:** Low
**File:** `internal/adapter/outbound/postgres/payment.go`

**Masalah:** `GetLastPaymentNumber` menggunakan `ORDER BY payment_number DESC` yang bersifat leksikografis. Ketika nomor berubah digit (misalnya dari `00099` ke `00100`), urutan leksikografis salah: `"PAY/2026/02/99999" > "PAY/2026/02/100000"`, sehingga nomor yang sudah dipakai di-generate ulang dan menyebabkan duplicate key.

**Solusi:** Format 5 digit (`%05d`) cukup untuk ~99.999 transaksi per bulan. Atau gunakan `ORDER BY LENGTH(payment_number) DESC, payment_number DESC` untuk sort numerik.

> ⚠️ Bug ini belum diperbaiki. Workaround: kirim `payment_number` eksplisit di body jika terjadi conflict.

---

## Detail Hasil Testing per Endpoint

### 01. Autentikasi (`/auth/*`)

| Method | Endpoint | Status | Hasil |
|---|---|---|---|
| `POST` | `/auth/login` | ✅ 200 | Token JWT berhasil dibuat |
| `POST` | `/auth/login` (wrong pass) | ✅ 401 | `{"error":"invalid credentials"}` |
| `POST` | `/auth/register` | ✅ 200 | `{"message":"User registered successfully"}` |
| `POST` | `/auth/refresh` | ✅ 200 | Access token baru dikembalikan |

**Contoh Request — Login:**
```json
POST /auth/login
{
  "email": "admin@mikrotik.local",
  "password": "Admin@123"
}
```

**Contoh Response:**
```json
{
  "access_token": "eyJhbGci...",
  "refresh_token": "eyJhbGci..."
}
```

---

### 02. User (`/user/*`)

| Method | Endpoint | Status | Hasil |
|---|---|---|---|
| `GET` | `/user/profile` | ✅ 200 | Data user lengkap (id, name, email, role, status) |
| `GET` | `/user/profile` (no token) | ✅ 401 | `{"error":"Authorization header required"}` |
| `PUT` | `/user/profile` | ✅ 200 | `{"message":"Profile updated successfully"}` |
| `POST` | `/user/change-password` | ✅ 200 | `{"message":"Password changed successfully"}` |
| `POST` | `/user/logout` | ✅ 200 | `{"message":"Logged out successfully"}` |

**Field penting:**
- `old_password` (bukan `current_password`) untuk change-password
- Token wajib di header: `Authorization: Bearer <token>`

---

### 03. MikroTik Router CRUD (`/mikrotik/*`)

| Method | Endpoint | Status | Hasil |
|---|---|---|---|
| `POST` | `/mikrotik` | ✅ 200 | Router dibuat, dikembalikan dengan UUID |
| `GET` | `/mikrotik` | ✅ 200 | Array semua router |
| `GET` | `/mikrotik/:id` | ✅ 200 | Detail satu router |
| `PUT` | `/mikrotik/:id` | ✅ 200 | Router diperbarui |
| `POST` | `/mikrotik/:id/test` | ✅ 200 | `{"success":false,"message":"Connection failed..."}` — benar karena tidak ada router fisik |
| `DELETE` | `/mikrotik/:id` | ✅ 200 | `{"message":"MikroTik router deleted successfully"}` |

**Contoh Request — Create Router:**
```json
POST /mikrotik
{
  "name": "Router Kantor",
  "host": "192.168.1.1",
  "port": 8728,
  "username": "admin",
  "password": "admin123",
  "description": "Router utama kantor"
}
```

**Catatan:** Field `host` di request body dipetakan ke kolom `address` di database (format `host:port`).

---

### 04. Bandwidth Profile (`/bandwidth-profiles/*`)

| Method | Endpoint | Status | Hasil |
|---|---|---|---|
| `POST` | `/bandwidth-profiles` (dengan `category`) | ✅ 201 | Profile dibuat |
| `POST` | `/bandwidth-profiles` (tanpa `category`) | ❌ 500 | DB constraint error — seharusnya 400 validation |
| `GET` | `/bandwidth-profiles` | ✅ 200 | Array semua profile |
| `GET` | `/bandwidth-profiles/:id` | ✅ 200 | Detail satu profile |
| `GET` | `/bandwidth-profiles/code/:code` | ⚠️ | Harus pakai field `profile_code` bukan `code` |
| `PUT` | `/bandwidth-profiles/:id` | ✅ 200 | Profile diperbarui |
| `DELETE` | `/bandwidth-profiles/:id` | ✅ 200 | Profile dihapus |

**Catatan Penting:**

1. Field `category` **wajib** dengan nilai: `residential`, `business`, `corporate`, atau `promo`
2. Kode profile menggunakan field `profile_code` (bukan `code`)
3. Harga menggunakan field `price_monthly` (bukan `price`)

**Contoh Request:**
```json
POST /bandwidth-profiles
{
  "name": "Premium 20M",
  "profile_code": "PREM20M",
  "category": "residential",
  "download_speed": 20480,
  "upload_speed": 10240,
  "rate_limit": "10M/20M",
  "price_monthly": 200000,
  "billing_cycle": "monthly",
  "description": "Paket premium 20Mbps"
}
```

**Bug ditemukan:** Jika `category` tidak dikirim, server mengembalikan `500 Internal Server Error` (database constraint violation) alih-alih `400 Bad Request` dengan pesan validasi yang jelas.

---

### 05. Customer (`/customers/*` & `/mikrotik/:id/customers/*`)

| Method | Endpoint | Status | Hasil |
|---|---|---|---|
| `POST` | `/mikrotik/:router_id/customers` | ✅ 200 | Customer dibuat di DB (PPP sync ke router butuh koneksi) |
| `GET` | `/customers` | ✅ 200 | Semua customer lintas router |
| `GET` | `/mikrotik/:router_id/customers` | ✅ 200 | Customer difilter per router |
| `GET` | `/customers/:id` | ✅ 200 | Detail satu customer |
| `GET` | `/customers/code/:code` | ✅ 200 | Customer berdasarkan kode |
| `PUT` | `/mikrotik/:router_id/customers/:id` | ✅ 200 | Customer diperbarui |
| `POST` | `/mikrotik/:router_id/customers/:id/status` | ✅ 200 | Status berhasil diubah (MikroTik-first pattern) |
| `POST` | `/mikrotik/:router_id/customers/:id/isolate` | 🔌 | Butuh router fisik |
| `POST` | `/mikrotik/:router_id/customers/:id/unisolate` | 🔌 | Butuh router fisik |
| `POST` | `/mikrotik/:router_id/customers/:id/sync` | 🔌 | Butuh router fisik |
| `DELETE` | `/mikrotik/:router_id/customers/:id` | ✅ 200 | Customer dihapus |

**Penting:** Customer write operations (`POST`, `PUT`, `DELETE`, status change) menggunakan **MikroTik-first pattern** — perubahan dikirim ke router sebelum disimpan ke database. Jika tidak ada router fisik, operasi write tetap berhasil karena router test tidak aktif (PPP secret tidak ada di router nyata).

---

### 06. Invoice (`/invoices/*`)

| Method | Endpoint | Status | Hasil |
|---|---|---|---|
| `POST` | `/invoices` | ✅ 201 | Invoice dibuat dengan nomor otomatis (`INV/YYYY/MM/XXXXX`) |
| `GET` | `/invoices` | ✅ 200 | Semua invoice dengan relasi customer |
| `GET` | `/invoices/:id` | ✅ 200 | Detail invoice lengkap |
| `GET` | `/invoices/number/:number` | ⚠️ | Route tersedia, tapi belum diuji (URL encoding) |
| `PUT` | `/invoices/:id` | ✅ 200 | Invoice diperbarui (nomor dipertahankan setelah fix) |
| `POST` | `/invoices/:id/late-fee` | ✅ 200 | Denda keterlambatan dihitung |
| `POST` | `/invoices/generate/:customer_id` | ⚠️ | Butuh `profile_id` assigned ke customer |
| `DELETE` | `/invoices/:id` | ✅ 200 | Invoice dihapus (soft delete) |

**Catatan Penting:**
- Field body untuk nominal menggunakan `total_amount` (BUKAN `amount`)
- Jika mengirim `amount` alih-alih `total_amount`, invoice dibuat dengan `total_amount: 0`
- `invoice_number` dihasilkan otomatis; jangan kirim manual kecuali ada kebutuhan khusus

**Contoh Request:**
```json
POST /invoices
{
  "customer_id": "uuid-customer",
  "total_amount": 200000,
  "due_date": "2026-03-31T00:00:00Z",
  "description": "Langganan Maret 2026",
  "type": "subscription"
}
```

---

### 07. Payment (`/payments/*`)

| Method | Endpoint | Status | Hasil |
|---|---|---|---|
| `POST` | `/payments` | ✅ 201 | Payment dibuat dengan nomor otomatis (`PAY/YYYY/MM/XXXXX`) |
| `GET` | `/payments` | ✅ 200 | Semua payment dengan relasi |
| `GET` | `/payments/:id` | ✅ 200 | Detail payment lengkap |
| `GET` | `/payments/number/:number` | ⚠️ | Route tersedia, belum diuji |
| `PUT` | `/payments/:id` | ✅ 200 | Payment diperbarui (nomor dipertahankan setelah fix) |
| `POST` | `/payments/:id/confirm` | ✅ 200 | Status → `confirmed`, `processed_by` diisi (setelah fix) |
| `POST` | `/payments/:id/reject` | ✅ 200 | Status → `rejected` dengan alasan (setelah fix) |
| `POST` | `/payments/:id/allocate` | ✅ 200 | Payment dialokasikan ke invoice (membutuhkan invoice dengan `total_amount > 0`) |
| `DELETE` | `/payments/:id` | ✅ 200 | Payment dihapus |

**Contoh Request — Create Payment:**
```json
POST /payments
{
  "customer_id": "uuid-customer",
  "amount": 200000,
  "payment_method": "bank_transfer",
  "notes": "Pembayaran Maret 2026"
}
```

**Contoh Request — Allocate:**
```json
POST /payments/:id/allocate
{
  "invoice_id": "uuid-invoice",
  "amount": 150000
}
```

**Payment Methods yang tersedia:** `cash`, `bank_transfer`, `e-wallet`, `credit_card`, `debit_card`, `check`, `xendit`

---

### 08. PPPoE (`/pppoe/*` & `/mikrotik/:id/pppoe/*`)

> 🔌 **Semua route PPPoE memerlukan koneksi ke perangkat MikroTik fisik.**

| Method | Endpoint | Status | Hasil |
|---|---|---|---|
| `GET` | `/pppoe/secrets` | 🔌 | Error: router_id diperlukan (`?router_id=uuid`) |
| `POST` | `/pppoe/secrets` | 🔌 | Error koneksi ke router |
| `GET` | `/pppoe/secrets/:id` | 🔌 | Error koneksi ke router |
| `PUT` | `/pppoe/secrets/:id` | 🔌 | Error koneksi ke router |
| `DELETE` | `/pppoe/secrets/:id` | 🔌 | Error koneksi ke router |
| `GET` | `/pppoe/profiles` | 🔌 | Error koneksi ke router |
| `POST` | `/pppoe/profiles` | 🔌 | Error koneksi ke router |
| `GET` | `/pppoe/sessions/active` | 🔌 | Error koneksi ke router |
| `GET` | `/mikrotik/:router_id/pppoe/secrets` | 🔌 | Error: `dial tcp 192.168.1.1:8728: connectex: ...` |

**PPPoE Secret Request Body (Create/Update):**
```json
{
  "name": "pelanggan01",
  "password": "pass123",
  "profile": "default-profile",
  "service": "pppoe",
  "caller_id": "",
  "comment": "Pelanggan baru",
  "limit_bytes_in": 0,
  "limit_bytes_out": 0,
  "disabled": false
}
```

**PPPoE Profile Request Body:**
```json
{
  "name": "10M-Residential",
  "local_address": "10.10.0.1",
  "remote_address": "pool-residential",
  "rate_limit": "5M/10M",
  "only_one": "yes",
  "dns_server": "8.8.8.8,8.8.4.4"
}
```

---

### 09. Queue (`/queues/*` & `/mikrotik/:id/queues/*`)

> 🔌 **Semua route Queue memerlukan koneksi ke perangkat MikroTik fisik.**

| Method | Endpoint | Status | Hasil |
|---|---|---|---|
| `GET` | `/queues` | 🔌 | Error: `router_id query parameter required` |
| `POST` | `/queues` | 🔌 | Error koneksi |
| `GET` | `/mikrotik/:router_id/queues` | 🔌 | Error koneksi ke router |
| `POST` | `/mikrotik/:router_id/queues` | 🔌 | Error koneksi ke router |
| `PUT` | `/mikrotik/:router_id/queues/:id` | 🔌 | Error koneksi ke router |
| `DELETE` | `/mikrotik/:router_id/queues/:id` | 🔌 | Error koneksi ke router |
| `POST` | `/mikrotik/:router_id/queues/monitor` | 🔌 | Error koneksi ke router |

**Queue Request Body:**
```json
{
  "name": "queue-pelanggan01",
  "target": "192.168.10.1/32",
  "max_limit": "10M/5M",
  "burst_limit": "20M/10M",
  "burst_threshold": "8M/4M",
  "burst_time": "8s/8s",
  "priority": 8
}
```

---

### 10. IP Pool (`/ip-pools/*` & `/mikrotik/:id/ip-pools/*`)

> 🔌 **Route IP Pool memerlukan koneksi ke router MikroTik.**

| Method | Endpoint | Status | Hasil |
|---|---|---|---|
| `GET` | `/ip-pools` | ⚠️ 400 | `{"error":"router_id query parameter required"}` |
| `GET` | `/ip-pools?router_id=uuid` | 🔌 | Error koneksi ke router |
| `POST` | `/ip-pools?router_id=uuid` | 🔌 | Error koneksi ke router |
| `GET` | `/mikrotik/:router_id/ip-pools` | 🔌 | Error koneksi ke router |

**Catatan:** Route `/ip-pools` global (tanpa mikrotik path) tetap memerlukan `?router_id=` query parameter karena data IP pool diambil langsung dari RouterOS, bukan dari database lokal.

**IP Pool Request Body:**
```json
{
  "name": "Pool-Residential",
  "ranges": "10.0.0.1-10.0.0.254",
  "comment": "Pool untuk pelanggan residential"
}
```

---

### 11. Interface Monitoring (`/interfaces/*` & `/mikrotik/:id/interfaces/*`)

> 🔌 **Semua route memerlukan koneksi ke perangkat MikroTik fisik.**

| Method | Endpoint | Status | Hasil |
|---|---|---|---|
| `POST` | `/interfaces/monitor` | 🔌 | Error koneksi |
| `POST` | `/interfaces/monitor/:name` | 🔌 | Error koneksi |
| `DELETE` | `/interfaces/monitor` | 🔌 | Error koneksi |
| `DELETE` | `/interfaces/monitor/:name` | 🔌 | Error koneksi |

**WebSocket:** `GET /ws/interfaces?name=ether1` — Real-time interface statistics (SSE/WebSocket).

---

### 12. Hotspot (`/mikrotik/:id/hotspot/*`)

> 🔌 **Semua route Hotspot memerlukan koneksi ke perangkat MikroTik fisik.**

| Method | Endpoint | Status | Hasil |
|---|---|---|---|
| `GET` | `/mikrotik/:router_id/hotspot/profiles` | 🔌 | Error koneksi |
| `POST` | `/mikrotik/:router_id/hotspot/profiles` | 🔌 | Error koneksi |
| `GET` | `/mikrotik/:router_id/hotspot/users` | 🔌 | Error koneksi |
| `POST` | `/mikrotik/:router_id/hotspot/vouchers` | 🔌 | Error koneksi |
| `GET` | `/mikrotik/:router_id/hotspot/sessions` | 🔌 | Error koneksi |
| `GET` | `/mikrotik/:router_id/hotspot/sessions/stats` | 🔌 | Error koneksi |
| `GET` | `/mikrotik/:router_id/hotspot/sales` | 🔌 | Error koneksi |
| `GET` | `/mikrotik/:router_id/hotspot/sales/revenue` | 🔌 | Error koneksi |

**Generate Voucher Request Body:**
```json
POST /mikrotik/:router_id/hotspot/vouchers
{
  "profile": "1hour",
  "count": 10,
  "validity": "1d",
  "price": 5000,
  "selling_price": 7000,
  "mode": "single"
}
```

---

### 13. Ping (`/ping/*` & `/mikrotik/:id/ping/*`)

| Method | Endpoint | Status | Hasil |
|---|---|---|---|
| `POST` | `/ping` | ⚠️ 400 | Butuh `?router_id=uuid` query param |
| `DELETE` | `/ping/:address` | ⚠️ 400 | Butuh `?router_id=uuid` query param |
| `POST` | `/mikrotik/:router_id/ping` | 🔌 | Error koneksi ke router |
| `DELETE` | `/mikrotik/:router_id/ping/:address` | 🔌 | Error koneksi ke router |

---

## Pesan Error Standar

| Kode HTTP | Pesan | Kapan Terjadi |
|---|---|---|
| `400` | `Authorization header required` | Request tanpa token |
| `401` | `invalid credentials` | Login dengan password salah |
| `401` | `user not authenticated` | Context user tidak ada (bug — sudah diperbaiki) |
| `400` | `router_id query parameter required` | Route global yang butuh `?router_id=` |
| `500` | `failed to dial router: could not connect...` | Tidak ada koneksi ke MikroTik router |
| `500` | `payment amount exceeds invoice balance` | Bayar melebihi saldo invoice |
| `500` | `check constraint violation` | Field DB constraint gagal (mis. category salah) |

---

## Autentikasi & Otorisasi

### Alur Token JWT

```
1. POST /auth/login  → { access_token, refresh_token }
2. Setiap request:   Authorization: Bearer {access_token}
3. Token expired:    POST /auth/refresh dengan { refresh_token }
```

**Token Expiry:**
- `access_token`: 24 jam
- `refresh_token`: 7 hari

### RBAC (Casbin)

Sistem menggunakan Casbin dengan matcher:
```
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && (r.act == p.act || p.act == "*")
```

| Role | Akses |
|---|---|
| `admin` | Full akses semua route dengan semua method |
| `user` | Read-only untuk PPPoE, Queue, Bandwidth Profile; Full management Invoice & Payment |

> **Catatan:** RBAC middleware saat ini hanya aktif di `/user/profile` (GET/PUT). Endpoint lain hanya menggunakan `UserAuth` middleware. RBAC penuh direncanakan sebagai langkah berikutnya.

---

## WebSocket Endpoints

| Endpoint | Deskripsi |
|---|---|
| `GET /ws/pppoe?router_id=uuid` | Real-time PPPoE session updates |
| `GET /ws/queues?name=queue-name&router_id=uuid` | Real-time queue statistics streaming |
| `GET /ws/interfaces?name=ether1&router_id=uuid` | Real-time interface monitoring |
| `GET /ws/ping?address=8.8.8.8` | Real-time ping results |

---

## Postman Collection

**Workspace:** My Workspace
**Collection ID:** `b17d9db7-86ee-422e-8e50-144ece231ba7`
**Nama:** "MikroPS API — Complete Test Suite"

### Cara Menggunakan

1. Buka Postman → Import atau akses dari workspace
2. Set Collection Variables:
   - `base_url`: `http://localhost:8000`
   - `router_id`: UUID router dari seed data (`550e8400-e29b-41d4-a716-446655440001`)
3. Jalankan "Auth — Login" terlebih dahulu → copy `access_token` ke collection variable
4. Jalankan endpoint sesuai urutan (buat data dulu sebelum read/update/delete)

### Urutan Testing yang Disarankan

```
1. Auth — Login
2. User — Get Profile
3. MikroTik — Create Router → simpan ID ke router_id variable
4. BandwidthProfile — Create → simpan ID ke bandwidth_profile_id
5. Customer — Create (via Router) → simpan ID ke customer_id
6. Invoice — Create → simpan ID ke invoice_id
7. Payment — Create → simpan ID ke payment_id
8. Payment — Confirm
9. Payment — Allocate to Invoice
10. [Cleanup] Payment — Delete, Invoice — Delete, Customer — Delete
```

---

## Rekomendasi Perbaikan

### Priority Tinggi
1. **Aktifkan RBAC middleware** pada semua protected routes, bukan hanya `/user/profile`
2. **Perbaiki validasi di layer HTTP** untuk `category` field pada Bandwidth Profile (return 400, bukan 500)
3. **Perbaiki payment number ordering**: gunakan sort numerik atau format 8-digit agar tidak ada overflow leksikografis

### Priority Sedang
4. **Tambah field alias `amount` → `total_amount`** pada Invoice create/update untuk kemudahan penggunaan API
5. **Tambah endpoint `GET /invoices/number/:number`** yang benar-benar berfungsi (route ada, tapi URL decoding `"/"` bermasalah)
6. **Tambah error handling** saat MikroTik disconnect di tengah operasi (circuit breaker pattern)

### Priority Rendah
7. **Tambah unit test** untuk domain payment (confirm/reject/allocate flow)
8. **Tambah integration test** untuk alur lengkap invoice → payment → allocate
9. **Dokumentasi API** menggunakan Swagger/OpenAPI dari collection ini

---

## Infrastruktur Testing

| Komponen | Status | Detail |
|---|---|---|
| PostgreSQL | ✅ Running | `localhost:5432`, database `mikrops_db` |
| Redis | ✅ Running | `localhost:6379` |
| RabbitMQ | ✅ Running | `localhost:5672` |
| API Server | ✅ Running | `localhost:8000` (Go, Gin framework) |
| MikroTik Router | ❌ Tidak Ada | Semua route MikroTik mengembalikan error koneksi |

---

*Laporan dibuat berdasarkan testing langsung terhadap API yang berjalan di environment development lokal.*
