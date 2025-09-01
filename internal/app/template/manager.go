package template

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/subconverter/subconverter-go/internal/pkg/errors"
	"github.com/subconverter/subconverter-go/internal/pkg/logger"
)

// TemplateManager manages template loading and rendering
type Manager struct {
	templatesDir string
	rulesDir     string
	logger       logger.Logger
	cache        map[string]*template.Template
}

// NewManager creates a new template manager
func NewManager(templatesDir, rulesDir string, logger logger.Logger) *Manager {
	return &Manager{
		templatesDir: templatesDir,
		rulesDir:     rulesDir,
		logger:       logger,
		cache:        make(map[string]*template.Template),
	}
}

// LoadTemplate loads a template from file
func (m *Manager) LoadTemplate(ctx context.Context, name string) (*template.Template, error) {
	if tmpl, exists := m.cache[name]; exists {
		return tmpl, nil
	}

	filePath := filepath.Join(m.templatesDir, name)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errors.NotFound("TEMPLATE_NOT_FOUND", fmt.Sprintf("template %s not found", name))
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to read template %s", name))
	}

	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse template %s", name))
	}

	m.cache[name] = tmpl
	return tmpl, nil
}

// RenderTemplate renders a template with data
func (m *Manager) RenderTemplate(ctx context.Context, name string, data interface{}) (string, error) {
	tmpl, err := m.LoadTemplate(ctx, name)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to render template %s", name))
	}

	return builder.String(), nil
}

// LoadRule loads a rule file from disk
func (m *Manager) LoadRule(ctx context.Context, rulePath string) ([]string, error) {
	fullPath := filepath.Join(m.rulesDir, rulePath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, errors.NotFound("RULE_NOT_FOUND", fmt.Sprintf("rule %s not found", rulePath))
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to read rule %s", rulePath))
	}

	lines := strings.Split(string(content), "\n")
	var rules []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			rules = append(rules, line)
		}
	}

	return rules, nil
}

// ListTemplates lists all available templates
func (m *Manager) ListTemplates(ctx context.Context) ([]string, error) {
	entries, err := os.ReadDir(m.templatesDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list templates")
	}

	var templates []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), ".tpl") || strings.HasSuffix(entry.Name(), ".yml") {
			templates = append(templates, entry.Name())
		}
	}

	return templates, nil
}

// ListRules lists all available rules
func (m *Manager) ListRules(ctx context.Context) ([]string, error) {
	var rules []string

	err := filepath.Walk(m.rulesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".list") {
			relPath, err := filepath.Rel(m.rulesDir, path)
			if err != nil {
				return err
			}
			rules = append(rules, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to list rules")
	}

	return rules, nil
}

// GetTemplatePath returns the full path to a template
func (m *Manager) GetTemplatePath(name string) string {
	return filepath.Join(m.templatesDir, name)
}

// GetRulesPath returns the full path to the rules directory
func (m *Manager) GetRulesPath() string {
	return m.rulesDir
}

// ClearCache clears the template cache
func (m *Manager) ClearCache() {
	m.cache = make(map[string]*template.Template)
}
