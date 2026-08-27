package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	CORS     CORSConfig     `yaml:"cors"`
	Logging  LoggingConfig  `yaml:"logging"`
	Metrics  MetricsConfig  `yaml:"metrics"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
}

func (c AppConfig) IsProduction() bool {
	return strings.EqualFold(c.Environment, "production") || c.Environment == "prod"
}

type ServerConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	ReadTimeout     string `yaml:"read_timeout"`
	WriteTimeout    string `yaml:"write_timeout"`
	IdleTimeout     string `yaml:"idle_timeout"`
	ShutdownTimeout string `yaml:"shutdown_timeout"`
	MaxBodyBytes    int64  `yaml:"max_body_bytes"`
}

func (c ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c ServerConfig) ReadTimeoutDuration() time.Duration {
	return parseDuration(c.ReadTimeout, 10*time.Second)
}

func (c ServerConfig) WriteTimeoutDuration() time.Duration {
	return parseDuration(c.WriteTimeout, 10*time.Second)
}

func (c ServerConfig) IdleTimeoutDuration() time.Duration {
	return parseDuration(c.IdleTimeout, 60*time.Second)
}

func (c ServerConfig) ShutdownTimeoutDuration() time.Duration {
	return parseDuration(c.ShutdownTimeout, 15*time.Second)
}

type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Name            string `yaml:"name"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	SSLMode         string `yaml:"sslmode"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime"`
}

func (c DatabaseConfig) DSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		sslMode,
	)
}

func (c DatabaseConfig) ConnMaxLifetimeDuration() time.Duration {
	return parseDuration(c.ConnMaxLifetime, time.Hour)
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type JWTConfig struct {
	AccessSecret  string `yaml:"access_secret"`
	RefreshSecret string `yaml:"refresh_secret"`
	AccessTTL     string `yaml:"access_ttl"`
	RefreshTTL    string `yaml:"refresh_ttl"`
}

func (c JWTConfig) AccessTTLDuration() time.Duration {
	return parseDuration(c.AccessTTL, 15*time.Minute)
}

func (c JWTConfig) RefreshTTLDuration() time.Duration {
	return parseDuration(c.RefreshTTL, 7*24*time.Hour)
}

type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins"`
	AllowMethods     []string `yaml:"allow_methods"`
	AllowHeaders     []string `yaml:"allow_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
}

type LoggingConfig struct {
	Level     string   `yaml:"level"`
	Directory string   `yaml:"directory"`
	Outputs   []string `yaml:"outputs"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:        "gin-api",
			Environment: "development",
		},
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			ReadTimeout:     "10s",
			WriteTimeout:    "10s",
			IdleTimeout:     "60s",
			ShutdownTimeout: "15s",
			MaxBodyBytes:    1 << 20,
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			Name:            "gin-api",
			Username:        "postgres",
			Password:        "postgres",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    10,
			ConnMaxLifetime: "1h",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
		JWT: JWTConfig{
			AccessSecret:  "change-me-access",
			RefreshSecret: "change-me-refresh",
			AccessTTL:     "15m",
			RefreshTTL:    "168h",
		},
		CORS: CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
			AllowCredentials: false,
		},
		Logging: LoggingConfig{
			Level:     "info",
			Directory: "logs",
			Outputs:   []string{"console"},
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
	}
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}
	if path == "" {
		path = "config.yaml"
	}

	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("read config: %w", err)
		}
	} else if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	applyEnv(cfg)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if _, err := time.ParseDuration(c.Server.ReadTimeout); err != nil {
		return fmt.Errorf("server.read_timeout: %w", err)
	}
	if _, err := time.ParseDuration(c.Server.WriteTimeout); err != nil {
		return fmt.Errorf("server.write_timeout: %w", err)
	}
	if _, err := time.ParseDuration(c.Server.IdleTimeout); err != nil {
		return fmt.Errorf("server.idle_timeout: %w", err)
	}
	if _, err := time.ParseDuration(c.Server.ShutdownTimeout); err != nil {
		return fmt.Errorf("server.shutdown_timeout: %w", err)
	}
	if _, err := time.ParseDuration(c.Database.ConnMaxLifetime); err != nil {
		return fmt.Errorf("database.conn_max_lifetime: %w", err)
	}
	if _, err := time.ParseDuration(c.JWT.AccessTTL); err != nil {
		return fmt.Errorf("jwt.access_ttl: %w", err)
	}
	if _, err := time.ParseDuration(c.JWT.RefreshTTL); err != nil {
		return fmt.Errorf("jwt.refresh_ttl: %w", err)
	}
	if c.JWT.AccessSecret == "" || c.JWT.RefreshSecret == "" {
		return fmt.Errorf("jwt secrets must not be empty")
	}
	if c.App.IsProduction() {
		if c.JWT.AccessSecret == "change-me-access" || c.JWT.RefreshSecret == "change-me-refresh" {
			return fmt.Errorf("jwt secrets must be changed in production")
		}
	}
	if c.Server.MaxBodyBytes <= 0 {
		return fmt.Errorf("server.max_body_bytes must be positive")
	}
	if c.Metrics.Path == "" {
		c.Metrics.Path = "/metrics"
	}
	return nil
}

func applyEnv(cfg *Config) {
	setString(&cfg.App.Name, "APP_NAME")
	setString(&cfg.App.Environment, "APP_ENVIRONMENT")

	setString(&cfg.Server.Host, "SERVER_HOST")
	setInt(&cfg.Server.Port, "SERVER_PORT")
	setString(&cfg.Server.ReadTimeout, "SERVER_READ_TIMEOUT")
	setString(&cfg.Server.WriteTimeout, "SERVER_WRITE_TIMEOUT")
	setString(&cfg.Server.IdleTimeout, "SERVER_IDLE_TIMEOUT")
	setString(&cfg.Server.ShutdownTimeout, "SERVER_SHUTDOWN_TIMEOUT")
	setInt64(&cfg.Server.MaxBodyBytes, "SERVER_MAX_BODY_BYTES")

	setString(&cfg.Database.Host, "DATABASE_HOST")
	setInt(&cfg.Database.Port, "DATABASE_PORT")
	setString(&cfg.Database.Name, "DATABASE_NAME")
	setString(&cfg.Database.Username, "DATABASE_USERNAME")
	setString(&cfg.Database.Password, "DATABASE_PASSWORD")
	setString(&cfg.Database.SSLMode, "DATABASE_SSLMODE")
	setInt(&cfg.Database.MaxOpenConns, "DATABASE_MAX_OPEN_CONNS")
	setInt(&cfg.Database.MaxIdleConns, "DATABASE_MAX_IDLE_CONNS")
	setString(&cfg.Database.ConnMaxLifetime, "DATABASE_CONN_MAX_LIFETIME")

	setString(&cfg.Redis.Host, "REDIS_HOST")
	setInt(&cfg.Redis.Port, "REDIS_PORT")
	setString(&cfg.Redis.Password, "REDIS_PASSWORD")
	setInt(&cfg.Redis.DB, "REDIS_DB")

	setString(&cfg.JWT.AccessSecret, "JWT_ACCESS_SECRET")
	setString(&cfg.JWT.RefreshSecret, "JWT_REFRESH_SECRET")
	setString(&cfg.JWT.AccessTTL, "JWT_ACCESS_TTL")
	setString(&cfg.JWT.RefreshTTL, "JWT_REFRESH_TTL")

	setString(&cfg.Logging.Level, "LOG_LEVEL")
	setString(&cfg.Logging.Directory, "LOG_DIRECTORY")
	if v := os.Getenv("LOG_OUTPUTS"); v != "" {
		cfg.Logging.Outputs = splitCSV(v)
	}

	if v := os.Getenv("CORS_ALLOW_ORIGINS"); v != "" {
		cfg.CORS.AllowOrigins = splitCSV(v)
	}

	setBool(&cfg.Metrics.Enabled, "METRICS_ENABLED")
	setString(&cfg.Metrics.Path, "METRICS_PATH")
}

func setString(dst *string, key string) {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		*dst = v
	}
}

func setInt(dst *int, key string) {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			*dst = n
		}
	}
}

func setInt64(dst *int64, key string) {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			*dst = n
		}
	}
}

func setBool(dst *bool, key string) {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			*dst = b
		}
	}
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseDuration(value string, fallback time.Duration) time.Duration {
	if value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}
