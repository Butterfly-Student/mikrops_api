package auth

import (
	"context"
	"os"
	"time"

	"github.com/palantir/stacktrace"
	"golang.org/x/crypto/bcrypt"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
	"mikrops/utils/jwt"
)

type AuthDomain interface {
	StaffLogin(ctx context.Context, req model.StaffAuthRequest) (model.AuthTokenResponse, error)
	CustomerLogin(ctx context.Context, req model.CustomerAuthRequest) (model.AuthTokenResponse, error)
	StaffRefreshToken(ctx context.Context, refreshToken string) (model.AuthTokenResponse, error)
	CustomerRefreshToken(ctx context.Context, refreshToken string) (model.AuthTokenResponse, error)
}

type authDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
}

func NewAuthDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
) AuthDomain {
	return &authDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
	}
}

func (d *authDomain) StaffLogin(ctx context.Context, req model.StaffAuthRequest) (model.AuthTokenResponse, error) {
	if req.Email == "" || req.Password == "" {
		return model.AuthTokenResponse{}, stacktrace.NewError("email and password are required")
	}

	// Find staff by email
	staff, err := d.databasePort.Staff().FindByEmail(req.Email)
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.Propagate(err, "failed to find staff")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(req.Password))
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.NewError("invalid credentials")
	}

	// Get role and permissions
	var role string
	var permissions []string
	if staff.Role != nil {
		role = staff.Role.Name
		// Load permissions from role
		perms, err := d.databasePort.Permission().FindByRoleID(staff.RoleID)
		if err == nil {
			for _, p := range perms {
				permissions = append(permissions, p.Name)
			}
		}
	}

	// Generate JWT tokens
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return model.AuthTokenResponse{}, stacktrace.NewError("JWT_SECRET not configured")
	}

	accessToken, err := jwt.GenerateStaffToken(staff.ID, staff.TenantID, role, permissions, jwtSecret)
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.Propagate(err, "failed to generate access token")
	}

	refreshToken, err := jwt.GenerateStaffRefreshToken(staff.ID, staff.TenantID, jwtSecret)
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.Propagate(err, "failed to generate refresh token")
	}

	// Update last login time
	now := time.Now()
	staff.LastLoginAt = &now
	err = d.databasePort.Staff().Update(staff.ID, staff.StaffInput)
	if err != nil {
		// Log error but don't fail the login
		// TODO: Add logging
	}

	return model.AuthTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    86400, // 24 hours
		TokenType:    "Bearer",
	}, nil
}

func (d *authDomain) CustomerLogin(ctx context.Context, req model.CustomerAuthRequest) (model.AuthTokenResponse, error) {
	if req.Username == "" || req.Password == "" || req.TenantSlug == "" {
		return model.AuthTokenResponse{}, stacktrace.NewError("username, password, and tenant_slug are required")
	}

	// Find tenant by slug
	tenants, err := d.databasePort.Tenant().FindByFilter(model.TenantFilter{Slugs: []string{req.TenantSlug}})
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.Propagate(err, "failed to find tenant")
	}
	if len(tenants) == 0 {
		return model.AuthTokenResponse{}, stacktrace.NewError("tenant not found")
	}
	tenant := tenants[0]

	// Find customer by username
	customer, err := d.databasePort.Customer().FindByUsername(req.Username)
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.Propagate(err, "failed to find customer")
	}

	// Verify tenant ID matches
	if customer.TenantID != tenant.ID {
		return model.AuthTokenResponse{}, stacktrace.NewError("invalid credentials")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(req.Password))
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.NewError("invalid credentials")
	}

	// Generate JWT tokens
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return model.AuthTokenResponse{}, stacktrace.NewError("JWT_SECRET not configured")
	}

	accessToken, err := jwt.GenerateCustomerToken(customer.ID, customer.TenantID, jwtSecret)
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.Propagate(err, "failed to generate access token")
	}

	refreshToken, err := jwt.GenerateCustomerRefreshToken(customer.ID, customer.TenantID, jwtSecret)
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.Propagate(err, "failed to generate refresh token")
	}

	return model.AuthTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    86400, // 24 hours
		TokenType:    "Bearer",
	}, nil
}

func (d *authDomain) StaffRefreshToken(ctx context.Context, refreshToken string) (model.AuthTokenResponse, error) {
	if refreshToken == "" {
		return model.AuthTokenResponse{}, stacktrace.NewError("refresh token is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return model.AuthTokenResponse{}, stacktrace.NewError("JWT_SECRET not configured")
	}

	// Parse refresh token
	claims, err := jwt.ParseRefreshToken(refreshToken, jwtSecret)
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.Propagate(err, "invalid refresh token")
	}

	staffID := claims.Subject
	tenantID := claims.ID

	// Verify staff still exists
	staff, err := d.databasePort.Staff().FindByID(staffID)
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.Propagate(err, "staff not found")
	}

	// Get role and permissions
	var role string
	var permissions []string
	if staff.Role != nil {
		role = staff.Role.Name
		perms, err := d.databasePort.Permission().FindByRoleID(staff.RoleID)
		if err == nil {
			for _, p := range perms {
				permissions = append(permissions, p.Name)
			}
		}
	}

	// Generate new access token
	accessToken, err := jwt.GenerateStaffToken(staffID, tenantID, role, permissions, jwtSecret)
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.Propagate(err, "failed to generate access token")
	}

	// Generate new refresh token
	newRefreshToken, err := jwt.GenerateStaffRefreshToken(staffID, tenantID, jwtSecret)
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.Propagate(err, "failed to generate refresh token")
	}

	return model.AuthTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    86400, // 24 hours
		TokenType:    "Bearer",
	}, nil
}

func (d *authDomain) CustomerRefreshToken(ctx context.Context, refreshToken string) (model.AuthTokenResponse, error) {
	if refreshToken == "" {
		return model.AuthTokenResponse{}, stacktrace.NewError("refresh token is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return model.AuthTokenResponse{}, stacktrace.NewError("JWT_SECRET not configured")
	}

	// Parse refresh token
	claims, err := jwt.ParseRefreshToken(refreshToken, jwtSecret)
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.Propagate(err, "invalid refresh token")
	}

	customerID := claims.Subject
	tenantID := claims.ID

	// Verify customer still exists
	_, err = d.databasePort.Customer().FindByID(customerID)
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.Propagate(err, "customer not found")
	}

	// Generate new access token
	accessToken, err := jwt.GenerateCustomerToken(customerID, tenantID, jwtSecret)
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.Propagate(err, "failed to generate access token")
	}

	// Generate new refresh token
	newRefreshToken, err := jwt.GenerateCustomerRefreshToken(customerID, tenantID, jwtSecret)
	if err != nil {
		return model.AuthTokenResponse{}, stacktrace.Propagate(err, "failed to generate refresh token")
	}

	return model.AuthTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    86400, // 24 hours
		TokenType:    "Bearer",
	}, nil
}
