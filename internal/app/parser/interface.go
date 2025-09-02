package parser

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/rogeecn/subconverter-go/internal/domain/proxy"
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
