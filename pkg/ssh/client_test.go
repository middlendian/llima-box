package ssh

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestNewClient tests client creation
func TestNewClient(t *testing.T) {
	tests := []struct {
		name         string
		instanceName string
		wantErr      bool
	}{
		{
			name:         "empty instance name",
			instanceName: "",
			wantErr:      true,
		},
		{
			name:         "non-existent instance",
			instanceName: "non-existent-instance",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.instanceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestClientConnection tests SSH connection (requires running VM)
// This is a manual test - run with: go test -v -run TestClientConnection
func TestClientConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if we should run integration tests
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test (set RUN_INTEGRATION_TESTS=true to run)")
	}

	instanceName := os.Getenv("LIMA_INSTANCE_NAME")
	if instanceName == "" {
		instanceName = "llima-box"
	}

	client, err := NewClient(instanceName)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	if !client.IsConnected() {
		t.Error("IsConnected() returned false after successful connection")
	}
}

// TestExec tests command execution (requires running VM)
func TestExec(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test (set RUN_INTEGRATION_TESTS=true to run)")
	}

	instanceName := os.Getenv("LIMA_INSTANCE_NAME")
	if instanceName == "" {
		instanceName = "llima-box"
	}

	client, err := NewClient(instanceName)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	tests := []struct {
		name    string
		cmd     string
		wantErr bool
	}{
		{
			name:    "simple command",
			cmd:     "echo hello",
			wantErr: false,
		},
		{
			name:    "whoami command",
			cmd:     "whoami",
			wantErr: false,
		},
		{
			name:    "pwd command",
			cmd:     "pwd",
			wantErr: false,
		},
		{
			name:    "failing command",
			cmd:     "exit 1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := client.Exec(tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exec(%q) error = %v, wantErr %v", tt.cmd, err, tt.wantErr)
			}
			if !tt.wantErr {
				t.Logf("Output: %s", output)
			}
		})
	}
}

// TestExecContext tests command execution with context
func TestExecContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test (set RUN_INTEGRATION_TESTS=true to run)")
	}

	instanceName := os.Getenv("LIMA_INSTANCE_NAME")
	if instanceName == "" {
		instanceName = "llima-box"
	}

	client, err := NewClient(instanceName)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	t.Run("successful command", func(t *testing.T) {
		ctx := context.Background()
		output, err := client.ExecContext(ctx, "echo hello")
		if err != nil {
			t.Errorf("ExecContext() error = %v", err)
		}
		t.Logf("Output: %s", output)
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := client.ExecContext(ctx, "sleep 5")
		if err == nil {
			t.Error("ExecContext() expected timeout error, got nil")
		}
	})
}

// TestConnectWithRetry tests retry logic
func TestConnectWithRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test (set RUN_INTEGRATION_TESTS=true to run)")
	}

	instanceName := os.Getenv("LIMA_INSTANCE_NAME")
	if instanceName == "" {
		instanceName = "llima-box"
	}

	client, err := NewClient(instanceName)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
	}

	if err := client.ConnectWithRetry(config); err != nil {
		t.Fatalf("ConnectWithRetry() failed: %v", err)
	}

	if !client.IsConnected() {
		t.Error("IsConnected() returned false after successful connection")
	}
}

// TestDefaultRetryConfig tests default retry configuration
func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("MaxAttempts = %d, want 3", config.MaxAttempts)
	}
	if config.InitialDelay != 1*time.Second {
		t.Errorf("InitialDelay = %v, want 1s", config.InitialDelay)
	}
	if config.MaxDelay != 10*time.Second {
		t.Errorf("MaxDelay = %v, want 10s", config.MaxDelay)
	}
	if config.Multiplier != 2.0 {
		t.Errorf("Multiplier = %f, want 2.0", config.Multiplier)
	}
}

// TestGetUser tests user retrieval
func TestGetUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test (set RUN_INTEGRATION_TESTS=true to run)")
	}

	instanceName := os.Getenv("LIMA_INSTANCE_NAME")
	if instanceName == "" {
		instanceName = "llima-box"
	}

	client, err := NewClient(instanceName)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	user := client.GetUser()
	if user == "" {
		t.Error("GetUser() returned empty string")
	}
	t.Logf("SSH user: %s", user)
}
