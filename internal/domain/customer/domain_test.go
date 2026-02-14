package customer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

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
		fullName := "John Doe"
		email := "john@example.com"
		phone := "081234567890"
		profileID := uuid.New()

		input := model.CustomerInput{
			FullName:  fullName,
			Email:     &email,
			Phone:     phone,
			ProfileID: &profileID,
		}

		expectedCustomer := &model.Customer{
			ID:           uuid.New(),
			CustomerCode: "CUST-20260215-0001",
			FullName:     fullName,
			Email:        &email,
			Phone:        phone,
			Status:       "pending",
		}

		mockCustomerDB.EXPECT().
			Create(gomock.Any()).
			DoAndReturn(func(customer *model.Customer) error {
				customer.ID = expectedCustomer.ID
				customer.CustomerCode = expectedCustomer.CustomerCode
				return nil
			}).Times(1)

		result, err := domain.CreateCustomer(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, fullName, result.FullName)
		assert.Equal(t, email, *result.Email)
		assert.Equal(t, "pending", result.Status)
	})

	t.Run("error - database error", func(t *testing.T) {
		fullName := "Jane Doe"
		input := model.CustomerInput{
			FullName: fullName,
		}

		mockCustomerDB.EXPECT().
			Create(gomock.Any()).
			Return(errors.New("database error")).
			Times(1)

		result, err := domain.CreateCustomer(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create customer")
	})
}

func TestGetCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()

	domain := NewCustomerDomain(mockDB, mockMikrotikPort)
	ctx := context.Background()

	t.Run("success - get customer", func(t *testing.T) {
		customerID := uuid.New().String()
		expectedCustomer := &model.Customer{
			ID:           uuid.MustParse(customerID),
			CustomerCode: "CUST-001",
			FullName:     "John Doe",
			Status:       "active",
		}

		mockCustomerDB.EXPECT().
			FindByID(customerID).
			Return(expectedCustomer, nil).
			Times(1)

		result, err := domain.GetCustomer(ctx, customerID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "CUST-001", result.CustomerCode)
		assert.Equal(t, "John Doe", result.FullName)
	})

	t.Run("error - customer not found", func(t *testing.T) {
		customerID := uuid.New().String()

		mockCustomerDB.EXPECT().
			FindByID(customerID).
			Return(nil, errors.New("not found")).
			Times(1)

		result, err := domain.GetCustomer(ctx, customerID)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - empty customer ID", func(t *testing.T) {
		result, err := domain.GetCustomer(ctx, "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "customer ID is required")
	})
}

func TestIsolateCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)
	mockMikrotikDB := mock_outbound_port.NewMockMikrotikDatabasePort(ctrl)
	mockBandwidthProfileDB := mock_outbound_port.NewMockBandwidthProfileDatabasePort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()
	mockDB.EXPECT().Mikrotik().Return(mockMikrotikDB).AnyTimes()
	mockDB.EXPECT().BandwidthProfile().Return(mockBandwidthProfileDB).AnyTimes()

	domain := NewCustomerDomain(mockDB, mockMikrotikPort)
	ctx := context.Background()

	t.Run("success - isolate customer", func(t *testing.T) {
		customerID := uuid.New().String()
		routerID := uuid.New()
		profileID := uuid.New()
		isolatedProfileID := uuid.New()
		pppSecretName := "user001"
		pppSecretPassword := "password123"

		customer := &model.Customer{
			ID:                uuid.MustParse(customerID),
			CustomerCode:      "CUST-001",
			FullName:          "John Doe",
			Status:            "active",
			RouterID:          &routerID,
			ProfileID:         &profileID,
			PppSecretName:     &pppSecretName,
			PppSecretPassword: &pppSecretPassword,
		}

		router := &model.MikrotikRouter{
			ID:       routerID,
			Name:     "Router-1",
			Address:  "192.168.1.1",
			Username: "admin",
			Password: "admin",
		}

		isolatedProfile := &model.BandwidthProfile{
			ID:             isolatedProfileID,
			ProfileCode:    "ISO-001",
			Name:           "Isolated Profile",
			Category:       "isolated",
			PppProfileName: "isolated-profile",
			DownloadSpeed:  128,
			UploadSpeed:    64,
		}

		// Expectations
		mockCustomerDB.EXPECT().
			FindByID(customerID).
			Return(customer, nil).
			Times(1)

		mockBandwidthProfileDB.EXPECT().
			FindByCategory("isolated").
			Return([]*model.BandwidthProfile{isolatedProfile}, nil).
			Times(1)

		mockMikrotikDB.EXPECT().
			FindByID(routerID.String()).
			Return(router, nil).
			Times(2) // Once for secret update, once for firewall

		mockMikrotikPort.EXPECT().
			UpdateSecret(router, gomock.Any()).
			Return(nil).
			Times(1)

		mockMikrotikPort.EXPECT().
			AddFirewallRule(router, gomock.Any()).
			Return(nil).
			AnyTimes()

		mockCustomerDB.EXPECT().
			Update(gomock.Any()).
			DoAndReturn(func(c *model.Customer) error {
				assert.Equal(t, "isolated", c.Status)
				return nil
			}).
			Times(1)

		err := domain.IsolateCustomer(ctx, customerID)

		assert.NoError(t, err)
	})

	t.Run("error - customer not found", func(t *testing.T) {
		customerID := uuid.New().String()

		mockCustomerDB.EXPECT().
			FindByID(customerID).
			Return(nil, errors.New("not found")).
			Times(1)

		err := domain.IsolateCustomer(ctx, customerID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "customer not found")
	})

	t.Run("error - isolated profile not found", func(t *testing.T) {
		customerID := uuid.New().String()
		routerID := uuid.New()
		profileID := uuid.New()

		customer := &model.Customer{
			ID:        uuid.MustParse(customerID),
			Status:    "active",
			RouterID:  &routerID,
			ProfileID: &profileID,
		}

		mockCustomerDB.EXPECT().
			FindByID(customerID).
			Return(customer, nil).
			Times(1)

		mockBandwidthProfileDB.EXPECT().
			FindByCategory("isolated").
			Return([]*model.BandwidthProfile{}, nil).
			Times(1)

		err := domain.IsolateCustomer(ctx, customerID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "isolated profile not found")
	})
}

func TestActivateCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)
	mockMikrotikDB := mock_outbound_port.NewMockMikrotikDatabasePort(ctrl)
	mockBandwidthProfileDB := mock_outbound_port.NewMockBandwidthProfileDatabasePort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()
	mockDB.EXPECT().Mikrotik().Return(mockMikrotikDB).AnyTimes()
	mockDB.EXPECT().BandwidthProfile().Return(mockBandwidthProfileDB).AnyTimes()

	domain := NewCustomerDomain(mockDB, mockMikrotikPort)
	ctx := context.Background()

	t.Run("success - activate customer", func(t *testing.T) {
		customerID := uuid.New().String()
		routerID := uuid.New()
		profileID := uuid.New()
		pppSecretName := "user001"
		pppSecretPassword := "password123"

		profile := &model.BandwidthProfile{
			ID:             profileID,
			ProfileCode:    "PROF-001",
			Name:           "10 Mbps",
			PppProfileName: "10mbps-profile",
		}

		customer := &model.Customer{
			ID:                uuid.MustParse(customerID),
			CustomerCode:      "CUST-001",
			FullName:          "John Doe",
			Status:            "isolated",
			RouterID:          &routerID,
			ProfileID:         &profileID,
			Profile:           profile,
			PppSecretName:     &pppSecretName,
			PppSecretPassword: &pppSecretPassword,
		}

		router := &model.MikrotikRouter{
			ID:       routerID,
			Name:     "Router-1",
			Address:  "192.168.1.1",
			Username: "admin",
			Password: "admin",
		}

		// Expectations
		mockCustomerDB.EXPECT().
			FindByID(customerID).
			Return(customer, nil).
			Times(1)

		mockBandwidthProfileDB.EXPECT().
			FindByID(profileID.String()).
			Return(profile, nil).
			Times(1)

		mockMikrotikDB.EXPECT().
			FindByID(routerID.String()).
			Return(router, nil).
			Times(2) // Once for secret update, once for firewall removal

		mockMikrotikPort.EXPECT().
			UpdateSecret(router, gomock.Any()).
			Return(nil).
			Times(1)

		mockMikrotikPort.EXPECT().
			RemoveFirewallRule(router, gomock.Any()).
			Return(nil).
			AnyTimes()

		mockCustomerDB.EXPECT().
			Update(gomock.Any()).
			DoAndReturn(func(c *model.Customer) error {
				assert.Equal(t, "active", c.Status)
				return nil
			}).
			Times(1)

		err := domain.ActivateCustomer(ctx, customerID)

		assert.NoError(t, err)
	})

	t.Run("error - customer has no profile", func(t *testing.T) {
		customerID := uuid.New().String()

		customer := &model.Customer{
			ID:      uuid.MustParse(customerID),
			Status:  "isolated",
			Profile: nil,
		}

		mockCustomerDB.EXPECT().
			FindByID(customerID).
			Return(customer, nil).
			Times(1)

		err := domain.ActivateCustomer(ctx, customerID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "customer has no profile assigned")
	})
}

func TestFindExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()

	domain := NewCustomerDomain(mockDB, mockMikrotikPort)
	ctx := context.Background()

	t.Run("success - find expired customers", func(t *testing.T) {
		now := time.Now()
		expiredDate := now.Add(-5 * 24 * time.Hour)

		expiredCustomers := []model.Customer{
			{
				ID:           uuid.New(),
				CustomerCode: "CUST-001",
				FullName:     "John Doe",
				Status:       "active",
				ExpiryDate:   &expiredDate,
			},
			{
				ID:           uuid.New(),
				CustomerCode: "CUST-002",
				FullName:     "Jane Smith",
				Status:       "active",
				ExpiryDate:   &expiredDate,
			},
		}

		mockCustomerDB.EXPECT().
			FindExpired().
			Return(expiredCustomers, nil).
			Times(1)

		result, err := domain.FindExpired(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("error - database error", func(t *testing.T) {
		mockCustomerDB.EXPECT().
			FindExpired().
			Return(nil, errors.New("database error")).
			Times(1)

		result, err := domain.FindExpired(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestFindExpiringSoon(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()

	domain := NewCustomerDomain(mockDB, mockMikrotikPort)
	ctx := context.Background()

	t.Run("success - find expiring soon", func(t *testing.T) {
		days := 3
		now := time.Now()
		soonDate := now.Add(2 * 24 * time.Hour)

		customers := []model.Customer{
			{
				ID:           uuid.New(),
				CustomerCode: "CUST-001",
				FullName:     "John Doe",
				Status:       "active",
				ExpiryDate:   &soonDate,
			},
		}

		mockCustomerDB.EXPECT().
			FindExpiringSoon(days).
			Return(customers, nil).
			Times(1)

		result, err := domain.FindExpiringSoon(ctx, days)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})
}
