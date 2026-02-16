package outbound_port

//go:generate mockgen -source=registry_database.go -destination=./../../../tests/mocks/port/mock_registry_database.go

type InTransaction func(repoRegistry DatabasePort) (interface{}, error)

type DatabasePort interface {
	Client() ClientDatabasePort
	User() UserDatabasePort
	Mikrotik() MikrotikDatabasePort
	BandwidthProfile() BandwidthProfileDatabasePort
	Customer() CustomerDatabasePort
	SystemSetting() SystemSettingDatabasePort
	Payment() PaymentDatabasePort
	PaymentAllocation() PaymentAllocationDatabasePort
	Invoice() InvoiceDatabasePort
	InvoiceItem() InvoiceItemDatabasePort
	CashCategory() CashCategoryDatabasePort
	CashTransaction() CashTransactionDatabasePort
	Notification() NotificationDatabasePort
	NotificationTemplate() NotificationTemplateDatabasePort
	ActivityLog() ActivityLogDatabasePort
	Refund() RefundDatabasePort
	DoInTransaction(txFunc InTransaction) (out interface{}, err error)
}

// DatabaseExecutor is now GORM's *gorm.DB
// We keep this interface for compatibility, but it now wraps gorm.DB
// NOTE: Commented out because mockgen cannot handle embedded pointer types
// type DatabaseExecutor interface {
// 	*gorm.DB
// }
