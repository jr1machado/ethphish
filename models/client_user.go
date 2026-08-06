package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

// clientSessionTokenLength is the size of the random token used for the
// client-portal session cookie.
const clientSessionTokenLength = 32

// ErrClientSessionInvalid is thrown when a client session cookie doesn't
// match a live, unexpired session.
var ErrClientSessionInvalid = errors.New("Client session is invalid or has expired")

// ClientUser is a lightweight identity for a contract approver. There is no
// password: identity is proven by clicking a magic link sent to a validated
// e-mail address (see ApprovalRequest / ClientSession).
type ClientUser struct {
	Id              int64      `json:"id"`
	TenantID        int64      `json:"-" gorm:"column:tenant_id;default:1"`
	Email           string     `json:"email" sql:"not null"`
	Name            string     `json:"name"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

func (c *ClientUser) TableName() string { return "client_users" }

// ClientSession is a server-side session record for the client approval
// portal, backing the "ethphish_client" cookie. Kept separate from the
// admin session store (middleware.Store) so a client approver never shares
// a cookie namespace with admin auth.
type ClientSession struct {
	Id               int64      `json:"id"`
	ClientUserId     int64      `json:"client_user_id"`
	ClientUser       ClientUser `json:"client_user,omitempty" gorm:"foreignkey:ClientUserId"`
	SessionTokenHash string     `json:"-"`
	ExpiresAt        time.Time  `json:"expires_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

func (c *ClientSession) TableName() string { return "client_sessions" }

// GetOrCreateClientUser finds an existing client user by (tenant, email) or
// creates one, marking the e-mail verified since it was only reachable via
// a magic link sent to that exact address.
func GetOrCreateClientUser(tenantID int64, email, name string) (ClientUser, error) {
	cu := ClientUser{}
	err := db.Where("tenant_id = ? AND email = ?", tenantID, email).First(&cu).Error
	if err == nil {
		return cu, nil
	}
	if err != gorm.ErrRecordNotFound {
		return cu, err
	}
	now := time.Now().UTC()
	cu = ClientUser{
		TenantID:        tenantID,
		Email:           email,
		Name:            name,
		EmailVerifiedAt: &now,
		CreatedAt:       now,
	}
	if err := db.Save(&cu).Error; err != nil {
		return ClientUser{}, err
	}
	return cu, nil
}

// CreateClientSession issues a new session for a client user, returning the
// stored session and the plaintext token to set as the session cookie.
func CreateClientSession(clientUserID int64, ttl time.Duration) (ClientSession, string, error) {
	token, err := generateOpaqueToken(clientSessionTokenLength)
	if err != nil {
		return ClientSession{}, "", err
	}
	cs := ClientSession{
		ClientUserId:     clientUserID,
		SessionTokenHash: hashToken(token),
		ExpiresAt:        time.Now().UTC().Add(ttl),
		CreatedAt:        time.Now().UTC(),
	}
	if err := db.Save(&cs).Error; err != nil {
		return ClientSession{}, "", err
	}
	return cs, token, nil
}

// GetClientSessionByToken resolves a session cookie's plaintext token to a
// live, unexpired session and its client user.
func GetClientSessionByToken(token string) (ClientSession, error) {
	cs := ClientSession{}
	err := db.Preload("ClientUser").Where("session_token_hash = ?", hashToken(token)).First(&cs).Error
	if err == gorm.ErrRecordNotFound {
		return cs, ErrClientSessionInvalid
	}
	if err != nil {
		return cs, err
	}
	if time.Now().UTC().After(cs.ExpiresAt) {
		return cs, ErrClientSessionInvalid
	}
	return cs, nil
}

// DeleteClientSession destroys a session (client logout).
func DeleteClientSession(token string) error {
	return db.Where("session_token_hash = ?", hashToken(token)).Delete(ClientSession{}).Error
}

// GetPendingApprovalsForClientUser lists the approval requests awaiting
// this client user, based on their verified e-mail matching a contract's
// configured approvers.
func GetPendingApprovalsForClientUser(clientUserID int64) ([]ApprovalRequestSummary, error) {
	cu := ClientUser{}
	if err := db.Where("id = ?", clientUserID).First(&cu).Error; err != nil {
		return nil, err
	}
	rows := []ApprovalRequestSummary{}
	err := db.Table("approval_requests ar").
		Select(`ar.id, c.id as contract_id, c.name as contract_name, c.client_name,
			cv.version_number, ar.status, ar.token_expires_at, ar.requested_at,
			ar.decided_at, ar.last_reminder_sent_at`).
		Joins("JOIN contract_versions cv ON cv.id = ar.contract_version_id").
		Joins("JOIN contracts c ON c.id = cv.contract_id").
		Joins("JOIN contract_approvers ca ON ca.contract_id = c.id").
		Where("ca.email = ? AND c.tenant_id = ?", cu.Email, cu.TenantID).
		Order("ar.requested_at desc").
		Scan(&rows).Error
	return rows, err
}
