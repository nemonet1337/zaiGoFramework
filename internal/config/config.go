package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Config システム全体の設定構造体
type Config struct {
	Database  DatabaseConfig  `yaml:"database"`
	API       APIConfig       `yaml:"api"`
	Inventory InventoryConfig `yaml:"inventory"`
	Log       LogConfig       `yaml:"log"`
}

// DatabaseConfig データベース接続設定
type DatabaseConfig struct {
	Host     string `yaml:"host" env:"DB_HOST"`
	Port     int    `yaml:"port" env:"DB_PORT"`
	User     string `yaml:"user" env:"DB_USER"`
	Password string `yaml:"password" env:"DB_PASSWORD"`
	DBName   string `yaml:"dbname" env:"DB_NAME"`
}

// APIConfig API サーバー設定
type APIConfig struct {
	Port            int           `yaml:"port" env:"API_PORT"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	EnableCORS      bool          `yaml:"enable_cors"`
	EnableAuth      bool          `yaml:"enable_auth"`
}

// InventoryConfig 在庫管理設定
type InventoryConfig struct {
	AllowNegativeStock  bool   `yaml:"allow_negative_stock"`
	DefaultLocation     string `yaml:"default_location"`
	AuditEnabled        bool   `yaml:"audit_enabled"`
	LowStockThreshold   int64  `yaml:"low_stock_threshold"`
	AlertTimeoutHours   int    `yaml:"alert_timeout_hours"`
}

// LogConfig ログ設定
type LogConfig struct {
	Level      string `yaml:"level" env:"LOG_LEVEL"`
	Format     string `yaml:"format"`
	OutputPath string `yaml:"output_path"`
}

// Load 設定をYAMLファイルと環境変数から読み込み
func Load() (*Config, error) {
	config := &Config{
		// デフォルト値設定
		Database: DatabaseConfig{
			Host:   "localhost",
			Port:   5432,
			User:   "postgres",
			DBName: "inventory",
		},
		API: APIConfig{
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
			EnableCORS:   true,
			EnableAuth:   false,
		},
		Inventory: InventoryConfig{
			AllowNegativeStock: false,
			DefaultLocation:    "DEFAULT",
			AuditEnabled:       true,
			LowStockThreshold:  10,
			AlertTimeoutHours:  24,
		},
		Log: LogConfig{
			Level:      "info",
			Format:     "json",
			OutputPath: "stdout",
		},
	}

	// YAML設定ファイル読み込み
	if err := loadFromYAML(config); err != nil {
		return nil, fmt.Errorf("YAML設定読み込みエラー: %w", err)
	}

	// 環境変数でオーバーライド
	if err := loadFromEnv(config); err != nil {
		return nil, fmt.Errorf("環境変数読み込みエラー: %w", err)
	}

	// バリデーション
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("設定バリデーションエラー: %w", err)
	}

	return config, nil
}

// loadFromYAML YAMLファイルから設定を読み込み
func loadFromYAML(config *Config) error {
	configPaths := []string{
		"config/app.yaml",
		"config.yaml",
		"app.yaml",
	}

	var yamlFile []byte
	var err error
	
	for _, path := range configPaths {
		if yamlFile, err = ioutil.ReadFile(path); err == nil {
			break
		}
	}

	if err != nil {
		// YAML設定ファイルが見つからない場合はスキップ（デフォルト値を使用）
		return nil
	}

	return yaml.Unmarshal(yamlFile, config)
}

// loadFromEnv 環境変数から設定をオーバーライド
func loadFromEnv(config *Config) error {
	return loadEnvToStruct(config)
}

// loadEnvToStruct 構造体のenvタグに基づいて環境変数を読み込み
func loadEnvToStruct(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("引数はstructのpointerである必要があります")
	}

	rv = rv.Elem()
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		// 埋め込み構造体の処理
		if field.Kind() == reflect.Struct && fieldType.Anonymous == false {
			if err := loadEnvToStruct(field.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			continue
		}

		envValue := os.Getenv(envTag)
		if envValue == "" {
			continue
		}

		if err := setFieldValue(field, envValue); err != nil {
			return fmt.Errorf("フィールド %s の設定に失敗: %w", fieldType.Name, err)
		}
	}

	return nil
}

// setFieldValue フィールドに環境変数の値を設定
func setFieldValue(field reflect.Value, value string) error {
	if !field.CanSet() {
		return fmt.Errorf("フィールドに書き込みできません")
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(duration))
		} else {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(intVal)
		}
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	default:
		return fmt.Errorf("サポートされていない型: %s", field.Kind())
	}

	return nil
}

// validate 設定をバリデーション
func (c *Config) validate() error {
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
	if !validLogLevels[c.Log.Level] {
		return fmt.Errorf("無効なログレベル: %s", c.Log.Level)
	}

	validLogFormats := map[string]bool{
		"json": true, "console": true,
	}
	if !validLogFormats[c.Log.Format] {
		return fmt.Errorf("無効なログフォーマット: %s", c.Log.Format)
	}

	return nil
}

// DSN generates PostgreSQL Data Source Name
// PostgreSQLデータソース名を生成
func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
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
