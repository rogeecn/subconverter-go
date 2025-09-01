package converter

import (
	"context"
	"testing"

	"github.com/subconverter/subconverter-go/internal/infra/config"
	"github.com/subconverter/subconverter-go/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestService_Convert(t *testing.T) {
	cfg := &config.Config{
		Cache: config.CacheConfig{
			TTL: 300,
		},
	}
	log := logger.New(logger.Config{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	})

	service := NewService(cfg, log)
	service.RegisterGenerators()

	t.Run("invalid target", func(t *testing.T) {
		_ = &ConvertRequest{
			Target: "invalid",
			URLs:   []string{"https://example.com/subscription"},
		}

		_, err := service.Convert(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("empty URLs", func(t *testing.T) {
		_ = &ConvertRequest{
			Target: "clash",
			URLs:   []string{},
		}

		_, err := service.Convert(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("valid request", func(t *testing.T) {
		_ = &ConvertRequest{
			Target: "clash",
			URLs:   []string{"https://example.com/subscription"},
		}

		// Mock HTTP client would be needed for real tests
		// This is a placeholder test
		assert.NotNil(t, service)
	})
}

func TestService_Validate(t *testing.T) {
	cfg := &config.Config{}
	log := logger.New(logger.Config{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	})

	service := NewService(cfg, log)
	service.RegisterGenerators()

	t.Run("valid URL format detection", func(t *testing.T) {
		content := "ss://YWVzLTI1Ni1nY206dGVzdA==@127.0.0.1:8388#Test"
		format := service.detectFormat(content)
		assert.Equal(t, "shadowsocks", format)
	})

	t.Run("vmess format detection", func(t *testing.T) {
		content := "vmess://eyJhZGQiOiIxMjcuMC4wLjEiLCJwb3J0IjoiODAiL..."
		format := service.detectFormat(content)
		assert.Equal(t, "vmess", format)
	})
}

func TestService_SupportedFormats(t *testing.T) {
	cfg := &config.Config{}
	log := logger.New(logger.Config{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	})

	service := NewService(cfg, log)
	service.RegisterGenerators()

	formats := service.SupportedFormats()
	assert.Contains(t, formats, "clash")
	assert.Contains(t, formats, "surge")
	assert.Contains(t, formats, "quantumult")
	assert.Contains(t, formats, "loon")
	assert.Contains(t, formats, "v2ray")
	assert.Contains(t, formats, "surfboard")
}

func TestService_Health(t *testing.T) {
	cfg := &config.Config{}
	log := logger.New(logger.Config{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	})

	service := NewService(cfg, log)
	err := service.Health(context.Background())
	assert.NoError(t, err)
}

func TestService_GetInfo(t *testing.T) {
	cfg := &config.Config{}
	log := logger.New(logger.Config{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	})

	service := NewService(cfg, log)
	service.RegisterGenerators()

	info, err := service.GetInfo(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", info.Version)
	assert.NotEmpty(t, info.SupportedTypes)
	assert.NotEmpty(t, info.Features)
}