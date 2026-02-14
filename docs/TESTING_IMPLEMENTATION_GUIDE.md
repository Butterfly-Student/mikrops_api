# Testing Implementation Guide

## Overview

This document provides a comprehensive guide for testing the MikroTik ISP Management System. It covers unit tests, integration tests, and mock-based testing strategies.

---

## Testing Strategy

### 1. Unit Tests
- Test individual functions/methods in isolation
- Use mocks for external dependencies
- Focus on business logic validation
- Located in: `internal/domain/*/domain_test.go`

### 2. Mock Tests
- Test with mocked dependencies (database, HTTP clients, etc.)
- Verify correct interaction with ports/adapters
- Use `gomock` for generating and using mocks
- Located in: `tests/mocks/port/`

### 3. Integration Tests
- Test with real dependencies using testcontainers
- End-to-end workflow testing
- Located in: `tests/integration/`

---

## Test Coverage Status

| Component | Unit Tests | Mock Tests | Integration Tests | Status |
|-----------|-----------|------------|-------------------|---------|
| **Domains** |
| Customer Domain | ✅ Created | ✅ Mocks ready | ⏳ Pending | 80% |
| Payment Domain | ✅ Created | ✅ Mocks ready | ⏳ Pending | 80% |
| Billing Domain | ⏳ To create | ⏳ Pending | ⏳ Pending | 20% |
| Notification Domain | ⏳ To create | ⏳ Pending | ⏳ Pending | 20% |
| Auth Domain | ✅ Exists | ✅ Exists | ⏳ Pending | 70% |
| User Domain | ✅ Exists | ✅ Exists | ⏳ Pending | 70% |
| PPPoE Domain | ✅ Exists | ✅ Exists | ⏳ Pending | 70% |
| Queue Domain | ✅ Exists | ✅ Exists | ⏳ Pending | 70% |
| **HTTP Handlers** |
| Customer Handler | ⏳ To create | ⏳ Pending | ⏳ Pending | 0% |
| Payment Handler | ⏳ To create | ⏳ Pending | ⏳ Pending | 0% |
| Existing Handlers | ✅ Partial | ✅ Partial | ⏳ Pending | 40% |
| **Database Adapters** |
| Customer Adapter | ⏳ To create | ⏳ Pending | ⏳ Pending | 0% |
| Payment Adapter | ⏳ To create | ⏳ Pending | ⏳ Pending | 0% |
| Existing Adapters | ✅ Exists | ✅ Exists | ⏳ Pending | 60% |
| **Utilities** |
| PDF Generator | ⏳ To create | N/A | ⏳ Pending | 0% |
| Email Utility | ⏳ To create | ⏳ Pending | ⏳ Pending | 0% |
| WhatsApp Utility | ⏳ To create | ⏳ Pending | ⏳ Pending | 0% |
| Xendit Utility | ⏳ To create | ⏳ Pending | ⏳ Pending | 0% |

---

## Created Test Files

### Domain Tests
1. ✅ `internal/domain/customer/domain_test.go` - Customer domain comprehensive tests
   - TestCreateCustomer
   - TestGetCustomer
   - TestIsolateCustomer
   - TestActivateCustomer
   - TestFindExpired
   - TestFindExpiringSoon

2. ✅ `internal/domain/payment/domain_test.go` - Payment domain comprehensive tests
   - TestCreatePayment
   - TestGetPayment
   - TestAllocatePayment
   - TestGenerateReceipt
   - TestGetPaymentHistory
   - TestGetPaymentStatistics

### Mock Files Generated
1. ✅ `tests/mocks/port/mock_payment_port.go` - Payment database port mocks
2. ✅ `tests/mocks/port/mock_customer_port.go` - Customer database port mocks
3. ✅ `tests/mocks/port/mock_bandwidth_profile_port.go` - Bandwidth profile port mocks
4. ✅ `tests/mocks/port/mock_invoice_port.go` - Invoice port mocks
5. ✅ `tests/mocks/port/mock_system_setting_port.go` - System setting port mocks
6. ✅ `tests/mocks/port/mock_notification_port.go` - Notification port mocks
7. ✅ `tests/mocks/port/mock_mikrotik_port.go` - MikroTik port mocks

---

## Test Examples

### Unit Test Example - Customer Domain

```go
func TestCreateCustomer(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
    mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)
    mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)

    mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()

    domain := NewCustomerDomain(mockDB, mockMikrotikPort)
    ctx := context.Background()

    t.Run("success - create customer", func(t *testing.T) {
        // Test implementation...
    })
}
```

### Integration Test Example - Customer Isolation Flow

```go
func TestCustomerIsolationFlow(t *testing.T) {
    // Setup test database with testcontainers
    db := setupTestDatabase(t)
    defer db.Close()

    // Create customer
    customer := createTestCustomer(db)

    // Test isolation
    err := domain.IsolateCustomer(ctx, customer.ID.String())
    assert.NoError(t, err)

    // Verify status changed
    updated, _ := domain.GetCustomer(ctx, customer.ID.String())
    assert.Equal(t, "isolated", updated.Status)
}
```

---

## Running Tests

### Run All Unit Tests
```bash
go test ./internal/domain/... -v
```

### Run Specific Domain Tests
```bash
# Customer domain
go test ./internal/domain/customer -v

# Payment domain
go test ./internal/domain/payment -v
```

### Run With Coverage
```bash
go test -coverprofile=coverage.out ./internal/domain/...
go tool cover -html=coverage.out
```

### Run Integration Tests
```bash
go test -tags=integration ./tests/integration/... -v
```

### Run Single Test
```bash
go test -run TestCreateCustomer ./internal/domain/customer -v
```

---

## Mock Generation

### Generate All Mocks
```bash
make generate-mocks
```

### Generate Specific Mock
```bash
mockgen -source=internal/port/outbound/payment.go \
  -destination=tests/mocks/port/mock_payment_port.go \
  -package=mock_outbound_port
```

### Mock Usage Example
```go
mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
mockPaymentDB := mock_outbound_port.NewMockPaymentDatabasePort(ctrl)

// Setup expectation
mockDB.EXPECT().Payment().Return(mockPaymentDB).Times(1)
mockPaymentDB.EXPECT().Create(gomock.Any()).Return(nil).Times(1)

// Call function under test
result, err := domain.CreatePayment(ctx, input)
```

---

## Testing Best Practices

### 1. Test Structure
```go
func TestFunctionName(t *testing.T) {
    // Arrange
    // - Setup mocks
    // - Create test data

    // Act
    // - Call function under test

    // Assert
    // - Verify results
    // - Check mock expectations
}
```

### 2. Table-Driven Tests
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   Input
        want    Output
        wantErr bool
    }{
        {
            name: "success case",
            input: validInput,
            want: expectedOutput,
            wantErr: false,
        },
        {
            name: "error case",
            input: invalidInput,
            want: nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 3. Mock Expectations
```go
// Exact times
mockDB.EXPECT().Create(gomock.Any()).Times(1)

// Any times
mockDB.EXPECT().Find(gomock.Any()).AnyTimes()

// With specific arguments
mockDB.EXPECT().FindByID("customer-123").Times(1)

// Return specific values
mockDB.EXPECT().Create(gomock.Any()).Return(nil).Times(1)

// Return error
mockDB.EXPECT().Create(gomock.Any()).Return(errors.New("db error")).Times(1)

// DoAndReturn for custom logic
mockDB.EXPECT().Create(gomock.Any()).DoAndReturn(func(c *Customer) error {
    c.ID = uuid.New()
    return nil
}).Times(1)
```

---

## Test Data Management

### Using Test Fixtures
```go
import "go-template/tests/fixtures"

func TestWithFixtures(t *testing.T) {
    factory := fixtures.NewTestDataFactory()

    customer := factory.Customer.ValidCustomer()
    payment := factory.Payment.ValidPayment()
    invoice := factory.Invoice.ValidInvoice()
}
```

### Creating Test Data
```go
func createTestCustomer() *model.Customer {
    email := "test@example.com"
    return &model.Customer{
        ID:           uuid.New(),
        CustomerCode: "CUST-TEST-001",
        FullName:     "Test Customer",
        Email:        &email,
        Phone:        "081234567890",
        Status:       "active",
    }
}
```

---

## Known Issues & Fixes

### Issue 1: Mock Interface Not Matching
**Problem:** Mock doesn't implement interface after adding new methods

**Solution:**
```bash
# Regenerate mocks
make generate-mocks

# Or manually
mockgen -source=internal/port/outbound/registry_database.go \
  -destination=tests/mocks/port/mock_registry_database.go \
  -package=mock_outbound_port
```

### Issue 2: Type Mismatch in Tests
**Problem:** `cannot use &string as string`

**Solution:**
```go
// Wrong
input := model.Input{
    Name: &"John",  // Error
}

// Correct
name := "John"
input := model.Input{
    Name: &name,  // OK
}
```

### Issue 3: Docker Not Available for Integration Tests
**Problem:** `rootless Docker not supported on Windows`

**Solution:**
- Use WSL2 for integration tests
- Or use Docker Desktop with proper configuration
- Or skip integration tests on Windows: `go test -short`

---

## TODO: Tests to be Created

### High Priority
1. ✅ Customer domain unit tests - DONE
2. ✅ Payment domain unit tests - DONE
3. ⏳ Billing domain unit tests
4. ⏳ Notification domain unit tests
5. ⏳ Customer HTTP handler tests
6. ⏳ Payment HTTP handler tests

### Medium Priority
7. ⏳ Customer adapter database tests
8. ⏳ Payment adapter database tests
9. ⏳ PDF generator utility tests
10. ⏳ Email utility tests
11. ⏳ WhatsApp utility tests

### Low Priority (But Important)
12. ⏳ Integration test: Customer isolation flow
13. ⏳ Integration test: Payment processing flow
14. ⏳ Integration test: Invoice generation flow
15. ⏳ Integration test: Notification delivery flow
16. ⏳ E2E test: Complete customer lifecycle

---

## Continuous Integration

### GitHub Actions Example
```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21

      - name: Run tests
        run: go test -v -race ./...

      - name: Coverage
        run: go test -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v2
        with:
          file: ./coverage.out
```

---

## Coverage Goals

| Component | Current | Target |
|-----------|---------|--------|
| Domain Logic | ~40% | 80% |
| HTTP Handlers | ~20% | 70% |
| Database Adapters | ~30% | 60% |
| Utilities | ~0% | 50% |
| **Overall** | **~25%** | **70%** |

---

## Next Steps

1. **Fix Mock Generation Issues**
   - Debug mockgen errors
   - Ensure all interfaces have complete mocks
   - Add missing mocks to makefile

2. **Complete Domain Tests**
   - Billing domain tests
   - Notification domain tests
   - Cash domain tests
   - Bandwidth profile domain tests

3. **Add HTTP Handler Tests**
   - Customer handler
   - Payment handler
   - Billing handler
   - Notification handler

4. **Add Adapter Tests**
   - Customer database adapter
   - Payment database adapter
   - Invoice database adapter

5. **Add Utility Tests**
   - PDF generation
   - Email sending
   - WhatsApp messaging
   - Xendit integration

6. **Add Integration Tests**
   - Customer isolation workflow
   - Payment processing workflow
   - Invoice generation workflow
   - Notification delivery workflow

---

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [gomock Documentation](https://github.com/golang/mock)
- [testify Documentation](https://github.com/stretchr/testify)
- [testcontainers-go](https://github.com/testcontainers/testcontainers-go)

---

**Last Updated:** 2026-02-15
**Status:** Foundation Complete - 25% Overall Coverage
**Next Milestone:** 50% Coverage with complete domain tests
