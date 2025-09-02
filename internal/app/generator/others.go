package generator

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/url"
    "strings"

    "github.com/rogeecn/subconverter-go/internal/domain/proxy"
    "github.com/rogeecn/subconverter-go/internal/domain/ruleset"
)

// QuantumultGenerator generates Quantumult configuration
type QuantumultGenerator struct{}

func NewQuantumultGenerator() *QuantumultGenerator { return &QuantumultGenerator{} }
func (g *QuantumultGenerator) Format() string      { return "quantumult" }
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

func NewLoonGenerator() *LoonGenerator       { return &LoonGenerator{} }
func (g *LoonGenerator) Format() string      { return "loon" }
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

func NewV2RayGenerator() *V2RayGenerator      { return &V2RayGenerator{} }
func (g *V2RayGenerator) Format() string      { return "v2ray" }
func (g *V2RayGenerator) ContentType() string { return "text/plain" }
func (g *V2RayGenerator) Generate(ctx context.Context, proxies []*proxy.Proxy, rulesets []*ruleset.RuleSet, options GenerateOptions) (string, error) {
    // Build newline-joined subscription: each proxy rendered to its native link
    // Then base64-encode the whole text as V2RayN-style subscription
    var links []string
    for _, p := range proxies {
        if link := buildLinkForProxy(p); link != "" {
            links = append(links, link)
        }
    }
    joined := strings.Join(links, "\n")
    // Standard base64 encoding
    encoded := base64.StdEncoding.EncodeToString([]byte(joined))
    return encoded, nil
}

// buildLinkForProxy renders a proxy into its standard URI form
func buildLinkForProxy(p *proxy.Proxy) string {
    switch p.Type {
    case proxy.TypeShadowsocks:
        // ss://<base64(method:password)>@host:port[#name][?plugin=...]
        user := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", p.Method, p.Password)))
        base := fmt.Sprintf("ss://%s@%s:%d", user, p.Server, p.Port)
        q := ""
        if p.Plugin != "" {
            // SIP002 plugin param, inline opts allowed
            opt := p.Plugin
            if p.PluginOpts != "" {
                opt = opt + ";" + p.PluginOpts
            }
            q = "?plugin=" + url.QueryEscape(opt)
        }
        frag := ""
        if p.Name != "" {
            frag = "#" + url.QueryEscape(p.Name)
        }
        return base + q + frag

    case proxy.TypeVMess:
        // vmess://base64(JSON)
        payload := map[string]string{
            "v":    "2",
            "ps":   p.Name,
            "add":  p.Server,
            "port": fmt.Sprintf("%d", p.Port),
            "id":   p.UUID,
            "aid":  fmt.Sprintf("%d", p.AID),
            "scy":  p.Method,
            "net":  strings.ToLower(string(p.Network)),
            "type": "",
            "host": p.Host,
            "path": p.Path,
            "tls":  "",
            "sni":  p.SNI,
        }
        if p.TLS != proxy.TLSNone {
            payload["tls"] = "tls"
        }
        // ALPN and FP are optional; commonly omitted
        if len(p.Alpn) > 0 {
            payload["alpn"] = strings.Join(p.Alpn, ",")
        }
        b, _ := json.Marshal(payload)
        b64 := base64.StdEncoding.EncodeToString(b)
        return "vmess://" + b64

    case proxy.TypeVLESS:
        // vless://uuid@host:port?encryption=none&security=...&type=...&host=...&path=...&sni=...#name
        u := &url.URL{Scheme: "vless", Host: fmt.Sprintf("%s:%d", p.Server, p.Port)}
        u.User = url.User(p.UUID)
        q := u.Query()
        q.Set("encryption", "none")
        if p.TLS != proxy.TLSNone {
            // support tls/reality
            if strings.EqualFold(p.Security, "reality") {
                q.Set("security", "reality")
                if p.Flow != "" {
                    q.Set("flow", p.Flow)
                }
                if p.ClientFingerprint != "" {
                    q.Set("fp", p.ClientFingerprint)
                }
                if p.RealityPublicKey != "" {
                    q.Set("pbk", p.RealityPublicKey)
                }
                if p.RealityShortID != "" {
                    q.Set("sid", p.RealityShortID)
                }
                if p.RealitySpiderX != "" {
                    q.Set("spx", p.RealitySpiderX)
                }
            } else {
                q.Set("security", "tls")
            }
        }
        // Network type (ws/grpc)
        if strings.EqualFold(string(p.Network), "tcp") == false && string(p.Network) != "" {
            q.Set("type", strings.ToLower(string(p.Network)))
        }
        if p.Host != "" {
            q.Set("host", p.Host)
        }
        if p.Path != "" {
            if q.Get("type") == "grpc" {
                q.Set("serviceName", p.Path)
            } else {
                q.Set("path", p.Path)
            }
        }
        if p.SNI != "" {
            q.Set("sni", p.SNI)
        }
        if len(p.Alpn) > 0 {
            q.Set("alpn", strings.Join(p.Alpn, ","))
        }
        u.RawQuery = q.Encode()
        if p.Name != "" {
            u.Fragment = url.QueryEscape(p.Name)
        }
        return u.String()

    case proxy.TypeTrojan:
        // trojan://password@host:port?security=tls&type=ws&host=...&path=...&sni=...#name
        u := &url.URL{Scheme: "trojan", Host: fmt.Sprintf("%s:%d", p.Server, p.Port)}
        u.User = url.User(p.Password)
        q := u.Query()
        if p.TLS == proxy.TLSNone {
            q.Set("security", "none")
        } else {
            q.Set("security", "tls")
        }
        // transport
        if strings.EqualFold(string(p.Network), "ws") {
            q.Set("type", "ws")
        } else if strings.EqualFold(string(p.Network), "grpc") {
            q.Set("type", "grpc")
        }
        if p.Host != "" {
            q.Set("host", p.Host)
        }
        if p.Path != "" {
            if q.Get("type") == "grpc" {
                q.Set("serviceName", p.Path)
            } else {
                q.Set("path", p.Path)
            }
        }
        if p.SNI != "" {
            q.Set("sni", p.SNI)
        }
        if p.SkipCertVerify {
            q.Set("allowInsecure", "1")
        }
        if len(p.Alpn) > 0 {
            q.Set("alpn", strings.Join(p.Alpn, ","))
        }
        u.RawQuery = q.Encode()
        if p.Name != "" {
            u.Fragment = url.QueryEscape(p.Name)
        }
        return u.String()
    }
    return ""
}

// SurfboardGenerator generates Surfboard configuration
type SurfboardGenerator struct{}

func NewSurfboardGenerator() *SurfboardGenerator  { return &SurfboardGenerator{} }
func (g *SurfboardGenerator) Format() string      { return "surfboard" }
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
