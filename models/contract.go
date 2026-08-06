package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

// Contract statuses
const (
	ContractDraft    string = "draft"
	ContractActive   string = "active"
	ContractArchived string = "archived"
)

// ErrContractNameNotSpecified is thrown when a contract is submitted without a name
var ErrContractNameNotSpecified = errors.New("Contract name not specified")

// ErrContractClientNotSpecified is thrown when a contract is submitted without a client name
var ErrContractClientNotSpecified = errors.New("Client name not specified")

// ErrContractNotFound is thrown when a contract can't be found
var ErrContractNotFound = errors.New("Contract not found")

// ErrApprovalRequired is thrown when a campaign linked to a contract is
// launched without a current, approved contract version.
var ErrApprovalRequired = errors.New("Campaign requires an approved contract before it can run")

// Contract represents a client contract that a campaign's scope can be
// bound to.
type Contract struct {
	Id         int64              `json:"id"`
	TenantID   int64              `json:"-" gorm:"column:tenant_id;default:1"`
	Name       string             `json:"name" sql:"not null"`
	ClientName string             `json:"client_name" sql:"not null"`
	Status     string             `json:"status" sql:"not null;default:'draft'"`
	CreatedBy  int64              `json:"-" gorm:"column:created_by"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
	Versions   []ContractVersion  `json:"versions,omitempty" gorm:"foreignkey:ContractId"`
	Approvers  []ContractApprover `json:"approvers,omitempty" gorm:"foreignkey:ContractId"`
}

// ContractVersion represents one uploaded revision of a contract document.
// Approvals bind to an exact ContractVersion, so a new upload always starts
// a fresh approval cycle.
type ContractVersion struct {
	Id               int64     `json:"id"`
	ContractId       int64     `json:"contract_id"`
	VersionNumber    int       `json:"version_number"`
	FilePath         string    `json:"-" gorm:"column:file_path"`
	FileName         string    `json:"file_name"`
	ScopeDescription string    `json:"scope_description"`
	UploadedBy       int64     `json:"-"`
	UploadedAt       time.Time `json:"uploaded_at"`
	Notes            string    `json:"notes"`
}

// ContractApprover is a named responsible party who receives approval
// requests for a contract.
type ContractApprover struct {
	Id         int64     `json:"id"`
	ContractId int64     `json:"contract_id"`
	Name       string    `json:"name" sql:"not null"`
	Email      string    `json:"email" sql:"not null"`
	CreatedAt  time.Time `json:"created_at"`

	// Magic-link token for this specific approver. Identity on redemption
	// is derived solely from this token — never from client-supplied
	// input — so one approver can't be impersonated by another who has a
	// copy of a *different* approver's link. See IssueApproverToken /
	// GetApproverByToken.
	TokenHash               string     `json:"-"`
	TokenExpiresAt          *time.Time `json:"-"`
	ActiveApprovalRequestId *int64     `json:"-"`
}

func (c *Contract) TableName() string { return "contracts" }

// Validate checks that a contract has the minimum required fields.
func (c *Contract) Validate() error {
	switch {
	case c.Name == "":
		return ErrContractNameNotSpecified
	case c.ClientName == "":
		return ErrContractClientNotSpecified
	}
	return nil
}

// GetContractsForTenant returns every contract belonging to the tenant,
// preloaded with its versions and approvers — the contracts UI reuses this
// list response to fill in the edit modal, so omitting these would leave
// the "Versions"/"Approvers" tables (and the "Request Approval" button)
// empty even though the data exists.
func GetContractsForTenant(tenantID int64) ([]Contract, error) {
	cs := []Contract{}
	err := db.Preload("Versions").Preload("Approvers").
		Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&cs).Error
	return cs, err
}

// GetContractForTenant returns a single tenant-scoped contract, preloaded
// with its versions and approvers.
func GetContractForTenant(id, tenantID int64) (Contract, error) {
	c := Contract{}
	err := db.Preload("Versions").Preload("Approvers").
		Where("id = ? AND tenant_id = ?", id, tenantID).First(&c).Error
	if err == gorm.ErrRecordNotFound {
		return c, ErrContractNotFound
	}
	return c, err
}

// PostContractForTenant creates a new contract scoped to the tenant.
func PostContractForTenant(c *Contract, tenantID, uid int64) error {
	if err := c.Validate(); err != nil {
		return err
	}
	c.TenantID = tenantID
	c.CreatedBy = uid
	if c.Status == "" {
		c.Status = ContractDraft
	}
	c.CreatedAt = time.Now().UTC()
	c.UpdatedAt = c.CreatedAt
	return db.Save(c).Error
}

// PutContractForTenant updates an existing tenant-scoped contract's mutable
// fields (name, client name, status).
func PutContractForTenant(c *Contract, tenantID int64) error {
	if err := c.Validate(); err != nil {
		return err
	}
	existing, err := GetContractForTenant(c.Id, tenantID)
	if err != nil {
		return err
	}
	existing.Name = c.Name
	existing.ClientName = c.ClientName
	if c.Status != "" {
		existing.Status = c.Status
	}
	existing.UpdatedAt = time.Now().UTC()
	return db.Save(&existing).Error
}

// DeleteContractForTenant removes a contract and its versions/approvers.
func DeleteContractForTenant(id, tenantID int64) error {
	if _, err := GetContractForTenant(id, tenantID); err != nil {
		return err
	}
	if err := db.Where("contract_id = ?", id).Delete(ContractVersion{}).Error; err != nil {
		return err
	}
	if err := db.Where("contract_id = ?", id).Delete(ContractApprover{}).Error; err != nil {
		return err
	}
	return db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(Contract{}).Error
}

// GetContractByID returns a single contract by id, without tenant scoping —
// used internally by the approval workflow, which already scopes through
// the approval request/contract version chain.
func GetContractByID(id int64) (Contract, error) {
	c := Contract{}
	err := db.Where("id = ?", id).First(&c).Error
	if err == gorm.ErrRecordNotFound {
		return c, ErrContractNotFound
	}
	return c, err
}

// AddContractVersion stores a new document version for a contract,
// auto-incrementing the version number. The caller is responsible for
// writing the file to disk (via reports.FileManager) before calling this
// and passing the resulting FilePath.
func AddContractVersion(contractID int64, v *ContractVersion, uid int64) (ContractVersion, error) {
	latest, err := GetLatestContractVersion(contractID)
	switch err {
	case nil:
		v.VersionNumber = latest.VersionNumber + 1
	case ErrContractVersionNotFound:
		v.VersionNumber = 1
	default:
		return ContractVersion{}, err
	}
	v.ContractId = contractID
	v.UploadedBy = uid
	v.UploadedAt = time.Now().UTC()
	if err := db.Save(v).Error; err != nil {
		return ContractVersion{}, err
	}
	return *v, nil
}

// ErrContractVersionNotFound is thrown when a contract has no versions yet.
var ErrContractVersionNotFound = errors.New("Contract has no versions")

// GetContractVersions returns every version of a contract, newest first.
func GetContractVersions(contractID int64) ([]ContractVersion, error) {
	vs := []ContractVersion{}
	err := db.Where("contract_id = ?", contractID).Order("version_number desc").Find(&vs).Error
	return vs, err
}

// GetLatestContractVersion returns the most recent version of a contract.
func GetLatestContractVersion(contractID int64) (ContractVersion, error) {
	v := ContractVersion{}
	err := db.Where("contract_id = ?", contractID).Order("version_number desc").First(&v).Error
	if err == gorm.ErrRecordNotFound {
		return v, ErrContractVersionNotFound
	}
	return v, err
}

// GetContractVersion returns a single contract version by id.
func GetContractVersion(id int64) (ContractVersion, error) {
	v := ContractVersion{}
	err := db.Where("id = ?", id).First(&v).Error
	if err == gorm.ErrRecordNotFound {
		return v, ErrContractVersionNotFound
	}
	return v, err
}

// GetContractVersionForTenant returns a contract version only if it
// belongs to a contract owned by the given tenant, for endpoints (like
// document download) that take a version id directly without an
// accompanying contract id in the URL.
func GetContractVersionForTenant(id, tenantID int64) (ContractVersion, error) {
	v, err := GetContractVersion(id)
	if err != nil {
		return v, err
	}
	if _, err := GetContractForTenant(v.ContractId, tenantID); err != nil {
		return ContractVersion{}, ErrContractVersionNotFound
	}
	return v, nil
}

// AddContractApprover adds a named responsible party to a contract.
func AddContractApprover(a *ContractApprover) error {
	a.CreatedAt = time.Now().UTC()
	return db.Save(a).Error
}

// GetContractApprovers returns every approver configured for a contract.
func GetContractApprovers(contractID int64) ([]ContractApprover, error) {
	as := []ContractApprover{}
	err := db.Where("contract_id = ?", contractID).Find(&as).Error
	return as, err
}

// DeleteContractApprover removes a single approver from a contract.
func DeleteContractApprover(id, contractID int64) error {
	return db.Where("id = ? AND contract_id = ?", id, contractID).Delete(ContractApprover{}).Error
}

// IsCampaignApproved re-checks contract approval at send time (as opposed
// to at creation time), so a contract version re-upload that happened while
// a campaign sat queued still blocks it from actually going out.
func IsCampaignApproved(contractID *int64) (bool, error) {
	return campaignApprovalOK(contractID)
}

// ErrApproverTokenInvalid is thrown when a magic-link token doesn't match
// any approver, or has expired, or its approval request is no longer
// pending.
var ErrApproverTokenInvalid = errors.New("Approval link is invalid or has expired")

// IssueApproverToken mints a fresh magic-link token for one specific
// approver on one specific approval request, storing only its hash.
// Redeeming the returned plaintext token (via GetApproverByToken) is the
// only way the client portal learns who is signing in — the approver's
// e-mail is never taken from client-supplied input.
func IssueApproverToken(approverID, approvalRequestID int64, ttl time.Duration) (string, error) {
	token, err := generateOpaqueToken(magicLinkTokenLength)
	if err != nil {
		return "", err
	}
	expires := time.Now().UTC().Add(ttl)
	err = db.Model(&ContractApprover{}).Where("id = ?", approverID).Updates(map[string]interface{}{
		"token_hash":                 hashToken(token),
		"token_expires_at":           expires,
		"active_approval_request_id": approvalRequestID,
	}).Error
	return token, err
}

// GetApproverByToken resolves a plaintext magic-link token to the exact
// approver it was issued to and the approval request it's for, rejecting
// expired tokens or approval requests that are no longer pending (already
// decided, or superseded by a newer contract version/token).
func GetApproverByToken(token string) (ContractApprover, ApprovalRequest, error) {
	a := ContractApprover{}
	err := db.Where("token_hash = ?", hashToken(token)).First(&a).Error
	if err == gorm.ErrRecordNotFound {
		return a, ApprovalRequest{}, ErrApproverTokenInvalid
	}
	if err != nil {
		return a, ApprovalRequest{}, err
	}
	if a.TokenExpiresAt == nil || time.Now().UTC().After(*a.TokenExpiresAt) || a.ActiveApprovalRequestId == nil {
		return a, ApprovalRequest{}, ErrApproverTokenInvalid
	}
	ar, err := GetApprovalRequest(*a.ActiveApprovalRequestId)
	if err != nil {
		return a, ApprovalRequest{}, ErrApproverTokenInvalid
	}
	if ar.Status != ApprovalPending || time.Now().UTC().After(ar.TokenExpiresAt) {
		return a, ar, ErrApproverTokenInvalid
	}
	return a, ar, nil
}

// campaignApprovalOK reports whether a campaign is clear to run. A campaign
// with no linked contract (contractID == nil) is always clear — contract
// linkage is opt-in. A linked campaign requires an approved approval_request
// for the contract's *current* version; uploading a new version orphans any
// prior approval, which is what invalidates approval after a material
// change.
func campaignApprovalOK(contractID *int64) (bool, error) {
	if contractID == nil {
		return true, nil
	}
	latest, err := GetLatestContractVersion(*contractID)
	if err == ErrContractVersionNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var ar ApprovalRequest
	err = db.Where("contract_version_id = ? AND status = ?", latest.Id, ApprovalApproved).
		Order("decided_at desc").First(&ar).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
