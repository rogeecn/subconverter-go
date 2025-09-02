package parser

import (
    "context"
    "encoding/base64"
    "strings"

    "github.com/rogeecn/subconverter-go/internal/domain/proxy"
    "gopkg.in/yaml.v3"
)

// Parser defines the interface for parsing different proxy protocols
type Parser interface {
	// Parse parses the subscription content into proxy configurations
	Parse(ctx context.Context, content string) ([]*proxy.Proxy, error)

	// Support checks if the parser supports the given content format
	Support(content string) bool

	// Type returns the type of proxy this parser handles
	Type() proxy.Type
}

// Manager manages multiple parsers and dispatches parsing tasks
type Manager struct {
	parsers []Parser
}

// NewManager creates a new parser manager with all available parsers
func NewManager() *Manager {
	return &Manager{
		parsers: []Parser{
			NewSSParser(),
			NewSSRParser(),
			NewVMessParser(),
			NewVLESSParser(),
			NewTrojanParser(),
			NewHysteriaParser(),
			NewHysteria2Parser(),
			NewSnellParser(),
			NewHTTPParser(),
			NewSocks5Parser(),
		},
	}
}

// Parse parses subscription content using appropriate parser
func (m *Manager) Parse(ctx context.Context, content string) ([]*proxy.Proxy, error) {
    var allProxies []*proxy.Proxy

    // Preprocess: try to decode whole content if it looks like Base64 subscription
    preprocessed := preprocessContent(content)

    // Special-case: Clash YAML subscription (multi-line document)
    // Detect quickly to avoid invoking line-based parsers on entire blob
    if looksLikeClashYAML(preprocessed) {
        if proxies, ok := parseClashYAML(preprocessed); ok && len(proxies) > 0 {
            allProxies = append(allProxies, proxies...)
            return allProxies, nil
        }
        // if detection said it looks like clash but parse failed, fall back to line parsing
    }

    // Split content by lines and parse each line
    lines := splitContent(preprocessed)

	for _, line := range lines {
		line = cleanLine(line)
		if line == "" {
			continue
		}

		for _, parser := range m.parsers {
			if parser.Support(line) {
				proxies, err := parser.Parse(ctx, line)
				if err != nil {
					// Log error but continue processing other lines
					continue
				}
				allProxies = append(allProxies, proxies...)
				break
			}
		}
	}

	return allProxies, nil
}

// AddParser adds a custom parser to the manager
func (m *Manager) AddParser(parser Parser) {
	m.parsers = append(m.parsers, parser)
}

// GetParsers returns all registered parsers
func (m *Manager) GetParsers() []Parser {
	return m.parsers
}

// splitContent splits content into lines
func splitContent(content string) []string {
	return strings.Split(content, "\n")
}

// cleanLine removes whitespace and comments from a line
func cleanLine(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
		return ""
	}
	return line
}

// preprocessContent attempts to decode base64-encoded subscription blobs.
// If decoding yields recognizable proxy schemes, the decoded text is used.
func preprocessContent(content string) string {
	text := strings.TrimSpace(content)
	// Quick path: if it already contains any scheme markers, return as-is.
	if strings.Contains(text, "ss://") || strings.Contains(text, "ssr://") ||
		strings.Contains(text, "vmess://") || strings.Contains(text, "vless://") ||
		strings.Contains(text, "trojan://") || strings.Contains(text, "hysteria://") ||
		strings.Contains(text, "hysteria2://") || strings.Contains(text, "snell://") ||
		strings.Contains(text, "socks5://") || strings.Contains(text, "socks://") ||
		strings.Contains(text, "http://") || strings.Contains(text, "https://") {
		return text
	}
	// Remove spaces and newlines for decoding attempts
	compact := strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', '\t', ' ':
			return -1
		}
		return r
	}, text)
	// Try URL-safe base64 first
	if decoded, err := base64.RawURLEncoding.DecodeString(compact); err == nil {
		d := string(decoded)
		if looksLikeSubscription(d) {
			return d
		}
	}
	// Try standard base64 (no padding and with padding)
	if decoded, err := base64.StdEncoding.DecodeString(compact); err == nil {
		d := string(decoded)
		if looksLikeSubscription(d) {
			return d
		}
	} else {
		// Try to pad to multiple of 4
		if m := len(compact) % 4; m != 0 {
			padded := compact + strings.Repeat("=", 4-m)
			if decoded2, err2 := base64.StdEncoding.DecodeString(padded); err2 == nil {
				d := string(decoded2)
				if looksLikeSubscription(d) {
					return d
				}
			}
		}
	}
	return content
}

func looksLikeSubscription(s string) bool {
    return strings.Contains(s, "ss://") || strings.Contains(s, "ssr://") ||
        strings.Contains(s, "vmess://") || strings.Contains(s, "vless://") ||
        strings.Contains(s, "trojan://") || strings.Contains(s, "hysteria://") ||
        strings.Contains(s, "hysteria2://") || strings.Contains(s, "snell://")
}

// looksLikeClashYAML performs a lightweight detection for Clash YAML docs.
func looksLikeClashYAML(s string) bool {
    if !strings.Contains(s, "proxies:") {
        return false
    }
    // Basic YAML shape check: contains ':' lines and likely multi-line content
    if !strings.Contains(s, ":") || !strings.Contains(s, "\n") {
        return false
    }
    return true
}

// parseClashYAML parses a Clash-style YAML and extracts proxies.
func parseClashYAML(s string) ([]*proxy.Proxy, bool) {
    type wsOpts struct {
        Path    string            `yaml:"path"`
        Headers map[string]string `yaml:"headers"`
    }
    type clashProxy struct {
        Name           string   `yaml:"name"`
        Type           string   `yaml:"type"`
        Server         string   `yaml:"server"`
        Port           int      `yaml:"port"`
        UUID           string   `yaml:"uuid"`
        AlterID        int      `yaml:"alterId"`
        Cipher         string   `yaml:"cipher"`
        TLS            bool     `yaml:"tls"`
        ServerName     string   `yaml:"servername"`
        SNI            string   `yaml:"sni"`
        Network        string   `yaml:"network"`
        UDP            bool     `yaml:"udp"`
        SkipCertVerify bool     `yaml:"skip-cert-verify"`
        Username       string   `yaml:"username"`
        Password       string   `yaml:"password"`
        Alpn           []string `yaml:"alpn"`
        WSOpts         *wsOpts  `yaml:"ws-opts"`
        // hysteria/hysteria2/snell minimal fields we might ignore for now
    }
    var data struct {
        Proxies []clashProxy `yaml:"proxies"`
    }
    if err := yaml.Unmarshal([]byte(s), &data); err != nil {
        return nil, false
    }
    if len(data.Proxies) == 0 {
        return nil, false
    }

    res := make([]*proxy.Proxy, 0, len(data.Proxies))
    for _, cp := range data.Proxies {
        p := &proxy.Proxy{
            Type:           proxy.Type(cp.Type),
            Name:           cp.Name,
            Server:         cp.Server,
            Port:           cp.Port,
            UUID:           cp.UUID,
            AID:            cp.AlterID,
            Method:         cp.Cipher,
            UDP:            cp.UDP,
            SkipCertVerify: cp.SkipCertVerify,
            Username:       cp.Username,
            Password:       cp.Password,
            Alpn:           cp.Alpn,
        }
        // TLS mapping
        if cp.TLS {
            p.TLS = proxy.TLSRequire
        }
        // SNI / servername
        if cp.ServerName != "" {
            p.SNI = cp.ServerName
        } else if cp.SNI != "" {
            p.SNI = cp.SNI
        }
        // Network
        switch strings.ToLower(cp.Network) {
        case "tcp":
            p.Network = proxy.NetworkTCP
        case "udp":
            p.Network = proxy.NetworkUDP
        case "tcp,udp":
            p.Network = proxy.NetworkTCPUDP
        case "ws":
            p.Network = proxy.NetworkTCP // Clash uses ws atop TCP; keep TCP here
        }
        // WS opts
        if cp.WSOpts != nil {
            p.Path = cp.WSOpts.Path
            if host, ok := cp.WSOpts.Headers["Host"]; ok {
                p.Host = host
            }
            if p.Host == "" && len(cp.WSOpts.Headers) > 0 {
                // pick any header key case-insensitively
                for k, v := range cp.WSOpts.Headers {
                    if strings.EqualFold(k, "host") {
                        p.Host = v
                        break
                    }
                }
            }
        }

        res = append(res, p)
    }
    return res, true
}
