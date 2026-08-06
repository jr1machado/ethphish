package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	log "github.com/gophish/gophish/logger"
)

var validConfig = []byte(`{
	"admin_server": {
		"listen_url": "127.0.0.1:3333",
		"use_tls": true,
		"cert_path": "gophish_admin.crt",
		"key_path": "gophish_admin.key"
	},
	"phish_server": {
		"listen_url": "0.0.0.0:8080",
		"use_tls": false,
		"cert_path": "example.crt",
		"key_path": "example.key"
	},
	"db_name": "postgres",
	"db_path": "host=postgres dbname=ethphish sslmode=disable",
	"migrations_prefix": "db/db_",
	"contact_address": ""
}`)

func createTemporaryConfig(t *testing.T) *os.File {
	f, err := ioutil.TempFile("", "gophish-config")
	if err != nil {
		t.Fatalf("unable to create temporary config: %v", err)
	}
	return f
}

func removeTemporaryConfig(t *testing.T, f *os.File) {
	err := f.Close()
	if err != nil {
		t.Fatalf("unable to remove temporary config: %v", err)
	}
}

func TestLoadConfig(t *testing.T) {
	f := createTemporaryConfig(t)
	defer removeTemporaryConfig(t, f)
	_, err := f.Write(validConfig)
	if err != nil {
		t.Fatalf("error writing config to temporary file: %v", err)
	}
	// Load the valid config
	conf, err := LoadConfig(f.Name())
	if err != nil {
		t.Fatalf("error loading config from temporary file: %v", err)
	}

	expectedConfig := &Config{}
	err = json.Unmarshal(validConfig, &expectedConfig)
	if err != nil {
		t.Fatalf("error unmarshaling config: %v", err)
	}
	expectedConfig.MigrationsPath = expectedConfig.MigrationsPath + expectedConfig.DBName + "/migrations"
	expectedConfig.TestFlag = false
	expectedConfig.AdminConf.CSRFKey = ""
	expectedConfig.Logging = &log.Config{}
	if !reflect.DeepEqual(expectedConfig, conf) {
		t.Fatalf("invalid config received. expected %#v got %#v", expectedConfig, conf)
	}

	// Load an invalid config
	_, err = LoadConfig("bogusfile")
	if err == nil {
		t.Fatalf("expected error when loading invalid config, but got %v", err)
	}
}

func TestLoadConfigEnvironmentOverrides(t *testing.T) {
	t.Setenv("ETHPHISH_ADMIN_LISTEN_URL", "0.0.0.0:3333")
	t.Setenv("ETHPHISH_ADMIN_USE_TLS", "false")
	t.Setenv("ETHPHISH_PHISH_LISTEN_URL", "0.0.0.0:8080")
	t.Setenv("ETHPHISH_REPORTS_STORAGE_PATH", "/var/lib/ethphish/reports")
	t.Setenv("ETHPHISH_DB_DRIVER", "postgres")
	t.Setenv("ETHPHISH_DB_DSN", "host=postgres dbname=ethphish sslmode=disable")
	t.Setenv("ETHPHISH_DB_MAX_OPEN_CONNECTIONS", "12")
	t.Setenv("ETHPHISH_DB_MAX_IDLE_CONNECTIONS", "4")
	t.Setenv("ETHPHISH_DB_CONNECTION_MAX_LIFETIME", "45m")

	f := createTemporaryConfig(t)
	defer removeTemporaryConfig(t, f)
	if _, err := f.Write(validConfig); err != nil {
		t.Fatalf("error writing config: %v", err)
	}
	conf, err := LoadConfig(f.Name())
	if err != nil {
		t.Fatalf("error loading config: %v", err)
	}
	if conf.AdminConf.ListenURL != "0.0.0.0:3333" || conf.AdminConf.UseTLS {
		t.Fatalf("admin environment overrides not applied: %#v", conf.AdminConf)
	}
	if conf.PhishConf.ListenURL != "0.0.0.0:8080" {
		t.Fatalf("phish environment override not applied: %s", conf.PhishConf.ListenURL)
	}
	if conf.ReportsConf.StoragePath != "/var/lib/ethphish/reports" {
		t.Fatalf("reports path not overridden: %s", conf.ReportsConf.StoragePath)
	}
	if conf.DBName != "postgres" || conf.DBPath != "host=postgres dbname=ethphish sslmode=disable" {
		t.Fatalf("database environment overrides not applied: %#v", conf)
	}
	if conf.DBMaxOpenConns != 12 || conf.DBMaxIdleConns != 4 || conf.DBConnMaxLife.String() != "45m0s" {
		t.Fatalf("database pool environment overrides not applied: %#v", conf)
	}
}

func TestLoadConfigRejectsDisabledPostgresTLSWhenRequired(t *testing.T) {
	f := createTemporaryConfig(t)
	defer removeTemporaryConfig(t, f)
	if _, err := f.Write(validConfig); err != nil {
		t.Fatalf("error writing config: %v", err)
	}
	t.Setenv("ETHPHISH_DB_DRIVER", "postgres")
	t.Setenv("ETHPHISH_DB_DSN", "host=db.example.test dbname=ethphish sslmode=disable")
	t.Setenv("ETHPHISH_DB_REQUIRE_TLS", "true")

	_, err := LoadConfig(f.Name())
	if err == nil || !strings.Contains(err.Error(), "PostgreSQL TLS is required") {
		t.Fatalf("expected PostgreSQL TLS validation error, got %v", err)
	}
}

func TestLoadConfigRejectsInvalidBooleanEnvironment(t *testing.T) {
	t.Setenv("ETHPHISH_ADMIN_USE_TLS", "sometimes")
	f := createTemporaryConfig(t)
	defer removeTemporaryConfig(t, f)
	if _, err := f.Write(validConfig); err != nil {
		t.Fatalf("error writing config: %v", err)
	}
	if _, err := LoadConfig(f.Name()); err == nil {
		t.Fatal("expected invalid boolean environment value to fail")
	}
}

func TestLoadConfigRejectsSQLite(t *testing.T) {
	f := createTemporaryConfig(t)
	defer removeTemporaryConfig(t, f)
	if _, err := f.Write(validConfig); err != nil {
		t.Fatalf("error writing config: %v", err)
	}
	t.Setenv("ETHPHISH_DB_DRIVER", "sqlite3")
	if _, err := LoadConfig(f.Name()); err == nil {
		t.Fatal("expected SQLite configuration to be rejected")
	}
}

func TestLoadConfigRejectsInvalidDatabasePoolEnvironment(t *testing.T) {
	t.Setenv("ETHPHISH_DB_MAX_OPEN_CONNECTIONS", "not-a-number")
	f := createTemporaryConfig(t)
	defer removeTemporaryConfig(t, f)
	if _, err := f.Write(validConfig); err != nil {
		t.Fatalf("error writing config: %v", err)
	}
	if _, err := LoadConfig(f.Name()); err == nil {
		t.Fatal("expected invalid database pool value to fail")
	}
}
