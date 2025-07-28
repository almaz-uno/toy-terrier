package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Telegram     TelegramConfig     `mapstructure:"telegram"`
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Scraper      ScraperConfig      `mapstructure:"scraper"`
	Facebook     FacebookConfig     `mapstructure:"facebook"`
	Notification NotificationConfig `mapstructure:"notification"`
	Log          LogConfig          `mapstructure:"log"`
	Metrics      MetricsConfig      `mapstructure:"metrics"`
	Security     SecurityConfig     `mapstructure:"security"`
}

type TelegramConfig struct {
	BotToken string        `mapstructure:"bot_token"`
	Debug    bool          `mapstructure:"debug"`
	Polling  PollingConfig `mapstructure:"polling"`
	AdminIDs []int64       `mapstructure:"admin_ids"`
}

type PollingConfig struct {
	Timeout        int      `mapstructure:"timeout"`         // Polling timeout in seconds
	Limit          int      `mapstructure:"limit"`           // Number of updates to fetch
	AllowedUpdates []string `mapstructure:"allowed_updates"` // Types of updates to receive
}

type NotificationConfig struct {
	BatchSize          int           `mapstructure:"batch_size"`
	Interval           time.Duration `mapstructure:"interval"`
	MaxRetries         int           `mapstructure:"max_retries"`
	RetryDelay         time.Duration `mapstructure:"retry_delay"`
	QueueCheckInterval time.Duration `mapstructure:"queue_check_interval"`
}

type SecurityConfig struct {
	AdminUsername string        `mapstructure:"admin_username"`
	AdminPassword string        `mapstructure:"admin_password"`
	JWTSecret     string        `mapstructure:"jwt_secret"`
	RateLimit     int           `mapstructure:"rate_limit"`
	RateWindow    time.Duration `mapstructure:"rate_window"`
}

type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	Host            string        `mapstructure:"host"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

// ScraperConfig holds scraper configuration
type ScraperConfig struct {
	ScrapeInterval time.Duration `mapstructure:"scrape_interval"`
	TimeoutSeconds int           `mapstructure:"timeout_seconds"`
	RequestDelay   time.Duration `mapstructure:"request_delay"`
	UserAgent      string        `mapstructure:"user_agent"`
}

type FacebookConfig struct {
	AccessToken string `mapstructure:"access_token"`
	AppID       string `mapstructure:"app_id"`
	AppSecret   string `mapstructure:"app_secret"`
	PageID      string `mapstructure:"page_id"`
	APIVersion  string `mapstructure:"api_version"`
}

type LogConfig struct {
	Level         string `mapstructure:"level"`
	Format        string `mapstructure:"format"`
	Output        string `mapstructure:"output"`
	ConsoleWriter bool   `mapstructure:"console_writer"`
	TimeFormat    string `mapstructure:"time_format"`
	Caller        bool   `mapstructure:"caller"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
	Port    int    `mapstructure:"port"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	if configPath != "" {
		// Use specific config file if provided
		viper.SetConfigFile(configPath)
	} else {
		// Use default config search paths
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./local")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(".")
	}

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Override with environment variables
	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setDefaults() {
	// Telegram defaults
	viper.SetDefault("telegram.debug", false)
	viper.SetDefault("telegram.polling.timeout", 60)
	viper.SetDefault("telegram.polling.limit", 100)
	viper.SetDefault("telegram.polling.allowed_updates", []string{"message", "callback_query"})

	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "10s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.shutdown_timeout", "30s")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "300s")
	viper.SetDefault("database.conn_max_idle_time", "120s")

	// Scraper defaults
	viper.SetDefault("scraper.user_agent", "Mozilla/5.0 (compatible; ToyTerrier/1.0)")
	viper.SetDefault("scraper.timeout", "30s")
	viper.SetDefault("scraper.retry_attempts", 3)
	viper.SetDefault("scraper.retry_delay", "5s")
	viper.SetDefault("scraper.scrape_interval", "1m")
	viper.SetDefault("scraper.timeout_seconds", 30)
	viper.SetDefault("scraper.request_delay", "2s")

	// Facebook defaults
	viper.SetDefault("facebook.api_version", "v18.0")

	// Notification defaults
	viper.SetDefault("notification.batch_size", 100)
	viper.SetDefault("notification.interval", "5m")
	viper.SetDefault("notification.max_retries", 3)
	viper.SetDefault("notification.retry_delay", "30s")
	viper.SetDefault("notification.queue_check_interval", "10s")

	// Log defaults
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.console_writer", false)
	viper.SetDefault("log.caller", false)

	// Metrics defaults
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.path", "/metrics")
	viper.SetDefault("metrics.port", 9090)

	// Security defaults
	viper.SetDefault("security.rate_limit", 100)
	viper.SetDefault("security.rate_window", "60s")
}
