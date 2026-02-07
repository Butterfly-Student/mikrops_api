package gormutil

import "gorm.io/gorm"

func TenantScope(tenantID string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}
