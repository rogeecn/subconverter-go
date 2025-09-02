package converter

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rogeecn/subconverter-go/internal/app/generator"
	"github.com/rogeecn/subconverter-go/internal/app/parser"
	"github.com/rogeecn/subconverter-go/internal/app/template"
	"github.com/rogeecn/subconverter-go/internal/domain/proxy"
	"github.com/rogeecn/subconverter-go/internal/infra/cache"
	"github.com/rogeecn/subconverter-go/internal/infra/config"
	"github.com/rogeecn/subconverter-go/internal/infra/http"
	"github.com/rogeecn/subconverter-go/internal/pkg/errors"
	"github.com/rogeecn/subconverter-go/internal/pkg/logger"
	"github.com/samber/lo"
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
			"target":   req.Target,
			"urls":     len(req.URLs),
			"duration": time.Since(start),
		}).Info("Conversion completed")
	}()

	// Validate request
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	// Log request basics
	s.logger.WithFields(map[string]interface{}{
		"target":          req.Target,
		"req_urls":        len(req.URLs),
		"include_remarks": len(req.Options.IncludeRemarks),
		"exclude_remarks": len(req.Options.ExcludeRemarks),
		"rename_rules":    len(req.Options.RenameRules),
		"emoji_rules":     len(req.Options.EmojiRules),
		"sort":            req.Options.Sort,
		"udp":             req.Options.UDP,
		"base_template":   req.Options.BaseTemplate,
	}).Info("Conversion request received")

	// Merge request URLs with configured extra links and check cache
	mergedURLs := make([]string, 0, len(req.URLs)+len(s.config.Subscription.ExtraLinks))
	mergedURLs = append(mergedURLs, req.URLs...)
	if len(s.config.Subscription.ExtraLinks) > 0 {
		mergedURLs = append(mergedURLs, s.config.Subscription.ExtraLinks...)
	}

	mergedReq := *req
	mergedReq.URLs = mergedURLs

	s.logger.WithFields(map[string]interface{}{"req_urls": len(req.URLs), "extra_links": len(s.config.Subscription.ExtraLinks), "merged_urls": len(mergedURLs)}).
		Debug("URLs merged")

	cacheKey := s.generateCacheKey(&mergedReq)
	s.logger.WithField("cache_key", cacheKey).Debug("Cache lookup start")
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var resp ConvertResponse
		if err := json.Unmarshal(cached, &resp); err == nil {
			s.logger.WithFields(map[string]interface{}{
				"cache_key": cacheKey,
				"proxies":   len(resp.Proxies),
				"bytes":     len(resp.Config),
			}).Info("Cache hit")
			return &resp, nil
		}
	}
	s.logger.WithField("cache_key", cacheKey).Debug("Cache miss")

	// Fetch subscriptions
	if len(mergedURLs) == 0 {
		return nil, errors.BadRequest("INVALID_URLS", "no subscription URLs or extra links provided")
	}
	s.logger.WithField("count", len(mergedURLs)).Info("Fetching subscriptions")
	allProxies, err := s.fetchSubscriptions(ctx, mergedURLs)
	if err != nil {
		return nil, err
	}
	s.logger.WithField("proxies", len(allProxies)).Info("Fetched and parsed subscriptions")

	// Apply filters
	before := len(allProxies)
	filteredProxies := s.applyFilters(allProxies, req.Options)
	s.logger.WithFields(map[string]interface{}{
		"before": before,
		"after":  len(filteredProxies),
	}).Info("Filters applied")

	// Generate configuration
	genStart := time.Now()
	s.logger.WithFields(map[string]interface{}{"target": req.Target, "proxies": len(filteredProxies)}).
		Info("Generating configuration")
	config, err := s.generatorManager.Generate(ctx, req.Target, filteredProxies, nil, generator.GenerateOptions{
		ProxyGroups:  s.buildProxyGroups(req.Options),
		Rules:        req.Options.Rules,
		SortProxies:  req.Options.Sort,
		UDPEnabled:   req.Options.UDP,
		RenameRules:  req.Options.RenameRules,
		EmojiRules:   req.Options.EmojiRules,
		BaseTemplate: req.Options.BaseTemplate,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate configuration")
	}
	s.logger.WithFields(map[string]interface{}{"bytes": len(config), "duration": time.Since(genStart)}).
		Info("Configuration generated")

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
		s.logger.WithFields(map[string]interface{}{"cache_key": cacheKey, "ttl_sec": s.config.Cache.TTL}).
			Debug("Cached conversion result")
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
	// Check more specific schemes first to avoid substring conflicts (e.g., vmess contains 'ss')
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
	if strings.Contains(content, "ss://") || strings.Contains(content, "ssr://") {
		return "shadowsocks"
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

			var text string
			if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
				s.logger.WithField("url", u).Debug("Fetching URL")
				content, err := s.httpClient.Get(ctx, u)
				if err != nil {
					results <- result{err: errors.Wrap(err, fmt.Sprintf("failed to fetch URL: %s", u))}
					return
				}
				text = string(content)
				s.logger.WithFields(map[string]interface{}{"url": u, "bytes": len(content)}).Debug("Fetched URL")
			} else {
				// treat as direct node link content (e.g., ss://, trojan://)
				text = u
				s.logger.WithField("scheme", strings.SplitN(u, "://", 2)[0]).Debug("Parsing direct node link")
			}

			proxies, err := s.parserManager.Parse(ctx, text)
			if err != nil {
				results <- result{err: errors.Wrap(err, fmt.Sprintf("failed to parse subscription: %s", u))}
				return
			}

			s.logger.WithFields(map[string]interface{}{"source": u, "proxies": len(proxies)}).
				Debug("Parsed proxies from source")
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
