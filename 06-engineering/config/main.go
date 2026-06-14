// Go 配置管理示例
//
// 本文件演示 Go 中常见的配置管理方案（零外部依赖）：
//   - 环境变量（os.Getenv, os.LookupEnv）
//   - 命令行标志（flag 包）
//   - JSON 配置文件解析
//   - 配置验证
//   - 分层配置：默认值 < 配置文件 < 环境变量 < 命令行参数
//
// 运行方式：
//
//	# 默认配置运行
//	go run main.go
//
//	# 通过环境变量覆盖
//	PORT=9090 DB_HOST=prod-db.example.com go run main.go
//
//	# 通过命令行参数覆盖
//	go run main.go -port=8080 -db-host=localhost -db-port=5432
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ============================================================
// 1. 配置结构体定义
// ============================================================

// AppConfig 应用配置，支持多层覆盖：
// 默认值 → 配置文件 → 环境变量 → 命令行参数
type AppConfig struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Logging  LoggingConfig  `json:"logging"`
	App      AppInfo        `json:"app"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	ReadTimeoutSec  int    `json:"read_timeout_sec"`
	WriteTimeoutSec int    `json:"write_timeout_sec"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	MaxConns int    `json:"max_conns"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `json:"level"`  // debug, info, warn, error
	Format string `json:"format"` // text, json
	Output string `json:"output"` // stdout, stderr, file
}

// AppInfo 应用信息
type AppInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Env     string `json:"env"` // dev, staging, production
}

// ============================================================
// 2. 默认配置
// ============================================================

// DefaultConfig 返回应用默认配置
func DefaultConfig() AppConfig {
	return AppConfig{
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			ReadTimeoutSec:  30,
			WriteTimeoutSec: 30,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "app_user",
			Password: "", // 生产环境必须通过环境变量或密钥管理服务设置
			DBName:   "myapp",
			MaxConns: 10,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		App: AppInfo{
			Name:    "go-example-app",
			Version: "1.0.0",
			Env:     "development",
		},
	}
}

// ============================================================
// 3. 配置文件读取（JSON）
// ============================================================

// LoadFromJSONFile 从 JSON 文件加载配置，覆盖默认值
// 文件格式示例见 config.json
func LoadFromJSONFile(cfg *AppConfig, filePath string) error {
	// 检查文件是否存在，不存在则跳过（非强制）
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("[config] 配置文件不存在，跳过: %s\n", filePath)
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败 %s: %w", filePath, err)
	}

	// 将 JSON 反序列化到已有配置上，仅覆盖 JSON 中出现的字段
	// 使用 json.Unmarshal 而非 json.Decoder，因为我们需要保留已有字段
	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("解析配置文件失败 %s: %w", filePath, err)
	}

	fmt.Printf("[config] 已加载配置文件: %s\n", filePath)
	return nil
}

// ============================================================
// 4. 环境变量覆盖
// ============================================================

// LoadFromEnv 从环境变量加载配置，覆盖现有值
//
// 支持的环境变量命名规则：
//   - APP_<SECTION>_<KEY> — 例如 APP_SERVER_PORT=9090
//   - 也支持直接的环境变量名，如 PORT, DB_HOST
func LoadFromEnv(cfg *AppConfig) {
	fmt.Println("[config] 从环境变量加载覆盖配置...")

	// 环境变量映射表：envKey → 通过闭包更新配置的函数
	envMappings := []struct {
		envKey string
		apply  func(value string) error
		desc   string
	}{
		// Server
		{"APP_SERVER_HOST", func(v string) error { cfg.Server.Host = v; return nil }, "服务监听地址"},
		{"APP_SERVER_PORT", func(v string) error { p, e := strconv.Atoi(v); cfg.Server.Port = p; return e }, "服务端口"},
		{"PORT", func(v string) error { p, e := strconv.Atoi(v); cfg.Server.Port = p; return e }, "服务端口（简写）"},
		{"APP_SERVER_READ_TIMEOUT", func(v string) error { t, e := strconv.Atoi(v); cfg.Server.ReadTimeoutSec = t; return e }, "读取超时"},
		{"APP_SERVER_WRITE_TIMEOUT", func(v string) error { t, e := strconv.Atoi(v); cfg.Server.WriteTimeoutSec = t; return e }, "写入超时"},

		// Database
		{"APP_DB_HOST", func(v string) error { cfg.Database.Host = v; return nil }, "数据库地址"},
		{"APP_DB_PORT", func(v string) error { p, e := strconv.Atoi(v); cfg.Database.Port = p; return e }, "数据库端口"},
		{"APP_DB_USER", func(v string) error { cfg.Database.User = v; return nil }, "数据库用户"},
		{"APP_DB_PASSWORD", func(v string) error { cfg.Database.Password = v; return nil }, "数据库密码"},
		{"APP_DB_NAME", func(v string) error { cfg.Database.DBName = v; return nil }, "数据库名称"},
		{"APP_DB_MAX_CONNS", func(v string) error { c, e := strconv.Atoi(v); cfg.Database.MaxConns = c; return e }, "最大连接数"},

		// Logging
		{"APP_LOG_LEVEL", func(v string) error { cfg.Logging.Level = v; return nil }, "日志级别"},
		{"APP_LOG_FORMAT", func(v string) error { cfg.Logging.Format = v; return nil }, "日志格式"},
		{"APP_LOG_OUTPUT", func(v string) error { cfg.Logging.Output = v; return nil }, "日志输出"},

		// App
		{"APP_ENV", func(v string) error { cfg.App.Env = v; return nil }, "运行环境"},
		{"APP_NAME", func(v string) error { cfg.App.Name = v; return nil }, "应用名称"},
		{"APP_VERSION", func(v string) error { cfg.App.Version = v; return nil }, "应用版本"},
	}

	for _, m := range envMappings {
		if value, ok := os.LookupEnv(m.envKey); ok {
			if err := m.apply(value); err != nil {
				fmt.Printf("[config] ⚠️ 环境变量 %s=%q 解析失败: %v\n", m.envKey, value, err)
			} else {
				fmt.Printf("[config]   环境变量 %s=%q (%s)\n", m.envKey, value, m.desc)
			}
		}
	}
}

// ============================================================
// 5. 命令行参数
// ============================================================

// LoadFromFlags 从命令行参数加载配置，覆盖现有值
// 命令行参数的优先级最高
func LoadFromFlags(cfg *AppConfig) {
	fmt.Println("[config] 从命令行参数加载覆盖配置...")

	// 定义所有命令行标志
	host := flag.String("host", cfg.Server.Host, "服务监听地址")
	port := flag.Int("port", cfg.Server.Port, "服务端口")
	dbHost := flag.String("db-host", cfg.Database.Host, "数据库地址")
	dbPort := flag.Int("db-port", cfg.Database.Port, "数据库端口")
	dbUser := flag.String("db-user", cfg.Database.User, "数据库用户")
	dbName := flag.String("db-name", cfg.Database.DBName, "数据库名称")
	dbMaxConns := flag.Int("db-max-conns", cfg.Database.MaxConns, "最大连接数")
	logLevel := flag.String("log-level", cfg.Logging.Level, "日志级别 (debug/info/warn/error)")
	logFormat := flag.String("log-format", cfg.Logging.Format, "日志格式 (text/json)")
	appEnv := flag.String("env", cfg.App.Env, "运行环境 (dev/staging/production)")
	configFile := flag.String("config", "", "JSON 配置文件路径")
	showHelp := flag.Bool("help", false, "显示帮助信息")

	// 自定义使用信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "用法: %s [选项]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n环境变量:\n")
		fmt.Fprintf(os.Stderr, "  所有选项也可以通过 APP_<SECTION>_<KEY> 环境变量设置\n")
		fmt.Fprintf(os.Stderr, "  例如: APP_SERVER_PORT=9090 APP_DB_HOST=prod-db.example.com\n")
	}

	flag.Parse()

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// 如果指定了配置文件，先加载（早于其他命令行参数，但会被后续同名字段覆盖）
	if *configFile != "" {
		if err := LoadFromJSONFile(cfg, *configFile); err != nil {
			fmt.Printf("[config] ⚠️ 加载配置文件失败: %v\n", err)
		}
	}

	// 仅当对应的 flag 被显式设置时才覆盖
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "host":
			cfg.Server.Host = *host
		case "port":
			cfg.Server.Port = *port
		case "db-host":
			cfg.Database.Host = *dbHost
		case "db-port":
			cfg.Database.Port = *dbPort
		case "db-user":
			cfg.Database.User = *dbUser
		case "db-name":
			cfg.Database.DBName = *dbName
		case "db-max-conns":
			cfg.Database.MaxConns = *dbMaxConns
		case "log-level":
			cfg.Logging.Level = *logLevel
		case "log-format":
			cfg.Logging.Format = *logFormat
		case "env":
			cfg.App.Env = *appEnv
		}
	})

	fmt.Printf("[config]   通过命令行设置: %s\n", formatVisitedFlags())
}

// formatVisitedFlags 格式化已访问的 flag 名称用于显示
func formatVisitedFlags() string {
	var names []string
	flag.Visit(func(f *flag.Flag) {
		names = append(names, fmt.Sprintf("-%s=%s", f.Name, f.Value.String()))
	})
	if len(names) == 0 {
		return "（无）"
	}
	return strings.Join(names, ", ")
}

// ============================================================
// 6. 配置验证
// ============================================================

// Validate 验证配置的合法性，返回所有验证错误
func (c *AppConfig) Validate() error {
	var errs []string

	// Server 验证
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errs = append(errs, fmt.Sprintf("端口号无效: %d（有效范围 1-65535）", c.Server.Port))
	}
	if c.Server.ReadTimeoutSec <= 0 {
		errs = append(errs, fmt.Sprintf("读取超时无效: %d（必须大于 0）", c.Server.ReadTimeoutSec))
	}
	if c.Server.WriteTimeoutSec <= 0 {
		errs = append(errs, fmt.Sprintf("写入超时无效: %d（必须大于 0）", c.Server.WriteTimeoutSec))
	}

	// Database 验证
	if c.Database.Host == "" {
		errs = append(errs, "数据库地址不能为空")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		errs = append(errs, fmt.Sprintf("数据库端口无效: %d", c.Database.Port))
	}
	if c.Database.DBName == "" {
		errs = append(errs, "数据库名称不能为空")
	}
	if c.Database.MaxConns <= 0 {
		errs = append(errs, fmt.Sprintf("最大连接数无效: %d（必须大于 0）", c.Database.MaxConns))
	}

	// Logging 验证
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[strings.ToLower(c.Logging.Level)] {
		errs = append(errs, fmt.Sprintf("日志级别无效: %s（有效值: debug, info, warn, error）", c.Logging.Level))
	}
	validFormats := map[string]bool{"text": true, "json": true}
	if !validFormats[strings.ToLower(c.Logging.Format)] {
		errs = append(errs, fmt.Sprintf("日志格式无效: %s（有效值: text, json）", c.Logging.Format))
	}

	// App 验证
	validEnvs := map[string]bool{"development": true, "staging": true, "production": true}
	if !validEnvs[strings.ToLower(c.App.Env)] {
		errs = append(errs, fmt.Sprintf("运行环境无效: %s（有效值: development, staging, production）", c.App.Env))
	}

	if len(errs) > 0 {
		return fmt.Errorf("配置验证失败:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

// ============================================================
// 7. 配置文件生成示例
// ============================================================

// generateSampleConfig 生成一个示例配置文件
func generateSampleConfig() string {
	cfg := AppConfig{
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			ReadTimeoutSec:  60,
			WriteTimeoutSec: 60,
		},
		Database: DatabaseConfig{
			Host:     "db.example.com",
			Port:     5432,
			User:     "prod_user",
			Password: "${DB_PASSWORD}", // 生产环境可通过环境变量替换
			DBName:   "production_db",
			MaxConns: 25,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		App: AppInfo{
			Name:    "myapp",
			Version: "2.0.0",
			Env:     "production",
		},
	}

	data, _ := json.MarshalIndent(cfg, "", "  ")
	return string(data)
}

// ============================================================
// 8. 主流程 — 分层配置加载
// ============================================================

// LoadConfig 执行分层配置加载：默认值 → 配置文件 → 环境变量 → 命令行
// 后面的层级覆盖前面的层级
func LoadConfig(configPath string) (*AppConfig, error) {
	fmt.Println("========================================")
	fmt.Println("配置加载流程")
	fmt.Println("========================================")

	// 第1层：默认值
	cfg := DefaultConfig()
	fmt.Printf("[config] 第1层 - 默认配置已加载\n\n")

	// 第2层：配置文件（如果存在）
	if configPath != "" {
		if err := LoadFromJSONFile(&cfg, configPath); err != nil {
			fmt.Printf("[config] ⚠️ 配置文件加载失败: %v\n", err)
		}
		fmt.Println()
	}

	// 第3层：环境变量
	LoadFromEnv(&cfg)
	fmt.Println()

	// 第4层：命令行参数（优先级最高）
	// 注意：flag.Parse() 在 LoadFromFlags 内部调用
	LoadFromFlags(&cfg)
	fmt.Println()

	// 配置验证
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}
	fmt.Println("[config] ✅ 配置验证通过")

	return &cfg, nil
}

func main() {
	fmt.Println("Go 配置管理示例")
	fmt.Println()

	// 默认配置文件路径（可通过命令行 -config 覆盖）
	defaultConfigPath := "config.json"
	if v := os.Getenv("APP_CONFIG_FILE"); v != "" {
		defaultConfigPath = v
	}

	cfg, err := LoadConfig(defaultConfigPath)
	if err != nil {
		fmt.Printf("\n❌ %v\n", err)
		os.Exit(1)
	}

	// 打印最终配置
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("最终配置")
	fmt.Println("========================================")
	printConfig(cfg)

	// 生成示例配置文件
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("示例配置文件（config.json）:")
	fmt.Println("========================================")
	fmt.Println(generateSampleConfig())

	// 提示文件变更监听模式（注释中展示）
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("配置文件热更新模式（注释说明）")
	fmt.Println("========================================")
	fmt.Println("// 生产环境通常会结合 fsnotify 实现配置热更新:")
	fmt.Println("// watcher, _ := fsnotify.NewWatcher()")
	fmt.Println("// watcher.Add(\"config.json\")")
	fmt.Println("// go func() {")
	fmt.Println("//     for event := range watcher.Events {")
	fmt.Println("//         if event.Op&fsnotify.Write != 0 {")
	fmt.Println("//             // 配置文件变更，重新加载")
	fmt.Println("//             newCfg, _ := LoadConfig(\"config.json\")")
	fmt.Println("//             // 原子替换全局配置指针")
	fmt.Println("//             atomic.StorePointer(&globalConfig, unsafe.Pointer(newCfg))")
	fmt.Println("//         }")
	fmt.Println("//     }")
	fmt.Println("// }()")
}

// printConfig 友好打印配置内容（隐藏密码）
func printConfig(cfg *AppConfig) {
	fmt.Printf("Server:\n")
	fmt.Printf("  地址: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("  读取超时: %ds\n", cfg.Server.ReadTimeoutSec)
	fmt.Printf("  写入超时: %ds\n", cfg.Server.WriteTimeoutSec)
	fmt.Printf("Database:\n")
	fmt.Printf("  主机: %s:%d\n", cfg.Database.Host, cfg.Database.Port)
	fmt.Printf("  用户: %s\n", cfg.Database.User)
	fmt.Printf("  密码: %s\n", maskPassword(cfg.Database.Password))
	fmt.Printf("  数据库: %s\n", cfg.Database.DBName)
	fmt.Printf("  最大连接: %d\n", cfg.Database.MaxConns)
	fmt.Printf("Logging:\n")
	fmt.Printf("  级别: %s\n", cfg.Logging.Level)
	fmt.Printf("  格式: %s\n", cfg.Logging.Format)
	fmt.Printf("  输出: %s\n", cfg.Logging.Output)
	fmt.Printf("App:\n")
	fmt.Printf("  名称: %s\n", cfg.App.Name)
	fmt.Printf("  版本: %s\n", cfg.App.Version)
	fmt.Printf("  环境: %s\n", cfg.App.Env)
}

func maskPassword(pwd string) string {
	if pwd == "" {
		return "（未设置）"
	}
	if len(pwd) <= 4 {
		return strings.Repeat("*", len(pwd))
	}
	return pwd[:2] + strings.Repeat("*", len(pwd)-4) + pwd[len(pwd)-2:]
}