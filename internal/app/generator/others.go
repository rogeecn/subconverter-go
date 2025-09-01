package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/subconverter/subconverter-go/internal/domain/proxy"
	"github.com/subconverter/subconverter-go/internal/domain/ruleset"
)

// QuantumultGenerator generates Quantumult configuration
type QuantumultGenerator struct{}

func NewQuantumultGenerator() *QuantumultGenerator { return &QuantumultGenerator{} }
func (g *QuantumultGenerator) Format() string { return "quantumult" }
func (g *QuantumultGenerator) ContentType() string { return "text/plain" }
func (g *QuantumultGenerator) Generate(ctx context.Context, proxies []*proxy.Proxy, rulesets []*ruleset.RuleSet, options GenerateOptions) (string, error) {
	var builder strings.Builder
	for _, proxy := range proxies {
		builder.WriteString(g.buildProxyLine(proxy))
		builder.WriteString("\n")
	}
	return builder.String(), nil
}

func (g *QuantumultGenerator) buildProxyLine(proxy *proxy.Proxy) string {
	switch proxy.Type {
	case "ss":
		return fmt.Sprintf("shadowsocks=%s:%d, method=%s, password=%s, tag=%s", proxy.Server, proxy.Port, proxy.Method, proxy.Password, proxy.Name)
	case "vmess":
		return fmt.Sprintf("vmess=%s:%d, method=none, password=%s, tag=%s", proxy.Server, proxy.Port, proxy.UUID, proxy.Name)
	default:
		return fmt.Sprintf("# Unsupported: %s", proxy.Name)
	}
}

// LoonGenerator generates Loon configuration
type LoonGenerator struct{}

func NewLoonGenerator() *LoonGenerator { return &LoonGenerator{} }
func (g *LoonGenerator) Format() string { return "loon" }
func (g *LoonGenerator) ContentType() string { return "text/plain" }
func (g *LoonGenerator) Generate(ctx context.Context, proxies []*proxy.Proxy, rulesets []*ruleset.RuleSet, options GenerateOptions) (string, error) {
	var builder strings.Builder
	for _, proxy := range proxies {
		builder.WriteString(g.buildProxyLine(proxy))
		builder.WriteString("\n")
	}
	return builder.String(), nil
}

func (g *LoonGenerator) buildProxyLine(proxy *proxy.Proxy) string {
	switch proxy.Type {
	case "ss":
		return fmt.Sprintf("%s = ss, %s, %d, encrypt-method=%s, password=%s", proxy.Name, proxy.Server, proxy.Port, proxy.Method, proxy.Password)
	case "vmess":
		return fmt.Sprintf("%s = vmess, %s, %d, username=%s", proxy.Name, proxy.Server, proxy.Port, proxy.UUID)
	default:
		return fmt.Sprintf("# %s = %s, %s, %d", proxy.Name, proxy.Type, proxy.Server, proxy.Port)
	}
}

// V2RayGenerator generates V2Ray configuration
type V2RayGenerator struct{}

func NewV2RayGenerator() *V2RayGenerator { return &V2RayGenerator{} }
func (g *V2RayGenerator) Format() string { return "v2ray" }
func (g *V2RayGenerator) ContentType() string { return "application/json" }
func (g *V2RayGenerator) Generate(ctx context.Context, proxies []*proxy.Proxy, rulesets []*ruleset.RuleSet, options GenerateOptions) (string, error) {
	result := fmt.Sprintf(`{"outbounds": [{"protocol": "freedom", "tag": "direct"}]}`)
	return result, nil
}

// SurfboardGenerator generates Surfboard configuration
type SurfboardGenerator struct{}

func NewSurfboardGenerator() *SurfboardGenerator { return &SurfboardGenerator{} }
func (g *SurfboardGenerator) Format() string { return "surfboard" }
func (g *SurfboardGenerator) ContentType() string { return "text/plain" }
func (g *SurfboardGenerator) Generate(ctx context.Context, proxies []*proxy.Proxy, rulesets []*ruleset.RuleSet, options GenerateOptions) (string, error) {
	var builder strings.Builder
	builder.WriteString("#!MANAGED-CONFIG\n\n")
	for _, proxy := range proxies {
		builder.WriteString(g.buildProxyLine(proxy))
		builder.WriteString("\n")
	}
	return builder.String(), nil
}

func (g *SurfboardGenerator) buildProxyLine(proxy *proxy.Proxy) string {
	switch proxy.Type {
	case "ss":
		return fmt.Sprintf("%s = ss, %s, %d, encrypt-method=%s, password=%s", proxy.Name, proxy.Server, proxy.Port, proxy.Method, proxy.Password)
	default:
		return fmt.Sprintf("# %s = %s, %s, %d", proxy.Name, proxy.Type, proxy.Server, proxy.Port)
	}
}