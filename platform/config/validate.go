package config

import "fmt"

func validate(c *Config) error {

	if c.Postgres.MaxOpenConns <= 0 {
		return fmt.Errorf("MAX_OPEN_CONNS must be > 0")
	}

	if c.Postgres.MaxIdleConns < 0 {
		return fmt.Errorf("MAX_IDLE_CONNS cannot be negative")
	}

	if c.App.Env == "production" {

		if c.Postgres.SSLMode != "require" {
			return fmt.Errorf("production requires PG_SSLMODE=require")
		}

	}

	return nil
}
