package models

import (
	"errors"
	"net/mail"
	"time"

	log "github.com/gophish/gophish/logger"
	"github.com/jinzhu/gorm"
)

// Template models hold the attributes for an email template to be sent to targets
type Template struct {
	Id             int64        `json:"id" gorm:"column:id; primary_key:yes"`
	TenantID       int64        `json:"-" gorm:"column:tenant_id;default:1"`
	UserId         int64        `json:"-" gorm:"column:user_id"`
	Name           string       `json:"name"`
	EnvelopeSender string       `json:"envelope_sender"`
	Subject        string       `json:"subject"`
	Text           string       `json:"text"`
	HTML           string       `json:"html" gorm:"column:html"`
	ModifiedDate   time.Time    `json:"modified_date"`
	Attachments    []Attachment `json:"attachments"`
}

// ErrTemplateNameNotSpecified is thrown when a template name is not specified
var ErrTemplateNameNotSpecified = errors.New("Template name not specified")

// ErrTemplateMissingParameter is thrown when a needed parameter is not provided
var ErrTemplateMissingParameter = errors.New("Need to specify at least plaintext or HTML content")

// Validate checks the given template to make sure values are appropriate and complete
func (t *Template) Validate() error {
	switch {
	case t.Name == "":
		return ErrTemplateNameNotSpecified
	case t.Text == "" && t.HTML == "":
		return ErrTemplateMissingParameter
	case t.EnvelopeSender != "":
		_, err := mail.ParseAddress(t.EnvelopeSender)
		if err != nil {
			return err
		}
	}
	if err := ValidateTemplate(t.HTML); err != nil {
		return err
	}
	if err := ValidateTemplate(t.Text); err != nil {
		return err
	}
	for _, a := range t.Attachments {
		if err := a.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// GetTemplates returns the templates owned by the given user.
func GetTemplates(uid int64) ([]Template, error) {
	ts := []Template{}
	err := db.Where("user_id=?", uid).Find(&ts).Error
	if err != nil {
		log.Error(err)
		return ts, err
	}
	for i := range ts {
		// Get Attachments
		err = db.Where("template_id=?", ts[i].Id).Find(&ts[i].Attachments).Error
		if err == nil && len(ts[i].Attachments) == 0 {
			ts[i].Attachments = make([]Attachment, 0)
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error(err)
			return ts, err
		}
	}
	return ts, err
}

// GetTemplatesForTenant returns templates only when both the user ownership
// and the selected tenant match. New multitenant handlers must use this
// function instead of the legacy user-only lookup.
func GetTemplatesForTenant(tenantID, uid int64) ([]Template, error) {
	ts := []Template{}
	err := withTenantTransaction(tenantID, func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id=? AND user_id=?", tenantID, uid).Find(&ts).Error; err != nil {
			return err
		}
		for i := range ts {
			err := tx.Where("template_id=?", ts[i].Id).Find(&ts[i].Attachments).Error
			if err == nil && len(ts[i].Attachments) == 0 {
				ts[i].Attachments = make([]Attachment, 0)
			}
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error(err)
		return ts, err
	}
	return ts, err
}

// GetTemplate returns the template, if it exists, specified by the given id and user_id.
func GetTemplate(id int64, uid int64) (Template, error) {
	t := Template{}
	err := db.Where("user_id=? and id=?", uid, id).Find(&t).Error
	if err != nil {
		log.Error(err)
		return t, err
	}

	// Get Attachments
	err = db.Where("template_id=?", t.Id).Find(&t.Attachments).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error(err)
		return t, err
	}
	if err == nil && len(t.Attachments) == 0 {
		t.Attachments = make([]Attachment, 0)
	}
	return t, err
}

// GetTemplateForTenant returns a template only from the selected tenant.
func GetTemplateForTenant(id, tenantID, uid int64) (Template, error) {
	t := Template{}
	err := withTenantTransaction(tenantID, func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id=? AND user_id=? AND id=?", tenantID, uid, id).Find(&t).Error; err != nil {
			return err
		}
		err := tx.Where("template_id=?", t.Id).Find(&t.Attachments).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if err == nil && len(t.Attachments) == 0 {
			t.Attachments = make([]Attachment, 0)
		}
		return nil
	})
	if err != nil {
		log.Error(err)
	}
	return t, err
}

// GetTemplateByName returns the template, if it exists, specified by the given name and user_id.
func GetTemplateByName(n string, uid int64) (Template, error) {
	t := Template{}
	err := db.Where("user_id=? and name=?", uid, n).Find(&t).Error
	if err != nil {
		log.Error(err)
		return t, err
	}

	// Get Attachments
	err = db.Where("template_id=?", t.Id).Find(&t.Attachments).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error(err)
		return t, err
	}
	if err == nil && len(t.Attachments) == 0 {
		t.Attachments = make([]Attachment, 0)
	}
	return t, err
}

// GetTemplateByNameForTenant prevents duplicate-name checks from leaking
// information across tenants.
func GetTemplateByNameForTenant(n string, tenantID, uid int64) (Template, error) {
	t := Template{}
	err := withTenantTransaction(tenantID, func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id=? AND user_id=? AND name=?", tenantID, uid, n).Find(&t).Error; err != nil {
			return err
		}
		err := tx.Where("template_id=?", t.Id).Find(&t.Attachments).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if err == nil && len(t.Attachments) == 0 {
			t.Attachments = make([]Attachment, 0)
		}
		return nil
	})
	if err != nil {
		log.Error(err)
	}
	return t, err
}

// PostTemplate creates a new template in the database.
func PostTemplate(t *Template) error {
	// Insert into the DB
	if err := t.Validate(); err != nil {
		return err
	}
	err := db.Save(t).Error
	if err != nil {
		log.Error(err)
		return err
	}

	// Save every attachment
	for i := range t.Attachments {
		t.Attachments[i].TemplateId = t.Id
		err := db.Save(&t.Attachments[i]).Error
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

// PostTemplateForTenant persists a template inside a tenant-bound
// transaction, ready for PostgreSQL RLS enforcement.
func PostTemplateForTenant(t *Template, tenantID int64) error {
	if err := t.Validate(); err != nil {
		return err
	}
	t.TenantID = tenantID
	return withTenantTransaction(tenantID, func(tx *gorm.DB) error {
		if err := tx.Save(t).Error; err != nil {
			return err
		}
		for i := range t.Attachments {
			t.Attachments[i].TemplateId = t.Id
			if err := tx.Save(&t.Attachments[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// PutTemplate edits an existing template in the database.
// Per the PUT Method RFC, it presumes all data for a template is provided.
func PutTemplate(t *Template) error {
	if err := t.Validate(); err != nil {
		return err
	}
	// Delete all attachments, and replace with new ones
	err := db.Where("template_id=?", t.Id).Delete(&Attachment{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error(err)
		return err
	}
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	for i := range t.Attachments {
		t.Attachments[i].TemplateId = t.Id
		err := db.Save(&t.Attachments[i]).Error
		if err != nil {
			log.Error(err)
			return err
		}
	}

	// Save final template
	err = db.Where("id=?", t.Id).Save(t).Error
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// PutTemplateForTenant verifies the current row belongs to the selected
// tenant before updating it. PostgreSQL RLS will provide the database-level
// backstop in the following Sprint 04 increment.
func PutTemplateForTenant(t *Template, tenantID, uid int64) error {
	if err := t.Validate(); err != nil {
		return err
	}
	t.TenantID = tenantID
	t.UserId = uid
	return withTenantTransaction(tenantID, func(tx *gorm.DB) error {
		var existing Template
		if err := tx.Where("id=? AND tenant_id=? AND user_id=?", t.Id, tenantID, uid).First(&existing).Error; err != nil {
			return err
		}
		if err := tx.Where("template_id=?", t.Id).Delete(&Attachment{}).Error; err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		for i := range t.Attachments {
			t.Attachments[i].TemplateId = t.Id
			if err := tx.Save(&t.Attachments[i]).Error; err != nil {
				return err
			}
		}
		return tx.Where("id=? AND tenant_id=? AND user_id=?", t.Id, tenantID, uid).Save(t).Error
	})
}

// DeleteTemplate deletes an existing template in the database.
// An error is returned if a template with the given user id and template id is not found.
func DeleteTemplate(id int64, uid int64) error {
	// Delete attachments
	err := db.Where("template_id=?", id).Delete(&Attachment{}).Error
	if err != nil {
		log.Error(err)
		return err
	}

	// Finally, delete the template itself
	err = db.Where("user_id=?", uid).Delete(Template{Id: id}).Error
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// DeleteTemplateForTenant removes a template only from the selected tenant.
func DeleteTemplateForTenant(id, tenantID, uid int64) error {
	return withTenantTransaction(tenantID, func(tx *gorm.DB) error {
		return tx.Where("id=? AND tenant_id=? AND user_id=?", id, tenantID, uid).Delete(&Template{}).Error
	})
}

// DeleteTemplates deletes multiple templates in the database.
// It verifies that each template belongs to the specified user before deletion.
func DeleteTemplates(ids []int64, uid int64) error {
	// Start a transaction
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, id := range ids {
		// Verify the template belongs to this user
		t := Template{}
		err := tx.Where("user_id=? and id=?", uid, id).First(&t).Error
		if err != nil {
			tx.Rollback()
			log.Error(err)
			return err
		}

		// Delete attachments
		err = tx.Where("template_id=?", id).Delete(&Attachment{}).Error
		if err != nil {
			tx.Rollback()
			log.Error(err)
			return err
		}

		// Delete the template
		err = tx.Where("user_id=?", uid).Delete(Template{Id: id}).Error
		if err != nil {
			tx.Rollback()
			log.Error(err)
			return err
		}
	}

	return tx.Commit().Error
}

// DeleteTemplatesForTenant is the bulk-delete equivalent scoped to one
// tenant. It intentionally does not report whether IDs existed elsewhere.
func DeleteTemplatesForTenant(ids []int64, tenantID, uid int64) error {
	return withTenantTransaction(tenantID, func(tx *gorm.DB) error {
		return tx.Where("id IN (?) AND tenant_id=? AND user_id=?", ids, tenantID, uid).Delete(&Template{}).Error
	})
}
