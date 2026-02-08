package internet_package_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	mock_outbound_port "mikrops/tests/mocks/port"
)

func TestInternetPackage(t *testing.T) {
	Convey("Test Internet Package", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockPackageDatabasePort := mock_outbound_port.NewMockInternetPackageDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().InternetPackage().Return(mockPackageDatabasePort).AnyTimes()

		packageDomain := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)

		input := model.InternetPackageInput{
			TenantID:     "tenant-123",
			Name:         "Package 10Mbps",
			Type:         "pppoe",
			UploadRate:   "10M",
			DownloadRate: "10M",
			Price:        100000,
			IsActive:     true,
		}

		output := model.InternetPackage{
			ID:                   "package-123",
			InternetPackageInput: input,
		}

		filter := model.InternetPackageFilter{IDs: []string{"package-123"}}

		Convey("Create", func() {
			Convey("Success", func() {
				mockPackageDatabasePort.EXPECT().Create(gomock.Any()).Return(output, nil).Times(1)

				result, err := packageDomain.InternetPackage().Create(context.Background(), input)
				So(err, ShouldBeNil)
				So(result.Name, ShouldEqual, "Package 10Mbps")
			})

			Convey("Database error", func() {
				mockPackageDatabasePort.EXPECT().Create(gomock.Any()).Return(model.InternetPackage{}, errors.New("error")).Times(1)

				_, err := packageDomain.InternetPackage().Create(context.Background(), input)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("FindByFilter", func() {
			Convey("Filter is empty", func() {
				_, err := packageDomain.InternetPackage().FindByFilter(context.Background(), model.InternetPackageFilter{})
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockPackageDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.InternetPackage{output}, nil).Times(1)

				results, err := packageDomain.InternetPackage().FindByFilter(context.Background(), filter)
				So(err, ShouldBeNil)
				So(results, ShouldNotBeEmpty)
			})
		})

		Convey("FindByID", func() {
			Convey("ID is empty", func() {
				_, err := packageDomain.InternetPackage().FindByID(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockPackageDatabasePort.EXPECT().FindByID(gomock.Any()).Return(output, nil).Times(1)

				result, err := packageDomain.InternetPackage().FindByID(context.Background(), "package-123")
				So(err, ShouldBeNil)
				So(result.Name, ShouldEqual, "Package 10Mbps")
			})
		})

		Convey("Update", func() {
			Convey("ID is empty", func() {
				err := packageDomain.InternetPackage().Update(context.Background(), "", input)
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockPackageDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				err := packageDomain.InternetPackage().Update(context.Background(), "package-123", input)
				So(err, ShouldBeNil)
			})
		})

		Convey("Delete", func() {
			Convey("ID is empty", func() {
				err := packageDomain.InternetPackage().Delete(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockPackageDatabasePort.EXPECT().Delete(gomock.Any()).Return(nil).Times(1)

				err := packageDomain.InternetPackage().Delete(context.Background(), "package-123")
				So(err, ShouldBeNil)
			})
		})
	})
}
