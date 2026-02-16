//go:build integration
// +build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	"go-template/internal/domain"
	"go-template/internal/model"
	"go-template/tests/helpers"
)

func TestNotificationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to setup postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	err = pgContainer.DB.AutoMigrate(
		&model.Notification{},
		&model.NotificationTemplate{},
		&model.Customer{},
		&model.Invoice{},
		&model.Payment{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate notification tables: %v", err)
	}

	dbAdapter := postgres_outbound_adapter.NewAdapter(pgContainer.DB)
	notificationDomain := domain.NewNotificationDomain(dbAdapter)

	Convey("Test Notification Integration with PostgreSQL", t, func() {
		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Notification{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.NotificationTemplate{})

		Convey("CreateNotification creates notification successfully", func() {
			notificationType := "email"
			recipient := "customer@example.com"
			subject := "Invoice Due Reminder"
			content := "Dear Customer, your invoice is due soon."

			input := model.NotificationInput{
				Type:      notificationType,
				Recipient: recipient,
				Subject:   &subject,
				Content:   content,
			}

			notification, err := notificationDomain.Create(ctx, input)
			So(err, ShouldBeNil)
			So(notification, ShouldNotBeNil)
			So(notification.Type, ShouldEqual, "email")
			So(notification.Recipient, ShouldEqual, "customer@example.com")
			So(notification.Subject, ShouldNotBeNil)
			So(*notification.Subject, ShouldEqual, "Invoice Due Reminder")
			So(notification.Content, ShouldEqual, "Dear Customer, your invoice is due soon.")
			So(notification.Status, ShouldEqual, "pending")
			So(notification.RetryCount, ShouldEqual, 0)

			Convey("GetNotification retrieves created notification", func() {
				found, err := notificationDomain.Get(ctx, notification.ID.String())
				So(err, ShouldBeNil)
				So(found.ID, ShouldEqual, notification.ID)
				So(found.Recipient, ShouldEqual, recipient)
			})

			Convey("ListNotifications returns notifications", func() {
				filter := model.NotificationFilter{}
				notifications, err := notificationDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(notifications), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("ListNotifications with type filter", func() {
				filter := model.NotificationFilter{
					Types: []string{"email"},
				}
				notifications, err := notificationDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(notifications), ShouldBeGreaterThanOrEqualTo, 1)
				for _, n := range notifications {
					So(n.Type, ShouldEqual, "email")
				}
			})

			Convey("ListNotifications with status filter", func() {
				filter := model.NotificationFilter{
					Status: []string{"pending"},
				}
				notifications, err := notificationDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(notifications), ShouldBeGreaterThanOrEqualTo, 1)
				for _, n := range notifications {
					So(n.Status, ShouldEqual, "pending")
				}
			})

			Convey("SendNotification marks as sent", func() {
				err := notificationDomain.Send(ctx, notification.ID.String())
				So(err, ShouldBeNil)

				updated, err := notificationDomain.Get(ctx, notification.ID.String())
				So(err, ShouldBeNil)
				So(updated.Status, ShouldEqual, "sent")
				So(updated.SentAt, ShouldNotBeNil)
			})

			Convey("MarkAsFailed marks notification as failed", func() {
				errorCode := "SMTP_ERROR"
				errorMsg := "Failed to connect to SMTP server"

				err := notificationDomain.MarkAsFailed(ctx, notification.ID.String(), errorCode, errorMsg)
				So(err, ShouldBeNil)

				updated, err := notificationDomain.Get(ctx, notification.ID.String())
				So(err, ShouldBeNil)
				So(updated.Status, ShouldEqual, "failed")
				So(updated.ErrorCode, ShouldNotBeNil)
				So(*updated.ErrorCode, ShouldEqual, "SMTP_ERROR")
				So(updated.ErrorMsg, ShouldNotBeNil)
				So(*updated.ErrorMsg, ShouldEqual, "Failed to connect to SMTP server")
			})

			Convey("RetryNotification increments retry count", func() {
				err := notificationDomain.Retry(ctx, notification.ID.String())
				So(err, ShouldBeNil)

				updated, err := notificationDomain.Get(ctx, notification.ID.String())
				So(err, ShouldBeNil)
				So(updated.Status, ShouldEqual, "retrying")
				So(updated.RetryCount, ShouldEqual, 1)
			})
		})

		Convey("Different notification types", func() {
			Convey("Email notification", func() {
				input := model.NotificationInput{
					Type:      "email",
					Recipient: "email@example.com",
					Subject:   func() *string { s := "Email Subject"; return &s }(),
					Content:   "Email content",
				}

				notification, err := notificationDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(notification.Type, ShouldEqual, "email")
			})

			Convey("WhatsApp notification", func() {
				input := model.NotificationInput{
					Type:      "whatsapp",
					Recipient: "6281234567890",
					Content:   "WhatsApp message content",
				}

				notification, err := notificationDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(notification.Type, ShouldEqual, "whatsapp")
			})

			Convey("SMS notification", func() {
				input := model.NotificationInput{
					Type:      "sms",
					Recipient: "6281234567890",
					Content:   "SMS message content",
				}

				notification, err := notificationDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(notification.Type, ShouldEqual, "sms")
			})
		})

		Convey("Notification with related entities", func() {
			// Create related entities
			profile := &model.BandwidthProfile{
				Name:         "Notif Test Profile",
				Category:     "pppoe",
				PriceMonthly: 100000,
				TaxRate:      0.11,
				IsActive:     true,
			}
			err := pgContainer.DB.Create(profile).Error
			So(err, ShouldBeNil)

			customerCode := "NOTIF-" + time.Now().Format("20060102150405")
			email := "notifcustomer@example.com"
			expiryDate := time.Now().AddDate(0, 1, 0)
			customer := &model.Customer{
				CustomerCode: customerCode,
				FullName:     "Notification Test Customer",
				Email:        &email,
				Phone:        "08123456789",
				Status:       "active",
				ExpiryDate:   &expiryDate,
				ProfileID:    &profile.ID,
			}
			err = pgContainer.DB.Create(customer).Error
			So(err, ShouldBeNil)

			// Create invoice
			invoice := &model.Invoice{
				CustomerID:  customer.ID,
				InvoiceNumber: "INV-" + time.Now().Format("20060102150405"),
				TotalAmount: 111000.0,
				Status:       "sent",
			}
			model.InvoicePrepare(invoice)
			err = pgContainer.DB.Create(invoice).Error
			So(err, ShouldBeNil)

			// Create payment
			payment := &model.Payment{
				CustomerID:    customer.ID,
				InvoiceID:     &invoice.ID,
				Amount:        111000.0,
				PaymentMethod: "bank_transfer",
				PaymentDate:   time.Now(),
				Status:        "confirmed",
			}
			model.PaymentPrepare(payment)
			err = pgContainer.DB.Create(payment).Error
			So(err, ShouldBeNil)

			Convey("Create notification with customer reference", func() {
				input := model.NotificationInput{
					Type:       "email",
					Recipient:  "customer@example.com",
					Subject:    func() *string { s := "Welcome"; return &s }(),
					Content:    "Welcome to our service",
					CustomerID: &customer.ID,
				}

				notification, err := notificationDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(notification.CustomerID, ShouldNotBeNil)
				So(*notification.CustomerID, ShouldEqual, customer.ID)

				Convey("Filter by customer", func() {
					filter := model.NotificationFilter{
						CustomerIDs: []uuid.UUID{customer.ID},
					}
					notifications, err := notificationDomain.List(ctx, filter)
					So(err, ShouldBeNil)
					So(len(notifications), ShouldBeGreaterThanOrEqualTo, 1)
				})
			})

			Convey("Create notification with invoice reference", func() {
				input := model.NotificationInput{
					Type:      "email",
					Recipient: "customer@example.com",
					Subject:   func() *string { s := "Invoice Reminder"; return &s }(),
					Content:   "Your invoice is due",
					InvoiceID: &invoice.ID,
				}

				notification, err := notificationDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(notification.InvoiceID, ShouldNotBeNil)
				So(*notification.InvoiceID, ShouldEqual, invoice.ID)

				Convey("Filter by invoice", func() {
					filter := model.NotificationFilter{
						InvoiceIDs: []uuid.UUID{invoice.ID},
					}
					notifications, err := notificationDomain.List(ctx, filter)
					So(err, ShouldBeNil)
					So(len(notifications), ShouldBeGreaterThanOrEqualTo, 1)
				})
			})

			Convey("Create notification with payment reference", func() {
				input := model.NotificationInput{
					Type:      "whatsapp",
					Recipient: "6281234567890",
					Content:   "Payment received successfully",
					PaymentID: &payment.ID,
				}

				notification, err := notificationDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(notification.PaymentID, ShouldNotBeNil)
				So(*notification.PaymentID, ShouldEqual, payment.ID)

				Convey("Filter by payment", func() {
					filter := model.NotificationFilter{
						PaymentIDs: []uuid.UUID{payment.ID},
					}
					notifications, err := notificationDomain.List(ctx, filter)
					So(err, ShouldBeNil)
					So(len(notifications), ShouldBeGreaterThanOrEqualTo, 1)
				})
			})
		})

		Convey("Scheduled notifications", func() {
			scheduledAt := time.Now().Add(time.Hour * 24)

			input := model.NotificationInput{
				Type:        "email",
				Recipient:   "scheduled@example.com",
				Subject:     func() *string { s := "Scheduled Email"; return &s }(),
				Content:     "This is a scheduled notification",
				ScheduledAt: &scheduledAt,
			}

			notification, err := notificationDomain.Create(ctx, input)
			So(err, ShouldBeNil)
			So(notification.ScheduledAt, ShouldNotBeNil)
			So(notification.ScheduledAt.Day(), ShouldEqual, scheduledAt.Day())

			Convey("Filter by scheduled time range", func() {
				startTime := time.Now()
				endTime := time.Now().Add(time.Hour * 48)

				filter := model.NotificationFilter{
					SentStart: &startTime,
					SentEnd:   &endTime,
				}
				notifications, err := notificationDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(notifications), ShouldBeGreaterThanOrEqualTo, 1)
			})
		})

		Convey("Notification with retry logic", func() {
			input := model.NotificationInput{
				Type:      "email",
				Recipient: "retry@example.com",
				Subject:   func() *string { s := "Test Retry"; return &s }(),
				Content:   "Test retry logic",
			}

			notification, err := notificationDomain.Create(ctx, input)
			So(err, ShouldBeNil)

			// Simulate multiple retries
			for i := 1; i <= 3; i++ {
				err := notificationDomain.Retry(ctx, notification.ID.String())
				So(err, ShouldBeNil)

				updated, err := notificationDomain.Get(ctx, notification.ID.String())
				So(err, ShouldBeNil)
				So(updated.RetryCount, ShouldEqual, i)
				So(updated.Status, ShouldEqual, "retrying")
			}

			Convey("Max retry count validation", func() {
				updated, err := notificationDomain.Get(ctx, notification.ID.String())
				So(err, ShouldBeNil)
				So(updated.RetryCount, ShouldEqual, 3)
			})
		})

		Convey("NotificationTemplate management", func() {
			Convey("CreateNotificationTemplate creates template successfully", func() {
				name := "invoice_reminder"
				templateType := "email"
				subject := "Invoice Payment Reminder - {{.InvoiceNumber}}"
				content := "Dear {{.CustomerName}}, your invoice {{.InvoiceNumber}} is due on {{.DueDate}}. Amount: {{.Amount}}"
				variables := `["InvoiceNumber", "CustomerName", "DueDate", "Amount"]`

				input := model.NotificationTemplateInput{
					Name:      name,
					Type:      templateType,
					Subject:   subject,
					Content:   content,
					IsActive:  func() *bool { b := true; return &b }(),
					Variables: variables,
				}

				template, err := notificationDomain.CreateTemplate(ctx, input)
				So(err, ShouldBeNil)
				So(template.Name, ShouldEqual, "invoice_reminder")
				So(template.Type, ShouldEqual, "email")
				So(template.Subject, ShouldEqual, "Invoice Payment Reminder - {{.InvoiceNumber}}")
				So(template.Content, ShouldEqual, "Dear {{.CustomerName}}, your invoice {{.InvoiceNumber}} is due on {{.DueDate}}. Amount: {{.Amount}}")
				So(template.IsActive, ShouldEqual, true)

				Convey("GetTemplate retrieves created template", func() {
					found, err := notificationDomain.GetTemplate(ctx, template.ID.String())
					So(err, ShouldBeNil)
					So(found.ID, ShouldEqual, template.ID)
					So(found.Name, ShouldEqual, name)
				})

				Convey("UpdateTemplate updates template", func() {
					newSubject := "Updated: Invoice Reminder"
					newContent := "Updated content"

					input := model.NotificationTemplateInput{
						Subject:  newSubject,
						Content:  newContent,
						IsActive: func() *bool { b := false; return &b }(),
					}

					updated, err := notificationDomain.UpdateTemplate(ctx, template.ID.String(), input)
					So(err, ShouldBeNil)
					So(updated.Subject, ShouldEqual, "Updated: Invoice Reminder")
					So(updated.Content, ShouldEqual, "Updated content")
					So(updated.IsActive, ShouldEqual, false)
				})

				Convey("DeleteTemplate removes template", func() {
					err := notificationDomain.DeleteTemplate(ctx, template.ID.String())
					So(err, ShouldBeNil)

					// Verify it's deleted
					_, err = notificationDomain.GetTemplate(ctx, template.ID.String())
					So(err, ShouldNotBeNil)
				})
			})

			Convey("Create duplicate template name returns error", func() {
				name := "duplicate_template"
				templateType := "email"

				input1 := model.NotificationTemplateInput{
					Name:     name,
					Type:     templateType,
					Subject:  "Subject 1",
					Content:  "Content 1",
					IsActive: func() *bool { b := true; return &b }(),
				}

				_, err := notificationDomain.CreateTemplate(ctx, input1)
				So(err, ShouldBeNil)

				input2 := model.NotificationTemplateInput{
					Name:     name,
					Type:     templateType,
					Subject:  "Subject 2",
					Content:  "Content 2",
					IsActive: func() *bool { b := true; return &b }(),
				}

				_, err = notificationDomain.CreateTemplate(ctx, input2)
				So(err, ShouldNotBeNil)
			})

			Convey("Different template types", func() {
				types := []string{"email", "whatsapp", "sms"}

				for _, templateType := range types {
					name := "template_" + templateType + "_" + time.Now().Format("20060102150405")
					input := model.NotificationTemplateInput{
						Name:     name,
						Type:     templateType,
						Subject:  "Test Subject",
						Content:  "Test content",
						IsActive: func() *bool { b := true; return &b }(),
					}

					template, err := notificationDomain.CreateTemplate(ctx, input)
					So(err, ShouldBeNil)
					So(template.Type, ShouldEqual, templateType)
				}
			})
		})

		Convey("Error handling", func() {
			Convey("GetNotification with invalid ID returns error", func() {
				_, err := notificationDomain.Get(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("GetTemplate with invalid ID returns error", func() {
				_, err := notificationDomain.GetTemplate(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("UpdateTemplate with invalid ID returns error", func() {
				input := model.NotificationTemplateInput{
					Subject: "Updated",
				}
				_, err := notificationDomain.UpdateTemplate(ctx, uuid.New().String(), input)
				So(err, ShouldNotBeNil)
			})

			Convey("DeleteTemplate with invalid ID returns error", func() {
				err := notificationDomain.DeleteTemplate(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("Send notification with invalid ID returns error", func() {
				err := notificationDomain.Send(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("MarkAsFailed with invalid ID returns error", func() {
				err := notificationDomain.MarkAsFailed(ctx, uuid.New().String(), "ERROR", "Error message")
				So(err, ShouldNotBeNil)
			})

			Convey("Retry with invalid ID returns error", func() {
				err := notificationDomain.Retry(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("Create notification with invalid type returns error", func() {
				input := model.NotificationInput{
					Type:      "invalid_type",
					Recipient: "test@example.com",
					Content:   "Test",
				}

				_, err := notificationDomain.Create(ctx, input)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Notification search", func() {
			// Create multiple notifications
			for i := 0; i < 3; i++ {
				input := model.NotificationInput{
					Type:      "email",
					Recipient: "search" + string(rune('1'+i)) + "@example.com",
					Subject:   func() *string { s := "Search Test " + string(rune('1'+i)); return &s }(),
					Content:   "Searchable content " + string(rune('1'+i)),
				}

				_, err := notificationDomain.Create(ctx, input)
				So(err, ShouldBeNil)
			}

			Convey("Search notifications by content", func() {
				searchTerm := "Searchable"
				filter := model.NotificationFilter{
					Search: &searchTerm,
				}
				notifications, err := notificationDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(notifications), ShouldBeGreaterThanOrEqualTo, 3)
			})

			Convey("Search notifications by subject", func() {
				searchTerm := "Search Test"
				filter := model.NotificationFilter{
					Search: &searchTerm,
				}
				notifications, err := notificationDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(notifications), ShouldBeGreaterThanOrEqualTo, 3)
			})
		})
	})
}
