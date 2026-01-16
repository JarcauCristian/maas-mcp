package templates

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type TemplateExecutor struct {
	TemplateId string
	Parameters map[string]any
}

func (t *TemplateExecutor) Execute() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		zap.L().Error(fmt.Sprintf("Failed to get current working directory err=%v", err))
		return "", err
	}

	templatePath := filepath.Join(currentDir, "internal/server/templates", t.TemplateId, "template.yaml")

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		zap.L().Error(fmt.Sprintf("Template file not found: %s", templatePath))
		return "", fmt.Errorf("template file not found: %s", templatePath)
	}

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		zap.L().Error(fmt.Sprintf("Failed to parse template file %s err=%v", templatePath, err))
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, t.Parameters)
	if err != nil {
		zap.L().Error(fmt.Sprintf("Failed to execute template %s err=%v", templatePath, err))
		return "", err
	}

	userData, err := t.injectScripts(buf.Bytes())
	if err != nil {
		zap.L().Error(fmt.Sprintf("Failed to inject scripts into user data err=%v", err))
		return "", err
	}

	encodedStr := base64.StdEncoding.EncodeToString(userData)
	return encodedStr, nil
}

type WriteFile struct {
	Path        string `yaml:"path"`
	Content     string `yaml:"content"`
	Encoding    string `yaml:"encoding,omitempty"`
	Permissions string `yaml:"permissions,omitempty"`
	Defer       bool   `yaml:"defer,omitempty"`
}

type CloudConfig struct {
	WriteFiles []WriteFile    `yaml:"write_files,omitempty"`
	Other      map[string]any `yaml:",inline"`
}

func (t *TemplateExecutor) injectScripts(userData []byte) ([]byte, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	scriptsDir := filepath.Join(currentDir, "scripts")
	scriptFiles, err := filepath.Glob(filepath.Join(scriptsDir, "*.sh"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob scripts directory: %w", err)
	}

	if len(scriptFiles) == 0 {
		zap.L().Info("No scripts found to inject")
		return userData, nil
	}

	var cloudConfig CloudConfig
	if err := yaml.Unmarshal(userData, &cloudConfig); err != nil {
		return nil, fmt.Errorf("failed to parse user data as YAML: %w", err)
	}

	templateVarRegex := regexp.MustCompile(`\{\{\s*\.(\w+)\s*\}\}`) // Regex use to get template values such as {{ .ValueName }}

	for _, scriptPath := range scriptFiles {
		scriptContent, err := os.ReadFile(scriptPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read script %s: %w", scriptPath, err)
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
		scriptName := filepath.Base(scriptPath)
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

func RetrieveExecutor(templateId string, parameters string) (*TemplateExecutor, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	templateDir := filepath.Join(currentDir, "internal/server/templates", templateId)

	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("template id %v does not exist", templateId)
	}

	descriptionPath := filepath.Join(templateDir, "description.json")
	if _, err := os.Stat(descriptionPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template description not found for id %v", templateId)
	}

	templatePath := filepath.Join(templateDir, "template.yaml")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template file not found for id %v", templateId)
	}

	var params map[string]any
	if err := json.Unmarshal([]byte(parameters), &params); err != nil {
		return nil, fmt.Errorf("failed to parse body: %v", err)
	}

	zap.L().Info(fmt.Sprintf("Creating generic template executor for template: %s", templateId))

	return &TemplateExecutor{
		TemplateId: templateId,
		Parameters: params,
	}, nil
}
