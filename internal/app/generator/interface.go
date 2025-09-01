package generator

import (
	"context"
	"fmt"

	"github.com/subconverter/subconverter-go/internal/domain/proxy"
	"github.com/subconverter/subconverter-go/internal/domain/ruleset"
)

// Generator defines the interface for generating configuration files
type Generator interface {
	// Generate creates a configuration file from proxies and rules
	Generate(ctx context.Context, proxies []*proxy.Proxy, rulesets []*ruleset.RuleSet, options GenerateOptions) (string, error)
	
	// ContentType returns the MIME type of the generated configuration
	ContentType() string
	
	// Format returns the format identifier
	Format() string
}

// GenerateOptions contains options for configuration generation
type GenerateOptions struct {
	ProxyGroups    []ProxyGroup   `json:"proxy_groups"`
	Rules          []string       `json:"rules"`
	Template       string         `json:"template,omitempty"`
	BaseTemplate   string         `json:"base_template,omitempty"`
	RenameRules    []RenameRule   `json:"rename_rules,omitempty"`
	EmojiRules     []EmojiRule    `json:"emoji_rules,omitempty"`
	SortProxies    bool           `json:"sort_proxies"`
	UDPEnabled     bool           `json:"udp_enabled"`
	SkipFailed     bool           `json:"skip_failed"`
	CustomOptions  map[string]interface{} `json:"custom_options,omitempty"`
}

// ProxyGroup represents a proxy group configuration
type ProxyGroup struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Proxies   []string `json:"proxies"`
	URL       string   `json:"url,omitempty"`
	Interval  int      `json:"interval,omitempty"`
	Tolerance int      `json:"tolerance,omitempty"`
	Filter    string   `json:"filter,omitempty"`
	Strategy  string   `json:"strategy,omitempty"`
}

// RenameRule defines rules for renaming proxies
type RenameRule struct {
	Match   string `json:"match"`
	Replace string `json:"replace"`
}

// EmojiRule defines rules for adding emojis to proxy names
type EmojiRule struct {
	Match string `json:"match"`
	Emoji string `json:"emoji"`
}

// Manager manages multiple generators
type Manager struct {
	generators map[string]Generator
}

// NewManager creates a new generator manager
func NewManager() *Manager {
	return &Manager{
		generators: make(map[string]Generator),
	}
}

// Register adds a generator to the manager
func (m *Manager) Register(format string, generator Generator) {
	m.generators[format] = generator
}

// Get returns a generator for the specified format
func (m *Manager) Get(format string) (Generator, bool) {
	generator, exists := m.generators[format]
	return generator, exists
}

// SupportedFormats returns all supported formats
func (m *Manager) SupportedFormats() []string {
	formats := make([]string, 0, len(m.generators))
	for format := range m.generators {
		formats = append(formats, format)
	}
	return formats
}

// Generate generates configuration for the specified format
func (m *Manager) Generate(ctx context.Context, format string, proxies []*proxy.Proxy, rulesets []*ruleset.RuleSet, options GenerateOptions) (string, error) {
	generator, exists := m.generators[format]
	if !exists {
		return "", fmt.Errorf("unsupported format: %s", format)
	}
	
	return generator.Generate(ctx, proxies, rulesets, options)
}