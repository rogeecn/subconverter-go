package proxy

import (
	"time"
)

type Type string

const (
	TypeShadowsocks  Type = "ss"
	TypeShadowsocksR Type = "ssr"
	TypeVMess        Type = "vmess"
	TypeVLESS        Type = "vless"
	TypeTrojan       Type = "trojan"
	TypeHysteria     Type = "hysteria"
	TypeHysteria2    Type = "hysteria2"
	TypeSnell        Type = "snell"
	TypeHTTP         Type = "http"
	TypeHTTPS        Type = "https"
	TypeSocks5       Type = "socks5"
)

type Network string

const (
	NetworkTCP    Network = "tcp"
	NetworkUDP    Network = "udp"
	NetworkTCPUDP Network = "tcp,udp"
)

type TLS string

const (
	TLSNone     TLS = "none"
	TLSRequest  TLS = "request"
	TLSRequire  TLS = "require"
	TLSVerify   TLS = "verify"
	TLSNoVerify TLS = "no-verify"
)

type Proxy struct {
	ID             string            `json:"id" yaml:"id"`
	Type           Type              `json:"type" yaml:"type"`
	Name           string            `json:"name" yaml:"name"`
	Server         string            `json:"server" yaml:"server"`
	Port           int               `json:"port" yaml:"port"`
	Password       string            `json:"password,omitempty" yaml:"password,omitempty"`
	Username       string            `json:"username,omitempty" yaml:"username,omitempty"`
	Method         string            `json:"method,omitempty" yaml:"method,omitempty"`
	UUID           string            `json:"uuid,omitempty" yaml:"uuid,omitempty"`
	AID            int               `json:"aid,omitempty" yaml:"aid,omitempty"`
	Security       string            `json:"security,omitempty" yaml:"security,omitempty"`
	Network        Network           `json:"network,omitempty" yaml:"network,omitempty"`
	TLS            TLS               `json:"tls,omitempty" yaml:"tls,omitempty"`
	SNI            string            `json:"sni,omitempty" yaml:"sni,omitempty"`
	Host           string            `json:"host,omitempty" yaml:"host,omitempty"`
	Path           string            `json:"path,omitempty" yaml:"path,omitempty"`
	Headers        map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Plugin         string            `json:"plugin,omitempty" yaml:"plugin,omitempty"`
	PluginOpts     string            `json:"plugin-opts,omitempty" yaml:"plugin-opts,omitempty"`
	UDP            bool              `json:"udp" yaml:"udp"`
	SkipCertVerify bool              `json:"skip-cert-verify,omitempty" yaml:"skip-cert-verify,omitempty"`
	Alpn           []string          `json:"alpn,omitempty" yaml:"alpn,omitempty"`
	Mux            bool              `json:"mux,omitempty" yaml:"mux,omitempty"`
	Congestion     string            `json:"congestion,omitempty" yaml:"congestion,omitempty"`
	UpMbps         int               `json:"up-mbps,omitempty" yaml:"up-mbps,omitempty"`
	DownMbps       int               `json:"down-mbps,omitempty" yaml:"down-mbps,omitempty"`
	Obfs           string            `json:"obfs,omitempty" yaml:"obfs,omitempty"`
	ObfsParam      string            `json:"obfs-param,omitempty" yaml:"obfs-param,omitempty"`
	Protocol       string            `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	ProtocolParam  string            `json:"protocol-param,omitempty" yaml:"protocol-param,omitempty"`
	// VLESS/Reality extras
	Flow              string    `json:"flow,omitempty" yaml:"flow,omitempty"`
	ClientFingerprint string    `json:"client-fingerprint,omitempty" yaml:"client-fingerprint,omitempty"`
	RealityPublicKey  string    `json:"reality-public-key,omitempty" yaml:"reality-public-key,omitempty"`
	RealityShortID    string    `json:"reality-short-id,omitempty" yaml:"reality-short-id,omitempty"`
	RealitySpiderX    string    `json:"reality-spider-x,omitempty" yaml:"reality-spider-x,omitempty"`
	CreatedAt         time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" yaml:"updated_at"`
}

type Params map[string]interface{}

type Statistics struct {
	Download int64 `json:"download,omitempty" yaml:"download,omitempty"`
	Upload   int64 `json:"upload,omitempty" yaml:"upload,omitempty"`
	Latency  int64 `json:"latency,omitempty" yaml:"latency,omitempty"`
}
