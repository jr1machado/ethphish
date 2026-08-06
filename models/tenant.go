package models

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrTenantSlugRequired  = errors.New("tenant slug is required")
	ErrTenantNameRequired  = errors.New("tenant name is required")
	ErrCompanyNameRequired = errors.New("company name is required")
)

// Tenant is the top-level isolation boundary for an EthPhish installation.
type Tenant struct {
	ID        int64     `json:"id" gorm:"column:id;primary_key"`
	Slug      string    `json:"slug" gorm:"column:slug"`
	Name      string    `json:"name" gorm:"column:name"`
	Active    bool      `json:"active" gorm:"column:active"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}

func (Tenant) TableName() string { return "tenants" }

func (t *Tenant) Validate() error {
	t.Slug = strings.ToLower(strings.TrimSpace(t.Slug))
	t.Name = strings.TrimSpace(t.Name)
	if t.Slug == "" {
		return ErrTenantSlugRequired
	}
	if t.Name == "" {
		return ErrTenantNameRequired
	}
	return nil
}

func PostTenant(t *Tenant) error {
	if err := t.Validate(); err != nil {
		return err
	}
	return db.Create(t).Error
}
func GetTenant(id int64) (Tenant, error) {
	var t Tenant
	return t, db.Where("id = ?", id).First(&t).Error
}
func GetTenantBySlug(slug string) (Tenant, error) {
	var t Tenant
	return t, db.Where("slug = ?", strings.ToLower(strings.TrimSpace(slug))).First(&t).Error
}

// Company is an organization boundary within one tenant.
type Company struct {
	ID        int64     `json:"id" gorm:"column:id;primary_key"`
	TenantID  int64     `json:"tenant_id" gorm:"column:tenant_id"`
	Name      string    `json:"name" gorm:"column:name"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}

func (Company) TableName() string { return "companies" }
func PostCompany(c *Company) error {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return ErrCompanyNameRequired
	}
	return db.Create(c).Error
}

// TenantUser grants one administrative user an explicit tenant/company scope.
type TenantUser struct {
	TenantID  int64     `json:"tenant_id" gorm:"column:tenant_id;primary_key"`
	UserID    int64     `json:"user_id" gorm:"column:user_id;primary_key"`
	CompanyID *int64    `json:"company_id" gorm:"column:company_id"`
	Role      string    `json:"role" gorm:"column:role"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}

func (TenantUser) TableName() string { return "tenant_users" }
func GrantTenantUser(grant *TenantUser) error {
	if grant.Role == "" {
		grant.Role = "member"
	}
	return db.Create(grant).Error
}
func GetTenantUser(tenantID, userID int64) (TenantUser, error) {
	var grant TenantUser
	return grant, db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&grant).Error
}

// GetTenantUsers returns all active tenant grants for a user. A user can have
// more than one grant, but the request layer must make the selected tenant
// explicit whenever there is more than one.
func GetTenantUsers(userID int64) ([]TenantUser, error) {
	var grants []TenantUser
	err := db.Joins("JOIN tenants ON tenants.id = tenant_users.tenant_id").
		Where("tenant_users.user_id = ? AND tenants.active = ?", userID, true).
		Order("tenant_users.tenant_id ASC").
		Find(&grants).Error
	return grants, err
}
