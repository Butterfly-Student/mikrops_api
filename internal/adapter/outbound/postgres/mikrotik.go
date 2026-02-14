package postgres_outbound_adapter

import (
	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"

	"gorm.io/gorm"
)

type mikrotikAdapter struct {
	db *gorm.DB
}

func NewMikrotikAdapter(db *gorm.DB) outbound_port.MikrotikDatabasePort {
	return &mikrotikAdapter{db: db}
}

func (a *mikrotikAdapter) Create(router *model.MikrotikRouter) error {
	return a.db.Create(router).Error
}

func (a *mikrotikAdapter) FindByID(id string) (*model.MikrotikRouter, error) {
	var router model.MikrotikRouter
	err := a.db.Where("id = ?", id).First(&router).Error
	if err != nil {
		return nil, err
	}
	return &router, nil
}

func (a *mikrotikAdapter) FindAll() ([]model.MikrotikRouter, error) {
	var routers []model.MikrotikRouter
	err := a.db.Find(&routers).Error
	return routers, err
}

func (a *mikrotikAdapter) Update(router *model.MikrotikRouter) error {
	return a.db.Save(router).Error
}

func (a *mikrotikAdapter) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&model.MikrotikRouter{}).Error
}
