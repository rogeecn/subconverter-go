package generator

import (
    "context"
    "fmt"
    "regexp"
    "strconv"
    "strings"

    "github.com/rogeecn/subconverter-go/internal/app/template"
    "github.com/rogeecn/subconverter-go/internal/domain/proxy"
    "github.com/rogeecn/subconverter-go/internal/domain/ruleset"
    "github.com/samber/lo"
    "gopkg.in/yaml.v3"
)

type ClashGenerator struct {
	templateManager *template.Manager
}

func NewClashGenerator(templateManager *template.Manager) *ClashGenerator {
	return &ClashGenerator{
		templateManager: templateManager,
	}
}

func (g *ClashGenerator) Format() string {
	return "clash"
}

func (g *ClashGenerator) ContentType() string {
	return "application/x-yaml"
}

func (g *ClashGenerator) Generate(ctx context.Context, proxies []*proxy.Proxy, rulesets []*ruleset.RuleSet, options GenerateOptions) (string, error) {
	// Load base template if specified
	if options.BaseTemplate != "" && g.templateManager != nil {
		templateData, err := g.templateManager.RenderTemplate(ctx, options.BaseTemplate, map[string]interface{}{
			"proxies":  proxies,
			"rulesets": rulesets,
			"options":  options,
		})
		if err == nil {
			return templateData, nil
		}
	}

	config := make(map[string]interface{})

	// Process proxies
	clashProxies := g.buildProxies(proxies)

	// Process proxy groups
	clashProxyGroups := g.buildProxyGroups(options.ProxyGroups, proxies)

	// Process rules
	clashRules := g.buildRules(rulesets, options.Rules)

	// Build configuration
	config["proxies"] = clashProxies
	config["proxy-groups"] = clashProxyGroups
	config["rules"] = clashRules

	// Add custom options
	for k, v := range options.CustomOptions {
		config[k] = v
	}

    data, err := yaml.Marshal(config)
    if err != nil {
        return "", fmt.Errorf("failed to marshal YAML: %v", err)
    }

    // Ensure unicode characters (e.g., emojis) are not shown as \UXXXXXXXX
    // Some environments may render surrogate escapes; convert them back.
    return unescapeYAMLUnicode(string(data)), nil
}

func (g *ClashGenerator) buildProxies(proxies []*proxy.Proxy) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(proxies))

	for _, p := range proxies {
		proxyMap := make(map[string]interface{})

		proxyMap["name"] = p.Name
		proxyMap["type"] = string(p.Type)
		proxyMap["server"] = p.Server
		proxyMap["port"] = p.Port

		switch p.Type {
		case "ss":
			proxyMap["cipher"] = p.Method
			proxyMap["password"] = p.Password
			if p.Plugin != "" {
				proxyMap["plugin"] = p.Plugin
				if p.PluginOpts != "" {
					proxyMap["plugin-opts"] = parsePluginOpts(p.PluginOpts)
				}
			}

		case "ssr":
			proxyMap["cipher"] = p.Method
			proxyMap["password"] = p.Password
			proxyMap["protocol"] = p.Protocol
			proxyMap["obfs"] = p.Obfs
			if p.ProtocolParam != "" {
				proxyMap["protocol-param"] = p.ProtocolParam
			}
			if p.ObfsParam != "" {
				proxyMap["obfs-param"] = p.ObfsParam
			}

		case "vmess":
			proxyMap["uuid"] = p.UUID
			proxyMap["alterId"] = p.AID
			proxyMap["cipher"] = p.Method
			if p.Network != "" {
				proxyMap["network"] = strings.ToLower(string(p.Network))
			}
			if p.TLS != proxy.TLSNone {
				proxyMap["tls"] = true
				if p.SNI != "" {
					proxyMap["servername"] = p.SNI
				}
				if p.SkipCertVerify {
					proxyMap["skip-cert-verify"] = true
				}
			}
			if p.Path != "" || p.Host != "" {
				proxyMap["ws-opts"] = map[string]interface{}{
					"path": p.Path,
					"headers": map[string]string{
						"Host": p.Host,
					},
				}
			}

		case "vless":
			proxyMap["uuid"] = p.UUID
			if p.Flow != "" {
				proxyMap["flow"] = p.Flow
			}
			if p.TLS != proxy.TLSNone {
				proxyMap["tls"] = true
				if p.SNI != "" {
					proxyMap["servername"] = p.SNI
				}
				if p.SkipCertVerify {
					proxyMap["skip-cert-verify"] = true
				}
			}
			if p.Path != "" || p.Host != "" {
				proxyMap["network"] = "ws"
				proxyMap["ws-opts"] = map[string]interface{}{
					"path": p.Path,
					"headers": map[string]string{
						"Host": p.Host,
					},
				}
			}
			if strings.EqualFold(p.Security, "reality") || p.RealityPublicKey != "" || p.RealityShortID != "" || p.RealitySpiderX != "" || p.ClientFingerprint != "" {
				if p.ClientFingerprint != "" {
					proxyMap["client-fingerprint"] = p.ClientFingerprint
				}
				reality := map[string]interface{}{}
				if p.RealityPublicKey != "" {
					reality["public-key"] = p.RealityPublicKey
				}
				if p.RealityShortID != "" {
					reality["short-id"] = p.RealityShortID
				}
				if p.RealitySpiderX != "" {
					reality["spider-x"] = p.RealitySpiderX
				}
				if len(reality) > 0 {
					proxyMap["reality-opts"] = reality
				}
			}

		case "trojan":
			proxyMap["password"] = p.Password
			if p.SNI != "" {
				proxyMap["sni"] = p.SNI
			}
			if p.SkipCertVerify {
				proxyMap["skip-cert-verify"] = true
			}

		case "hysteria":
			proxyMap["auth"] = p.Password
			proxyMap["up"] = fmt.Sprintf("%d Mbps", p.UpMbps)
			proxyMap["down"] = fmt.Sprintf("%d Mbps", p.DownMbps)
			if p.SNI != "" {
				proxyMap["sni"] = p.SNI
			}
			if p.SkipCertVerify {
				proxyMap["skip-cert-verify"] = true
			}

		case "hysteria2":
			proxyMap["password"] = p.Password
			if p.SNI != "" {
				proxyMap["sni"] = p.SNI
			}
			if p.SkipCertVerify {
				proxyMap["skip-cert-verify"] = true
			}
			if len(p.Alpn) > 0 {
				proxyMap["alpn"] = p.Alpn
			}

		case "snell":
			proxyMap["psk"] = p.Password
			proxyMap["version"] = 3

		case "http", "https":
			proxyMap["username"] = p.Username
			proxyMap["password"] = p.Password
			if p.Type == "https" {
				proxyMap["tls"] = true
				if p.SNI != "" {
					proxyMap["sni"] = p.SNI
				}
			}
		}

		if p.UDP {
			proxyMap["udp"] = true
		}

		result = append(result, proxyMap)
	}

	return result
}

func (g *ClashGenerator) buildProxyGroups(groups []ProxyGroup, proxies []*proxy.Proxy) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(groups))

	// Default groups if none provided
	if len(groups) == 0 {
		groups = []ProxyGroup{
			{
				Name:    "🚀 节点选择",
				Type:    "select",
				Proxies: []string{"♻️ 自动选择", "🔯 故障转移", "DIRECT"},
			},
			{
				Name:     "♻️ 自动选择",
				Type:     "url-test",
				Proxies:  []string{},
				URL:      "http://www.gstatic.com/generate_204",
				Interval: 300,
			},
			{
				Name:     "🔯 故障转移",
				Type:     "fallback",
				Proxies:  []string{},
				URL:      "http://www.gstatic.com/generate_204",
				Interval: 300,
			},
		}
	}

	// Build proxy names list
	proxyNames := lo.Map(proxies, func(p *proxy.Proxy, _ int) string {
		return p.Name
	})

	for _, group := range groups {
		groupMap := make(map[string]interface{})
		groupMap["name"] = group.Name
		groupMap["type"] = group.Type

		// Build proxies list
		proxiesList := make([]string, 0)
		if len(group.Proxies) > 0 {
			proxiesList = append(proxiesList, group.Proxies...)
		}

		// Apply filter if specified
		if group.Filter != "" {
			filtered := lo.Filter(proxyNames, func(name string, _ int) bool {
				return strings.Contains(name, group.Filter)
			})
			proxiesList = append(proxiesList, filtered...)
		} else {
			proxiesList = append(proxiesList, proxyNames...)
		}

		groupMap["proxies"] = proxiesList

		if group.URL != "" {
			groupMap["url"] = group.URL
		}
		if group.Interval > 0 {
			groupMap["interval"] = group.Interval
		}
		if group.Tolerance > 0 {
			groupMap["tolerance"] = group.Tolerance
		}

		result = append(result, groupMap)
	}

	return result
}

func (g *ClashGenerator) buildRules(rulesets []*ruleset.RuleSet, customRules []string) []string {
	result := make([]string, 0)

	// Add rules from rulesets
	for _, ruleset := range rulesets {
		if !ruleset.Enabled {
			continue
		}
		for _, rule := range ruleset.Rules {
			clashRule := g.convertRule(rule)
			if clashRule != "" {
				result = append(result, clashRule)
			}
		}
	}

	// Add custom rules
	result = append(result, customRules...)

	// Add default rule
	if len(result) == 0 {
		result = append(result, "MATCH,DIRECT")
	}

	return result
}

func (g *ClashGenerator) convertRule(rule ruleset.Rule) string {
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
	case ruleset.RuleTypeGeoIP:
		return fmt.Sprintf("GEOIP,%s,%s", rule.Value, rule.Proxy)
	case ruleset.RuleTypeFinal, ruleset.RuleTypeMatch:
		return fmt.Sprintf("MATCH,%s", rule.Proxy)
	default:
		return ""
	}
}

func parsePluginOpts(opts string) map[string]interface{} {
	result := make(map[string]interface{})
	parts := strings.Split(opts, ";")
	for _, part := range parts {
		if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return result
}

// unescapeYAMLUnicode replaces YAML-style \UXXXXXXXX escapes with actual runes for better readability
func unescapeYAMLUnicode(s string) string {
    re := regexp.MustCompile(`\\U([0-9a-fA-F]{8})`)
    return re.ReplaceAllStringFunc(s, func(m string) string {
        sub := re.FindStringSubmatch(m)
        if len(sub) != 2 {
            return m
        }
        v, err := strconv.ParseUint(sub[1], 16, 32)
        if err != nil {
            return m
        }
        r := rune(v)
        return string(r)
    })
}
