// Package approvals sends the transactional e-mails for the contract
// approval workflow (request, reminder, decision notification) and runs
// the background reminder/expiration cron.
package approvals

import (
	"context"
	"fmt"
	"net/mail"
	"time"

	"github.com/gophish/gomail"
	"github.com/gophish/gophish/mailer"
	"github.com/gophish/gophish/models"
)

// MagicLinkTTL is how long an approval magic link (and the client session
// it establishes) stays valid before requiring a resend.
const MagicLinkTTL = 7 * 24 * time.Hour

// PortalLoginTokenTTL is how long a self-service client-portal login link
// stays valid. Short-lived by design — it's a login link, not an
// approval window, so it doesn't need to survive days like MagicLinkTTL.
const PortalLoginTokenTTL = 15 * time.Minute

// approvalMail is a one-off transactional email, following the same
// mailer.Mail contract as models.MailLog (see models/maillog.go) but for
// plain admin-to-client notices rather than phishing simulation content.
type approvalMail struct {
	smtp    models.SMTP
	to      string
	subject string
	body    string
}

func (m *approvalMail) Backoff(reason error) error { return nil }
func (m *approvalMail) Error(err error) error      { return err }
func (m *approvalMail) Success() error             { return nil }

func (m *approvalMail) GetDialer() (mailer.Dialer, error) {
	return m.smtp.GetDialer()
}

func (m *approvalMail) GetSmtpFrom() (string, error) {
	f, err := mail.ParseAddress(m.smtp.FromAddress)
	if err != nil {
		return "", err
	}
	return f.Address, nil
}

func (m *approvalMail) Generate(msg *gomail.Message) error {
	f, err := mail.ParseAddress(m.smtp.FromAddress)
	if err != nil {
		return err
	}
	msg.SetAddressHeader("From", f.Address, f.Name)
	msg.SetHeader("To", m.to)
	msg.SetHeader("Subject", m.subject)
	msg.SetBody("text/plain", m.body)
	return nil
}

// tenantSMTP returns the tenant's first configured sending profile.
// EthPhish has no separate "system mailer" concept yet, so approval e-mails
// ride whichever sending profile the admin already configured.
// ponytail: picks the first profile rather than a designated "system"
// profile; add a dedicated system-mailer config if multiple profiles ever
// need to disambiguate which one sends approval mail.
func tenantSMTP(tenantID, uid int64) (models.SMTP, error) {
	smtps, err := models.GetSMTPsForTenant(tenantID, uid)
	if err != nil {
		return models.SMTP{}, err
	}
	if len(smtps) == 0 {
		return models.SMTP{}, fmt.Errorf("no sending profile configured for this tenant")
	}
	return smtps[0], nil
}

func send(tenantID, uid int64, to, subject, body string) error {
	smtp, err := tenantSMTP(tenantID, uid)
	if err != nil {
		return err
	}
	m := &approvalMail{smtp: smtp, to: to, subject: subject, body: body}
	return mailer.SendOne(context.Background(), m)
}

// sendSystem is send's equivalent for e-mail with no acting admin user —
// the client portal's self-service login is triggered by the client, not
// by an admin action, so there's no uid to resolve an owned SMTP profile
// with; it uses whichever sending profile the tenant has configured.
func sendSystem(tenantID int64, to, subject, body string) error {
	smtp, err := models.GetAnySMTPForTenant(tenantID)
	if err != nil {
		return fmt.Errorf("no sending profile configured for this tenant")
	}
	m := &approvalMail{smtp: smtp, to: to, subject: subject, body: body}
	return mailer.SendOne(context.Background(), m)
}

// SendApprovalRequestEmail notifies a contract approver that a version is
// awaiting their decision, with a link to the client approval portal that
// logs them in via the embedded magic-link token.
func SendApprovalRequestEmail(tenantID, uid int64, approverEmail, approverName, contractName, clientName, portalURL string) error {
	subject := fmt.Sprintf("Action required: approve %q", contractName)
	body := fmt.Sprintf(
		"Hello %s,\n\n%s has requested your approval for the contract %q before the related campaign can run.\n\nReview and respond here:\n%s\n\nThis link expires after a few days; if it does, ask your EthPhish contact for a new one.\n\n-- EthPhish",
		approverName, clientName, contractName, portalURL,
	)
	return send(tenantID, uid, approverEmail, subject, body)
}

// SendReminderEmail re-notifies an approver about a still-pending approval
// request.
func SendReminderEmail(tenantID, uid int64, approverEmail, approverName, contractName, portalURL string) error {
	subject := fmt.Sprintf("Reminder: approval still pending for %q", contractName)
	body := fmt.Sprintf(
		"Hello %s,\n\nThis is a reminder that your approval is still pending for %q.\n\nReview and respond here:\n%s\n\nThis reminder replaces any earlier approval e-mail — links from previous messages for this contract no longer work.\n\n-- EthPhish",
		approverName, contractName, portalURL,
	)
	return send(tenantID, uid, approverEmail, subject, body)
}

// SendPortalLoginEmail delivers a self-service client-portal login link.
// Unlike the approval e-mails above, this isn't triggered by an admin
// action — a client requested it themselves from the portal's login form.
func SendPortalLoginEmail(tenantID int64, to, name, portalURL string) error {
	subject := "Your EthPhish client portal login link"
	body := fmt.Sprintf(
		"Hello %s,\n\nUse this link to sign in to the client portal:\n%s\n\nThis link expires shortly and can only be used once. If you didn't request it, you can ignore this e-mail.\n\n-- EthPhish",
		name, portalURL,
	)
	return sendSystem(tenantID, to, subject, body)
}

// SendTrainingAssignmentEmail notifies a target they've been assigned an
// awareness training, with their unique access link. Admin-triggered
// (via the "Assign to Group" action), so it uses send (uid-scoped SMTP
// profile), not sendSystem.
func SendTrainingAssignmentEmail(tenantID, uid int64, to, name, trainingName, link string) error {
	subject := fmt.Sprintf("Training assigned: %q", trainingName)
	body := fmt.Sprintf(
		"Hello %s,\n\nYou've been assigned the training %q.\n\nStart here:\n%s\n\n-- EthPhish",
		name, trainingName, link,
	)
	return send(tenantID, uid, to, subject, body)
}

// SendDecisionNotificationEmail tells the admin who requested an approval
// what the client decided.
func SendDecisionNotificationEmail(tenantID, uid int64, adminRecipient, contractName, status, comment string) error {
	subject := fmt.Sprintf("Decision recorded: %q is %s", contractName, status)
	body := fmt.Sprintf("The contract %q was marked %q by the client.", contractName, status)
	if comment != "" {
		body += fmt.Sprintf("\n\nComment:\n%s", comment)
	}
	return send(tenantID, uid, adminRecipient, subject, body)
}
