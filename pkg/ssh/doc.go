// Package ssh provides SSH client functionality for connecting to Lima VM instances.
//
// This package wraps the standard golang.org/x/crypto/ssh package and integrates
// with Lima's SSH configuration to provide seamless connection to Lima VMs.
//
// # Basic Usage
//
// Create a client and execute a command:
//
//	client, err := ssh.NewClient("llima-box")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	output, err := client.Exec("whoami")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(output)
//
// # Interactive Shell
//
// Launch an interactive shell:
//
//	client, err := ssh.NewClient("llima-box")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	err = client.ExecInteractive("bash")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # SSH Agent Forwarding
//
// SSH agent forwarding is automatically enabled when:
// - SSH_AUTH_SOCK environment variable is set
// - The socket exists and is accessible
// - Running in an interactive session with a PTY
//
// This allows Git operations and other SSH-based tools to work
// seamlessly inside the VM using your host's SSH keys.
//
// # Connection Management
//
// The client automatically connects on first command execution.
// You can also explicitly connect:
//
//	client, err := ssh.NewClient("llima-box")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	if err := client.Connect(); err != nil {
//		log.Fatal(err)
//	}
//
// # Error Handling
//
// All methods return descriptive errors. Connection failures,
// command execution errors, and SSH configuration issues are
// properly wrapped with context.
package ssh
