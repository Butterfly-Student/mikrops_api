package payment_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
	mock_outbound_port "mikrops/tests/mocks/port"
)

func TestPayment(t *testing.T) {
	Convey("Test Payment", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockPaymentDatabasePort := mock_outbound_port.NewMockPaymentDatabasePort(mockCtrl)
		mockInvoiceDatabasePort := mock_outbound_port.NewMockInvoiceDatabasePort(mockCtrl)

		mockDatabasePort.EXPECT().Payment().Return(mockPaymentDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().Invoice().Return(mockInvoiceDatabasePort).AnyTimes()

		paymentDomain := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)

		input := model.PaymentInput{
			TenantID:        "tenant-123",
			InvoiceID:       "invoice-123",
			PaymentMethodID: "pm-123",
			Amount:          110000,
			PaymentDate:     time.Now(),
			Status:          model.PaymentStatusPending,
		}

		payment := model.Payment{
			ID:           "payment-123",
			PaymentInput: input,
		}

		invoice := model.Invoice{
			ID: "invoice-123",
			InvoiceInput: model.InvoiceInput{
				Status: model.InvoiceStatusUnpaid,
			},
		}

		Convey("Create", func() {
			Convey("Success", func() {
				mockPaymentDatabasePort.EXPECT().Create(gomock.Any()).Return(payment, nil).Times(1)

				result, err := paymentDomain.Payment().Create(context.Background(), input)
				So(err, ShouldBeNil)
				So(result.Amount, ShouldEqual, 110000)
			})
		})

		Convey("FindByFilter", func() {
			filter := model.PaymentFilter{IDs: []string{"payment-123"}}

			Convey("Filter is empty", func() {
				_, err := paymentDomain.Payment().FindByFilter(context.Background(), model.PaymentFilter{})
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockPaymentDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Payment{payment}, nil).Times(1)

				results, err := paymentDomain.Payment().FindByFilter(context.Background(), filter)
				So(err, ShouldBeNil)
				So(results, ShouldNotBeEmpty)
			})
		})

		Convey("FindByID", func() {
			Convey("Success", func() {
				mockPaymentDatabasePort.EXPECT().FindByID(gomock.Any()).Return(payment, nil).Times(1)

				_, err := paymentDomain.Payment().FindByID(context.Background(), "payment-123")
				So(err, ShouldBeNil)
			})
		})

		Convey("Update", func() {
			Convey("ID is empty", func() {
				err := paymentDomain.Payment().Update(context.Background(), "", input)
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockPaymentDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				err := paymentDomain.Payment().Update(context.Background(), "payment-123", input)
				So(err, ShouldBeNil)
			})
		})

		Convey("Delete", func() {
			Convey("ID is empty", func() {
				err := paymentDomain.Payment().Delete(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockPaymentDatabasePort.EXPECT().Delete(gomock.Any()).Return(nil).Times(1)

				err := paymentDomain.Payment().Delete(context.Background(), "payment-123")
				So(err, ShouldBeNil)
			})
		})

		Convey("Verify", func() {
			staffID := "staff-123"

			Convey("ID is empty", func() {
				err := paymentDomain.Payment().Verify(context.Background(), "", "staff-123")
				So(err, ShouldNotBeNil)
			})

			Convey("StaffID is empty", func() {
				err := paymentDomain.Payment().Verify(context.Background(), "payment-123", "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success with transaction", func() {
				mockDatabasePort.EXPECT().DoInTransaction(gomock.Any()).DoAndReturn(
					func(fn func(outbound_port.DatabasePort) (interface{}, error)) (interface{}, error) {
						// Mock transaction ports
						mockTxPaymentPort := mock_outbound_port.NewMockPaymentDatabasePort(mockCtrl)
						mockTxInvoicePort := mock_outbound_port.NewMockInvoiceDatabasePort(mockCtrl)
						mockTxDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)

						mockTxDatabasePort.EXPECT().Payment().Return(mockTxPaymentPort).AnyTimes()
						mockTxDatabasePort.EXPECT().Invoice().Return(mockTxInvoicePort).AnyTimes()

						mockTxPaymentPort.EXPECT().FindByID(gomock.Any()).Return(payment, nil).Times(1)
						mockTxInvoicePort.EXPECT().FindByID(gomock.Any()).Return(invoice, nil).Times(1)

						mockTxPaymentPort.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
							func(id string, data model.PaymentInput) error {
								So(data.Status, ShouldEqual, model.PaymentStatusVerified)
								So(data.VerifiedBy, ShouldNotBeNil)
								So(*data.VerifiedBy, ShouldEqual, staffID)
								return nil
							},
						).Times(1)

						mockTxInvoicePort.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
							func(id string, data model.InvoiceInput) error {
								So(data.Status, ShouldEqual, model.InvoiceStatusPaid)
								So(data.PaidAt, ShouldNotBeNil)
								return nil
							},
						).Times(1)

						return fn(mockTxDatabasePort)
					},
				).Times(1)

				err := paymentDomain.Payment().Verify(context.Background(), "payment-123", staffID)
				So(err, ShouldBeNil)
			})

			Convey("Payment not found", func() {
				mockDatabasePort.EXPECT().DoInTransaction(gomock.Any()).DoAndReturn(
					func(fn func(outbound_port.DatabasePort) (interface{}, error)) (interface{}, error) {
						mockTxPaymentPort := mock_outbound_port.NewMockPaymentDatabasePort(mockCtrl)
						mockTxInvoicePort := mock_outbound_port.NewMockInvoiceDatabasePort(mockCtrl)
						mockTxDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)

						mockTxDatabasePort.EXPECT().Payment().Return(mockTxPaymentPort).AnyTimes()
						mockTxDatabasePort.EXPECT().Invoice().Return(mockTxInvoicePort).AnyTimes()

						mockTxPaymentPort.EXPECT().FindByID(gomock.Any()).Return(model.Payment{}, errors.New("not found")).Times(1)

						return fn(mockTxDatabasePort)
					},
				).Times(1)

				err := paymentDomain.Payment().Verify(context.Background(), "payment-123", staffID)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Reject", func() {
			notes := "Invalid payment proof"
			staffID := "staff-123"

			Convey("ID is empty", func() {
				err := paymentDomain.Payment().Reject(context.Background(), "", staffID, notes)
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockPaymentDatabasePort.EXPECT().FindByID(gomock.Any()).Return(payment, nil).Times(1)
				mockPaymentDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
					func(id string, data model.PaymentInput) error {
						So(data.Status, ShouldEqual, model.PaymentStatusRejected)
						So(data.VerifiedBy, ShouldNotBeNil)
						So(data.VerifiedAt, ShouldNotBeNil)
						So(data.Notes, ShouldEqual, notes)
						return nil
					},
				).Times(1)

				err := paymentDomain.Payment().Reject(context.Background(), "payment-123", staffID, notes)
				So(err, ShouldBeNil)
			})
		})
	})
}
