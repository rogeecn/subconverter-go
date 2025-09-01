package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/subconverter/subconverter-go/internal/domain/proxy"
	"github.com/subconverter/subconverter-go/internal/domain/ruleset"
)

type SurgeGenerator struct{}

func NewSurgeGenerator() *SurgeGenerator {
	return &SurgeGenerator{}
}

func (g *SurgeGenerator) Format() string {
	return "surge"
}

func (g *SurgeGenerator) ContentType() string {
	return "text/plain"
}

func (g *SurgeGenerator) Generate(ctx context.Context, proxies []*proxy.Proxy, rulesets []*ruleset.RuleSet, options GenerateOptions) (string, error) {
	var builder strings.Builder
	
	// Header
	builder.WriteString("#!MANAGED-CONFIG https://example.com interval=86400 strict=false\n\n")
	
	// General section
	builder.WriteString("[General]\n")
	builder.WriteString("loglevel = notify\n")
	builder.WriteString("dns-server = 8.8.8.8, 1.1.1.1\n")
	builder.WriteString("skip-proxy = 127.0.0.1, 192.168.0.0/16\n\n")
	
	// Proxy section
	builder.WriteString("[Proxy]\n")
	for _, proxy := range proxies {
		line := g.buildProxyLine(proxy)
		builder.WriteString(line)
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
	
	// Proxy Group section
	builder.WriteString("[Proxy Group]\n")
	for _, group := range options.ProxyGroups {
		line := g.buildProxyGroupLine(group)
		builder.WriteString(line)
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
	
	// Rule section
	builder.WriteString("[Rule]\n")
	for _, ruleset := range rulesets {
		if !ruleset.Enabled {
			continue
		}
		for _, rule := range ruleset.Rules {
			line := g.buildRuleLine(rule)
			builder.WriteString(line)
			builder.WriteString("\n")
		}
	}
	
	// Add default rule
	builder.WriteString("FINAL,DIRECT\n")
	
	return builder.String(), nil
}

func (g *SurgeGenerator) buildProxyLine(proxy *proxy.Proxy) string {
	var parts []string
	
	switch proxy.Type {
	case "ss":
		parts = []string{
			proxy.Name,
			"ss",
			proxy.Server,
			fmt.Sprintf("%d", proxy.Port),
			"encrypt-method=" + proxy.Method,
			"password=" + proxy.Password,
		}
	case "vmess":
		parts = []string{
			proxy.Name,
			"vmess",
			proxy.Server,
			fmt.Sprintf("%d", proxy.Port),
			"username=" + proxy.UUID,
			"ws=true",
			"ws-path=" + proxy.Path,
			"tls=true",
		}
	case "trojan":
		parts = []string{
			proxy.Name,
			"trojan",
			proxy.Server,
			fmt.Sprintf("%d", proxy.Port),
			"password=" + proxy.Password,
			"tls=true",
		}
	default:
		return fmt.Sprintf("# Unsupported proxy type: %s", proxy.Type)
	}
	
	return strings.Join(parts, " = ")
}

func (g *SurgeGenerator) buildProxyGroupLine(group ProxyGroup) string {
	var parts []string
	parts = append(parts, group.Name)
	parts = append(parts, "=", group.Type)
	
	if group.URL != "" {
		parts = append(parts, "url="+group.URL)
	}
	if group.Interval > 0 {
		parts = append(parts, fmt.Sprintf("interval=%d", group.Interval))
	}
	
	parts = append(parts, group.Proxies...)
	return strings.Join(parts, ", ")
}

func (g *SurgeGenerator) buildRuleLine(rule ruleset.Rule) string {
	switch rule.Type {
	case ruleset.RuleTypeDomain:
		return fmt.Sprintf("DOMAIN,%s,%s", rule.Value, rule.Proxy)
	case ruleset.RuleTypeDomainSuffix:
		return fmt.Sprintf("DOMAIN-SUFFIX,%s,%s", rule.Value, rule.Proxy)
	case ruleset.RuleTypeDomainKeyword:
		return fmt.Sprintf("DOMAIN-KEYWORD,%s,%s", rule.Value, rule.Proxy)
	case ruleset.RuleTypeIPCIDR:
		return fmt.Sprintf("IP-CIDR,%s,%s", rule.Value, rule.Proxy)
	case ruleset.RuleTypeIPCIDR6:
		return fmt.Sprintf("IP-CIDR6,%s,%s", rule.Value, rule.Proxy)
	case ruleset.RuleTypeFinal, ruleset.RuleTypeMatch:
		return fmt.Sprintf("FINAL,%s", rule.Proxy)
	default:
		return fmt.Sprintf("# Unsupported rule type: %s", rule.Type)
	}
}