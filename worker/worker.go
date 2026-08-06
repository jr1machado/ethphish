package worker

import (
	"context"
	"strconv"
	"time"

	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/mailer"
	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/queue"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

// mailQueueConcurrency is how many messages are processed off the RabbitMQ
// mail queue at once.
const mailQueueConcurrency = 4

// Worker is an interface that defines the operations needed for a background worker
type Worker interface {
	Start()
	LaunchCampaign(c models.Campaign)
	SendTestEmail(s *models.EmailRequest) error
	SendTestSMS(s *models.SMSRequest) error
}

// DefaultWorker is the background worker that handles watching for new campaigns and sending emails and SMS messages appropriately.
type DefaultWorker struct {
	mailer    mailer.Mailer
	smsMailer *mailer.SMSWorker
	mailQueue *queue.Client
}

// New creates a new worker object to handle the creation of campaigns
func New(options ...func(Worker) error) (Worker, error) {
	defaultMailer := mailer.NewMailWorker()
	defaultSMSMailer := mailer.NewSMSWorker()
	w := &DefaultWorker{
		mailer:    defaultMailer,
		smsMailer: defaultSMSMailer,
	}
	for _, opt := range options {
		if err := opt(w); err != nil {
			return nil, err
		}
	}
	return w, nil
}

// WithMailer sets the mailer for a given worker.
// By default, workers use a standard, default mailworker.
func WithMailer(m mailer.Mailer) func(*DefaultWorker) error {
	return func(w *DefaultWorker) error {
		w.mailer = m
		return nil
	}
}

// WithRabbitMQURL connects to RabbitMQ and routes campaign email dispatch
// through a durable queue instead of the in-process mailer channel. Without
// this option (empty url, or url unset) the worker falls back to the direct
// in-process path unchanged, which is what every non-RabbitMQ test relies on.
func WithRabbitMQURL(url string) func(Worker) error {
	return func(iw Worker) error {
		if url == "" {
			return nil
		}
		w, ok := iw.(*DefaultWorker)
		if !ok {
			return nil
		}
		client, err := queue.Connect(url)
		if err != nil {
			return err
		}
		w.mailQueue = client
		return nil
	}
}

// processCampaigns loads maillogs scheduled to be sent before the provided
// time and sends them to the mailer.
func (w *DefaultWorker) processCampaigns(t time.Time) error {
	ms, err := models.GetQueuedMailLogs(t.UTC())
	if err != nil {
		log.Error(err)
		return err
	}
	// Lock the MailLogs (they will be unlocked after processing)
	err = models.LockMailLogs(ms, true)
	if err != nil {
		return err
	}
	campaignCache := make(map[int64]models.Campaign)
	// We'll group the maillogs by campaign ID to (roughly) group
	// them by sending profile. This lets the mailer re-use the Sender
	// instead of having to re-connect to the SMTP server for every
	// email.
	msg := make(map[int64][]*models.MailLog)
	for _, m := range ms {
		// We cache the campaign here to greatly reduce the time it takes to
		// generate the message (ref #1726)
		c, ok := campaignCache[m.CampaignId]
		if !ok {
			c, err = models.GetCampaignMailContext(m.CampaignId, m.UserId)
			if err != nil {
				return err
			}
			campaignCache[c.Id] = c
		}
		m.CacheCampaign(&c)
		msg[m.CampaignId] = append(msg[m.CampaignId], m)
	}

	// Next, we process each group of maillogs in parallel
	for cid, msc := range msg {
		go func(cid int64, msc []*models.MailLog) {
			c := campaignCache[cid]
			if ok, err := models.IsCampaignApproved(c.ContractID); err != nil {
				log.Error(err)
				return
			} else if !ok {
				log.WithFields(logrus.Fields{
					"campaign_id": cid,
				}).Warn("Skipping send: campaign's contract approval is missing or stale")
				return
			}
			if c.Status == models.CampaignQueued {
				err := c.UpdateStatus(models.CampaignInProgress)
				if err != nil {
					log.Error(err)
					return
				}
			}
			log.WithFields(logrus.Fields{
				"num_emails": len(msc),
			}).Info("Sending emails to mailer for processing")
			w.dispatchMailLogs(msc)
		}(cid, msc)
	}
	return nil
}

// dispatchMailLogs hands off each MailLog for delivery. When a RabbitMQ
// queue is configured each MailLog is published as its own durable message
// (see queue.Client.Consume for the retry/DLQ semantics); otherwise it falls
// back to the direct in-process mailer channel, unchanged from before this
// queue existed.
func (w *DefaultWorker) dispatchMailLogs(ms []*models.MailLog) {
	if w.mailQueue == nil {
		mailEntries := make([]mailer.Mail, len(ms))
		for i, m := range ms {
			mailEntries[i] = m
		}
		w.mailer.Queue(mailEntries)
		return
	}
	for _, m := range ms {
		body := []byte(strconv.FormatInt(m.Id, 10))
		if err := w.mailQueue.Publish(context.Background(), queue.MailSendQueue, body, 0); err != nil {
			log.WithFields(logrus.Fields{
				"maillog_id": m.Id,
			}).Error("publishing maillog to RabbitMQ, falling back to direct send: ", err)
			w.mailer.Queue([]mailer.Mail{m})
		}
	}
}

// consumeMailQueue processes MailLogs published to RabbitMQ until ctx is
// done. It is a no-op when no RabbitMQ queue is configured.
func (w *DefaultWorker) consumeMailQueue(ctx context.Context) {
	if w.mailQueue == nil {
		return
	}
	err := w.mailQueue.Consume(ctx, mailQueueConcurrency, w.handleMailMessage)
	if err != nil {
		log.Error("consuming RabbitMQ mail queue: ", err)
	}
}

// handleMailMessage loads the MailLog named by body and sends it.
//
// A returned nil error means the message was fully handled: either the
// MailLog no longer exists (already sent or permanently failed by an
// earlier delivery of this same message — an idempotent no-op) or
// mailer.SendOne ran, which itself persists the outcome via
// MailLog.Success/Error/Backoff regardless of whether the send worked. A
// non-nil error means processing itself failed (for example the database
// was unreachable) and the caller should retry the message.
func (w *DefaultWorker) handleMailMessage(ctx context.Context, body []byte) error {
	id, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		log.Error("malformed maillog id in mail queue message: ", err)
		return nil
	}
	m, err := models.GetMailLogByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	return mailer.SendOne(ctx, m)
}

// Start launches the worker to poll the database every minute for any pending maillogs
// and smslogs that need to be processed.
func (w *DefaultWorker) Start() {
	log.Info("Background Worker Started Successfully - Waiting for Campaigns")
	ctx := context.Background()
	go w.mailer.Start(ctx)
	go w.smsMailer.Start(ctx)
	go w.consumeMailQueue(ctx)
	for t := range time.Tick(1 * time.Minute) {
		err := w.processCampaigns(t)
		if err != nil {
			log.Error(err)
			continue
		}

		err = w.processSMSCampaigns(t)
		if err != nil {
			log.Error(err)
			continue
		}
	}
}

// processSMSCampaigns loads smslogs scheduled to be sent before the provided
// time and sends them to the SMS mailer.
func (w *DefaultWorker) processSMSCampaigns(t time.Time) error {
	ss, err := models.GetQueuedSMSLogs(t.UTC())
	if err != nil {
		log.Error(err)
		return err
	}
	// Lock the SMSLogs (they will be unlocked after processing)
	err = models.LockSMSLogs(ss, true)
	if err != nil {
		return err
	}
	campaignCache := make(map[int64]models.Campaign)
	// We'll group the smslogs by campaign ID to (roughly) group
	// them by sending profile. This lets the mailer re-use the Sender
	// instead of having to re-connect to the SMS provider for every
	// message.
	msg := make(map[int64][]mailer.SMSMail)
	for _, s := range ss {
		// We cache the campaign here to greatly reduce the time it takes to
		// generate the message
		c, ok := campaignCache[s.CampaignId]
		if !ok {
			c, err = models.GetCampaignSMSContext(s.CampaignId, s.UserId)
			if err != nil {
				return err
			}
			campaignCache[c.Id] = c
		}
		s.CacheCampaign(&c)
		msg[s.CampaignId] = append(msg[s.CampaignId], s)
	}

	// Next, we process each group of smslogs in parallel
	for cid, ssc := range msg {
		go func(cid int64, ssc []mailer.SMSMail) {
			c := campaignCache[cid]
			if ok, err := models.IsCampaignApproved(c.ContractID); err != nil {
				log.Error(err)
				return
			} else if !ok {
				log.WithFields(logrus.Fields{
					"campaign_id": cid,
				}).Warn("Skipping send: campaign's contract approval is missing or stale")
				return
			}
			if c.Status == models.CampaignQueued {
				err := c.UpdateStatus(models.CampaignInProgress)
				if err != nil {
					log.Error(err)
					return
				}
			}
			log.WithFields(logrus.Fields{
				"num_sms": len(ssc),
			}).Info("Sending SMS messages to SMS mailer for processing")
			w.smsMailer.Queue(ssc)
		}(cid, ssc)
	}
	return nil
}

// LaunchCampaign starts a campaign
func (w *DefaultWorker) LaunchCampaign(c models.Campaign) {
	// Handle different campaign types
	if c.Type == "sms" {
		w.launchSMSCampaign(c)
	} else {
		w.launchEmailCampaign(c)
	}
}

// launchEmailCampaign starts an email campaign
func (w *DefaultWorker) launchEmailCampaign(c models.Campaign) {
	ms, err := models.GetMailLogsByCampaign(c.Id)
	if err != nil {
		log.Error(err)
		return
	}
	models.LockMailLogs(ms, true)
	mailEntries := []*models.MailLog{}
	currentTime := time.Now().UTC()
	campaignMailCtx, err := models.GetCampaignMailContext(c.Id, c.UserId)
	if err != nil {
		log.Error(err)
		return
	}
	for _, m := range ms {
		// Only send the emails scheduled to be sent for the past minute to
		// respect the campaign scheduling options
		if m.SendDate.After(currentTime) {
			m.Unlock()
			continue
		}
		err = m.CacheCampaign(&campaignMailCtx)
		if err != nil {
			log.Error(err)
			return
		}
		mailEntries = append(mailEntries, m)
	}
	w.dispatchMailLogs(mailEntries)
}

// launchSMSCampaign starts an SMS campaign
func (w *DefaultWorker) launchSMSCampaign(c models.Campaign) {
	ss, err := models.GetSMSLogsByCampaign(c.Id)
	if err != nil {
		log.Error(err)
		return
	}
	models.LockSMSLogs(ss, true)
	// This is required since you cannot pass a slice of values
	// that implements an interface as a slice of that interface.
	smsEntries := []mailer.SMSMail{}
	currentTime := time.Now().UTC()
	campaignSMSCtx, err := models.GetCampaignSMSContext(c.Id, c.UserId)
	if err != nil {
		log.Error(err)
		return
	}
	for _, s := range ss {
		// Only send the SMS messages scheduled to be sent for the past minute to
		// respect the campaign scheduling options
		if s.SendDate.After(currentTime) {
			s.Unlock()
			continue
		}
		err = s.CacheCampaign(&campaignSMSCtx)
		if err != nil {
			log.Error(err)
			return
		}
		smsEntries = append(smsEntries, s)
	}
	w.smsMailer.Queue(smsEntries)
}

// SendTestEmail sends a test email
func (w *DefaultWorker) SendTestEmail(s *models.EmailRequest) error {
	go func() {
		ms := []mailer.Mail{s}
		w.mailer.Queue(ms)
	}()
	return <-s.ErrorChan
}

// SendTestSMS sends a test SMS
func (w *DefaultWorker) SendTestSMS(s *models.SMSRequest) error {
	go func() {
		ms := []mailer.SMSMail{s}
		w.smsMailer.Queue(ms)
	}()
	return <-s.ErrorChan
}
