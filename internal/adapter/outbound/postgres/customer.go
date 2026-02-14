package postgres_outbound_adapter

import (
	"errors"
	"time"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"gorm.io/gorm"
)

const tableCustomer = "customers"

type CustomerAdapter struct {
	db *gorm.DB
}

func NewCustomerAdapter(db *gorm.DB) outbound_port.CustomerDatabasePort {
	return &CustomerAdapter{db: db}
}

func (a *CustomerAdapter) Create(customer *model.Customer) error {
	model.CustomerPrepare(customer)
	return a.db.Create(customer).Error
}

func (a *CustomerAdapter) FindByID(id string) (*model.Customer, error) {
	var customer model.Customer
	err := a.db.Preload("Router").Preload("Profile").Where("id = ?", id).First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("customer not found")
		}
		return nil, err
	}
	return &customer, nil
}

func (a *CustomerAdapter) FindByCustomerCode(code string) (*model.Customer, error) {
	var customer model.Customer
	err := a.db.Preload("Router").Preload("Profile").Where("customer_code = ?", code).First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("customer not found")
		}
		return nil, err
	}
	return &customer, nil
}

func (a *CustomerAdapter) FindByRouterID(routerID string) ([]model.Customer, error) {
	var customers []model.Customer
	err := a.db.Preload("Profile").Where("router_id = ?", routerID).Find(&customers).Error
	return customers, err
}

func (a *CustomerAdapter) FindByProfileID(profileID string) ([]model.Customer, error) {
	var customers []model.Customer
	err := a.db.Where("profile_id = ?", profileID).Find(&customers).Error
	return customers, err
}

func (a *CustomerAdapter) FindAll() ([]model.Customer, error) {
	var customers []model.Customer
	err := a.db.Preload("Router").Preload("Profile").Order("created_at DESC").Find(&customers).Error
	return customers, err
}

func (a *CustomerAdapter) Find(filter model.CustomerFilter) ([]model.Customer, error) {
	var customers []model.Customer
	query := a.db.Model(&model.Customer{}).Preload("Router").Preload("Profile")

	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.CustomerCodes) > 0 {
		query = query.Where("customer_code IN ?", filter.CustomerCodes)
	}
	if len(filter.RouterIDs) > 0 {
		query = query.Where("router_id IN ?", filter.RouterIDs)
	}
	if len(filter.ProfileIDs) > 0 {
		query = query.Where("profile_id IN ?", filter.ProfileIDs)
	}
	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}
	if filter.Email != nil {
		query = query.Where("email = ?", *filter.Email)
	}
	if filter.Phone != nil {
		query = query.Where("phone = ?", *filter.Phone)
	}
	if filter.Search != nil {
		search := "%" + *filter.Search + "%"
		query = query.Where("full_name ILIKE ? OR phone ILIKE ? OR email ILIKE ?", search, search, search)
	}

	query = query.Where("deleted_at IS NULL")

	if filter.IsActive != nil {
		if *filter.IsActive {
			query = query.Where("status IN ?", []string{"active", "isolated"})
		} else {
			query = query.Where("status IN ?", []string{"pending", "suspended", "terminated"})
		}
	}

	if filter.IsExpired != nil {
		if *filter.IsExpired {
			query = query.Where("expiry_date < ?", time.Now())
		} else {
			query = query.Where("expiry_date >= ?", time.Now())
		}
	}

	if filter.WillExpireSoon != nil {
		soonDate := time.Now().AddDate(0, 0, *filter.WillExpireSoon)
		query = query.Where("expiry_date BETWEEN ? AND ?", time.Now(), soonDate)
	}

	err := query.Order("created_at DESC").Find(&customers).Error
	return customers, err
}

func (a *CustomerAdapter) FindExpired() ([]model.Customer, error) {
	var customers []model.Customer
	err := a.db.Where("expiry_date < ? AND status = ?", time.Now(), "active").Find(&customers).Error
	return customers, err
}

func (a *CustomerAdapter) FindExpiringSoon(days int) ([]model.Customer, error) {
	var customers []model.Customer
	soonDate := time.Now().AddDate(0, 0, days)
	err := a.db.Where("expiry_date BETWEEN ? AND ? AND status = ?", time.Now(), soonDate, "active").Find(&customers).Error
	return customers, err
}

func (a *CustomerAdapter) Update(customer *model.Customer) error {
	return a.db.Save(customer).Error
}

func (a *CustomerAdapter) Delete(id string) error {
	return a.db.Delete(&model.Customer{}, "id = ?", id).Error
}

func (a *CustomerAdapter) FindByPppSecretName(name string) (*model.Customer, error) {
	var customer model.Customer
	err := a.db.Where("ppp_secret_name = ?", name).First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("customer not found")
		}
		return nil, err
	}
	return &customer, nil
}
