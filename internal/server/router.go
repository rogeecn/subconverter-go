package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rogeecn/subconverter-go/internal/app/converter"
	"github.com/rogeecn/subconverter-go/internal/app/generator"
	"github.com/rogeecn/subconverter-go/internal/infra/config"
	"github.com/rogeecn/subconverter-go/internal/pkg/errors"
)

// Router manages HTTP routes
type Router struct {
	app     *fiber.App
	service *converter.Service
	config  *config.Config
}

// NewRouter creates a new router
func NewRouter(service *converter.Service, cfg *config.Config) *Router {
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "SubConverter-Go",
		AppName:       "SubConverter Go",
		ReadTimeout:   30 * time.Second,
		WriteTimeout:  30 * time.Second,
		IdleTimeout:   60 * time.Second,
	})

	return &Router{
		app:     app,
		service: service,
		config:  cfg,
	}
}

// SetupRoutes configures all routes
func (r *Router) SetupRoutes() {
	// Middleware
	r.app.Use(recover.New())
	r.app.Use(logger.New(logger.Config{
		Format: "${time} ${method} ${path} - ${status} ${latency}\n",
	}))

	if r.config.Security.CORS.Enabled {
		r.app.Use(cors.New(cors.Config{
			AllowOrigins: strings.Join(r.config.Security.CORS.Origins, ","),
			AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		}))
	}

	if r.config.Security.RateLimit.Enabled {
		r.app.Use(limiter.New(limiter.Config{
			Max:        r.config.Security.RateLimit.Requests,
			Expiration: parseDuration(r.config.Security.RateLimit.Window),
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.IP()
			},
		}))
	}

	r.app.Get("/health", r.handleHealth)
	r.app.Get("/", r.handleConvert)
}

// handleConvertGet adds a GET-based conversion for direct subscription merge URLs
// Example:
// /?target=clash&url=...&url=...&sort=1&udp=1&include=HK&exclude=TEST
func (r *Router) handleConvert(c *fiber.Ctx) error {
	target := c.Query("target")
	if target == "" {
		target = "clash" // default to Clash YAML for convenience
	}

	// Collect URLs from repeated url params and from comma-separated urls
	q := c.Context().QueryArgs()
	urls := make([]string, 0)
	// repeated url params
	if vals := q.PeekMulti("url"); len(vals) > 0 {
		for _, v := range vals {
			if len(v) > 0 {
				urls = append(urls, string(v))
			}
		}
	}
	// single urls comma-separated
	if s := c.Query("urls"); s != "" {
		for _, u := range strings.Split(s, ",") {
			u = strings.TrimSpace(u)
			if u != "" {
				urls = append(urls, u)
			}
		}
	}
	if len(urls) == 0 {
		// allow single url param alias
		if s := c.Query("url"); s != "" {
			urls = append(urls, s)
		}
	}
	if len(urls) == 0 {
		return r.errorResponse(c, errors.BadRequest("INVALID_URLS", "at least one url is required"))
	}

	// Build options from query
	opts := converter.Options{}

	// include/exclude remarks
	if vals := q.PeekMulti("include"); len(vals) > 0 {
		for _, v := range vals {
			opts.IncludeRemarks = append(opts.IncludeRemarks, string(v))
		}
	}
	if s := c.Query("include_remarks"); s != "" {
		opts.IncludeRemarks = append(opts.IncludeRemarks, strings.Split(s, ",")...)
	}
	if vals := q.PeekMulti("exclude"); len(vals) > 0 {
		for _, v := range vals {
			opts.ExcludeRemarks = append(opts.ExcludeRemarks, string(v))
		}
	}
	if s := c.Query("exclude_remarks"); s != "" {
		opts.ExcludeRemarks = append(opts.ExcludeRemarks, strings.Split(s, ",")...)
	}

	// sort / udp
	if b := c.Query("sort"); b != "" {
		opts.Sort = isTruthy(b)
	}
	if b := c.Query("udp"); b != "" {
		opts.UDP = isTruthy(b)
	}

	// rules
	if vals := q.PeekMulti("rule"); len(vals) > 0 {
		for _, v := range vals {
			opts.Rules = append(opts.Rules, string(v))
		}
	}
	if s := c.Query("rules"); s != "" {
		opts.Rules = append(opts.Rules, strings.Split(s, ",")...)
	}

	// rename rules: rename=old->new (repeatable)
	if vals := q.PeekMulti("rename"); len(vals) > 0 {
		for _, v := range vals {
			parts := strings.SplitN(string(v), "->", 2)
			if len(parts) == 2 {
				opts.RenameRules = append(opts.RenameRules, generator.RenameRule{Match: parts[0], Replace: parts[1]})
			}
		}
	}
	// emoji rules: emoji=match:😊 (repeatable)
	if vals := q.PeekMulti("emoji"); len(vals) > 0 {
		for _, v := range vals {
			parts := strings.SplitN(string(v), ":", 2)
			if len(parts) == 2 {
				opts.EmojiRules = append(opts.EmojiRules, generator.EmojiRule{Match: parts[0], Emoji: parts[1]})
			}
		}
	}

	// base template
	if s := c.Query("base_template"); s != "" {
		opts.BaseTemplate = s
	}
	if s := c.Query("base"); s != "" {
		opts.BaseTemplate = s
	}

	// Build request and convert
	req := converter.ConvertRequest{Target: target, URLs: urls, Options: opts}
	resp, err := r.service.Convert(c.Context(), &req)
	if err != nil {
		return r.errorResponse(c, err)
	}

	// Set content type and return
	generator, exists := r.service.GeneratorManager().Get(target)
	if !exists {
		return r.errorResponse(c, fmt.Errorf("unsupported format: %s", target))
	}
	c.Set("Content-Type", generator.ContentType())
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=config.%s", target))
	return c.SendString(resp.Config)
}

func isTruthy(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// handleHealth returns health status
func (r *Router) handleHealth(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)

	// Check service health
	if err := r.service.Health(ctx); err != nil {
		services["service"] = "unhealthy"
		services["error"] = err.Error()
	} else {
		services["service"] = "healthy"
	}

	return c.JSON(converter.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
		Services:  services,
	})
}

// errorResponse returns a standardized error response
func (r *Router) errorResponse(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*errors.Error); ok {
		return c.Status(appErr.Status).JSON(map[string]interface{}{
			"error":   appErr.Message,
			"code":    appErr.Code,
			"details": appErr.Details,
		})
	}

	return c.Status(500).JSON(map[string]interface{}{
		"error": err.Error(),
		"code":  "INTERNAL_ERROR",
	})
}

// App returns the fiber app
func (r *Router) App() *fiber.App {
	return r.app
}

// parseDuration parses duration string
func parseDuration(durationStr string) time.Duration {
	duration, _ := time.ParseDuration(durationStr)
	if duration == 0 {
		duration = time.Minute
	}
	return duration
}
