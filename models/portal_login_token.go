package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

// portalLoginTokenLength is the size of the random token embedded in a
// self-service portal login link. Shorter-lived than an approval magic
// link (see portalLoginTokenTTL in approvals), so a smaller length is
// fine — it never needs to survive days of guessing attempts.
const portalLoginTokenLength = 32

// ErrPortalLoginTokenInvalid is thrown when a self-service login token
// doesn't match any pending request, has expired, or was already used.
var ErrPortalLoginTokenInvalid = errors.New("Login link is invalid or has expired")

// PortalLoginToken is a single-use magic link for the client portal's
// self-service login ("enter your e-mail, we send you a link"). Kept
// separate from ContractApprover.TokenHash, which is a single mutable
// slot tied to one active ApprovalRequest per approver — reusing it here
// would silently invalidate an in-flight approval link.
type PortalLoginToken struct {
	Id        int64      `json:"id"`
	TenantID  int64      `json:"-" gorm:"column:tenant_id;default:1"`
	Email     string     `json:"email"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func (t *PortalLoginToken) TableName() string { return "portal_login_tokens" }

// ApproverTenantMatch is one (tenant, name) a client's e-mail address is
// registered as a contract approver for.
type ApproverTenantMatch struct {
	TenantID int64  `json:"tenant_id"`
	Name     string `json:"name"`
}

// FindTenantsForApproverEmail returns every tenant the given e-mail is a
// configured contract approver for, deduplicated by tenant. The
// self-service login form doesn't know the caller's tenant up front —
// this is the lookup that resolves it, the same way an approval magic
// link resolves identity from the token rather than trusting client
// input.
func FindTenantsForApproverEmail(email string) ([]ApproverTenantMatch, error) {
	matches := []ApproverTenantMatch{}
	err := db.Table("contract_approvers ca").
		Select("DISTINCT c.tenant_id, ca.name").
		Joins("JOIN contracts c ON c.id = ca.contract_id").
		Where("ca.email = ?", email).
		Scan(&matches).Error
	return matches, err
}

// CreatePortalLoginToken mints a fresh self-service login token for one
// tenant/e-mail pair, storing only its hash. Returns the plaintext token —
// the only time it exists — for the caller to embed in the e-mailed link.
func CreatePortalLoginToken(tenantID int64, email string, ttl time.Duration) (string, error) {
	token, err := generateOpaqueToken(portalLoginTokenLength)
	if err != nil {
		return "", err
	}
	plt := PortalLoginToken{
		TenantID:  tenantID,
		Email:     email,
		TokenHash: hashToken(token),
		ExpiresAt: time.Now().UTC().Add(ttl),
		CreatedAt: time.Now().UTC(),
	}
	if err := db.Save(&plt).Error; err != nil {
		return "", err
	}
	return token, nil
}

// RedeemPortalLoginToken validates a plaintext self-service login token —
// unexpired and not already used — and marks it used, returning the
// tenant and e-mail it was issued for. A token can only ever be redeemed
// once, even if it hasn't expired yet.
func RedeemPortalLoginToken(token string) (PortalLoginToken, error) {
	plt := PortalLoginToken{}
	err := db.Where("token_hash = ?", hashToken(token)).First(&plt).Error
	if err == gorm.ErrRecordNotFound {
		return plt, ErrPortalLoginTokenInvalid
	}
	if err != nil {
		return plt, err
	}
	if plt.UsedAt != nil || time.Now().UTC().After(plt.ExpiresAt) {
		return plt, ErrPortalLoginTokenInvalid
	}
	now := time.Now().UTC()
	plt.UsedAt = &now
	if err := db.Save(&plt).Error; err != nil {
		return plt, err
	}
	return plt, nil
}
