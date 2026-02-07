package activity

import (
	"context"

	"github.com/google/uuid"
)

type key int

const (
	TransactionID key = iota
	Action
	ClientID
	Payload
	Result
	TenantID
	StaffID
	CustomerID
	UserRole
	UserPermissions
)

func NewContext(action string) context.Context {
	trxID := uuid.New().String()
	ctx := context.WithValue(context.Background(), TransactionID, trxID)
	return context.WithValue(ctx, Action, action)
}

func GetTransactionID(ctx context.Context) (string, bool) {
	trxID, ok := ctx.Value(TransactionID).(string)
	return trxID, ok
}

func WithAction(ctx context.Context, action string) context.Context {
	return context.WithValue(ctx, Action, action)
}

func GetAction(ctx context.Context) (string, bool) {
	action, ok := ctx.Value(Action).(string)
	return action, ok
}

func WithClientID(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, ClientID, clientID)
}

func GetClientID(ctx context.Context) (string, bool) {
	clientID, ok := ctx.Value(ClientID).(string)
	return clientID, ok
}

func WithPayload(ctx context.Context, payload interface{}) context.Context {
	return context.WithValue(ctx, Payload, payload)
}

func GetPayload(ctx context.Context) interface{} {
	return ctx.Value(Payload)
}

func WithResult(ctx context.Context, payload interface{}) context.Context {
	return context.WithValue(ctx, Result, payload)
}

func GetResult(ctx context.Context) interface{} {
	return ctx.Value(Result)
}

func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantID, tenantID)
}

func GetTenantID(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(TenantID).(string)
	return tenantID, ok
}

func WithStaffID(ctx context.Context, staffID string) context.Context {
	return context.WithValue(ctx, StaffID, staffID)
}

func GetStaffID(ctx context.Context) (string, bool) {
	staffID, ok := ctx.Value(StaffID).(string)
	return staffID, ok
}

func WithCustomerID(ctx context.Context, customerID string) context.Context {
	return context.WithValue(ctx, CustomerID, customerID)
}

func GetCustomerID(ctx context.Context) (string, bool) {
	customerID, ok := ctx.Value(CustomerID).(string)
	return customerID, ok
}

func WithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, UserRole, role)
}

func GetUserRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRole).(string)
	return role, ok
}

func WithUserPermissions(ctx context.Context, permissions []string) context.Context {
	return context.WithValue(ctx, UserPermissions, permissions)
}

func GetUserPermissions(ctx context.Context) ([]string, bool) {
	permissions, ok := ctx.Value(UserPermissions).([]string)
	return permissions, ok
}

func GetFields(ctx context.Context) map[string]interface{} {
	fields := make(map[string]interface{})

	if id, ok := GetTransactionID(ctx); ok {
		fields["transaction_id"] = id
	}

	if action, ok := GetAction(ctx); ok {
		fields["action"] = action
	}

	if clientID, ok := GetClientID(ctx); ok {
		fields["client_id"] = clientID
	}

	if tenantID, ok := GetTenantID(ctx); ok {
		fields["tenant_id"] = tenantID
	}

	if staffID, ok := GetStaffID(ctx); ok {
		fields["staff_id"] = staffID
	}

	if customerID, ok := GetCustomerID(ctx); ok {
		fields["customer_id"] = customerID
	}

	fields["payload"] = GetPayload(ctx)
	fields["result"] = GetResult(ctx)

	return fields
}
