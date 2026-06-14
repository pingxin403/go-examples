// Go HTTP 中间件示例
//
// 本文件演示 Go 标准库 net/http 中的中间件模式（不使用任何第三方框架）：
//   - 日志中间件（请求日志记录）
//   - 认证中间件（Bearer Token 验证）
//   - 限流中间件（基于令牌桶）
//   - 中间件链（组合多个中间件）
//   - 超时中间件
//   - 恢复中间件（panic 恢复）
//
// Go 中 HTTP 中间件的本质是：接收一个 http.Handler，返回一个新的 http.Handler，
// 在它前后执行逻辑。签名：func(next http.Handler) http.Handler
//
// 运行方式：
//
//	go run main.go
//	# 在另一个终端：
//	curl -i http://localhost:8080/api/public/hello
//	curl -i http://localhost:8080/api/protected/data
//	curl -i -H "Authorization: Bearer valid-token-123" http://localhost:8080/api/protected/data
//	# 限流测试：
//	for i in {1..15}; do curl -i http://localhost:8080/api/public/hello; done
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// ============================================================
// 上下文键类型 — 用于在中间件间传递数据
// ============================================================

type contextKey string

const (
	// UserIDKey 用户 ID 上下文键（认证中间件设置）
	UserIDKey contextKey = "user_id"
	// RequestIDKey 请求 ID 上下文键（日志中间件设置）
	RequestIDKey contextKey = "request_id"
	// StartTimeKey 请求开始时间（日志中间件设置）
	StartTimeKey contextKey = "start_time"
)

// GetUserID 从 context 获取用户 ID
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// GetRequestID 从 context 获取请求 ID
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// ============================================================
// 1. 日志中间件（Logging Middleware）
// ============================================================

// LoggingMiddleware 记录每个请求的方法、路径、状态码、耗时
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 为请求生成唯一 ID
		requestID := fmt.Sprintf("req-%d", rand.Int63()%1000000)
		start := time.Now()

		// 将请求 ID 和开始时间存入 context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		ctx = context.WithValue(ctx, StartTimeKey, start)
		r = r.WithContext(ctx)

		// 使用自定义 ResponseWriter 捕获状态码
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// 调用下一个 handler（执行业务逻辑）
		next.ServeHTTP(lrw, r)

		// 请求结束后记录日志
		duration := time.Since(start)
		log.Printf("[%s] %s %s → %d (%s)\n",
			requestID,
			r.Method,
			r.URL.Path,
			lrw.statusCode,
			duration,
		)
	})
}

// loggingResponseWriter 包装 http.ResponseWriter，捕获状态码
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// ============================================================
// 2. 认证中间件（Authentication Middleware）
// ============================================================

// 模拟有效的 token 集合
var validTokens = map[string]string{
	"valid-token-123": "user-001",
	"valid-token-456": "user-002",
	"admin-token-789": "admin-001",
}

// AuthMiddleware 验证 Bearer Token，通过后设置 user_id 到 context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从 Authorization 头提取 token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error":   "missing_authorization_header",
				"message": "请求头缺少 Authorization",
			})
			return
		}

		// 提取 Bearer token
		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error":   "invalid_auth_format",
				"message": "Authorization 格式应为: Bearer <token>",
			})
			return
		}

		token := authHeader[len(bearerPrefix):]

		// 验证 token
		userID, ok := validTokens[token]
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error":   "invalid_token",
				"message": "Token 无效或已过期",
			})
			return
		}

		// 将用户 ID 存入 context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		r = r.WithContext(ctx)

		// 继续处理请求
		next.ServeHTTP(w, r)
	})
}

// ============================================================
// 3. 限流中间件（Rate Limiting Middleware）
// ============================================================

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	mu          sync.Mutex
	rate        float64   // 每秒填充速率
	capacity    float64   // 桶容量
	tokens      float64   // 当前令牌数
	lastRefill  time.Time // 上次填充时间
}

// NewTokenBucket 创建令牌桶限流器
func NewTokenBucket(rate, capacity float64) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity, // 初始满桶
		lastRefill: time.Now(),
	}
}

// Allow 判断是否允许通过（消费一个令牌）
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// 计算需要补充的令牌数
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now

	// 消费令牌
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// RateLimitMiddleware 限流中间件
// rate: 每秒允许请求数, capacity: 突发容量
func RateLimitMiddleware(rate, capacity float64) func(http.Handler) http.Handler {
	bucket := NewTokenBucket(rate, capacity)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !bucket.Allow() {
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", 1/rate))
				writeJSON(w, http.StatusTooManyRequests, map[string]string{
					"error":   "rate_limit_exceeded",
					"message": fmt.Sprintf("请求过于频繁，每秒最多 %.0f 次", rate),
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================
// 4. 超时中间件（Timeout Middleware）
// ============================================================

// TimeoutMiddleware 为请求设置超时控制
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 创建带超时的 context
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)

			// 创建一个 channel 来接收 handler 完成信号
			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			// 等待 handler 完成或超时
			select {
			case <-done:
				// handler 正常完成
			case <-ctx.Done():
				// 超时
				log.Printf("[超时] %s %s 超过 %v\n", r.Method, r.URL.Path, timeout)
				writeJSON(w, http.StatusGatewayTimeout, map[string]string{
					"error":   "request_timeout",
					"message": fmt.Sprintf("请求处理超时 (%v)", timeout),
				})
			}
		})
	}
}

// ============================================================
// 5. 恢复中间件（Recovery Middleware）
// ============================================================

// RecoveryMiddleware 捕获 handler 中的 panic，防止进程崩溃
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[PANIC] %s %s: %v\n", r.Method, r.URL.Path, rec)
				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error":   "internal_server_error",
					"message": "服务器内部错误",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ============================================================
// 6. 中间件链（Middleware Chain）
// ============================================================

// Middleware 中间件类型别名
type Middleware func(http.Handler) http.Handler

// Chain 将多个中间件组合成一个链。
// 执行顺序：最后一个传入的中间件最先执行（洋葱模型）。
//
// 例如 Chain(h, A, B, C) 执行顺序: A → B → C → h → C → B → A
func Chain(final http.Handler, middlewares ...Middleware) http.Handler {
	// 从后向前包装
	for i := len(middlewares) - 1; i >= 0; i-- {
		final = middlewares[i](final)
	}
	return final
}

// ============================================================
// 7. 业务 Handler
// ============================================================

// handleHello 公共接口 — 不需要认证
func handleHello(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"message":    "你好，世界！",
		"request_id": GetRequestID(r.Context()),
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, data)
}

// handleProtectedData 受保护接口 — 需要认证
func handleProtectedData(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	data := map[string]interface{}{
		"message":    "这是受保护的数据",
		"user_id":    userID,
		"request_id": GetRequestID(r.Context()),
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, data)
}

// handleSlowRequest 模拟慢请求（测试超时中间件）
func handleSlowRequest(w http.ResponseWriter, r *http.Request) {
	// 模拟耗时操作
	time.Sleep(3 * time.Second)
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "慢请求已完成",
	})
}

// handlePanic 模拟 panic（测试恢复中间件）
func handlePanic(w http.ResponseWriter, r *http.Request) {
	panic("模拟的服务器内部错误")
}

// ============================================================
// 辅助函数
// ============================================================

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// ============================================================
// 路由设置
// ============================================================

func setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// 公共路由 — 基础中间件
	publicHandler := Chain(
		http.HandlerFunc(handleHello),
		LoggingMiddleware,
		RateLimitMiddleware(5, 10), // 每秒 5 个请求，突发 10 个
		RecoveryMiddleware,
	)
	mux.Handle("/api/public/hello", publicHandler)

	// 受保护路由 — 需要认证
	protectedHandler := Chain(
		http.HandlerFunc(handleProtectedData),
		LoggingMiddleware,
		RateLimitMiddleware(10, 20),
		AuthMiddleware,
		RecoveryMiddleware,
	)
	mux.Handle("/api/protected/data", protectedHandler)

	// 慢请求路由 — 演示超时
	slowHandler := Chain(
		http.HandlerFunc(handleSlowRequest),
		LoggingMiddleware,
		TimeoutMiddleware(2*time.Second), // 2 秒超时
		RecoveryMiddleware,
	)
	mux.Handle("/api/slow", slowHandler)

	// Panic 路由 — 演示恢复
	panicHandler := Chain(
		http.HandlerFunc(handlePanic),
		LoggingMiddleware,
		RecoveryMiddleware,
	)
	mux.Handle("/api/panic", panicHandler)

	return mux
}

func main() {
	fmt.Println("========================================")
	fmt.Println("Go HTTP 中间件示例")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("可用端点:")
	fmt.Println("  GET /api/public/hello    — 公共接口（限流: 5次/秒）")
	fmt.Println("  GET /api/protected/data  — 需认证（Token: valid-token-123）")
	fmt.Println("  GET /api/slow            — 模拟慢请求（2秒超时）")
	fmt.Println("  GET /api/panic           — 模拟 panic（测试恢复）")
	fmt.Println()
	fmt.Println("测试命令:")
	fmt.Println("  curl http://localhost:8080/api/public/hello")
	fmt.Println("  curl -H 'Authorization: Bearer valid-token-123' http://localhost:8080/api/protected/data")
	fmt.Println("  curl http://localhost:8080/api/slow")
	fmt.Println("  curl http://localhost:8080/api/panic")
	fmt.Println()

	handler := setupRoutes()

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("🚀 服务启动在 http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}