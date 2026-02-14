package cash

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/log"
)

type CashDomain interface {
	CreateCategory(ctx context.Context, input model.CashCategoryInput) (*model.CashCategory, error)
	GetCategory(ctx context.Context, id string) (*model.CashCategory, error)
	ListCategories(ctx context.Context, filter model.CashCategoryFilter) ([]model.CashCategory, error)
	UpdateCategory(ctx context.Context, id string, input model.CashCategoryInput) (*model.CashCategory, error)
	DeleteCategory(ctx context.Context, id string) error
	CreateTransaction(ctx context.Context, input model.CashTransactionInput) (*model.CashTransaction, error)
	GetTransaction(ctx context.Context, id string) (*model.CashTransaction, error)
	ListTransactions(ctx context.Context, filter model.CashTransactionFilter) ([]model.CashTransaction, error)
	UpdateTransaction(ctx context.Context, id string, input model.CashTransactionInput) (*model.CashTransaction, error)
	DeleteTransaction(ctx context.Context, id string) error
	ApproveTransaction(ctx context.Context, id string, userID string) error
	RejectTransaction(ctx context.Context, id string, userID string, reason string) error
	AutoRecordPayment(ctx context.Context, payment *model.Payment) error
	GetBalance(ctx context.Context, startDate time.Time, endDate time.Time) (float64, error)
}

type domain struct {
	dbPort outbound_port.DatabasePort
}

func NewCashDomain(
	dbPort outbound_port.DatabasePort,
) CashDomain {
	return &domain{
		dbPort: dbPort,
	}
}

func (d *domain) CreateCategory(ctx context.Context, input model.CashCategoryInput) (*model.CashCategory, error) {
	category := &model.CashCategory{
		Code:      "",
		Name:      "",
		Type:      "income",
		IsActive:  true,
		SortOrder: 0,
	}

	if input.Code != nil {
		category.Code = *input.Code
	}
	if input.Name != nil {
		category.Name = *input.Name
	}
	if input.Type != nil {
		category.Type = *input.Type
	}
	if input.ParentCategoryID != nil {
		category.ParentCategoryID = input.ParentCategoryID
	}
	if input.Description != nil {
		category.Description = input.Description
	}
	if input.IsSystem != nil {
		category.IsSystem = *input.IsSystem
	}
	if input.IsActive != nil {
		category.IsActive = *input.IsActive
	}
	if input.SortOrder != nil {
		category.SortOrder = *input.SortOrder
	}

	err := d.dbPort.CashCategory().Create(category)
	if err != nil {
		return nil, err
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Created cash category %s", category.Code))
	return category, nil
}

func (d *domain) GetCategory(ctx context.Context, id string) (*model.CashCategory, error) {
	return d.dbPort.CashCategory().FindByID(id)
}

func (d *domain) ListCategories(ctx context.Context, filter model.CashCategoryFilter) ([]model.CashCategory, error) {
	if filter.IsEmpty() {
		return d.dbPort.CashCategory().FindAll()
	}
	return d.dbPort.CashCategory().Find(filter)
}

func (d *domain) UpdateCategory(ctx context.Context, id string, input model.CashCategoryInput) (*model.CashCategory, error) {
	category, err := d.dbPort.CashCategory().FindByID(id)
	if err != nil {
		return nil, errors.New("category not found")
	}

	if input.Name != nil {
		category.Name = *input.Name
	}
	if input.Type != nil {
		category.Type = *input.Type
	}
	if input.ParentCategoryID != nil {
		category.ParentCategoryID = input.ParentCategoryID
	}
	if input.Description != nil {
		category.Description = input.Description
	}
	if input.IsActive != nil {
		category.IsActive = *input.IsActive
	}
	if input.SortOrder != nil {
		category.SortOrder = *input.SortOrder
	}

	err = d.dbPort.CashCategory().Update(category)
	if err != nil {
		return nil, err
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Updated cash category %s", category.Code))
	return category, nil
}

func (d *domain) DeleteCategory(ctx context.Context, id string) error {
	category, err := d.dbPort.CashCategory().FindByID(id)
	if err != nil {
		return err
	}

	if category.IsSystem {
		return errors.New("cannot delete system category")
	}

	err = d.dbPort.CashCategory().Delete(id)
	if err != nil {
		return err
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Deleted cash category %s", id))
	return nil
}

func (d *domain) CreateTransaction(ctx context.Context, input model.CashTransactionInput) (*model.CashTransaction, error) {
	transaction := &model.CashTransaction{
		TransactionNumber: "",
		TransactionDate:   time.Now(),
		Type:              "income",
		CategoryID:        input.CategoryID,
		Amount:            0,
		RequiresApproval:  false,
	}

	if input.TransactionNumber != nil {
		transaction.TransactionNumber = *input.TransactionNumber
	}
	if input.TransactionDate != nil {
		transaction.TransactionDate = *input.TransactionDate
	}
	if input.Type != nil {
		transaction.Type = *input.Type
	}
	if input.Amount != nil {
		transaction.Amount = *input.Amount
	}
	if input.PaymentMethod != nil {
		transaction.PaymentMethod = input.PaymentMethod
	}
	if input.Description != nil {
		transaction.Description = input.Description
	}
	if input.ReferenceType != nil {
		transaction.ReferenceType = input.ReferenceType
	}
	if input.ReferenceID != nil {
		transaction.ReferenceID = input.ReferenceID
	}
	if input.CustomerID != nil {
		transaction.CustomerID = input.CustomerID
	}
	if input.AccountName != nil {
		transaction.AccountName = input.AccountName
	}
	if input.AccountNumber != nil {
		transaction.AccountNumber = input.AccountNumber
	}
	if input.ProofImage != nil {
		transaction.ProofImage = input.ProofImage
	}
	if input.ReceiptNumber != nil {
		transaction.ReceiptNumber = input.ReceiptNumber
	}
	if input.RequiresApproval != nil {
		transaction.RequiresApproval = *input.RequiresApproval
	}
	if input.Notes != nil {
		transaction.Notes = input.Notes
	}

	model.CashTransactionPrepare(transaction)

	err := d.dbPort.CashTransaction().Create(transaction)
	if err != nil {
		return nil, err
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Created cash transaction %s", transaction.TransactionNumber))
	return transaction, nil
}

func (d *domain) GetTransaction(ctx context.Context, id string) (*model.CashTransaction, error) {
	return d.dbPort.CashTransaction().FindByID(id)
}

func (d *domain) ListTransactions(ctx context.Context, filter model.CashTransactionFilter) ([]model.CashTransaction, error) {
	if filter.IsEmpty() {
		return d.dbPort.CashTransaction().FindAll()
	}
	return d.dbPort.CashTransaction().Find(filter)
}

func (d *domain) UpdateTransaction(ctx context.Context, id string, input model.CashTransactionInput) (*model.CashTransaction, error) {
	transaction, err := d.dbPort.CashTransaction().FindByID(id)
	if err != nil {
		return nil, errors.New("transaction not found")
	}

	if input.Amount != nil {
		transaction.Amount = *input.Amount
	}
	if input.PaymentMethod != nil {
		transaction.PaymentMethod = input.PaymentMethod
	}
	if input.Description != nil {
		transaction.Description = input.Description
	}
	if input.ReferenceType != nil {
		transaction.ReferenceType = input.ReferenceType
	}
	if input.ReferenceID != nil {
		transaction.ReferenceID = input.ReferenceID
	}
	if input.CustomerID != nil {
		transaction.CustomerID = input.CustomerID
	}
	if input.AccountName != nil {
		transaction.AccountName = input.AccountName
	}
	if input.AccountNumber != nil {
		transaction.AccountNumber = input.AccountNumber
	}
	if input.ProofImage != nil {
		transaction.ProofImage = input.ProofImage
	}
	if input.ReceiptNumber != nil {
		transaction.ReceiptNumber = input.ReceiptNumber
	}
	if input.Notes != nil {
		transaction.Notes = input.Notes
	}

	err = d.dbPort.CashTransaction().Update(transaction)
	if err != nil {
		return nil, err
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Updated cash transaction %s", transaction.TransactionNumber))
	return transaction, nil
}

func (d *domain) DeleteTransaction(ctx context.Context, id string) error {
	err := d.dbPort.CashTransaction().Delete(id)
	if err != nil {
		return err
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Deleted cash transaction %s", id))
	return nil
}

func (d *domain) ApproveTransaction(ctx context.Context, id string, userID string) error {
	transaction, err := d.dbPort.CashTransaction().FindByID(id)
	if err != nil {
		return err
	}

	if transaction.ApprovalStatus != "pending" {
		return errors.New("transaction is not pending")
	}

	transaction.ApprovalStatus = "approved"
	userUUID, _ := uuid.Parse(userID)
	transaction.ApprovedBy = &userUUID
	now := time.Now()
	transaction.ApprovedAt = &now

	return d.dbPort.CashTransaction().Update(transaction)
}

func (d *domain) RejectTransaction(ctx context.Context, id string, userID string, reason string) error {
	transaction, err := d.dbPort.CashTransaction().FindByID(id)
	if err != nil {
		return err
	}

	if transaction.ApprovalStatus != "pending" {
		return errors.New("transaction is not pending")
	}

	transaction.ApprovalStatus = "rejected"
	userUUID, _ := uuid.Parse(userID)
	transaction.ApprovedBy = &userUUID
	if reason != "" {
		transaction.Notes = &reason
	}

	return d.dbPort.CashTransaction().Update(transaction)
}

func (d *domain) AutoRecordPayment(ctx context.Context, payment *model.Payment) error {
	incomeCategory, err := d.dbPort.CashCategory().FindByCode("INC-SUB")
	if err != nil {
		return err
	}

	description := ""
	referenceType := "payment"

	transaction := model.CashTransaction{
		TransactionNumber: generatePaymentCashNumber(payment.PaymentNumber),
		TransactionDate:   payment.PaymentDate,
		Type:              "income",
		CategoryID:        incomeCategory.ID,
		Amount:            payment.AllocatedAmount,
		PaymentMethod:     &payment.PaymentMethod,
		Description:       &description,
		ReferenceType:     &referenceType,
		ReferenceID:       &payment.ID,
		CustomerID:        &payment.CustomerID,
		RequiresApproval:  false,
	}

	model.CashTransactionPrepare(&transaction)

	return d.dbPort.CashTransaction().Create(&transaction)
}

func (d *domain) GetBalance(ctx context.Context, startDate time.Time, endDate time.Time) (float64, error) {
	income, err := d.dbPort.CashTransaction().GetIncomeTotal(startDate, endDate)
	if err != nil {
		return 0, err
	}

	expense, err := d.dbPort.CashTransaction().GetExpenseTotal(startDate, endDate)
	if err != nil {
		return 0, err
	}

	return income - expense, nil
}

func generatePaymentCashNumber(paymentNumber string) string {
	return fmt.Sprintf("PAY-%s-TRX", paymentNumber)
}
