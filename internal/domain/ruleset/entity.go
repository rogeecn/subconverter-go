package ruleset

import (
	"time"
)

type Type string

const (
	TypeSurge      Type = "surge"
	TypeClash      Type = "clash"
	TypeQuantumult Type = "quantumult"
	TypeLoon       Type = "loon"
	TypeSurfboard  Type = "surfboard"
	TypeV2Ray      Type = "v2ray"
)

type RuleType string

const (
	RuleTypeDomain      RuleType = "DOMAIN"
	RuleTypeDomainSuffix RuleType = "DOMAIN-SUFFIX"
	RuleTypeDomainKeyword RuleType = "DOMAIN-KEYWORD"
	RuleTypeIPCIDR      RuleType = "IP-CIDR"
	RuleTypeIPCIDR6     RuleType = "IP-CIDR6"
	RuleTypeGeoIP       RuleType = "GEOIP"
	RuleTypeUserAgent   RuleType = "USER-AGENT"
	RuleTypeURLRegex    RuleType = "URL-REGEX"
	RuleTypeFinal       RuleType = "FINAL"
	RuleTypeMatch       RuleType = "MATCH"
)

type Rule struct {
	Type      RuleType `json:"type" yaml:"type"`
	Value     string   `json:"value" yaml:"value"`
	Proxy     string   `json:"proxy,omitempty" yaml:"proxy,omitempty"`
	NoResolve bool     `json:"no-resolve,omitempty" yaml:"no-resolve,omitempty"`
	Policy    string   `json:"policy,omitempty" yaml:"policy,omitempty"`
}

type RuleSet struct {
	ID        string    `json:"id" yaml:"id"`
	Name      string    `json:"name" yaml:"name"`
	Type      Type      `json:"type" yaml:"type"`
	Rules     []Rule    `json:"rules" yaml:"rules"`
	Source    string    `json:"source,omitempty" yaml:"source,omitempty"`
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
	Enabled   bool      `json:"enabled" yaml:"enabled"`
}

type ProxyGroupType string

const (
	ProxyGroupSelect     ProxyGroupType = "select"
	ProxyGroupURLTest    ProxyGroupType = "url-test"
	ProxyGroupFallback   ProxyGroupType = "fallback"
	ProxyGroupLoadBalance ProxyGroupType = "load-balance"
	ProxyGroupRelay      ProxyGroupType = "relay"
)

type ProxyGroup struct {
	Name      string         `json:"name" yaml:"name"`
	Type      ProxyGroupType `json:"type" yaml:"type"`
	Proxies   []string      `json:"proxies" yaml:"proxies"`
	URL       string        `json:"url,omitempty" yaml:"url,omitempty"`
	Interval  int           `json:"interval,omitempty" yaml:"interval,omitempty"`
	Tolerance int           `json:"tolerance,omitempty" yaml:"tolerance,omitempty"`
	Filter    string        `json:"filter,omitempty" yaml:"filter,omitempty"`
	Strategy  string        `json:"strategy,omitempty" yaml:"strategy,omitempty"`
}