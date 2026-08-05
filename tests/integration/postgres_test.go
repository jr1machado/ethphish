package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gophish/gophish/config"
	tenantctx "github.com/gophish/gophish/context"
	appcrypto "github.com/gophish/gophish/crypto"
	mid "github.com/gophish/gophish/middleware"
	"github.com/gophish/gophish/migration"
	"github.com/gophish/gophish/models"
	_ "github.com/mattn/go-sqlite3"
)

const postgresDSNEnv = "ETHPHISH_TEST_POSTGRES_DSN"
const postgresMigrationsPathEnv = "ETHPHISH_TEST_MIGRATIONS_PATH"

func postgresMigrationsPath(t *testing.T) string {
	t.Helper()
	if path := os.Getenv(postgresMigrationsPathEnv); path != "" {
		return path
	}
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locating integration test source")
	}
	return filepath.Join(filepath.Dir(sourceFile), "..", "..", "db", "db_postgres", "migrations")
}

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

	migrationsPath := postgresMigrationsPath(t)

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

// TestPostgresTenantFoundation verifies that the Sprint 04 tenant, company,
// and user-grant relations are persisted by PostgreSQL with their intended
// boundaries. It does not send messages or contact any external service.
func TestPostgresTenantFoundation(t *testing.T) {
	setupPostgres(t)

	suffix := time.Now().UTC().Format("20060102150405.000000000")
	tenant := models.Tenant{Slug: "integration-tenant-" + suffix, Name: "Integration tenant " + suffix, Active: true}
	if err := models.PostTenant(&tenant); err != nil {
		t.Fatalf("creating tenant: %v", err)
	}
	company := models.Company{TenantID: tenant.ID, Name: "Integration company " + suffix}
	if err := models.PostCompany(&company); err != nil {
		t.Fatalf("creating company: %v", err)
	}
	grant := models.TenantUser{
		TenantID:  tenant.ID,
		UserID:    1,
		CompanyID: &company.ID,
		Role:      "tenant_admin",
	}
	if err := models.GrantTenantUser(&grant); err != nil {
		t.Fatalf("granting tenant access: %v", err)
	}

	storedTenant, err := models.GetTenantBySlug(strings.ToUpper(tenant.Slug))
	if err != nil || storedTenant.ID != tenant.ID || !storedTenant.Active {
		t.Fatalf("stored tenant = %#v, err = %v", storedTenant, err)
	}
	storedGrant, err := models.GetTenantUser(tenant.ID, 1)
	if err != nil || storedGrant.CompanyID == nil || *storedGrant.CompanyID != company.ID || storedGrant.Role != "tenant_admin" {
		t.Fatalf("stored tenant grant = %#v, err = %v", storedGrant, err)
	}
	grants, err := models.GetTenantUsers(1)
	if err != nil || len(grants) == 0 {
		t.Fatalf("user tenant grants = %#v, err = %v", grants, err)
	}
	template := models.Template{
		TenantID: tenant.ID,
		UserId:   1,
		Name:     "Tenant-scoped template " + suffix,
		Subject:  "Authorized simulation",
		Text:     "Training message",
	}
	if err := models.PostTemplate(&template); err != nil {
		t.Fatalf("creating tenant-scoped template: %v", err)
	}
	storedTemplate, err := models.GetTemplate(template.Id, 1)
	if err != nil || storedTemplate.TenantID != tenant.ID {
		t.Fatalf("stored tenant-scoped template = %#v, err = %v", storedTemplate, err)
	}
	otherTenant := models.Tenant{Slug: "integration-other-tenant-" + suffix, Name: "Integration other tenant " + suffix, Active: true}
	if err := models.PostTenant(&otherTenant); err != nil {
		t.Fatalf("creating second tenant: %v", err)
	}
	if err := models.GrantTenantUser(&models.TenantUser{TenantID: otherTenant.ID, UserID: 1, Role: "tenant_admin"}); err != nil {
		t.Fatalf("granting second tenant access: %v", err)
	}
	otherTemplate := models.Template{
		TenantID: otherTenant.ID,
		UserId:   1,
		Name:     "Other tenant template " + suffix,
		Subject:  "Authorized simulation",
		Text:     "Training message",
	}
	if err := models.PostTemplate(&otherTemplate); err != nil {
		t.Fatalf("creating second tenant template: %v", err)
	}
	if _, err := models.GetTemplateForTenant(otherTemplate.Id, tenant.ID, 1); err == nil {
		t.Fatal("tenant-scoped lookup returned another tenant's template")
	}
	page := models.Page{TenantID: tenant.ID, UserId: 1, Name: "Tenant page " + suffix, HTML: "<p>Training</p>"}
	if err := models.PostPageForTenant(&page, tenant.ID); err != nil {
		t.Fatalf("creating tenant-scoped page: %v", err)
	}
	if _, err := models.GetPageForTenant(page.Id, otherTenant.ID, 1); err == nil {
		t.Fatal("tenant-scoped lookup returned another tenant's page")
	}

	protected := mid.ResolveTenantScope(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope, err := tenantctx.RequireTenantScope(r)
		if err != nil || scope.TenantID != tenant.ID || scope.UserID != 1 {
			http.Error(w, "unexpected tenant scope", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(mid.TenantIDHeader, fmt.Sprintf("%d", tenant.ID))
	req = tenantctx.Set(req, "user", models.User{Id: 1})
	response := httptest.NewRecorder()
	protected.ServeHTTP(response, req)
	if response.Code != http.StatusNoContent {
		t.Fatalf("authorized tenant scope returned %d: %s", response.Code, response.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(mid.TenantIDHeader, fmt.Sprintf("%d", tenant.ID+999))
	req = tenantctx.Set(req, "user", models.User{Id: 1})
	response = httptest.NewRecorder()
	protected.ServeHTTP(response, req)
	if response.Code != http.StatusForbidden {
		t.Fatalf("cross-tenant request returned %d, want %d", response.Code, http.StatusForbidden)
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
	migrationsPath := postgresMigrationsPath(t)
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
