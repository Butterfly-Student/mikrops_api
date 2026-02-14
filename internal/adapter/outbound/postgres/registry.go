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

func (s *adapter) Mikrotik() outbound_port.MikrotikDatabasePort {
	return NewMikrotikAdapter(s.db)
}

func (s *adapter) BandwidthProfile() outbound_port.BandwidthProfileDatabasePort {
	return NewBandwidthProfileAdapter(s.db)
}

func (s *adapter) Customer() outbound_port.CustomerDatabasePort {
	return NewCustomerAdapter(s.db)
}

func (s *adapter) SystemSetting() outbound_port.SystemSettingDatabasePort {
	return NewSystemSettingAdapter(s.db)
}

func (s *adapter) Payment() outbound_port.PaymentDatabasePort {
	return NewPaymentAdapter(s.db)
}

func (s *adapter) PaymentAllocation() outbound_port.PaymentAllocationDatabasePort {
	return NewPaymentAllocationAdapter(s.db)
}

func (s *adapter) Invoice() outbound_port.InvoiceDatabasePort {
	return NewInvoiceAdapter(s.db)
}

func (s *adapter) InvoiceItem() outbound_port.InvoiceItemDatabasePort {
	return NewInvoiceItemAdapter(s.db)
}

func (s *adapter) CashCategory() outbound_port.CashCategoryDatabasePort {
	return NewCashCategoryAdapter(s.db)
}

func (s *adapter) CashTransaction() outbound_port.CashTransactionDatabasePort {
	return NewCashTransactionAdapter(s.db)
}

func (s *adapter) Notification() outbound_port.NotificationDatabasePort {
	return NewNotificationAdapter(s.db)
}

func (s *adapter) NotificationTemplate() outbound_port.NotificationTemplateDatabasePort {
	return NewNotificationTemplateAdapter(s.db)
}
