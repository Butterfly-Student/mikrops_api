package gin_inbound_adapter_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"

	gin_inbound_adapter "go-template/internal/adapter/inbound/gin"
	"go-template/internal/domain"
	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestBillingHandler(t *testing.T) {
	Convey("Test Billing HTTP Handler", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockInvoiceDB := mock_outbound_port.NewMockInvoiceDatabasePort(mockCtrl)
		mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(mockCtrl)

		mockBandwidthProfileDB := mock_outbound_port.NewMockBandwidthProfileDatabasePort(mockCtrl)

		mockDatabasePort.EXPECT().Invoice().Return(mockInvoiceDB).AnyTimes()
		mockDatabasePort.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()
		mockDatabasePort.EXPECT().SystemSetting().Return(mock_outbound_port.NewMockSystemSettingDatabasePort(mockCtrl)).AnyTimes()
		mockDatabasePort.EXPECT().BandwidthProfile().Return(mockBandwidthProfileDB).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, nil, nil, nil, nil, nil, nil, nil)
		handler := gin_inbound_adapter.NewBillingHandler(dom)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.POST("/invoices", handler.CreateInvoice)
		router.GET("/invoices/:id", handler.GetInvoice)
		router.GET("/invoices", handler.ListInvoices)

		Convey("CreateInvoice", func() {
			Convey("Success", func() {
				customerID := uuid.New()

				reqBody := map[string]interface{}{
					"customer_id": customerID.String(),
					"items": []map[string]interface{}{
						{
							"description": "Monthly subscription",
							"quantity":    1,
							"unit_price":  100000.0,
						},
					},
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/invoices", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				profileID := uuid.New()
				customer := &model.Customer{
					ID:        customerID,
					FullName:  "John Doe",
					ProfileID: &profileID,
				}

				profile := &model.BandwidthProfile{
					ID:            profileID,
					ProfileCode:   "BW-100MBPS",
					Name:          "100 Mbps",
					DownloadSpeed: 100000000,
					UploadSpeed:   100000000,
				}

				mockCustomerDB.EXPECT().
					FindByID(customerID.String()).
					Return(customer, nil).
					Times(1)

				mockBandwidthProfileDB.EXPECT().
					FindByID(profileID.String()).
					Return(profile, nil).
					Times(1)

				expectedInvoice := &model.Invoice{
					ID:          uuid.New(),
					CustomerID:  customerID,
					TotalAmount: 100000.0,
					Status:      "draft",
				}

				mockDatabasePort.EXPECT().
					DoInTransaction(gomock.Any()).
					DoAndReturn(func(txFunc interface{}) (interface{}, error) {
						return expectedInvoice, nil
					}).Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)
			})
		})

		Convey("GetInvoice", func() {
			Convey("Success", func() {
				invoiceID := uuid.New()

				expectedInvoice := &model.Invoice{
					ID:          invoiceID,
					CustomerID:  uuid.New(),
					TotalAmount: 100000.0,
					Status:      "draft",
				}

				mockInvoiceDB.EXPECT().
					FindByID(invoiceID.String()).
					Return(expectedInvoice, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/invoices/"+invoiceID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Not Found", func() {
				invoiceID := uuid.New()

				mockInvoiceDB.EXPECT().
					FindByID(invoiceID.String()).
					Return(nil, errors.New("invoice not found")).
					Times(1)

				req := httptest.NewRequest("GET", "/invoices/"+invoiceID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("ListInvoices", func() {
			Convey("Success", func() {
				invoices := []model.Invoice{
					{
						ID:          uuid.New(),
						CustomerID:  uuid.New(),
						TotalAmount: 100000.0,
						Status:      "draft",
					},
					{
						ID:          uuid.New(),
						CustomerID:  uuid.New(),
						TotalAmount: 200000.0,
						Status:      "sent",
					},
				}

				mockInvoiceDB.EXPECT().
					FindAll().
					Return(invoices, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/invoices", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response []model.Invoice
				json.Unmarshal(w.Body.Bytes(), &response)
				So(len(response), ShouldEqual, 2)
			})
		})
	})
}
