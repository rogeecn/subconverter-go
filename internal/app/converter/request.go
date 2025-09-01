package converter

import (
	"github.com/subconverter/subconverter-go/internal/app/generator"
	"github.com/subconverter/subconverter-go/internal/domain/proxy"
	"github.com/subconverter/subconverter-go/internal/domain/ruleset"
)

// ConvertRequest represents a conversion request
type ConvertRequest struct {
	Target    string         `json:"target" validate:"required"`
	URLs      []string       `json:"urls" validate:"required,gt=0"`
	ConfigURL string         `json:"config,omitempty"`
	Options   Options        `json:"options,omitempty"`
}

// Options contains conversion options
type Options struct {
	IncludeRemarks []string                `json:"include_remarks,omitempty"`
	ExcludeRemarks []string                `json:"exclude_remarks,omitempty"`
	RenameRules    []generator.RenameRule  `json:"rename_rules,omitempty"`
	EmojiRules     []generator.EmojiRule   `json:"emoji_rules,omitempty"`
	Sort           bool                    `json:"sort,omitempty"`
	UDP            bool                    `json:"udp,omitempty"`
	ProxyGroups    []generator.ProxyGroup  `json:"proxy_groups,omitempty"`
	Rules          []string                `json:"rules,omitempty"`
	BaseTemplate   string                  `json:"base_template,omitempty"`
	CustomOptions  map[string]interface{}  `json:"custom_options,omitempty"`
}

// ConvertResponse represents a conversion response
type ConvertResponse struct {
	Config    string              `json:"config"`
	Format    string              `json:"format"`
	Proxies   []*proxy.Proxy      `json:"proxies"`
	RuleSets  []*ruleset.RuleSet  `json:"rule_sets,omitempty"`
	Generated string              `json:"generated"`
}

// BatchConvertRequest represents a batch conversion request
type BatchConvertRequest struct {
	Requests []ConvertRequest `json:"requests" validate:"required,gt=0"`
}

// BatchConvertResponse represents a batch conversion response
type BatchConvertResponse struct {
	Results []ConvertResponse `json:"results"`
	Errors  []string          `json:"errors,omitempty"`
}

// ValidateRequest represents a validation request
type ValidateRequest struct {
	URL string `json:"url" validate:"required,url"`
}

// ValidateResponse represents a validation response
type ValidateResponse struct {
	Valid   bool   `json:"valid"`
	Format  string `json:"format,omitempty"`
	Proxies int    `json:"proxies"`
	Error   string `json:"error,omitempty"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// InfoResponse represents service information
type InfoResponse struct {
	Version        string   `json:"version"`
	SupportedTypes []string `json:"supported_types"`
	Features       []string `json:"features"`
}