package outbound_port

//go:generate mockgen -source=registry_database.go -destination=./../../../tests/mocks/port/mock_registry_database.go
type InTransaction func(repoRegistry DatabasePort) (interface{}, error)

type DatabasePort interface {
	Client() ClientDatabasePort
	Tenant() TenantDatabasePort
	Role() RoleDatabasePort
	Permission() PermissionDatabasePort
	Staff() StaffDatabasePort
	Nas() NasDatabasePort
	InternetPackage() InternetPackageDatabasePort
	Customer() CustomerDatabasePort
	Subscription() SubscriptionDatabasePort
	PaymentMethod() PaymentMethodDatabasePort
	Invoice() InvoiceDatabasePort
	Payment() PaymentDatabasePort
	TenantSetting() TenantSettingDatabasePort
	CustomerRegistration() CustomerRegistrationDatabasePort
	PppoeAccount() PppoeAccountDatabasePort
	ActivityLog() ActivityLogDatabasePort
	MikrotikSyncLog() MikrotikSyncLogDatabasePort

	DoInTransaction(txFunc InTransaction) (out interface{}, err error)
	WithTenantScope(tenantID string) DatabasePort
}
