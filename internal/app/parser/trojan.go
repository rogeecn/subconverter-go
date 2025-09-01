package parser

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/subconverter/subconverter-go/internal/domain/proxy"
)

type TrojanParser struct{}

func NewTrojanParser() *TrojanParser {
	return &TrojanParser{}
}

func (p *TrojanParser) Type() proxy.Type {
	return proxy.TypeTrojan
}

func (p *TrojanParser) Support(content string) bool {
	return strings.HasPrefix(content, "trojan://")
}

func (p *TrojanParser) Parse(ctx context.Context, content string) ([]*proxy.Proxy, error) {
	if !p.Support(content) {
		return nil, fmt.Errorf("invalid trojan URL format")
	}

	u, err := url.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trojan URL: %v", err)
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	name := u.Fragment
	if name == "" {
		name = fmt.Sprintf("Trojan-%s", u.Hostname())
	}

	// Parse query parameters
	query := u.Query()

	// Parse network
	network := proxy.NetworkTCP
	if net := query.Get("type"); net != "" {
		switch strings.ToLower(net) {
		case "tcp":
			network = proxy.NetworkTCP
		case "udp":
			network = proxy.NetworkUDP
		case "grpc":
			network = proxy.NetworkTCP // gRPC uses TCP
		case "ws":
			network = proxy.NetworkTCP // WebSocket uses TCP
		}
	}

	// Parse TLS
	tls := proxy.TLSRequire // Trojan always uses TLS

	// Parse security settings
	skipCertVerify := false
	if query.Get("allowInsecure") == "1" || query.Get("allowInsecure") == "true" {
		skipCertVerify = true
	}

	// Parse SNI
	sni := query.Get("sni")
	if sni == "" {
		sni = query.Get("peer")
	}

	// Parse ALPN
	var alpn []string
	if alpnStr := query.Get("alpn"); alpnStr != "" {
		alpn = strings.Split(alpnStr, ",")
	}

	// Parse transport settings
	host := query.Get("host")
	path := query.Get("path")

	// Handle gRPC settings
	if network == proxy.NetworkTCP && query.Get("type") == "grpc" {
		path = query.Get("serviceName")
	}

	// Handle WebSocket settings
	if network == proxy.NetworkTCP && query.Get("type") == "ws" {
		if host == "" {
			host = query.Get("host")
		}
	}

	result := &proxy.Proxy{
		ID:             uuid.New().String(),
		Type:           proxy.Type("trojan"),
		Name:           name,
		Server:         u.Hostname(),
		Port:           port,
		Password:       u.User.Username(),
		Network:        network,
		TLS:            tls,
		SNI:            sni,
		Host:           host,
		Path:           path,
		SkipCertVerify: skipCertVerify,
		Alpn:           alpn,
		UDP:            network == proxy.NetworkUDP || query.Get("udp") == "true",
	}

	return []*proxy.Proxy{result}, nil
}