package templates

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestNewTemplateExecutor(t *testing.T) {
	t.Run("creates executor for existing template", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		_ = store.Create(GenericTemplate{
			Id:          "executor_test",
			Name:        "Executor Test",
			Description: "Test executor creation",
		})
		params := `{"ServerName": "test-server"}`

		// Act
		executor, err := NewTemplateExecutor(store, "executor_test", params)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if executor == nil {
			t.Fatal("expected executor to be non-nil")
		}
		if executor.templateID != "executor_test" {
			t.Errorf("expected templateID 'executor_test', got '%s'", executor.templateID)
		}
		if executor.store != store {
			t.Error("expected executor to reference the same store")
		}
	})

	t.Run("returns error for nonexistent template", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		params := `{"key": "value"}`

		// Act
		executor, err := NewTemplateExecutor(store, "nonexistent", params)

		// Assert
		if err == nil {
			t.Fatal("expected error for nonexistent template")
		}
		if executor != nil {
			t.Error("expected executor to be nil on error")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("returns error for invalid JSON parameters", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		_ = store.Create(GenericTemplate{
			Id:          "json_test",
			Name:        "JSON Test",
			Description: "Test invalid JSON",
		})
		invalidParams := `{invalid json}`

		// Act
		executor, err := NewTemplateExecutor(store, "json_test", invalidParams)

		// Assert
		if err == nil {
			t.Fatal("expected error for invalid JSON parameters")
		}
		if executor != nil {
			t.Error("expected executor to be nil on error")
		}
		if !strings.Contains(err.Error(), "failed to parse parameters") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("parses empty JSON object successfully", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		_ = store.Create(GenericTemplate{
			Id:          "empty_params",
			Name:        "Empty Params",
			Description: "Test empty params",
		})
		emptyParams := `{}`

		// Act
		executor, err := NewTemplateExecutor(store, "empty_params", emptyParams)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if executor == nil {
			t.Fatal("expected executor to be non-nil")
		}
		if len(executor.parameters) != 0 {
			t.Errorf("expected empty parameters, got %d", len(executor.parameters))
		}
	})

	t.Run("parses complex parameters successfully", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		_ = store.Create(GenericTemplate{
			Id:          "complex_params",
			Name:        "Complex Params",
			Description: "Test complex params",
		})
		complexParams := `{"name": "server1", "port": 8080, "enabled": true, "tags": ["web", "prod"]}`

		// Act
		executor, err := NewTemplateExecutor(store, "complex_params", complexParams)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if executor.parameters["name"] != "server1" {
			t.Errorf("expected name 'server1', got '%v'", executor.parameters["name"])
		}
		if executor.parameters["port"] != float64(8080) {
			t.Errorf("expected port 8080, got '%v'", executor.parameters["port"])
		}
		if executor.parameters["enabled"] != true {
			t.Errorf("expected enabled true, got '%v'", executor.parameters["enabled"])
		}
	})
}

func TestTemplateExecutor_Execute(t *testing.T) {
	t.Run("executes template and returns base64 encoded result", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		_ = store.Create(GenericTemplate{
			Id:             "execute_test",
			Name:           "Execute Test",
			Description:    "Test execution",
			UpdatePackages: true,
			Packages:       []string{"nginx"},
		})
		executor, _ := NewTemplateExecutor(store, "execute_test", `{}`)

		// Act
		result, err := executor.Execute()

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result == "" {
			t.Error("expected non-empty result")
		}

		// Verify it's valid base64
		decoded, err := base64.StdEncoding.DecodeString(result)
		if err != nil {
			t.Fatalf("expected valid base64, got error: %v", err)
		}
		if !strings.Contains(string(decoded), "#cloud-config") {
			t.Error("expected decoded content to contain '#cloud-config'")
		}
	})

	t.Run("substitutes template parameters", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		_ = store.Create(GenericTemplate{
			Id:          "param_sub_test",
			Name:        "Param Sub Test",
			Description: "Test parameter substitution",
			Parameters: []Parameter{
				{Name: "ServerName", Description: "Server name"},
			},
			Commands: []string{"echo {{ .ServerName }}"},
		})
		executor, _ := NewTemplateExecutor(store, "param_sub_test", `{"ServerName": "my-server"}`)

		// Act
		result, err := executor.Execute()

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		decoded, _ := base64.StdEncoding.DecodeString(result)
		if !strings.Contains(string(decoded), "my-server") {
			t.Errorf("expected 'my-server' in output, got: %s", string(decoded))
		}
	})

	t.Run("handles template with files", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		_ = store.Create(GenericTemplate{
			Id:          "files_test",
			Name:        "Files Test",
			Description: "Test with files",
			Files: []File{
				{Path: "/etc/test.conf", Content: "test content"},
			},
		})
		executor, _ := NewTemplateExecutor(store, "files_test", `{}`)

		// Act
		result, err := executor.Execute()

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		decoded, _ := base64.StdEncoding.DecodeString(result)
		if !strings.Contains(string(decoded), "/etc/test.conf") {
			t.Errorf("expected '/etc/test.conf' in output, got: %s", string(decoded))
		}
	})

	t.Run("handles template with packages", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		_ = store.Create(GenericTemplate{
			Id:          "packages_test",
			Name:        "Packages Test",
			Description: "Test with packages",
			Packages:    []string{"vim", "curl", "wget"},
		})
		executor, _ := NewTemplateExecutor(store, "packages_test", `{}`)

		// Act
		result, err := executor.Execute()

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		decoded, _ := base64.StdEncoding.DecodeString(result)
		decodedStr := string(decoded)
		if !strings.Contains(decodedStr, "vim") {
			t.Error("expected 'vim' in packages")
		}
		if !strings.Contains(decodedStr, "curl") {
			t.Error("expected 'curl' in packages")
		}
		if !strings.Contains(decodedStr, "wget") {
			t.Error("expected 'wget' in packages")
		}
	})

	t.Run("handles template with commands", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		_ = store.Create(GenericTemplate{
			Id:          "commands_test",
			Name:        "Commands Test",
			Description: "Test with commands",
			Commands:    []string{"systemctl start nginx", "echo done"},
		})
		executor, _ := NewTemplateExecutor(store, "commands_test", `{}`)

		// Act
		result, err := executor.Execute()

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		decoded, _ := base64.StdEncoding.DecodeString(result)
		if !strings.Contains(string(decoded), "systemctl start nginx") {
			t.Error("expected 'systemctl start nginx' in commands")
		}
	})
}

func TestToEnvVarName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple lowercase", "server", "SERVER"},
		{"camelCase", "serverName", "SERVER_NAME"},
		{"PascalCase", "ServerName", "SERVER_NAME"},
		{"multiple capitals", "serverNamePort", "SERVER_NAME_PORT"},
		{"all caps inserts underscores", "SERVER", "S_E_R_V_E_R"},
		{"single char", "s", "S"},
		{"empty string", "", ""},
		{"consecutive capitals", "HTTPServer", "H_T_T_P_SERVER"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			result := toEnvVarName(tt.input)

			// Assert
			if result != tt.expected {
				t.Errorf("toEnvVarName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRetrieveExecutor(t *testing.T) {
	t.Run("retrieves executor using global store", func(t *testing.T) {
		// Arrange
		store := MustTemplateStore()
		_ = store.Create(GenericTemplate{
			Id:          "global_executor_test",
			Name:        "Global Executor Test",
			Description: "Test global executor",
		})
		defer store.Delete("global_executor_test")

		// Act
		executor, err := RetrieveExecutor("global_executor_test", `{}`)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if executor == nil {
			t.Fatal("expected executor to be non-nil")
		}
	})

	t.Run("returns error for nonexistent template in global store", func(t *testing.T) {
		// Arrange - nothing to arrange, template doesn't exist

		// Act
		executor, err := RetrieveExecutor("definitely_nonexistent_template", `{}`)

		// Assert
		if err == nil {
			t.Fatal("expected error for nonexistent template")
		}
		if executor != nil {
			t.Error("expected executor to be nil")
		}
	})
}
