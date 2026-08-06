package approvals

import (
	"fmt"
	"time"

	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
)

// ReminderInterval is how long a pending approval sits before it gets a
// reminder e-mail (and how often it gets another one after that).
const ReminderInterval = 48 * time.Hour

// tickInterval is how often the scheduler checks for expired/reminder-due
// approval requests.
const tickInterval = 1 * time.Hour

// StartScheduler runs the approval reminder/expiration cron in the
// background until the process exits. Mirrors how imap.Monitor and
// worker.DefaultWorker are started as a `go` goroutine from gophish.go.
func StartScheduler(portalBaseURL string) {
	log.Info("Approval reminder/expiration scheduler started")
	for range time.Tick(tickInterval) {
		runOnce(portalBaseURL)
	}
}

func runOnce(portalBaseURL string) {
	expired, err := models.ExpirePendingApprovals()
	if err != nil {
		log.Error(err)
	} else if expired > 0 {
		log.Infof("Expired %d pending approval request(s)", expired)
	}

	due, err := models.PendingApprovalsNeedingReminder(ReminderInterval)
	if err != nil {
		log.Error(err)
		return
	}
	for _, ar := range due {
		if err := sendReminder(ar, portalBaseURL); err != nil {
			log.Error(err)
			continue
		}
		if err := models.MarkReminderSent(ar.Id); err != nil {
			log.Error(err)
		}
	}
}

func sendReminder(ar models.ApprovalRequest, portalBaseURL string) error {
	contract, err := models.GetContractForApprovalRequest(ar)
	if err != nil {
		return err
	}
	approvers, err := models.GetContractApprovers(contract.Id)
	if err != nil {
		return err
	}
	if len(approvers) == 0 {
		return nil
	}
	for _, approver := range approvers {
		token, err := models.IssueApproverToken(approver.Id, ar.Id, MagicLinkTTL)
		if err != nil {
			log.Error(err)
			continue
		}
		portalURL := fmt.Sprintf("%s/approvals/login?token=%s", portalBaseURL, token)
		if err := SendReminderEmail(contract.TenantID, contract.CreatedBy, approver.Email, approver.Name, contract.Name, portalURL); err != nil {
			log.Error(err)
		}
	}
	return nil
}
