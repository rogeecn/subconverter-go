package parser

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/subconverter/subconverter-go/internal/domain/proxy"
)

type SSParser struct{}

func NewSSParser() *SSParser {
	return &SSParser{}
}

func (p *SSParser) Type() proxy.Type {
	return proxy.TypeShadowsocks
}

func (p *SSParser) Support(content string) bool {
	return strings.HasPrefix(content, "ss://")
}

func (p *SSParser) Parse(ctx context.Context, content string) ([]*proxy.Proxy, error) {
	if !p.Support(content) {
		return nil, fmt.Errorf("invalid shadowsocks URL format")
	}

	// Remove the ss:// prefix
	content = strings.TrimPrefix(content, "ss://")
	
	// Try to parse as base64 first
	if decoded, err := base64.RawURLEncoding.DecodeString(content); err == nil {
		return p.parseLegacy(string(decoded))
	}
	
	// Try standard base64
	if decoded, err := base64.StdEncoding.DecodeString(content); err == nil {
		return p.parseLegacy(string(decoded))
	}
	
	// Try to parse as SIP002 format
	return p.parseSIP002(content)
}

func (p *SSParser) parseLegacy(decoded string) ([]*proxy.Proxy, error) {
	parts := strings.SplitN(decoded, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid shadowsocks legacy format")
	}

	methodPassword := strings.SplitN(parts[0], ":", 2)
	if len(methodPassword) != 2 {
		return nil, fmt.Errorf("invalid method and password format")
	}

	serverPort := strings.SplitN(parts[1], ":", 2)
	if len(serverPort) != 2 {
		return nil, fmt.Errorf("invalid server and port format")
	}

	port, err := strconv.Atoi(serverPort[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	result := &proxy.Proxy{
		ID:       uuid.New().String(),
		Type:     proxy.Type("ss"),
		Name:     fmt.Sprintf("SS-%s", serverPort[0]),
		Server:   serverPort[0],
		Port:     port,
		Method:   methodPassword[0],
		Password: methodPassword[1],
		UDP:      true,
	}

	return []*proxy.Proxy{result}, nil
}

func (p *SSParser) parseSIP002(content string) ([]*proxy.Proxy, error) {
	u, err := url.Parse("ss://" + content)
	if err != nil {
		return nil, fmt.Errorf("invalid SIP002 URL: %v", err)
	}

	// Parse user info
	userInfo, err := url.PathUnescape(u.User.String())
	if err != nil {
		return nil, fmt.Errorf("invalid user info: %v", err)
	}

	methodPassword := strings.SplitN(userInfo, ":", 2)
	if len(methodPassword) != 2 {
		return nil, fmt.Errorf("invalid method and password format")
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	// Parse plugin info
	plugin := u.Query().Get("plugin")
	pluginOpts := u.Query().Get("plugin-opts")

	// Parse name
	name := u.Fragment
	if name == "" {
		name = fmt.Sprintf("SS-%s", u.Hostname())
	}

	result := &proxy.Proxy{
		ID:       uuid.New().String(),
		Type:     proxy.Type("ss"),
		Name:     name,
		Server:   u.Hostname(),
		Port:     port,
		Method:   methodPassword[0],
		Password: methodPassword[1],
		Plugin:   plugin,
		PluginOpts: pluginOpts,
		UDP:      true,
	}

	return []*proxy.Proxy{result}, nil
}