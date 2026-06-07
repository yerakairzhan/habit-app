// Package config loads and validates application configuration.
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all runtime configuration loaded from env / config file.
type Config struct {
	Server   ServerConfig   `mapstructure:",squash"`
	Database DatabaseConfig `mapstructure:",squash"`
	JWT      JWTConfig      `mapstructure:",squash"`
}

type ServerConfig struct {
	Port            int    `mapstructure:"server_port"`
	GracefulTimeout int    `mapstructure:"server_graceful_timeout_seconds"`
	Env             string `mapstructure:"server_env"` // "development" | "production"
}

type DatabaseConfig struct {
	Host     string `mapstructure:"database_host"`
	Port     int    `mapstructure:"database_port"`
	Name     string `mapstructure:"database_name"`
	User     string `mapstructure:"database_user"`
	Password string `mapstructure:"database_password"`
	SSLMode  string `mapstructure:"database_ssl_mode"`
	MaxConns int    `mapstructure:"database_max_conns"` // Switched back to int
}

type JWTConfig struct {
	Secret string `mapstructure:"jwt_secret"`
}

// Load reads configuration from environment variables (prefixed HABIT_) and
// an optional config.yaml file in the working directory.
func Load() (*Config, error) {
	v := viper.New()

	// Defaults match the flat keys
	v.SetDefault("server_port", 50051)
	v.SetDefault("server_graceful_timeout_seconds", 10)
	v.SetDefault("server_env", "production")
	v.SetDefault("database_host", "localhost")
	v.SetDefault("database_port", 5432)
	v.SetDefault("database_ssl_mode", "disable")
	v.SetDefault("database_max_conns", 20)

	// Config file (optional)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/habit-tracker/")
	_ = v.ReadInConfig() // not fatal if missing

	// Environment variable binding
	v.SetEnvPrefix("HABIT")
	v.AutomaticEnv()

	// CRITICAL FIX: Explicitly bind the keys so Viper maps "HABIT_JWT_SECRET" to "jwt_secret"
	_ = v.BindEnv("jwt_secret")
	_ = v.BindEnv("database_name")
	_ = v.BindEnv("database_user")
	_ = v.BindEnv("database_password")
	_ = v.BindEnv("database_host")
	_ = v.BindEnv("server_port")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config unmarshal: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}
	return &cfg, nil
}

func validate(c *Config) error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("HABIT_JWT_SECRET must be set")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("HABIT_JWT_SECRET must be at least 32 characters")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("HABIT_DATABASE_NAME must be set")
	}
	if c.Database.User == "" {
		return fmt.Errorf("HABIT_DATABASE_USER must be set")
	}
	return nil
}

// DSN returns the PostgreSQL data source name.
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}
