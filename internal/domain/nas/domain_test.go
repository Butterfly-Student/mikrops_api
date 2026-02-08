package nas_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	mock_outbound_port "mikrops/tests/mocks/port"
	"mikrops/utils/crypto"
)

func TestNas(t *testing.T) {
	Convey("Test NAS", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")
		defer os.Unsetenv("ENCRYPTION_KEY")

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(mockCtrl)
		mockNasDatabasePort := mock_outbound_port.NewMockNasDatabasePort(mockCtrl)
		mockTenantDatabasePort := mock_outbound_port.NewMockTenantDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().Nas().Return(mockNasDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().Tenant().Return(mockTenantDatabasePort).AnyTimes()
		mockHttpPort.EXPECT().Mikrotik().Return(mockMikrotikPort).AnyTimes()

		nasDomain := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)

		encryptPassword, _ := crypto.Encrypt("password123")

		input := model.NasInput{
			TenantID: "tenant-123",
			Name:     "Test NAS",
			Host:     "192.168.1.1",
			Username: "admin",
			Password: "password123",
			ApiPort:  8728,
			RestPort: 80,
			IsActive: true,
		}

		output := model.Nas{
			ID: "nas-123",
			NasInput: model.NasInput{
				TenantID:          "tenant-123",
				Name:              "Test NAS",
				Host:              "192.168.1.1",
				Username:          "admin",
				PasswordEncrypted: encryptPassword,
				ApiPort:           8728,
				RestPort:          80,
				IsActive:          true,
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			},
		}

		tenant := model.Tenant{
			ID: "tenant-123",
			TenantInput: model.TenantInput{
				MaxNas: 3,
			},
		}

		Convey("Create", func() {
			Convey("Max NAS limit reached", func() {
				mockNasDatabasePort.EXPECT().CountByTenantID(gomock.Any()).Return(3, nil).Times(1)
				mockTenantDatabasePort.EXPECT().FindByID(gomock.Any()).Return(tenant, nil).Times(1)

				_, err := nasDomain.Nas().Create(context.Background(), input)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "maximum number of NAS")
			})

			Convey("Success with password encryption", func() {
				mockNasDatabasePort.EXPECT().CountByTenantID(gomock.Any()).Return(1, nil).Times(1)
				mockTenantDatabasePort.EXPECT().FindByID(gomock.Any()).Return(tenant, nil).Times(1)
				mockNasDatabasePort.EXPECT().Create(gomock.Any()).DoAndReturn(
					func(data model.NasInput) (model.Nas, error) {
						// Verify password was encrypted
						So(data.PasswordEncrypted, ShouldNotBeEmpty)
						So(data.Password, ShouldBeEmpty)
						return output, nil
					},
				).Times(1)

				result, err := nasDomain.Nas().Create(context.Background(), input)
				So(err, ShouldBeNil)
				So(result.Name, ShouldEqual, "Test NAS")
			})

			Convey("Database error on count", func() {
				mockNasDatabasePort.EXPECT().CountByTenantID(gomock.Any()).Return(0, errors.New("error")).Times(1)

				_, err := nasDomain.Nas().Create(context.Background(), input)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("FindByFilter", func() {
			filter := model.NasFilter{IDs: []string{"nas-123"}}

			Convey("Filter is empty", func() {
				_, err := nasDomain.Nas().FindByFilter(context.Background(), model.NasFilter{})
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Nas{output}, nil).Times(1)

				results, err := nasDomain.Nas().FindByFilter(context.Background(), filter)
				So(err, ShouldBeNil)
				So(results, ShouldNotBeEmpty)
			})
		})

		Convey("FindByID", func() {
			Convey("ID is empty", func() {
				_, err := nasDomain.Nas().FindByID(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().FindByID(gomock.Any()).Return(output, nil).Times(1)

				result, err := nasDomain.Nas().FindByID(context.Background(), "nas-123")
				So(err, ShouldBeNil)
				So(result.Name, ShouldEqual, "Test NAS")
			})
		})

		Convey("Update", func() {
			Convey("ID is empty", func() {
				err := nasDomain.Nas().Update(context.Background(), "", input)
				So(err, ShouldNotBeNil)
			})

			Convey("Success with password re-encryption", func() {
				mockNasDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
					func(id string, data model.NasInput) error {
						if data.Password != "" {
							So(data.PasswordEncrypted, ShouldNotBeEmpty)
							So(data.Password, ShouldBeEmpty)
						}
						return nil
					},
				).Times(1)

				err := nasDomain.Nas().Update(context.Background(), "nas-123", input)
				So(err, ShouldBeNil)
			})
		})

		Convey("Delete", func() {
			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().Delete(gomock.Any()).Return(nil).Times(1)

				err := nasDomain.Nas().Delete(context.Background(), "nas-123")
				So(err, ShouldBeNil)
			})
		})

		Convey("CountByTenantID", func() {
			Convey("TenantID is empty", func() {
				_, err := nasDomain.Nas().CountByTenantID(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().CountByTenantID(gomock.Any()).Return(2, nil).Times(1)

				count, err := nasDomain.Nas().CountByTenantID(context.Background(), "tenant-123")
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 2)
			})
		})

		Convey("TestConnection", func() {
			Convey("ID is empty", func() {
				err := nasDomain.Nas().TestConnection(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("NAS not found", func() {
				mockNasDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Nas{}, errors.New("not found")).Times(1)

				err := nasDomain.Nas().TestConnection(context.Background(), "nas-123")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().FindByID(gomock.Any()).Return(output, nil).Times(1)
				mockMikrotikPort.EXPECT().TestConnection(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				err := nasDomain.Nas().TestConnection(context.Background(), "nas-123")
				So(err, ShouldBeNil)
			})
		})
	})
}
