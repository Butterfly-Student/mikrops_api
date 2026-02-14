package outbound_port

import "gorm.io/gorm"

type InTransaction func(repoRegistry DatabasePort) (interface{}, error)

type DatabasePort interface {
	Client() ClientDatabasePort
	User() UserDatabasePort
	Mikrotik() MikrotikDatabasePort
	BandwidthProfile() BandwidthProfileDatabasePort
	Customer() CustomerDatabasePort
	SystemSetting() SystemSettingDatabasePort
	Invoice() InvoiceDatabasePort
	Payment() PaymentDatabasePort
	DoInTransaction(txFunc InTransaction) (out interface{}, err error)
}

// DatabaseExecutor is now GORM's *gorm.DB
// We keep this interface for compatibility, but it now wraps gorm.DB
type DatabaseExecutor interface {
	*gorm.DB
}
