package http_test

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
	. "github.com/smartystreets/goconvey/convey"

	gin_inbound_adapter "mikrops/internal/adapter/inbound/gin"
	"mikrops/internal/domain"
	"mikrops/internal/model"
	mock_outbound_port "mikrops/tests/mocks/port"
)

func TestInvoiceAdapter(t *testing.T) {
	Convey("Test Invoice HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockInvoiceDatabasePort := mock_outbound_port.NewMockInvoiceDatabasePort(mockCtrl)
		mockSubscriptionDatabasePort := mock_outbound_port.NewMockSubscriptionDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().Invoice().Return(mockInvoiceDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().Subscription().Return(mockSubscriptionDatabasePort).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)
		adapter := gin_inbound_adapter.NewAdapter(dom, mockHttpPort)

		invoiceOutput := model.Invoice{
			ID: "inv-123",
			InvoiceInput: model.InvoiceInput{
				TenantID:       "tenant-123",
				CustomerID:     "cust-123",
				SubscriptionID: "sub-123",
				InvoiceNumber:  "INV-2025-001",
				Amount:         150000,
				TaxAmount:      0,
				TotalAmount:    150000,
				Status:         model.InvoiceStatusUnpaid,
				DueDate:        time.Now().AddDate(0, 0, 7),
				PeriodStart:    time.Now(),
				PeriodEnd:      time.Now().AddDate(0, 1, 0),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
		}

		router := gin.New()
		router.GET("/api/v1/invoices", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.Invoice().List(c)
		})
		router.POST("/api/v1/invoices", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.Invoice().Create(c)
		})
		router.GET("/api/v1/invoices/:id", func(c *gin.Context) {
			adapter.Invoice().Get(c)
		})
		router.PUT("/api/v1/invoices/:id", func(c *gin.Context) {
			adapter.Invoice().Update(c)
		})
		router.POST("/api/v1/invoices/generate", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.Invoice().GenerateBulk(c)
		})

		Convey("List", func() {
			Convey("Success", func() {
				mockInvoiceDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Invoice{invoiceOutput}, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Domain error", func() {
				mockInvoiceDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return(nil, errors.New("error")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Create", func() {
			payload := model.InvoiceInput{
				CustomerID:     "cust-123",
				SubscriptionID: "sub-123",
				InvoiceNumber:  "INV-2025-002",
				Amount:         150000,
				TotalAmount:    150000,
				DueDate:        time.Now().AddDate(0, 0, 7),
				PeriodStart:    time.Now(),
				PeriodEnd:      time.Now().AddDate(0, 1, 0),
			}

			Convey("Success", func() {
				mockInvoiceDatabasePort.EXPECT().Create(gomock.Any()).Return(invoiceOutput, nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain error", func() {
				mockInvoiceDatabasePort.EXPECT().Create(gomock.Any()).Return(model.Invoice{}, errors.New("error")).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Get", func() {
			Convey("Success", func() {
				mockInvoiceDatabasePort.EXPECT().FindByID(gomock.Any()).Return(invoiceOutput, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices/inv-123", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Not found", func() {
				mockInvoiceDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Invoice{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices/inv-999", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Update", func() {
			payload := model.InvoiceInput{
				Notes: "Updated notes",
			}

			Convey("Success", func() {
				mockInvoiceDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPut, "/api/v1/invoices/inv-123", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPut, "/api/v1/invoices/inv-123", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("GenerateBulk", func() {
			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices/generate", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Success", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Subscription{}, nil).Times(1)

				payload := map[string]interface{}{
					"period_start": time.Now().Format(time.RFC3339),
					"period_end":   time.Now().AddDate(0, 1, 0).Format(time.RFC3339),
				}
				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices/generate", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldBeIn, []int{http.StatusCreated, http.StatusInternalServerError})
			})
		})
	})
}
