package retry

import (
	"context"
	"fmt"
	"time"
)

// Config holds retry configuration
type Config struct {
	MaxAttempts int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
}

// DefaultConfig returns sensible retry defaults
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		InitialWait: 100 * time.Millisecond,
		MaxWait:     5 * time.Second,
		Multiplier:  2.0,
	}
}

// Do executes a function with exponential backoff retry
func Do(ctx context.Context, cfg Config, fn func(ctx context.Context) error) error {
	var lastErr error
	wait := cfg.InitialWait

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		// Try the operation
		if err := fn(ctx); err == nil {
			return nil // Success!
		} else {
			lastErr = err
		}

		// Don't wait after the last attempt
		if attempt == cfg.MaxAttempts {
			break
		}

		// Wait with exponential backoff
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(wait):
			// Calculate next wait time with exponential backoff
			wait = time.Duration(float64(wait) * cfg.Multiplier)
			if wait > cfg.MaxWait {
				wait = cfg.MaxWait
			}
		}
	}

	return fmt.Errorf("max retry attempts reached (%d): %w", cfg.MaxAttempts, lastErr)
}
