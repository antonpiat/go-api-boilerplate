package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDefaultsAndEnvOverride(t *testing.T) {
	t.Setenv("DATABASE_PASSWORD", "secret-from-env")
	t.Setenv("JWT_ACCESS_SECRET", "access-from-env")
	t.Setenv("JWT_REFRESH_SECRET", "refresh-from-env")
	t.Setenv("APP_ENVIRONMENT", "development")
	t.Setenv("LOG_OUTPUTS", "console,file")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`
app:
  name: test-api
  environment: development
jwt:
  access_secret: change-me-access
  refresh_secret: change-me-refresh
  access_ttl: 15m
  refresh_ttl: 168h
`), 0o644)
	require.NoError(t, err)

	cfg, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, "test-api", cfg.App.Name)
	require.Equal(t, "secret-from-env", cfg.Database.Password)
	require.Equal(t, "access-from-env", cfg.JWT.AccessSecret)
	require.Equal(t, []string{"console", "file"}, cfg.Logging.Outputs)
}

func TestValidateRejectsDefaultJWTInProduction(t *testing.T) {
	cfg := DefaultConfig()
	cfg.App.Environment = "production"
	err := cfg.Validate()
	require.Error(t, err)
}

func TestDSN(t *testing.T) {
	cfg := DefaultConfig().Database
	require.Contains(t, cfg.DSN(), "postgres://postgres:postgres@localhost:5432/gin-api")
}
