package postgres_outbound_adapter_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	postgres_outbound_adapter "mikrops/internal/adapter/outbound/postgres"
	"mikrops/internal/model"
)

func initGormWithMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}

	return gormDB, mock, db
}

func TestClientAdapter(t *testing.T) {
	Convey("Test Postgres Client Adapter", t, func() {
		gormDB, mock, db := initGormWithMock(t)
		defer db.Close()

		adapter := postgres_outbound_adapter.NewClientAdapter(gormDB)

		now := time.Now()
		inputs := []model.ClientInput{
			{
				Name:      "Test Client",
				BearerKey: "test-key",
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		filter := model.ClientFilter{
			IDs: []int{1},
		}

		Convey("Upsert", func() {
			Convey("Success", func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "clients"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()

				err := adapter.Upsert(inputs)
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})
		})

		Convey("FindByFilter", func() {
			Convey("Success", func() {
				rows := sqlmock.NewRows([]string{"id", "name", "bearer_key", "created_at", "updated_at"}).
					AddRow(1, "Test Client", "test-key", now, now)

				mock.ExpectQuery(`SELECT \* FROM "clients"`).
					WillReturnRows(rows)

				results, err := adapter.FindByFilter(filter, false)
				So(err, ShouldBeNil)
				So(len(results), ShouldEqual, 1)
				So(results[0].Name, ShouldEqual, "Test Client")
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("Query error", func() {
				mock.ExpectQuery(`SELECT \* FROM "clients"`).
					WillReturnError(sqlmock.ErrCancelled)

				_, err := adapter.FindByFilter(filter, false)
				So(err, ShouldNotBeNil)
			})

			Convey("Empty result", func() {
				rows := sqlmock.NewRows([]string{"id", "name", "bearer_key", "created_at", "updated_at"})

				mock.ExpectQuery(`SELECT \* FROM "clients"`).
					WillReturnRows(rows)

				results, err := adapter.FindByFilter(filter, false)
				So(err, ShouldBeNil)
				So(len(results), ShouldEqual, 0)
			})
		})

		Convey("DeleteByFilter", func() {
			Convey("Success", func() {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "clients"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()

				err := adapter.DeleteByFilter(filter)
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})
		})

		Convey("IsExists", func() {
			Convey("Exists", func() {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT count\(\*\) FROM "clients"`).
					WillReturnRows(rows)

				exists, err := adapter.IsExists("test-key")
				So(err, ShouldBeNil)
				So(exists, ShouldBeTrue)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("Not exists", func() {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT count\(\*\) FROM "clients"`).
					WillReturnRows(rows)

				exists, err := adapter.IsExists("nonexistent")
				So(err, ShouldBeNil)
				So(exists, ShouldBeFalse)
			})

			Convey("Query error", func() {
				mock.ExpectQuery(`SELECT count\(\*\) FROM "clients"`).
					WillReturnError(sqlmock.ErrCancelled)

				_, err := adapter.IsExists("test-key")
				So(err, ShouldNotBeNil)
			})
		})
	})
}
