package payment_method_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	mock_outbound_port "mikrops/tests/mocks/port"
)

func TestPaymentMethod(t *testing.T) {
	Convey("Test Payment Method", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockPaymentMethodDatabasePort := mock_outbound_port.NewMockPaymentMethodDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().PaymentMethod().Return(mockPaymentMethodDatabasePort).AnyTimes()

		paymentMethodDomain := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)

		input := model.PaymentMethodInput{
			TenantID:      "tenant-123",
			Name:          "Bank Transfer",
			AccountNumber: "1234567890",
			AccountName:   "Test Account",
			IsActive:      true,
		}

		output := model.PaymentMethod{
			ID:                 "pm-123",
			PaymentMethodInput: input,
		}

		filter := model.PaymentMethodFilter{IDs: []string{"pm-123"}}

		Convey("Create", func() {
			Convey("Success", func() {
				mockPaymentMethodDatabasePort.EXPECT().Create(gomock.Any()).Return(output, nil).Times(1)

				result, err := paymentMethodDomain.PaymentMethod().Create(context.Background(), input)
				So(err, ShouldBeNil)
				So(result.Name, ShouldEqual, "Bank Transfer")
			})
		})

		Convey("FindByFilter", func() {
			Convey("Filter is empty", func() {
				_, err := paymentMethodDomain.PaymentMethod().FindByFilter(context.Background(), model.PaymentMethodFilter{})
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockPaymentMethodDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.PaymentMethod{output}, nil).Times(1)

				results, err := paymentMethodDomain.PaymentMethod().FindByFilter(context.Background(), filter)
				So(err, ShouldBeNil)
				So(results, ShouldNotBeEmpty)
			})
		})

		Convey("FindByID", func() {
			Convey("Success", func() {
				mockPaymentMethodDatabasePort.EXPECT().FindByID(gomock.Any()).Return(output, nil).Times(1)

				result, err := paymentMethodDomain.PaymentMethod().FindByID(context.Background(), "pm-123")
				So(err, ShouldBeNil)
				So(result.Name, ShouldEqual, "Bank Transfer")
			})
		})

		Convey("Update", func() {
			Convey("Success", func() {
				mockPaymentMethodDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				err := paymentMethodDomain.PaymentMethod().Update(context.Background(), "pm-123", input)
				So(err, ShouldBeNil)
			})
		})

		Convey("Delete", func() {
			Convey("Success", func() {
				mockPaymentMethodDatabasePort.EXPECT().Delete(gomock.Any()).Return(nil).Times(1)

				err := paymentMethodDomain.PaymentMethod().Delete(context.Background(), "pm-123")
				So(err, ShouldBeNil)
			})
		})
	})
}
