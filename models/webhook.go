package models

import (
	"errors"

	log "github.com/gophish/gophish/logger"
	"github.com/jinzhu/gorm"
)

// Webhook represents the webhook model
type Webhook struct {
	Id       int64  `json:"id" gorm:"column:id; primary_key:yes"`
	TenantID int64  `json:"-" gorm:"column:tenant_id;default:1"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Secret   string `json:"secret"`
	IsActive bool   `json:"is_active"`
}

// ErrURLNotSpecified indicates there was no URL specified
var ErrURLNotSpecified = errors.New("URL can't be empty")

// ErrNameNotSpecified indicates there was no name specified
var ErrNameNotSpecified = errors.New("Name can't be empty")

// GetWebhooks returns the webhooks
func GetWebhooks() ([]Webhook, error) {
	whs := []Webhook{}
	err := db.Find(&whs).Error
	return whs, err
}

// GetActiveWebhooks returns the active webhooks
func GetActiveWebhooks() ([]Webhook, error) {
	whs := []Webhook{}
	err := db.Where("is_active=?", true).Find(&whs).Error
	return whs, err
}

// GetWebhook returns the webhook that the given id corresponds to.
// If no webhook is found, an error is returned.
func GetWebhook(id int64) (Webhook, error) {
	wh := Webhook{}
	err := db.Where("id=?", id).First(&wh).Error
	return wh, err
}

// PostWebhook creates a new webhook in the database.
func PostWebhook(wh *Webhook) error {
	err := wh.Validate()
	if err != nil {
		log.Error(err)
		return err
	}
	err = db.Save(wh).Error
	if err != nil {
		log.Error(err)
	}
	return err
}

// PutWebhook edits an existing webhook in the database.
func PutWebhook(wh *Webhook) error {
	err := wh.Validate()
	if err != nil {
		log.Error(err)
		return err
	}
	err = db.Save(wh).Error
	return err
}

// DeleteWebhook deletes an existing webhook in the database.
// An error is returned if a webhook with the given id isn't found.
func DeleteWebhook(id int64) error {
	err := db.Where("id=?", id).Delete(&Webhook{}).Error
	return err
}

func (wh *Webhook) Validate() error {
	if wh.URL == "" {
		return ErrURLNotSpecified
	}
	if wh.Name == "" {
		return ErrNameNotSpecified
	}
	return nil
}

// GetWebhooksForTenant returns the webhooks owned by the given tenant.
func GetWebhooksForTenant(tenantID int64) ([]Webhook, error) {
	whs := []Webhook{}
	err := withTenantTransaction(tenantID, func(tx *gorm.DB) error {
		return tx.Where("tenant_id=?", tenantID).Find(&whs).Error
	})
	return whs, err
}

// GetActiveWebhooksForTenant returns the active webhooks owned by the given
// tenant. Campaign event delivery must use this instead of the legacy
// global lookup, so an event from one tenant's campaign is never sent to
// another tenant's endpoint.
func GetActiveWebhooksForTenant(tenantID int64) ([]Webhook, error) {
	whs := []Webhook{}
	err := withTenantTransaction(tenantID, func(tx *gorm.DB) error {
		return tx.Where("tenant_id=? AND is_active=?", tenantID, true).Find(&whs).Error
	})
	return whs, err
}

// GetWebhookForTenant returns the webhook scoped to the given tenant.
func GetWebhookForTenant(id, tenantID int64) (Webhook, error) {
	wh := Webhook{}
	err := withTenantTransaction(tenantID, func(tx *gorm.DB) error {
		return tx.Where("id=? AND tenant_id=?", id, tenantID).First(&wh).Error
	})
	return wh, err
}

// PostWebhookForTenant creates a new webhook inside a tenant-bound
// transaction.
func PostWebhookForTenant(wh *Webhook, tenantID int64) error {
	if err := wh.Validate(); err != nil {
		log.Error(err)
		return err
	}
	wh.TenantID = tenantID
	return withTenantTransaction(tenantID, func(tx *gorm.DB) error {
		return tx.Save(wh).Error
	})
}

// PutWebhookForTenant verifies the row belongs to the selected tenant before
// updating it.
func PutWebhookForTenant(wh *Webhook, tenantID int64) error {
	if err := wh.Validate(); err != nil {
		log.Error(err)
		return err
	}
	wh.TenantID = tenantID
	return withTenantTransaction(tenantID, func(tx *gorm.DB) error {
		var existing Webhook
		if err := tx.Where("id=? AND tenant_id=?", wh.Id, tenantID).First(&existing).Error; err != nil {
			return err
		}
		return tx.Where("id=? AND tenant_id=?", wh.Id, tenantID).Save(wh).Error
	})
}

// DeleteWebhookForTenant removes a webhook only from the selected tenant.
func DeleteWebhookForTenant(id, tenantID int64) error {
	return withTenantTransaction(tenantID, func(tx *gorm.DB) error {
		return tx.Where("id=? AND tenant_id=?", id, tenantID).Delete(&Webhook{}).Error
	})
}
