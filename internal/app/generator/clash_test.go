package generator

import (
	"context"
	"strings"
	"testing"

	"github.com/subconverter/subconverter-go/internal/domain/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClashGenerator_Generate(t *testing.T) {
	generator := NewClashGenerator(nil)
	ctx := context.Background()

	proxies := []*proxy.Proxy{
		{
			Type:     proxy.Shadowsocks,
			Server:   "127.0.0.1",
			Port:     8388,
			Password: "test",
			Method:   "aes-256-gcm",
			Name:     "Test-SS",
		},
		{
			Type:     proxy.VMess,
			Server:   "127.0.0.1",
			Port:     443,
			UUID:     "12345678-1234-1234-1234-123456789012",
			AlterID:  0,
			Security: "auto",
			Network:  "tcp",
			TLS:      true,
			Name:     "Test-VMess",
		},
		{
			Type:     proxy.Trojan,
			Server:   "127.0.0.1",
			Port:     443,
			Password: "test",
			SNI:      "example.com",
			TLS:      true,
			Name:     "Test-Trojan",
		},
	}

	proxyGroups := []ProxyGroup{
		{
			Name:    "🚀 节点选择",
			Type:    "select",
			Proxies: []string{"♻️ 自动选择", "🔯 故障转移", "DIRECT"},
		},
		{
			Name:     "♻️ 自动选择",
			Type:     "url-test",
			Proxies:  []string{"Test-SS", "Test-VMess", "Test-Trojan"},
			URL:      "http://www.gstatic.com/generate_204",
			Interval: 300,
		},
	}

	rules := []string{
		"DOMAIN-SUFFIX,google.com,🚀 节点选择",
		"DOMAIN-SUFFIX,github.com,🚀 节点选择",
		"GEOIP,CN,DIRECT",
		"MATCH,🚀 节点选择",
	}

	config, err := generator.Generate(ctx, proxies, proxyGroups, rules, GenerateOptions{
		SortProxies: true,
		UDPEnabled:  true,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, config)

	// Verify YAML structure
	assert.Contains(t, config, "port: 7890")
	assert.Contains(t, config, "socks-port: 7891")
	assert.Contains(t, config, "allow-lan: true")
	assert.Contains(t, config, "mode: Rule")
	assert.Contains(t, config, "log-level: info")

	// Verify proxies
	assert.Contains(t, config, "proxies:")
	assert.Contains(t, config, "- name: Test-SS")
	assert.Contains(t, config, "  type: ss")
	assert.Contains(t, config, "  server: 127.0.0.1")
	assert.Contains(t, config, "  port: 8388")
	assert.Contains(t, config, "  cipher: aes-256-gcm")
	assert.Contains(t, config, "  password: test")

	// Verify proxy groups
	assert.Contains(t, config, "proxy-groups:")
	assert.Contains(t, config, "- name: 🚀 节点选择")
	assert.Contains(t, config, "  type: select")
	assert.Contains(t, config, "- name: ♻️ 自动选择")
	assert.Contains(t, config, "  type: url-test")

	// Verify rules
	assert.Contains(t, config, "rules:")
	assert.Contains(t, config, "- DOMAIN-SUFFIX,google.com,🚀 节点选择")
	assert.Contains(t, config, "- GEOIP,CN,DIRECT")
}

func TestClashGenerator_ContentType(t *testing.T) {
	generator := NewClashGenerator(nil)
	assert.Equal(t, "application/x-yaml", generator.ContentType())
}

func TestClashGenerator_Name(t *testing.T) {
	generator := NewClashGenerator(nil)
	assert.Equal(t, "clash", generator.Name())
}

func TestClashGenerator_EmptyProxies(t *testing.T) {
	generator := NewClashGenerator(nil)
	ctx := context.Background()

	config, err := generator.Generate(ctx, []*proxy.Proxy{}, []ProxyGroup{}, []string{}, GenerateOptions{})
	require.NoError(t, err)
	assert.Contains(t, config, "proxies: []")
}

func BenchmarkClashGenerator_Generate(b *testing.B) {
	generator := NewClashGenerator(nil)
	ctx := context.Background()

	proxies := []*proxy.Proxy{
		{
			Type:     proxy.Shadowsocks,
			Server:   "127.0.0.1",
			Port:     8388,
			Password: "test",
			Method:   "aes-256-gcm",
			Name:     "Test-SS",
		},
	}

	proxyGroups := []ProxyGroup{
		{
			Name:    "🚀 节点选择",
			Type:    "select",
			Proxies: []string{"DIRECT"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.Generate(ctx, proxies, proxyGroups, []string{}, GenerateOptions{})
	}
}