package templates

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

//go:embed scripts/*
var scriptsInjectFS embed.FS

// TemplateExecutor executes a template with parameters
type TemplateExecutor struct {
	store      *TemplateStore
	scriptsFS  embed.FS
	templateID string
	parameters map[string]any
}

// NewTemplateExecutor creates a new executor for the given template
func NewTemplateExecutor(store *TemplateStore, templateID string, parameters string) (*TemplateExecutor, error) {
	if !store.Exists(templateID) {
		return nil, fmt.Errorf("template %s does not exist", templateID)
	}

	var params map[string]any
	if err := json.Unmarshal([]byte(parameters), &params); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	zap.L().Info(fmt.Sprintf("Creating template executor for template: %s", templateID))

	return &TemplateExecutor{
		store:      store,
		scriptsFS:  scriptsInjectFS,
		templateID: templateID,
		parameters: params,
	}, nil
}

// Execute renders the template with parameters and returns base64 encoded user data
func (e *TemplateExecutor) Execute() (string, error) {
	content, err := e.store.GetContent(e.templateID)
	if err != nil {
		zap.L().Error(fmt.Sprintf("Template not found: %s, err=%v", e.templateID, err))
		return "", fmt.Errorf("template not found: %s", e.templateID)
	}

	tmpl, err := template.New(e.templateID).Parse(content)
	if err != nil {
		zap.L().Error(fmt.Sprintf("Failed to parse template %s err=%v", e.templateID, err))
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, e.parameters); err != nil {
		zap.L().Error(fmt.Sprintf("Failed to execute template %s err=%v", e.templateID, err))
		return "", err
	}

	userData, err := e.injectScripts(buf.Bytes())
	if err != nil {
		zap.L().Error(fmt.Sprintf("Failed to inject scripts into user data err=%v", err))
		return "", err
	}

	return base64.StdEncoding.EncodeToString(userData), nil
}

// WriteFile represents a file entry in cloud-config
type WriteFile struct {
	Path        string `yaml:"path"`
	Content     string `yaml:"content"`
	Encoding    string `yaml:"encoding,omitempty"`
	Permissions string `yaml:"permissions,omitempty"`
	Defer       bool   `yaml:"defer,omitempty"`
}

// CloudConfig represents a cloud-init configuration
type CloudConfig struct {
	WriteFiles []WriteFile    `yaml:"write_files,omitempty"`
	Other      map[string]any `yaml:",inline"`
}

func (e *TemplateExecutor) injectScripts(userData []byte) ([]byte, error) {
	// Read script files from the embedded filesystem
	entries, err := e.scriptsFS.ReadDir("scripts")
	if err != nil {
		return nil, fmt.Errorf("failed to read scripts directory: %w", err)
	}

	// Filter for .sh files
	var scriptFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sh") {
			scriptFiles = append(scriptFiles, entry.Name())
		}
	}

	if len(scriptFiles) == 0 {
		zap.L().Info("No scripts found to inject")
		return userData, nil
	}

	var cloudConfig CloudConfig
	if err := yaml.Unmarshal(userData, &cloudConfig); err != nil {
		return nil, fmt.Errorf("failed to parse user data as YAML: %w", err)
	}

	templateVarRegex := regexp.MustCompile(`\{\{\s*\.(\w+)\s*\}\}`)

	for _, scriptName := range scriptFiles {
		scriptContent, err := e.scriptsFS.ReadFile(filepath.Join("scripts", scriptName))
		if err != nil {
			return nil, fmt.Errorf("failed to read script %s: %w", scriptName, err)
		}

		renderedScript := templateVarRegex.ReplaceAllStringFunc(string(scriptContent), func(match string) string {
			submatches := templateVarRegex.FindStringSubmatch(match)
			if len(submatches) < 2 {
				return match
			}
			varName := submatches[1]

			envName := toEnvVarName(varName)
			if envValue := os.Getenv(envName); envValue != "" {
				return envValue
			}
			return match
		})

		encodedContent := base64.StdEncoding.EncodeToString([]byte(renderedScript))
		destPath := fmt.Sprintf("/var/lib/cloud/scripts/per-once/zzzz-%s", scriptName)

		cloudConfig.WriteFiles = append(cloudConfig.WriteFiles, WriteFile{
			Path:        destPath,
			Content:     encodedContent,
			Encoding:    "b64",
			Permissions: "0755",
			Defer:       true,
		})
	}

	result, err := yaml.Marshal(&cloudConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cloud config: %w", err)
	}

	if !bytes.HasPrefix(userData, []byte("#cloud-config")) {
		return result, nil
	}

	return append([]byte("#cloud-config\n"), result...), nil
}

func toEnvVarName(camelCase string) string {
	var result strings.Builder
	for i, r := range camelCase {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToUpper(result.String())
}
