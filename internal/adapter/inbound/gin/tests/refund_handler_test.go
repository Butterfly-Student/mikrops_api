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

func TestRefundHandler(t *testing.T) {
	Convey("Test Refund HTTP Handler", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockRefundDB := mock_outbound_port.NewMockRefundDatabasePort(mockCtrl)

		mockDatabasePort.EXPECT().Refund().Return(mockRefundDB).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, nil, nil, nil, nil, nil, nil, nil)
		handler := gin_inbound_adapter.NewRefundHttpHandler(dom)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.POST("/refunds", handler.CreateRefund)
		router.GET("/refunds/:id", handler.GetRefund)
		router.GET("/refunds", handler.ListRefunds)
		router.PUT("/refunds/:id", handler.UpdateRefund)
		router.DELETE("/refunds/:id", handler.DeleteRefund)
		router.POST("/refunds/:id/approve", handler.ApproveRefund)
		router.POST("/refunds/:id/reject", handler.RejectRefund)
		router.POST("/refunds/:id/process", handler.ProcessRefund)
		router.POST("/refunds/:id/complete", handler.CompleteRefund)
		router.GET("/refunds/pending", handler.GetPendingRefunds)

		Convey("CreateRefund", func() {
			Convey("Success", func() {
				paymentID := uuid.New()
				refundAmount := 100000.0
				refundType := "partial"
				refundReason := "Customer request"
				refundMethod := "bank_transfer"
				bankName := "BCA"
				bankAccountName := "John Doe"
				bankAccountNumber := "1234567890"

				reqBody := map[string]interface{}{
					"payment_id":          paymentID,
					"refund_amount":       &refundAmount,
					"refund_type":         &refundType,
					"refund_reason":       &refundReason,
					"refund_method":       &refundMethod,
					"bank_name":           &bankName,
					"bank_account_name":   &bankAccountName,
					"bank_account_number": &bankAccountNumber,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/refunds", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				expectedRefund := &model.Refund{
					ID:          uuid.New(),
					PaymentID:   paymentID,
					RefundAmount: refundAmount,
					RefundType:   refundType,
					Status:      "pending",
				}

				mockRefundDB.EXPECT().
					Create(gomock.Any()).
					Return(expectedRefund, nil).
					Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest("POST", "/refunds", bytes.NewBuffer([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("GetRefund", func() {
			Convey("Success", func() {
				refundID := uuid.New()

				expectedRefund := &model.Refund{
					ID:          refundID,
					PaymentID:   uuid.New(),
					RefundAmount: 100000.0,
					Status:      "pending",
				}

				mockRefundDB.EXPECT().
					FindByID(refundID.String()).
					Return(expectedRefund, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/refunds/"+refundID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response model.Refund
				json.Unmarshal(w.Body.Bytes(), &response)
				So(response.ID, ShouldEqual, refundID)
			})

			Convey("Not Found", func() {
				refundID := uuid.New()

				mockRefundDB.EXPECT().
					FindByID(refundID.String()).
					Return(nil, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/refunds/"+refundID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("ListRefunds", func() {
			Convey("Success", func() {
				refunds := []model.Refund{
					{
						ID:          uuid.New(),
						PaymentID:   uuid.New(),
						RefundAmount: 100000.0,
						Status:      "pending",
					},
					{
						ID:          uuid.New(),
						PaymentID:   uuid.New(),
						RefundAmount: 50000.0,
						Status:      "approved",
					},
				}

				mockRefundDB.EXPECT().
					FindByFilter(gomock.Any(), false).
					Return(refunds, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/refunds", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				So(response["total"], ShouldEqual, 2)
			})
		})

		Convey("UpdateRefund", func() {
			Convey("Success", func() {
				refundID := uuid.New()
				refundReason := "Updated reason"

				reqBody := map[string]interface{}{
					"refund_reason": &refundReason,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("PUT", "/refunds/"+refundID.String(), bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				updatedRefund := &model.Refund{
					ID:          refundID,
					PaymentID:   uuid.New(),
					RefundAmount: 100000.0,
					RefundReason: refundReason,
					Status:      "pending",
				}

				mockRefundDB.EXPECT().
					Update(refundID.String(), gomock.Any()).
					Return(updatedRefund, nil).
					Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("DeleteRefund", func() {
			Convey("Success", func() {
				refundID := uuid.New()

				mockRefundDB.EXPECT().
					Delete(refundID.String()).
					Return(nil).
					Times(1)

				req := httptest.NewRequest("DELETE", "/refunds/"+refundID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Error", func() {
				refundID := uuid.New()

				mockRefundDB.EXPECT().
					Delete(refundID.String()).
					Return(errors.New("failed to delete refund")).
					Times(1)

				req := httptest.NewRequest("DELETE", "/refunds/"+refundID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("ApproveRefund", func() {
			Convey("Success", func() {
				refundID := uuid.New()
				approvedBy := "admin@example.com"

				reqBody := map[string]interface{}{
					"approved_by": approvedBy,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/refunds/"+refundID.String()+"/approve", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				approvedRefund := &model.Refund{
					ID:          refundID,
					PaymentID:   uuid.New(),
					RefundAmount: 100000.0,
					Status:      "approved",
					ApprovedBy:  &[]uuid.UUID{uuid.New()}[0],
				}

				mockRefundDB.EXPECT().
					Approve(refundID.String(), approvedBy).
					Return(approvedRefund, nil).
					Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Invalid JSON", func() {
				refundID := uuid.New()

				req := httptest.NewRequest("POST", "/refunds/"+refundID.String()+"/approve", bytes.NewBuffer([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("RejectRefund", func() {
			Convey("Success", func() {
				refundID := uuid.New()
				rejectedBy := "admin@example.com"
				reason := "Duplicate request"

				reqBody := map[string]interface{}{
					"rejected_by": rejectedBy,
					"reason":      reason,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/refunds/"+refundID.String()+"/reject", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				rejectedRefund := &model.Refund{
					ID:              refundID,
					PaymentID:       uuid.New(),
					RefundAmount:    100000.0,
					Status:          "rejected",
					ApprovedBy:      &[]uuid.UUID{uuid.New()}[0],
					RejectionReason: &reason,
				}

				mockRefundDB.EXPECT().
					Reject(refundID.String(), rejectedBy, reason).
					Return(rejectedRefund, nil).
					Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Invalid JSON", func() {
				refundID := uuid.New()

				req := httptest.NewRequest("POST", "/refunds/"+refundID.String()+"/reject", bytes.NewBuffer([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("ProcessRefund", func() {
			Convey("Success", func() {
				refundID := uuid.New()
				processedBy := "admin@example.com"

				reqBody := map[string]interface{}{
					"processed_by": processedBy,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/refunds/"+refundID.String()+"/process", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				processedRefund := &model.Refund{
					ID:          refundID,
					PaymentID:   uuid.New(),
					RefundAmount: 100000.0,
					Status:      "processing",
					ProcessedBy: &[]uuid.UUID{uuid.New()}[0],
				}

				mockRefundDB.EXPECT().
					Process(refundID.String(), processedBy).
					Return(processedRefund, nil).
					Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("CompleteRefund", func() {
			Convey("Success", func() {
				refundID := uuid.New()
				processedBy := "admin@example.com"
				xenditRefundID := "xendit-refund-123"

				reqBody := map[string]interface{}{
					"processed_by":    processedBy,
					"xendit_refund_id": xenditRefundID,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/refunds/"+refundID.String()+"/complete", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				completedRefund := &model.Refund{
					ID:            refundID,
					PaymentID:     uuid.New(),
					RefundAmount:  100000.0,
					Status:        "completed",
					ProcessedBy:   &[]uuid.UUID{uuid.New()}[0],
					XenditRefundID: &xenditRefundID,
				}

				mockRefundDB.EXPECT().
					Complete(refundID.String(), processedBy, xenditRefundID).
					Return(completedRefund, nil).
					Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("GetPendingRefunds", func() {
			Convey("Success", func() {
				refunds := []model.Refund{
					{
						ID:          uuid.New(),
						PaymentID:   uuid.New(),
						RefundAmount: 100000.0,
						Status:      "pending",
					},
				}

				mockRefundDB.EXPECT().
					FindPendingRefunds().
					Return(refunds, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/refunds/pending", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				So(response["total"], ShouldEqual, 1)
			})
		})
	})
}
