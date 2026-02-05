package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// TemplateCache manages template caching
type TemplateCache struct {
	cache map[string]*Jinja2Template
	mutex sync.RWMutex
}

// NewTemplateCache creates a new template cache
func NewTemplateCache() *TemplateCache {
	return &TemplateCache{
		cache: make(map[string]*Jinja2Template),
	}
}

// GetTemplate retrieves or loads a template
func (tc *TemplateCache) GetTemplate(templatePath string) (*Jinja2Template, error) {
	tc.mutex.RLock()
	if tmpl, exists := tc.cache[templatePath]; exists {
		tc.mutex.RUnlock()
		return tmpl, nil
	}
	tc.mutex.RUnlock()

	// Load template from file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	// Create new template
	tmpl, err := NewJinja2Template(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to create template from %s: %w", templatePath, err)
	}

	// Cache the template
	tc.mutex.Lock()
	tc.cache[templatePath] = tmpl
	tc.mutex.Unlock()

	return tmpl, nil
}

// ClearCache clears the template cache
func (tc *TemplateCache) ClearCache() {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.cache = make(map[string]*Jinja2Template)
}

// GetTemplatePaths returns common template paths
func GetTemplatePaths() []string {
	return []string{
		"./providers/models/template_files/",
		"./templates/",
		"/etc/agentforge/templates/",
		"../../providers/models/template_files/",
	}
}

// FindTemplate finds a template file in common locations
func FindTemplate(templateName string) (string, error) {
	// Try with .j2 extension first
	if !strings.HasSuffix(templateName, ".j2") {
		templateName += ".j2"
	}

	for _, basePath := range GetTemplatePaths() {
		fullPath := filepath.Join(basePath, templateName)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("template %s not found in any of the standard paths", templateName)
}
