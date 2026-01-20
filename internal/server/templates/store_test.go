package templates

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestNewTemplateStore(t *testing.T) {
	// Arrange & Act
	store := NewTemplateStore()

	// Assert
	if store == nil {
		t.Fatal("expected store to be non-nil")
	}
	if store.runtime == nil {
		t.Fatal("expected runtime map to be initialized")
	}
	if len(store.runtime) != 0 {
		t.Errorf("expected empty runtime map, got %d items", len(store.runtime))
	}
}

func TestTemplateStore_Create(t *testing.T) {
	t.Run("creates template successfully", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		gt := GenericTemplate{
			Id:          "test_template",
			Name:        "Test Template",
			Description: "A test template",
			Parameters: []Parameter{
				{Name: "ServerName", Description: "The server name"},
			},
			UpdatePackages:  true,
			UpgradePackages: false,
			Packages:        []string{"nginx"},
			Commands:        []string{"systemctl start nginx"},
		}

		// Act
		err := store.Create(gt)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !store.Exists("test_template") {
			t.Error("expected template to exist after creation")
		}
	})

	t.Run("returns error when template already exists", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		gt := GenericTemplate{
			Id:          "duplicate_template",
			Name:        "Duplicate Template",
			Description: "A duplicate template",
		}
		_ = store.Create(gt)

		// Act
		err := store.Create(gt)

		// Assert
		if err == nil {
			t.Fatal("expected error when creating duplicate template")
		}
		if err.Error() != "template duplicate_template already exists" {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestTemplateStore_Delete(t *testing.T) {
	t.Run("deletes existing template successfully", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		gt := GenericTemplate{
			Id:          "to_delete",
			Name:        "To Delete",
			Description: "Template to delete",
		}
		_ = store.Create(gt)

		// Act
		err := store.Delete("to_delete")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if store.Exists("to_delete") {
			t.Error("expected template to not exist after deletion")
		}
	})

	t.Run("returns error when template does not exist", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()

		// Act
		err := store.Delete("nonexistent")

		// Assert
		if err == nil {
			t.Fatal("expected error when deleting nonexistent template")
		}
		if err.Error() != "template nonexistent not found" {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestTemplateStore_Exists(t *testing.T) {
	t.Run("returns true for existing template", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		gt := GenericTemplate{
			Id:          "existing",
			Name:        "Existing",
			Description: "Existing template",
		}
		_ = store.Create(gt)

		// Act
		exists := store.Exists("existing")

		// Assert
		if !exists {
			t.Error("expected Exists to return true for existing template")
		}
	})

	t.Run("returns false for nonexistent template", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()

		// Act
		exists := store.Exists("nonexistent")

		// Assert
		if exists {
			t.Error("expected Exists to return false for nonexistent template")
		}
	})
}

func TestTemplateStore_ListIDs(t *testing.T) {
	t.Run("returns empty slice when no templates", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()

		// Act
		ids := store.ListIDs()

		// Assert
		if len(ids) != 0 {
			t.Errorf("expected empty slice, got %d items", len(ids))
		}
	})

	t.Run("returns all template IDs", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		templates := []GenericTemplate{
			{Id: "template_1", Name: "Template 1", Description: "First"},
			{Id: "template_2", Name: "Template 2", Description: "Second"},
			{Id: "template_3", Name: "Template 3", Description: "Third"},
		}
		for _, gt := range templates {
			_ = store.Create(gt)
		}

		// Act
		ids := store.ListIDs()

		// Assert
		if len(ids) != 3 {
			t.Errorf("expected 3 IDs, got %d", len(ids))
		}
		idMap := make(map[string]bool)
		for _, id := range ids {
			idMap[id] = true
		}
		for _, gt := range templates {
			if !idMap[gt.Id] {
				t.Errorf("expected ID %s to be in list", gt.Id)
			}
		}
	})
}

func TestTemplateStore_ListDescriptions(t *testing.T) {
	t.Run("returns empty slice when no templates", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()

		// Act
		descriptions := store.ListDescriptions()

		// Assert
		if len(descriptions) != 0 {
			t.Errorf("expected empty slice, got %d items", len(descriptions))
		}
	})

	t.Run("returns all template descriptions", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		_ = store.Create(GenericTemplate{
			Id:          "desc_test",
			Name:        "Description Test",
			Description: "Testing descriptions",
		})

		// Act
		descriptions := store.ListDescriptions()

		// Assert
		if len(descriptions) != 1 {
			t.Fatalf("expected 1 description, got %d", len(descriptions))
		}
		if descriptions[0].ID != "desc_test" {
			t.Errorf("expected ID 'desc_test', got '%s'", descriptions[0].ID)
		}
	})
}

func TestTemplateStore_ListContents(t *testing.T) {
	t.Run("returns empty map when no templates", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()

		// Act
		contents := store.ListContents()

		// Assert
		if len(contents) != 0 {
			t.Errorf("expected empty map, got %d items", len(contents))
		}
	})

	t.Run("returns all template contents", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		_ = store.Create(GenericTemplate{
			Id:          "content_test",
			Name:        "Content Test",
			Description: "Testing contents",
			Packages:    []string{"vim"},
		})

		// Act
		contents := store.ListContents()

		// Assert
		if len(contents) != 1 {
			t.Fatalf("expected 1 content entry, got %d", len(contents))
		}
		if _, exists := contents["content_test"]; !exists {
			t.Error("expected content for 'content_test' to exist")
		}
	})
}

func TestTemplateStore_GetDescription(t *testing.T) {
	t.Run("returns description for existing template", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		_ = store.Create(GenericTemplate{
			Id:          "get_desc",
			Name:        "Get Description",
			Description: "Test get description",
		})

		// Act
		desc, err := store.GetDescription("get_desc")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if desc.ID != "get_desc" {
			t.Errorf("expected ID 'get_desc', got '%s'", desc.ID)
		}
	})

	t.Run("returns error for nonexistent template", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()

		// Act
		_, err := store.GetDescription("nonexistent")

		// Assert
		if err == nil {
			t.Fatal("expected error for nonexistent template")
		}
	})
}

func TestTemplateStore_GetContent(t *testing.T) {
	t.Run("returns content for existing template", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		_ = store.Create(GenericTemplate{
			Id:             "get_content",
			Name:           "Get Content",
			Description:    "Test get content",
			UpdatePackages: true,
		})

		// Act
		content, err := store.GetContent("get_content")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if content == "" {
			t.Error("expected non-empty content")
		}
		if !strings.Contains(content, "#cloud-config") {
			t.Error("expected content to contain '#cloud-config'")
		}
	})

	t.Run("returns error for nonexistent template", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()

		// Act
		_, err := store.GetContent("nonexistent")

		// Assert
		if err == nil {
			t.Fatal("expected error for nonexistent template")
		}
	})
}

func TestTemplateStore_Get(t *testing.T) {
	t.Run("returns full template for existing ID", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()
		_ = store.Create(GenericTemplate{
			Id:          "full_template",
			Name:        "Full Template",
			Description: "Test full template",
		})

		// Act
		tmpl, err := store.Get("full_template")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if tmpl.Description.ID != "full_template" {
			t.Errorf("expected ID 'full_template', got '%s'", tmpl.Description.ID)
		}
		if tmpl.Content == "" {
			t.Error("expected non-empty content")
		}
	})

	t.Run("returns error for nonexistent template", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()

		// Act
		_, err := store.Get("nonexistent")

		// Assert
		if err == nil {
			t.Fatal("expected error for nonexistent template")
		}
	})
}

func TestTemplateStore_ListMetaTemplateFiles(t *testing.T) {
	// Arrange
	store := NewTemplateStore()

	// Act
	files, err := store.ListMetaTemplateFiles()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(files) < 2 {
		t.Errorf("expected at least 2 meta template files, got %d", len(files))
	}

	hasDescription := false
	hasTemplate := false
	for _, f := range files {
		if f == "description.json.templ" {
			hasDescription = true
		}
		if f == "template.yaml.templ" {
			hasTemplate = true
		}
	}
	if !hasDescription {
		t.Error("expected 'description.json.templ' in meta template files")
	}
	if !hasTemplate {
		t.Error("expected 'template.yaml.templ' in meta template files")
	}
}

func TestTemplateStore_GetMetaTemplateContent(t *testing.T) {
	t.Run("returns content for existing meta template", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()

		// Act
		content, err := store.GetMetaTemplateContent("description.json.templ")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(content) == 0 {
			t.Error("expected non-empty content")
		}
	})

	t.Run("returns error for nonexistent meta template", func(t *testing.T) {
		// Arrange
		store := NewTemplateStore()

		// Act
		_, err := store.GetMetaTemplateContent("nonexistent.templ")

		// Assert
		if err == nil {
			t.Fatal("expected error for nonexistent meta template")
		}
	})
}

func TestTemplateStore_ConcurrentAccess(t *testing.T) {
	// Arrange
	store := NewTemplateStore()
	var wg sync.WaitGroup
	numGoroutines := 10

	// Act - Create templates concurrently
	for i := range numGoroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			gt := GenericTemplate{
				Id:          fmt.Sprintf("concurrent_%d", idx),
				Name:        fmt.Sprintf("Concurrent %d", idx),
				Description: "Concurrent test",
			}
			_ = store.Create(gt)
		}(i)
	}
	wg.Wait()

	// Assert
	ids := store.ListIDs()
	if len(ids) != numGoroutines {
		t.Errorf("expected %d templates, got %d", numGoroutines, len(ids))
	}

	// Act - Read and delete concurrently
	for i := range numGoroutines {
		wg.Add(2)
		go func(idx int) {
			defer wg.Done()
			_ = store.Exists(fmt.Sprintf("concurrent_%d", idx))
		}(i)
		go func(idx int) {
			defer wg.Done()
			_ = store.Delete(fmt.Sprintf("concurrent_%d", idx))
		}(i)
	}
	wg.Wait()

	// Assert - All should be deleted
	ids = store.ListIDs()
	if len(ids) != 0 {
		t.Errorf("expected 0 templates after deletion, got %d", len(ids))
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"lowercase", "hello", "Hello"},
		{"uppercase", "HELLO", "Hello"},
		{"mixed case", "hELLO", "Hello"},
		{"single char", "a", "A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			result := capitalize(tt.input)

			// Assert
			if result != tt.expected {
				t.Errorf("capitalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMustTemplateStore(t *testing.T) {
	// Arrange & Act
	store1 := MustTemplateStore()
	store2 := MustTemplateStore()

	// Assert
	if store1 == nil {
		t.Fatal("expected store to be non-nil")
	}
	if store1 != store2 {
		t.Error("expected MustTemplateStore to return the same instance")
	}
}
