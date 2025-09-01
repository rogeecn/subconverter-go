package converter

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/samber/lo"
	"github.com/subconverter/subconverter-go/internal/app/generator"
	"github.com/subconverter/subconverter-go/internal/app/parser"
	"github.com/subconverter/subconverter-go/internal/app/template"
	"github.com/subconverter/subconverter-go/internal/domain/proxy"
	"github.com/subconverter/subconverter-go/internal/infra/cache"
	"github.com/subconverter/subconverter-go/internal/infra/config"
	"github.com/subconverter/subconverter-go/internal/infra/http"
	"github.com/subconverter/subconverter-go/internal/pkg/errors"
	"github.com/subconverter/subconverter-go/internal/pkg/logger"
)

// Service provides the core conversion functionality
type Service struct {
	parserManager    *parser.Manager
	generatorManager *generator.Manager
	templateManager  *template.Manager
	cache            cache.Cache
	config           *config.Config
	httpClient       *http.Client
	logger           *logger.Logger
}

// NewService creates a new conversion service
func NewService(cfg *config.Config, log *logger.Logger) *Service {
	templateManager := template.NewManager(cfg.Generator.TemplatesDir, cfg.Generator.RulesDir, *log)
	
	return &Service{
		parserManager:    parser.NewManager(),
		generatorManager: generator.NewManager(),
		templateManager:  templateManager,
		cache:            cache.NewMemoryCache(),
		config:           cfg,
		httpClient:       http.NewClient(),
		logger:           log,
	}
}

// Convert converts subscription URLs to target format
func (s *Service) Convert(ctx context.Context, req *ConvertRequest) (*ConvertResponse, error) {
	start := time.Now()
	defer func() {
		s.logger.WithFields(map[string]interface{}{
			"target":    req.Target,
			"urls":      len(req.URLs),
			"duration":  time.Since(start),
		}).Info("Conversion completed")
	}()

	// Validate request
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	// Check cache
	cacheKey := s.generateCacheKey(req)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var resp ConvertResponse
		if err := json.Unmarshal(cached, &resp); err == nil {
			return &resp, nil
		}
	}

	// Fetch subscriptions
	allProxies, err := s.fetchSubscriptions(ctx, req.URLs)
	if err != nil {
		return nil, err
	}

	// Apply filters
	filteredProxies := s.applyFilters(allProxies, req.Options)

	// Generate configuration
	config, err := s.generatorManager.Generate(ctx, req.Target, filteredProxies, nil, generator.GenerateOptions{
		ProxyGroups:   s.buildProxyGroups(req.Options),
		Rules:         req.Options.Rules,
		SortProxies:   req.Options.Sort,
		UDPEnabled:    req.Options.UDP,
		RenameRules:   req.Options.RenameRules,
		EmojiRules:    req.Options.EmojiRules,
		BaseTemplate:  req.Options.BaseTemplate,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate configuration")
	}

	// Build response
	resp := &ConvertResponse{
		Config:    config,
		Format:    req.Target,
		Proxies:   filteredProxies,
		Generated: time.Now().Format(time.RFC3339),
	}

	// Cache the response
	if cacheData, err := json.Marshal(resp); err == nil {
		s.cache.Set(ctx, cacheKey, cacheData, time.Duration(s.config.Cache.TTL)*time.Second)
	}

	return resp, nil
}

// Validate validates a subscription URL
func (s *Service) Validate(ctx context.Context, req *ValidateRequest) (*ValidateResponse, error) {
	content, err := s.httpClient.Get(ctx, req.URL)
	if err != nil {
		return &ValidateResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	proxies, err := s.parserManager.Parse(ctx, string(content))
	if err != nil {
		return &ValidateResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	format := s.detectFormat(string(content))

	return &ValidateResponse{
		Valid:   true,
		Format:  format,
		Proxies: len(proxies),
	}, nil
}

// GetInfo returns service information
func (s *Service) GetInfo(ctx context.Context) (*InfoResponse, error) {
	formats := s.SupportedFormats()
	return &InfoResponse{
		Version:        "1.0.0",
		SupportedTypes: formats,
		Features: []string{
			"High-performance conversion",
			"Multiple protocol support",
			"Cloud-native architecture",
			"Caching support",
			"Rate limiting",
			"Health checks",
		},
	}, nil
}

func (s *Service) detectFormat(content string) string {
	// Simple format detection based on content patterns
	if strings.Contains(content, "ss://") || strings.Contains(content, "ssr://") {
		return "shadowsocks"
	}
	if strings.Contains(content, "vmess://") {
		return "vmess"
	}
	if strings.Contains(content, "trojan://") {
		return "trojan"
	}
	if strings.Contains(content, "vless://") {
		return "vless"
	}
	if strings.Contains(content, "hysteria://") {
		return "hysteria"
	}
	if strings.Contains(content, "hysteria2://") {
		return "hysteria2"
	}
	if strings.Contains(content, "snell://") {
		return "snell"
	}
	return "unknown"
}

// Health checks the service health
func (s *Service) Health(ctx context.Context) error {
	// Check cache health
	if err := s.cache.Health(ctx); err != nil {
		return errors.Wrap(err, "cache health check failed")
	}

	// Check HTTP client
	if err := s.httpClient.Health(ctx); err != nil {
		return errors.Wrap(err, "http client health check failed")
	}

	return nil
}

func (s *Service) validateRequest(req *ConvertRequest) error {
	if req.Target == "" {
		return errors.BadRequest("INVALID_TARGET", "target format is required")
	}

	if len(req.URLs) == 0 {
		return errors.BadRequest("INVALID_URLS", "at least one subscription URL is required")
	}

	// Check if target format is supported
	if _, exists := s.generatorManager.Get(req.Target); !exists {
		return errors.BadRequest("UNSUPPORTED_TARGET", fmt.Sprintf("target format '%s' is not supported", req.Target))
	}

	return nil
}

func (s *Service) fetchSubscriptions(ctx context.Context, urls []string) ([]*proxy.Proxy, error) {
	type result struct {
		proxies []*proxy.Proxy
		err     error
	}

	results := make(chan result, len(urls))
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()

			content, err := s.httpClient.Get(ctx, u)
			if err != nil {
				results <- result{err: errors.Wrap(err, fmt.Sprintf("failed to fetch URL: %s", u))}
				return
			}

			proxies, err := s.parserManager.Parse(ctx, string(content))
			if err != nil {
				results <- result{err: errors.Wrap(err, fmt.Sprintf("failed to parse subscription: %s", u))}
				return
			}

			results <- result{proxies: proxies}
		}(url)
	}

	wg.Wait()
	close(results)

	// Collect results
	var allProxies []*proxy.Proxy
	for r := range results {
		if r.err != nil {
			s.logger.WithError(r.err).Warn("Failed to process subscription")
			continue
		}
		allProxies = append(allProxies, r.proxies...)
	}

	if len(allProxies) == 0 {
		return nil, errors.BadRequest("NO_PROXIES", "no valid proxies found in subscriptions")
	}

	return allProxies, nil
}

func (s *Service) applyFilters(proxies []*proxy.Proxy, options Options) []*proxy.Proxy {
	result := proxies

	// Apply include filters
	if len(options.IncludeRemarks) > 0 {
		result = lo.Filter(result, func(p *proxy.Proxy, _ int) bool {
			return lo.SomeBy(options.IncludeRemarks, func(pattern string) bool {
				return strings.Contains(p.Name, pattern)
			})
		})
	}

	// Apply exclude filters
	if len(options.ExcludeRemarks) > 0 {
		result = lo.Filter(result, func(p *proxy.Proxy, _ int) bool {
			return !lo.SomeBy(options.ExcludeRemarks, func(pattern string) bool {
				return strings.Contains(p.Name, pattern)
			})
		})
	}

	// Apply rename rules
	if len(options.RenameRules) > 0 {
		for _, p := range result {
			for _, rule := range options.RenameRules {
				p.Name = strings.ReplaceAll(p.Name, rule.Match, rule.Replace)
			}
		}
	}

	// Apply emoji rules
	if len(options.EmojiRules) > 0 {
		for _, p := range result {
			for _, rule := range options.EmojiRules {
				if strings.Contains(p.Name, rule.Match) {
					p.Name = rule.Emoji + " " + p.Name
				}
			}
		}
	}

	// Sort proxies
	if options.Sort {
		sort.Slice(result, func(i, j int) bool {
			return result[i].Name < result[j].Name
		})
	}

	// Remove duplicates
	seen := make(map[string]bool)
	unique := make([]*proxy.Proxy, 0, len(result))
	for _, p := range result {
		key := fmt.Sprintf("%s:%d:%s", p.Server, p.Port, p.Type)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, p)
		}
	}

	return unique
}

func (s *Service) buildProxyGroups(options Options) []generator.ProxyGroup {
	if len(options.ProxyGroups) > 0 {
		return options.ProxyGroups
	}

	// Default proxy groups
	return []generator.ProxyGroup{
		{
			Name:    "🚀 节点选择",
			Type:    "select",
			Proxies: []string{"♻️ 自动选择", "🔯 故障转移", "DIRECT"},
		},
		{
			Name:     "♻️ 自动选择",
			Type:     "url-test",
			Proxies:  []string{},
			URL:      "http://www.gstatic.com/generate_204",
			Interval: 300,
		},
		{
			Name:     "🔯 故障转移",
			Type:     "fallback",
			Proxies:  []string{},
			URL:      "http://www.gstatic.com/generate_204",
			Interval: 300,
		},
	}
}

func (s *Service) generateCacheKey(req *ConvertRequest) string {
	key := fmt.Sprintf("convert:%s:%s", req.Target, strings.Join(req.URLs, ","))
	return key
}

// RegisterGenerators registers all available generators
func (s *Service) RegisterGenerators() {
	s.generatorManager.Register("clash", generator.NewClashGenerator(s.templateManager))
	s.generatorManager.Register("surge", generator.NewSurgeGenerator())
	s.generatorManager.Register("quantumult", generator.NewQuantumultGenerator())
	s.generatorManager.Register("loon", generator.NewLoonGenerator())
	s.generatorManager.Register("v2ray", generator.NewV2RayGenerator())
	s.generatorManager.Register("surfboard", generator.NewSurfboardGenerator())
}

// SupportedFormats returns all supported formats
func (s *Service) GeneratorManager() *generator.Manager {
	return s.generatorManager
}

func (s *Service) HTTPClient() *http.Client {
	return s.httpClient
}

func (s *Service) ParserManager() *parser.Manager {
	return s.parserManager
}

func (s *Service) DetectFormat(content string) string {
	return s.detectFormat(content)
}

func (s *Service) SupportedFormats() []string {
	return s.generatorManager.SupportedFormats()
}