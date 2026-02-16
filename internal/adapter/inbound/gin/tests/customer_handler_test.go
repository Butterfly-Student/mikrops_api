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

func TestCustomerHandler(t *testing.T) {
	Convey("Test Customer HTTP Handler", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(mockCtrl)
		mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(mockCtrl)
		mockBandwidthProfileDB := mock_outbound_port.NewMockBandwidthProfileDatabasePort(mockCtrl)

		mockDatabasePort.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()
		mockDatabasePort.EXPECT().BandwidthProfile().Return(mockBandwidthProfileDB).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, nil, nil, nil, mockMikrotikPort, nil, nil, nil)
		handler := gin_inbound_adapter.NewCustomerHandler(dom)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.POST("/customers", handler.CreateCustomer)
		router.GET("/customers/:id", handler.GetCustomer)
		router.GET("/customers", handler.ListCustomers)
		router.PUT("/customers/:id", handler.UpdateCustomer)
		router.DELETE("/customers/:id", handler.DeleteCustomer)

		Convey("CreateCustomer", func() {
			Convey("Success", func() {
				customerCode := "CUST-001"
				fullName := "John Doe"
				phone := "081234567890"
				profileID := uuid.New()
				expiryDate := "2024-12-31T00:00:00Z"

				reqBody := map[string]interface{}{
					"customer_code": customerCode,
					"full_name":     fullName,
					"phone":         phone,
					"profile_id":    profileID,
					"expiry_date":   expiryDate,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/customers", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				expectedCustomer := &model.Customer{
					ID:           uuid.New(),
					CustomerCode: customerCode,
					FullName:     fullName,
					Phone:        phone,
				}

				mockDatabasePort.EXPECT().
					DoInTransaction(gomock.Any()).
					DoAndReturn(func(txFunc interface{}) (interface{}, error) {
						return expectedCustomer, nil
					}).Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest("POST", "/customers", bytes.NewBuffer([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("GetCustomer", func() {
			Convey("Success", func() {
				customerID := uuid.New()

				expectedCustomer := &model.Customer{
					ID:           customerID,
					CustomerCode: "CUST-001",
					FullName:     "John Doe",
					Phone:        "081234567890",
				}

				mockCustomerDB.EXPECT().
					FindByID(customerID.String()).
					Return(expectedCustomer, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/customers/"+customerID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response model.Customer
				json.Unmarshal(w.Body.Bytes(), &response)
				So(response.ID, ShouldEqual, customerID)
			})

			Convey("Not Found", func() {
				customerID := uuid.New()

				mockCustomerDB.EXPECT().
					FindByID(customerID.String()).
					Return(nil, errors.New("customer not found")).
					Times(1)

				req := httptest.NewRequest("GET", "/customers/"+customerID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("ListCustomers", func() {
			Convey("Success", func() {
				customers := []model.Customer{
					{
						ID:           uuid.New(),
						CustomerCode: "CUST-001",
						FullName:     "John Doe",
						Phone:        "081234567890",
					},
					{
						ID:           uuid.New(),
						CustomerCode: "CUST-002",
						FullName:     "Jane Smith",
						Phone:        "081234567891",
					},
				}

				mockCustomerDB.EXPECT().
					FindAll().
					Return(customers, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/customers", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response []model.Customer
				json.Unmarshal(w.Body.Bytes(), &response)
				So(len(response), ShouldEqual, 2)
			})
		})

		Convey("UpdateCustomer", func() {
			Convey("Success", func() {
				customerID := uuid.New()
				fullName := "John Doe Updated"

				reqBody := map[string]interface{}{
					"full_name": fullName,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("PUT", "/customers/"+customerID.String(), bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				updatedCustomer := &model.Customer{
					ID:           customerID,
					CustomerCode: "CUST-001",
					FullName:     fullName,
					Phone:        "081234567890",
				}

				mockDatabasePort.EXPECT().
					DoInTransaction(gomock.Any()).
					DoAndReturn(func(txFunc interface{}) (interface{}, error) {
						return updatedCustomer, nil
					}).Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("DeleteCustomer", func() {
			Convey("Success", func() {
				customerID := uuid.New()

				customer := &model.Customer{
					ID:           customerID,
					CustomerCode: "CUST-001",
					FullName:     "John Doe",
				}

				mockCustomerDB.EXPECT().
					FindByID(customerID.String()).
					Return(customer, nil).
					Times(1)

				mockCustomerDB.EXPECT().
					Delete(customerID.String()).
					Return(nil).
					Times(1)

				req := httptest.NewRequest("DELETE", "/customers/"+customerID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}
