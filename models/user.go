package models

import (
	"errors"
	"time"

	log "github.com/gophish/gophish/logger"
)

// ErrModifyingOnlyAdmin occurs when there is an attempt to modify the only
// user account with the Admin role in such a way that there will be no user
// accounts left in Gophish with that role.
var ErrModifyingOnlyAdmin = errors.New("Cannot remove the only administrator")

// User represents the user model for gophish.
type User struct {
	Id                     int64     `json:"id"`
	Username               string    `json:"username" sql:"not null;unique"`
	Hash                   string    `json:"-"`
	ApiKey                 string    `json:"api_key" sql:"not null;unique"`
	Role                   Role      `json:"role" gorm:"association_autoupdate:false;association_autocreate:false"`
	RoleID                 int64     `json:"-"`
	PasswordChangeRequired bool      `json:"password_change_required"`
	AccountLocked          bool      `json:"account_locked"`
	LastLogin              time.Time `json:"last_login"`
}

// GetUser returns the user that the given id corresponds to. If no user is found, an
// error is thrown.
func GetUser(id int64) (User, error) {
	u := User{}
	err := db.Preload("Role").Where("id=?", id).First(&u).Error
	return u, err
}

// GetUsers returns the users registered in Gophish
func GetUsers() ([]User, error) {
	us := []User{}
	err := db.Preload("Role").Find(&us).Error
	return us, err
}

// GetUserByAPIKey returns the user that the given API Key corresponds to. If no user is found, an
// error is thrown.
func GetUserByAPIKey(key string) (User, error) {
	u := User{}
	err := db.Preload("Role").Where("api_key = ?", key).First(&u).Error
	return u, err
}

// GetUserByUsername returns the user that the given username corresponds to. If no user is found, an
// error is thrown.
func GetUserByUsername(username string) (User, error) {
	u := User{}
	err := db.Preload("Role").Where("username = ?", username).First(&u).Error
	return u, err
}

// PutUser updates the given user. A newly created legacy user is granted the
// default tenant so installations upgraded from the single-tenant model keep
// working. Multitenant API flows must use PutUserForTenant instead.
func PutUser(u *User) error {
	if u.Id != 0 {
		return db.Save(u).Error
	}
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Save(u).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Create(&TenantUser{TenantID: 1, UserID: u.Id, Role: "member"}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// PutUserForTenant creates a user and its first tenant grant atomically. It
// is used by multitenant API flows to avoid a transient or unintended default
// tenant membership.
func PutUserForTenant(u *User, tenantID int64, companyID *int64, role string) error {
	if u.Id != 0 {
		return db.Save(u).Error
	}
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Save(u).Error; err != nil {
		tx.Rollback()
		return err
	}
	if role == "" {
		role = "member"
	}
	if err := tx.Create(&TenantUser{TenantID: tenantID, UserID: u.Id, CompanyID: companyID, Role: role}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// EnsureEnoughAdmins ensures that there is more than one user account in
// Gophish with the Admin role. This function is meant to be called before
// modifying a user account with the Admin role in a non-revokable way.
func EnsureEnoughAdmins() error {
	role, err := GetRoleBySlug(RoleAdmin)
	if err != nil {
		return err
	}
	var adminCount int
	err = db.Model(&User{}).Where("role_id=?", role.ID).Count(&adminCount).Error
	if err != nil {
		return err
	}
	if adminCount == 1 {
		return ErrModifyingOnlyAdmin
	}
	return nil
}

// DeleteUser deletes the given user. To ensure that there is always at least
// one user account with the Admin role, this function will refuse to delete
// the last Admin.
func DeleteUser(id int64) error {
	existing, err := GetUser(id)
	if err != nil {
		return err
	}
	// If the user is an admin, we need to verify that it's not the last one.
	if existing.Role.Slug == RoleAdmin {
		err = EnsureEnoughAdmins()
		if err != nil {
			return err
		}
	}
	campaigns, err := GetCampaigns(id)
	if err != nil {
		return err
	}
	// Delete the campaigns
	log.Infof("Deleting campaigns for user ID %d", id)
	for _, campaign := range campaigns {
		err = DeleteCampaign(campaign.Id)
		if err != nil {
			return err
		}
	}
	log.Infof("Deleting pages for user ID %d", id)
	// Delete the landing pages
	pages, err := GetPages(id)
	if err != nil {
		return err
	}
	for _, page := range pages {
		err = DeletePage(page.Id, id)
		if err != nil {
			return err
		}
	}
	// Delete the templates
	log.Infof("Deleting templates for user ID %d", id)
	templates, err := GetTemplates(id)
	if err != nil {
		return err
	}
	for _, template := range templates {
		err = DeleteTemplate(template.Id, id)
		if err != nil {
			return err
		}
	}
	// Delete the groups
	log.Infof("Deleting groups for user ID %d", id)
	groups, err := GetGroups(id)
	if err != nil {
		return err
	}
	for _, group := range groups {
		err = DeleteGroup(&group)
		if err != nil {
			return err
		}
	}
	// Delete the sending profiles
	log.Infof("Deleting sending profiles for user ID %d", id)
	profiles, err := GetSMTPs(id)
	if err != nil {
		return err
	}
	for _, profile := range profiles {
		err = DeleteSMTP(profile.Id, id)
		if err != nil {
			return err
		}
	}
	// Finally, delete the user
	err = db.Where("id=?", id).Delete(&User{}).Error
	return err
}
