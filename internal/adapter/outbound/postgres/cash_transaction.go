package postgres_outbound_adapter

import (
	"errors"
	"time"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"gorm.io/gorm"
)

const tableCashTransaction = "cash_transactions"

type CashTransactionAdapter struct {
	db *gorm.DB
}

func NewCashTransactionAdapter(db *gorm.DB) outbound_port.CashTransactionDatabasePort {
	return &CashTransactionAdapter{db: db}
}

func (a *CashTransactionAdapter) Create(transaction *model.CashTransaction) error {
	model.CashTransactionPrepare(transaction)
	return a.db.Create(transaction).Error
}

func (a *CashTransactionAdapter) FindByID(id string) (*model.CashTransaction, error) {
	var transaction model.CashTransaction
	err := a.db.Preload("Category").Preload("Customer").Preload("ApprovedByUser").Preload("ProcessedByUser").Where("id = ? AND deleted_at IS NULL", id).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}
	return &transaction, nil
}

func (a *CashTransactionAdapter) FindAll() ([]model.CashTransaction, error) {
	var transactions []model.CashTransaction
	if err := a.db.Preload("Category").Preload("Customer").Where("deleted_at IS NULL").Order("transaction_date DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (a *CashTransactionAdapter) Find(filter model.CashTransactionFilter) ([]model.CashTransaction, error) {
	var transactions []model.CashTransaction
	query := a.db.Preload("Category").Preload("Customer").Preload("ApprovedByUser").Where("deleted_at IS NULL")

	if !filter.IsEmpty() {
		if len(filter.IDs) > 0 {
			query = query.Where("id IN ?", filter.IDs)
		}
		if len(filter.TransactionNumber) > 0 {
			query = query.Where("transaction_number IN ?", filter.TransactionNumber)
		}
		if filter.Type != nil {
			query = query.Where("type = ?", *filter.Type)
		}
		if len(filter.CategoryIDs) > 0 {
			query = query.Where("category_id IN ?", filter.CategoryIDs)
		}
		if filter.ReferenceType != nil {
			query = query.Where("reference_type = ?", *filter.ReferenceType)
		}
		if filter.ReferenceID != nil {
			query = query.Where("reference_id = ?", *filter.ReferenceID)
		}
		if len(filter.CustomerIDs) > 0 {
			query = query.Where("customer_id IN ?", filter.CustomerIDs)
		}
		if filter.DateStart != nil {
			query = query.Where("transaction_date >= ?", *filter.DateStart)
		}
		if filter.DateEnd != nil {
			query = query.Where("transaction_date <= ?", *filter.DateEnd)
		}
		if filter.AmountMin != nil {
			query = query.Where("amount >= ?", *filter.AmountMin)
		}
		if filter.AmountMax != nil {
			query = query.Where("amount <= ?", *filter.AmountMax)
		}
		if filter.ApprovalStatus != nil {
			query = query.Where("approval_status = ?", *filter.ApprovalStatus)
		}
		if filter.RequiresApproval != nil {
			query = query.Where("requires_approval = ?", *filter.RequiresApproval)
		}
		if filter.Search != nil {
			search := "%" + *filter.Search + "%"
			query = query.Where("transaction_number ILIKE ? OR description ILIKE ? OR receipt_number ILIKE ?", search, search, search)
		}
	}

	if err := query.Order("transaction_date DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (a *CashTransactionAdapter) Update(transaction *model.CashTransaction) error {
	result := a.db.Save(transaction)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}

func (a *CashTransactionAdapter) Delete(id string) error {
	result := a.db.Where("id = ?", id).Update("deleted_at", "NOW()")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}

func (a *CashTransactionAdapter) FindByNumber(number string) (*model.CashTransaction, error) {
	var transaction model.CashTransaction
	err := a.db.Preload("Category").Preload("Customer").Where("transaction_number = ? AND deleted_at IS NULL", number).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}
	return &transaction, nil
}

func (a *CashTransactionAdapter) FindByReference(refType string, refID string) ([]model.CashTransaction, error) {
	var transactions []model.CashTransaction
	if err := a.db.Where("reference_type = ? AND reference_id = ? AND deleted_at IS NULL", refType, refID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (a *CashTransactionAdapter) GetIncomeTotal(startDate time.Time, endDate time.Time) (float64, error) {
	var total float64
	err := a.db.Table(tableCashTransaction).
		Where("type = ? AND transaction_date >= ? AND transaction_date <= ? AND deleted_at IS NULL", "income", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (a *CashTransactionAdapter) GetExpenseTotal(startDate time.Time, endDate time.Time) (float64, error) {
	var total float64
	err := a.db.Table(tableCashTransaction).
		Where("type = ? AND transaction_date >= ? AND transaction_date <= ? AND deleted_at IS NULL", "expense", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}
