//go:build integration
// +build integration

package integration_test

import (
	"context"
	"encoding/json"
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

func TestActivityIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to setup postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	err = pgContainer.DB.AutoMigrate(&model.ActivityLog{})
	if err != nil {
		t.Fatalf("Failed to migrate activity log table: %v", err)
	}

	dbAdapter := postgres_outbound_adapter.NewAdapter(pgContainer.DB)
	activityDomain := domain.NewActivityDomain(dbAdapter)

	Convey("Test Activity Log Integration with PostgreSQL", t, func() {
		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.ActivityLog{})

		Convey("CreateActivityLog creates log successfully", func() {
			userID := uuid.New()
			entityID := uuid.New()
			description := "Created new customer"

			input := model.ActivityLogInput{
				UserID:      &userID,
				Action:      "create",
				EntityType:  "customer",
				EntityID:    &entityID,
				Description: description,
				IPAddress:   "192.168.1.100",
				UserAgent:   "Mozilla/5.0 Test Browser",
			}

			activity, err := activityDomain.Create(ctx, input)
			So(err, ShouldBeNil)
			So(activity, ShouldNotBeNil)
			So(activity.Action, ShouldEqual, "create")
			So(activity.EntityType, ShouldEqual, "customer")
			So(activity.Description, ShouldEqual, description)
			So(activity.IPAddress, ShouldEqual, "192.168.1.100")

			Convey("GetActivityLog retrieves created log", func() {
				found, err := activityDomain.Get(ctx, activity.ID.String())
				So(err, ShouldBeNil)
				So(found.ID, ShouldEqual, activity.ID)
				So(found.Description, ShouldEqual, description)
			})

			Convey("ListActivityLogs returns logs", func() {
				filter := model.ActivityLogFilter{}
				logs, err := activityDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(logs), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("ListActivityLogs with action filter", func() {
				filter := model.ActivityLogFilter{
					Actions: []string{"create"},
				}
				logs, err := activityDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(logs), ShouldBeGreaterThanOrEqualTo, 1)
				for _, log := range logs {
					So(log.Action, ShouldEqual, "create")
				}
			})

			Convey("ListActivityLogs with entity type filter", func() {
				filter := model.ActivityLogFilter{
					EntityTypes: []string{"customer"},
				}
				logs, err := activityDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(logs), ShouldBeGreaterThanOrEqualTo, 1)
				for _, log := range logs {
					So(log.EntityType, ShouldEqual, "customer")
				}
			})

			Convey("ListActivityLogs with user filter", func() {
				filter := model.ActivityLogFilter{
					UserIDs: []uuid.UUID{userID},
				}
				logs, err := activityDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(logs), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("ListActivityLogs with entity filter", func() {
				filter := model.ActivityLogFilter{
					EntityIDs: []uuid.UUID{entityID},
				}
				logs, err := activityDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(logs), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("ListActivityLogs with date range filter", func() {
				now := time.Now()
				startTime := now.Add(-time.Hour)
				endTime := now.Add(time.Hour)

				filter := model.ActivityLogFilter{
					CreatedAtStart: &startTime,
					CreatedAtEnd:   &endTime,
				}
				logs, err := activityDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(logs), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("ListActivityLogs with pagination", func() {
				// Create additional logs
				for i := 0; i < 5; i++ {
					newEntityID := uuid.New()
					input := model.ActivityLogInput{
						UserID:      &userID,
						Action:      "create",
						EntityType:  "customer",
						EntityID:    &newEntityID,
						Description: "Test log " + string(rune('0'+i)),
					}
					_, err := activityDomain.Create(ctx, input)
					So(err, ShouldBeNil)
				}

				filter := model.ActivityLogFilter{
					Limit:  3,
					Offset: 0,
				}
				logs, err := activityDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(logs), ShouldBeLessThanOrEqualTo, 3)
			})

			Convey("ListActivityLogs with search", func() {
				searchTerm := "Created new"
				filter := model.ActivityLogFilter{
					Search: &searchTerm,
				}
				logs, err := activityDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(logs), ShouldBeGreaterThanOrEqualTo, 1)
			})
		})

		Convey("CreateActivityLog with old and new values", func() {
			userID := uuid.New()
			entityID := uuid.New()

			oldValues := map[string]interface{}{
				"status": "pending",
				"name":   "Old Name",
			}
			oldValuesJSON, _ := json.Marshal(oldValues)

			newValues := map[string]interface{}{
				"status": "active",
				"name":   "New Name",
			}
			newValuesJSON, _ := json.Marshal(newValues)

			input := model.ActivityLogInput{
				UserID:      &userID,
				Action:      "update",
				EntityType:  "customer",
				EntityID:    &entityID,
				Description: "Updated customer status",
				OldValues:   string(oldValuesJSON),
				NewValues:   string(newValuesJSON),
			}

			activity, err := activityDomain.Create(ctx, input)
			So(err, ShouldBeNil)
			So(activity.OldValues, ShouldNotBeEmpty)
			So(activity.NewValues, ShouldNotBeEmpty)

			Convey("Verify old and new values are stored correctly", func() {
				var oldData map[string]interface{}
				err := json.Unmarshal([]byte(activity.OldValues), &oldData)
				So(err, ShouldBeNil)
				So(oldData["status"], ShouldEqual, "pending")

				var newData map[string]interface{}
				err = json.Unmarshal([]byte(activity.NewValues), &newData)
				So(err, ShouldBeNil)
				So(newData["status"], ShouldEqual, "active")
			})
		})

		Convey("Different action types", func() {
			userID := uuid.New()

			actions := []string{
				"create",
				"update",
				"delete",
				"login",
				"logout",
				"isolate",
				"reactivate",
				"send_notification",
				"sync_to_mikrotik",
				"generate_invoice",
				"apply_payment",
			}

			entityTypes := []string{
				"customer",
				"invoice",
				"payment",
				"bandwidth_profile",
				"mikrotik_router",
				"user",
				"cash_transaction",
				"cash_category",
			}

			for _, action := range actions {
				for _, entityType := range entityTypes {
					entityID := uuid.New()
					input := model.ActivityLogInput{
						UserID:      &userID,
						Action:      action,
						EntityType:  entityType,
						EntityID:    &entityID,
						Description: "Test " + action + " on " + entityType,
					}

					activity, err := activityDomain.Create(ctx, input)
					So(err, ShouldBeNil)
					So(activity.Action, ShouldEqual, action)
					So(activity.EntityType, ShouldEqual, entityType)
				}
			}

			Convey("Filter by multiple actions", func() {
				filter := model.ActivityLogFilter{
					Actions: []string{"create", "update", "delete"},
				}
				logs, err := activityDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(logs), ShouldBeGreaterThanOrEqualTo, 3)
			})

			Convey("Filter by multiple entity types", func() {
				filter := model.ActivityLogFilter{
					EntityTypes: []string{"customer", "invoice", "payment"},
				}
				logs, err := activityDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(logs), ShouldBeGreaterThanOrEqualTo, 3)
			})
		})

		Convey("Activity log tracking for customer lifecycle", func() {
			userID := uuid.New()
			customerID := uuid.New()

			Convey("Track customer creation", func() {
				input := model.ActivityLogInput{
					UserID:      &userID,
					Action:      "create",
					EntityType:  "customer",
					EntityID:    &customerID,
					Description: "Created customer John Doe",
					NewValues:   `{"name": "John Doe", "status": "pending"}`,
					IPAddress:   "192.168.1.100",
					UserAgent:   "Mozilla/5.0",
				}

				activity, err := activityDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(activity.Action, ShouldEqual, "create")
			})

			Convey("Track customer activation", func() {
				input := model.ActivityLogInput{
					UserID:      &userID,
					Action:      "update",
					EntityType:  "customer",
					EntityID:    &customerID,
					Description: "Activated customer John Doe",
					OldValues:   `{"status": "pending"}`,
					NewValues:   `{"status": "active"}`,
					IPAddress:   "192.168.1.100",
					UserAgent:   "Mozilla/5.0",
				}

				activity, err := activityDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(activity.Action, ShouldEqual, "update")
			})

			Convey("Track customer isolation", func() {
				input := model.ActivityLogInput{
					UserID:      &userID,
					Action:      "isolate",
					EntityType:  "customer",
					EntityID:    &customerID,
					Description: "Isolated customer John Doe due to non-payment",
					OldValues:   `{"status": "active"}`,
					NewValues:   `{"status": "isolated"}`,
					IPAddress:   "192.168.1.100",
				}

				activity, err := activityDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(activity.Action, ShouldEqual, "isolate")
			})

			Convey("Track customer payment", func() {
				paymentID := uuid.New()
				input := model.ActivityLogInput{
					UserID:      &userID,
					Action:      "apply_payment",
					EntityType:  "payment",
					EntityID:    &paymentID,
					Description: "Applied payment of Rp 100.000",
					NewValues:   `{"amount": 100000, "customer_id": "` + customerID.String() + `"}`,
					IPAddress:   "192.168.1.100",
				}

				activity, err := activityDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(activity.Action, ShouldEqual, "apply_payment")
			})

			Convey("Get all activity for a customer", func() {
				filter := model.ActivityLogFilter{
					EntityTypes: []string{"customer", "payment"},
					EntityIDs:   []uuid.UUID{customerID},
				}
				logs, err := activityDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(logs), ShouldBeGreaterThanOrEqualTo, 1)
			})
		})

		Convey("Error handling", func() {
			Convey("GetActivityLog with invalid ID returns error", func() {
				_, err := activityDomain.Get(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("CreateActivityLog with invalid action returns error", func() {
				userID := uuid.New()
				entityID := uuid.New()

				input := model.ActivityLogInput{
					UserID:      &userID,
					Action:      "invalid_action",
					EntityType:  "customer",
					EntityID:    &entityID,
					Description: "Test",
				}

				_, err := activityDomain.Create(ctx, input)
				So(err, ShouldNotBeNil)
			})

			Convey("CreateActivityLog with invalid entity type returns error", func() {
				userID := uuid.New()
				entityID := uuid.New()

				input := model.ActivityLogInput{
					UserID:      &userID,
					Action:      "create",
					EntityType:  "invalid_entity",
					EntityID:    &entityID,
					Description: "Test",
				}

				_, err := activityDomain.Create(ctx, input)
				So(err, ShouldNotBeNil)
			})

			Convey("CreateActivityLog without description returns error", func() {
				userID := uuid.New()
				entityID := uuid.New()

				input := model.ActivityLogInput{
					UserID:     &userID,
					Action:     "create",
					EntityType: "customer",
					EntityID:   &entityID,
				}

				_, err := activityDomain.Create(ctx, input)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Activity log statistics and reporting", func() {
			userID := uuid.New()

			// Create activities for different entities
			for i := 0; i < 10; i++ {
				entityID := uuid.New()
				action := "create"
				entityType := "customer"

				if i%2 == 0 {
					action = "update"
				}
				if i%3 == 0 {
					entityType = "invoice"
				}

				input := model.ActivityLogInput{
					UserID:      &userID,
					Action:      action,
					EntityType:  entityType,
					EntityID:    &entityID,
					Description: "Activity " + string(rune('0'+i)),
				}

				_, err := activityDomain.Create(ctx, input)
				So(err, ShouldBeNil)
			}

			Convey("Count activities by user", func() {
				filter := model.ActivityLogFilter{
					UserIDs: []uuid.UUID{userID},
				}
				logs, err := activityDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(logs), ShouldEqual, 10)
			})

			Convey("Count activities by action type", func() {
				filter := model.ActivityLogFilter{
					Actions: []string{"create"},
				}
				logs, err := activityDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(logs), ShouldBeGreaterThan, 0)
			})

			Convey("Get recent activities", func() {
				now := time.Now()
				startTime := now.Add(-time.Hour)

				filter := model.ActivityLogFilter{
					CreatedAtStart: &startTime,
					Limit:          5,
				}
				logs, err := activityDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(logs), ShouldBeLessThanOrEqualTo, 5)
			})
		})

		Convey("Activity log without user", func() {
			entityID := uuid.New()

			input := model.ActivityLogInput{
				Action:      "create",
				EntityType:  "customer",
				EntityID:    &entityID,
				Description: "System created customer",
				IPAddress:   "127.0.0.1",
			}

			activity, err := activityDomain.Create(ctx, input)
			So(err, ShouldBeNil)
			So(activity.UserID, ShouldBeNil)
		})
	})
}
