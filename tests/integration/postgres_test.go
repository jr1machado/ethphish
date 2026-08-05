package integration

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/models"
)

const postgresDSNEnv = "ETHPHISH_TEST_POSTGRES_DSN"

// setupPostgres verifies the deployment path used in development and CI: a
// fresh PostgreSQL database accepts every migration and is available through
// the application readiness primitive. It is opt-in for local development to
// keep the normal unit-test suite self-contained.
func setupPostgres(t *testing.T) {
	t.Helper()
	dsn := os.Getenv(postgresDSNEnv)
	if dsn == "" {
		t.Skipf("set %s to run PostgreSQL integration tests", postgresDSNEnv)
	}

	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locating integration test source")
	}
	migrationsPath := filepath.Join(filepath.Dir(sourceFile), "..", "..", "db", "db_postgres", "migrations")

	t.Setenv(models.InitialAdminPassword, "integration-test-password")
	t.Setenv(models.InitialAdminApiToken, "integration-test-api-token")
	conf := &config.Config{
		DBName:         "postgres",
		DBPath:         dsn,
		MigrationsPath: migrationsPath,
		DBMaxOpenConns: 5,
		DBMaxIdleConns: 2,
		DBConnMaxLife:  time.Minute,
	}
	if err := models.Setup(conf); err != nil {
		t.Fatalf("setting up PostgreSQL and applying migrations: %v", err)
	}
	if err := models.Ping(); err != nil {
		t.Fatalf("checking PostgreSQL readiness: %v", err)
	}
}

func TestPostgresMigrationsAndReadiness(t *testing.T) {
	setupPostgres(t)
	if _, err := models.GetUserByUsername(models.DefaultAdminUsername); err != nil {
		t.Fatalf("checking migrated administrative user: %v", err)
	}
}

// TestPostgresCampaignPersistence exercises the core authorized-simulation
// persistence path without sending an email or contacting an external system.
func TestPostgresCampaignPersistence(t *testing.T) {
	setupPostgres(t)

	group := models.Group{
		Name:   "PostgreSQL integration group",
		UserId: 1,
		Targets: []models.Target{
			{BaseRecipient: models.BaseRecipient{Email: "participant@example.test", FirstName: "Participant"}},
		},
	}
	if err := models.PostGroup(&group); err != nil {
		t.Fatalf("creating target group: %v", err)
	}
	template := models.Template{
		Name:    "PostgreSQL integration template",
		Subject: "Approved simulation",
		Text:    "Training message",
		HTML:    "<p>Training message</p>",
		UserId:  1,
	}
	if err := models.PostTemplate(&template); err != nil {
		t.Fatalf("creating email template: %v", err)
	}
	page := models.Page{Name: "PostgreSQL integration page", HTML: "<html>Training</html>", UserId: 1}
	if err := models.PostPage(&page); err != nil {
		t.Fatalf("creating landing page: %v", err)
	}
	smtp := models.SMTP{
		Name:        "PostgreSQL integration profile",
		Host:        "smtp.example.test:2525",
		FromAddress: "training@example.test",
		UserId:      1,
	}
	if err := models.PostSMTP(&smtp); err != nil {
		t.Fatalf("creating SMTP profile: %v", err)
	}
	campaign := models.Campaign{
		Name:     "PostgreSQL integration campaign",
		UserId:   1,
		Template: template,
		Page:     page,
		SMTP:     smtp,
		Groups:   []models.Group{group},
	}
	if err := models.PostCampaign(&campaign, campaign.UserId); err != nil {
		t.Fatalf("persisting campaign: %v", err)
	}
	stored, err := models.GetCampaign(campaign.Id, campaign.UserId)
	if err != nil {
		t.Fatalf("reading campaign: %v", err)
	}
	if stored.Name != campaign.Name || len(stored.Results) != 1 {
		t.Fatalf("stored campaign = %#v; want name %q and one result", stored, campaign.Name)
	}
	mailLogs, err := models.GetMailLogsByCampaign(campaign.Id)
	if err != nil {
		t.Fatalf("reading campaign mail logs: %v", err)
	}
	if len(mailLogs) != 1 {
		t.Fatalf("mail log count = %d, want 1", len(mailLogs))
	}
}

// TestPostgresSMSCampaignPersistence covers the SMS persistence path without
// creating a network client or attempting delivery.
func TestPostgresSMSCampaignPersistence(t *testing.T) {
	setupPostgres(t)

	group := models.Group{
		Name:   "PostgreSQL SMS integration group",
		UserId: 1,
		Targets: []models.Target{
			{BaseRecipient: models.BaseRecipient{Email: "participant@example.test", Phone: "+15551234567"}},
		},
	}
	if err := models.PostGroup(&group); err != nil {
		t.Fatalf("creating SMS target group: %v", err)
	}
	template := models.SMSTemplate{Name: "PostgreSQL SMS template", Text: "Authorized training", UserId: 1}
	if err := models.PostSMSTemplate(&template); err != nil {
		t.Fatalf("creating SMS template: %v", err)
	}
	page := models.Page{Name: "PostgreSQL SMS page", HTML: "<html>Training</html>", UserId: 1}
	if err := models.PostPage(&page); err != nil {
		t.Fatalf("creating SMS landing page: %v", err)
	}
	profile := models.SMS{
		Name:           "PostgreSQL SMS profile",
		Provider:       "twilio",
		From:           "+15557654321",
		ProviderConfig: `{"account_sid":"integration","auth_token":"integration"}`,
		UserId:         1,
	}
	if err := models.PostSMS(&profile); err != nil {
		t.Fatalf("creating SMS profile: %v", err)
	}
	campaign := models.Campaign{
		Name:        "PostgreSQL SMS campaign",
		UserId:      1,
		Type:        "sms",
		SMSTemplate: template,
		Page:        page,
		SMS:         profile,
		Groups:      []models.Group{group},
	}
	if err := models.PostCampaign(&campaign, campaign.UserId); err != nil {
		t.Fatalf("persisting SMS campaign: %v", err)
	}
	stored, err := models.GetCampaign(campaign.Id, campaign.UserId)
	if err != nil {
		t.Fatalf("reading SMS campaign: %v", err)
	}
	if stored.Type != "sms" || len(stored.Results) != 1 {
		t.Fatalf("stored SMS campaign has type %q and %d results; want sms and 1", stored.Type, len(stored.Results))
	}
	logs, err := models.GetSMSLogsByCampaign(campaign.Id)
	if err != nil {
		t.Fatalf("reading SMS campaign logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("SMS log count = %d, want 1", len(logs))
	}
}
