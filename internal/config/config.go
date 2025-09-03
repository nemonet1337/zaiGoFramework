package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds application configuration
// アプリケーション設定を保持
type Config struct {
	Database  DatabaseConfig  `yaml:"database"`
	API       APIConfig       `yaml:"api"`
	Inventory InventoryConfig `yaml:"inventory"`
	Logging   LoggingConfig   `yaml:"logging"`
}

// DatabaseConfig holds database configuration
// データベース設定を保持
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// APIConfig holds API server configuration
// APIサーバー設定を保持
type APIConfig struct {
	Port           int           `yaml:"port"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	EnableCORS     bool          `yaml:"enable_cors"`
	EnableMetrics  bool          `yaml:"enable_metrics"`
}

// InventoryConfig holds inventory-specific configuration
// 在庫固有の設定を保持
type InventoryConfig struct {
	AllowNegativeStock   bool  `yaml:"allow_negative_stock"`
	DefaultLocation      string `yaml:"default_location"`
	AuditEnabled         bool  `yaml:"audit_enabled"`
	LowStockThreshold    int64 `yaml:"low_stock_threshold"`
	AlertTimeoutHours    int   `yaml:"alert_timeout_hours"`
}

// LoggingConfig holds logging configuration
// ログ設定を保持
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"` // json, console
	Output string `yaml:"output"` // stdout, file
}

// Load loads configuration from environment variables
// 環境変数から設定を読み込み
func Load() (*Config, error) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "inventory"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "inventory_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		API: APIConfig{
			Port:          getEnvAsInt("API_PORT", 8080),
			ReadTimeout:   getEnvAsDuration("API_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:  getEnvAsDuration("API_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:   getEnvAsDuration("API_IDLE_TIMEOUT", 60*time.Second),
			EnableCORS:    getEnvAsBool("API_ENABLE_CORS", true),
			EnableMetrics: getEnvAsBool("API_ENABLE_METRICS", true),
		},
		Inventory: InventoryConfig{
			AllowNegativeStock: getEnvAsBool("INVENTORY_ALLOW_NEGATIVE_STOCK", false),
			DefaultLocation:    getEnv("INVENTORY_DEFAULT_LOCATION", "DEFAULT"),
			AuditEnabled:       getEnvAsBool("INVENTORY_AUDIT_ENABLED", true),
			LowStockThreshold:  getEnvAsInt64("INVENTORY_LOW_STOCK_THRESHOLD", 10),
			AlertTimeoutHours:  getEnvAsInt("INVENTORY_ALERT_TIMEOUT_HOURS", 24),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
	}

	// バリデーション
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("設定バリデーションに失敗しました: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
// 設定をバリデーション
func (c *Config) Validate() error {
	// データベース設定チェック
	if c.Database.Host == "" {
		return fmt.Errorf("データベースホストが指定されていません")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("無効なデータベースポート: %d", c.Database.Port)
	}
	if c.Database.User == "" {
		return fmt.Errorf("データベースユーザーが指定されていません")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("データベース名が指定されていません")
	}

	// API設定チェック
	if c.API.Port <= 0 || c.API.Port > 65535 {
		return fmt.Errorf("無効なAPIポート: %d", c.API.Port)
	}

	// 在庫設定チェック
	if c.Inventory.DefaultLocation == "" {
		return fmt.Errorf("デフォルトロケーションが指定されていません")
	}
	if c.Inventory.LowStockThreshold < 0 {
		return fmt.Errorf("低在庫閾値は0以上である必要があります")
	}

	// ログ設定チェック
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true, "fatal": true,
	}
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("無効なログレベル: %s", c.Logging.Level)
	}

	validLogFormats := map[string]bool{
		"json": true, "console": true,
	}
	if !validLogFormats[c.Logging.Format] {
		return fmt.Errorf("無効なログフォーマット: %s", c.Logging.Format)
	}

	return nil
}

// DSN generates PostgreSQL Data Source Name
// PostgreSQLデータソース名を生成
func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// ヘルパー関数

// getEnv gets environment variable with default value
// デフォルト値付きで環境変数を取得
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as integer with default value
// デフォルト値付きで環境変数を整数として取得
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsInt64 gets environment variable as int64 with default value
// デフォルト値付きで環境変数をint64として取得
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if int64Value, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int64Value
		}
	}
	return defaultValue
}

// getEnvAsBool gets environment variable as boolean with default value
// デフォルト値付きで環境変数をbooleanとして取得
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsDuration gets environment variable as duration with default value
// デフォルト値付きで環境変数をdurationとして取得
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
