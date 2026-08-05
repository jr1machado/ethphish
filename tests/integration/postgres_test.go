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

// TestPostgresMigrationsAndReadiness verifies the deployment path used in
// development and CI: a fresh PostgreSQL database accepts every migration and
// is available through the application readiness primitive. It is opt-in for
// local development to keep the normal unit-test suite self-contained.
func TestPostgresMigrationsAndReadiness(t *testing.T) {
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
	if _, err := models.GetUserByUsername(models.DefaultAdminUsername); err != nil {
		t.Fatalf("checking migrated administrative user: %v", err)
	}
}
