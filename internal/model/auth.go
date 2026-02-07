package model

type StaffAuthRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CustomerAuthRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	TenantSlug string `json:"tenant_slug" binding:"required"`
}

type AuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type StaffJWTClaims struct {
	StaffID     string   `json:"staff_id"`
	TenantID    string   `json:"tenant_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type CustomerJWTClaims struct {
	CustomerID string `json:"customer_id"`
	TenantID   string `json:"tenant_id"`
}
