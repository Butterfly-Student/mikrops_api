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

func TestNotificationHandler(t *testing.T) {
	Convey("Test Notification HTTP Handler", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockNotificationDB := mock_outbound_port.NewMockNotificationDatabasePort(mockCtrl)

		mockDatabasePort.EXPECT().Notification().Return(mockNotificationDB).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, nil, nil, nil, nil, nil, nil, nil)
		handler := gin_inbound_adapter.NewNotificationHandler(dom)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.POST("/notifications", handler.CreateNotification)
		router.GET("/notifications/:id", handler.GetNotification)
		router.GET("/notifications", handler.ListNotifications)

		Convey("CreateNotification", func() {
			Convey("Success", func() {
				customerID := uuid.New()
				subject := "Payment Confirmation"
				content := "Your payment has been confirmed"

				reqBody := map[string]interface{}{
					"customer_id": customerID.String(),
					"type":        "email",
					"recipient":   "customer@example.com",
					"subject":     subject,
					"content":     content,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/notifications", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				expectedNotification := &model.Notification{
					ID:         uuid.New(),
					CustomerID: &customerID,
					Type:       "email",
					Recipient:  "customer@example.com",
					Subject:    &subject,
					Content:    content,
					Status:     "pending",
				}

				mockNotificationDB.EXPECT().
					Create(gomock.Any()).
					DoAndReturn(func(notification *model.Notification) error {
						notification.ID = expectedNotification.ID
						return nil
					}).Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest("POST", "/notifications", bytes.NewBuffer([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("GetNotification", func() {
			Convey("Success", func() {
				notificationID := uuid.New()
				subject := "Payment Confirmation"

				expectedNotification := &model.Notification{
					ID:        notificationID,
					Type:      "email",
					Recipient: "customer@example.com",
					Subject:   &subject,
					Content:   "Your payment has been confirmed",
					Status:    "sent",
				}

				mockNotificationDB.EXPECT().
					FindByID(notificationID.String()).
					Return(expectedNotification, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/notifications/"+notificationID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Not Found", func() {
				notificationID := uuid.New()

				mockNotificationDB.EXPECT().
					FindByID(notificationID.String()).
					Return(nil, errors.New("notification not found")).
					Times(1)

				req := httptest.NewRequest("GET", "/notifications/"+notificationID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("ListNotifications", func() {
			Convey("Success", func() {
				notifications := []model.Notification{
					{
						ID:        uuid.New(),
						Type:      "email",
						Recipient: "customer1@example.com",
						Content:   "Test notification 1",
						Status:    "sent",
					},
					{
						ID:        uuid.New(),
						Type:      "email",
						Recipient: "customer2@example.com",
						Content:   "Test notification 2",
						Status:    "pending",
					},
				}

				mockNotificationDB.EXPECT().
					FindAll().
					Return(notifications, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/notifications", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response struct {
					Count         int                    `json:"count"`
					Notifications []model.Notification `json:"notifications"`
				}
				json.Unmarshal(w.Body.Bytes(), &response)
				So(len(response.Notifications), ShouldEqual, 2)
				So(response.Count, ShouldEqual, 2)
			})
		})
	})
}
