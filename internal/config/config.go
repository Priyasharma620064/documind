package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Embedding   EmbeddingConfig   `mapstructure:"embedding"`
	Storage     StorageConfig     `mapstructure:"storage"`
	Repositories []RepoConfig    `mapstructure:"repositories"`
	Ingestion   IngestionConfig   `mapstructure:"ingestion"`
	Search      SearchConfig      `mapstructure:"search"`
	Evaluation  EvaluationConfig  `mapstructure:"evaluation"`
	Logging     LoggingConfig     `mapstructure:"logging"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	HTTPAddr     string        `mapstructure:"http_addr"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// EmbeddingConfig holds embedding model settings.
type EmbeddingConfig struct {
	Provider   string `mapstructure:"provider"`
	Model      string `mapstructure:"model"`
	Endpoint   string `mapstructure:"endpoint"`
	Dimensions int    `mapstructure:"dimensions"`
	BatchSize  int    `mapstructure:"batch_size"`
}

// StorageConfig holds storage paths.
type StorageConfig struct {
	DataDir      string `mapstructure:"data_dir"`
	SQLitePath   string `mapstructure:"sqlite_path"`
	VectorDBPath string `mapstructure:"vector_db_path"`
}

// RepoConfig holds repository tracking configuration.
type RepoConfig struct {
	Name         string   `mapstructure:"name"`
	URL          string   `mapstructure:"url"`
	Branches     []string `mapstructure:"branches"`
	ContentPaths []string `mapstructure:"content_paths"`
}

// IngestionConfig holds ingestion pipeline settings.
type IngestionConfig struct {
	CloneDepth     int      `mapstructure:"clone_depth"`
	FileExtensions []string `mapstructure:"file_extensions"`
	MaxFileSizeMB  int      `mapstructure:"max_file_size_mb"`
	Workers        int      `mapstructure:"workers"`
}

// SearchConfig holds search engine settings.
type SearchConfig struct {
	TopK         int     `mapstructure:"top_k"`
	MinScore     float64 `mapstructure:"min_score"`
	HybridWeight float64 `mapstructure:"hybrid_weight"`
}

// EvaluationConfig holds documentation quality evaluation settings.
type EvaluationConfig struct {
	LinkCheckTimeout            time.Duration `mapstructure:"link_check_timeout"`
	LinkCheckConcurrency        int           `mapstructure:"link_check_concurrency"`
	FreshnessThresholdDays      int           `mapstructure:"freshness_threshold_days"`
	DuplicateSimilarityThreshold float64      `mapstructure:"duplicate_similarity_threshold"`
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load reads configuration from file, environment, and defaults.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.documind")
		v.AddConfigPath("/etc/documind")
	}

	// Environment variables
	v.SetEnvPrefix("DOCUMIND")
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
		slog.Warn("No config file found, using defaults")
	} else {
		slog.Info("Loaded config", "file", v.ConfigFileUsed())
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &cfg, nil
}

// setDefaults configures default values.
func setDefaults(v *viper.Viper) {
	// Server
	v.SetDefault("server.http_addr", ":8080")
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")

	// Embedding
	v.SetDefault("embedding.provider", "ollama")
	v.SetDefault("embedding.model", "nomic-embed-text")
	v.SetDefault("embedding.endpoint", "http://localhost:11434")
	v.SetDefault("embedding.dimensions", 768)
	v.SetDefault("embedding.batch_size", 32)

	// Storage
	v.SetDefault("storage.data_dir", "data")
	v.SetDefault("storage.sqlite_path", "data/documind.db")
	v.SetDefault("storage.vector_db_path", "data/vectordb")

	// Ingestion
	v.SetDefault("ingestion.clone_depth", 1)
	v.SetDefault("ingestion.file_extensions", []string{".md", ".yaml", ".yml", ".go", ".json"})
	v.SetDefault("ingestion.max_file_size_mb", 10)
	v.SetDefault("ingestion.workers", 4)

	// Search
	v.SetDefault("search.top_k", 10)
	v.SetDefault("search.min_score", 0.3)
	v.SetDefault("search.hybrid_weight", 0.7)

	// Evaluation
	v.SetDefault("evaluation.link_check_timeout", "10s")
	v.SetDefault("evaluation.link_check_concurrency", 10)
	v.SetDefault("evaluation.freshness_threshold_days", 90)
	v.SetDefault("evaluation.duplicate_similarity_threshold", 0.90)

	// Logging
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "text")
}

// SetupLogging configures the global slog logger based on config.
func SetupLogging(cfg LoggingConfig) {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}
