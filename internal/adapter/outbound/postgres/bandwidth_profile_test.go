package postgres_outbound_adapter_test

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	"go-template/internal/model"
	"go-template/tests/helpers"
)

func TestBandwidthProfileAdapter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer pgContainer.Terminate(ctx)

	err = pgContainer.DB.AutoMigrate(&model.BandwidthProfile{})
	if err != nil {
		t.Fatal(err)
	}

	adapter := postgres_outbound_adapter.NewBandwidthProfileAdapter(pgContainer.DB)

	Convey("Test Postgres Bandwidth Profile Adapter (Integration)", t, func() {
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.BandwidthProfile{})

		isActive := true
		isVisible := true
		profile := &model.BandwidthProfile{
			ProfileCode:       "PLAN-" + time.Now().Format("20060102150405"),
			Name:              "Test Plan 10Mbps",
			Description:       strPtr("Test bandwidth plan"),
			ServiceType:       model.ServiceTypePPPoE,
			Category:          model.ProfileCategoryResidential,
			PppProfileName:    "ppp-10m",
			DownloadSpeed:     10000,
			UploadSpeed:       10000,
			PriceMonthly:      200000,
			PriceInstallation: 500000,
			IsActive:          &isActive,
			IsVisible:         &isVisible,
		}

		Convey("Create", func() {
			Convey("Insert new profile", func() {
				err := adapter.Create(ctx, profile)
				So(err, ShouldBeNil)
				So(profile.ID.String(), ShouldNotBeEmpty)

				var count int64
				pgContainer.DB.Model(&model.BandwidthProfile{}).Count(&count)
				So(count, ShouldEqual, 1)
			})
		})

		Convey("FindByID", func() {
			adapter.Create(ctx, profile)

			Convey("Find existing profile", func() {
				found, err := adapter.FindByID(ctx, profile.ID.String())
				So(err, ShouldBeNil)
				So(found.ProfileCode, ShouldEqual, profile.ProfileCode)
				So(found.Name, ShouldEqual, profile.Name)
			})

			Convey("Find non-existent profile", func() {
				_, err := adapter.FindByID(ctx, "550e8400-e29b-41d4-a716-446655449999")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("FindByCode", func() {
			adapter.Create(ctx, profile)

			Convey("Find by profile code", func() {
				found, err := adapter.FindByCode(ctx, profile.ProfileCode)
				So(err, ShouldBeNil)
				So(found.ID, ShouldEqual, profile.ID)
			})
		})

		Convey("FindAll", func() {
			adapter.Create(ctx, profile)

			Convey("Find all without filter", func() {
				profiles, err := adapter.FindAll(ctx, nil)
				So(err, ShouldBeNil)
				So(len(profiles), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("Find with category filter", func() {
				category := model.ProfileCategoryResidential
				filter := &model.BandwidthProfileFilter{Category: &category}
				profiles, err := adapter.FindAll(ctx, filter)
				So(err, ShouldBeNil)
				So(len(profiles), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("Find with price range filter", func() {
				minPrice := float64(100000)
				maxPrice := float64(300000)
				filter := &model.BandwidthProfileFilter{
					MinPrice: &minPrice,
					MaxPrice: &maxPrice,
				}
				profiles, err := adapter.FindAll(ctx, filter)
				So(err, ShouldBeNil)
				So(len(profiles), ShouldBeGreaterThanOrEqualTo, 1)
			})
		})

		Convey("Update", func() {
			adapter.Create(ctx, profile)

			Convey("Update profile data", func() {
				profile.Name = "Updated Plan Name"
				err := adapter.Update(ctx, profile)
				So(err, ShouldBeNil)

				found, _ := adapter.FindByID(ctx, profile.ID.String())
				So(found.Name, ShouldEqual, "Updated Plan Name")
			})
		})

		Convey("Delete", func() {
			adapter.Create(ctx, profile)

			Convey("Soft delete profile", func() {
				err := adapter.Delete(ctx, profile.ID.String())
				So(err, ShouldBeNil)

				_, err = adapter.FindByID(ctx, profile.ID.String())
				So(err, ShouldNotBeNil)
			})
		})
	})
}
