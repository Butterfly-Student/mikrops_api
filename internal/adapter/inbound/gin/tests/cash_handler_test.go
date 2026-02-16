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

func TestCashHandler(t *testing.T) {
	Convey("Test Cash HTTP Handler", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockCashCategoryDB := mock_outbound_port.NewMockCashCategoryDatabasePort(mockCtrl)
		mockCashTransactionDB := mock_outbound_port.NewMockCashTransactionDatabasePort(mockCtrl)

		mockDatabasePort.EXPECT().CashCategory().Return(mockCashCategoryDB).AnyTimes()
		mockDatabasePort.EXPECT().CashTransaction().Return(mockCashTransactionDB).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, nil, nil, nil, nil, nil, nil, nil)
		handler := gin_inbound_adapter.NewCashHandler(dom)

		gin.SetMode(gin.TestMode)
		router := gin.New()

		// Category routes
		router.POST("/cash/categories", handler.CreateCashCategory)
		router.GET("/cash/categories/:id", handler.GetCashCategory)
		router.GET("/cash/categories", handler.ListCashCategories)
		router.PUT("/cash/categories/:id", handler.UpdateCashCategory)
		router.DELETE("/cash/categories/:id", handler.DeleteCashCategory)

		// Transaction routes
		router.POST("/cash/transactions", handler.CreateCashTransaction)
		router.GET("/cash/transactions/:id", handler.GetCashTransaction)
		router.GET("/cash/transactions", handler.ListCashTransactions)
		router.PUT("/cash/transactions/:id", handler.UpdateCashTransaction)
		router.DELETE("/cash/transactions/:id", handler.DeleteCashTransaction)
		router.POST("/cash/transactions/:id/approve", handler.ApproveCashTransaction)
		router.POST("/cash/transactions/:id/reject", handler.RejectCashTransaction)
		router.GET("/cash/balance", handler.GetCashBalance)

		Convey("CreateCashCategory", func() {
			Convey("Success", func() {
				code := "INC-001"
				name := "Payment Received"
				cashType := "income"

				reqBody := map[string]interface{}{
					"code": &code,
					"name": &name,
					"type": &cashType,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/cash/categories", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				expectedCategory := &model.CashCategory{
					ID:   uuid.New(),
					Code: code,
					Name: name,
					Type: cashType,
				}

				mockCashCategoryDB.EXPECT().
					Create(gomock.Any()).
					DoAndReturn(func(category *model.CashCategory) error {
						category.ID = expectedCategory.ID
						return nil
					}).Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest("POST", "/cash/categories", bytes.NewBuffer([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("GetCashCategory", func() {
			Convey("Success", func() {
				categoryID := uuid.New()

				expectedCategory := &model.CashCategory{
					ID:   categoryID,
					Code: "INC-001",
					Name: "Payment Received",
					Type: "income",
				}

				mockCashCategoryDB.EXPECT().
					FindByID(categoryID.String()).
					Return(expectedCategory, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/cash/categories/"+categoryID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response model.CashCategory
				json.Unmarshal(w.Body.Bytes(), &response)
				So(response.ID, ShouldEqual, categoryID)
			})

			Convey("Not Found", func() {
				categoryID := uuid.New()

				mockCashCategoryDB.EXPECT().
					FindByID(categoryID.String()).
					Return(nil, errors.New("category not found")).
					Times(1)

				req := httptest.NewRequest("GET", "/cash/categories/"+categoryID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("ListCashCategories", func() {
			Convey("Success", func() {
				categories := []model.CashCategory{
					{
						ID:   uuid.New(),
						Code: "INC-001",
						Name: "Payment Received",
						Type: "income",
					},
					{
						ID:   uuid.New(),
						Code: "EXP-001",
						Name: "Office Rent",
						Type: "expense",
					},
				}

				mockCashCategoryDB.EXPECT().
					Find(gomock.Any()).
					Return(categories, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/cash/categories", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response []model.CashCategory
				json.Unmarshal(w.Body.Bytes(), &response)
				So(len(response), ShouldEqual, 2)
			})
		})

		Convey("UpdateCashCategory", func() {
			Convey("Success", func() {
				categoryID := uuid.New()
				name := "Updated Category Name"

				reqBody := map[string]interface{}{
					"name": &name,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("PUT", "/cash/categories/"+categoryID.String(), bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				updatedCategory := &model.CashCategory{
					ID:   categoryID,
					Code: "INC-001",
					Name: name,
					Type: "income",
				}

				mockCashCategoryDB.EXPECT().
					FindByID(categoryID.String()).
					Return(updatedCategory, nil).
					Times(1)

				mockCashCategoryDB.EXPECT().
					Update(gomock.Any()).
					Return(nil).
					Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("DeleteCashCategory", func() {
			Convey("Success", func() {
				categoryID := uuid.New()

				mockCashCategoryDB.EXPECT().
					Delete(categoryID.String()).
					Return(nil).
					Times(1)

				req := httptest.NewRequest("DELETE", "/cash/categories/"+categoryID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("CreateCashTransaction", func() {
			Convey("Success", func() {
				categoryID := uuid.New()
				amount := 100000.0
				description := "Payment received"

				reqBody := map[string]interface{}{
					"category_id":  categoryID,
					"amount":       &amount,
					"description":  &description,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/cash/transactions", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				mockCashTransactionDB.EXPECT().
					Create(gomock.Any()).
					DoAndReturn(func(transaction *model.CashTransaction) error {
						transaction.ID = uuid.New()
						return nil
					}).Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)
			})
		})

		Convey("GetCashTransaction", func() {
			Convey("Success", func() {
				transactionID := uuid.New()

				expectedTransaction := &model.CashTransaction{
					ID:     transactionID,
					Amount: 100000.0,
					ApprovalStatus: "pending",
				}

				mockCashTransactionDB.EXPECT().
					FindByID(transactionID.String()).
					Return(expectedTransaction, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/cash/transactions/"+transactionID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("ListCashTransactions", func() {
			Convey("Success", func() {
				transactions := []model.CashTransaction{
					{
						ID:     uuid.New(),
						Amount: 100000.0,
						ApprovalStatus: "approved",
					},
					{
						ID:     uuid.New(),
						Amount: 50000.0,
						ApprovalStatus: "pending",
					},
				}

				mockCashTransactionDB.EXPECT().
					Find(gomock.Any()).
					Return(transactions, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/cash/transactions", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("ApproveCashTransaction", func() {
			Convey("Success", func() {
				transactionID := uuid.New()

				existingTransaction := &model.CashTransaction{
					ID:     transactionID,
					Amount: 100000.0,
					ApprovalStatus: "pending",
				}

				mockCashTransactionDB.EXPECT().
					FindByID(transactionID.String()).
					Return(existingTransaction, nil).
					Times(1)

				mockCashTransactionDB.EXPECT().
					Update(gomock.Any()).
					Return(nil).
					Times(1)

				req := httptest.NewRequest("POST", "/cash/transactions/"+transactionID.String()+"/approve", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("GetCashBalance", func() {
			Convey("Success", func() {
				_ = time.Now().Add(-30 * 24 * time.Hour)
				_ = time.Now()

				incomeTransactions := []model.CashTransaction{
					{Amount: 100000.0, ApprovalStatus: "approved"},
					{Amount: 150000.0, ApprovalStatus: "approved"},
				}

				expenseTransactions := []model.CashTransaction{
					{Amount: 50000.0, ApprovalStatus: "approved"},
				}

				mockCashCategoryDB.EXPECT().
					Find(gomock.Any()).
					Return([]model.CashCategory{}, nil).
					Times(2)

				mockCashTransactionDB.EXPECT().
					Find(gomock.Any()).
					Return(incomeTransactions, nil).
					Times(1)

				mockCashTransactionDB.EXPECT().
					Find(gomock.Any()).
					Return(expenseTransactions, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/cash/balance", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}
