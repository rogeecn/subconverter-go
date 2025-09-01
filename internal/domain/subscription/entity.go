package subscription

import (
	"time"

	"github.com/subconverter/subconverter-go/internal/domain/proxy"
	"github.com/subconverter/subconverter-go/internal/domain/ruleset"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusExpired  Status = "expired"
	StatusError    Status = "error"
)

type Subscription struct {
	ID          string            `json:"id" yaml:"id"`
	Name        string            `json:"name" yaml:"name"`
	URL         string            `json:"url" yaml:"url"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Status      Status            `json:"status" yaml:"status"`
	Proxies     []*proxy.Proxy    `json:"proxies" yaml:"proxies"`
	RuleSets    []*ruleset.RuleSet `json:"rule_sets,omitempty" yaml:"rule_sets,omitempty"`
	CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty" yaml:"expires_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type SubscriptionRequest struct {
	URL         string            `json:"url" validate:"required,url"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type SubscriptionResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Description string    `json:"description,omitempty"`
	Status      Status    `json:"status"`
	ProxyCount  int       `json:"proxy_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type UpdateRequest struct {
	Name        *string            `json:"name,omitempty"`
	Description *string            `json:"description,omitempty"`
	Metadata    map[string]string  `json:"metadata,omitempty"`
	RuleSets    []string           `json:"rule_sets,omitempty"`
}

type FilterOptions struct {
	IncludeTypes []proxy.Type `json:"include_types,omitempty"`
	ExcludeTypes []proxy.Type `json:"exclude_types,omitempty"`
	IncludeNames []string     `json:"include_names,omitempty"`
	ExcludeNames []string     `json:"exclude_names,omitempty"`
	CountryCodes []string     `json:"country_codes,omitempty"`
	MinSpeed     int64        `json:"min_speed,omitempty"`
	MaxLatency   int64        `json:"max_latency,omitempty"`
	SortBy       string       `json:"sort_by,omitempty"`
	SortOrder    string       `json:"sort_order,omitempty"`
}

type Statistics struct {
	TotalSubscriptions int64     `json:"total_subscriptions"`
	TotalProxies       int64     `json:"total_proxies"`
	ActiveSubscriptions int64    `json:"active_subscriptions"`
	LastUpdate         time.Time `json:"last_update"`
}