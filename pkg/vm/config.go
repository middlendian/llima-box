package vm

import _ "embed"

//go:embed lima.yaml
var embeddedConfig string

// GetEmbeddedConfig returns the embedded Lima configuration YAML
func GetEmbeddedConfig() (string, error) {
	return embeddedConfig, nil
}
