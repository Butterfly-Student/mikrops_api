package fixtures

import (
	"fmt"
	"time"

	"mikrops/internal/model"
)

// ============================================================================
// Client Test Data
// ============================================================================

type ClientTestData struct{}

func NewClientTestData() *ClientTestData {
	return &ClientTestData{}
}

func (c *ClientTestData) ValidClientInput() model.ClientInput {
	now := time.Now()
	return model.ClientInput{
		Name:      "Test Client",
		BearerKey: "test-bearer-key-" + now.Format("20060102150405"),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *ClientTestData) ValidClient() model.Client {
	return model.Client{
		ID:          1,
		ClientInput: c.ValidClientInput(),
	}
}

func (c *ClientTestData) ValidClientFilter() model.ClientFilter {
	return model.ClientFilter{
		IDs:        []int{1},
		Names:      []string{"Test Client"},
		BearerKeys: []string{"test-bearer-key"},
	}
}

func (c *ClientTestData) MultipleClients(count int) []model.Client {
	clients := make([]model.Client, count)
	now := time.Now()
	for i := 0; i < count; i++ {
		clients[i] = model.Client{
			ID: i + 1,
			ClientInput: model.ClientInput{
				Name:      "Client " + string(rune('A'+i)),
				BearerKey: "key-" + string(rune('a'+i)),
				CreatedAt: now,
				UpdatedAt: now,
			},
		}
	}
	return clients
}

func (c *ClientTestData) MultipleClientInputs(count int) []model.ClientInput {
	inputs := make([]model.ClientInput, count)
	now := time.Now()
	for i := 0; i < count; i++ {
		inputs[i] = model.ClientInput{
			Name:      "Client " + string(rune('A'+i)),
			BearerKey: "key-" + string(rune('a'+i)) + "-" + now.Format("150405"),
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	return inputs
}

// ============================================================================
// Tenant Test Data
// ============================================================================

type TenantTestData struct{}

func NewTenantTestData() *TenantTestData {
	return &TenantTestData{}
}

func (t *TenantTestData) ValidTenantInput() model.TenantInput {
	now := time.Now()
	return model.TenantInput{
		Name:             "Test Tenant",
		Slug:             "test-tenant",
		Email:            "tenant@example.com",
		Phone:            "+6281234567890",
		Address:          "Jl. Test No. 123",
		LogoURL:          "https://example.com/logo.png",
		MaxNas:           3,
		SubscriptionPlan: "basic",
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func (t *TenantTestData) ValidTenant() model.Tenant {
	return model.Tenant{
		ID:          "tenant-123",
		TenantInput: t.ValidTenantInput(),
	}
}

func (t *TenantTestData) ValidTenantFilter() model.TenantFilter {
	isActive := true
	return model.TenantFilter{
		IDs:      []string{"tenant-123"},
		Slugs:    []string{"test-tenant"},
		Emails:   []string{"tenant@example.com"},
		IsActive: &isActive,
	}
}

func (t *TenantTestData) MultipleTenants(count int) []model.Tenant {
	tenants := make([]model.Tenant, count)
	now := time.Now()
	for i := 0; i < count; i++ {
		tenants[i] = model.Tenant{
			ID: fmt.Sprintf("tenant-%d", i+1),
			TenantInput: model.TenantInput{
				Name:             fmt.Sprintf("Tenant %d", i+1),
				Slug:             fmt.Sprintf("tenant-%d", i+1),
				Email:            fmt.Sprintf("tenant%d@example.com", i+1),
				Phone:            fmt.Sprintf("+628123456%04d", i),
				MaxNas:           3,
				SubscriptionPlan: "basic",
				IsActive:         true,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		}
	}
	return tenants
}

// ============================================================================
// Staff Test Data
// ============================================================================

type StaffTestData struct{}

func NewStaffTestData() *StaffTestData {
	return &StaffTestData{}
}

func (s *StaffTestData) ValidStaffInput() model.StaffInput {
	now := time.Now()
	return model.StaffInput{
		TenantID:     "tenant-123",
		RoleID:       1,
		Email:        "staff@example.com",
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxyz1234567890",
		Password:     "password123",
		FullName:     "Test Staff",
		Phone:        "+6281234567891",
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (s *StaffTestData) ValidStaff() model.Staff {
	return model.Staff{
		ID:         "staff-123",
		StaffInput: s.ValidStaffInput(),
	}
}

func (s *StaffTestData) ValidStaffFilter() model.StaffFilter {
	isActive := true
	return model.StaffFilter{
		IDs:       []string{"staff-123"},
		TenantIDs: []string{"tenant-123"},
		Emails:    []string{"staff@example.com"},
		RoleIDs:   []int{1},
		IsActive:  &isActive,
	}
}

func (s *StaffTestData) MultipleStaffs(count int) []model.Staff {
	staffs := make([]model.Staff, count)
	now := time.Now()
	for i := 0; i < count; i++ {
		staffs[i] = model.Staff{
			ID: fmt.Sprintf("staff-%d", i+1),
			StaffInput: model.StaffInput{
				TenantID:     "tenant-123",
				RoleID:       1,
				Email:        fmt.Sprintf("staff%d@example.com", i+1),
				PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxyz1234567890",
				FullName:     fmt.Sprintf("Staff %d", i+1),
				Phone:        fmt.Sprintf("+628123457%04d", i),
				IsActive:     true,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		}
	}
	return staffs
}

// ============================================================================
// NAS Test Data
// ============================================================================

type NasTestData struct{}

func NewNasTestData() *NasTestData {
	return &NasTestData{}
}

func (n *NasTestData) ValidNasInput() model.NasInput {
	now := time.Now()
	return model.NasInput{
		TenantID:          "tenant-123",
		Name:              "Test NAS",
		Host:              "192.168.1.1",
		ApiPort:           8728,
		RestPort:          80,
		Username:          "admin",
		Password:          "password123",
		PasswordEncrypted: "encrypted_password_here",
		UseSSL:            false,
		RouterOsVersion:   "7.12",
		Identity:          "MikroTik",
		IsActive:          true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func (n *NasTestData) ValidNas() model.Nas {
	return model.Nas{
		ID:       "nas-123",
		NasInput: n.ValidNasInput(),
	}
}

func (n *NasTestData) ValidNasFilter() model.NasFilter {
	isActive := true
	return model.NasFilter{
		IDs:       []string{"nas-123"},
		TenantIDs: []string{"tenant-123"},
		Hosts:     []string{"192.168.1.1"},
		IsActive:  &isActive,
	}
}

func (n *NasTestData) MultipleNas(count int) []model.Nas {
	nases := make([]model.Nas, count)
	now := time.Now()
	for i := 0; i < count; i++ {
		nases[i] = model.Nas{
			ID: fmt.Sprintf("nas-%d", i+1),
			NasInput: model.NasInput{
				TenantID:          "tenant-123",
				Name:              fmt.Sprintf("NAS %d", i+1),
				Host:              fmt.Sprintf("192.168.1.%d", i+10),
				ApiPort:           8728,
				RestPort:          80,
				Username:          "admin",
				PasswordEncrypted: "encrypted_password",
				UseSSL:            false,
				RouterOsVersion:   "7.12",
				IsActive:          true,
				CreatedAt:         now,
				UpdatedAt:         now,
			},
		}
	}
	return nases
}

// ============================================================================
// Internet Package Test Data
// ============================================================================

type InternetPackageTestData struct{}

func NewInternetPackageTestData() *InternetPackageTestData {
	return &InternetPackageTestData{}
}

func (i *InternetPackageTestData) ValidInternetPackageInput() model.InternetPackageInput {
	now := time.Now()
	return model.InternetPackageInput{
		TenantID:      "tenant-123",
		Name:          "Paket 10Mbps",
		Description:   "Paket internet 10Mbps unlimited",
		Type:          "pppoe",
		UploadRate:    "10M",
		DownloadRate:  "10M",
		UploadBurst:   "15M",
		DownloadBurst: "15M",
		Price:         300000,
		BillingCycle:  "monthly",
		ValidityDays:  30,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (i *InternetPackageTestData) ValidInternetPackage() model.InternetPackage {
	return model.InternetPackage{
		ID:                   "package-123",
		InternetPackageInput: i.ValidInternetPackageInput(),
	}
}

func (i *InternetPackageTestData) ValidInternetPackageFilter() model.InternetPackageFilter {
	isActive := true
	return model.InternetPackageFilter{
		IDs:           []string{"package-123"},
		TenantIDs:     []string{"tenant-123"},
		Types:         []string{"pppoe"},
		Names:         []string{"Paket 10Mbps"},
		BillingCycles: []string{"monthly"},
		IsActive:      &isActive,
	}
}

func (i *InternetPackageTestData) MultipleInternetPackages(count int) []model.InternetPackage {
	packages := make([]model.InternetPackage, count)
	now := time.Now()
	speeds := []string{"5M", "10M", "20M", "50M", "100M"}
	prices := []int64{150000, 300000, 500000, 1000000, 2000000}

	for idx := 0; idx < count; idx++ {
		speedIdx := idx % len(speeds)
		packages[idx] = model.InternetPackage{
			ID: fmt.Sprintf("package-%d", idx+1),
			InternetPackageInput: model.InternetPackageInput{
				TenantID:     "tenant-123",
				Name:         fmt.Sprintf("Paket %s", speeds[speedIdx]),
				Description:  fmt.Sprintf("Paket internet %s unlimited", speeds[speedIdx]),
				Type:         "pppoe",
				UploadRate:   speeds[speedIdx],
				DownloadRate: speeds[speedIdx],
				Price:        prices[speedIdx],
				BillingCycle: "monthly",
				ValidityDays: 30,
				IsActive:     true,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		}
	}
	return packages
}

// ============================================================================
// Customer Test Data
// ============================================================================

type CustomerTestData struct{}

func NewCustomerTestData() *CustomerTestData {
	return &CustomerTestData{}
}

func (c *CustomerTestData) ValidCustomerInput() model.CustomerInput {
	now := time.Now()
	nasID := "nas-123"
	return model.CustomerInput{
		TenantID:       "tenant-123",
		FullName:       "Test Customer",
		Email:          "customer@example.com",
		Phone:          "+6281234567892",
		Address:        "Jl. Customer No. 456",
		IdentityNumber: "1234567890123456",
		Username:       "testcustomer",
		PasswordHash:   "$2a$10$abcdefghijklmnopqrstuvwxyz1234567890",
		Password:       "password123",
		PppoeUsername:  "pppoe_test",
		PppoePassword:  "pppoe_pass",
		StaticIP:       "10.10.10.10",
		NasID:          &nasID,
		IsActive:       true,
		RegisteredAt:   now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (c *CustomerTestData) ValidCustomer() model.Customer {
	return model.Customer{
		ID:            "customer-123",
		CustomerInput: c.ValidCustomerInput(),
	}
}

func (c *CustomerTestData) ValidCustomerFilter() model.CustomerFilter {
	isActive := true
	return model.CustomerFilter{
		IDs:            []string{"customer-123"},
		TenantIDs:      []string{"tenant-123"},
		Emails:         []string{"customer@example.com"},
		Usernames:      []string{"testcustomer"},
		PppoeUsernames: []string{"pppoe_test"},
		IsActive:       &isActive,
	}
}

func (c *CustomerTestData) MultipleCustomers(count int) []model.Customer {
	customers := make([]model.Customer, count)
	now := time.Now()
	nasID := "nas-123"

	for i := 0; i < count; i++ {
		customers[i] = model.Customer{
			ID: fmt.Sprintf("customer-%d", i+1),
			CustomerInput: model.CustomerInput{
				TenantID:       "tenant-123",
				FullName:       fmt.Sprintf("Customer %d", i+1),
				Email:          fmt.Sprintf("customer%d@example.com", i+1),
				Phone:          fmt.Sprintf("+628123458%04d", i),
				Address:        fmt.Sprintf("Jl. Customer %d", i+1),
				IdentityNumber: fmt.Sprintf("%016d", i+1),
				Username:       fmt.Sprintf("customer%d", i+1),
				PasswordHash:   "$2a$10$abcdefghijklmnopqrstuvwxyz1234567890",
				PppoeUsername:  fmt.Sprintf("pppoe_%d", i+1),
				PppoePassword:  "pppoe_pass",
				NasID:          &nasID,
				IsActive:       true,
				RegisteredAt:   now,
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		}
	}
	return customers
}

// ============================================================================
// Subscription Test Data
// ============================================================================

type SubscriptionTestData struct{}

func NewSubscriptionTestData() *SubscriptionTestData {
	return &SubscriptionTestData{}
}

func (s *SubscriptionTestData) ValidSubscriptionInput() model.SubscriptionInput {
	now := time.Now()
	endDate := now.AddDate(0, 1, 0)
	return model.SubscriptionInput{
		TenantID:           "tenant-123",
		CustomerID:         "customer-123",
		PackageID:          "package-123",
		NasID:              "nas-123",
		Status:             model.SubscriptionStatusActive,
		StartDate:          now,
		EndDate:            endDate,
		AutoRenew:          true,
		MikrotikQueueName:  "queue_customer_123",
		MikrotikSecretName: "pppoe_test",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func (s *SubscriptionTestData) ValidSubscription() model.Subscription {
	return model.Subscription{
		ID:                "subscription-123",
		SubscriptionInput: s.ValidSubscriptionInput(),
	}
}

func (s *SubscriptionTestData) ValidSubscriptionFilter() model.SubscriptionFilter {
	autoRenew := true
	return model.SubscriptionFilter{
		IDs:         []string{"subscription-123"},
		TenantIDs:   []string{"tenant-123"},
		CustomerIDs: []string{"customer-123"},
		PackageIDs:  []string{"package-123"},
		Statuses:    []string{model.SubscriptionStatusActive},
		AutoRenew:   &autoRenew,
	}
}

func (s *SubscriptionTestData) MultipleSubscriptions(count int) []model.Subscription {
	subscriptions := make([]model.Subscription, count)
	now := time.Now()
	endDate := now.AddDate(0, 1, 0)

	for i := 0; i < count; i++ {
		subscriptions[i] = model.Subscription{
			ID: fmt.Sprintf("subscription-%d", i+1),
			SubscriptionInput: model.SubscriptionInput{
				TenantID:           "tenant-123",
				CustomerID:         fmt.Sprintf("customer-%d", i+1),
				PackageID:          "package-123",
				NasID:              "nas-123",
				Status:             model.SubscriptionStatusActive,
				StartDate:          now,
				EndDate:            endDate,
				AutoRenew:          true,
				MikrotikQueueName:  fmt.Sprintf("queue_customer_%d", i+1),
				MikrotikSecretName: fmt.Sprintf("pppoe_%d", i+1),
				CreatedAt:          now,
				UpdatedAt:          now,
			},
		}
	}
	return subscriptions
}

// ============================================================================
// Payment Method Test Data
// ============================================================================

type PaymentMethodTestData struct{}

func NewPaymentMethodTestData() *PaymentMethodTestData {
	return &PaymentMethodTestData{}
}

func (p *PaymentMethodTestData) ValidPaymentMethodInput() model.PaymentMethodInput {
	now := time.Now()
	return model.PaymentMethodInput{
		TenantID:      "tenant-123",
		Name:          "BCA Transfer",
		Type:          "bank_transfer",
		AccountName:   "PT Test ISP",
		AccountNumber: "1234567890",
		BankName:      "Bank Central Asia",
		Instructions:  "Transfer ke rekening BCA 1234567890 a.n PT Test ISP",
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (p *PaymentMethodTestData) ValidPaymentMethod() model.PaymentMethod {
	return model.PaymentMethod{
		ID:                 "payment-method-123",
		PaymentMethodInput: p.ValidPaymentMethodInput(),
	}
}

func (p *PaymentMethodTestData) ValidPaymentMethodFilter() model.PaymentMethodFilter {
	isActive := true
	return model.PaymentMethodFilter{
		IDs:       []string{"payment-method-123"},
		TenantIDs: []string{"tenant-123"},
		Types:     []string{"bank_transfer"},
		IsActive:  &isActive,
	}
}

func (p *PaymentMethodTestData) MultiplePaymentMethods(count int) []model.PaymentMethod {
	methods := make([]model.PaymentMethod, count)
	now := time.Now()
	types := []string{"bank_transfer", "e_wallet", "qris", "cash"}
	names := []string{"BCA Transfer", "GoPay", "QRIS", "Cash"}

	for i := 0; i < count; i++ {
		typeIdx := i % len(types)
		methods[i] = model.PaymentMethod{
			ID: fmt.Sprintf("payment-method-%d", i+1),
			PaymentMethodInput: model.PaymentMethodInput{
				TenantID:      "tenant-123",
				Name:          names[typeIdx],
				Type:          types[typeIdx],
				AccountName:   "PT Test ISP",
				AccountNumber: fmt.Sprintf("%010d", i+1),
				BankName:      "Bank Central Asia",
				IsActive:      true,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
		}
	}
	return methods
}

// ============================================================================
// Invoice Test Data
// ============================================================================

type InvoiceTestData struct{}

func NewInvoiceTestData() *InvoiceTestData {
	return &InvoiceTestData{}
}

func (i *InvoiceTestData) ValidInvoiceInput() model.InvoiceInput {
	now := time.Now()
	dueDate := now.AddDate(0, 0, 7)
	periodStart := now
	periodEnd := now.AddDate(0, 1, 0)

	return model.InvoiceInput{
		TenantID:       "tenant-123",
		CustomerID:     "customer-123",
		SubscriptionID: "subscription-123",
		InvoiceNumber:  fmt.Sprintf("INV-%s", now.Format("20060102150405")),
		Amount:         300000,
		TaxAmount:      30000,
		TotalAmount:    330000,
		Status:         model.InvoiceStatusUnpaid,
		DueDate:        dueDate,
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
		Notes:          "Monthly subscription fee",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (i *InvoiceTestData) ValidInvoice() model.Invoice {
	return model.Invoice{
		ID:           "invoice-123",
		InvoiceInput: i.ValidInvoiceInput(),
	}
}

func (i *InvoiceTestData) ValidInvoiceFilter() model.InvoiceFilter {
	return model.InvoiceFilter{
		IDs:             []string{"invoice-123"},
		TenantIDs:       []string{"tenant-123"},
		CustomerIDs:     []string{"customer-123"},
		SubscriptionIDs: []string{"subscription-123"},
		Statuses:        []string{model.InvoiceStatusUnpaid},
	}
}

func (i *InvoiceTestData) MultipleInvoices(count int) []model.Invoice {
	invoices := make([]model.Invoice, count)
	now := time.Now()

	for idx := 0; idx < count; idx++ {
		dueDate := now.AddDate(0, 0, 7)
		periodStart := now.AddDate(0, -idx, 0)
		periodEnd := periodStart.AddDate(0, 1, 0)

		invoices[idx] = model.Invoice{
			ID: fmt.Sprintf("invoice-%d", idx+1),
			InvoiceInput: model.InvoiceInput{
				TenantID:       "tenant-123",
				CustomerID:     fmt.Sprintf("customer-%d", idx+1),
				SubscriptionID: fmt.Sprintf("subscription-%d", idx+1),
				InvoiceNumber:  fmt.Sprintf("INV-%s-%04d", now.Format("200601"), idx+1),
				Amount:         300000,
				TaxAmount:      30000,
				TotalAmount:    330000,
				Status:         model.InvoiceStatusUnpaid,
				DueDate:        dueDate,
				PeriodStart:    periodStart,
				PeriodEnd:      periodEnd,
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		}
	}
	return invoices
}

// ============================================================================
// Payment Test Data
// ============================================================================

type PaymentTestData struct{}

func NewPaymentTestData() *PaymentTestData {
	return &PaymentTestData{}
}

func (p *PaymentTestData) ValidPaymentInput() model.PaymentInput {
	now := time.Now()
	return model.PaymentInput{
		TenantID:        "tenant-123",
		InvoiceID:       "invoice-123",
		PaymentMethodID: "payment-method-123",
		Amount:          330000,
		PaymentDate:     now,
		ProofURL:        "https://example.com/proof.jpg",
		Status:          model.PaymentStatusPending,
		Notes:           "Payment for monthly subscription",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (p *PaymentTestData) ValidPayment() model.Payment {
	return model.Payment{
		ID:           "payment-123",
		PaymentInput: p.ValidPaymentInput(),
	}
}

func (p *PaymentTestData) ValidPaymentFilter() model.PaymentFilter {
	return model.PaymentFilter{
		IDs:              []string{"payment-123"},
		TenantIDs:        []string{"tenant-123"},
		InvoiceIDs:       []string{"invoice-123"},
		PaymentMethodIDs: []string{"payment-method-123"},
		Statuses:         []string{model.PaymentStatusPending},
	}
}

func (p *PaymentTestData) MultiplePayments(count int) []model.Payment {
	payments := make([]model.Payment, count)
	now := time.Now()
	statuses := []string{
		model.PaymentStatusPending,
		model.PaymentStatusVerified,
		model.PaymentStatusRejected,
	}

	for i := 0; i < count; i++ {
		statusIdx := i % len(statuses)
		payments[i] = model.Payment{
			ID: fmt.Sprintf("payment-%d", i+1),
			PaymentInput: model.PaymentInput{
				TenantID:        "tenant-123",
				InvoiceID:       fmt.Sprintf("invoice-%d", i+1),
				PaymentMethodID: "payment-method-123",
				Amount:          330000,
				PaymentDate:     now,
				ProofURL:        fmt.Sprintf("https://example.com/proof%d.jpg", i+1),
				Status:          statuses[statusIdx],
				CreatedAt:       now,
				UpdatedAt:       now,
			},
		}
	}
	return payments
}
