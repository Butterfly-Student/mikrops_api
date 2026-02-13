package outbound_port

import "gorm.io/gorm"

//go:generate mockgen -source=registry_database.go -destination=./../../../tests/mocks/port/mock_registry_database.go
type InTransaction func(repoRegistry DatabasePort) (interface{}, error)

type DatabasePort interface {
	Client() ClientDatabasePort
	User() UserDatabasePort
	DoInTransaction(txFunc InTransaction) (out interface{}, err error)
}

// DatabaseExecutor is now GORM's *gorm.DB
// We keep this interface for compatibility, but it now wraps gorm.DB
type DatabaseExecutor interface {
	*gorm.DB
}
