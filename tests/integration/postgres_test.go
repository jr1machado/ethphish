package integration

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gophish/gophish/config"
	appcrypto "github.com/gophish/gophish/crypto"
	"github.com/gophish/gophish/migration"
	"github.com/gophish/gophish/models"
	_ "github.com/mattn/go-sqlite3"
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

// TestPostgresSQLiteImport validates the approved migration path against an
// empty, schema-only PostgreSQL database. It contains synthetic data only.
func TestPostgresSQLiteImport(t *testing.T) {
	dsn := os.Getenv(postgresDSNEnv)
	if dsn == "" {
		t.Skipf("set %s to run PostgreSQL integration tests", postgresDSNEnv)
	}
	adminDSN := strings.Replace(dsn, "dbname=ethphish", "dbname=postgres", 1)
	importDSN := strings.Replace(dsn, "dbname=ethphish", "dbname=ethphish_import_test", 1)
	admin, err := sql.Open("postgres", adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	if _, err := admin.Exec(`DROP DATABASE IF EXISTS ethphish_import_test WITH (FORCE); CREATE DATABASE ethphish_import_test`); err != nil {
		t.Fatalf("creating isolated import database: %v", err)
	}
	target, err := sql.Open("postgres", importDSN)
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	_, sourceFile, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(sourceFile), "..", "..", "db", "db_postgres", "migrations")
	latest, err := migration.Latest(migrationsPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.Apply(context.Background(), "postgres", migrationsPath, latest, target); err != nil {
		t.Fatalf("preparing schema-only destination: %v", err)
	}
	sourcePath := filepath.Join(t.TempDir(), "legacy.db")
	source, err := sql.Open("sqlite3", sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := source.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, username TEXT NOT NULL, hash TEXT, api_key TEXT NOT NULL); INSERT INTO users VALUES (7, 'legacy-admin', 'synthetic-hash', 'synthetic-api-key')`); err != nil {
		t.Fatal(err)
	}
	source.Close()
	readonly, err := migration.OpenSQLiteReadOnly(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	defer readonly.Close()
	report, err := migration.Import(context.Background(), readonly, target)
	if err != nil {
		t.Fatalf("importing synthetic SQLite data: %v", err)
	}
	if !report.Reconciled {
		t.Fatalf("import report is not reconciled: %#v", report.Tables)
	}
	var username string
	if err := target.QueryRow(`SELECT username FROM users WHERE id = 7`).Scan(&username); err != nil || username != "legacy-admin" {
		t.Fatalf("validating imported user: username=%q err=%v", username, err)
	}
}

// TestPostgresOperationalPersistence covers IMAP, encryption, webhooks and
// report status persistence without connecting to any external endpoint.
func TestPostgresOperationalPersistence(t *testing.T) {
	setupPostgres(t)
	key, err := appcrypto.GenerateEncryptionKey()
	if err != nil {
		t.Fatal(err)
	}
	if err := appcrypto.InitEncryptionWithKey(key); err != nil {
		t.Fatal(err)
	}
	secret, err := appcrypto.Encrypt("synthetic-secret")
	if err != nil {
		t.Fatal(err)
	}
	smtp := models.SMTP{Name: "encrypted smtp", Host: "127.0.0.1:2525", FromAddress: "training@example.test", Password: secret, UserId: 1}
	if err := models.PostSMTP(&smtp); err != nil {
		t.Fatal(err)
	}
	im := models.IMAP{Host: "127.0.0.1", Port: 993, Username: "training", Password: secret, TLS: true, IMAPFreq: 60}
	if err := models.PostIMAP(&im, 1); err != nil {
		t.Fatal(err)
	}
	storedIMAP, err := models.GetIMAPById(im.Id, 1)
	if err != nil {
		t.Fatal(err)
	}
	if plain, err := appcrypto.Decrypt(storedIMAP.Password); err != nil || plain != "synthetic-secret" {
		t.Fatalf("decrypting IMAP password: %q %v", plain, err)
	}
	wrongKey, _ := appcrypto.GenerateEncryptionKey()
	appcrypto.InitEncryptionWithKey(wrongKey)
	if _, err := appcrypto.Decrypt(storedIMAP.Password); err == nil {
		t.Fatal("wrong encryption key decrypted persisted secret")
	}
	appcrypto.InitEncryptionWithKey(key)
	webhook := models.Webhook{Name: "persistence only", URL: "https://webhook.example.test/events", Secret: secret, IsActive: true}
	if err := models.PostWebhook(&webhook); err != nil {
		t.Fatal(err)
	}
	active, err := models.GetActiveWebhooks()
	if err != nil || len(active) == 0 {
		t.Fatalf("reading active webhooks: %v", err)
	}
	report := models.Report{UserId: 1, CampaignIds: "[]", Format: "json"}
	if err := models.PostReport(&report); err != nil {
		t.Fatal(err)
	}
	if err := models.UpdateReportStatus(report.Id, models.ReportStatusProcessing); err != nil {
		t.Fatal(err)
	}
	if err := models.UpdateReportStatus(report.Id, models.ReportStatusCompleted); err != nil {
		t.Fatal(err)
	}
	storedReport, err := models.GetReport(report.Id, 1)
	if err != nil || storedReport.Status != models.ReportStatusCompleted || storedReport.StartedAt == nil || storedReport.CompletedAt == nil {
		t.Fatalf("report status transition: %#v %v", storedReport, err)
	}
}
