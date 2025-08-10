// Package configx provides configuration loading and validation against contracts schemas
// This adheres to spec.md requirement: "Go 端負責 Config 驗證與載入"
package configx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Config represents the platform configuration structure (SSOT-aligned, no backward-compat fields)
// observability 對齊 otel_init.go 的讀取模式：
//   observability:
//     mode: lgtm_local|grafana_cloud|gcp
//     otlp:
//       protocol: grpc|http
//       endpoint: host:port (grpc) | http(s)://host:port (http)
//       insecure: true|false
//       headers: { key: value }
//     logs:
//       mode: file|stdout|off
//       file:
//         path: ./var/log/detectviz/detectviz.log
//         max_size_mb: 50
//         max_backups: 7
//         max_age_days: 14
//         compress: true
//     profiling:
//       enabled: true|false
//       pprof_address: 127.0.0.1:6060
//       application_name: go-platform
//       tags: { service.name: go-platform, deployment.environment: dev }
//   resource/sampling 為可選，用於後續擴充
// 其他區塊維持最小可運行：grpc / plugin / memory

type Config struct {
	Env string `yaml:"env" json:"env"`

	Observability struct {
		Mode string `yaml:"mode" json:"mode"`

		OTLP struct {
			Protocol string            `yaml:"protocol" json:"protocol"` // grpc|http
			Endpoint string            `yaml:"endpoint" json:"endpoint"`
			Insecure bool              `yaml:"insecure" json:"insecure"`
			Headers  map[string]string `yaml:"headers" json:"headers"`
		} `yaml:"otlp" json:"otlp"`

		Logs struct {
			Mode string `yaml:"mode" json:"mode"` // file|stdout|off
			File struct {
				Path       string `yaml:"path" json:"path"`
				MaxSizeMB  int    `yaml:"max_size_mb" json:"max_size_mb"`
				MaxBackups int    `yaml:"max_backups" json:"max_backups"`
				MaxAgeDays int    `yaml:"max_age_days" json:"max_age_days"`
				Compress   bool   `yaml:"compress" json:"compress"`
			} `yaml:"file" json:"file"`
		} `yaml:"logs" json:"logs"`

		Resource struct {
			ServiceName    string `yaml:"service_name" json:"service_name"`
			ServiceVersion string `yaml:"service_version" json:"service_version"`
			Environment    string `yaml:"environment" json:"environment"`
		} `yaml:"resource" json:"resource"`

		Sampling struct {
			Ratio float64 `yaml:"ratio" json:"ratio"`
		} `yaml:"sampling" json:"sampling"`

		Profiling struct {
			Enabled         bool              `yaml:"enabled" json:"enabled"`
			PprofAddress    string            `yaml:"pprof_address" json:"pprof_address"` // 127.0.0.1:6060
			ApplicationName string            `yaml:"application_name" json:"application_name"`
			Tags            map[string]string `yaml:"tags" json:"tags"`
		} `yaml:"profiling" json:"profiling"`
	} `yaml:"observability" json:"observability"`

	GRPC struct {
		Listen       string `yaml:"listen" json:"listen"`
		MaxRecvBytes int    `yaml:"max_recv_bytes" json:"max_recv_bytes"`
		MaxSendBytes int    `yaml:"max_send_bytes" json:"max_send_bytes"`
		TLS          struct {
			Enabled  bool   `yaml:"enabled" json:"enabled"`
			CertFile string `yaml:"cert_file" json:"cert_file"`
			KeyFile  string `yaml:"key_file" json:"key_file"`
		} `yaml:"tls" json:"tls"`
	} `yaml:"grpc" json:"grpc"`

	Plugin struct {
		Paths    []string `yaml:"paths" json:"paths"`
		Registry string   `yaml:"registry" json:"registry"`
	} `yaml:"plugin" json:"plugin"`

	Memory struct {
		Backend           string `yaml:"backend" json:"backend"` // inmem|redis|weaviate|chroma|vertex
		DSN               string `yaml:"dsn" json:"dsn"`
		DefaultTTLSeconds int    `yaml:"default_ttl_seconds" json:"default_ttl_seconds"`
	} `yaml:"memory" json:"memory"`
}

// LoadAndValidate loads configuration file with unified precedence and validates against SSOT schema.
// Precedence (highest → lowest):
//  1. explicit argument --config
//  2. env DETECTVIZ_CONFIG_FILE
//  3. ./config.yaml (working directory)
//  4. ./contracts/config.yaml (team-shared overlay)
//  5. ./contracts/samples/config.yaml (SSOT sample; last resort)
func LoadAndValidate(configPath string) (*Config, error) {
	resolved, tried, err := resolveConfigPath(configPath)
	if err != nil {
		return nil, err
	}

	// Read configuration file
	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", resolved, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config (%s): %w", resolved, err)
	}

	applyDefaults(&cfg)

	// 環境變數覆蓋（與 Python 對齊）
	applyEnvOverrides(&cfg)

	// Validate against contracts schema
	if err := validateWithSchema(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	zap.L().Info("Configuration loaded and validated",
		zap.String("config.path", resolved),
		zap.Strings("config.search_tried", tried),
		zap.String("env", cfg.Env),
		zap.String("observability.mode", cfg.Observability.Mode),
		zap.String("otlp.protocol", cfg.Observability.OTLP.Protocol),
		zap.String("otlp.endpoint", cfg.Observability.OTLP.Endpoint),
	)

	return &cfg, nil
}

// resolveConfigPath resolves the effective config path according to the unified precedence.
// It returns (resolvedPath, triedCandidates, error).
func resolveConfigPath(arg string) (string, []string, error) {
	// If user provided an explicit path via flag, honor it strictly.
	if strings.TrimSpace(arg) != "" {
		abs := arg
		if !filepath.IsAbs(abs) {
			if a, err := filepath.Abs(abs); err == nil {
				abs = a
			}
		}
		if st, err := os.Stat(abs); err == nil && !st.IsDir() {
			return abs, []string{abs}, nil
		}
		return "", []string{abs}, fmt.Errorf("config file not found: %s", abs)
	}

	tried := []string{}

	// env override
	if env := strings.TrimSpace(os.Getenv("DETECTVIZ_CONFIG_FILE")); env != "" {
		tried = append(tried, env)
		if st, err := os.Stat(env); err == nil && !st.IsDir() {
			return env, tried, nil
		}
	}

	// working dir
	candidates := []string{"./config.yaml", "config.yaml"}
	for _, p := range candidates {
		tried = append(tried, p)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, tried, nil
		}
	}

	// contracts-based fallbacks
	if cd := findContractsDir(); cd != "" {
		p1 := filepath.Join(cd, "config.yaml")
		p2 := filepath.Join(cd, "samples", "config.yaml")
		for _, p := range []string{p1, p2} {
			tried = append(tried, p)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p, tried, nil
			}
		}
	}

	return "", tried, fmt.Errorf("no config file found; tried %v", tried)
}

// validateWithSchema validates config against contracts/schemas/config.schema.json
func validateWithSchema(cfg *Config) error {
	// Find contracts directory - should be a sibling of go-platform
	contractsDir := findContractsDir()
	if contractsDir == "" {
		return fmt.Errorf("contracts directory not found")
	}

	schemaPath := filepath.Join(contractsDir, "schemas", "config.schema.json")

	// Check if schema file exists
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		return fmt.Errorf("config schema not found at %s", schemaPath)
	}

	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
	documentLoader := gojsonschema.NewGoLoader(cfg)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, fmt.Sprintf("  - %s", desc.String()))
		}
		return fmt.Errorf("configuration validation failed:\n%v", errors)
	}

	return nil
}

// findContractsDir locates the contracts directory relative to go-platform
func findContractsDir() string {
	// Try relative path from go-platform
	paths := []string{
		"contracts",
		"../contracts",       // From go-platform root
		"../../contracts",    // From go-platform/internal
		"../../../contracts", // From go-platform/internal/configx
	}

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
	}

	return ""
}

// applyDefaults sets default values according to spec.md requirements
func applyDefaults(cfg *Config) {
	// Ensure non-nil maps/slices for schema validation
	if cfg.Observability.OTLP.Headers == nil {
		cfg.Observability.OTLP.Headers = map[string]string{}
	}
	if cfg.Plugin.Paths == nil {
		cfg.Plugin.Paths = []string{}
	}
	if cfg.Plugin.Registry == "" {
		cfg.Plugin.Registry = "file"
	}
	// gRPC payload limits must be >= 1MB (schema minimum)
	if cfg.GRPC.MaxRecvBytes < 1048576 {
		cfg.GRPC.MaxRecvBytes = 4 * 1024 * 1024
	}
	if cfg.GRPC.MaxSendBytes < 1048576 {
		cfg.GRPC.MaxSendBytes = 4 * 1024 * 1024
	}

	// Default observability mode
	if cfg.Observability.Mode == "" {
		cfg.Observability.Mode = "lgtm_local"
	}

	// Default OTLP protocol & endpoint
	if cfg.Observability.OTLP.Protocol == "" {
		cfg.Observability.OTLP.Protocol = "grpc"
	}
	if cfg.Observability.OTLP.Endpoint == "" {
		if cfg.Observability.OTLP.Protocol == "http" {
			cfg.Observability.OTLP.Endpoint = "http://localhost:4318"
		} else {
			cfg.Observability.OTLP.Endpoint = "localhost:4317"
		}
	}
	// Default logs
	if cfg.Observability.Logs.Mode == "" {
		cfg.Observability.Logs.Mode = "file"
	}
	if cfg.Observability.Logs.File.Path == "" {
		cfg.Observability.Logs.File.Path = "./var/log/detectviz/detectviz.log"
	}
	if cfg.Observability.Logs.File.MaxSizeMB == 0 {
		cfg.Observability.Logs.File.MaxSizeMB = 50
	}
	if cfg.Observability.Logs.File.MaxBackups == 0 {
		cfg.Observability.Logs.File.MaxBackups = 7
	}
	if cfg.Observability.Logs.File.MaxAgeDays == 0 {
		cfg.Observability.Logs.File.MaxAgeDays = 14
	}
	// compress 預設 true（YAML 未設定時為零值 false，這裡置 true）
	if !cfg.Observability.Logs.File.Compress {
		cfg.Observability.Logs.File.Compress = true
	}

	// Default gRPC settings（ToolBridge）
	if cfg.GRPC.Listen == "" {
		cfg.GRPC.Listen = "0.0.0.0:6606"
	}

	// Default memory backend
	if cfg.Memory.Backend == "" {
		cfg.Memory.Backend = "inmem"
	}

	// Default profiling（僅支援 pprof）
	if cfg.Observability.Profiling.PprofAddress == "" {
		cfg.Observability.Profiling.PprofAddress = "127.0.0.1:6060"
	}
	if cfg.Observability.Profiling.ApplicationName == "" {
		cfg.Observability.Profiling.ApplicationName = "go-platform"
	}
	if cfg.Observability.Profiling.Tags == nil {
		cfg.Observability.Profiling.Tags = map[string]string{}
	}
}

// GetObservabilityConfig returns a nested map (for otel_init.go)
func (c *Config) GetObservabilityConfig() map[string]interface{} {
	config := map[string]interface{}{
		"mode": c.Observability.Mode,
		"otlp": map[string]interface{}{
			"protocol": c.Observability.OTLP.Protocol,
			"endpoint": c.Observability.OTLP.Endpoint,
			"insecure": c.Observability.OTLP.Insecure,
			"headers":  c.Observability.OTLP.Headers,
		},
		"logs": map[string]interface{}{
			"mode": c.Observability.Logs.Mode,
			"file": map[string]interface{}{
				"path":         c.Observability.Logs.File.Path,
				"max_size_mb":  c.Observability.Logs.File.MaxSizeMB,
				"max_backups":  c.Observability.Logs.File.MaxBackups,
				"max_age_days": c.Observability.Logs.File.MaxAgeDays,
				"compress":     c.Observability.Logs.File.Compress,
			},
		},
		"resource": map[string]string{
			"service.name":           c.Observability.Resource.ServiceName,
			"service.version":        c.Observability.Resource.ServiceVersion,
			"deployment.environment": c.Observability.Resource.Environment,
		},
		"sampling": c.Observability.Sampling.Ratio,
	}

	// Add profiling configuration if enabled（僅 pprof）
	if c.Observability.Profiling.Enabled {
		config["profiling"] = map[string]interface{}{
			"enabled":          c.Observability.Profiling.Enabled,
			"pprof_address":    c.Observability.Profiling.PprofAddress,
			"application_name": c.Observability.Profiling.ApplicationName,
			"tags":             c.Observability.Profiling.Tags,
		}
	}

	return config
}

// applyEnvOverrides 以 DETECTVIZ_ 前綴環境變數覆蓋設定
func applyEnvOverrides(c *Config) {
	setStr(&c.Env, os.Getenv("DETECTVIZ_ENV"))

	// gRPC
	setStr(&c.GRPC.Listen, os.Getenv("DETECTVIZ__GRPC__LISTEN"))
	if v := os.Getenv("DETECTVIZ__GRPC__MAX_RECV_BYTES"); v != "" {
		if iv, err := atoi(v); err == nil {
			c.GRPC.MaxRecvBytes = iv
		}
	}
	if v := os.Getenv("DETECTVIZ__GRPC__MAX_SEND_BYTES"); v != "" {
		if iv, err := atoi(v); err == nil {
			c.GRPC.MaxSendBytes = iv
		}
	}

	// Observability base
	setStr(&c.Observability.Mode, os.Getenv("DETECTVIZ__OBSERVABILITY__MODE"))
	setStr(&c.Observability.OTLP.Protocol, os.Getenv("DETECTVIZ__OBSERVABILITY__OTLP__PROTOCOL"))
	setStr(&c.Observability.OTLP.Endpoint, os.Getenv("DETECTVIZ__OBSERVABILITY__OTLP__ENDPOINT"))
	if v := os.Getenv("DETECTVIZ__OBSERVABILITY__OTLP__INSECURE"); v != "" {
		if bv, err := parseBool(v); err == nil {
			c.Observability.OTLP.Insecure = bv
		}
	}
	if v := os.Getenv("DETECTVIZ__OBSERVABILITY__OTLP__HEADERS"); v != "" {
		c.Observability.OTLP.Headers = parseHeadersCSV(v)
	}

	// Logs
	setStr(&c.Observability.Logs.Mode, os.Getenv("DETECTVIZ__OBSERVABILITY__LOGS__MODE"))
	setStr(&c.Observability.Logs.File.Path, os.Getenv("DETECTVIZ__OBSERVABILITY__LOGS__FILE__PATH"))
	if v := os.Getenv("DETECTVIZ__OBSERVABILITY__LOGS__FILE__MAX_SIZE_MB"); v != "" {
		if iv, err := atoi(v); err == nil {
			c.Observability.Logs.File.MaxSizeMB = iv
		}
	}
	if v := os.Getenv("DETECTVIZ__OBSERVABILITY__LOGS__FILE__MAX_BACKUPS"); v != "" {
		if iv, err := atoi(v); err == nil {
			c.Observability.Logs.File.MaxBackups = iv
		}
	}
	if v := os.Getenv("DETECTVIZ__OBSERVABILITY__LOGS__FILE__MAX_AGE_DAYS"); v != "" {
		if iv, err := atoi(v); err == nil {
			c.Observability.Logs.File.MaxAgeDays = iv
		}
	}
	if v := os.Getenv("DETECTVIZ__OBSERVABILITY__LOGS__FILE__COMPRESS"); v != "" {
		if bv, err := parseBool(v); err == nil {
			c.Observability.Logs.File.Compress = bv
		}
	}

	// Profiling (pprof)
	if v := os.Getenv("DETECTVIZ__OBSERVABILITY__PROFILING__ENABLED"); v != "" {
		if bv, err := parseBool(v); err == nil {
			c.Observability.Profiling.Enabled = bv
		}
	}
	setStr(&c.Observability.Profiling.PprofAddress, os.Getenv("DETECTVIZ__OBSERVABILITY__PROFILING__PPROF_ADDRESS"))
	setStr(&c.Observability.Profiling.ApplicationName, os.Getenv("DETECTVIZ__OBSERVABILITY__PROFILING__APPLICATION_NAME"))
	if v := os.Getenv("DETECTVIZ__OBSERVABILITY__PROFILING__TAGS"); v != "" {
		c.Observability.Profiling.Tags = parseHeadersCSV(v)
	}

	// Resource
	setStr(&c.Observability.Resource.ServiceName, os.Getenv("DETECTVIZ__OBSERVABILITY__RESOURCE__SERVICE_NAME"))
	setStr(&c.Observability.Resource.ServiceVersion, os.Getenv("DETECTVIZ__OBSERVABILITY__RESOURCE__SERVICE_VERSION"))
	setStr(&c.Observability.Resource.Environment, os.Getenv("DETECTVIZ__OBSERVABILITY__RESOURCE__ENVIRONMENT"))

	// Sampling
	if v := os.Getenv("DETECTVIZ__OBSERVABILITY__SAMPLING__RATIO"); v != "" {
		if fv, err := atof(v); err == nil {
			c.Observability.Sampling.Ratio = fv
		}
	}

	// Plugins & Memory
	if v := os.Getenv("DETECTVIZ__PLUGIN__PATHS"); v != "" {
		c.Plugin.Paths = splitCSV(v)
	}
	setStr(&c.Plugin.Registry, os.Getenv("DETECTVIZ__PLUGIN__REGISTRY"))

	setStr(&c.Memory.Backend, os.Getenv("DETECTVIZ__MEMORY__BACKEND"))
	setStr(&c.Memory.DSN, os.Getenv("DETECTVIZ__MEMORY__DSN"))
	if v := os.Getenv("DETECTVIZ__MEMORY__DEFAULT_TTL_SECONDS"); v != "" {
		if iv, err := atoi(v); err == nil {
			c.Memory.DefaultTTLSeconds = iv
		}
	}
}

func setStr(dst *string, v string) {
	if v != "" {
		*dst = v
	}
}

func atoi(s string) (int, error) {
	var x int
	_, err := fmt.Sscanf(s, "%d", &x)
	return x, err
}

func parseBool(s string) (bool, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "1", "true", "t", "yes", "y", "on":
		return true, nil
	case "0", "false", "f", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool: %s", s)
	}
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// atof 解析浮點數（用於 sampling.ratio）
func atof(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%f", &f)
	return f, err
}

// parseHeadersCSV 解析以逗號分隔的 k=v 列表到 map
// 例如："authorization=Bearer xxx, x-foo=bar" → {"authorization":"Bearer xxx", "x-foo":"bar"}
func parseHeadersCSV(s string) map[string]string {
	out := map[string]string{}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		if k != "" {
			out[k] = v
		}
	}
	return out
}
