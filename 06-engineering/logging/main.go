// Go 日志处理示例
//
// 本文件演示 Go 中三种主要的日志方案：
//   - 标准库 log 包 — 简单、直接，适合基础日志需求
//   - log/slog 包 — Go 1.21 引入的结构化日志，支持 JSON/Text 格式和日志级别
//   - 自定义 Handler — 扩展 slog 实现企业级日志格式
//
// 运行方式：
//
//	go run main.go
package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"
)

// ============================================================
// 1. 标准 log 包 — 最基本的日志输出
// ============================================================

func demoStandardLog() {
	log.Println("=== 1. 标准 log 包 ===")

	// log.Print / log.Printf / log.Println — 输出到 stderr
	log.Print("这是 log.Print 输出")
	log.Printf("这是 log.Printf 输出（格式化: %s=%d）", "count", 42)
	log.Println("这是 log.Println 输出（自动追加换行）")

	// log.Fatal — 打印后调用 os.Exit(1)
	// 取消注释体验：log.Fatal("致命错误，程序退出")

	// 自定义 logger：设置前缀和标志位
	customLogger := log.New(
		os.Stdout,                    // 输出目标
		"[CUSTOM] ",                  // 前缀
		log.Ldate|log.Ltime|log.Lshortfile, // 标志位：日期+时间+源码位置
	)
	customLogger.Println("这是自定义 logger 的输出")
	customLogger.Printf("请求处理耗时: %dms", 156)

	log.Println() // 空行分隔
}

// ============================================================
// 2. log/slog 包 — Go 1.21 结构化日志
// ============================================================

func demoSlog() {
	log.Println("=== 2. slog 结构化日志 ===")

	// ---------- 2a. 默认 TextHandler ----------
	log.Println("--- TextHandler（默认）---")
	textLogger := slog.Default()
	textLogger.Info("服务启动成功",
		"port", 8080,
		"env", "production",
		"version", "1.2.3",
	)

	// ---------- 2b. JSONHandler（推荐生产环境使用）----------
	log.Println("--- JSONHandler（生产环境推荐）---")
	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // 设置最低日志级别
	}))
	slog.SetDefault(jsonLogger) // 设置为全局默认 logger

	jsonLogger.Debug("调试信息",
		"module", "auth",
		"user_id", "u-1001",
		"latency_ms", 12,
	)
	jsonLogger.Info("用户登录成功",
		"user_id", "u-1001",
		"ip", "192.168.1.100",
		"login_time", time.Now().Format(time.RFC3339),
	)
	jsonLogger.Warn("API 响应缓慢",
		"path", "/api/v1/users",
		"duration_ms", 2530,
		"threshold_ms", 1000,
	)
	jsonLogger.Error("数据库连接失败",
		"db_host", "postgres.example.com",
		"db_port", 5432,
		"error", "connection refused",
		"retry_count", 3,
	)

	// 恢复默认 logger 为 TextHandler 以免影响后续输出
	slog.SetDefault(textLogger)
	log.Println()
}

// ============================================================
// 3. 日志级别演示 — Debug < Info < Warn < Error
// ============================================================

func demoLogLevels() {
	log.Println("=== 3. 日志级别 ===")

	// 创建一个只输出 Warn 及以上级别的 logger（过滤 Debug 和 Info）
	leveledHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})
	leveledLogger := slog.New(leveledHandler)

	log.Println("下面只显示 Warn 和 Error 级别的日志（Debug 和 Info 被过滤）:")
	leveledLogger.Debug("这条不会出现（级别太低）")
	leveledLogger.Info("这条也不会出现（级别太低）")
	leveledLogger.Warn("⚠️ 警告：磁盘使用率超过 80%",
		"disk", "/dev/sda1",
		"used_percent", 83,
	)
	leveledLogger.Error("❌ 错误：磁盘写入失败",
		"disk", "/dev/sda1",
		"error", "no space left on device",
	)

	// 动态调整日志级别
	var programLevel = new(slog.LevelVar)
	programLevel.Set(slog.LevelDebug) // 初始为 Debug 级别
	dynamicLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: programLevel,
	}))
	log.Println("动态级别（当前 Debug），所有日志都会显示:")
	dynamicLogger.Debug("调试信息可见")
	dynamicLogger.Info("常规信息可见")

	programLevel.Set(slog.LevelWarn) // 调高到 Warn
	log.Println("级别调高到 Warn，Debug 和 Info 被过滤:")
	dynamicLogger.Debug("这条被过滤了")
	dynamicLogger.Info("这条也被过滤了")
	dynamicLogger.Warn("警告可见")
	dynamicLogger.Error("错误可见")

	log.Println()
}

// ============================================================
// 4. 自定义 Handler — 实现企业级日志格式
// ============================================================

// CustomHandler 自定义 slog.Handler，在每条日志前添加时间戳和调用者信息
type CustomHandler struct {
	handler slog.Handler
}

// NewCustomHandler 创建自定义 Handler
func NewCustomHandler(output *os.File, opts *slog.HandlerOptions) *CustomHandler {
	return &CustomHandler{
		handler: slog.NewJSONHandler(output, opts),
	}
}

// Enabled 判断指定级别是否启用
func (h *CustomHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.handler.Enabled(context.Background(), level)
}

// Handle 处理日志记录，添加自定义字段
func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	// 在每条日志中添加固定字段
	r.AddAttrs(
		slog.String("app_name", "go-examples"),
		slog.String("hostname", getHostname()),
		slog.String("timestamp", r.Time.Format(time.RFC3339Nano)),
	)
	return h.handler.Handle(ctx, r)
}

// WithAttrs 返回附加了指定属性的 Handler 副本
func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomHandler{handler: h.handler.WithAttrs(attrs)}
}

// WithGroup 返回指定分组的 Handler 副本
func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return &CustomHandler{handler: h.handler.WithGroup(name)}
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func demoCustomHandler() {
	log.Println("=== 4. 自定义 Handler ===")

	customLogger := slog.New(NewCustomHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	customLogger.Info("用户注册成功",
		"user_id", "u-2001",
		"email", "user@example.com",
		"source", "wechat",
	)
	customLogger.Warn("登录异常检测",
		"user_id", "u-2001",
		"attempts", 5,
		"ip", "10.0.0.1",
		"reason", "too many failed attempts",
	)

	log.Println()
}

// ============================================================
// 5. 上下文日志 — 在请求链路中传递日志上下文
// ============================================================

// 自定义上下文键类型，避免与其他包冲突
type contextKey string

const requestIDKey contextKey = "request_id"

// WithRequestID 将请求 ID 存入 context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID 从 context 中取出请求 ID
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// ContextHandler 包装另一个 Handler，自动从 context 提取请求 ID 并加入日志
type ContextHandler struct {
	handler slog.Handler
}

func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if requestID := GetRequestID(ctx); requestID != "" {
		r.AddAttrs(slog.String("request_id", requestID))
	}
	return h.handler.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{handler: h.handler.WithGroup(name)}
}

func demoContextualLogging() {
	log.Println("=== 5. 上下文日志 ===")

	// 创建支持上下文的 logger
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	ctxHandler := &ContextHandler{handler: baseHandler}
	ctxLogger := slog.New(ctxHandler)

	// 模拟请求处理：不同请求有不同 ID
	requestIDs := []string{"req-abc-001", "req-abc-002", "req-abc-003"}
	for _, reqID := range requestIDs {
		ctx := WithRequestID(context.Background(), reqID)
		// 日志自动携带 request_id 字段
		ctxLogger.InfoContext(ctx, "请求处理开始",
			"path", "/api/orders",
			"method", "POST",
		)
		ctxLogger.InfoContext(ctx, "数据库查询完成",
			"query", "SELECT * FROM orders",
			"rows", 10,
			"duration_ms", 45,
		)
		ctxLogger.WarnContext(ctx, "请求处理缓慢",
			"total_duration_ms", 3200,
		)
	}
}

func main() {
	log.Println("==========================================")
	log.Println("Go 日志处理示例")
	log.Println("==========================================")
	log.Println()

	demoStandardLog()
	demoSlog()
	demoLogLevels()
	demoCustomHandler()
	demoContextualLogging()

	log.Println()
	log.Println("所有日志示例执行完毕")
}