package postgres_outbound_adapter

import (
	"gorm.io/gorm"

	outbound_port "go-template/internal/port/outbound"
)

type adapter struct {
	db *gorm.DB
}

func NewAdapter(db *gorm.DB) outbound_port.DatabasePort {
	return &adapter{
		db: db,
	}
}

// DoInTransaction executes a function within a database transaction
func (s *adapter) DoInTransaction(txFunc outbound_port.InTransaction) (out interface{}, err error) {
	var result interface{}
	var txErr error

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create a new adapter with the transaction
		txAdapter := &adapter{
			db: tx,
		}

		// Execute the transaction function
		result, txErr = txFunc(txAdapter)
		return txErr
	})

	if err != nil {
		// Transaction was rolled back
		return nil, err
	}

	// Transaction was committed
	return result, nil
}

func (s *adapter) Client() outbound_port.ClientDatabasePort {
	return NewClientAdapter(s.db)
}

func (s *adapter) User() outbound_port.UserDatabasePort {
	return NewUserAdapter(s.db)
}
