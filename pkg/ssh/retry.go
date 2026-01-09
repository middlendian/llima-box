package ssh

import (
	"fmt"
	"time"
)

// RetryConfig configures retry behavior for SSH operations
type RetryConfig struct {
	// MaxAttempts is the maximum number of attempts (default: 3)
	MaxAttempts int
	// InitialDelay is the delay before the first retry (default: 1s)
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries (default: 10s)
	MaxDelay time.Duration
	// Multiplier is the backoff multiplier (default: 2)
	Multiplier float64
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}
}

// ConnectWithRetry attempts to connect with exponential backoff
func (c *Client) ConnectWithRetry(config RetryConfig) error {
	var lastErr error

	delay := config.InitialDelay
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := c.Connect()
		if err == nil {
			return nil
		}

		lastErr = err

		if attempt < config.MaxAttempts {
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * config.Multiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return fmt.Errorf("failed to connect after %d attempts: %w", config.MaxAttempts, lastErr)
}

// ExecWithRetry executes a command with retry on failure
func (c *Client) ExecWithRetry(cmd string, config RetryConfig) (string, error) {
	var lastErr error
	var output string

	delay := config.InitialDelay
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		var err error
		output, err = c.Exec(cmd)
		if err == nil {
			return output, nil
		}

		lastErr = err

		// If connection is broken, try to reconnect
		if !c.IsConnected() {
			_ = c.Close()
			if err := c.Connect(); err != nil {
				lastErr = fmt.Errorf("reconnection failed: %w", err)
			}
		}

		if attempt < config.MaxAttempts {
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * config.Multiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return output, fmt.Errorf("command failed after %d attempts: %w", config.MaxAttempts, lastErr)
}
