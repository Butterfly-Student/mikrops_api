//go:build integration
// +build integration

package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/casbin/casbin/v3"
	casbinmodel "github.com/casbin/casbin/v3/model"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	"go-template/internal/domain"
	"go-template/internal/model"
	"go-template/tests/helpers"
	"go-template/utils/hash"
	"go-template/utils/token"
)

func TestAuthIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set JWT environment variables
	os.Setenv("JWT_SECRET", "test-secret-key-for-integration-testing")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key-for-integration-testing")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("JWT_REFRESH_SECRET")

	ctx := context.Background()

	// Use shared helper for container setup
	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to setup postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	// Use GORM AutoMigrate
	err = pgContainer.DB.AutoMigrate(&model.AdminUser{})
	if err != nil {
		t.Fatalf("Failed to migrate table: %v", err)
	}

	adapter := postgres_outbound_adapter.NewUserAdapter(pgContainer.DB)
	dbAdapter := postgres_outbound_adapter.NewAdapter(pgContainer.DB)
	enforcer, _ := casbin.NewEnforcer(casbinmodel.NewModel())
	dom := domain.NewDomain(dbAdapter, nil, nil, nil, nil, enforcer)

	Convey("Test Auth Integration with PostgreSQL", t, func() {
		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.AdminUser{})

		Convey("Full Authentication Flow", func() {
			email := "auth-integration-" + time.Now().Format("20060102150405") + "@example.com"
			password := "TestPassword123!"
			hashedPassword, _ := hash.HashPassword(password)
			isActive := true

			Convey("Register creates a new user", func() {
				user := &model.AdminUser{
					FullName:     "Auth Integration User",
					Email:        email,
					PasswordHash: hashedPassword,
					Role:         model.AdminRoleCS,
					IsActive:     &isActive,
				}
				err := adapter.Create(user)
				So(err, ShouldBeNil)
				So(user.ID.String(), ShouldNotBeEmpty)

				Convey("Login returns access and refresh tokens", func() {
					loginReq := model.LoginRequest{
						Email:    email,
						Password: password,
					}
					loginRes, err := dom.Auth().Login(loginReq)
					So(err, ShouldBeNil)
					So(loginRes.AccessToken, ShouldNotBeEmpty)
					So(loginRes.RefreshToken, ShouldNotBeEmpty)

					Convey("Refresh token generates new access token", func() {
						refreshReq := model.RefreshTokenRequest{
							RefreshToken: loginRes.RefreshToken,
						}
						newLoginRes, err := dom.Auth().RefreshToken(refreshReq)
						So(err, ShouldBeNil)
						So(newLoginRes.AccessToken, ShouldNotBeEmpty)
						So(newLoginRes.RefreshToken, ShouldNotBeEmpty)
					})

					Convey("Change password updates user password", func() {
						changePwdReq := model.ChangePasswordRequest{
							OldPassword: password,
							NewPassword: "NewPassword456!",
						}
						err := dom.Auth().ChangePassword(user.ID, changePwdReq)
						So(err, ShouldBeNil)

						// Verify new password works
						loginReq2 := model.LoginRequest{
							Email:    email,
							Password: "NewPassword456!",
						}
						loginRes2, err := dom.Auth().Login(loginReq2)
						So(err, ShouldBeNil)
						So(loginRes2.AccessToken, ShouldNotBeEmpty)
					})
				})
			})

			Convey("Register with duplicate email returns error", func() {
				user1 := &model.AdminUser{
					FullName:     "User One",
					Email:        email,
					PasswordHash: hashedPassword,
					Role:         model.AdminRoleCS,
					IsActive:     &isActive,
				}
				err := adapter.Create(user1)
				So(err, ShouldBeNil)

				user2 := &model.AdminUser{
					FullName:     "User Two",
					Email:        email,
					PasswordHash: hashedPassword,
					Role:         model.AdminRoleCS,
					IsActive:     &isActive,
				}
				err = adapter.Create(user2)
				So(err, ShouldNotBeNil)
			})

			Convey("Login with invalid credentials returns error", func() {
				loginReq := model.LoginRequest{
					Email:    "nonexistent@example.com",
					Password: password,
				}
				_, err := dom.Auth().Login(loginReq)
				So(err, ShouldNotBeNil)
			})

			Convey("Change password with wrong old password returns error", func() {
				user := &model.AdminUser{
					FullName:     "Password Change User",
					Email:        "password-change-" + time.Now().Format("20060102150405") + "@example.com",
					PasswordHash: hashedPassword,
					Role:         model.AdminRoleCS,
					IsActive:     &isActive,
				}
				err := adapter.Create(user)
				So(err, ShouldBeNil)

				changePwdReq := model.ChangePasswordRequest{
					OldPassword: "WrongPassword",
					NewPassword: "NewPassword456!",
				}
				err = dom.Auth().ChangePassword(user.ID, changePwdReq)
				So(err, ShouldNotBeNil)
			})

			Convey("Refresh with invalid token returns error", func() {
				refreshReq := model.RefreshTokenRequest{
					RefreshToken: "invalid-refresh-token",
				}
				_, err := dom.Auth().RefreshToken(refreshReq)
				So(err, ShouldNotBeNil)
			})

			Convey("Token validation works correctly", func() {
				testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440123")
				accessToken, _ := token.GenerateAccessToken(testUUID.String(), "admin")
				So(accessToken, ShouldNotBeEmpty)

				refreshToken, _ := token.GenerateRefreshToken(testUUID.String())
				So(refreshToken, ShouldNotBeEmpty)

				// Validate access token
				claims, err := token.ValidateToken(accessToken, false)
				So(err, ShouldBeNil)
				So(claims["sub"], ShouldEqual, testUUID.String())
				So(claims["role"], ShouldEqual, "admin")

				// Validate refresh token
				refreshClaims, err := token.ValidateToken(refreshToken, true)
				So(err, ShouldBeNil)
				So(refreshClaims["sub"], ShouldEqual, testUUID.String())

				// Invalid token
				_, err = token.ValidateToken("invalid-token", false)
				So(err, ShouldNotBeNil)
			})
		})
	})
}
