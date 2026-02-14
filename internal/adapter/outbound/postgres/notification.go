package postgres_outbound_adapter

import (
	"errors"
	"time"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"gorm.io/gorm"
)

const tableNotification = "notifications"

type NotificationAdapter struct {
	db *gorm.DB
}

func NewNotificationAdapter(db *gorm.DB) outbound_port.NotificationDatabasePort {
	return &NotificationAdapter{db: db}
}

func (a *NotificationAdapter) Create(notification *model.Notification) error {
	return a.db.Create(notification).Error
}

func (a *NotificationAdapter) FindByID(id string) (*model.Notification, error) {
	var notification model.Notification
	err := a.db.Where("id = ? AND deleted_at IS NULL", id).First(&notification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("notification not found")
		}
		return nil, err
	}
	return &notification, nil
}

func (a *NotificationAdapter) FindAll() ([]model.Notification, error) {
	var notifications []model.Notification
	if err := a.db.Where("deleted_at IS NULL").Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (a *NotificationAdapter) Find(filter model.NotificationFilter) ([]model.Notification, error) {
	var notifications []model.Notification
	query := a.db.Where("deleted_at IS NULL")

	if !filter.IsEmpty() {
		if len(filter.IDs) > 0 {
			query = query.Where("id IN ?", filter.IDs)
		}
		if len(filter.Types) > 0 {
			query = query.Where("type IN ?", filter.Types)
		}
		if len(filter.Status) > 0 {
			query = query.Where("status IN ?", filter.Status)
		}
		if len(filter.Recipients) > 0 {
			query = query.Where("recipient IN ?", filter.Recipients)
		}
		if len(filter.CustomerIDs) > 0 {
			query = query.Where("customer_id IN ?", filter.CustomerIDs)
		}
		if len(filter.InvoiceIDs) > 0 {
			query = query.Where("invoice_id IN ?", filter.InvoiceIDs)
		}
		if len(filter.PaymentIDs) > 0 {
			query = query.Where("payment_id IN ?", filter.PaymentIDs)
		}
		if filter.SentStart != nil {
			query = query.Where("sent_at >= ?", *filter.SentStart)
		}
		if filter.SentEnd != nil {
			query = query.Where("sent_at <= ?", *filter.SentEnd)
		}
		if filter.Search != nil {
			search := "%" + *filter.Search + "%"
			query = query.Where("recipient ILIKE ? OR content ILIKE ?", search, search)
		}
	}

	if err := query.Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (a *NotificationAdapter) Update(notification *model.Notification) error {
	result := a.db.Save(notification)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}

func (a *NotificationAdapter) Delete(id string) error {
	result := a.db.Where("id = ?", id).Update("deleted_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}

func (a *NotificationAdapter) FindByStatus(statuses []string) ([]model.Notification, error) {
	var notifications []model.Notification
	if err := a.db.Where("status IN ?", statuses).Where("deleted_at IS NULL").Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (a *NotificationAdapter) FindByRecipient(recipient string) ([]model.Notification, error) {
	var notifications []model.Notification
	if err := a.db.Where("recipient = ?", recipient).Where("deleted_at IS NULL").Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}
