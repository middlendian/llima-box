package env

import (
	"testing"
)

func TestIsValidEnvironmentName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid environment name",
			input: "my-project-a1b2",
			want:  true,
		},
		{
			name:  "valid with longer base name",
			input: "my-cool-app-c3d4",
			want:  true,
		},
		{
			name:  "invalid - no hash",
			input: "my-project",
			want:  false,
		},
		{
			name:  "invalid - hash too short",
			input: "my-project-a1b",
			want:  false,
		},
		{
			name:  "invalid - hash too long",
			input: "my-project-a1b2c",
			want:  false,
		},
		{
			name:  "invalid - starts with number",
			input: "123-project-a1b2",
			want:  false,
		},
		{
			name:  "invalid - uppercase",
			input: "My-Project-a1b2",
			want:  false,
		},
		{
			name:  "invalid - non-hex hash",
			input: "my-project-wxyz",
			want:  false,
		},
		{
			name:  "valid - single letter base",
			input: "a-a1b2",
			want:  true,
		},
		{
			name:  "valid - with numbers in base",
			input: "app123-a1b2",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidEnvironmentName(tt.input)
			if got != tt.want {
				t.Errorf("IsValidEnvironmentName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestEnvironmentStruct(t *testing.T) {
	env := &Environment{
		Name:        "test-env-a1b2",
		ProjectPath: "/Users/test/project",
	}

	if env.Name != "test-env-a1b2" {
		t.Errorf("Expected Name to be 'test-env-a1b2', got %s", env.Name)
	}
	if env.ProjectPath != "/Users/test/project" {
		t.Errorf("Expected ProjectPath to be '/Users/test/project', got %s", env.ProjectPath)
	}
}
