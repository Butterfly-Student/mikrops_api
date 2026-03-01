package postgres_outbound_adapter

import (
	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"

	"gorm.io/gorm"
)

type userAdapter struct {
	db *gorm.DB
}

func NewUserAdapter(db *gorm.DB) outbound_port.UserDatabasePort {
	return &userAdapter{
		db: db,
	}
}

func (r *userAdapter) Create(user *model.AdminUser) error {
	return r.db.Create(user).Error
}

func (r *userAdapter) FindByEmail(email string) (*model.AdminUser, error) {
	var user model.AdminUser
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userAdapter) FindByID(id string) (*model.AdminUser, error) {
	var user model.AdminUser
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userAdapter) Update(user model.AdminUser) error {
	return r.db.Save(&user).Error
}
