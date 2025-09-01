package parser

import (
	"context"
	"testing"

	"github.com/subconverter/subconverter-go/internal/domain/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSParser_Parse(t *testing.T) {
	parser := NewSSParser()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    string
		expected *proxy.Proxy
		wantErr  bool
	}{
		{
			name:  "valid legacy SS",
			input: "ss://YWVzLTI1Ni1nY206dGVzdA==@127.0.0.1:8388#Test",
			expected: &proxy.Proxy{
				Type:     proxy.Shadowsocks,
				Server:   "127.0.0.1",
				Port:     8388,
				Password: "test",
				Method:   "aes-256-gcm",
				Name:     "Test",
			},
			wantErr: false,
		},
		{
			name:  "valid SIP002 SS",
			input: "ss://YWVzLTI1Ni1nY206dGVzdA==@127.0.0.1:8388/?plugin=obfs-local%3Bobfs%3Dhttp#Test",
			expected: &proxy.Proxy{
				Type:     proxy.Shadowsocks,
				Server:   "127.0.0.1",
				Port:     8388,
				Password: "test",
				Method:   "aes-256-gcm",
				Plugin:   "obfs-local",
				PluginOpts: "obfs=http",
				Name:     "Test",
			},
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "invalid://test",
			expected: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(ctx, tt.input)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, result, 1)
			
			proxy := result[0]
			assert.Equal(t, tt.expected.Type, proxy.Type)
			assert.Equal(t, tt.expected.Server, proxy.Server)
			assert.Equal(t, tt.expected.Port, proxy.Port)
			assert.Equal(t, tt.expected.Password, proxy.Password)
			assert.Equal(t, tt.expected.Method, proxy.Method)
			assert.Equal(t, tt.expected.Name, proxy.Name)
			
			if tt.expected.Plugin != "" {
				assert.Equal(t, tt.expected.Plugin, proxy.Plugin)
				assert.Equal(t, tt.expected.PluginOpts, proxy.PluginOpts)
			}
		})
	}
}

func TestSSParser_Supports(t *testing.T) {
	parser := NewSSParser()
	
	assert.True(t, parser.Supports("ss://test"))
	assert.False(t, parser.Supports("vmess://test"))
	assert.False(t, parser.Supports("invalid://test"))
}

func BenchmarkSSParser_Parse(b *testing.B) {
	parser := NewSSParser()
	ctx := context.Background()
	input := "ss://YWVzLTI1Ni1nY206dGVzdA==@127.0.0.1:8388#Test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(ctx, input)
	}
}