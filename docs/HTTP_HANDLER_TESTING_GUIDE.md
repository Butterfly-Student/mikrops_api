# HTTP Handler Testing Guide

## 📚 Overview

Dokumentasi ini menjelaskan testing untuk HTTP handlers di layer `internal/adapter/inbound/gin`. HTTP handler tests memastikan bahwa API endpoints bekerja dengan baik dalam menerima request dan mengirim response.

---

## 📁 Lokasi Test Files

**Path:** `internal/adapter/inbound/gin/tests/`

Semua test files berada dalam **satu folder khusus** terpisah dari handler code untuk organisasi yang lebih baik.

**Struktur:**
```
internal/adapter/inbound/gin/
├── tests/                           # 📂 Semua test files
│   ├── auth_test.go
│   ├── bandwidth_profile_handler_test.go
│   ├── billing_handler_test.go
│   ├── client_test.go
│   ├── customer_handler_test.go
│   ├── middleware_test.go
│   ├── notification_handler_test.go
│   ├── payment_handler_test.go
│   ├── pppoe_test.go
│   ├── queue_test.go
│   ├── system_setting_handler_test.go
│   └── user_test.go
├── auth.go                          # Handler files
├── bandwidth_profile_handler.go
├── billing_handler.go
└── ...
```

---

## ✅ Daftar HTTP Handler Tests

### **Existing Tests** (Sudah Ada Sebelumnya)

1. ✅ **auth_test.go** - Authentication & authorization endpoints
2. ✅ **client_test.go** - Client management endpoints
3. ✅ **middleware_test.go** - Middleware functionality
4. ✅ **pppoe_test.go** - PPPoE secret management endpoints
5. ✅ **queue_test.go** - Queue management endpoints
6. ✅ **user_test.go** - User CRUD endpoints

### **New Tests** (Baru Dibuat)

7. ✅ **customer_handler_test.go** - Customer management endpoints
8. ✅ **payment_handler_test.go** - Payment processing endpoints
9. ✅ **bandwidth_profile_handler_test.go** - Bandwidth profile endpoints
10. ✅ **billing_handler_test.go** - Invoice/billing endpoints
11. ✅ **system_setting_handler_test.go** - System settings endpoints
12. ✅ **notification_handler_test.go** - Notification endpoints

**Total: 12 HTTP Handler Test Files** 🎉

---

## 🎯 Apa yang Di-Test di HTTP Handler Tests?

### 1. **Request Parsing**
- ✅ JSON body parsing
- ✅ URL parameters (`/customers/:id`)
- ✅ Query parameters (`?status=active`)
- ✅ Headers (Content-Type, Authorization)

### 2. **Response Format**
- ✅ HTTP status codes (200, 201, 400, 404, 500)
- ✅ JSON response structure
- ✅ Error messages format

### 3. **Routing**
- ✅ Correct endpoint paths
- ✅ HTTP methods (GET, POST, PUT, DELETE)

### 4. **Validation**
- ✅ Invalid JSON handling
- ✅ Missing required fields
- ✅ Invalid data types

### 5. **Error Handling**
- ✅ Database errors
- ✅ Not found errors
- ✅ Business logic errors

---

## 📝 Struktur HTTP Handler Test

```go
package gin_inbound_adapter_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/golang/mock/gomock"
    . "github.com/smartystreets/goconvey/convey"

    gin_inbound_adapter "go-template/internal/adapter/inbound/gin"
    "go-template/internal/domain"
    mock_outbound_port "go-template/tests/mocks/port"
)

func TestCustomerHandler(t *testing.T) {
    Convey("Test Customer HTTP Handler", t, func() {
        // Setup mocks
        mockCtrl := gomock.NewController(t)
        defer mockCtrl.Finish()

        mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
        mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(mockCtrl)

        mockDatabasePort.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()

        // Create domain with mocks
        dom := domain.NewDomain(mockDatabasePort, nil, nil, nil, nil, nil, nil, nil)
        handler := gin_inbound_adapter.NewCustomerHandler(dom)

        // Setup test router
        gin.SetMode(gin.TestMode)
        router := gin.New()
        router.POST("/customers", handler.CreateCustomer)
        router.GET("/customers/:id", handler.GetCustomer)

        Convey("CreateCustomer", func() {
            Convey("Success", func() {
                // Prepare request
                reqBody := map[string]interface{}{
                    "full_name": "John Doe",
                    "phone": "081234567890",
                }
                bodyBytes, _ := json.Marshal(reqBody)
                req := httptest.NewRequest("POST", "/customers", bytes.NewBuffer(bodyBytes))
                req.Header.Set("Content-Type", "application/json")
                w := httptest.NewRecorder()

                // Setup mock expectations
                mockDatabasePort.EXPECT().
                    DoInTransaction(gomock.Any()).
                    Return(&model.Customer{...}, nil).
                    Times(1)

                // Execute request
                router.ServeHTTP(w, req)

                // Assert response
                So(w.Code, ShouldEqual, http.StatusCreated)
            })

            Convey("Invalid JSON", func() {
                req := httptest.NewRequest("POST", "/customers", bytes.NewBuffer([]byte("invalid")))
                w := httptest.NewRecorder()

                router.ServeHTTP(w, req)

                So(w.Code, ShouldEqual, http.StatusBadRequest)
            })
        })
    })
}
```

---

## 🔍 Detail Test Coverage Per Handler

### 1. **Customer Handler Test**
**File:** `customer_handler_test.go`

**Endpoints Tested:**
- ✅ `POST /customers` - Create customer
- ✅ `GET /customers/:id` - Get customer by ID
- ✅ `GET /customers` - List customers with filters
- ✅ `PUT /customers/:id` - Update customer
- ✅ `DELETE /customers/:id` - Delete customer

**Test Cases:**
- Success scenarios
- Invalid JSON
- Not found errors
- Database errors

---

### 2. **Payment Handler Test**
**File:** `payment_handler_test.go`

**Endpoints Tested:**
- ✅ `POST /payments` - Create payment
- ✅ `GET /payments/:id` - Get payment by ID
- ✅ `GET /payments` - List payments
- ✅ `POST /payments/:id/confirm` - Confirm payment
- ✅ `POST /payments/:id/reject` - Reject payment

**Test Cases:**
- Payment creation with customer validation
- Payment confirmation flow
- Payment rejection with reason
- Invalid payment scenarios

---

### 3. **Bandwidth Profile Handler Test**
**File:** `bandwidth_profile_handler_test.go`

**Endpoints Tested:**
- ✅ `POST /profiles` - Create bandwidth profile
- ✅ `GET /profiles/:id` - Get profile by ID
- ✅ `GET /profiles` - List all profiles
- ✅ `PUT /profiles/:id` - Update profile
- ✅ `DELETE /profiles/:id` - Delete profile

**Test Cases:**
- Profile creation with speed validation
- Profile retrieval and listing
- Not found scenarios

---

### 4. **Billing Handler Test**
**File:** `billing_handler_test.go`

**Endpoints Tested:**
- ✅ `POST /invoices` - Create invoice
- ✅ `GET /invoices/:id` - Get invoice by ID
- ✅ `GET /invoices` - List invoices

**Test Cases:**
- Invoice creation with items
- Customer validation
- Invoice not found
- List with filters

---

### 5. **System Setting Handler Test**
**File:** `system_setting_handler_test.go`

**Endpoints Tested:**
- ✅ `GET /settings` - List all settings
- ✅ `GET /settings/:key` - Get setting by key
- ✅ `PUT /settings/:key` - Update setting

**Test Cases:**
- Get setting by key
- Setting not found
- Update setting value
- List all settings

---

### 6. **Notification Handler Test**
**File:** `notification_handler_test.go`

**Endpoints Tested:**
- ✅ `POST /notifications` - Create notification
- ✅ `GET /notifications/:id` - Get notification by ID
- ✅ `GET /notifications` - List notifications

**Test Cases:**
- Create email notification
- Create WhatsApp notification
- Notification not found
- List with status filter

---

## 🚀 Cara Menjalankan HTTP Handler Tests

### **Test All Handlers**
```bash
go test ./internal/adapter/inbound/gin/tests/... -v
```

### **Test Specific Handler**
```bash
# Customer handler
go test -v -run TestCustomerHandler ./internal/adapter/inbound/gin/tests

# Payment handler
go test -v -run TestPaymentHandler ./internal/adapter/inbound/gin/tests

# Auth handler
go test -v -run TestAuthAdapter ./internal/adapter/inbound/gin/tests
```

### **Test dengan Coverage**
```bash
go test -coverprofile=coverage_http.out ./internal/adapter/inbound/gin/tests/...
go tool cover -html=coverage_http.out -o coverage_http.html
```

### **Test Parallel (Faster)**
```bash
go test -v -parallel 4 ./internal/adapter/inbound/gin/tests/...
```

---

## 🔄 HTTP Handler Test vs Integration Test

| Aspek | HTTP Handler Test | Integration Test |
|-------|------------------|------------------|
| **Location** | `internal/adapter/inbound/gin/tests/` | `tests/integration/` |
| **Dependencies** | Mocked (domain layer) | Real (database, etc.) |
| **Speed** | ⚡ Fast (ms) | 🐌 Slower (seconds) |
| **Database** | ❌ Mock | ✅ Real PostgreSQL |
| **Scope** | HTTP layer only | Full E2E flow |
| **Purpose** | Test API interface | Test full workflow |

---

## 💡 Best Practices

### ✅ **DO:**
- Test all HTTP methods (GET, POST, PUT, DELETE)
- Test success and error scenarios
- Test invalid JSON inputs
- Test missing required fields
- Use `httptest.NewRecorder()` for responses
- Use `httptest.NewRequest()` for requests
- Mock domain layer dependencies
- Assert HTTP status codes
- Verify response JSON structure

### ❌ **DON'T:**
- Don't use real database in handler tests
- Don't test business logic (that's domain test)
- Don't skip error scenarios
- Don't hardcode URLs (use variables)
- Don't ignore response validation

---

## 📊 Test Statistics

```
Total Handler Test Files:  12
Total Test Cases:         ~80+
Average Tests per File:    6-8
Test Coverage Target:      80%
```

---

## 🎯 Next Steps

### **For New Handlers:**
1. Create handler file: `internal/adapter/inbound/gin/new_handler.go`
2. Create test file: `internal/adapter/inbound/gin/tests/new_handler_test.go`
3. Follow existing test structure
4. Test all CRUD operations
5. Test error scenarios
6. Run tests: `go test -v -run TestNewHandler ./internal/adapter/inbound/gin/tests`

### **To Add Tests:**
```bash
# Copy existing test as template
cp internal/adapter/inbound/gin/tests/customer_handler_test.go \
   internal/adapter/inbound/gin/tests/new_handler_test.go

# Edit and customize for your handler
# Run tests
go test -v ./internal/adapter/inbound/gin/tests/...
```

---

## 🔗 Related Documentation

- [Testing Implementation Guide](./TESTING_IMPLEMENTATION_GUIDE.md)
- [Domain Testing Guide](./DOMAIN_TESTING_GUIDE.md)
- [Integration Testing Guide](../tests/integration/README.md)
- [API Documentation](./API_DOCUMENTATION.md)

---

**Last Updated:** 2026-02-15
**Status:** ✅ Complete - 12 handler test files created
**Coverage:** ~70-80% of HTTP endpoints

---

## 📞 Support

Jika ada pertanyaan tentang HTTP handler testing:
1. Check existing test files sebagai referensi
2. Follow the test structure pattern
3. Ensure all CRUD operations are tested
4. Don't forget error scenarios!

**Happy Testing! 🚀**
