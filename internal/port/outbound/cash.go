package outbound_port

import (
	"time"

	"go-template/internal/model"
)

type CashCategoryDatabasePort interface {
	Create(category *model.CashCategory) error
	FindByID(id string) (*model.CashCategory, error)
	FindAll() ([]model.CashCategory, error)
	Find(filter model.CashCategoryFilter) ([]model.CashCategory, error)
	Update(category *model.CashCategory) error
	Delete(id string) error
	FindByCode(code string) (*model.CashCategory, error)
	FindByType(cashType string) ([]model.CashCategory, error)
}

type CashTransactionDatabasePort interface {
	Create(transaction *model.CashTransaction) error
	FindByID(id string) (*model.CashTransaction, error)
	FindAll() ([]model.CashTransaction, error)
	Find(filter model.CashTransactionFilter) ([]model.CashTransaction, error)
	Update(transaction *model.CashTransaction) error
	Delete(id string) error
	FindByNumber(number string) (*model.CashTransaction, error)
	FindByReference(refType string, refID string) ([]model.CashTransaction, error)
	GetIncomeTotal(startDate time.Time, endDate time.Time) (float64, error)
	GetExpenseTotal(startDate time.Time, endDate time.Time) (float64, error)
}
