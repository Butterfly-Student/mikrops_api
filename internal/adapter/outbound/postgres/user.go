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

func (r *userAdapter) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userAdapter) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userAdapter) FindByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userAdapter) Update(user model.User) error {
	return r.db.Save(&user).Error
}
