package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

// Approval request statuses
const (
	ApprovalPending          string = "pending"
	ApprovalApproved         string = "approved"
	ApprovalRejected         string = "rejected"
	ApprovalChangesRequested string = "changes_requested"
	ApprovalExpired          string = "expired"
)

// magicLinkTokenLength is the size of the random token embedded in an
// approval e-mail link.
const magicLinkTokenLength = 32

// ErrApprovalNotFound is thrown when an approval request can't be found.
var ErrApprovalNotFound = errors.New("Approval request not found")

// ErrApprovalTokenInvalid is thrown when a magic-link token doesn't match
// any pending approval request, or has expired.
var ErrApprovalTokenInvalid = errors.New("Approval link is invalid or has expired")

// ErrApprovalAlreadyDecided is thrown when a client tries to decide an
// approval request that is no longer pending.
var ErrApprovalAlreadyDecided = errors.New("This approval request has already been decided")

// ApprovalRequest represents one round of client sign-off for an exact
// ContractVersion, optionally tied to the campaign it's gating.
type ApprovalRequest struct {
	Id                 int64             `json:"id"`
	ContractVersionId  int64             `json:"contract_version_id"`
	CampaignId         *int64            `json:"campaign_id,omitempty"`
	Status             string            `json:"status"`
	MagicLinkTokenHash string            `json:"-"`
	TokenExpiresAt     time.Time         `json:"token_expires_at"`
	RequestedAt        time.Time         `json:"requested_at"`
	DecidedAt          *time.Time        `json:"decided_at,omitempty"`
	DecidedBy          *int64            `json:"decided_by,omitempty"`
	LastReminderSentAt *time.Time        `json:"last_reminder_sent_at,omitempty"`
	Comments           []ApprovalComment `json:"comments,omitempty" gorm:"foreignkey:ApprovalRequestId"`
}

func (a *ApprovalRequest) TableName() string { return "approval_requests" }

// ApprovalComment is a single message on an approval request's thread,
// left by either the admin side or the client side.
type ApprovalComment struct {
	Id                int64     `json:"id"`
	ApprovalRequestId int64     `json:"approval_request_id"`
	AuthorType        string    `json:"author_type"`
	AuthorId          int64     `json:"author_id"`
	Body              string    `json:"body" sql:"not null"`
	CreatedAt         time.Time `json:"created_at"`
}

// CreateApprovalRequest opens a new approval cycle for a contract version.
// It returns the stored request and the plaintext magic-link token — the
// only time the plaintext token exists; only its hash is persisted.
func CreateApprovalRequest(contractVersionID int64, campaignID *int64, ttl time.Duration) (ApprovalRequest, string, error) {
	token, err := generateOpaqueToken(magicLinkTokenLength)
	if err != nil {
		return ApprovalRequest{}, "", err
	}
	now := time.Now().UTC()
	ar := ApprovalRequest{
		ContractVersionId:  contractVersionID,
		CampaignId:         campaignID,
		Status:             ApprovalPending,
		MagicLinkTokenHash: hashToken(token),
		TokenExpiresAt:     now.Add(ttl),
		RequestedAt:        now,
	}
	if err := db.Save(&ar).Error; err != nil {
		return ApprovalRequest{}, "", err
	}
	return ar, token, nil
}

// GetApprovalRequest returns a single approval request by id.
func GetApprovalRequest(id int64) (ApprovalRequest, error) {
	ar := ApprovalRequest{}
	err := db.Preload("Comments").Where("id = ?", id).First(&ar).Error
	if err == gorm.ErrRecordNotFound {
		return ar, ErrApprovalNotFound
	}
	return ar, err
}

// GetApprovalRequestByToken looks up a pending approval request by its
// plaintext magic-link token, rejecting expired or already-used tokens.
func GetApprovalRequestByToken(token string) (ApprovalRequest, error) {
	ar := ApprovalRequest{}
	err := db.Preload("Comments").
		Where("magic_link_token_hash = ?", hashToken(token)).First(&ar).Error
	if err == gorm.ErrRecordNotFound {
		return ar, ErrApprovalTokenInvalid
	}
	if err != nil {
		return ar, err
	}
	if ar.Status != ApprovalPending || time.Now().UTC().After(ar.TokenExpiresAt) {
		return ar, ErrApprovalTokenInvalid
	}
	return ar, nil
}

// approvalRequestSummary is the joined shape the Central de Aprovações
// table needs — one row per approval request with its contract context.
type ApprovalRequestSummary struct {
	Id                 int64      `json:"id"`
	ContractId         int64      `json:"contract_id"`
	ContractName       string     `json:"contract_name"`
	ClientName         string     `json:"client_name"`
	VersionNumber      int        `json:"version_number"`
	Status             string     `json:"status"`
	TokenExpiresAt     time.Time  `json:"token_expires_at"`
	RequestedAt        time.Time  `json:"requested_at"`
	DecidedAt          *time.Time `json:"decided_at,omitempty"`
	LastReminderSentAt *time.Time `json:"last_reminder_sent_at,omitempty"`
}

// GetApprovalRequestsForTenant returns every approval request belonging to
// the tenant's contracts, joined with contract/version context for display
// in the Central de Aprovações.
func GetApprovalRequestsForTenant(tenantID int64) ([]ApprovalRequestSummary, error) {
	rows := []ApprovalRequestSummary{}
	err := db.Table("approval_requests ar").
		Select(`ar.id, c.id as contract_id, c.name as contract_name, c.client_name,
			cv.version_number, ar.status, ar.token_expires_at, ar.requested_at,
			ar.decided_at, ar.last_reminder_sent_at`).
		Joins("JOIN contract_versions cv ON cv.id = ar.contract_version_id").
		Joins("JOIN contracts c ON c.id = cv.contract_id").
		Where("c.tenant_id = ?", tenantID).
		Order("ar.requested_at desc").
		Scan(&rows).Error
	return rows, err
}

// GetContractForApprovalRequest resolves an approval request back to the
// contract it belongs to, via its contract version.
func GetContractForApprovalRequest(ar ApprovalRequest) (Contract, error) {
	v, err := GetContractVersion(ar.ContractVersionId)
	if err != nil {
		return Contract{}, err
	}
	return GetContractByID(v.ContractId)
}

// AddApprovalComment appends a comment to an approval request's thread.
func AddApprovalComment(approvalRequestID int64, authorType string, authorID int64, body string) error {
	c := ApprovalComment{
		ApprovalRequestId: approvalRequestID,
		AuthorType:        authorType,
		AuthorId:          authorID,
		Body:              body,
		CreatedAt:         time.Now().UTC(),
	}
	return db.Save(&c).Error
}

// DecideApproval records a client's decision (approved/rejected/
// changes_requested) on a pending approval request, optionally with a
// comment.
func DecideApproval(id int64, decidedBy int64, status string, comment string) error {
	ar, err := GetApprovalRequest(id)
	if err != nil {
		return err
	}
	if ar.Status != ApprovalPending {
		return ErrApprovalAlreadyDecided
	}
	now := time.Now().UTC()
	ar.Status = status
	ar.DecidedAt = &now
	ar.DecidedBy = &decidedBy
	if err := db.Save(&ar).Error; err != nil {
		return err
	}
	if comment != "" {
		return AddApprovalComment(id, "client", decidedBy, comment)
	}
	return nil
}

// ExpirePendingApprovals marks every pending approval request whose token
// has expired as expired, returning how many were updated. Called by the
// reminder/expiration cron.
func ExpirePendingApprovals() (int64, error) {
	res := db.Model(&ApprovalRequest{}).
		Where("status = ? AND token_expires_at < ?", ApprovalPending, time.Now().UTC()).
		Update("status", ApprovalExpired)
	return res.RowsAffected, res.Error
}

// PendingApprovalsNeedingReminder returns pending approval requests older
// than the reminder interval that haven't been reminded within it.
func PendingApprovalsNeedingReminder(interval time.Duration) ([]ApprovalRequest, error) {
	cutoff := time.Now().UTC().Add(-interval)
	ars := []ApprovalRequest{}
	err := db.Where("status = ? AND requested_at < ? AND (last_reminder_sent_at IS NULL OR last_reminder_sent_at < ?)",
		ApprovalPending, cutoff, cutoff).Find(&ars).Error
	return ars, err
}

// MarkReminderSent stamps the last-reminder timestamp for an approval
// request.
func MarkReminderSent(id int64) error {
	now := time.Now().UTC()
	return db.Model(&ApprovalRequest{}).Where("id = ?", id).Update("last_reminder_sent_at", now).Error
}
