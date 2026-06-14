// main_test.go — 对 logging 包中可测试函数的完整测试套件
//
// 本文件测试:
//   - WithRequestID / GetRequestID — 上下文信息存取
//   - getHostname — 主机名获取
//   - NewCustomHandler / CustomHandler — 自定义 Handler 创建与行为
//   - ContextHandler — 上下文日志 Handler
package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
)

// ============================================================
// 1. 上下文请求 ID 存取
// ============================================================

// TestWithAndGetRequestID 测试 WithRequestID 和 GetRequestID 的配对使用
func TestWithAndGetRequestID(t *testing.T) {
	tests := []struct {
		name      string
		requestID string
		want      string
	}{
		{name: "正常请求 ID", requestID: "req-abc-001", want: "req-abc-001"},
		{name: "空请求 ID", requestID: "", want: ""},
		{name: "带特殊字符", requestID: "req_xyz-999/1.0", want: "req_xyz-999/1.0"},
		{name: "长 ID", requestID: "a-very-long-request-id-that-spans-multiple-segments-1234567890", want: "a-very-long-request-id-that-spans-multiple-segments-1234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithRequestID(context.Background(), tt.requestID)
			got := GetRequestID(ctx)
			if got != tt.want {
				t.Errorf("GetRequestID() = %q; 期望 %q", got, tt.want)
			}
		})
	}
}

// TestGetRequestID_NoValue 测试在不含请求 ID 的 context 中获取时返回空字符串
func TestGetRequestID_NoValue(t *testing.T) {
	got := GetRequestID(context.Background())
	if got != "" {
		t.Errorf("从空 context 获取 GetRequestID() = %q; 期望空字符串", got)
	}
}

// TestGetRequestID_WrongType 测试 context 中存放了非 string 类型的值时返回空字符串
func TestGetRequestID_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), requestIDKey, 12345) // int 类型
	got := GetRequestID(ctx)
	if got != "" {
		t.Errorf("类型不匹配时 GetRequestID() = %q; 期望空字符串", got)
	}
}

// ============================================================
// 2. getHostname
// ============================================================

// TestGetHostname 测试 getHostname 返回非空字符串
func TestGetHostname(t *testing.T) {
	hostname := getHostname()
	if hostname == "" {
		t.Error("getHostname() 返回了空字符串")
	}
	// 主机名应与 os.Hostname() 一致
	expected, _ := os.Hostname()
	if hostname != expected {
		t.Errorf("getHostname() = %q; 期望 %q", hostname, expected)
	}
}

// ============================================================
// 3. CustomHandler
// ============================================================

// TestNewCustomHandler 测试创建 CustomHandler 基本属性
func TestNewCustomHandler(t *testing.T) {
	t.Run("使用 JSONHandler 创建", func(t *testing.T) {
		handler := NewCustomHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
		if handler == nil {
			t.Fatal("NewCustomHandler() 返回了 nil")
		}
		if !handler.Enabled(context.Background(), slog.LevelInfo) {
			t.Error("Info 级别应当启用")
		}
		if handler.Enabled(context.Background(), slog.LevelDebug) {
			t.Error("Debug 级别应当禁用（Level=Info）")
		}
	})

	t.Run("Debug 级别启用", func(t *testing.T) {
		handler := NewCustomHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
		if !handler.Enabled(context.Background(), slog.LevelDebug) {
			t.Error("Level=Debug 时 Debug 级别应当启用")
		}
	})
}

// TestCustomHandler_Handle 测试 CustomHandler 输出中包含 app_name 和 hostname 字段
func TestCustomHandler_Handle(t *testing.T) {
	tmpFile := createTempFile(t)
	defer tmpFile.Close()

	handler := NewCustomHandler(tmpFile, &slog.HandlerOptions{Level: slog.LevelInfo})

	logger := slog.New(handler)
	logger.Info("测试消息", "key", "value")

	output := readFileContent(t, tmpFile)
	t.Logf("CustomHandler 输出: %s", output)

	// JSONHandler 输出应为 JSON，包含固定附加字段
	if !strings.Contains(output, "app_name") {
		t.Error("输出中缺少 app_name 字段")
	}
	if !strings.Contains(output, "hostname") {
		t.Error("输出中缺少 hostname 字段")
	}
	if !strings.Contains(output, "timestamp") {
		t.Error("输出中缺少 timestamp 字段")
	}
}

// TestCustomHandler_WithAttrs 测试 WithAttrs 返回的 Handler 包含附加属性
func TestCustomHandler_WithAttrs(t *testing.T) {
	tmpFile := createTempFile(t)
	defer tmpFile.Close()

	handler := NewCustomHandler(tmpFile, &slog.HandlerOptions{Level: slog.LevelInfo})

	// 附加额外属性
	attrsLogger := slog.New(handler.WithAttrs([]slog.Attr{slog.String("module", "auth")}))
	attrsLogger.Info("带属性的日志")

	output := readFileContent(t, tmpFile)
	if !strings.Contains(output, "auth") {
		t.Error("WithAttrs 后的日志输出中缺少附加属性 'auth'")
	}
}

// ============================================================
// 4. ContextHandler
// ============================================================

// TestContextHandler 测试 ContextHandler 从 context 提取 request_id 并加入日志
func TestContextHandler(t *testing.T) {
	tmpFile := createTempFile(t)
	defer tmpFile.Close()

	baseHandler := slog.NewJSONHandler(tmpFile, &slog.HandlerOptions{Level: slog.LevelInfo})
	ctxHandler := &ContextHandler{handler: baseHandler}
	ctxLogger := slog.New(ctxHandler)

	// 设置带 request ID 的 context
	ctx := WithRequestID(context.Background(), "req-test-001")
	ctxLogger.InfoContext(ctx, "带 request_id 的日志", "path", "/api/test")

	output := readFileContent(t, tmpFile)
	t.Logf("ContextHandler 输出: %s", output)

	if !strings.Contains(output, "req-test-001") {
		t.Error("ContextHandler 输出中应包含 request_id")
	}
	if !strings.Contains(output, "request_id") {
		t.Error("ContextHandler 输出中应包含 request_id 字段名")
	}
}

// TestContextHandler_NoRequestID 测试 context 没有 request_id 时正常工作
func TestContextHandler_NoRequestID(t *testing.T) {
	tmpFile := createTempFile(t)
	defer tmpFile.Close()

	baseHandler := slog.NewJSONHandler(tmpFile, &slog.HandlerOptions{Level: slog.LevelInfo})
	ctxHandler := &ContextHandler{handler: baseHandler}
	ctxLogger := slog.New(ctxHandler)

	// 不设置 request ID
	ctxLogger.InfoContext(context.Background(), "无 request_id 的日志")

	output := readFileContent(t, tmpFile)
	if !strings.Contains(output, "无 request_id") {
		t.Error("无 request_id 时 ContextHandler 应能正常输出日志")
	}
}

// TestContextHandler_Enabled 测试 ContextHandler 的 Enabled 方法
func TestContextHandler_Enabled(t *testing.T) {
	handler := &ContextHandler{
		handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}),
	}

	tests := []struct {
		name  string
		level slog.Level
		want  bool
	}{
		{name: "Debug 禁用", level: slog.LevelDebug, want: false},
		{name: "Info 禁用", level: slog.LevelInfo, want: false},
		{name: "Warn 启用", level: slog.LevelWarn, want: true},
		{name: "Error 启用", level: slog.LevelError, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.Enabled(context.Background(), tt.level)
			if got != tt.want {
				t.Errorf("Enabled(%v) = %v; 期望 %v", tt.level, got, tt.want)
			}
		})
	}
}

// TestContextHandler_WithAttrs 测试 ContextHandler.WithAttrs 返回非 nil
func TestContextHandler_WithAttrs(t *testing.T) {
	handler := &ContextHandler{
		handler: slog.NewJSONHandler(os.Stdout, nil),
	}
	result := handler.WithAttrs([]slog.Attr{slog.String("k", "v")})
	if result == nil {
		t.Error("WithAttrs 返回了 nil")
	}
}

// TestContextHandler_WithGroup 测试 ContextHandler.WithGroup 返回非 nil
func TestContextHandler_WithGroup(t *testing.T) {
	handler := &ContextHandler{
		handler: slog.NewJSONHandler(os.Stdout, nil),
	}
	result := handler.WithGroup("test_group")
	if result == nil {
		t.Error("WithGroup 返回了 nil")
	}
}

// ============================================================
// 辅助函数
// ============================================================

// createTempFile 创建临时文件用于捕获日志输出
func createTempFile(t *testing.T) *os.File {
	t.Helper()
	f, err := os.CreateTemp("", "log-test-*.json")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	return f
}

// readFileContent 读取临时文件内容并 seek 到开头
func readFileContent(t *testing.T, f *os.File) string {
	t.Helper()
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatalf("seek 文件失败: %v", err)
	}
	data := make([]byte, 4096)
	n, err := f.Read(data)
	if err != nil && n == 0 {
		t.Fatalf("读取文件失败: %v", err)
	}
	return string(data[:n])
}