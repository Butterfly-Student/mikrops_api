package staff_test

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

func TestStaff(t *testing.T) {
	Convey("Test Staff", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockStaffDatabasePort := mock_outbound_port.NewMockStaffDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().Staff().Return(mockStaffDatabasePort).AnyTimes()

		staffDomain := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)

		input := model.StaffInput{
			TenantID: "tenant-123",
			RoleID:   1,
			Email:    "staff@example.com",
			Password: "password123",
			FullName: "Test Staff",
			Phone:    "081234567890",
			IsActive: true,
		}

		output := model.Staff{
			ID: "staff-123",
			StaffInput: model.StaffInput{
				TenantID:     "tenant-123",
				RoleID:       1,
				Email:        "staff@example.com",
				PasswordHash: "$2a$10$hashedpassword",
				FullName:     "Test Staff",
				Phone:        "081234567890",
				IsActive:     true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		}

		filter := model.StaffFilter{
			IDs:       []string{"staff-123"},
			TenantIDs: []string{"tenant-123"},
			Emails:    []string{"staff@example.com"},
		}

		Convey("Create", func() {
			Convey("Success with password hashing", func() {
				mockStaffDatabasePort.EXPECT().Create(gomock.Any()).DoAndReturn(
					func(data model.StaffInput) (model.Staff, error) {
						// Verify password was hashed
						So(data.PasswordHash, ShouldNotBeEmpty)
						So(data.Password, ShouldBeEmpty)
						return output, nil
					},
				).Times(1)

				result, err := staffDomain.Staff().Create(context.Background(), input)
				So(err, ShouldBeNil)
				So(result.Email, ShouldEqual, "staff@example.com")
			})

			Convey("Database error", func() {
				mockStaffDatabasePort.EXPECT().Create(gomock.Any()).Return(model.Staff{}, errors.New("error")).Times(1)

				_, err := staffDomain.Staff().Create(context.Background(), input)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("FindByFilter", func() {
			Convey("Filter is empty", func() {
				_, err := staffDomain.Staff().FindByFilter(context.Background(), model.StaffFilter{})
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockStaffDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Staff{output}, nil).Times(1)

				results, err := staffDomain.Staff().FindByFilter(context.Background(), filter)
				So(err, ShouldBeNil)
				So(results, ShouldNotBeEmpty)
			})
		})

		Convey("FindByID", func() {
			Convey("ID is empty", func() {
				_, err := staffDomain.Staff().FindByID(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockStaffDatabasePort.EXPECT().FindByID(gomock.Any()).Return(output, nil).Times(1)

				result, err := staffDomain.Staff().FindByID(context.Background(), "staff-123")
				So(err, ShouldBeNil)
				So(result.Email, ShouldEqual, "staff@example.com")
			})
		})

		Convey("FindByEmail", func() {
			Convey("Email is empty", func() {
				_, err := staffDomain.Staff().FindByEmail(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockStaffDatabasePort.EXPECT().FindByEmail(gomock.Any()).Return(output, nil).Times(1)

				result, err := staffDomain.Staff().FindByEmail(context.Background(), "staff@example.com")
				So(err, ShouldBeNil)
				So(result.Email, ShouldEqual, "staff@example.com")
			})
		})

		Convey("Update", func() {
			Convey("ID is empty", func() {
				err := staffDomain.Staff().Update(context.Background(), "", input)
				So(err, ShouldNotBeNil)
			})

			Convey("Success with password re-hashing", func() {
				mockStaffDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
					func(id string, data model.StaffInput) error {
						// Verify password was re-hashed if provided
						if data.Password != "" {
							So(data.PasswordHash, ShouldNotBeEmpty)
							So(data.Password, ShouldBeEmpty)
						}
						return nil
					},
				).Times(1)

				err := staffDomain.Staff().Update(context.Background(), "staff-123", input)
				So(err, ShouldBeNil)
			})
		})

		Convey("Delete", func() {
			Convey("ID is empty", func() {
				err := staffDomain.Staff().Delete(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockStaffDatabasePort.EXPECT().Delete(gomock.Any()).Return(nil).Times(1)

				err := staffDomain.Staff().Delete(context.Background(), "staff-123")
				So(err, ShouldBeNil)
			})
		})
	})
}
