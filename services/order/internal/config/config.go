package config

import (
	"fmt"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog/log"
)

var (
	cfg  *Config
	once sync.Once
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port string `env:"PORT" env-default:"50051" yaml:"port"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string `env:"DB_HOST" env-required:"true" yaml:"host"`
	Port            int    `env:"DB_PORT" env-default:"5434" yaml:"port"`
	User            string `env:"DB_USER" env-required:"true" yaml:"user"`
	Password        string `env:"DB_PASSWORD" env-required:"true" yaml:"password"`
	DBName          string `env:"DB_NAME" env-required:"true" yaml:"db_name"`
	SSLMode         string `env:"DB_SSLMODE" env-default:"disable" yaml:"ssl_mode"`
	MaxOpenConns    int    `env:"MAX_OPEN_CONNS" env-default:"25" yaml:"max_open_conns"`
	MaxIdleConns    int    `env:"MAX_IDLE_CONNS" env-default:"5" yaml:"max_idle_conns"`
	ConnMaxLifetime int    `env:"CONN_MAX_LIFETIME" env-default:"5" yaml:"conn_max_lifetime"`
}

// Load loads and validates configuration from environment variables
func Init() error {
	var err error
	once.Do(func() {
		cfg = &Config{}

		if loadErr := cleanenv.ReadEnv(cfg); loadErr != nil {
			err = fmt.Errorf("failed to read config: %w", loadErr)
		}

		log.Info().Msg("✅ Configuration loaded")
	})
	return err
}

// Get returns the global config instance
func Get() *Config {
	if cfg == nil {
		panic("configuration not initialized, call config.Init() first")
	}
	return cfg
}

// Convenience getters for easy access
func Server() ServerConfig {
	return Get().Server
}

func Database() DatabaseConfig {
	return Get().Database
}
