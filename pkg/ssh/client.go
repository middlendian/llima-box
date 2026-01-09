package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/lima-vm/lima/pkg/store"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// Client wraps SSH connection to a Lima VM instance
type Client struct {
	instanceName string
	instance     *store.Instance
	sshConfig    *ssh.ClientConfig
	client       *ssh.Client
}

// NewClient creates a new SSH client for the given Lima instance
func NewClient(instanceName string) (*Client, error) {
	if instanceName == "" {
		return nil, fmt.Errorf("instance name cannot be empty")
	}

	// Get instance details
	inst, err := store.Inspect(instanceName)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect instance %s: %w", instanceName, err)
	}

	// Check if instance is running
	if inst.Status != store.StatusRunning {
		return nil, fmt.Errorf("instance %s is not running (status: %s)", instanceName, inst.Status)
	}

	return &Client{
		instanceName: instanceName,
		instance:     inst,
	}, nil
}

// Connect establishes SSH connection to the Lima VM
func (c *Client) Connect() error {
	if c.client != nil {
		return nil // Already connected
	}

	// Get SSH user from instance config
	user := "lima" // Default user
	if c.instance.Config != nil && c.instance.Config.User.Name != nil {
		user = *c.instance.Config.User.Name
	}

	// Get SSH key paths - Lima stores keys in $LIMA_HOME/_config/
	limaDir := store.Directory()
	keyPaths := []string{
		filepath.Join(limaDir, "_config", "user"),
		filepath.Join(c.instance.Dir, "ssh_key"),
	}

	// Load SSH keys
	authMethods := []ssh.AuthMethod{}
	for _, keyPath := range keyPaths {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			continue // Skip invalid keys
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			continue // Skip invalid keys
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("no valid SSH keys found in %v", keyPaths)
	}

	// Create SSH client config
	c.sshConfig = &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Lima VMs are trusted
		Timeout:         10 * time.Second,
	}

	// Connect to SSH using instance hostname and port
	host := c.instance.Hostname
	if host == "" {
		host = "127.0.0.1"
	}
	port := c.instance.SSHLocalPort
	if port == 0 {
		port = 22
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	client, err := ssh.Dial("tcp", addr, c.sshConfig)
	if err != nil {
		return fmt.Errorf("failed to dial SSH at %s: %w", addr, err)
	}

	c.client = client
	return nil
}

// Exec executes a command on the VM and returns the output
// This is for non-interactive commands
func (c *Client) Exec(cmd string) (string, error) {
	if c.client == nil {
		if err := c.Connect(); err != nil {
			return "", err
		}
	}

	// Create a session
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Run command and capture output
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

// ExecContext executes a command with context support
func (c *Client) ExecContext(ctx context.Context, cmd string) (string, error) {
	if c.client == nil {
		if err := c.Connect(); err != nil {
			return "", err
		}
	}

	// Create a session
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Create channel for command completion
	done := make(chan error, 1)
	var output []byte

	go func() {
		output, err = session.CombinedOutput(cmd)
		done <- err
	}()

	// Wait for command or context cancellation
	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGKILL)
		return "", ctx.Err()
	case err := <-done:
		if err != nil {
			return string(output), fmt.Errorf("command failed: %w", err)
		}
		return string(output), nil
	}
}

// ExecInteractive executes a command interactively with terminal support
// This is for commands that need user interaction (like shells)
func (c *Client) ExecInteractive(cmd string) error {
	if c.client == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// Create a session
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Setup SSH agent forwarding if available
	if err := setupAgentForwarding(session); err != nil {
		// SSH agent forwarding is optional, continue without it
		fmt.Fprintf(os.Stderr, "Warning: SSH agent forwarding not available: %v\n", err)
	}

	// Connect stdin, stdout, stderr
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// Get terminal size
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		// Request pseudo terminal
		state, err := term.MakeRaw(fd)
		if err != nil {
			return fmt.Errorf("failed to make terminal raw: %w", err)
		}
		defer term.Restore(fd, state)

		width, height, err := term.GetSize(fd)
		if err != nil {
			width, height = 80, 24 // Default size
		}

		// Request PTY
		if err := session.RequestPty("xterm-256color", height, width, ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}); err != nil {
			return fmt.Errorf("failed to request PTY: %w", err)
		}

		// Handle terminal resize
		go handleTerminalResize(session, fd)
	}

	// Run command
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// ExecPipe executes a command and returns pipes for stdin, stdout, stderr
// This is useful for programmatic interaction with commands
func (c *Client) ExecPipe(cmd string) (stdin io.WriteCloser, stdout, stderr io.Reader, err error) {
	if c.client == nil {
		if err := c.Connect(); err != nil {
			return nil, nil, nil, err
		}
	}

	// Create a session
	session, err := c.client.NewSession()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Get pipes
	stdin, err = session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, nil, nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err = session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, nil, nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err = session.StderrPipe()
	if err != nil {
		session.Close()
		return nil, nil, nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	if err := session.Start(cmd); err != nil {
		session.Close()
		return nil, nil, nil, fmt.Errorf("failed to start command: %w", err)
	}

	return stdin, stdout, stderr, nil
}

// Close closes the SSH connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// IsConnected returns true if the client has an active connection
func (c *Client) IsConnected() bool {
	return c.client != nil
}

// GetUser returns the SSH user for the connection
func (c *Client) GetUser() string {
	if c.sshConfig != nil {
		return c.sshConfig.User
	}
	return ""
}

// GetInstanceName returns the Lima instance name
func (c *Client) GetInstanceName() string {
	return c.instanceName
}

// setupAgentForwarding sets up SSH agent forwarding for the session
func setupAgentForwarding(session *ssh.Session) error {
	// Check if SSH_AUTH_SOCK is set
	authSock := os.Getenv("SSH_AUTH_SOCK")
	if authSock == "" {
		return fmt.Errorf("SSH_AUTH_SOCK not set")
	}

	// Verify the socket exists
	if _, err := os.Stat(authSock); err != nil {
		return fmt.Errorf("SSH agent socket not found: %w", err)
	}

	// Request agent forwarding
	if err := session.RequestAgentForwarding(); err != nil {
		return fmt.Errorf("failed to request agent forwarding: %w", err)
	}

	return nil
}

// handleTerminalResize monitors terminal size changes and updates the remote PTY
func handleTerminalResize(session *ssh.Session, fd int) {
	// This is a simplified version - a full implementation would use SIGWINCH
	// For now, we'll just set the initial size
	// A production version would listen for terminal resize signals
}

// GetSSHConfigPath returns the path to Lima's SSH config for the instance
func (c *Client) GetSSHConfigPath() string {
	return filepath.Join(c.instance.Dir, "ssh.config")
}
