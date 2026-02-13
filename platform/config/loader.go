package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("read env: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
