package subscription_test

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
	"mikrops/utils/crypto"
	mock_outbound_port "mikrops/tests/mocks/port"
)

func TestSubscription(t *testing.T) {
	os.Setenv("ENCRYPTION_KEY", "01234567890123456789012345678901")

	Convey("Test Subscription", t, func() {
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
		mockPppoeAccountDatabasePort := mock_outbound_port.NewMockPppoeAccountDatabasePort(mockCtrl)

		mockDatabasePort.EXPECT().Subscription().Return(mockSubscriptionDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().InternetPackage().Return(mockInternetPackageDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().Nas().Return(mockNasDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().PppoeAccount().Return(mockPppoeAccountDatabasePort).AnyTimes()
		mockHttpPort.EXPECT().Mikrotik().Return(mockMikrotikPort).AnyTimes()

		subscriptionDomain := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)

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

		pppoeAccountID := "pppoe-account-123"

		pppoeAccount := model.PppoeAccount{
			ID: pppoeAccountID,
			PppoeAccountInput: model.PppoeAccountInput{
				Username:          "pppoe_abc123",
				PasswordEncrypted: "encrypted_password",
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

		input := model.SubscriptionInput{
			TenantID:       "tenant-123",
			CustomerID:     "customer-123",
			PackageID:      "package-123",
			NasID:          "nas-123",
			PppoeAccountID: &pppoeAccountID,
			StartDate:      time.Now(),
			EndDate:        time.Now().AddDate(0, 1, 0),
			Status:         model.SubscriptionStatusActive,
		}

		output := model.Subscription{
			ID: "sub-123",
			SubscriptionInput: model.SubscriptionInput{
				TenantID:           "tenant-123",
				CustomerID:         "customer-123",
				PackageID:          "package-123",
				NasID:              "nas-123",
				StartDate:          time.Now(),
				EndDate:            time.Now().AddDate(0, 1, 0),
				Status:             model.SubscriptionStatusActive,
				MikrotikSecretName: "pppoe_abc123",
				MikrotikQueueName:  "queue_pppoe_abc123",
			},
		}

		Convey("Create", func() {
			Convey("Success with MikroTik sync", func() {
				mockPppoeAccountDatabasePort.EXPECT().FindByID(pppoeAccountID).Return(pppoeAccount, nil).Times(1)
				mockInternetPackageDatabasePort.EXPECT().FindByID(gomock.Any()).Return(pkg, nil).Times(1)
				mockSubscriptionDatabasePort.EXPECT().Create(gomock.Any()).DoAndReturn(
					func(data model.SubscriptionInput) (model.Subscription, error) {
						So(data.MikrotikSecretName, ShouldNotBeEmpty)
						So(data.MikrotikQueueName, ShouldNotBeEmpty)
						return output, nil
					},
				).Times(1)
				mockNasDatabasePort.EXPECT().FindByID("nas-123").Return(nas, nil).Times(1)
				mockMikrotikPort.EXPECT().CreatePPPoESecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockMikrotikPort.EXPECT().CreateSimpleQueue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				result, err := subscriptionDomain.Subscription().Create(context.Background(), input)
				So(err, ShouldBeNil)
				So(result.MikrotikSecretName, ShouldEqual, "pppoe_abc123")
			})

			Convey("PPPoE account not found", func() {
				mockInternetPackageDatabasePort.EXPECT().FindByID(gomock.Any()).Return(pkg, nil).Times(1)
				mockPppoeAccountDatabasePort.EXPECT().FindByID(pppoeAccountID).Return(model.PppoeAccount{}, errors.New("not found")).Times(1)

				_, err := subscriptionDomain.Subscription().Create(context.Background(), input)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("FindByFilter", func() {
			filter := model.SubscriptionFilter{IDs: []string{"sub-123"}}

			Convey("Filter is empty", func() {
				_, err := subscriptionDomain.Subscription().FindByFilter(context.Background(), model.SubscriptionFilter{})
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Subscription{output}, nil).Times(1)

				results, err := subscriptionDomain.Subscription().FindByFilter(context.Background(), filter)
				So(err, ShouldBeNil)
				So(results, ShouldNotBeEmpty)
			})
		})

		Convey("FindByID", func() {
			Convey("ID is empty", func() {
				_, err := subscriptionDomain.Subscription().FindByID(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByID(gomock.Any()).Return(output, nil).Times(1)

				_, err := subscriptionDomain.Subscription().FindByID(context.Background(), "sub-123")
				So(err, ShouldBeNil)
			})
		})

		Convey("Activate", func() {
			Convey("ID is empty", func() {
				err := subscriptionDomain.Subscription().Activate(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByID(gomock.Any()).Return(output, nil).Times(1)
				mockSubscriptionDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockNasDatabasePort.EXPECT().FindByID("nas-123").Return(nas, nil).Times(1)
				mockMikrotikPort.EXPECT().ListPPPoESecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.MikrotikPPPoESecret{
					{ID: "s1", Name: "pppoe_abc123"},
				}, nil).Times(1)
				mockMikrotikPort.EXPECT().EnablePPPoESecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "s1").Return(nil).Times(1)

				err := subscriptionDomain.Subscription().Activate(context.Background(), "sub-123")
				So(err, ShouldBeNil)
			})
		})

		Convey("Suspend", func() {
			Convey("Success", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByID(gomock.Any()).Return(output, nil).Times(1)
				mockSubscriptionDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockNasDatabasePort.EXPECT().FindByID("nas-123").Return(nas, nil).Times(1)
				mockMikrotikPort.EXPECT().ListPPPoESecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.MikrotikPPPoESecret{
					{ID: "s1", Name: "pppoe_abc123"},
				}, nil).Times(1)
				mockMikrotikPort.EXPECT().DisablePPPoESecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "s1").Return(nil).Times(1)

				err := subscriptionDomain.Subscription().Suspend(context.Background(), "sub-123")
				So(err, ShouldBeNil)
			})
		})

		Convey("SetVacation", func() {
			start := time.Now()
			end := time.Now().AddDate(0, 0, 7)

			Convey("Success", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByID(gomock.Any()).Return(output, nil).Times(1)
				mockSubscriptionDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockNasDatabasePort.EXPECT().FindByID("nas-123").Return(nas, nil).Times(1)
				mockMikrotikPort.EXPECT().ListPPPoESecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.MikrotikPPPoESecret{
					{ID: "s1", Name: "pppoe_abc123"},
				}, nil).Times(1)
				mockMikrotikPort.EXPECT().DisablePPPoESecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "s1").Return(nil).Times(1)

				err := subscriptionDomain.Subscription().SetVacation(context.Background(), "sub-123", start, end)
				So(err, ShouldBeNil)
			})
		})

		Convey("Cancel", func() {
			Convey("Success", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByID(gomock.Any()).Return(output, nil).Times(1)
				mockSubscriptionDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockNasDatabasePort.EXPECT().FindByID("nas-123").Return(nas, nil).Times(1)
				mockMikrotikPort.EXPECT().ListPPPoESecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.MikrotikPPPoESecret{
					{ID: "s1", Name: "pppoe_abc123"},
				}, nil).Times(1)
				mockMikrotikPort.EXPECT().DeletePPPoESecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "s1").Return(nil).Times(1)
				mockMikrotikPort.EXPECT().ListSimpleQueues(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.MikrotikSimpleQueue{
					{ID: "q1", Name: "queue_pppoe_abc123"},
				}, nil).Times(1)
				mockMikrotikPort.EXPECT().DeleteSimpleQueue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "q1").Return(nil).Times(1)

				err := subscriptionDomain.Subscription().Cancel(context.Background(), "sub-123")
				So(err, ShouldBeNil)
			})
		})

		Convey("FindExpiring", func() {
			Convey("Days must be positive", func() {
				_, err := subscriptionDomain.Subscription().FindExpiring(context.Background(), 0, "tenant-123")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockSubscriptionDatabasePort.EXPECT().FindExpiring(gomock.Any(), gomock.Any()).Return([]model.Subscription{output}, nil).Times(1)

				results, err := subscriptionDomain.Subscription().FindExpiring(context.Background(), 7, "tenant-123")
				So(err, ShouldBeNil)
				So(results, ShouldNotBeEmpty)
			})
		})
	})
}
