//go:build integration
// +build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	mikrotik_outbound_adapter "go-template/internal/adapter/outbound/mikrotik"
	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	"go-template/internal/domain/customer"
	"go-template/internal/model"
	"go-template/tests/helpers"
)

func TestCustomerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to setup postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	err = pgContainer.DB.AutoMigrate(
		&model.Customer{},
		&model.BandwidthProfile{},
		&model.MikrotikRouter{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate customer tables: %v", err)
	}

	dbAdapter := postgres_outbound_adapter.NewAdapter(pgContainer.DB)
	mikrotikAdapter := mikrotik_outbound_adapter.NewMikrotikClientAdapter()
	bandwidthProfileAdapter := postgres_outbound_adapter.NewBandwidthProfileAdapter(pgContainer.DB)
	customerDomain := customer.NewCustomerDomain(dbAdapter, mikrotikAdapter, bandwidthProfileAdapter)

	Convey("Test Customer Integration with PostgreSQL", t, func() {
		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Customer{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.BandwidthProfile{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.MikrotikRouter{})

		Convey("Setup test data", func() {
			// Create bandwidth profile
			isActiveProfile := true
			profile := &model.BandwidthProfile{
				Name:         "Test Customer Profile",
				Category:     "pppoe",
				PriceMonthly: 100000,
				TaxRate:      0.11,
				IsActive:     &isActiveProfile,
			}
			err := dbAdapter.BandwidthProfile().Create(profile)
			So(err, ShouldBeNil)

			// Create Mikrotik router
			routerID := uuid.New()
			router := &model.MikrotikRouter{
				ID:       routerID,
				Name:     "Customer Test Router",
				Address:  "192.168.88.1:8728",
				Username: "admin",
				Password: "admin",
				IsActive: func() *bool { b := true; return &b }(),
			}
			err = dbAdapter.Mikrotik().Create(router)
			So(err, ShouldBeNil)

			Convey("CreateCustomer creates customer successfully", func() {
				customerCode := "CUST-" + time.Now().Format("20060102150405")
				email := "customer@test.com"
				expiryDate := time.Now().AddDate(0, 1, 0)
				activationDate := time.Now()
				pppSecretName := "ppp-customer-test"
				pppSecretPassword := "test123"
				staticIP := "192.168.88.100"
				macAddress := "00:11:22:33:44:55"
				billingCycle := "monthly"
				billingDay := 1
				autoIsolate := true
				gracePeriodDays := 3

				input := model.CustomerInput{
					CustomerCode:      &customerCode,
					FullName:          "Test Customer",
					Email:             &email,
					Phone:             "08123456789",
					Address:           func() *string { s := "Test Address 123"; return &s }(),
					Status:            func() *string { s := "active"; return &s }(),
					ActivationDate:    &activationDate,
					ExpiryDate:        &expiryDate,
					RouterID:          &routerID,
					PppSecretName:     &pppSecretName,
					PppSecretPassword: &pppSecretPassword,
					StaticIP:          &staticIP,
					MACAddress:        &macAddress,
					ProfileID:         &profile.ID,
					BillingCycle:      &billingCycle,
					BillingDay:        &billingDay,
					AutoIsolate:       &autoIsolate,
					GracePeriodDays:   &gracePeriodDays,
				}

				customer, err := customerDomain.CreateCustomer(ctx, input)
				So(err, ShouldBeNil)
				So(customer, ShouldNotBeNil)
				So(customer.CustomerCode, ShouldEqual, customerCode)
				So(customer.FullName, ShouldEqual, "Test Customer")
				So(customer.Email, ShouldNotBeNil)
				So(*customer.Email, ShouldEqual, "customer@test.com")
				So(customer.Status, ShouldEqual, "active")
				So(customer.ProfileID, ShouldEqual, profile.ID)
				So(customer.RouterID, ShouldNotBeNil)
				So(*customer.RouterID, ShouldEqual, routerID)

				Convey("GetCustomer retrieves created customer", func() {
					found, err := customerDomain.GetCustomer(ctx, customer.ID.String())
					So(err, ShouldBeNil)
					So(found.ID, ShouldEqual, customer.ID)
					So(found.CustomerCode, ShouldEqual, customerCode)
				})

				Convey("ListCustomers returns customers", func() {
					filter := model.CustomerFilter{}
					customers, err := customerDomain.ListCustomers(ctx, filter)
					So(err, ShouldBeNil)
					So(len(customers), ShouldBeGreaterThanOrEqualTo, 1)
				})

				Convey("ListCustomers with profile filter", func() {
					filter := model.CustomerFilter{
						ProfileIDs: []uuid.UUID{profile.ID},
					}
					customers, err := customerDomain.ListCustomers(ctx, filter)
					So(err, ShouldBeNil)
					So(len(customers), ShouldBeGreaterThanOrEqualTo, 1)
				})

				Convey("ListCustomers with status filter", func() {
					filter := model.CustomerFilter{
						Status: []string{"active"},
					}
					customers, err := customerDomain.ListCustomers(ctx, filter)
					So(err, ShouldBeNil)
					So(len(customers), ShouldBeGreaterThanOrEqualTo, 1)
				})

				Convey("UpdateCustomer updates customer fields", func() {
					newEmail := "updated@test.com"
					newPhone := "08987654321"
					newStatus := "suspended"

					input := model.CustomerInput{
						Email: &newEmail,
						Phone: newPhone,
						Status: &newStatus,
					}

					updated, err := customerDomain.UpdateCustomer(ctx, customer.ID.String(), input)
					So(err, ShouldBeNil)
					So(updated.Email, ShouldNotBeNil)
					So(*updated.Email, ShouldEqual, "updated@test.com")
					So(updated.Phone, ShouldEqual, "08987654321")
					So(updated.Status, ShouldEqual, "suspended")
				})

				Convey("DeleteCustomer soft deletes customer", func() {
					err := customerDomain.DeleteCustomer(ctx, customer.ID.String())
					So(err, ShouldBeNil)

					// Verify soft delete
					_, err = customerDomain.GetCustomer(ctx, customer.ID.String())
					So(err, ShouldNotBeNil)
				})
			})

			Convey("CreateCustomer with duplicate code returns error", func() {
				customerCode := "DUP-" + time.Now().Format("20060102150405")
				email1 := "dup1@test.com"
				email2 := "dup2@test.com"
				expiryDate := time.Now().AddDate(0, 1, 0)

				// Create first customer
				input1 := model.CustomerInput{
					CustomerCode: &customerCode,
					FullName:     "Duplicate Customer 1",
					Email:        &email1,
					Phone:        "08123456789",
					ExpiryDate:   &expiryDate,
					RouterID:     &routerID,
					ProfileID:    &profile.ID,
				}

				_, err := customerDomain.CreateCustomer(ctx, input1)
				So(err, ShouldBeNil)

				// Try to create second customer with same code
				input2 := model.CustomerInput{
					CustomerCode: &customerCode,
					FullName:     "Duplicate Customer 2",
					Email:        &email2,
					Phone:        "08123456789",
					ExpiryDate:   &expiryDate,
					RouterID:     &routerID,
					ProfileID:    &profile.ID,
				}

				_, err = customerDomain.CreateCustomer(ctx, input2)
				So(err, ShouldNotBeNil)
			})

			Convey("CreateCustomer with duplicate email returns error", func() {
				customerCode1 := "EMAIL1-" + time.Now().Format("20060102150405")
				customerCode2 := "EMAIL2-" + time.Now().Format("20060102150405")
				email := "duplicate@test.com"
				expiryDate := time.Now().AddDate(0, 1, 0)

				// Create first customer
				input1 := model.CustomerInput{
					CustomerCode: &customerCode1,
					FullName:     "Email Duplicate 1",
					Email:        &email,
					Phone:        "08123456789",
					ExpiryDate:   &expiryDate,
					RouterID:     &routerID,
					ProfileID:    &profile.ID,
				}

				_, err := customerDomain.CreateCustomer(ctx, input1)
				So(err, ShouldBeNil)

				// Try to create second customer with same email
				input2 := model.CustomerInput{
					CustomerCode: &customerCode2,
					FullName:     "Email Duplicate 2",
					Email:        &email,
					Phone:        "08987654321",
					ExpiryDate:   &expiryDate,
					RouterID:     &routerID,
					ProfileID:    &profile.ID,
				}

				_, err = customerDomain.CreateCustomer(ctx, input2)
				So(err, ShouldNotBeNil)
			})

			Convey("CreateCustomer with invalid profile returns error", func() {
				customerCode := "INVPROF-" + time.Now().Format("20060102150405")
				expiryDate := time.Now().AddDate(0, 1, 0)
				invalidProfileID := uuid.New()

				input := model.CustomerInput{
					CustomerCode: &customerCode,
					FullName:     "Invalid Profile Customer",
					Phone:        "08123456789",
					ExpiryDate:   &expiryDate,
					RouterID:     &routerID,
					ProfileID:    &invalidProfileID,
				}

				_, err := customerDomain.CreateCustomer(ctx, input)
				So(err, ShouldNotBeNil)
			})

			Convey("GetCustomer with invalid ID returns error", func() {
				_, err := customerDomain.GetCustomer(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("UpdateCustomer with invalid ID returns error", func() {
				input := model.CustomerInput{
					FullName: "Updated Name",
				}
				_, err := customerDomain.UpdateCustomer(ctx, uuid.New().String(), input)
				So(err, ShouldNotBeNil)
			})

			Convey("DeleteCustomer with invalid ID returns error", func() {
				err := customerDomain.DeleteCustomer(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Customer Filtering", func() {
			// Create test data
			isActiveFilter := true
			profile := &model.BandwidthProfile{
				Name:         "Filter Test Profile",
				Category:     "pppoe",
				PriceMonthly: 150000,
				TaxRate:      0.11,
				IsActive:     &isActiveFilter,
			}
			err := dbAdapter.BandwidthProfile().Create(profile)
			So(err, ShouldBeNil)

			isActiveFilter2 := true
			profile2 := &model.BandwidthProfile{
				Name:         "Filter Test Profile 2",
				Category:     "pppoe",
				PriceMonthly: 200000,
				TaxRate:      0.11,
				IsActive:     &isActiveFilter2,
			}
			err = dbAdapter.BandwidthProfile().Create(profile2)
			So(err, ShouldBeNil)

			routerID := uuid.New()
			router := &model.MikrotikRouter{
				ID:       routerID,
				Name:     "Filter Test Router",
				Address:  "192.168.88.2:8728",
				Username: "admin",
				Password: "admin",
				IsActive: func() *bool { b := true; return &b }(),
			}
			err = dbAdapter.Mikrotik().Create(router)
			So(err, ShouldBeNil)

			expiryDate := time.Now().AddDate(0, 1, 0)

			// Create multiple customers
			for i := 1; i <= 3; i++ {
				customerCode := "FILT" + string(rune('0'+i)) + "-" + time.Now().Format("20060102150405")
				email := func() *string { s := "filter" + string(rune('0'+i)) + "@test.com"; return &s }()
				status := "active"
				if i == 3 {
					status = "pending"
				}

				profileID := profile.ID
				if i == 2 {
					profileID = profile2.ID
				}

				input := model.CustomerInput{
					CustomerCode: &customerCode,
					FullName:     "Filter Customer " + string(rune('0'+i)),
					Email:        email,
					Phone:        "0812345678" + string(rune('0'+i)),
					Status:       &status,
					ExpiryDate:   &expiryDate,
					RouterID:     &routerID,
					ProfileID:    &profileID,
				}

				_, err := customerDomain.CreateCustomer(ctx, input)
				So(err, ShouldBeNil)
			}

			Convey("Filter by multiple statuses", func() {
				filter := model.CustomerFilter{
					Status: []string{"active", "pending"},
				}
				customers, err := customerDomain.ListCustomers(ctx, filter)
				So(err, ShouldBeNil)
				So(len(customers), ShouldBeGreaterThanOrEqualTo, 3)
			})

			Convey("Filter by profile IDs", func() {
				filter := model.CustomerFilter{
					ProfileIDs: []uuid.UUID{profile.ID},
				}
				customers, err := customerDomain.ListCustomers(ctx, filter)
				So(err, ShouldBeNil)
				So(len(customers), ShouldBeGreaterThanOrEqualTo, 2)
			})

			Convey("Filter by search term", func() {
				searchTerm := "Filter Customer"
				filter := model.CustomerFilter{
					Search: &searchTerm,
				}
				customers, err := customerDomain.ListCustomers(ctx, filter)
				So(err, ShouldBeNil)
				So(len(customers), ShouldBeGreaterThanOrEqualTo, 3)
			})
		})

		Convey("Customer Status Changes", func() {
			// Create test data
			isActiveStatus := true
			profile := &model.BandwidthProfile{
				Name:         "Status Test Profile",
				Category:     "pppoe",
				PriceMonthly: 100000,
				TaxRate:      0.11,
				IsActive:     &isActiveStatus,
			}
			err := dbAdapter.BandwidthProfile().Create(profile)
			So(err, ShouldBeNil)

			routerID := uuid.New()
			router := &model.MikrotikRouter{
				ID:       routerID,
				Name:     "Status Test Router",
				Address:  "192.168.88.3:8728",
				Username: "admin",
				Password: "admin",
				IsActive: func() *bool { b := true; return &b }(),
			}
			err = dbAdapter.Mikrotik().Create(router)
			So(err, ShouldBeNil)

			customerCode := "STAT-" + time.Now().Format("20060102150405")
			email := "status@test.com"
			expiryDate := time.Now().AddDate(0, 1, 0)
			activationDate := time.Now()

			input := model.CustomerInput{
				CustomerCode:   &customerCode,
				FullName:       "Status Test Customer",
				Email:          &email,
				Phone:          "08123456789",
				Status:         func() *string { s := "pending"; return &s }(),
				ActivationDate: &activationDate,
				ExpiryDate:     &expiryDate,
				RouterID:       &routerID,
				ProfileID:      &profile.ID,
			}

			customer, err := customerDomain.CreateCustomer(ctx, input)
			So(err, ShouldBeNil)

			Convey("Activate customer", func() {
				activeStatus := "active"
				input := model.CustomerInput{
					Status: &activeStatus,
				}

				updated, err := customerDomain.UpdateCustomer(ctx, customer.ID.String(), input)
				So(err, ShouldBeNil)
				So(updated.Status, ShouldEqual, "active")
			})

			Convey("Suspend customer", func() {
				suspendedStatus := "suspended"
				input := model.CustomerInput{
					Status: &suspendedStatus,
				}

				updated, err := customerDomain.UpdateCustomer(ctx, customer.ID.String(), input)
				So(err, ShouldBeNil)
				So(updated.Status, ShouldEqual, "suspended")
			})

			Convey("Isolate customer", func() {
				isolatedStatus := "isolated"
				input := model.CustomerInput{
					Status: &isolatedStatus,
				}

				updated, err := customerDomain.UpdateCustomer(ctx, customer.ID.String(), input)
				So(err, ShouldBeNil)
				So(updated.Status, ShouldEqual, "isolated")
			})
		})
	})
}
