package mikrotik_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	"mikrops/utils/crypto"
	mock_outbound_port "mikrops/tests/mocks/port"
)

func TestMikrotik(t *testing.T) {
	// Set encryption key for tests
	os.Setenv("ENCRYPTION_KEY", "01234567890123456789012345678901")

	Convey("Test Mikrotik", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)
		mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(mockCtrl)

		mockSubscriptionDatabasePort := mock_outbound_port.NewMockSubscriptionDatabasePort(mockCtrl)
		mockInternetPackageDatabasePort := mock_outbound_port.NewMockInternetPackageDatabasePort(mockCtrl)
		mockNasDatabasePort := mock_outbound_port.NewMockNasDatabasePort(mockCtrl)

		mockDatabasePort.EXPECT().Subscription().Return(mockSubscriptionDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().InternetPackage().Return(mockInternetPackageDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().Nas().Return(mockNasDatabasePort).AnyTimes()
		mockHttpPort.EXPECT().Mikrotik().Return(mockMikrotikPort).AnyTimes()

		mikrotikDomain := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)

		encryptedPassword, _ := crypto.Encrypt("naspassword")

		nas := model.Nas{
			ID: "nas-123",
			NasInput: model.NasInput{
				Host:              "192.168.1.1",
				ApiPort:           8728,
				RestPort:          80,
				Username:          "admin",
				PasswordEncrypted: encryptedPassword,
				UseSSL:            false,
			},
		}

		subscription := model.Subscription{
			ID: "sub-123",
			SubscriptionInput: model.SubscriptionInput{
				CustomerID:         "customer-123",
				PackageID:          "package-123",
				NasID:              "nas-123",
				MikrotikSecretName: "pppoe_abc123",
				MikrotikQueueName:  "queue_pppoe_abc123",
			},
		}

		pkg := model.InternetPackage{
			ID: "package-123",
			InternetPackageInput: model.InternetPackageInput{
				Name:         "Basic",
				UploadRate:   "10M",
				DownloadRate: "10M",
			},
		}

		Convey("SyncSubscription", func() {
			Convey("SubscriptionID is empty", func() {
				err := mikrotikDomain.Mikrotik().SyncSubscription(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Subscription not found", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Subscription{}, errors.New("not found")).Times(1)

				err := mikrotikDomain.Mikrotik().SyncSubscription(context.Background(), "sub-123")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByID(gomock.Any()).Return(subscription, nil).Times(1)
				mockInternetPackageDatabasePort.EXPECT().FindByID(gomock.Any()).Return(pkg, nil).Times(1)
				mockNasDatabasePort.EXPECT().FindByID("nas-123").Return(nas, nil).Times(1)
				mockMikrotikPort.EXPECT().CreatePPPoESecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockMikrotikPort.EXPECT().CreateSimpleQueue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				err := mikrotikDomain.Mikrotik().SyncSubscription(context.Background(), "sub-123")
				So(err, ShouldBeNil)
			})
		})

		Convey("GetActiveConnections", func() {
			Convey("NasID is empty", func() {
				_, err := mikrotikDomain.Mikrotik().GetActiveConnections(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().FindByID("nas-123").Return(nas, nil).Times(1)
				mockMikrotikPort.EXPECT().GetActiveConnections(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.MikrotikConnection{}, nil).Times(1)

				connections, err := mikrotikDomain.Mikrotik().GetActiveConnections(context.Background(), "nas-123")
				So(err, ShouldBeNil)
				So(connections, ShouldNotBeNil)
			})
		})

		Convey("GetBandwidth", func() {
			target := "pppoe_abc123"

			Convey("NasID is empty", func() {
				_, err := mikrotikDomain.Mikrotik().GetBandwidth(context.Background(), "", target)
				So(err, ShouldNotBeNil)
			})

			Convey("Target is empty", func() {
				_, err := mikrotikDomain.Mikrotik().GetBandwidth(context.Background(), "nas-123", "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().FindByID("nas-123").Return(nas, nil).Times(1)
				mockMikrotikPort.EXPECT().MonitorBandwidth(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.MikrotikBandwidth{}, nil).Times(1)

				bandwidth, err := mikrotikDomain.Mikrotik().GetBandwidth(context.Background(), "nas-123", target)
				So(err, ShouldBeNil)
				So(bandwidth, ShouldNotBeNil)
			})
		})

		Convey("RemoveSubscription", func() {
			Convey("SubscriptionID is empty", func() {
				err := mikrotikDomain.Mikrotik().RemoveSubscription(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Subscription not found", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Subscription{}, errors.New("not found")).Times(1)

				err := mikrotikDomain.Mikrotik().RemoveSubscription(context.Background(), "sub-123")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByID(gomock.Any()).Return(subscription, nil).Times(1)
				mockNasDatabasePort.EXPECT().FindByID("nas-123").Return(nas, nil).Times(1)
				mockMikrotikPort.EXPECT().ListPPPoESecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.MikrotikPPPoESecret{
					{ID: "secret-1", Name: "pppoe_abc123"},
				}, nil).Times(1)
				mockMikrotikPort.EXPECT().DeletePPPoESecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "secret-1").Return(nil).Times(1)
				mockMikrotikPort.EXPECT().ListSimpleQueues(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.MikrotikSimpleQueue{
					{ID: "queue-1", Name: "queue_pppoe_abc123"},
				}, nil).Times(1)
				mockMikrotikPort.EXPECT().DeleteSimpleQueue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "queue-1").Return(nil).Times(1)

				err := mikrotikDomain.Mikrotik().RemoveSubscription(context.Background(), "sub-123")
				So(err, ShouldBeNil)
			})
		})
	})
}
