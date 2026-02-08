package postgres_outbound_adapter

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"

	outbound_port "mikrops/internal/port/outbound"
)

type adapter struct {
	db       *gorm.DB
	tx       *gorm.DB
	tenantID string
}

func NewAdapter(db *gorm.DB) outbound_port.DatabasePort {
	return &adapter{
		db: db,
	}
}

func (s *adapter) getDB() *gorm.DB {
	db := s.db
	if s.tx != nil {
		db = s.tx
	}

	// Apply tenant scoping if tenantID is set
	if s.tenantID != "" {
		db = db.Where("tenant_id = ?", s.tenantID)
	}

	return db
}

func (s *adapter) DoInTransaction(txFunc outbound_port.InTransaction) (out interface{}, err error) {
	if s.tx != nil {
		return txFunc(s)
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			switch x := p.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("unknown panic")
			}
		} else if err != nil {
			xerr := tx.Rollback().Error
			if xerr != nil {
				err = errors.Wrap(err, xerr.Error())
			}
		} else {
			err = tx.Commit().Error
		}
	}()

	reg := &adapter{
		db:       s.db,
		tx:       tx,
		tenantID: s.tenantID,
	}
	out, err = txFunc(reg)
	if err != nil {
		if out != nil {
			return out, err
		}
		return nil, err
	}
	return
}

func (s *adapter) WithTenantScope(tenantID string) outbound_port.DatabasePort {
	return &adapter{
		db:       s.db,
		tx:       s.tx,
		tenantID: tenantID,
	}
}

// Adapter accessors
func (s *adapter) Client() outbound_port.ClientDatabasePort {
	return NewClientAdapter(s.getDB())
}

func (s *adapter) Tenant() outbound_port.TenantDatabasePort {
	return NewTenantAdapter(s.getDB())
}

func (s *adapter) Role() outbound_port.RoleDatabasePort {
	return NewRoleAdapter(s.getDB())
}

func (s *adapter) Permission() outbound_port.PermissionDatabasePort {
	return NewPermissionAdapter(s.getDB())
}

func (s *adapter) Staff() outbound_port.StaffDatabasePort {
	return NewStaffAdapter(s.getDB())
}

func (s *adapter) Nas() outbound_port.NasDatabasePort {
	return NewNasAdapter(s.getDB())
}

func (s *adapter) InternetPackage() outbound_port.InternetPackageDatabasePort {
	return NewInternetPackageAdapter(s.getDB())
}

func (s *adapter) Customer() outbound_port.CustomerDatabasePort {
	return NewCustomerAdapter(s.getDB())
}

func (s *adapter) Subscription() outbound_port.SubscriptionDatabasePort {
	return NewSubscriptionAdapter(s.getDB())
}

func (s *adapter) PaymentMethod() outbound_port.PaymentMethodDatabasePort {
	return NewPaymentMethodAdapter(s.getDB())
}

func (s *adapter) Invoice() outbound_port.InvoiceDatabasePort {
	return NewInvoiceAdapter(s.getDB())
}

func (s *adapter) Payment() outbound_port.PaymentDatabasePort {
	return NewPaymentAdapter(s.getDB())
}
