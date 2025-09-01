package config

import (

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Log    LogConfig    `yaml:"log"`
	Redis  RedisConfig  `yaml:"redis"`
	Cache  CacheConfig  `yaml:"cache"`
	Security SecurityConfig `yaml:"security"`
	Parser ParserConfig `yaml:"parser"`
	Generator GeneratorConfig `yaml:"generator"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
	Mode string `yaml:"mode"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

type RedisConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
}

type CacheConfig struct {
	TTL     int `yaml:"ttl"`
	MaxSize int `yaml:"max_size"`
}

type SecurityConfig struct {
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	CORS      CORSConfig      `yaml:"cors"`
}

type RateLimitConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Requests  int    `yaml:"requests"`
	Window    string `yaml:"window"`
}

type CORSConfig struct {
	Enabled bool     `yaml:"enabled"`
	Origins []string `yaml:"origins"`
}

type ParserConfig struct {
	Timeout int `yaml:"timeout"`
	MaxSize int `yaml:"max_size"`
}

type GeneratorConfig struct {
	TemplatesDir string `yaml:"templates_dir"`
	RulesDir     string `yaml:"rules_dir"`
}

// Load loads configuration from file and environment
func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.mode", "release")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("redis.enabled", false)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.database", 0)
	viper.SetDefault("cache.ttl", 300)
	viper.SetDefault("cache.max_size", 1000)
	viper.SetDefault("security.rate_limit.enabled", true)
	viper.SetDefault("security.rate_limit.requests", 100)
	viper.SetDefault("security.rate_limit.window", "1m")
	viper.SetDefault("security.cors.enabled", true)
	viper.SetDefault("security.cors.origins", []string{"*"})
	viper.SetDefault("parser.timeout", 30)
	viper.SetDefault("parser.max_size", 10485760)
	viper.SetDefault("generator.templates_dir", "./base/base")
	viper.SetDefault("generator.rules_dir", "./base/rules")

	viper.AutomaticEnv()

	var cfg Config
	if err := viper.ReadInConfig(); err != nil {
		// Use defaults if config file not found
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}

	return &cfg
}