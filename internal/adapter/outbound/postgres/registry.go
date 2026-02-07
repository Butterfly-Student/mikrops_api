package postgres_outbound_adapter

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"

	outbound_port "mikrops/internal/port/outbound"
)

type adapter struct {
	db *gorm.DB
	tx *gorm.DB
}

func NewAdapter(db *gorm.DB) outbound_port.DatabasePort {
	return &adapter{
		db: db,
	}
}

func (s *adapter) getDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
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
		db: s.db,
		tx: tx,
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

func (s *adapter) Client() outbound_port.ClientDatabasePort {
	return NewClientAdapter(s.getDB())
}
