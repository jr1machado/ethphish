package models

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/gophish/gophish/auth"
	"github.com/gophish/gophish/config"
	dbmigration "github.com/gophish/gophish/migration"

	log "github.com/gophish/gophish/logger"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq" // Blank import needed to import PostgreSQL
)

var db *gorm.DB
var conf *config.Config

const MaxDatabaseConnectionAttempts int = 10

// postgresMigrationLockID serializes Goose migrations across application
// instances that share a PostgreSQL database.
const postgresMigrationLockID int64 = 0x4554485048495348

// DefaultAdminUsername is the default username for the administrative user
const DefaultAdminUsername = "admin"

// InitialAdminPassword is the environment variable that specifies which
// password to use for the initial root login instead of generating one
// randomly
const InitialAdminPassword = "GOPHISH_INITIAL_ADMIN_PASSWORD"

// InitialAdminApiToken is the environment variable that specifies the
// API token to seed the initial root login instead of generating one
// randomly
const InitialAdminApiToken = "GOPHISH_INITIAL_ADMIN_API_TOKEN"

const (
	CampaignInProgress    string = "In progress"
	CampaignQueued        string = "Queued"
	CampaignCreated       string = "Created"
	CampaignEmailsSent    string = "Emails Sent"
	CampaignComplete      string = "Completed"
	EventSent             string = "Email Sent"
	EventSMSSent          string = "SMS Sent"
	EventSendingError     string = "Error Sending Email"
	EventSMSError         string = "Error Sending SMS"
	EventOpened           string = "Email Opened"
	EventClicked          string = "Clicked Link"
	EventDataSubmit       string = "Submitted Data"
	EventReported         string = "Email Reported"
	EventReplied          string = "Email Replied"
	EventProxyRequest     string = "Proxied request"
	EventMFACodeSent      string = "MFA Code Sent"
	EventMFACodeSendError string = "MFA Code Send Error"
	EventMFACodeVerified  string = "MFA Code Verified"
	EventMFACodeFailed    string = "MFA Code Failed"
	StatusSuccess         string = "Success"
	StatusQueued          string = "Queued"
	StatusSending         string = "Sending"
	StatusUnknown         string = "Unknown"
	StatusScheduled       string = "Scheduled"
	StatusRetry           string = "Retrying"
	Error                 string = "Error"
)

// ErrInvalidCampaignID is thrown when a campaign ID is provided that doesn't match
// the expected campaign ID
var ErrInvalidCampaignID = fmt.Errorf("incorrect campaign provided for caching")

// Flash is used to hold flash information for use in templates.
type Flash struct {
	Type    string
	Message string
}

// Response contains the attributes found in an API response
type Response struct {
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// Copy of auth.GenerateSecureKey to prevent cyclic import with auth library
func generateSecureKey() string {
	k := make([]byte, 32)
	io.ReadFull(rand.Reader, k)
	return fmt.Sprintf("%x", k)
}

func createTemporaryPassword(u *User) error {
	var temporaryPassword string
	if envPassword := os.Getenv(InitialAdminPassword); envPassword != "" {
		temporaryPassword = envPassword
	} else {
		// This will result in a 16 character password which could be viewed as an
		// inconvenience, but it should be ok for now.
		temporaryPassword = auth.GenerateSecureKey(auth.MinPasswordLength)
	}
	hash, err := auth.GeneratePasswordHash(temporaryPassword)
	if err != nil {
		return err
	}
	u.Hash = hash
	// Anytime a temporary password is created, we will force the user
	// to change their password
	u.PasswordChangeRequired = true
	err = db.Save(u).Error
	if err != nil {
		return err
	}
	log.Infof("Please login with the username admin and the password %s", temporaryPassword)
	return nil
}

// Setup initializes the database and runs any needed migrations.
//
// First, it establishes a connection to the database, then runs any migrations
// newer than the version the database is on.
//
// Once the database is up-to-date, we create an admin user (if needed) that
// has a randomly generated API key and password.
func Setup(c *config.Config) error {
	// Setup the package-scoped config
	conf = c
	// Get the latest possible migration
	latest, err := dbmigration.Latest(conf.MigrationsPath)
	if err != nil {
		log.Error(err)
		return err
	}

	// Register certificates for tls encrypted db connections
	if conf.DBSSLCaPath != "" {
		switch conf.DBName {
		case "mysql":
			rootCertPool := x509.NewCertPool()
			pem, err := ioutil.ReadFile(conf.DBSSLCaPath)
			if err != nil {
				log.Error(err)
				return err
			}
			if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
				log.Error("Failed to append PEM.")
				return err
			}
			mysql.RegisterTLSConfig("ssl_ca", &tls.Config{
				RootCAs: rootCertPool,
			})
			// Default database is sqlite3, which supports no tls, as connection
			// is file based
		default:
		}
	}

	// Open our database connection
	i := 0
	for {
		db, err = gorm.Open(conf.DBName, conf.DBPath)
		if err == nil {
			break
		}
		if err != nil && i >= MaxDatabaseConnectionAttempts {
			log.Error(err)
			return err
		}
		i += 1
		log.Warn("waiting for database to be up...")
		time.Sleep(5 * time.Second)
	}
	db.LogMode(false)
	db.SetLogger(log.Logger)
	if err != nil {
		log.Error(err)
		return err
	}
	// Migrate up to the latest version
	err = runMigrations(conf.DBName, conf.MigrationsPath, latest, db.DB())
	if err != nil {
		log.Error(err)
		return err
	}
	configureConnectionPool(db.DB(), conf)
	// Ensure preset URL templates exist
	err = EnsurePresetURLTemplates(db)
	if err != nil {
		log.Error(err)
		return err
	}
	// Create the admin user if it doesn't exist
	var userCount int64
	var adminUser User
	db.Model(&User{}).Count(&userCount)
	adminRole, err := GetRoleBySlug(RoleAdmin)
	if err != nil {
		log.Error(err)
		return err
	}
	if userCount == 0 {
		adminUser := User{
			Username:               DefaultAdminUsername,
			Role:                   adminRole,
			RoleID:                 adminRole.ID,
			PasswordChangeRequired: true,
		}

		if envToken := os.Getenv(InitialAdminApiToken); envToken != "" {
			adminUser.ApiKey = envToken
		} else {
			adminUser.ApiKey = auth.GenerateSecureKey(auth.APIKeyLength)
		}

		err = PutUser(&adminUser)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	// If this is the first time the user is installing Gophish, then we will
	// generate a temporary password for the admin user.
	//
	// We do this here instead of in the block above where the admin is created
	// since there's the chance the user executes Gophish and has some kind of
	// error, then tries restarting it. If they didn't grab the password out of
	// the logs, then they would have lost it.
	//
	// By doing the temporary password here, we will regenerate that temporary
	// password until the user is able to reset the admin password.
	if adminUser.Username == "" {
		adminUser, err = GetUserByUsername(DefaultAdminUsername)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	if adminUser.PasswordChangeRequired {
		err = createTemporaryPassword(&adminUser)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

func runMigrations(driver, migrationsPath string, latest int64, database *sql.DB) error {
	if driver != "postgres" {
		return dbmigration.Apply(context.Background(), driver, migrationsPath, latest, database)
	}

	// An advisory lock makes concurrent starts wait for the migration owner.
	// The temporary one-connection pool guarantees that Goose uses the session
	// holding the lock.
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	if _, err := database.Exec("SELECT pg_advisory_lock($1)", postgresMigrationLockID); err != nil {
		return fmt.Errorf("acquiring PostgreSQL migration lock: %w", err)
	}
	defer func() {
		if _, err := database.Exec("SELECT pg_advisory_unlock($1)", postgresMigrationLockID); err != nil {
			log.Errorf("releasing PostgreSQL migration lock: %v", err)
		}
	}()
	return dbmigration.Apply(context.Background(), driver, migrationsPath, latest, database)
}

func configureConnectionPool(database *sql.DB, config *config.Config) {
	maxOpen := config.DBMaxOpenConns
	maxIdle := config.DBMaxIdleConns
	maxLifetime := config.DBConnMaxLife
	if config.DBName == "sqlite3" {
		maxOpen = 1
		// An in-memory SQLite database exists per connection. Keep the single
		// connection idle between Goose statements so the schema is retained.
		maxIdle = 1
	} else {
		if maxOpen == 0 {
			maxOpen = 10
		}
		if maxIdle == 0 {
			maxIdle = 5
		}
		if maxLifetime == 0 {
			maxLifetime = 30 * time.Minute
		}
	}
	database.SetMaxOpenConns(maxOpen)
	database.SetMaxIdleConns(maxIdle)
	database.SetConnMaxLifetime(maxLifetime)
}

// Ping reports whether the currently configured database is reachable. It is
// used by readiness checks and never exposes connection details.
func Ping() error {
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return db.DB().Ping()
}
