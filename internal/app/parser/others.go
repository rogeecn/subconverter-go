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

// VLESSParser parses VLESS protocol URLs
type VLESSParser struct{}

func NewVLESSParser() *VLESSParser { return &VLESSParser{} }
func (p *VLESSParser) Type() proxy.Type { return proxy.Type("vless") }
func (p *VLESSParser) Support(content string) bool { return strings.HasPrefix(content, "vless://") }
func (p *VLESSParser) Parse(ctx context.Context, content string) ([]*proxy.Proxy, error) {
    if !p.Support(content) {
        return nil, fmt.Errorf("invalid vless URL format")
    }
    u, err := url.Parse(content)
    if err != nil {
        return nil, fmt.Errorf("failed to parse vless URL: %v", err)
    }
    port, _ := strconv.Atoi(u.Port())
    name := u.Fragment
    if name == "" {
        name = fmt.Sprintf("VLESS-%s", u.Hostname())
    }
    q := u.Query()

    // Base fields
    result := &proxy.Proxy{
        ID:     uuid.New().String(),
        Type:   proxy.Type("vless"),
        Name:   name,
        Server: u.Hostname(),
        Port:   port,
        UUID:   u.User.Username(),
        UDP:    true,
    }

    // Security/TLS/Reality
    sec := strings.ToLower(q.Get("security"))
    result.Security = sec
    if sec == "tls" || sec == "reality" {
        result.TLS = proxy.TLSRequire
    }
    if sni := q.Get("sni"); sni != "" {
        result.SNI = sni
    }
    if q.Get("allowInsecure") == "1" || strings.EqualFold(q.Get("insecure"), "1") || strings.EqualFold(q.Get("insecure"), "true") {
        result.SkipCertVerify = true
    }

    // VLESS Reality extras
    if sec == "reality" {
        result.Flow = q.Get("flow")
        result.ClientFingerprint = q.Get("fp")
        result.RealityPublicKey = q.Get("pbk")
        result.RealityShortID = q.Get("sid")
        if spx := q.Get("spx"); spx != "" {
            result.RealitySpiderX = spx
        }
    }

    // Transport (WS/GRPC)
    net := strings.ToLower(q.Get("type"))
    if net == "" {
        net = strings.ToLower(q.Get("network"))
    }
    switch net {
    case "ws":
        result.Network = proxy.NetworkTCP
        // WebSocket parameters
        result.Path = q.Get("path")
        result.Host = q.Get("host")
    case "grpc":
        result.Network = proxy.NetworkTCP
        // serviceName as path equivalent
        if svc := q.Get("serviceName"); svc != "" {
            result.Path = svc
        }
    default:
        // default TCP
        result.Network = proxy.NetworkTCP
    }

    // ALPN
    if alpn := q.Get("alpn"); alpn != "" {
        result.Alpn = strings.Split(alpn, ",")
    }

    return []*proxy.Proxy{result}, nil
}

// HysteriaParser parses Hysteria protocol URLs
type HysteriaParser struct{}

func NewHysteriaParser() *HysteriaParser { return &HysteriaParser{} }
func (p *HysteriaParser) Type() proxy.Type { return proxy.Type("hysteria") }
func (p *HysteriaParser) Support(content string) bool { return strings.HasPrefix(content, "hysteria://") }
func (p *HysteriaParser) Parse(ctx context.Context, content string) ([]*proxy.Proxy, error) {
	if !p.Support(content) {
		return nil, fmt.Errorf("invalid hysteria URL format")
	}
	u, err := url.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hysteria URL: %v", err)
	}
	port, _ := strconv.Atoi(u.Port())
	name := u.Fragment
	if name == "" {
		name = fmt.Sprintf("Hysteria-%s", u.Hostname())
	}
	result := &proxy.Proxy{
		ID: uuid.New().String(),
		Type: proxy.Type("hysteria"),
		Name: name,
		Server: u.Hostname(),
		Port: port,
		Password: u.User.Username(),
		UDP: true,
	}
	return []*proxy.Proxy{result}, nil
}

// Hysteria2Parser parses Hysteria2 protocol URLs
type Hysteria2Parser struct{}

func NewHysteria2Parser() *Hysteria2Parser { return &Hysteria2Parser{} }
func (p *Hysteria2Parser) Type() proxy.Type { return proxy.Type("hysteria2") }
func (p *Hysteria2Parser) Support(content string) bool { return strings.HasPrefix(content, "hysteria2://") }
func (p *Hysteria2Parser) Parse(ctx context.Context, content string) ([]*proxy.Proxy, error) {
	if !p.Support(content) {
		return nil, fmt.Errorf("invalid hysteria2 URL format")
	}
	u, err := url.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hysteria2 URL: %v", err)
	}
	port, _ := strconv.Atoi(u.Port())
	name := u.Fragment
	if name == "" {
		name = fmt.Sprintf("Hysteria2-%s", u.Hostname())
	}
    q := u.Query()
    result := &proxy.Proxy{
        ID: uuid.New().String(),
        Type: proxy.Type("hysteria2"),
        Name: name,
        Server: u.Hostname(),
        Port: port,
        Password: u.User.Username(),
        UDP: true,
    }
    if sni := q.Get("sni"); sni != "" { result.SNI = sni }
    if alpn := q.Get("alpn"); alpn != "" { result.Alpn = strings.Split(alpn, ",") }
    if q.Get("insecure") == "1" || strings.EqualFold(q.Get("insecure"), "true") { result.SkipCertVerify = true }
    return []*proxy.Proxy{result}, nil
}

// SnellParser parses Snell protocol URLs
type SnellParser struct{}

func NewSnellParser() *SnellParser { return &SnellParser{} }
func (p *SnellParser) Type() proxy.Type { return proxy.Type("snell") }
func (p *SnellParser) Support(content string) bool { return strings.HasPrefix(content, "snell://") }
func (p *SnellParser) Parse(ctx context.Context, content string) ([]*proxy.Proxy, error) {
	if !p.Support(content) {
		return nil, fmt.Errorf("invalid snell URL format")
	}
	u, err := url.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snell URL: %v", err)
	}
	port, _ := strconv.Atoi(u.Port())
	name := u.Fragment
	if name == "" {
		name = fmt.Sprintf("Snell-%s", u.Hostname())
	}
	result := &proxy.Proxy{
		ID: uuid.New().String(),
		Type: proxy.Type("snell"),
		Name: name,
		Server: u.Hostname(),
		Port: port,
		Password: u.User.Username(),
		UDP: true,
	}
	return []*proxy.Proxy{result}, nil
}

// HTTPParser parses HTTP/HTTPS protocol URLs
type HTTPParser struct{}

func NewHTTPParser() *HTTPParser { return &HTTPParser{} }
func (p *HTTPParser) Type() proxy.Type { return proxy.Type("http") }
func (p *HTTPParser) Support(content string) bool { 
	return strings.HasPrefix(content, "http://") || strings.HasPrefix(content, "https://") 
}
func (p *HTTPParser) Parse(ctx context.Context, content string) ([]*proxy.Proxy, error) {
	u, err := url.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTTP URL: %v", err)
	}
	port, _ := strconv.Atoi(u.Port())
	if port == 0 {
		if u.Scheme == "https" {
			port = 443
		} else {
			port = 80
		}
	}
	name := u.Fragment
	if name == "" {
		name = fmt.Sprintf("HTTP-%s", u.Hostname())
	}
	proxyType := proxy.Type("http")
	if u.Scheme == "https" {
		proxyType = proxy.Type("https")
	}
	result := &proxy.Proxy{
		ID: uuid.New().String(),
		Type: proxyType,
		Name: name,
		Server: u.Hostname(),
		Port: port,
		Username: u.User.Username(),
		Password: "",
		UDP: false,
	}
	if u.User != nil {
		result.Password, _ = u.User.Password()
	}
	return []*proxy.Proxy{result}, nil
}

// Socks5Parser parses SOCKS5 protocol URLs
type Socks5Parser struct{}

func NewSocks5Parser() *Socks5Parser { return &Socks5Parser{} }
func (p *Socks5Parser) Type() proxy.Type { return proxy.Type("socks5") }
func (p *Socks5Parser) Support(content string) bool { 
	return strings.HasPrefix(content, "socks5://") || strings.HasPrefix(content, "socks://") 
}
func (p *Socks5Parser) Parse(ctx context.Context, content string) ([]*proxy.Proxy, error) {
	if !strings.HasPrefix(content, "socks5://") {
		content = strings.Replace(content, "socks://", "socks5://", 1)
	}
	u, err := url.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SOCKS5 URL: %v", err)
	}
	port, _ := strconv.Atoi(u.Port())
	if port == 0 {
		port = 1080
	}
	name := u.Fragment
	if name == "" {
		name = fmt.Sprintf("SOCKS5-%s", u.Hostname())
	}
	result := &proxy.Proxy{
		ID: uuid.New().String(),
		Type: proxy.Type("socks5"),
		Name: name,
		Server: u.Hostname(),
		Port: port,
		Username: u.User.Username(),
		Password: "",
		UDP: true,
	}
	if u.User != nil {
		result.Password, _ = u.User.Password()
	}
	return []*proxy.Proxy{result}, nil
}
