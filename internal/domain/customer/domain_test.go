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

func TestCustomerDomain_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()

	domain := NewCustomerDomain(mockDB, mockMikrotikPort)

	t.Run("success", func(t *testing.T) {
		input := model.CustomerInput{
			CustomerCode: "CUST001",
			FullName:     "Test Customer",
			Phone:        "08123456789",
		}

		mockCustomerDB.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, c *model.Customer) error {
			assert.Equal(t, input.CustomerCode, c.CustomerCode)
			assert.Equal(t, input.FullName, c.FullName)
			assert.Equal(t, input.Phone, c.Phone)
			return nil
		})

		customer, err := domain.Create(context.Background(), input)
		assert.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, input.CustomerCode, customer.CustomerCode)
		assert.Equal(t, input.FullName, customer.FullName)
		assert.Equal(t, model.CustomerStatusPending, customer.Status)
	})

	t.Run("database error", func(t *testing.T) {
		input := model.CustomerInput{
			CustomerCode: "CUST002",
			FullName:     "Test Customer 2",
			Phone:        "08123456790",
		}

		mockCustomerDB.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("db error"))

		customer, err := domain.Create(context.Background(), input)
		assert.Error(t, err)
		assert.Nil(t, customer)
	})
}

func TestCustomerDomain_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()

	domain := NewCustomerDomain(mockDB, mockMikrotikPort)

	t.Run("success", func(t *testing.T) {
		customerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		expectedCustomer := &model.Customer{
			ID:           customerID,
			CustomerCode: "CUST001",
			FullName:     "Test Customer",
			Phone:        "08123456789",
		}

		mockCustomerDB.EXPECT().FindByID(gomock.Any(), customerID.String()).Return(expectedCustomer, nil)

		customer, err := domain.GetByID(context.Background(), customerID.String())
		assert.NoError(t, err)
		assert.Equal(t, expectedCustomer, customer)
	})

	t.Run("customer not found", func(t *testing.T) {
		customerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")

		mockCustomerDB.EXPECT().FindByID(gomock.Any(), customerID.String()).Return(nil, errors.New("not found"))

		customer, err := domain.GetByID(context.Background(), customerID.String())
		assert.Error(t, err)
		assert.Nil(t, customer)
	})
}

func TestCustomerDomain_GetByCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()

	domain := NewCustomerDomain(mockDB, mockMikrotikPort)

	t.Run("success", func(t *testing.T) {
		expectedCustomer := &model.Customer{
			ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			CustomerCode: "CUST001",
			FullName:     "Test Customer",
			Phone:        "08123456789",
		}

		mockCustomerDB.EXPECT().FindByCode(gomock.Any(), "CUST001").Return(expectedCustomer, nil)

		customer, err := domain.GetByCode(context.Background(), "CUST001")
		assert.NoError(t, err)
		assert.Equal(t, expectedCustomer, customer)
	})
}

func TestCustomerDomain_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()

	domain := NewCustomerDomain(mockDB, mockMikrotikPort)

	t.Run("success", func(t *testing.T) {
		status := model.CustomerStatusActive
		filter := &model.CustomerFilter{
			Status: &status,
		}

		expectedCustomers := []model.Customer{
			{
				ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
				CustomerCode: "CUST001",
				FullName:     "Test Customer 1",
				Status:       model.CustomerStatusActive,
			},
			{
				ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
				CustomerCode: "CUST002",
				FullName:     "Test Customer 2",
				Status:       model.CustomerStatusActive,
			},
		}

		mockCustomerDB.EXPECT().FindAll(gomock.Any(), filter).Return(expectedCustomers, nil)

		customers, err := domain.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, customers, 2)
	})
}

func TestCustomerDomain_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()

	domain := NewCustomerDomain(mockDB, mockMikrotikPort)

	t.Run("success", func(t *testing.T) {
		customerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		existingCustomer := &model.Customer{
			ID:           customerID,
			CustomerCode: "CUST001",
			FullName:     "Old Name",
			Phone:        "08123456789",
		}

		input := model.CustomerInput{
			CustomerCode: "CUST001",
			FullName:     "New Name",
			Phone:        "08123456789",
		}

		mockCustomerDB.EXPECT().FindByID(gomock.Any(), customerID.String()).Return(existingCustomer, nil)
		mockCustomerDB.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		customer, err := domain.Update(context.Background(), customerID.String(), input)
		assert.NoError(t, err)
		assert.Equal(t, input.FullName, customer.FullName)
	})
}

func TestCustomerDomain_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()

	domain := NewCustomerDomain(mockDB, mockMikrotikPort)

	t.Run("success", func(t *testing.T) {
		customerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

		mockCustomerDB.EXPECT().FindByID(gomock.Any(), customerID.String()).Return(&model.Customer{ID: customerID}, nil)
		mockCustomerDB.EXPECT().Delete(gomock.Any(), customerID.String()).Return(nil)

		err := domain.Delete(context.Background(), customerID.String())
		assert.NoError(t, err)
	})
}

func TestCustomerDomain_ChangeStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()

	domain := NewCustomerDomain(mockDB, mockMikrotikPort)

	t.Run("change to active", func(t *testing.T) {
		customerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		existingCustomer := &model.Customer{
			ID:           customerID,
			CustomerCode: "CUST001",
			FullName:     "Test Customer",
			Status:       model.CustomerStatusPending,
		}

		mockCustomerDB.EXPECT().FindByID(gomock.Any(), customerID.String()).Return(existingCustomer, nil)
		mockCustomerDB.EXPECT().UpdateStatus(gomock.Any(), customerID.String(), model.CustomerStatusActive).Return(nil)

		err := domain.ChangeStatus(context.Background(), customerID.String(), model.CustomerStatusActive)
		assert.NoError(t, err)
	})

	t.Run("change to isolated - returns error (moved to subscription domain)", func(t *testing.T) {
		customerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
		existingCustomer := &model.Customer{
			ID:           customerID,
			CustomerCode: "CUST002",
			FullName:     "Test Customer 2",
			Status:       model.CustomerStatusActive,
		}

		mockCustomerDB.EXPECT().FindByID(gomock.Any(), customerID.String()).Return(existingCustomer, nil)

		err := domain.Isolate(context.Background(), customerID.String())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Subscription")
	})
}

func TestCustomerDomain_IsActive(t *testing.T) {
	t.Run("customer is active", func(t *testing.T) {
		customer := &model.Customer{
			ID:     uuid.New(),
			Status: model.CustomerStatusActive,
		}
		assert.True(t, customer.IsActive())
	})

	t.Run("customer is not active", func(t *testing.T) {
		customer := &model.Customer{
			ID:     uuid.New(),
			Status: model.CustomerStatusPending,
		}
		assert.False(t, customer.IsActive())
	})
}

func TestCustomerDomain_IsIsolated(t *testing.T) {
	t.Run("customer is isolated", func(t *testing.T) {
		customer := &model.Customer{
			ID:     uuid.New(),
			Status: model.CustomerStatusIsolated,
		}
		assert.True(t, customer.IsIsolated())
	})

	t.Run("customer is not isolated", func(t *testing.T) {
		customer := &model.Customer{
			ID:     uuid.New(),
			Status: model.CustomerStatusActive,
		}
		assert.False(t, customer.IsIsolated())
	})
}

func TestCustomerDomain_CanBeIsolated(t *testing.T) {
	t.Run("can be isolated - active with auto_isolate true", func(t *testing.T) {
		autoIsolate := true
		customer := &model.Customer{
			ID:          uuid.New(),
			Status:      model.CustomerStatusActive,
			AutoIsolate: &autoIsolate,
		}
		assert.True(t, customer.CanBeIsolated())
	})

	t.Run("cannot be isolated - active with auto_isolate false", func(t *testing.T) {
		autoIsolate := false
		customer := &model.Customer{
			ID:          uuid.New(),
			Status:      model.CustomerStatusActive,
			AutoIsolate: &autoIsolate,
		}
		assert.False(t, customer.CanBeIsolated())
	})

	t.Run("cannot be isolated - not active", func(t *testing.T) {
		autoIsolate := true
		customer := &model.Customer{
			ID:          uuid.New(),
			Status:      model.CustomerStatusPending,
			AutoIsolate: &autoIsolate,
		}
		assert.False(t, customer.CanBeIsolated())
	})
}

func TestCustomerDomain_GetGracePeriodDays(t *testing.T) {
	t.Run("custom grace period", func(t *testing.T) {
		gracePeriod := 7
		customer := &model.Customer{
			ID:              uuid.New(),
			GracePeriodDays: &gracePeriod,
		}
		assert.Equal(t, 7, customer.GetGracePeriodDays())
	})

	t.Run("default grace period", func(t *testing.T) {
		customer := &model.Customer{
			ID: uuid.New(),
		}
		assert.Equal(t, 3, customer.GetGracePeriodDays())
	})
}

func TestCustomerInput_ToModel(t *testing.T) {
	t.Run("converts input to model", func(t *testing.T) {
		status := "active"
		input := model.CustomerInput{
			CustomerCode: "CUST001",
			FullName:     "Test Customer",
			Email:        strPtr("test@example.com"),
			Phone:        "08123456789",
			Status:       &status,
		}

		customer := input.ToModel()
		assert.Equal(t, input.CustomerCode, customer.CustomerCode)
		assert.Equal(t, input.FullName, customer.FullName)
		assert.Equal(t, input.Email, customer.Email)
		assert.Equal(t, input.Phone, customer.Phone)
		assert.Equal(t, model.CustomerStatusActive, customer.Status)
	})

	t.Run("default status is pending", func(t *testing.T) {
		input := model.CustomerInput{
			CustomerCode: "CUST001",
			FullName:     "Test Customer",
			Phone:        "08123456789",
		}

		customer := input.ToModel()
		assert.Equal(t, model.CustomerStatusPending, customer.Status)
	})
}

func strPtr(s string) *string {
	return &s
}

// Helper function for time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
