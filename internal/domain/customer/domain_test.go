package customer_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	mock_outbound_port "mikrops/tests/mocks/port"
)

func TestCustomer(t *testing.T) {
	Convey("Test Customer", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockCustomerDatabasePort := mock_outbound_port.NewMockCustomerDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().Customer().Return(mockCustomerDatabasePort).AnyTimes()

		customerDomain := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)

		input := model.CustomerInput{
			TenantID: "tenant-123",
			FullName: "Test Customer",
			Email:    "customer@example.com",
			Phone:    "081234567890",
			Address:  "Test Address",
			Password: "password123",
		}

		output := model.Customer{
			ID: "customer-123",
			CustomerInput: model.CustomerInput{
				TenantID:     "tenant-123",
				FullName:     "Test Customer",
				Email:        "customer@example.com",
				Phone:        "081234567890",
				Address:      "Test Address",
				PasswordHash: "$2a$10$hashedpassword",
				IsActive:     true,
				CreatedAt:    time.Now(),
			},
		}

		Convey("Create", func() {
			Convey("Success", func() {
				mockCustomerDatabasePort.EXPECT().Create(gomock.Any()).DoAndReturn(
					func(data model.CustomerInput) (model.Customer, error) {
						// Verify password was hashed
						So(data.PasswordHash, ShouldNotBeEmpty)
						So(data.Password, ShouldBeEmpty)
						return output, nil
					},
				).Times(1)

				result, err := customerDomain.Customer().Create(context.Background(), input)
				So(err, ShouldBeNil)
				So(result.FullName, ShouldEqual, "Test Customer")
			})

			Convey("Database error", func() {
				mockCustomerDatabasePort.EXPECT().Create(gomock.Any()).Return(model.Customer{}, errors.New("error")).Times(1)

				_, err := customerDomain.Customer().Create(context.Background(), input)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("FindByFilter", func() {
			filter := model.CustomerFilter{IDs: []string{"customer-123"}}

			Convey("Filter is empty", func() {
				_, err := customerDomain.Customer().FindByFilter(context.Background(), model.CustomerFilter{})
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockCustomerDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Customer{output}, nil).Times(1)

				results, err := customerDomain.Customer().FindByFilter(context.Background(), filter)
				So(err, ShouldBeNil)
				So(results, ShouldNotBeEmpty)
			})
		})

		Convey("FindByID", func() {
			Convey("ID is empty", func() {
				_, err := customerDomain.Customer().FindByID(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockCustomerDatabasePort.EXPECT().FindByID(gomock.Any()).Return(output, nil).Times(1)

				result, err := customerDomain.Customer().FindByID(context.Background(), "customer-123")
				So(err, ShouldBeNil)
				So(result.FullName, ShouldEqual, "Test Customer")
			})
		})

		Convey("FindByUsername", func() {
			Convey("Username is empty", func() {
				_, err := customerDomain.Customer().FindByUsername(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockCustomerDatabasePort.EXPECT().FindByUsername(gomock.Any()).Return(output, nil).Times(1)

				result, err := customerDomain.Customer().FindByUsername(context.Background(), "customer123")
				So(err, ShouldBeNil)
				So(result.FullName, ShouldEqual, "Test Customer")
			})
		})

		Convey("Update", func() {
			Convey("ID is empty", func() {
				err := customerDomain.Customer().Update(context.Background(), "", input)
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockCustomerDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				err := customerDomain.Customer().Update(context.Background(), "customer-123", input)
				So(err, ShouldBeNil)
			})
		})

		Convey("Delete", func() {
			Convey("Success", func() {
				mockCustomerDatabasePort.EXPECT().Delete(gomock.Any()).Return(nil).Times(1)

				err := customerDomain.Customer().Delete(context.Background(), "customer-123")
				So(err, ShouldBeNil)
			})
		})
	})
}
