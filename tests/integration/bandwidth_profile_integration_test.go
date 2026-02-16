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

	mikrotik_outbound_adapter "go-template/internal/adapter/outbound/mikrotik"
	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	"go-template/internal/domain/bandwidth_profile"
	"go-template/internal/model"
	"go-template/tests/helpers"
)

func TestBandwidthProfileIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to setup postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	err = pgContainer.DB.AutoMigrate(&model.BandwidthProfile{})
	if err != nil {
		t.Fatalf("Failed to migrate bandwidth profile table: %v", err)
	}

	dbAdapter := postgres_outbound_adapter.NewAdapter(pgContainer.DB)
	mikrotikAdapter := mikrotik_outbound_adapter.NewMikrotikClientAdapter()
	bandwidthProfileDomain := bandwidth_profile.NewBandwidthProfileDomain(dbAdapter, mikrotikAdapter)

	Convey("Test BandwidthProfile Integration with PostgreSQL", t, func() {
		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.BandwidthProfile{})

		Convey("CreateBandwidthProfile creates profile successfully", func() {
			profileCode := "PROFILE-" + time.Now().Format("20060102150405")
			name := "Test Profile 10Mbps"
			description := "10Mbps download, 5Mbps upload"
			category := "pppoe"
			pppProfileName := "pppoe-10mbps"
			downloadSpeed := int64(10240)
			uploadSpeed := int64(5120)
			priceMonthly := 100000.0
			priceInstallation := 50000.0
			taxRate := 0.11
			isActive := true
			isVisible := true

			input := model.BandwidthProfileInput{
				ProfileCode:       &profileCode,
				Name:              name,
				Description:       &description,
				Category:          category,
				PppProfileName:    pppProfileName,
				DownloadSpeed:     downloadSpeed,
				UploadSpeed:       uploadSpeed,
				PriceMonthly:      priceMonthly,
				PriceInstallation: priceInstallation,
				TaxRate:           taxRate,
				IsActive:          &isActive,
				IsVisible:         &isVisible,
			}

			profile, err := bandwidthProfileDomain.CreateProfile(ctx, input)
			So(err, ShouldBeNil)
			So(profile, ShouldNotBeNil)
			So(profile.ProfileCode, ShouldEqual, profileCode)
			So(profile.Name, ShouldEqual, "Test Profile 10Mbps")
			So(profile.Description, ShouldNotBeNil)
			So(*profile.Description, ShouldEqual, "10Mbps download, 5Mbps upload")
			So(profile.Category, ShouldEqual, "pppoe")
			So(profile.PppProfileName, ShouldEqual, "pppoe-10mbps")
			So(profile.DownloadSpeed, ShouldEqual, int64(10240))
			So(profile.UploadSpeed, ShouldEqual, int64(5120))
			So(profile.PriceMonthly, ShouldEqual, 100000.0)
			So(profile.PriceInstallation, ShouldEqual, 50000.0)
			So(profile.TaxRate, ShouldEqual, 0.11)
			So(profile.IsActive, ShouldNotBeNil)
			So(*profile.IsActive, ShouldEqual, true)
			So(profile.IsVisible, ShouldNotBeNil)
			So(*profile.IsVisible, ShouldEqual, true)
			So(profile.Priority, ShouldNotBeNil)
			So(*profile.Priority, ShouldEqual, 8)

			Convey("GetBandwidthProfile retrieves created profile", func() {
				found, err := bandwidthProfileDomain.GetProfile(ctx, profile.ID.String())
				So(err, ShouldBeNil)
				So(found.ID, ShouldEqual, profile.ID)
				So(found.ProfileCode, ShouldEqual, profileCode)
			})

			Convey("ListBandwidthProfiles returns profiles", func() {
				filter := model.BandwidthProfileFilter{}
				profiles, err := bandwidthProfileDomain.ListProfiles(ctx, filter)
				So(err, ShouldBeNil)
				So(len(profiles), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("ListBandwidthProfiles with category filter", func() {
				filter := model.BandwidthProfileFilter{
					Categories: []string{"pppoe"},
				}
				profiles, err := bandwidthProfileDomain.ListProfiles(ctx, filter)
				So(err, ShouldBeNil)
				So(len(profiles), ShouldBeGreaterThanOrEqualTo, 1)
				for _, p := range profiles {
					So(p.Category, ShouldEqual, "pppoe")
				}
			})

			Convey("ListBandwidthProfiles with active filter", func() {
				filter := model.BandwidthProfileFilter{
					IsActive: &isActive,
				}
				profiles, err := bandwidthProfileDomain.ListProfiles(ctx, filter)
				So(err, ShouldBeNil)
				So(len(profiles), ShouldBeGreaterThanOrEqualTo, 1)
				for _, p := range profiles {
					So(p.IsActive, ShouldNotBeNil)
					So(*p.IsActive, ShouldEqual, true)
				}
			})

			Convey("ListBandwidthProfiles with visible filter", func() {
				filter := model.BandwidthProfileFilter{
					IsVisible: &isVisible,
				}
				profiles, err := bandwidthProfileDomain.ListProfiles(ctx, filter)
				So(err, ShouldBeNil)
				So(len(profiles), ShouldBeGreaterThanOrEqualTo, 1)
				for _, p := range profiles {
					So(p.IsVisible, ShouldNotBeNil)
					So(*p.IsVisible, ShouldEqual, true)
				}
			})

			Convey("UpdateBandwidthProfile updates profile", func() {
				newName := "Updated Profile 10Mbps"
				newDescription := "Updated description"
				newPrice := 150000.0

				input := model.BandwidthProfileInput{
					Name:        newName,
					Description: &newDescription,
					PriceMonthly: newPrice,
				}

				updated, err := bandwidthProfileDomain.UpdateProfile(ctx, profile.ID.String(), input)
				So(err, ShouldBeNil)
				So(updated.Name, ShouldEqual, "Updated Profile 10Mbps")
				So(updated.Description, ShouldNotBeNil)
				So(*updated.Description, ShouldEqual, "Updated description")
				So(updated.PriceMonthly, ShouldEqual, 150000.0)
			})

			Convey("DeleteBandwidthProfile soft deletes profile", func() {
				err := bandwidthProfileDomain.DeleteProfile(ctx, profile.ID.String())
				So(err, ShouldBeNil)

				// Verify soft delete
				_, err = bandwidthProfileDomain.GetProfile(ctx, profile.ID.String())
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Different profile categories", func() {
			categories := []string{"pppoe", "ip-static", "hotspot", "isolated"}

			for _, category := range categories {
				profileCode := "PROF-" + category + "-" + time.Now().Format("20060102150405")
				name := category + " profile"
				pppProfileName := "profile-" + category
				downloadSpeed := int64(10240)
				uploadSpeed := int64(5120)
				priceMonthly := 100000.0

				input := model.BandwidthProfileInput{
					ProfileCode:    &profileCode,
					Name:           name,
					Category:       category,
					PppProfileName: pppProfileName,
					DownloadSpeed:  downloadSpeed,
					UploadSpeed:    uploadSpeed,
					PriceMonthly:   priceMonthly,
				}

				profile, err := bandwidthProfileDomain.CreateProfile(ctx, input)
				So(err, ShouldBeNil)
				So(profile.Category, ShouldEqual, category)
			}

			Convey("Filter by multiple categories", func() {
				filter := model.BandwidthProfileFilter{
					Categories: []string{"pppoe", "ip-static"},
				}
				profiles, err := bandwidthProfileDomain.ListProfiles(ctx, filter)
				So(err, ShouldBeNil)
				So(len(profiles), ShouldBeGreaterThanOrEqualTo, 2)
			})
		})

		Convey("Profile with burst settings", func() {
			profileCode := "BURST-" + time.Now().Format("20060102150405")
			name := "Burst Profile"
			category := "pppoe"
			pppProfileName := "burst-profile"
			downloadSpeed := int64(10240)
			uploadSpeed := int64(5120)
			burstDownload := int64(20480)
			burstUpload := int64(10240)
			burstThreshold := 80
			burstTime := 30
			priceMonthly := 150000.0

			input := model.BandwidthProfileInput{
				ProfileCode:    &profileCode,
				Name:           name,
				Category:       category,
				PppProfileName: pppProfileName,
				DownloadSpeed:  downloadSpeed,
				UploadSpeed:    uploadSpeed,
				BurstDownload:  &burstDownload,
				BurstUpload:    &burstUpload,
				BurstThreshold: &burstThreshold,
				BurstTime:      &burstTime,
				PriceMonthly:   priceMonthly,
			}

			profile, err := bandwidthProfileDomain.CreateProfile(ctx, input)
			So(err, ShouldBeNil)
			So(profile.BurstDownload, ShouldNotBeNil)
			So(*profile.BurstDownload, ShouldEqual, int64(20480))
			So(profile.BurstUpload, ShouldNotBeNil)
			So(*profile.BurstUpload, ShouldEqual, int64(10240))
			So(profile.BurstThreshold, ShouldNotBeNil)
			So(*profile.BurstThreshold, ShouldEqual, 80)
			So(profile.BurstTime, ShouldNotBeNil)
			So(*profile.BurstTime, ShouldEqual, 30)
		})

		Convey("Profile with priority and queue settings", func() {
			profileCode := "PRIORITY-" + time.Now().Format("20060102150405")
			name := "Priority Profile"
			category := "pppoe"
			pppProfileName := "priority-profile"
			downloadSpeed := int64(10240)
			uploadSpeed := int64(5120)
			priority := 5
			queueType := "pcq-upload"
			sharedUsers := 5
			queueName := "queue-priority"
			priceMonthly := 200000.0

			input := model.BandwidthProfileInput{
				ProfileCode:    &profileCode,
				Name:           name,
				Category:       category,
				PppProfileName: pppProfileName,
				DownloadSpeed:  downloadSpeed,
				UploadSpeed:    uploadSpeed,
				Priority:       &priority,
				QueueType:      &queueType,
				SharedUsers:    &sharedUsers,
				QueueName:      &queueName,
				PriceMonthly:   priceMonthly,
			}

			profile, err := bandwidthProfileDomain.CreateProfile(ctx, input)
			So(err, ShouldBeNil)
			So(profile.Priority, ShouldNotBeNil)
			So(*profile.Priority, ShouldEqual, 5)
			So(profile.QueueType, ShouldNotBeNil)
			So(*profile.QueueType, ShouldEqual, "pcq-upload")
			So(profile.SharedUsers, ShouldNotBeNil)
			So(*profile.SharedUsers, ShouldEqual, 5)
			So(profile.QueueName, ShouldNotBeNil)
			So(*profile.QueueName, ShouldEqual, "queue-priority")
		})

		Convey("CreateBandwidthProfile with duplicate code returns error", func() {
			profileCode := "DUP-" + time.Now().Format("20060102150405")
			name := "Duplicate Profile"
			category := "pppoe"
			pppProfileName := "dup-profile"
			downloadSpeed := int64(10240)
			uploadSpeed := int64(5120)
			priceMonthly := 100000.0

			input := model.BandwidthProfileInput{
				ProfileCode:    &profileCode,
				Name:           name,
				Category:       category,
				PppProfileName: pppProfileName,
				DownloadSpeed:  downloadSpeed,
				UploadSpeed:    uploadSpeed,
				PriceMonthly:   priceMonthly,
			}

			_, err := bandwidthProfileDomain.CreateProfile(ctx, input)
			So(err, ShouldBeNil)

			// Try to create duplicate
			input2 := model.BandwidthProfileInput{
				ProfileCode:    &profileCode,
				Name:           "Duplicate Profile 2",
				Category:       category,
				PppProfileName: "dup-profile-2",
				DownloadSpeed:  downloadSpeed,
				UploadSpeed:    uploadSpeed,
				PriceMonthly:   priceMonthly,
			}

			_, err = bandwidthProfileDomain.CreateProfile(ctx, input2)
			So(err, ShouldNotBeNil)
		})

		Convey("GetBandwidthProfile with invalid ID returns error", func() {
			_, err := bandwidthProfileDomain.GetProfile(ctx, uuid.New().String())
			So(err, ShouldNotBeNil)
		})

		Convey("UpdateBandwidthProfile with invalid ID returns error", func() {
			input := model.BandwidthProfileInput{
				Name: "Updated Name",
			}
			_, err := bandwidthProfileDomain.UpdateProfile(ctx, uuid.New().String(), input)
			So(err, ShouldNotBeNil)
		})

		Convey("DeleteBandwidthProfile with invalid ID returns error", func() {
			err := bandwidthProfileDomain.DeleteProfile(ctx, uuid.New().String())
			So(err, ShouldNotBeNil)
		})

		Convey("Sort order filtering", func() {
			// Create multiple profiles with different sort orders
			for i := 0; i < 3; i++ {
				profileCode := "SORT-" + time.Now().Format("20060102150405") + string(rune('0'+i))
				name := "Sort Profile " + string(rune('1'+i))
				category := "pppoe"
				pppProfileName := "sort-profile-" + string(rune('0'+i))
				downloadSpeed := int64(10240)
				uploadSpeed := int64(5120)
				priceMonthly := 100000.0
				sortOrder := i + 1

				input := model.BandwidthProfileInput{
					ProfileCode:    &profileCode,
					Name:           name,
					Category:       category,
					PppProfileName: pppProfileName,
					DownloadSpeed:  downloadSpeed,
					UploadSpeed:    uploadSpeed,
					PriceMonthly:   priceMonthly,
					SortOrder:      &sortOrder,
				}

				_, err := bandwidthProfileDomain.CreateProfile(ctx, input)
				So(err, ShouldBeNil)
			}

			Convey("List all profiles", func() {
				filter := model.BandwidthProfileFilter{}
				profiles, err := bandwidthProfileDomain.ListProfiles(ctx, filter)
				So(err, ShouldBeNil)
				So(len(profiles), ShouldBeGreaterThanOrEqualTo, 3)
			})
		})

		Convey("Active and inactive profiles", func() {
			// Create active profile
			activeCode := "ACTIVE-" + time.Now().Format("20060102150405")
			active := true
			activeInput := model.BandwidthProfileInput{
				ProfileCode:    &activeCode,
				Name:           "Active Profile",
				Category:       "pppoe",
				PppProfileName: "active-profile",
				DownloadSpeed:  int64(10240),
				UploadSpeed:    int64(5120),
				PriceMonthly:   100000.0,
				IsActive:       &active,
			}

			activeProfile, err := bandwidthProfileDomain.CreateProfile(ctx, activeInput)
			So(err, ShouldBeNil)

			// Create inactive profile
			inactiveCode := "INACTIVE-" + time.Now().Format("20060102150405")
			inactive := false
			inactiveInput := model.BandwidthProfileInput{
				ProfileCode:    &inactiveCode,
				Name:           "Inactive Profile",
				Category:       "pppoe",
				PppProfileName: "inactive-profile",
				DownloadSpeed:  int64(10240),
				UploadSpeed:    int64(5120),
				PriceMonthly:   100000.0,
				IsActive:       &inactive,
			}

			inactiveProfile, err := bandwidthProfileDomain.CreateProfile(ctx, inactiveInput)
			So(err, ShouldBeNil)

			Convey("Filter active profiles", func() {
				isActive := true
				filter := model.BandwidthProfileFilter{
					IsActive: &isActive,
				}
				profiles, err := bandwidthProfileDomain.ListProfiles(ctx, filter)
				So(err, ShouldBeNil)
				So(len(profiles), ShouldBeGreaterThanOrEqualTo, 1)

				// Verify our active profile is in the list
				found := false
				for _, p := range profiles {
					if p.ID == activeProfile.ID {
						found = true
						break
					}
				}
				So(found, ShouldBeTrue)
			})

			Convey("Filter inactive profiles", func() {
				isActive := false
				filter := model.BandwidthProfileFilter{
					IsActive: &isActive,
				}
				profiles, err := bandwidthProfileDomain.ListProfiles(ctx, filter)
				So(err, ShouldBeNil)

				// Verify our inactive profile is in the list
				found := false
				for _, p := range profiles {
					if p.ID == inactiveProfile.ID {
						found = true
						break
					}
				}
				So(found, ShouldBeTrue)
			})
		})
	})
}
