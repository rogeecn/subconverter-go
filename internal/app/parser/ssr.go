package parser

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/subconverter/subconverter-go/internal/domain/proxy"
)

type SSRParser struct{}

func NewSSRParser() *SSRParser {
	return &SSRParser{}
}

func (p *SSRParser) Type() proxy.Type {
	return proxy.TypeShadowsocksR
}

func (p *SSRParser) Support(content string) bool {
	return strings.HasPrefix(content, "ssr://")
}

func (p *SSRParser) Parse(ctx context.Context, content string) ([]*proxy.Proxy, error) {
	if !p.Support(content) {
		return nil, fmt.Errorf("invalid shadowsocksr URL format")
	}

	// Remove the ssr:// prefix
	content = strings.TrimPrefix(content, "ssr://")
	
	// Decode base64
	decoded, err := base64.RawURLEncoding.DecodeString(content)
	if err != nil {
		// Try standard base64
		decoded, err = base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %v", err)
		}
	}

	// Parse SSR format: server:port:protocol:method:obfs:password_base64/?params_base64
	parts := strings.SplitN(string(decoded), "/?", 2)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid SSR format")
	}

	// Parse basic info
	basicParts := strings.Split(parts[0], ":")
	if len(basicParts) != 6 {
		return nil, fmt.Errorf("invalid SSR basic format")
	}

	server := basicParts[0]
	port, err := strconv.Atoi(basicParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	protocol := basicParts[2]
	method := basicParts[3]
	obfs := basicParts[4]

	// Decode password
	passwordDecoded, err := base64.RawURLEncoding.DecodeString(basicParts[5])
	if err != nil {
		passwordDecoded, err = base64.StdEncoding.DecodeString(basicParts[5])
		if err != nil {
			return nil, fmt.Errorf("failed to decode password: %v", err)
		}
	}
	password := string(passwordDecoded)

	// Parse parameters
	params := make(map[string]string)
	if len(parts) > 1 {
		paramPairs := strings.Split(parts[1], "&")
		for _, pair := range paramPairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				key := kv[0]
				value := kv[1]
				
				// Decode parameter values
				if decoded, err := base64.RawURLEncoding.DecodeString(value); err == nil {
					value = string(decoded)
				} else if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
					value = string(decoded)
				}
				
				params[key] = value
			}
		}
	}

	name := params["remarks"]
	if name == "" {
		name = fmt.Sprintf("SSR-%s", server)
	}

	obfsParam := params["obfsparam"]
	protocolParam := params["protoparam"]
	group := params["group"]
	if group != "" && name == "" {
		name = group
	}

	result := &proxy.Proxy{
		ID:            uuid.New().String(),
		Type:          proxy.Type("ssr"),
		Name:          name,
		Server:        server,
		Port:          port,
		Password:      password,
		Method:        method,
		Protocol:      protocol,
		Obfs:          obfs,
		ProtocolParam: protocolParam,
		ObfsParam:     obfsParam,
		UDP:           true,
	}

	return []*proxy.Proxy{result}, nil
}