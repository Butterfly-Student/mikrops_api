package notification

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestCreateNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockNotificationDB := mock_outbound_port.NewMockNotificationDatabasePort(ctrl)

	mockDB.EXPECT().Notification().Return(mockNotificationDB).AnyTimes()

	domain := NewNotificationDomain(mockDB, nil, nil)
	ctx := context.Background()

	t.Run("success - create notification", func(t *testing.T) {
		customerID := uuid.New()
		notifType := "email"
		recipient := "customer@example.com"
		subject := "Payment Confirmation"
		content := "Your payment has been confirmed"

		input := model.NotificationInput{
			CustomerID: &customerID,
			Type:       notifType,
			Recipient:  recipient,
			Subject:    &subject,
			Content:    content,
		}

		mockNotificationDB.EXPECT().
			Create(gomock.Any()).
			DoAndReturn(func(notification *model.Notification) error {
				notification.ID = uuid.New()
				notification.Status = "pending"
				return nil
			}).Times(1)

		result, err := domain.CreateNotification(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "pending", result.Status)
	})

	t.Run("error - database error", func(t *testing.T) {
		customerID := uuid.New()

		input := model.NotificationInput{
			CustomerID: &customerID,
			Type:       "email",
			Recipient:  "test@example.com",
			Content:    "Test content",
		}

		mockNotificationDB.EXPECT().
			Create(gomock.Any()).
			Return(errors.New("database error")).
			Times(1)

		result, err := domain.CreateNotification(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockNotificationDB := mock_outbound_port.NewMockNotificationDatabasePort(ctrl)

	mockDB.EXPECT().Notification().Return(mockNotificationDB).AnyTimes()

	domain := NewNotificationDomain(mockDB, nil, nil)
	ctx := context.Background()

	t.Run("success - get notification", func(t *testing.T) {
		notificationID := uuid.New()

		expectedNotification := &model.Notification{
			ID:        notificationID,
			Type:      "email",
			Recipient: "customer@example.com",
			Content:   "Payment confirmation",
			Status:    "sent",
		}

		mockNotificationDB.EXPECT().
			FindByID(notificationID.String()).
			Return(expectedNotification, nil).
			Times(1)

		result, err := domain.GetNotification(ctx, notificationID.String())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, notificationID, result.ID)
	})
}

func TestCreateTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockNotificationTemplateDB := mock_outbound_port.NewMockNotificationTemplateDatabasePort(ctrl)

	mockDB.EXPECT().NotificationTemplate().Return(mockNotificationTemplateDB).AnyTimes()

	domain := NewNotificationDomain(mockDB, nil, nil)
	ctx := context.Background()

	t.Run("success - create template", func(t *testing.T) {
		name := "Payment Confirmation"

		input := model.NotificationTemplateInput{
			Name:    name,
			Type:    "email",
			Subject: "Payment Confirmed - {{.CustomerName}}",
			Content: "Your payment of {{.Amount}} has been confirmed",
		}

		mockNotificationTemplateDB.EXPECT().
			Create(gomock.Any()).
			DoAndReturn(func(template *model.NotificationTemplate) error {
				template.ID = uuid.New()
				return nil
			}).Times(1)

		result, err := domain.CreateTemplate(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestSendPaymentConfirmation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)
	mockPaymentDB := mock_outbound_port.NewMockPaymentDatabasePort(ctrl)
	mockNotificationDB := mock_outbound_port.NewMockNotificationDatabasePort(ctrl)
	mockNotificationTemplateDB := mock_outbound_port.NewMockNotificationTemplateDatabasePort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()
	mockDB.EXPECT().Payment().Return(mockPaymentDB).AnyTimes()
	mockDB.EXPECT().Notification().Return(mockNotificationDB).AnyTimes()
	mockDB.EXPECT().NotificationTemplate().Return(mockNotificationTemplateDB).AnyTimes()

	domain := NewNotificationDomain(mockDB, nil, nil)
	ctx := context.Background()

	t.Run("success - send payment confirmation", func(t *testing.T) {
		customerID := uuid.New()
		paymentID := uuid.New()
		amount := 100000.0

		email := "customer@example.com"
		phone := "081234567890"

		customer := &model.Customer{
			ID:       customerID,
			FullName: "John Doe",
			Email:    &email,
			Phone:    phone,
		}

		payment := &model.Payment{
			ID:     paymentID,
			Amount: amount,
		}

		mockCustomerDB.EXPECT().
			FindByID(customerID.String()).
			Return(customer, nil).
			Times(1)

		mockPaymentDB.EXPECT().
			FindByID(paymentID.String()).
			Return(payment, nil).
			Times(1)

		mockNotificationTemplateDB.EXPECT().
			FindByName(gomock.Any()).
			Return(nil, errors.New("template not found")).
			AnyTimes()

		mockNotificationDB.EXPECT().
			Create(gomock.Any()).
			Return(nil).
			AnyTimes()

		err := domain.SendPaymentConfirmation(ctx, customerID.String(), paymentID.String(), amount)

		assert.NoError(t, err)
	})
}

func TestRetryFailedNotifications(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockNotificationDB := mock_outbound_port.NewMockNotificationDatabasePort(ctrl)

	mockDB.EXPECT().Notification().Return(mockNotificationDB).AnyTimes()

	domain := NewNotificationDomain(mockDB, nil, nil)
	ctx := context.Background()

	t.Run("success - retry failed notifications", func(t *testing.T) {
		failedNotifications := []model.Notification{
			{
				ID:        uuid.New(),
				Type:      "email",
				Recipient: "test@example.com",
				Content:   "Test content",
				Status:    "failed",
			},
		}

		mockNotificationDB.EXPECT().
			Find(gomock.Any()).
			Return(failedNotifications, nil).
			Times(1)

		result, err := domain.RetryFailedNotifications(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}
