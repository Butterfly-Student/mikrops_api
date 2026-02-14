package gin_inbound_adapter_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"

	gin_inbound_adapter "go-template/internal/adapter/inbound/gin"
	"go-template/internal/domain"
	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestPaymentHandler(t *testing.T) {
	Convey("Test Payment HTTP Handler", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockPaymentDB := mock_outbound_port.NewMockPaymentDatabasePort(mockCtrl)
		mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(mockCtrl)

		mockDatabasePort.EXPECT().Payment().Return(mockPaymentDB).AnyTimes()
		mockDatabasePort.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, nil, nil, nil, nil, nil, nil, nil)
		handler := gin_inbound_adapter.NewPaymentAdapter(dom)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.POST("/payments", handler.CreatePayment)
		router.GET("/payments/:id", handler.GetPayment)
		router.GET("/payments", handler.ListPayments)
		router.PUT("/payments/:id", handler.UpdatePayment)
		router.DELETE("/payments/:id", handler.DeletePayment)

		Convey("CreatePayment", func() {
			Convey("Success", func() {
				customerID := uuid.New()
				amount := 100000.0

				reqBody := map[string]interface{}{
					"customer_id":    customerID.String(),
					"amount":         amount,
					"payment_method": "bank_transfer",
					"payment_date":   time.Now().Format(time.RFC3339),
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/payments", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				customer := &model.Customer{
					ID:       customerID,
					FullName: "John Doe",
				}

				mockCustomerDB.EXPECT().
					FindByID(customerID.String()).
					Return(customer, nil).
					Times(1)

				mockPaymentDB.EXPECT().
					Create(gomock.Any()).
					DoAndReturn(func(payment *model.Payment) error {
						payment.ID = uuid.New()
						return nil
					}).Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest("POST", "/payments", bytes.NewBuffer([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("GetPayment", func() {
			Convey("Success", func() {
				paymentID := uuid.New()

				expectedPayment := &model.Payment{
					ID:         paymentID,
					CustomerID: uuid.New(),
					Amount:     100000.0,
					Status:     "pending",
				}

				mockPaymentDB.EXPECT().
					FindByID(paymentID.String()).
					Return(expectedPayment, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/payments/"+paymentID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response model.Payment
				json.Unmarshal(w.Body.Bytes(), &response)
				So(response.ID, ShouldEqual, paymentID)
			})

			Convey("Not Found", func() {
				paymentID := uuid.New()

				mockPaymentDB.EXPECT().
					FindByID(paymentID.String()).
					Return(nil, errors.New("payment not found")).
					Times(1)

				req := httptest.NewRequest("GET", "/payments/"+paymentID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("ListPayments", func() {
			Convey("Success", func() {
				payments := []model.Payment{
					{
						ID:         uuid.New(),
						CustomerID: uuid.New(),
						Amount:     100000.0,
						Status:     "pending",
					},
					{
						ID:         uuid.New(),
						CustomerID: uuid.New(),
						Amount:     200000.0,
						Status:     "confirmed",
					},
				}

				mockPaymentDB.EXPECT().
					Find(gomock.Any()).
					Return(payments, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/payments", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response []model.Payment
				json.Unmarshal(w.Body.Bytes(), &response)
				So(len(response), ShouldEqual, 2)
			})
		})

		Convey("UpdatePayment", func() {
			Convey("Success", func() {
				paymentID := uuid.New()
				customerID := uuid.New()
				newAmount := 150000.0

				reqBody := map[string]interface{}{
					"customer_id":    customerID.String(),
					"amount":         newAmount,
					"payment_method": "bank_transfer",
					"status":         "confirmed",
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("PUT", "/payments/"+paymentID.String(), bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				existingPayment := &model.Payment{
					ID:         paymentID,
					CustomerID: customerID,
					Amount:     100000.0,
					Status:     "pending",
				}

				mockPaymentDB.EXPECT().
					FindByID(paymentID.String()).
					Return(existingPayment, nil).
					Times(1)

				mockPaymentDB.EXPECT().
					Update(gomock.Any()).
					Return(nil).
					Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("DeletePayment", func() {
			Convey("Success", func() {
				paymentID := uuid.New()

				mockPaymentDB.EXPECT().
					Delete(paymentID.String()).
					Return(nil).
					Times(1)

				req := httptest.NewRequest("DELETE", "/payments/"+paymentID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}
