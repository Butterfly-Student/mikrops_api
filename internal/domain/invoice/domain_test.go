package invoice_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	mock_outbound_port "mikrops/tests/mocks/port"
)

func TestInvoice(t *testing.T) {
	Convey("Test Invoice", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockInvoiceDatabasePort := mock_outbound_port.NewMockInvoiceDatabasePort(mockCtrl)
		mockSubscriptionDatabasePort := mock_outbound_port.NewMockSubscriptionDatabasePort(mockCtrl)
		mockInternetPackageDatabasePort := mock_outbound_port.NewMockInternetPackageDatabasePort(mockCtrl)

		mockDatabasePort.EXPECT().Invoice().Return(mockInvoiceDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().Subscription().Return(mockSubscriptionDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().InternetPackage().Return(mockInternetPackageDatabasePort).AnyTimes()

		invoiceDomain := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)

		input := model.InvoiceInput{
			TenantID:       "tenant-123",
			CustomerID:     "customer-123",
			SubscriptionID: "sub-123",
			Amount:         100000,
			TaxAmount:      10000,
			TotalAmount:    110000,
			DueDate:        time.Now().AddDate(0, 0, 7),
			PeriodStart:    time.Now(),
			PeriodEnd:      time.Now().AddDate(0, 1, 0),
			Status:         model.InvoiceStatusUnpaid,
		}

		output := model.Invoice{
			ID:           "invoice-123",
			InvoiceInput: input,
		}

		Convey("Create", func() {
			Convey("Success with invoice number generation", func() {
				mockInvoiceDatabasePort.EXPECT().Create(gomock.Any()).DoAndReturn(
					func(data model.InvoiceInput) (model.Invoice, error) {
						So(data.InvoiceNumber, ShouldNotBeEmpty)
						So(data.TotalAmount, ShouldEqual, data.Amount+data.TaxAmount)
						// Return the input data with the generated invoice number
						result := output
						result.InvoiceNumber = data.InvoiceNumber
						return result, nil
					},
				).Times(1)

				result, err := invoiceDomain.Invoice().Create(context.Background(), input)
				So(err, ShouldBeNil)
				So(result.InvoiceNumber, ShouldNotBeEmpty)
			})
		})

		Convey("FindByFilter", func() {
			filter := model.InvoiceFilter{IDs: []string{"invoice-123"}}

			Convey("Filter is empty", func() {
				_, err := invoiceDomain.Invoice().FindByFilter(context.Background(), model.InvoiceFilter{})
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockInvoiceDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Invoice{output}, nil).Times(1)

				results, err := invoiceDomain.Invoice().FindByFilter(context.Background(), filter)
				So(err, ShouldBeNil)
				So(results, ShouldNotBeEmpty)
			})
		})

		Convey("FindByID", func() {
			Convey("Success", func() {
				mockInvoiceDatabasePort.EXPECT().FindByID(gomock.Any()).Return(output, nil).Times(1)

				_, err := invoiceDomain.Invoice().FindByID(context.Background(), "invoice-123")
				So(err, ShouldBeNil)
			})
		})

		Convey("Update", func() {
			Convey("Success", func() {
				mockInvoiceDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				err := invoiceDomain.Invoice().Update(context.Background(), "invoice-123", input)
				So(err, ShouldBeNil)
			})
		})

		Convey("Delete", func() {
			Convey("ID is empty", func() {
				err := invoiceDomain.Invoice().Delete(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockInvoiceDatabasePort.EXPECT().Delete(gomock.Any()).Return(nil).Times(1)

				err := invoiceDomain.Invoice().Delete(context.Background(), "invoice-123")
				So(err, ShouldBeNil)
			})
		})

		Convey("GenerateBulk", func() {
			periodStart := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
			periodEnd := time.Date(2026, 2, 28, 23, 59, 59, 0, time.UTC)

			subscription := model.Subscription{
				ID: "sub-123",
				SubscriptionInput: model.SubscriptionInput{
					TenantID:   "tenant-123",
					CustomerID: "customer-123",
					PackageID:  "package-123",
				},
			}

			pkg := model.InternetPackage{
				ID: "package-123",
				InternetPackageInput: model.InternetPackageInput{
					Name:  "Package 10Mbps",
					Price: 100000,
				},
			}

			Convey("TenantID is empty", func() {
				err := invoiceDomain.Invoice().GenerateBulk(context.Background(), "", periodStart, periodEnd)
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Subscription{subscription}, nil).Times(1)
				mockInternetPackageDatabasePort.EXPECT().FindByID(gomock.Any()).Return(pkg, nil).Times(1)
				mockInvoiceDatabasePort.EXPECT().BulkCreate(gomock.Any()).DoAndReturn(
					func(data []model.InvoiceInput) error {
						So(data, ShouldNotBeEmpty)
						So(data[0].Amount, ShouldEqual, 100000)
						So(data[0].InvoiceNumber, ShouldNotBeEmpty)
						return nil
					},
				).Times(1)

				err := invoiceDomain.Invoice().GenerateBulk(context.Background(), "tenant-123", periodStart, periodEnd)
				So(err, ShouldBeNil)
			})

			Convey("No active subscriptions", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Subscription{}, nil).Times(1)

				err := invoiceDomain.Invoice().GenerateBulk(context.Background(), "tenant-123", periodStart, periodEnd)
				So(err, ShouldBeNil)
			})
		})
	})
}
