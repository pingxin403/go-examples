// main_test.go — 对 config 包中可测试函数的完整测试套件
//
// 本文件测试:
//   - DefaultConfig — 默认配置生成
//   - Validate — 配置验证（正常/异常）
//   - maskPassword — 密码脱敏
//   - generateSampleConfig — 示例配置生成
//   - LoadFromJSONFile — JSON 配置文件加载
package main

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================
// 1. DefaultConfig
// ============================================================

// TestDefaultConfig 测试默认配置的字段值
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	t.Run("Server 默认值", func(t *testing.T) {
		if cfg.Server.Host != "0.0.0.0" {
			t.Errorf("Server.Host = %q; 期望 %q", cfg.Server.Host, "0.0.0.0")
		}
		if cfg.Server.Port != 8080 {
			t.Errorf("Server.Port = %d; 期望 %d", cfg.Server.Port, 8080)
		}
		if cfg.Server.ReadTimeoutSec != 30 {
			t.Errorf("Server.ReadTimeoutSec = %d; 期望 %d", cfg.Server.ReadTimeoutSec, 30)
		}
		if cfg.Server.WriteTimeoutSec != 30 {
			t.Errorf("Server.WriteTimeoutSec = %d; 期望 %d", cfg.Server.WriteTimeoutSec, 30)
		}
	})

	t.Run("Database 默认值", func(t *testing.T) {
		if cfg.Database.Host != "localhost" {
			t.Errorf("Database.Host = %q; 期望 %q", cfg.Database.Host, "localhost")
		}
		if cfg.Database.Port != 5432 {
			t.Errorf("Database.Port = %d; 期望 %d", cfg.Database.Port, 5432)
		}
		if cfg.Database.User != "app_user" {
			t.Errorf("Database.User = %q; 期望 %q", cfg.Database.User, "app_user")
		}
		if cfg.Database.Password != "" {
			t.Errorf("Database.Password = %q; 期望空字符串", cfg.Database.Password)
		}
		if cfg.Database.DBName != "myapp" {
			t.Errorf("Database.DBName = %q; 期望 %q", cfg.Database.DBName, "myapp")
		}
		if cfg.Database.MaxConns != 10 {
			t.Errorf("Database.MaxConns = %d; 期望 %d", cfg.Database.MaxConns, 10)
		}
	})

	t.Run("Logging 默认值", func(t *testing.T) {
		if cfg.Logging.Level != "info" {
			t.Errorf("Logging.Level = %q; 期望 %q", cfg.Logging.Level, "info")
		}
		if cfg.Logging.Format != "json" {
			t.Errorf("Logging.Format = %q; 期望 %q", cfg.Logging.Format, "json")
		}
		if cfg.Logging.Output != "stdout" {
			t.Errorf("Logging.Output = %q; 期望 %q", cfg.Logging.Output, "stdout")
		}
	})

	t.Run("App 默认值", func(t *testing.T) {
		if cfg.App.Name != "go-example-app" {
			t.Errorf("App.Name = %q; 期望 %q", cfg.App.Name, "go-example-app")
		}
		if cfg.App.Version != "1.0.0" {
			t.Errorf("App.Version = %q; 期望 %q", cfg.App.Version, "1.0.0")
		}
		if cfg.App.Env != "development" {
			t.Errorf("App.Env = %q; 期望 %q", cfg.App.Env, "development")
		}
	})
}

// ============================================================
// 2. Validate — 配置验证
// ============================================================

// TestValidate_Valid 测试有效配置通过验证
func TestValidate_Valid(t *testing.T) {
	cfg := DefaultConfig()
	err := cfg.Validate()
	if err != nil {
		t.Errorf("默认配置应通过验证，但得到: %v", err)
	}
}

// TestValidate_InvalidPort 测试端口号无效
func TestValidate_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{name: "端口为 0", port: 0},
		{name: "端口为负数", port: -1},
		{name: "端口超出范围", port: 65536},
	}

	for _, tt := range tests {
		t.Run("ServerPort_"+tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Server.Port = tt.port
			if err := cfg.Validate(); err == nil {
				t.Errorf("Server.Port=%d 应验证失败", tt.port)
			}
		})
	}
}

// TestValidate_InvalidTimeout 测试超时值无效
func TestValidate_InvalidTimeout(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Server.ReadTimeoutSec = 0
	cfg.Server.WriteTimeoutSec = -5

	err := cfg.Validate()
	if err == nil {
		t.Fatal("超时值为 0 及负数时应验证失败")
	}
}

// TestValidate_EmptyDBHost 测试数据库地址为空
func TestValidate_EmptyDBHost(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Database.Host = ""
	err := cfg.Validate()
	if err == nil {
		t.Fatal("数据库地址为空时应验证失败")
	}
}

// TestValidate_EmptyDBName 测试数据库名称为空
func TestValidate_EmptyDBName(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Database.DBName = ""
	err := cfg.Validate()
	if err == nil {
		t.Fatal("数据库名称为空时应验证失败")
	}
}

// TestValidate_InvalidDBPort 测试数据库端口无效
func TestValidate_InvalidDBPort(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Database.Port = 0
	err := cfg.Validate()
	if err == nil {
		t.Fatal("数据库端口为 0 时应验证失败")
	}
}

// TestValidate_InvalidMaxConns 测试最大连接数无效
func TestValidate_InvalidMaxConns(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Database.MaxConns = 0
	err := cfg.Validate()
	if err == nil {
		t.Fatal("MaxConns 为 0 时应验证失败")
	}
}

// TestValidate_InvalidLogLevel 测试日志级别无效
func TestValidate_InvalidLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{name: "未知级别", level: "critical"},
		{name: "拼写错误", level: "infor"},
		{name: "空字符串", level: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Logging.Level = tt.level
			err := cfg.Validate()
			if err == nil {
				t.Errorf("日志级别 %q 应验证失败", tt.level)
			}
		})
	}
}

// TestValidate_InvalidLogFormat 测试日志格式无效
func TestValidate_InvalidLogFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{name: "未知格式", format: "xml"},
		{name: "空字符串", format: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Logging.Format = tt.format
			err := cfg.Validate()
			if err == nil {
				t.Errorf("日志格式 %q 应验证失败", tt.format)
			}
		})
	}
}

// TestValidate_InvalidEnv 测试运行环境无效
func TestValidate_InvalidEnv(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{name: "未知环境", env: "test"},
		{name: "拼写错误", env: "devlopment"},
		{name: "空字符串", env: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.App.Env = tt.env
			err := cfg.Validate()
			if err == nil {
				t.Errorf("运行环境 %q 应验证失败", tt.env)
			}
		})
	}
}

// TestValidate_MultipleErrors 测试多个验证错误同时返回
func TestValidate_MultipleErrors(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Server.Port = 0
	cfg.Database.Host = ""
	cfg.Logging.Level = "critical"
	cfg.App.Env = "test"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("多个配置错误时应返回错误")
	}

	errMsg := err.Error()
	// 应包含多个错误信息
	if !contains(errMsg, "端口号") {
		t.Errorf("错误消息应包含端口号验证错误: %s", errMsg)
	}
	if !contains(errMsg, "数据库地址") {
		t.Errorf("错误消息应包含数据库地址验证错误: %s", errMsg)
	}
	if !contains(errMsg, "日志级别") {
		t.Errorf("错误消息应包含日志级别验证错误: %s", errMsg)
	}
	if !contains(errMsg, "运行环境") {
		t.Errorf("错误消息应包含运行环境验证错误: %s", errMsg)
	}
}

// ============================================================
// 3. maskPassword — 密码脱敏
// ============================================================

// TestMaskPassword 测试密码脱敏函数
func TestMaskPassword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "空密码", input: "", expected: "（未设置）"},
		{name: "短密码 1 位", input: "a", expected: "*"},
		{name: "短密码 4 位", input: "abcd", expected: "****"},
		{name: "正常密码 8 位", input: "secret123", expected: "se*****23"},
		{name: "长密码 17 位", input: "mysecretpassword!", expected: "my*************d!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskPassword(tt.input)
			if got != tt.expected {
				t.Errorf("maskPassword(%q) = %q; 期望 %q", tt.input, got, tt.expected)
			}
		})
	}
}

// ============================================================
// 4. generateSampleConfig
// ============================================================

// TestGenerateSampleConfig 测试示例配置生成
func TestGenerateSampleConfig(t *testing.T) {
	result := generateSampleConfig()

	t.Run("返回非空字符串", func(t *testing.T) {
		if result == "" {
			t.Error("generateSampleConfig() 返回了空字符串")
		}
	})

	t.Run("输出为合法 JSON", func(t *testing.T) {
		if result[0] != '{' {
			t.Error("输出应以 '{' 开头（JSON 对象）")
		}
	})

	t.Run("包含关键字段", func(t *testing.T) {
		if !contains(result, "myapp") {
			t.Error("示例配置中应包含应用名称 'myapp'")
		}
		if !contains(result, "production") {
			t.Error("示例配置中应包含环境 'production'")
		}
		if !contains(result, "db.example.com") {
			t.Error("示例配置中应包含数据库地址 'db.example.com'")
		}
	})
}

// ============================================================
// 5. LoadFromJSONFile
// ============================================================

// TestLoadFromJSONFile 测试从 JSON 文件加载配置
func TestLoadFromJSONFile(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	t.Run("加载合法 JSON 文件", func(t *testing.T) {
		jsonContent := `{"server": {"port": 9090, "host": "0.0.0.0"}}`
		configPath := filepath.Join(tmpDir, "valid.json")
		if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
			t.Fatalf("写入临时配置文件失败: %v", err)
		}

		cfg := DefaultConfig()
		if err := LoadFromJSONFile(&cfg, configPath); err != nil {
			t.Fatalf("LoadFromJSONFile 失败: %v", err)
		}

		if cfg.Server.Port != 9090 {
			t.Errorf("Server.Port = %d; 期望 %d", cfg.Server.Port, 9090)
		}
		// 未在 JSON 中指定的字段应保留默认值
		if cfg.Server.ReadTimeoutSec != 30 {
			t.Errorf("未覆盖字段 ReadTimeoutSec = %d; 期望 30", cfg.Server.ReadTimeoutSec)
		}
	})

	t.Run("不存在的文件应返回 nil（跳过）", func(t *testing.T) {
		cfg := DefaultConfig()
		err := LoadFromJSONFile(&cfg, filepath.Join(tmpDir, "nonexistent.json"))
		if err != nil {
			t.Errorf("不存在的文件应返回 nil（跳过），但得到: %v", err)
		}
	})

	t.Run("非法 JSON 应返回错误", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "invalid.json")
		if err := os.WriteFile(configPath, []byte("{invalid json}"), 0644); err != nil {
			t.Fatalf("写入临时配置文件失败: %v", err)
		}

		cfg := DefaultConfig()
		err := LoadFromJSONFile(&cfg, configPath)
		if err == nil {
			t.Error("非法 JSON 应返回错误")
		}
	})
}

// TestLoadFromJSONFile_PartialOverride 测试 JSON 文件部分覆盖默认配置
func TestLoadFromJSONFile_PartialOverride(t *testing.T) {
	tmpDir := t.TempDir()
	jsonContent := `{
		"logging": {"level": "debug"},
		"app": {"env": "staging"}
	}`
	configPath := filepath.Join(tmpDir, "partial.json")
	if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("写入临时配置文件失败: %v", err)
	}

	cfg := DefaultConfig()
	if err := LoadFromJSONFile(&cfg, configPath); err != nil {
		t.Fatalf("LoadFromJSONFile 失败: %v", err)
	}

	// 被覆盖的字段
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q; 期望 %q", cfg.Logging.Level, "debug")
	}
	if cfg.App.Env != "staging" {
		t.Errorf("App.Env = %q; 期望 %q", cfg.App.Env, "staging")
	}
	// 未被覆盖的字段应保留默认值
	if cfg.Server.Port != 8080 {
		t.Errorf("未覆盖字段 Server.Port = %d; 期望 %d", cfg.Server.Port, 8080)
	}
	if cfg.Database.MaxConns != 10 {
		t.Errorf("未覆盖字段 Database.MaxConns = %d; 期望 %d", cfg.Database.MaxConns, 10)
	}
}

// ============================================================
// 辅助函数
// ============================================================

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

// containsStr 简单子串查找，避免使用 strings 包之外的非标准库
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}