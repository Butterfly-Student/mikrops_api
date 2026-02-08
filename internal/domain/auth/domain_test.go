package auth_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/bcrypt"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	mock_outbound_port "mikrops/tests/mocks/port"
)

func TestAuth(t *testing.T) {
	Convey("Test Auth", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockStaffDatabasePort := mock_outbound_port.NewMockStaffDatabasePort(mockCtrl)
		mockCustomerDatabasePort := mock_outbound_port.NewMockCustomerDatabasePort(mockCtrl)
		mockTenantDatabasePort := mock_outbound_port.NewMockTenantDatabasePort(mockCtrl)

		mockDatabasePort.EXPECT().Staff().Return(mockStaffDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().Customer().Return(mockCustomerDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().Tenant().Return(mockTenantDatabasePort).AnyTimes()

		authDomain := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)

		// Set JWT secret for testing
		os.Setenv("JWT_SECRET", "test-secret-key")
		defer os.Unsetenv("JWT_SECRET")

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

		staff := model.Staff{
			ID: "staff-123",
			StaffInput: model.StaffInput{
				TenantID:     "tenant-123",
				Email:        "staff@example.com",
				PasswordHash: string(hashedPassword),
			},
		}

		customer := model.Customer{
			ID: "customer-123",
			CustomerInput: model.CustomerInput{
				TenantID:     "tenant-123",
				Username:     "customer123",
				PasswordHash: string(hashedPassword),
			},
		}

		tenant := model.Tenant{
			ID: "tenant-123",
			TenantInput: model.TenantInput{
				Slug: "test-tenant",
			},
		}

		Convey("StaffLogin", func() {
			req := model.StaffAuthRequest{
				Email:    "staff@example.com",
				Password: "password123",
			}

			Convey("Email or password is empty", func() {
				emptyReq := model.StaffAuthRequest{}
				_, err := authDomain.Auth().StaffLogin(context.Background(), emptyReq)
				So(err, ShouldNotBeNil)
			})

			Convey("Staff not found", func() {
				mockStaffDatabasePort.EXPECT().FindByEmail(gomock.Any()).Return(model.Staff{}, errors.New("not found")).Times(1)

				_, err := authDomain.Auth().StaffLogin(context.Background(), req)
				So(err, ShouldNotBeNil)
			})

			Convey("Invalid password", func() {
				mockStaffDatabasePort.EXPECT().FindByEmail(gomock.Any()).Return(staff, nil).Times(1)

				invalidReq := req
				invalidReq.Password = "wrongpassword"
				_, err := authDomain.Auth().StaffLogin(context.Background(), invalidReq)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid credentials")
			})

			Convey("Success", func() {
				mockStaffDatabasePort.EXPECT().FindByEmail(gomock.Any()).Return(staff, nil).Times(1)
				mockStaffDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				result, err := authDomain.Auth().StaffLogin(context.Background(), req)
				So(err, ShouldBeNil)
				So(result.AccessToken, ShouldNotBeEmpty)
				So(result.RefreshToken, ShouldNotBeEmpty)
				So(result.TokenType, ShouldEqual, "Bearer")
			})
		})

		Convey("CustomerLogin", func() {
			req := model.CustomerAuthRequest{
				Username:   "customer123",
				Password:   "password123",
				TenantSlug: "test-tenant",
			}

			Convey("Required fields are empty", func() {
				emptyReq := model.CustomerAuthRequest{}
				_, err := authDomain.Auth().CustomerLogin(context.Background(), emptyReq)
				So(err, ShouldNotBeNil)
			})

			Convey("Tenant not found", func() {
				mockTenantDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Tenant{}, nil).Times(1)

				_, err := authDomain.Auth().CustomerLogin(context.Background(), req)
				So(err, ShouldNotBeNil)
			})

			Convey("Customer not found", func() {
				mockTenantDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Tenant{tenant}, nil).Times(1)
				mockCustomerDatabasePort.EXPECT().FindByUsername(gomock.Any()).Return(model.Customer{}, errors.New("not found")).Times(1)

				_, err := authDomain.Auth().CustomerLogin(context.Background(), req)
				So(err, ShouldNotBeNil)
			})

			Convey("Invalid password", func() {
				mockTenantDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Tenant{tenant}, nil).Times(1)
				mockCustomerDatabasePort.EXPECT().FindByUsername(gomock.Any()).Return(customer, nil).Times(1)

				invalidReq := req
				invalidReq.Password = "wrongpassword"
				_, err := authDomain.Auth().CustomerLogin(context.Background(), invalidReq)
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Tenant{tenant}, nil).Times(1)
				mockCustomerDatabasePort.EXPECT().FindByUsername(gomock.Any()).Return(customer, nil).Times(1)

				result, err := authDomain.Auth().CustomerLogin(context.Background(), req)
				So(err, ShouldBeNil)
				So(result.AccessToken, ShouldNotBeEmpty)
				So(result.RefreshToken, ShouldNotBeEmpty)
			})
		})

		Convey("StaffRefreshToken", func() {
			Convey("Refresh token is empty", func() {
				_, err := authDomain.Auth().StaffRefreshToken(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			// Note: Full refresh token testing requires actual JWT parsing
			// which needs more complex setup. Basic validation is covered above.
		})

		Convey("CustomerRefreshToken", func() {
			Convey("Refresh token is empty", func() {
				_, err := authDomain.Auth().CustomerRefreshToken(context.Background(), "")
				So(err, ShouldNotBeNil)
			})
		})
	})
}
