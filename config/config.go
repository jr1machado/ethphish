package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/gophish/gophish/logger"
)

// AdminServer represents the Admin server configuration details
type AdminServer struct {
	ListenURL            string   `json:"listen_url"`
	UseTLS               bool     `json:"use_tls"`
	CertPath             string   `json:"cert_path"`
	KeyPath              string   `json:"key_path"`
	CSRFKey              string   `json:"csrf_key"`
	AllowedInternalHosts []string `json:"allowed_internal_hosts"`
}

// PhishServer represents the Phish server configuration details
type PhishServer struct {
	ListenURL string `json:"listen_url"`
	UseTLS    bool   `json:"use_tls"`
	CertPath  string `json:"cert_path"`
	KeyPath   string `json:"key_path"`
}

// Reports represents the reports configuration details
type Reports struct {
	StoragePath string `json:"storage_path"`
}

// OIDC holds optional OpenID Connect settings for admin UI login.
type OIDC struct {
	Enabled              bool   `json:"enabled"`
	Issuer               string `json:"issuer"`
	ClientID             string `json:"client_id"`
	RedirectURL          string `json:"redirect_url"`
	RequiredGroup        string `json:"required_group"`
	GroupsClaim          string `json:"groups_claim"`
	UsernameFromEmail    string `json:"username_from_email"`
	AllowUnverifiedEmail bool   `json:"allow_unverified_email"`
}

// Config represents the configuration information.
type Config struct {
	AdminConf      AdminServer   `json:"admin_server"`
	PhishConf      PhishServer   `json:"phish_server"`
	DBName         string        `json:"db_name"`
	DBPath         string        `json:"db_path"`
	DBSSLCaPath    string        `json:"db_sslca_path"`
	DBRequireTLS   bool          `json:"db_require_tls"`
	DBMaxOpenConns int           `json:"db_max_open_connections"`
	DBMaxIdleConns int           `json:"db_max_idle_connections"`
	DBConnMaxLife  time.Duration `json:"db_connection_max_lifetime"`
	MigrationsPath string        `json:"migrations_prefix"`
	TestFlag       bool          `json:"test_flag"`
	ContactAddress string        `json:"contact_address"`
	Logging        *log.Config   `json:"logging"`
	ReportsConf    Reports       `json:"reports"`
	OIDC           OIDC          `json:"oidc"`
}

// Version contains the current gophish version
var Version = ""

// AnglerPhishVersion contains the current anglerphish version
var AnglerPhishVersion = ""

// LoadConfig loads the configuration from the specified filepath
func LoadConfig(filepath string) (*Config, error) {
	// Get the config file
	configFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = json.Unmarshal(configFile, config)
	if err != nil {
		return nil, err
	}
	if config.Logging == nil {
		config.Logging = &log.Config{}
	}
	if err := applyEnvironment(config); err != nil {
		return nil, err
	}
	if config.DBRequireTLS && config.DBName == "postgres" && postgresTLSDisabled(config.DBPath) {
		return nil, fmt.Errorf("PostgreSQL TLS is required but db_path sets sslmode=disable")
	}
	if config.DBName != "postgres" {
		return nil, fmt.Errorf("only PostgreSQL is supported by the server runtime; configure ETHPHISH_DB_DRIVER=postgres")
	}
	// Choosing the migrations directory based on the database used.
	config.MigrationsPath = config.MigrationsPath + config.DBName + "/migrations"
	// Explicitly set the TestFlag to false to prevent config.json overrides
	config.TestFlag = false
	return config, nil
}

// applyEnvironment overlays deployment-specific values without modifying the
// checked-in configuration file. Empty variables intentionally keep the file
// value, preventing an incomplete environment from erasing safe defaults.
func applyEnvironment(config *Config) error {
	setString := func(name string, target *string) {
		if value, ok := os.LookupEnv(name); ok && value != "" {
			*target = value
		}
	}
	setBool := func(name string, target *bool) error {
		value, ok := os.LookupEnv(name)
		if !ok || value == "" {
			return nil
		}
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean in %s: %w", name, err)
		}
		*target = parsed
		return nil
	}
	setInt := func(name string, target *int) error {
		value, ok := os.LookupEnv(name)
		if !ok || value == "" {
			return nil
		}
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			return fmt.Errorf("invalid integer in %s", name)
		}
		*target = parsed
		return nil
	}
	setDuration := func(name string, target *time.Duration) error {
		value, ok := os.LookupEnv(name)
		if !ok || value == "" {
			return nil
		}
		parsed, err := time.ParseDuration(value)
		if err != nil || parsed < 0 {
			return fmt.Errorf("invalid duration in %s", name)
		}
		*target = parsed
		return nil
	}

	setString("ETHPHISH_ADMIN_LISTEN_URL", &config.AdminConf.ListenURL)
	setString("ETHPHISH_ADMIN_CERT_PATH", &config.AdminConf.CertPath)
	setString("ETHPHISH_ADMIN_KEY_PATH", &config.AdminConf.KeyPath)
	setString("ETHPHISH_ADMIN_CSRF_KEY", &config.AdminConf.CSRFKey)
	setString("ETHPHISH_PHISH_LISTEN_URL", &config.PhishConf.ListenURL)
	setString("ETHPHISH_PHISH_CERT_PATH", &config.PhishConf.CertPath)
	setString("ETHPHISH_PHISH_KEY_PATH", &config.PhishConf.KeyPath)
	setString("ETHPHISH_DB_DRIVER", &config.DBName)
	setString("ETHPHISH_DB_DSN", &config.DBPath)
	setString("ETHPHISH_DB_SSL_CA_PATH", &config.DBSSLCaPath)
	setString("ETHPHISH_CONTACT_ADDRESS", &config.ContactAddress)
	setString("ETHPHISH_REPORTS_STORAGE_PATH", &config.ReportsConf.StoragePath)
	setString("ETHPHISH_OIDC_ISSUER", &config.OIDC.Issuer)
	setString("ETHPHISH_OIDC_CLIENT_ID", &config.OIDC.ClientID)
	setString("ETHPHISH_OIDC_REDIRECT_URL", &config.OIDC.RedirectURL)
	setString("ETHPHISH_OIDC_REQUIRED_GROUP", &config.OIDC.RequiredGroup)
	if err := setInt("ETHPHISH_DB_MAX_OPEN_CONNECTIONS", &config.DBMaxOpenConns); err != nil {
		return err
	}
	if err := setInt("ETHPHISH_DB_MAX_IDLE_CONNECTIONS", &config.DBMaxIdleConns); err != nil {
		return err
	}
	if err := setDuration("ETHPHISH_DB_CONNECTION_MAX_LIFETIME", &config.DBConnMaxLife); err != nil {
		return err
	}

	if err := setBool("ETHPHISH_ADMIN_USE_TLS", &config.AdminConf.UseTLS); err != nil {
		return err
	}
	if err := setBool("ETHPHISH_PHISH_USE_TLS", &config.PhishConf.UseTLS); err != nil {
		return err
	}
	if err := setBool("ETHPHISH_OIDC_ENABLED", &config.OIDC.Enabled); err != nil {
		return err
	}
	if err := setBool("ETHPHISH_DB_REQUIRE_TLS", &config.DBRequireTLS); err != nil {
		return err
	}
	if value, ok := os.LookupEnv("ETHPHISH_ALLOWED_INTERNAL_HOSTS"); ok {
		config.AdminConf.AllowedInternalHosts = splitList(value)
	}
	return nil
}

func postgresTLSDisabled(dsn string) bool {
	for _, option := range strings.Fields(dsn) {
		if strings.EqualFold(option, "sslmode=disable") {
			return true
		}
	}
	return false
}

func splitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
