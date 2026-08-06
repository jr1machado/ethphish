package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gophish/gophish/approvals"
	"github.com/gophish/gophish/config"
	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
)

// approvalPortalBaseURL is the public base URL clients use to reach the
// approval portal (mounted on the phishing server). See
// InitApprovalServices and config.Config.ApprovalPortalBaseURL.
var approvalPortalBaseURL string

// InitApprovalServices wires up the public base URL used to build magic
// links in approval e-mails.
func InitApprovalServices(cfg *config.Config) {
	if cfg != nil {
		approvalPortalBaseURL = cfg.ApprovalPortalBaseURL
		trainingPortalBaseURL = cfg.ApprovalPortalBaseURL
	}
}

// approvalNotifyResponse wraps an approval request with how many of its
// configured approvers were actually e-mailed, so the UI can tell a real
// send failure (e.g. no SMTP profile configured) apart from success instead
// of always flashing "sent" just because the request record was created.
type approvalNotifyResponse struct {
	models.ApprovalRequest
	ApproversNotified int `json:"approvers_notified"`
	ApproversTotal    int `json:"approvers_total"`
}

// approvalRequestForTenant loads an approval request and verifies it
// belongs (via its contract version's contract) to the given tenant,
// returning models.ErrApprovalNotFound if it doesn't — so a tenant can
// never view, comment on, resend, or export another tenant's approvals.
func approvalRequestForTenant(id, tenantID int64) (models.ApprovalRequest, error) {
	ar, err := models.GetApprovalRequest(id)
	if err != nil {
		return ar, err
	}
	contract, err := models.GetContractForApprovalRequest(ar)
	if err != nil {
		return models.ApprovalRequest{}, err
	}
	if contract.TenantID != tenantID {
		return models.ApprovalRequest{}, models.ErrApprovalNotFound
	}
	return ar, nil
}

// Approvals lists every approval request for the tenant (the Central de
// Aprovações table) and issues new approval requests.
func (as *Server) Approvals(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	switch r.Method {
	case "GET":
		rows, err := models.GetApprovalRequestsForTenant(scope.TenantID)
		if err != nil {
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, rows, http.StatusOK)
	case "POST":
		as.IssueApproval(w, r)
	default:
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
	}
}

type issueApprovalRequest struct {
	ContractVersionID int64  `json:"contract_version_id"`
	CampaignID        *int64 `json:"campaign_id,omitempty"`
}

// IssueApproval handles POST /api/approvals — opens a new approval cycle.
func (as *Server) IssueApproval(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	req := issueApprovalRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
		return
	}
	version, err := models.GetContractVersion(req.ContractVersionID)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}
	contract, err := models.GetContractForTenant(version.ContractId, scope.TenantID)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}
	approvers, err := models.GetContractApprovers(contract.Id)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	if len(approvers) == 0 {
		JSONResponse(w, models.Response{Success: false, Message: "Contract has no configured approvers"}, http.StatusBadRequest)
		return
	}
	ar, _, err := models.CreateApprovalRequest(version.Id, req.CampaignID, approvals.MagicLinkTTL)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	userID := ctx.Get(r, "user_id").(int64)
	notified := 0
	for _, approver := range approvers {
		token, err := models.IssueApproverToken(approver.Id, ar.Id, approvals.MagicLinkTTL)
		if err != nil {
			log.Error(err)
			continue
		}
		portalURL := fmt.Sprintf("%s/approvals/login?token=%s", approvalPortalBaseURL, token)
		if err := approvals.SendApprovalRequestEmail(scope.TenantID, userID, approver.Email, approver.Name, contract.Name, contract.ClientName, portalURL); err != nil {
			log.Error(err)
			continue
		}
		notified++
	}
	JSONResponse(w, approvalNotifyResponse{ar, notified, len(approvers)}, http.StatusCreated)
}

// ApprovalID handles retrieving a single approval request with its
// comment thread, and resending its magic link.
func (as *Server) ApprovalID(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 0, 64)
	switch r.Method {
	case "GET":
		ar, err := approvalRequestForTenant(id, scope.TenantID)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
			return
		}
		JSONResponse(w, ar, http.StatusOK)
	default:
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
	}
}

// ApprovalResend re-issues the magic link for a pending approval request.
func (as *Server) ApprovalResend(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
		return
	}
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 0, 64)
	ar, err := approvalRequestForTenant(id, scope.TenantID)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}
	if ar.Status != models.ApprovalPending {
		JSONResponse(w, models.Response{Success: false, Message: models.ErrApprovalAlreadyDecided.Error()}, http.StatusBadRequest)
		return
	}
	contract, err := models.GetContractForApprovalRequest(ar)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	approvers, err := models.GetContractApprovers(contract.Id)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	userID := ctx.Get(r, "user_id").(int64)
	notified := 0
	for _, approver := range approvers {
		token, err := models.IssueApproverToken(approver.Id, ar.Id, approvals.MagicLinkTTL)
		if err != nil {
			log.Error(err)
			continue
		}
		portalURL := fmt.Sprintf("%s/approvals/login?token=%s", approvalPortalBaseURL, token)
		if err := approvals.SendReminderEmail(scope.TenantID, userID, approver.Email, approver.Name, contract.Name, portalURL); err != nil {
			log.Error(err)
			continue
		}
		notified++
	}
	// Mirrors the cron scheduler (approvals/scheduler.go runOnce): stamp the
	// reminder timestamp whenever a resend was attempted, matching what the
	// Approval Center's "Last Reminder" column expects to see change.
	if err := models.MarkReminderSent(ar.Id); err != nil {
		log.Error(err)
	}
	JSONResponse(w, approvalNotifyResponse{ar, notified, len(approvers)}, http.StatusOK)
}

// ApprovalComment lets the admin side add a comment to the thread.
func (as *Server) ApprovalComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
		return
	}
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 0, 64)
	if _, err := approvalRequestForTenant(id, scope.TenantID); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}
	body := struct {
		Body string `json:"body"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Body == "" {
		JSONResponse(w, models.Response{Success: false, Message: "Comment body is required"}, http.StatusBadRequest)
		return
	}
	userID := ctx.Get(r, "user_id").(int64)
	if err := models.AddApprovalComment(id, "admin", userID, body.Body); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	JSONResponse(w, models.Response{Success: true, Message: "Comment added"}, http.StatusCreated)
}

// ApprovalExport returns the approval record (contract, version, decision,
// full comment thread) as proof-of-acceptance evidence that can be saved
// or archived by the admin.
func (as *Server) ApprovalExport(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 0, 64)
	ar, err := approvalRequestForTenant(id, scope.TenantID)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}
	version, err := models.GetContractVersion(ar.ContractVersionId)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	contract, err := models.GetContractByID(version.ContractId)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	evidence := struct {
		ApprovalRequest models.ApprovalRequest `json:"approval_request"`
		ContractVersion models.ContractVersion `json:"contract_version"`
		Contract        models.Contract        `json:"contract"`
	}{ar, version, contract}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"approval-%d-evidence.json\"", id))
	JSONResponse(w, evidence, http.StatusOK)
}
