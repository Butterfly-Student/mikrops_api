package outbound_port

//go:generate mockgen -source=registry_database.go -destination=./../../../tests/mocks/port/mock_registry_database.go
type InTransaction func(repoRegistry DatabasePort) (interface{}, error)

type DatabasePort interface {
	Client() ClientDatabasePort
	DoInTransaction(txFunc InTransaction) (out interface{}, err error)
}
