package models

import (
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"
)

// withTenantTransaction executes database work with a tenant value bound to
// the current transaction. PostgreSQL's SET LOCAL is connection-safe: its
// value disappears at commit or rollback and cannot leak through the pool to
// another request.
func withTenantTransaction(tenantID int64, work func(*gorm.DB) error) error {
	if tenantID <= 0 {
		return fmt.Errorf("tenant ID is required")
	}
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if tx.Dialect().GetName() == "postgres" {
		if err := tx.Exec("SELECT set_config('ethphish.tenant_id', ?, true)", strconv.FormatInt(tenantID, 10)).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := work(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
