package controllers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gophish/gophish/approvals"
	"github.com/gophish/gophish/auth"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

// registerClientPortalRoutes mounts the client portal — the ongoing,
// self-service area a tenant's contract approver can return to on their
// own to browse campaigns and indicators, as opposed to /approvals/*,
// which is only reachable via an admin-issued approval magic link and
// only ever shows the one decision it was issued for. Both areas share
// the same session cookie (see setClientSessionCookie) and CSRF
// protection pattern, scoped to their own path prefix.
func (ps *PhishingServer) registerClientPortalRoutes(router *mux.Router) {
	sub := mux.NewRouter()
	sub.HandleFunc("/portal/login", ps.PortalLoginForm).Methods(http.MethodGet)
	sub.HandleFunc("/portal/login", ps.PortalLoginRequest).Methods(http.MethodPost)
	sub.HandleFunc("/portal/login/verify", ps.PortalLoginVerify).Methods(http.MethodGet)
	sub.HandleFunc("/portal", requirePortalSession(ps.PortalDashboard)).Methods(http.MethodGet)
	sub.HandleFunc("/portal/campaigns/{id:[0-9]+}", requirePortalSession(ps.PortalCampaign)).Methods(http.MethodGet)
	sub.HandleFunc("/portal/reports", requirePortalSession(ps.PortalReportCSV)).Methods(http.MethodGet)
	sub.HandleFunc("/portal/logout", requirePortalSession(ps.PortalLogout)).Methods(http.MethodPost)

	csrfKey := []byte(ps.config.CSRFKey)
	if len(csrfKey) == 0 {
		csrfKey = []byte(auth.GenerateSecureKey(auth.APIKeyLength))
	}
	protected := csrf.Protect(csrfKey, csrf.Path("/portal"), csrf.Secure(ps.config.UseTLS))(sub)
	router.PathPrefix("/portal").Handler(protected)
}

func requirePortalSession(handler func(http.ResponseWriter, *http.Request, models.ClientSession)) http.HandlerFunc {
	return requireClientSessionRedirectingTo("/portal/login", handler)
}

// PortalLoginForm renders the "enter your e-mail" self-service login form.
func (ps *PhishingServer) PortalLoginForm(w http.ResponseWriter, r *http.Request) {
	approvalPortalTemplate("portal_login").Execute(w, struct {
		Token   string
		Message string
	}{csrf.Token(r), ""})
}

// portalLoginRequestMessage is shown regardless of whether the submitted
// e-mail matched a known approver, so the form never confirms or denies
// that a given address is registered — the same anti-enumeration
// reasoning as a password-reset form.
const portalLoginRequestMessage = "If that e-mail is registered as a contract approver, a login link is on its way. It expires in a few minutes."

// PortalLoginRequest issues a fresh single-use login token for every
// tenant the submitted e-mail is a configured contract approver for, and
// e-mails each one a link. See models.FindTenantsForApproverEmail for how
// identity is resolved without trusting a client-supplied tenant.
func (ps *PhishingServer) PortalLoginRequest(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.FormValue("email")
	if email != "" {
		matches, err := models.FindTenantsForApproverEmail(email)
		if err != nil {
			log.Error(err)
		}
		for _, match := range matches {
			token, err := models.CreatePortalLoginToken(match.TenantID, email, approvals.PortalLoginTokenTTL)
			if err != nil {
				log.Error(err)
				continue
			}
			portalURL := fmt.Sprintf("%s/portal/login/verify?token=%s", ps.approvalPortalBaseURL, token)
			if err := approvals.SendPortalLoginEmail(match.TenantID, email, match.Name, portalURL); err != nil {
				log.Error(err)
			}
		}
	}
	approvalPortalTemplate("portal_login").Execute(w, struct {
		Token   string
		Message string
	}{csrf.Token(r), portalLoginRequestMessage})
}

// PortalLoginVerify consumes a self-service login token and issues a
// client-portal session, the same shape ApprovalLogin issues for the
// approval-decision flow.
func (ps *PhishingServer) PortalLoginVerify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	plt, err := models.RedeemPortalLoginToken(token)
	if err != nil {
		approvalPortalTemplate("portal_login").Execute(w, struct {
			Token   string
			Message string
		}{csrf.Token(r), "This login link is invalid or has expired. Request a new one below."})
		return
	}
	cu, err := models.GetOrCreateClientUser(plt.TenantID, plt.Email, "")
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	_, sessionToken, err := models.CreateClientSession(cu.Id, clientSessionTTL)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	setClientSessionCookie(w, ps, sessionToken)
	http.Redirect(w, r, "/portal", http.StatusFound)
}

// PortalDashboard lists every campaign belonging to the client's tenant,
// with aggregate stats only — never a named target row. See
// models.GetCampaignSummariesForTenantAllUsers.
func (ps *PhishingServer) PortalDashboard(w http.ResponseWriter, r *http.Request, cs models.ClientSession) {
	summaries, err := models.GetCampaignSummariesForTenantAllUsers(cs.ClientUser.TenantID)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	approvalPortalTemplate("portal_dashboard").Execute(w, struct {
		Summaries models.CampaignSummaries
		Token     string
	}{summaries, csrf.Token(r)})
}

// PortalCampaign shows one campaign's aggregate breakdown. Tenant
// membership is the only authorization check — same model as the
// dashboard, not restricted to campaigns tied to a contract this client
// approves.
func (ps *PhishingServer) PortalCampaign(w http.ResponseWriter, r *http.Request, cs models.ClientSession) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	summary, err := models.GetCampaignSummaryForTenantAllUsers(id, cs.ClientUser.TenantID)
	if err != nil || summary.Id == 0 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	approvalPortalTemplate("portal_campaign").Execute(w, struct {
		Summary models.CampaignSummary
	}{summary})
}

// PortalReportCSV streams an aggregate-only CSV of every campaign in the
// client's tenant — one row per campaign, counts only, no per-target
// data. This intentionally doesn't reuse the admin Word/Excel report
// pipeline (reports.GenerateReport): that pipeline fetches campaign data
// through models.GetCampaign/GetCampaignResults hardcoded to user ID 1
// (see reports/python_service.go FetchCampaignData), so it silently drops
// results for campaigns owned by any other admin user — unsafe to expose
// tenant-wide until that's fixed. The aggregate stats used here come from
// getCampaignStats, which has no such owner assumption.
func (ps *PhishingServer) PortalReportCSV(w http.ResponseWriter, r *http.Request, cs models.ClientSession) {
	summaries, err := models.GetCampaignSummariesForTenantAllUsers(cs.ClientUser.TenantID)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"ethphish_portal_report.csv\"")
	cw := csv.NewWriter(w)
	cw.Write([]string{"Campaign", "Type", "Status", "Launch Date", "Sent", "Opened", "Clicked", "Submitted Data", "Reported", "Errors"})
	for _, c := range summaries.Campaigns {
		cw.Write([]string{
			c.Name,
			c.Type,
			c.Status,
			c.LaunchDate.Format("2006-01-02"),
			strconv.FormatInt(c.Stats.EmailsSent, 10),
			strconv.FormatInt(c.Stats.OpenedEmail, 10),
			strconv.FormatInt(c.Stats.ClickedLink, 10),
			strconv.FormatInt(c.Stats.SubmittedData, 10),
			strconv.FormatInt(c.Stats.EmailReported, 10),
			strconv.FormatInt(c.Stats.Error, 10),
		})
	}
	cw.Flush()
}

// PortalLogout destroys the client's portal session.
func (ps *PhishingServer) PortalLogout(w http.ResponseWriter, r *http.Request, cs models.ClientSession) {
	clearClientSession(w, r)
	http.Redirect(w, r, "/portal/login", http.StatusFound)
}
