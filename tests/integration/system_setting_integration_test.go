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

func TestSystemSettingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to setup postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	err = pgContainer.DB.AutoMigrate(&model.SystemSetting{})
	if err != nil {
		t.Fatalf("Failed to migrate system setting table: %v", err)
	}

	dbAdapter := postgres_outbound_adapter.NewAdapter(pgContainer.DB)
	systemSettingDomain := domain.NewSystemSettingDomain(dbAdapter)

	Convey("Test SystemSetting Integration with PostgreSQL", t, func() {
		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.SystemSetting{})

		Convey("CreateSystemSetting creates setting successfully", func() {
			key := "test.setting." + time.Now().Format("20060102150405")
			value := "test value"
			valueType := "string"
			category := "test"
			description := "Test setting description"

			input := model.SystemSettingInput{
				Key:         &key,
				Value:       &value,
				ValueType:   &valueType,
				Category:    &category,
				Description: &description,
			}

			setting, err := systemSettingDomain.Create(ctx, input)
			So(err, ShouldBeNil)
			So(setting, ShouldNotBeNil)
			So(setting.Key, ShouldEqual, key)
			So(setting.Value, ShouldNotBeNil)
			So(*setting.Value, ShouldEqual, "test value")
			So(setting.ValueType, ShouldEqual, "string")
			So(setting.Category, ShouldNotBeNil)
			So(*setting.Category, ShouldEqual, "test")
			So(setting.Description, ShouldNotBeNil)
			So(*setting.Description, ShouldEqual, "Test setting description")
			So(setting.IsPublic, ShouldNotBeNil)
			So(*setting.IsPublic, ShouldEqual, false)

			Convey("GetSystemSetting retrieves created setting", func() {
				found, err := systemSettingDomain.Get(ctx, setting.ID.String())
				So(err, ShouldBeNil)
				So(found.ID, ShouldEqual, setting.ID)
				So(found.Key, ShouldEqual, key)
			})

			Convey("FindByKey retrieves setting by key", func() {
				found, err := systemSettingDomain.FindByKey(key)
				So(err, ShouldBeNil)
				So(found.Key, ShouldEqual, key)
				So(found.Value, ShouldNotBeNil)
				So(*found.Value, ShouldEqual, "test value")
			})

			Convey("ListSystemSettings returns settings", func() {
				filter := model.SystemSettingFilter{}
				settings, err := systemSettingDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(settings), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("ListSystemSettings with category filter", func() {
				filter := model.SystemSettingFilter{
					Category: &category,
				}
				settings, err := systemSettingDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(settings), ShouldBeGreaterThanOrEqualTo, 1)
				for _, s := range settings {
					So(s.Category, ShouldNotBeNil)
					So(*s.Category, ShouldEqual, "test")
				}
			})

			Convey("UpdateSystemSetting updates setting", func() {
				newValue := "updated value"
				newDescription := "Updated description"

				input := model.SystemSettingInput{
					Value:       &newValue,
					Description: &newDescription,
				}

				updated, err := systemSettingDomain.Update(ctx, setting.ID.String(), input)
				So(err, ShouldBeNil)
				So(updated.Value, ShouldNotBeNil)
				So(*updated.Value, ShouldEqual, "updated value")
				So(updated.Description, ShouldNotBeNil)
				So(*updated.Description, ShouldEqual, "Updated description")
			})

			Convey("UpsertSystemSetting creates or updates", func() {
				// Update existing setting
				upsertValue := "upserted value"
				input := model.SystemSettingInput{
					Key:   &key,
					Value: &upsertValue,
				}

				setting, err := systemSettingDomain.Upsert(ctx, input)
				So(err, ShouldBeNil)
				So(setting.Value, ShouldNotBeNil)
				So(*setting.Value, ShouldEqual, "upserted value")

				// Create new setting
				newKey := "new.upsert.setting." + time.Now().Format("20060102150405")
				newValue := "new upsert value"
				input = model.SystemSettingInput{
					Key:   &newKey,
					Value: &newValue,
				}

				setting, err = systemSettingDomain.Upsert(ctx, input)
				So(err, ShouldBeNil)
				So(setting.Key, ShouldEqual, newKey)
				So(setting.Value, ShouldNotBeNil)
				So(*setting.Value, ShouldEqual, "new upsert value")
			})

			Convey("DeleteSystemSetting removes setting", func() {
				err := systemSettingDomain.Delete(ctx, setting.ID.String())
				So(err, ShouldBeNil)

				// Verify it's deleted
				_, err = systemSettingDomain.Get(ctx, setting.ID.String())
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Different value types", func() {
			Convey("String value type", func() {
				key := "setting.string." + time.Now().Format("20060102150405")
				value := "string value"
				valueType := "string"

				input := model.SystemSettingInput{
					Key:       &key,
					Value:     &value,
					ValueType: &valueType,
				}

				setting, err := systemSettingDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(setting.ValueType, ShouldEqual, "string")

				// Test GetString
				result := setting.GetString("default")
				So(result, ShouldEqual, "string value")
			})

			Convey("Number value type", func() {
				key := "setting.number." + time.Now().Format("20060102150405")
				value := "42"
				valueType := "number"

				input := model.SystemSettingInput{
					Key:       &key,
					Value:     &value,
					ValueType: &valueType,
				}

				setting, err := systemSettingDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(setting.ValueType, ShouldEqual, "number")

				// Test GetInt
				result := setting.GetInt(0)
				So(result, ShouldEqual, 42)
			})

			Convey("Boolean value type", func() {
				key := "setting.boolean." + time.Now().Format("20060102150405")
				value := "true"
				valueType := "boolean"

				input := model.SystemSettingInput{
					Key:       &key,
					Value:     &value,
					ValueType: &valueType,
				}

				setting, err := systemSettingDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(setting.ValueType, ShouldEqual, "boolean")

				// Test GetBool
				result := setting.GetBool(false)
				So(result, ShouldEqual, true)
			})

			Convey("Float value type", func() {
				key := "setting.float." + time.Now().Format("20060102150405")
				value := "3.14"
				valueType := "number"

				input := model.SystemSettingInput{
					Key:       &key,
					Value:     &value,
					ValueType: &valueType,
				}

				setting, err := systemSettingDomain.Create(ctx, input)
				So(err, ShouldBeNil)

				// Test GetFloat
				result := setting.GetFloat(0.0)
				So(result, ShouldEqual, 3.14)
			})
		})

		Convey("Get value with defaults", func() {
			Convey("GetString returns default when value is nil", func() {
				key := "setting.nilstring." + time.Now().Format("20060102150405")
				valueType := "string"

				input := model.SystemSettingInput{
					Key:       &key,
					ValueType: &valueType,
				}

				setting, err := systemSettingDomain.Create(ctx, input)
				So(err, ShouldBeNil)

				result := setting.GetString("default value")
				So(result, ShouldEqual, "default value")
			})

			Convey("GetInt returns default when value is nil", func() {
				key := "setting.nilint." + time.Now().Format("20060102150405")
				valueType := "number"

				input := model.SystemSettingInput{
					Key:       &key,
					ValueType: &valueType,
				}

				setting, err := systemSettingDomain.Create(ctx, input)
				So(err, ShouldBeNil)

				result := setting.GetInt(100)
				So(result, ShouldEqual, 100)
			})

			Convey("GetBool returns default when value is nil", func() {
				key := "setting.nilbool." + time.Now().Format("20060102150405")
				valueType := "boolean"

				input := model.SystemSettingInput{
					Key:       &key,
					ValueType: &valueType,
				}

				setting, err := systemSettingDomain.Create(ctx, input)
				So(err, ShouldBeNil)

				result := setting.GetBool(true)
				So(result, ShouldEqual, true)
			})
		})

		Convey("IsPublic flag", func() {
			key := "setting.public." + time.Now().Format("20060102150405")
			value := "public value"
			valueType := "string"
			isPublic := true

			input := model.SystemSettingInput{
				Key:       &key,
				Value:     &value,
				ValueType: &valueType,
				IsPublic:  &isPublic,
			}

			setting, err := systemSettingDomain.Create(ctx, input)
			So(err, ShouldBeNil)
			So(setting.IsPublic, ShouldNotBeNil)
			So(*setting.IsPublic, ShouldEqual, true)

			Convey("Filter by public settings", func() {
				filter := model.SystemSettingFilter{
					IsPublic: &isPublic,
				}
				settings, err := systemSettingDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(settings), ShouldBeGreaterThanOrEqualTo, 1)
				for _, s := range settings {
					So(s.IsPublic, ShouldNotBeNil)
					So(*s.IsPublic, ShouldEqual, true)
				}
			})
		})

		Convey("Multiple keys filter", func() {
			// Create multiple settings
			keys := []string{}
			for i := 0; i < 3; i++ {
				key := "setting.multi." + time.Now().Format("20060102150405") + string(rune('0'+i))
				value := "value " + string(rune('0'+i))
				valueType := "string"

				input := model.SystemSettingInput{
					Key:       &key,
					Value:     &value,
					ValueType: &valueType,
				}

				_, err := systemSettingDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				keys = append(keys, key)
			}

			Convey("Filter by multiple keys", func() {
				filter := model.SystemSettingFilter{
					Keys: keys,
				}
				settings, err := systemSettingDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(settings), ShouldEqual, 3)
			})
		})

		Convey("Error handling", func() {
			Convey("GetSystemSetting with invalid ID returns error", func() {
				_, err := systemSettingDomain.Get(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("UpdateSystemSetting with invalid ID returns error", func() {
				value := "test"
				input := model.SystemSettingInput{
					Value: &value,
				}
				_, err := systemSettingDomain.Update(ctx, uuid.New().String(), input)
				So(err, ShouldNotBeNil)
			})

			Convey("DeleteSystemSetting with invalid ID returns error", func() {
				err := systemSettingDomain.Delete(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("FindByKey with non-existent key returns error", func() {
				_, err := systemSettingDomain.FindByKey("non.existent.key")
				So(err, ShouldNotBeNil)
			})

			Convey("Create duplicate key returns error", func() {
				key := "duplicate.key." + time.Now().Format("20060102150405")
				value := "first"
				valueType := "string"

				input := model.SystemSettingInput{
					Key:       &key,
					Value:     &value,
					ValueType: &valueType,
				}

				_, err := systemSettingDomain.Create(ctx, input)
				So(err, ShouldBeNil)

				// Try to create duplicate
				input2 := model.SystemSettingInput{
					Key:       &key,
					Value:     &value,
					ValueType: &valueType,
				}

				_, err = systemSettingDomain.Create(ctx, input2)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Common system settings", func() {
			settings := []struct {
				key        string
				value      string
				valueType  string
				category   string
				description string
			}{
				{"company.name", "PT Internet Provider", "string", "company", "Company name"},
				{"company.address", "Jl. Raya No. 123", "string", "company", "Company address"},
				{"company.phone", "021-12345678", "string", "company", "Company phone"},
				{"company.email", "info@isp.com", "string", "company", "Company email"},
				{"invoice.due_days", "7", "number", "invoice", "Default invoice due days"},
				{"invoice.late_fee_enabled", "true", "boolean", "invoice", "Enable late fees"},
				{"invoice.late_fee_amount", "5000", "number", "invoice", "Late fee amount per day"},
				{"payment.auto_confirm", "false", "boolean", "payment", "Auto confirm payments"},
			}

			for _, s := range settings {
				key := s.key + "." + time.Now().Format("20060102150405")
				input := model.SystemSettingInput{
					Key:         &key,
					Value:       &s.value,
					ValueType:   &s.valueType,
					Category:    &s.category,
					Description: &s.description,
				}

				setting, err := systemSettingDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(setting.Key, ShouldEqual, key)
			}

			Convey("Retrieve company settings", func() {
				category := "company"
				filter := model.SystemSettingFilter{
					Category: &category,
				}
				settings, err := systemSettingDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(settings), ShouldBeGreaterThanOrEqualTo, 4)
			})

			Convey("Retrieve invoice settings", func() {
				category := "invoice"
				filter := model.SystemSettingFilter{
					Category: &category,
				}
				settings, err := systemSettingDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(settings), ShouldBeGreaterThanOrEqualTo, 3)
			})
		})
	})
}
