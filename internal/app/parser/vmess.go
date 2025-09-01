package parser

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/subconverter/subconverter-go/internal/domain/proxy"
)

type VMessParser struct{}

func NewVMessParser() *VMessParser {
	return &VMessParser{}
}

func (p *VMessParser) Type() proxy.Type {
	return proxy.TypeVMess
}

func (p *VMessParser) Support(content string) bool {
	return strings.HasPrefix(content, "vmess://")
}

func (p *VMessParser) Parse(ctx context.Context, content string) ([]*proxy.Proxy, error) {
	if !p.Support(content) {
		return nil, fmt.Errorf("invalid vmess URL format")
	}

	// Remove the vmess:// prefix
	content = strings.TrimPrefix(content, "vmess://")
	
	// Decode base64
	decoded, err := base64.RawURLEncoding.DecodeString(content)
	if err != nil {
		// Try standard base64
		decoded, err = base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %v", err)
		}
	}

	var config struct {
		V  string `json:"v"`
		PS string `json:"ps"`
		Add string `json:"add"`
		Port string `json:"port"`
		ID string `json:"id"`
		AID string `json:"aid"`
		Scy string `json:"scy"`
		Net string `json:"net"`
		Type string `json:"type"`
		Host string `json:"host"`
		Path string `json:"path"`
		TLS string `json:"tls"`
		SNI string `json:"sni"`
		Alpn string `json:"alpn"`
		FP string `json:"fp"`
	}

	if err := json.Unmarshal(decoded, &config); err != nil {
		return nil, fmt.Errorf("failed to parse vmess config: %v", err)
	}

	port, err := strconv.Atoi(config.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	aid, err := strconv.Atoi(config.AID)
	if err != nil {
		aid = 0
	}

	name := config.PS
	if name == "" {
		name = fmt.Sprintf("VMess-%s", config.Add)
	}

	// Parse network
	network := proxy.NetworkTCP
	if config.Net != "" {
		switch strings.ToLower(config.Net) {
		case "tcp":
			network = proxy.NetworkTCP
		case "udp":
			network = proxy.NetworkUDP
		case "tcp,udp":
			network = proxy.NetworkTCPUDP
		}
	}

	// Parse TLS
	tls := proxy.TLSNone
	if strings.ToLower(config.TLS) == "tls" {
		tls = proxy.TLSRequire
	}

	// Parse headers
	headers := make(map[string]string)
	if config.Host != "" {
		headers["Host"] = config.Host
	}

	// Parse ALPN
	var alpn []string
	if config.Alpn != "" {
		alpn = strings.Split(config.Alpn, ",")
	}

	result := &proxy.Proxy{
		ID:       uuid.New().String(),
		Type:     proxy.Type("vmess"),
		Name:     name,
		Server:   config.Add,
		Port:     port,
		UUID:     config.ID,
		AID:      aid,
		Method:   config.Scy,
		Network:  network,
		TLS:      tls,
		SNI:      config.SNI,
		Host:     config.Host,
		Path:     config.Path,
		Headers:  headers,
		Alpn:     alpn,
		UDP:      network == proxy.NetworkUDP || network == proxy.NetworkTCPUDP,
	}

	return []*proxy.Proxy{result}, nil
}