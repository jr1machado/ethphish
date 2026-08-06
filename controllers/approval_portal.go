package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gophish/gophish/approvals"
	"github.com/gophish/gophish/auth"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

// clientSessionCookieName backs the client-portal session, kept entirely
// separate from the admin session cookie (middleware.Store / "gophish").
const clientSessionCookieName = "ethphish_client"

// clientSessionTTL matches approvals.MagicLinkTTL — a client's session
// lasts as long as the magic link that created it would have.
const clientSessionTTL = approvals.MagicLinkTTL

// registerApprovalPortalRoutes mounts the client-facing contract approval
// portal on the phishing server — the public, internet-facing server (see
// controllers/phish.go) — under /approvals. CSRF protection is scoped to
// just this subtree via its own sub-router: the rest of the phishing
// server intentionally has none, since credential-capture forms and
// tracking pixels are legitimate cross-origin POSTs by design.
func (ps *PhishingServer) registerApprovalPortalRoutes(router *mux.Router) {
	sub := mux.NewRouter()
	sub.HandleFunc("/approvals/login", ps.ApprovalLogin).Methods(http.MethodGet)
	sub.HandleFunc("/approvals", requireClientSession(ps.ApprovalDashboard)).Methods(http.MethodGet)
	sub.HandleFunc("/approvals/{id:[0-9]+}", requireClientSession(ps.ApprovalView)).Methods(http.MethodGet)
	// POST-only: gorilla/csrf only enforces its token check on unsafe
	// methods, so a decision/logout route reachable via GET would let an
	// attacker trigger it cross-site with a plain link or <img> tag.
	sub.HandleFunc("/approvals/{id:[0-9]+}/decision", requireClientSession(ps.ApprovalDecision)).Methods(http.MethodPost)
	sub.HandleFunc("/approvals/logout", requireClientSession(ps.ApprovalLogout)).Methods(http.MethodPost)

	csrfKey := []byte(ps.config.CSRFKey)
	if len(csrfKey) == 0 {
		csrfKey = []byte(auth.GenerateSecureKey(auth.APIKeyLength))
	}
	protected := csrf.Protect(csrfKey, csrf.Path("/approvals"), csrf.Secure(ps.config.UseTLS))(sub)
	router.PathPrefix("/approvals").Handler(protected)
}

func requireClientSession(handler func(http.ResponseWriter, *http.Request, models.ClientSession)) http.HandlerFunc {
	return requireClientSessionRedirectingTo("/approvals/login", handler)
}

// requireClientSessionRedirectingTo is the shared guard behind both
// requireClientSession (/approvals/*) and requirePortalSession
// (/portal/*) — same cookie, same session store, different login page to
// bounce back to so a client landing on either area unauthenticated ends
// up on that area's own entry point.
func requireClientSessionRedirectingTo(loginPath string, handler func(http.ResponseWriter, *http.Request, models.ClientSession)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(clientSessionCookieName)
		if err != nil {
			http.Redirect(w, r, loginPath, http.StatusFound)
			return
		}
		cs, err := models.GetClientSessionByToken(cookie.Value)
		if err != nil {
			http.Redirect(w, r, loginPath, http.StatusFound)
			return
		}
		handler(w, r, cs)
	}
}

// setClientSessionCookie issues the shared client-portal session cookie.
// Path is "/" — not "/approvals" — because the same session now backs both
// the approval-decision area and the broader client portal (/portal/*);
// scoping it to one path would make it invisible on the other.
func setClientSessionCookie(w http.ResponseWriter, ps *PhishingServer, sessionToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     clientSessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   ps.config.UseTLS,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().UTC().Add(clientSessionTTL),
	})
}

func approvalPortalTemplate(name string) *template.Template {
	t, err := template.ParseFiles("templates/" + name + ".html")
	if err != nil {
		log.Error(err)
	}
	return t
}

// ApprovalLogin consumes a magic-link token and issues a client-portal
// session cookie. The approver's identity is derived entirely from the
// token — via models.GetApproverByToken, which was minted for exactly one
// approver by IssueApproverToken — never from client-supplied input, so
// one approver's link can't be replayed to impersonate a different named
// approver on the same contract. See approvals.SendApprovalRequestEmail
// for how the link is built.
func (ps *PhishingServer) ApprovalLogin(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	approver, ar, err := models.GetApproverByToken(token)
	if err != nil {
		renderClientLogin(w, "This approval link is invalid or has expired.")
		return
	}
	contract, err := models.GetContractByID(approver.ContractId)
	if err != nil {
		renderClientLogin(w, "This approval link is invalid or has expired.")
		return
	}
	cu, err := models.GetOrCreateClientUser(contract.TenantID, approver.Email, approver.Name)
	if err != nil {
		log.Error(err)
		renderClientLogin(w, "Something went wrong. Please try again later.")
		return
	}
	_, sessionToken, err := models.CreateClientSession(cu.Id, clientSessionTTL)
	if err != nil {
		log.Error(err)
		renderClientLogin(w, "Something went wrong. Please try again later.")
		return
	}
	setClientSessionCookie(w, ps, sessionToken)
	http.Redirect(w, r, fmt.Sprintf("/approvals/%d", ar.Id), http.StatusFound)
}

func renderClientLogin(w http.ResponseWriter, message string) {
	approvalPortalTemplate("client_login").Execute(w, struct{ Message string }{message})
}

// ApprovalDashboard lists every approval request awaiting the logged-in
// client user's decision.
func (ps *PhishingServer) ApprovalDashboard(w http.ResponseWriter, r *http.Request, cs models.ClientSession) {
	rows, err := models.GetPendingApprovalsForClientUser(cs.ClientUserId)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	approvalPortalTemplate("client_dashboard").Execute(w, struct {
		Approvals []models.ApprovalRequestSummary
		Token     string
	}{rows, csrf.Token(r)})
}

// clientAuthorizedForContract checks the session's client user e-mail is a
// configured approver for the contract behind an approval request, and
// that the contract belongs to the same tenant the client user was
// created under — a client user is only ever created within one tenant's
// magic-link flow (see ApprovalLogin), but this keeps a same-email
// approver reused across tenants from crossing tenant boundaries.
func clientAuthorizedForContract(cs models.ClientSession, contract models.Contract) (bool, error) {
	if cs.ClientUser.TenantID != contract.TenantID {
		return false, nil
	}
	approvers, err := models.GetContractApprovers(contract.Id)
	if err != nil {
		return false, err
	}
	for _, a := range approvers {
		if a.Email == cs.ClientUser.Email {
			return true, nil
		}
	}
	return false, nil
}

// ApprovalView renders one contract version, its scope, and the comment
// thread, with approve/reject/request-changes actions.
func (ps *PhishingServer) ApprovalView(w http.ResponseWriter, r *http.Request, cs models.ClientSession) {
	id := mux.Vars(r)["id"]
	ar, err := models.GetApprovalRequest(atoi64(id))
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	version, err := models.GetContractVersion(ar.ContractVersionId)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	contract, err := models.GetContractByID(version.ContractId)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	authorized, err := clientAuthorizedForContract(cs, contract)
	if err != nil || !authorized {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	approvalPortalTemplate("client_approval").Execute(w, struct {
		ApprovalRequest models.ApprovalRequest
		Version         models.ContractVersion
		Contract        models.Contract
		Token           string
	}{ar, version, contract, csrf.Token(r)})
}

// ApprovalDecision records the client's approve/reject/request-changes
// decision and notifies the admin who requested it.
func (ps *PhishingServer) ApprovalDecision(w http.ResponseWriter, r *http.Request, cs models.ClientSession) {
	id := atoi64(mux.Vars(r)["id"])
	ar, err := models.GetApprovalRequest(id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	version, err := models.GetContractVersion(ar.ContractVersionId)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	contract, err := models.GetContractByID(version.ContractId)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	authorized, err := clientAuthorizedForContract(cs, contract)
	if err != nil || !authorized {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	r.ParseForm()
	action := r.FormValue("action")
	comment := r.FormValue("comment")
	var status string
	switch action {
	case "approve":
		status = models.ApprovalApproved
	case "reject":
		status = models.ApprovalRejected
	case "changes":
		status = models.ApprovalChangesRequested
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}
	if err := models.DecideApproval(id, cs.ClientUserId, status, comment); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if admin, err := models.GetUser(contract.CreatedBy); err == nil {
		if err := approvals.SendDecisionNotificationEmail(contract.TenantID, contract.CreatedBy, admin.Username, contract.Name, status, comment); err != nil {
			log.Error(err)
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/approvals/%d", id), http.StatusFound)
}

// ApprovalLogout destroys the client-portal session.
func (ps *PhishingServer) ApprovalLogout(w http.ResponseWriter, r *http.Request, cs models.ClientSession) {
	clearClientSession(w, r)
	http.Redirect(w, r, "/approvals/login", http.StatusFound)
}

// clearClientSession destroys the session record and its cookie, shared
// by both the approval and portal logout handlers.
func clearClientSession(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(clientSessionCookieName); err == nil {
		models.DeleteClientSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: clientSessionCookieName, Value: "", Path: "/", MaxAge: -1})
}

func atoi64(s string) int64 {
	var n int64
	fmt.Sscanf(s, "%d", &n)
	return n
}
