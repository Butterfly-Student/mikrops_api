package postgres_outbound_adapter

import (
	"errors"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"gorm.io/gorm"
)

const tableNotificationTemplate = "notification_templates"

type NotificationTemplateAdapter struct {
	db *gorm.DB
}

func NewNotificationTemplateAdapter(db *gorm.DB) outbound_port.NotificationTemplateDatabasePort {
	return &NotificationTemplateAdapter{db: db}
}

func (a *NotificationTemplateAdapter) Create(template *model.NotificationTemplate) error {
	return a.db.Create(template).Error
}

func (a *NotificationTemplateAdapter) FindByID(id string) (*model.NotificationTemplate, error) {
	var template model.NotificationTemplate
	err := a.db.Where("id = ? AND deleted_at IS NULL", id).First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("notification template not found")
		}
		return nil, err
	}
	return &template, nil
}

func (a *NotificationTemplateAdapter) FindAll() ([]model.NotificationTemplate, error) {
	var templates []model.NotificationTemplate
	if err := a.db.Where("deleted_at IS NULL").Order("name ASC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (a *NotificationTemplateAdapter) FindByName(name string) (*model.NotificationTemplate, error) {
	var template model.NotificationTemplate
	err := a.db.Where("name = ? AND deleted_at IS NULL", name).First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("notification template not found")
		}
		return nil, err
	}
	return &template, nil
}

func (a *NotificationTemplateAdapter) Update(template *model.NotificationTemplate) error {
	result := a.db.Save(template)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}

func (a *NotificationTemplateAdapter) Delete(id string) error {
	result := a.db.Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}
