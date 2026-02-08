package tenant_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	mock_outbound_port "mikrops/tests/mocks/port"
)

func TestTenant(t *testing.T) {
	Convey("Test Tenant", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockTenantDatabasePort := mock_outbound_port.NewMockTenantDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().Tenant().Return(mockTenantDatabasePort).AnyTimes()

		tenantDomain := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)

		input := model.TenantInput{
			Name:  "Test Tenant",
			Slug:  "test-tenant",
			Email: "test@example.com",
		}

		output := model.Tenant{
			ID: "tenant-123",
			TenantInput: model.TenantInput{
				Name:      "Test Tenant",
				Slug:      "test-tenant",
				Email:     "test@example.com",
				MaxNas:    3,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		filter := model.TenantFilter{
			IDs:   []string{"tenant-123"},
			Slugs: []string{"test-tenant"},
		}

		Convey("Create", func() {
			Convey("Slug already exists", func() {
				mockTenantDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Tenant{output}, nil).Times(1)

				_, err := tenantDomain.Tenant().Create(context.Background(), input)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "slug already exists")
			})

			Convey("Database error on slug check", func() {
				mockTenantDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return(nil, errors.New("db error")).Times(1)

				_, err := tenantDomain.Tenant().Create(context.Background(), input)
				So(err, ShouldNotBeNil)
			})

			Convey("Database error on create", func() {
				mockTenantDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Tenant{}, nil).Times(1)
				mockTenantDatabasePort.EXPECT().Create(gomock.Any()).Return(model.Tenant{}, errors.New("create error")).Times(1)

				_, err := tenantDomain.Tenant().Create(context.Background(), input)
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Tenant{}, nil).Times(1)
				mockTenantDatabasePort.EXPECT().Create(gomock.Any()).Return(output, nil).Times(1)

				result, err := tenantDomain.Tenant().Create(context.Background(), input)
				So(err, ShouldBeNil)
				So(result.Name, ShouldEqual, "Test Tenant")
				So(result.Slug, ShouldEqual, "test-tenant")
			})
		})

		Convey("FindByFilter", func() {
			Convey("Filter is empty", func() {
				_, err := tenantDomain.Tenant().FindByFilter(context.Background(), model.TenantFilter{})
				So(err, ShouldNotBeNil)
			})

			Convey("Database error", func() {
				mockTenantDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return(nil, errors.New("error")).Times(1)

				_, err := tenantDomain.Tenant().FindByFilter(context.Background(), filter)
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Tenant{output}, nil).Times(1)

				results, err := tenantDomain.Tenant().FindByFilter(context.Background(), filter)
				So(err, ShouldBeNil)
				So(results, ShouldNotBeEmpty)
				So(results[0].Name, ShouldEqual, "Test Tenant")
			})
		})

		Convey("FindByID", func() {
			Convey("ID is empty", func() {
				_, err := tenantDomain.Tenant().FindByID(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Database error", func() {
				mockTenantDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Tenant{}, errors.New("error")).Times(1)

				_, err := tenantDomain.Tenant().FindByID(context.Background(), "tenant-123")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().FindByID(gomock.Any()).Return(output, nil).Times(1)

				result, err := tenantDomain.Tenant().FindByID(context.Background(), "tenant-123")
				So(err, ShouldBeNil)
				So(result.Name, ShouldEqual, "Test Tenant")
			})
		})

		Convey("Update", func() {
			Convey("ID is empty", func() {
				err := tenantDomain.Tenant().Update(context.Background(), "", input)
				So(err, ShouldNotBeNil)
			})

			Convey("Slug already exists for different tenant", func() {
				differentTenant := output
				differentTenant.ID = "different-id"
				mockTenantDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Tenant{differentTenant}, nil).Times(1)

				err := tenantDomain.Tenant().Update(context.Background(), "tenant-123", input)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "slug already exists")
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Tenant{}, nil).Times(1)
				mockTenantDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				err := tenantDomain.Tenant().Update(context.Background(), "tenant-123", input)
				So(err, ShouldBeNil)
			})
		})

		Convey("Delete", func() {
			Convey("ID is empty", func() {
				err := tenantDomain.Tenant().Delete(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Database error", func() {
				mockTenantDatabasePort.EXPECT().Delete(gomock.Any()).Return(errors.New("error")).Times(1)

				err := tenantDomain.Tenant().Delete(context.Background(), "tenant-123")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().Delete(gomock.Any()).Return(nil).Times(1)

				err := tenantDomain.Tenant().Delete(context.Background(), "tenant-123")
				So(err, ShouldBeNil)
			})
		})
	})
}
