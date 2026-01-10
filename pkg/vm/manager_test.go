package vm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mockExecutor implements commandExecutor for testing
type mockExecutor struct {
	// responses maps command args to response data
	responses map[string][]byte
	// errors maps command args to errors
	errors map[string]error
	// calls tracks all executed commands
	calls [][]string
}

func newMockExecutor() *mockExecutor {
	return &mockExecutor{
		responses: make(map[string][]byte),
		errors:    make(map[string]error),
		calls:     make([][]string, 0),
	}
}

func (m *mockExecutor) exec(_ context.Context, _ string, args ...string) ([]byte, error) {
	m.calls = append(m.calls, args)
	key := strings.Join(args, " ")

	if err, ok := m.errors[key]; ok {
		return nil, err
	}

	if resp, ok := m.responses[key]; ok {
		return resp, nil
	}

	return nil, fmt.Errorf("unexpected command: %s", key)
}

func (m *mockExecutor) setResponse(args []string, data []byte) {
	key := strings.Join(args, " ")
	m.responses[key] = data
}

func (m *mockExecutor) setError(args []string, err error) {
	key := strings.Join(args, " ")
	m.errors[key] = err
}

func (m *mockExecutor) assertCalled(t *testing.T, expectedArgs []string) {
	t.Helper()
	key := strings.Join(expectedArgs, " ")
	for _, call := range m.calls {
		if strings.Join(call, " ") == key {
			return
		}
	}
	t.Errorf("expected command not called: %s", key)
}

// loadTestData loads test data from the testdata directory
func loadTestData(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join("testdata", filename)
	// #nosec G304 -- Test helper reading from known testdata directory
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to load test data %s: %v", filename, err)
	}
	return data
}

// TestListInstances_SingleObject tests parsing a single instance object (the bug scenario)
func TestListInstances_SingleObject(t *testing.T) {
	mock := newMockExecutor()
	mock.setResponse([]string{"--tty=false", "list", "--json"}, loadTestData(t, "list_single_instance.json"))

	mgr := newManagerWithExecutor("llima-box", mock)

	instances, err := mgr.listInstances()
	if err != nil {
		t.Fatalf("listInstances failed: %v", err)
	}

	if len(instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(instances))
	}

	if instances[0].Name != "llima-box" {
		t.Errorf("expected name 'llima-box', got '%s'", instances[0].Name)
	}

	if instances[0].Status != "Running" {
		t.Errorf("expected status 'Running', got '%s'", instances[0].Status)
	}
}

// TestListInstances_Array tests parsing an array of instances
func TestListInstances_Array(t *testing.T) {
	mock := newMockExecutor()
	mock.setResponse([]string{"--tty=false", "list", "--json"}, loadTestData(t, "list_multiple_instances.json"))

	mgr := newManagerWithExecutor("llima-box", mock)

	instances, err := mgr.listInstances()
	if err != nil {
		t.Fatalf("listInstances failed: %v", err)
	}

	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}

	if instances[0].Name != "llima-box" {
		t.Errorf("expected first instance name 'llima-box', got '%s'", instances[0].Name)
	}

	if instances[1].Name != "default" {
		t.Errorf("expected second instance name 'default', got '%s'", instances[1].Name)
	}
}

// TestListInstances_Empty tests parsing an empty array
func TestListInstances_Empty(t *testing.T) {
	mock := newMockExecutor()
	mock.setResponse([]string{"--tty=false", "list", "--json"}, loadTestData(t, "list_empty.json"))

	mgr := newManagerWithExecutor("llima-box", mock)

	instances, err := mgr.listInstances()
	if err != nil {
		t.Fatalf("listInstances failed: %v", err)
	}

	if len(instances) != 0 {
		t.Fatalf("expected 0 instances, got %d", len(instances))
	}
}

// TestListInstances_InvalidJSON tests handling of invalid JSON
func TestListInstances_InvalidJSON(t *testing.T) {
	mock := newMockExecutor()
	mock.setResponse([]string{"--tty=false", "list", "--json"}, []byte("invalid json"))

	mgr := newManagerWithExecutor("llima-box", mock)

	_, err := mgr.listInstances()
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "failed to parse limactl list output") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

// TestExists_InstanceExists tests checking if an instance exists
func TestExists_InstanceExists(t *testing.T) {
	tests := []struct {
		name         string
		instanceName string
		dataFile     string
		want         bool
	}{
		{
			name:         "single instance exists",
			instanceName: "llima-box",
			dataFile:     "list_single_instance.json",
			want:         true,
		},
		{
			name:         "instance in array exists",
			instanceName: "llima-box",
			dataFile:     "list_multiple_instances.json",
			want:         true,
		},
		{
			name:         "instance does not exist",
			instanceName: "nonexistent",
			dataFile:     "list_single_instance.json",
			want:         false,
		},
		{
			name:         "no instances",
			instanceName: "llima-box",
			dataFile:     "list_empty.json",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.setResponse([]string{"--tty=false", "list", "--json"}, loadTestData(t, tt.dataFile))

			mgr := newManagerWithExecutor(tt.instanceName, mock)

			exists, err := mgr.Exists()
			if err != nil {
				t.Fatalf("Exists failed: %v", err)
			}

			if exists != tt.want {
				t.Errorf("expected exists=%v, got %v", tt.want, exists)
			}
		})
	}
}

// TestIsRunning tests checking if an instance is running
func TestIsRunning(t *testing.T) {
	tests := []struct {
		name     string
		dataFile string
		want     bool
		wantErr  bool
	}{
		{
			name:     "running instance",
			dataFile: "list_running_instance.json",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "stopped instance",
			dataFile: "list_stopped_instance.json",
			want:     false,
			wantErr:  false,
		},
		{
			name:     "instance not found",
			dataFile: "list_empty.json",
			want:     false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.setResponse([]string{"--tty=false", "list", "--json"}, loadTestData(t, tt.dataFile))

			mgr := newManagerWithExecutor("llima-box", mock)

			running, err := mgr.IsRunning()
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got error=%v", tt.wantErr, err)
			}

			if !tt.wantErr && running != tt.want {
				t.Errorf("expected running=%v, got %v", tt.want, running)
			}
		})
	}
}

// TestGetInstance tests retrieving instance details
func TestGetInstance(t *testing.T) {
	tests := []struct {
		name         string
		instanceName string
		dataFile     string
		wantName     string
		wantStatus   string
		wantErr      bool
	}{
		{
			name:         "get single instance",
			instanceName: "llima-box",
			dataFile:     "list_single_instance.json",
			wantName:     "llima-box",
			wantStatus:   "Running",
			wantErr:      false,
		},
		{
			name:         "get instance from array",
			instanceName: "default",
			dataFile:     "list_multiple_instances.json",
			wantName:     "default",
			wantStatus:   "Stopped",
			wantErr:      false,
		},
		{
			name:         "instance not found",
			instanceName: "nonexistent",
			dataFile:     "list_single_instance.json",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.setResponse([]string{"--tty=false", "list", "--json"}, loadTestData(t, tt.dataFile))

			mgr := newManagerWithExecutor(tt.instanceName, mock)

			inst, err := mgr.GetInstance()
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got error=%v", tt.wantErr, err)
			}

			if tt.wantErr {
				return
			}

			if inst.Name != tt.wantName {
				t.Errorf("expected name=%s, got %s", tt.wantName, inst.Name)
			}

			if inst.Status != tt.wantStatus {
				t.Errorf("expected status=%s, got %s", tt.wantStatus, inst.Status)
			}
		})
	}
}

// TestStart tests starting an instance
func TestStart(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  string
		dataFile       string
		expectStartCmd bool
		wantErr        bool
	}{
		{
			name:           "start stopped instance",
			initialStatus:  "Stopped",
			dataFile:       "list_stopped_instance.json",
			expectStartCmd: true,
			wantErr:        false,
		},
		{
			name:           "start already running instance",
			initialStatus:  "Running",
			dataFile:       "list_running_instance.json",
			expectStartCmd: false,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.setResponse([]string{"--tty=false", "list", "--json"}, loadTestData(t, tt.dataFile))
			mock.setResponse([]string{"--tty=false", "start", "llima-box"}, []byte{})

			mgr := newManagerWithExecutor("llima-box", mock)

			err := mgr.Start(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got error=%v", tt.wantErr, err)
			}

			if tt.expectStartCmd {
				mock.assertCalled(t, []string{"--tty=false", "start", "llima-box"})
			}
		})
	}
}

// TestStop tests stopping an instance
func TestStop(t *testing.T) {
	mock := newMockExecutor()
	mock.setResponse([]string{"--tty=false", "list", "--json"}, loadTestData(t, "list_running_instance.json"))
	mock.setResponse([]string{"--tty=false", "stop", "llima-box"}, []byte{})

	mgr := newManagerWithExecutor("llima-box", mock)

	err := mgr.Stop(context.Background())
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	mock.assertCalled(t, []string{"--tty=false", "stop", "llima-box"})
}

// TestDelete tests deleting an instance
func TestDelete(t *testing.T) {
	tests := []struct {
		name        string
		force       bool
		expectedCmd []string
	}{
		{
			name:        "delete without force",
			force:       false,
			expectedCmd: []string{"--tty=false", "delete", "llima-box"},
		},
		{
			name:        "delete with force",
			force:       true,
			expectedCmd: []string{"--tty=false", "delete", "llima-box", "--force"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.setResponse([]string{"--tty=false", "list", "--json"}, loadTestData(t, "list_running_instance.json"))
			mock.setResponse(tt.expectedCmd, []byte{})

			mgr := newManagerWithExecutor("llima-box", mock)

			err := mgr.Delete(context.Background(), tt.force)
			if err != nil {
				t.Fatalf("Delete failed: %v", err)
			}

			mock.assertCalled(t, tt.expectedCmd)
		})
	}
}

// TestGetInstanceName tests getting the instance name
func TestGetInstanceName(t *testing.T) {
	tests := []struct {
		name         string
		instanceName string
		want         string
	}{
		{
			name:         "custom instance name",
			instanceName: "test-instance",
			want:         "test-instance",
		},
		{
			name:         "default instance name",
			instanceName: "",
			want:         DefaultInstanceName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := newManagerWithExecutor(tt.instanceName, newMockExecutor())

			if got := mgr.GetInstanceName(); got != tt.want {
				t.Errorf("expected instance name=%s, got %s", tt.want, got)
			}
		})
	}
}

// TestCommandExecutionErrors tests handling of command execution errors
func TestCommandExecutionErrors(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*mockExecutor)
		test   func(*Manager) error
		errMsg string
	}{
		{
			name: "list command error",
			setup: func(m *mockExecutor) {
				m.setError([]string{"--tty=false", "list", "--json"}, fmt.Errorf("command failed"))
			},
			test: func(mgr *Manager) error {
				_, err := mgr.listInstances()
				return err
			},
			errMsg: "command failed",
		},
		{
			name: "start command error",
			setup: func(m *mockExecutor) {
				m.setResponse([]string{"--tty=false", "list", "--json"}, loadTestData(t, "list_stopped_instance.json"))
				m.setError([]string{"--tty=false", "start", "llima-box"}, fmt.Errorf("start failed"))
			},
			test: func(mgr *Manager) error {
				return mgr.Start(context.Background())
			},
			errMsg: "failed to start instance",
		},
		{
			name: "stop command error",
			setup: func(m *mockExecutor) {
				m.setResponse([]string{"--tty=false", "list", "--json"}, loadTestData(t, "list_running_instance.json"))
				m.setError([]string{"--tty=false", "stop", "llima-box"}, fmt.Errorf("stop failed"))
			},
			test: func(mgr *Manager) error {
				return mgr.Stop(context.Background())
			},
			errMsg: "failed to stop instance",
		},
		{
			name: "delete command error",
			setup: func(m *mockExecutor) {
				m.setResponse([]string{"--tty=false", "list", "--json"}, loadTestData(t, "list_running_instance.json"))
				m.setError([]string{"--tty=false", "delete", "llima-box"}, fmt.Errorf("delete failed"))
			},
			test: func(mgr *Manager) error {
				return mgr.Delete(context.Background(), false)
			},
			errMsg: "failed to delete instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			tt.setup(mock)

			mgr := newManagerWithExecutor("llima-box", mock)

			err := tt.test(mgr)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing '%s', got: %v", tt.errMsg, err)
			}
		})
	}
}

// TestGetConfigPath tests getting the configuration path
func TestGetConfigPath(t *testing.T) {
	mock := newMockExecutor()
	mock.setResponse([]string{"--tty=false", "list", "--json"}, loadTestData(t, "list_running_instance.json"))

	mgr := newManagerWithExecutor("llima-box", mock)

	configPath, err := mgr.GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath failed: %v", err)
	}

	expectedSuffix := "/llima-box/lima.yaml"
	if !strings.HasSuffix(configPath, expectedSuffix) {
		t.Errorf("expected path to end with '%s', got: %s", expectedSuffix, configPath)
	}
}
