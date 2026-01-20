package templates

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
	"sync"
	"text/template"
)

//go:embed template/*
var metaTemplateFS embed.FS
var (
	globalStore *TemplateStore
	once        sync.Once
)

// TemplateStore manages both embedded meta-templates and runtime templates
type TemplateStore struct {
	metaFS    embed.FS
	runtime   map[string]Template
	runtimeMu sync.RWMutex
}

// NewTemplateStore creates a new TemplateStore with the embedded meta-templates
func NewTemplateStore() *TemplateStore {
	return &TemplateStore{
		metaFS:  metaTemplateFS,
		runtime: make(map[string]Template),
	}
}

// ListIDs returns all runtime template IDs
func (s *TemplateStore) ListIDs() []string {
	s.runtimeMu.RLock()
	defer s.runtimeMu.RUnlock()

	ids := make([]string, 0, len(s.runtime))
	for id := range s.runtime {
		ids = append(ids, id)
	}
	return ids
}

// ListDescriptions returns all runtime template descriptions
func (s *TemplateStore) ListDescriptions() []Description {
	s.runtimeMu.RLock()
	defer s.runtimeMu.RUnlock()

	descriptions := make([]Description, 0, len(s.runtime))
	for _, t := range s.runtime {
		descriptions = append(descriptions, t.Description)
	}
	return descriptions
}

// ListContents returns all runtime template contents (template.yaml)
func (s *TemplateStore) ListContents() map[string]string {
	s.runtimeMu.RLock()
	defer s.runtimeMu.RUnlock()

	contents := make(map[string]string, len(s.runtime))
	for id, t := range s.runtime {
		contents[id] = t.Content
	}
	return contents
}

// GetDescription returns the description for a specific template ID
func (s *TemplateStore) GetDescription(templateID string) (Description, error) {
	s.runtimeMu.RLock()
	defer s.runtimeMu.RUnlock()

	t, exists := s.runtime[templateID]
	if !exists {
		return Description{}, fmt.Errorf("template %s not found", templateID)
	}
	return t.Description, nil
}

// GetContent returns the template.yaml content for a specific template ID
func (s *TemplateStore) GetContent(templateID string) (string, error) {
	s.runtimeMu.RLock()
	defer s.runtimeMu.RUnlock()

	t, exists := s.runtime[templateID]
	if !exists {
		return "", fmt.Errorf("template %s not found", templateID)
	}
	return t.Content, nil
}

// Get returns the full template for a specific ID
func (s *TemplateStore) Get(templateID string) (Template, error) {
	s.runtimeMu.RLock()
	defer s.runtimeMu.RUnlock()

	t, exists := s.runtime[templateID]
	if !exists {
		return Template{}, fmt.Errorf("template %s not found", templateID)
	}
	return t, nil
}

// Exists checks if a template exists
func (s *TemplateStore) Exists(templateID string) bool {
	s.runtimeMu.RLock()
	defer s.runtimeMu.RUnlock()

	_, exists := s.runtime[templateID]
	return exists
}

// Create generates a new template from the meta-templates and stores it in runtime
func (s *TemplateStore) Create(gt GenericTemplate) error {
	s.runtimeMu.Lock()
	defer s.runtimeMu.Unlock()

	if _, exists := s.runtime[gt.Id]; exists {
		return fmt.Errorf("template %s already exists", gt.Id)
	}

	// Generate description.json
	descContent, err := s.executeMetaTemplate("description.json.templ", gt)
	if err != nil {
		return fmt.Errorf("failed to generate description.json: %w", err)
	}

	// Generate template.yaml
	yamlContent, err := s.executeMetaTemplate("template.yaml.templ", gt)
	if err != nil {
		return fmt.Errorf("failed to generate template.yaml: %w", err)
	}

	// Parse the generated description
	var desc Description
	if err := json.Unmarshal([]byte(descContent), &desc); err != nil {
		return fmt.Errorf("failed to parse generated description: %w", err)
	}

	s.runtime[gt.Id] = Template{
		Description: desc,
		Content:     yamlContent,
	}

	return nil
}

// Delete removes a runtime template by ID
func (s *TemplateStore) Delete(templateID string) error {
	s.runtimeMu.Lock()
	defer s.runtimeMu.Unlock()

	if _, exists := s.runtime[templateID]; !exists {
		return fmt.Errorf("template %s not found", templateID)
	}

	delete(s.runtime, templateID)
	return nil
}

// executeMetaTemplate executes a meta-template file against the generic template data
func (s *TemplateStore) executeMetaTemplate(filename string, gt GenericTemplate) (string, error) {
	content, err := s.metaFS.ReadFile("template/" + filename)
	if err != nil {
		return "", fmt.Errorf("failed to read meta template %s: %w", filename, err)
	}

	funcMap := template.FuncMap{
		"Capitalize": capitalize,
		"ToLower":    strings.ToLower,
		"sub": func(a, b int) int {
			return a - b
		},
	}

	tmpl, err := template.New(filename).Funcs(funcMap).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse meta template %s: %w", filename, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, gt); err != nil {
		return "", fmt.Errorf("failed to execute meta template %s: %w", filename, err)
	}

	return buf.String(), nil
}

// ListMetaTemplateFiles returns all files in the embedded meta-template directory
func (s *TemplateStore) ListMetaTemplateFiles() ([]string, error) {
	entries, err := fs.ReadDir(s.metaFS, "template")
	if err != nil {
		return nil, fmt.Errorf("failed to read meta template directory: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// GetMetaTemplateContent returns the content of a specific meta-template file
func (s *TemplateStore) GetMetaTemplateContent(filename string) ([]byte, error) {
	return s.metaFS.ReadFile("template/" + filename)
}

func capitalize(value string) string {
	if value == "" {
		return ""
	}
	return strings.ToUpper(string(value[0])) + strings.ToLower(value[1:])
}

func getTemplateStore() *TemplateStore {
	once.Do(func() {
		globalStore = NewTemplateStore()
	})

	return globalStore
}

func MustTemplateStore() *TemplateStore {
	return getTemplateStore()
}

func RetrieveExecutor(templateID string, parameters string) (*TemplateExecutor, error) {
	return NewTemplateExecutor(globalStore, templateID, parameters)
}
