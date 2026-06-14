// main_test.go — 对 middleware 包中可测试函数的完整测试套件
//
// 本文件测试:
//   - TokenBucket — 令牌桶限流器
//   - GetUserID / GetRequestID — 上下文数据存取
//   - Chain — 中间件链组合
//   - writeJSON — JSON 响应写入
//   - 各中间件逻辑（日志、认证、限流、超时、恢复）
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// ============================================================
// 1. TokenBucket — 令牌桶限流器
// ============================================================

// TestNewTokenBucket 测试令牌桶创建时的初始状态
func TestNewTokenBucket(t *testing.T) {
	tb := NewTokenBucket(10, 5)
	if tb == nil {
		t.Fatal("NewTokenBucket 返回了 nil")
	}

	// 初始满桶，应能立即消费 tokens
	for i := 0; i < 5; i++ {
		if !tb.Allow() {
			t.Fatalf("第 %d 次 Allow 应返回 true（初始满桶）", i+1)
		}
	}
	// 第 6 次应被拒绝（桶空）
	if tb.Allow() {
		t.Error("桶空后 Allow 应返回 false")
	}
}

// TestTokenBucket_Refill 测试令牌桶按时间自动填充
func TestTokenBucket_Refill(t *testing.T) {
	tb := NewTokenBucket(100, 10) // 每秒 100 个，填充很快

	// 清空令牌桶
	for tb.Allow() {
	}

	// 等待足够时间填充至少一个令牌
	time.Sleep(20 * time.Millisecond)

	if !tb.Allow() {
		t.Error("等待后应能获取到令牌（自动填充）")
	}
}

// TestTokenBucket_ConcurrentAccess 测试令牌桶并发安全
func TestTokenBucket_ConcurrentAccess(t *testing.T) {
	tb := NewTokenBucket(1000, 100)

	var wg sync.WaitGroup
	allowed := make(chan bool, 200)

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed <- tb.Allow()
		}()
	}

	wg.Wait()
	close(allowed)

	// 统计通过的请求数（不应超过容量 100）
	count := 0
	for v := range allowed {
		if v {
			count++
		}
	}
	// 并发访问下可能略超容量，允许 10% 的余量
	if count > 110 {
		t.Errorf("并发访问下通过的请求数 = %d; 期望不超过容量 100 (允许10%%余量)", count)
	}
}

// TestTokenBucket_RateLimit 测试令牌桶速率限制接近预期速率
func TestTokenBucket_RateLimit(t *testing.T) {
	rate := 50.0
	capacity := 10.0
	tb := NewTokenBucket(rate, capacity)

	// 在 100ms 内消费尽可能多的令牌
	start := time.Now()
	count := 0
	for time.Since(start) < 100*time.Millisecond {
		if tb.Allow() {
			count++
		}
		time.Sleep(1 * time.Millisecond) // 避免忙等
	}

	// 100ms 内大约应获得 rate*0.1 个令牌（~5个）
	t.Logf("100ms 内通过 %d 个请求", count)
	if count < 1 {
		t.Error("100ms 内应至少通过 1 个请求")
	}
}

// ============================================================
// 2. GetUserID / GetRequestID — 上下文数据存取
// ============================================================

// TestGetUserID 测试从 context 获取用户 ID
func TestGetUserID(t *testing.T) {
	t.Run("正常获取", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserIDKey, "user-001")
		got := GetUserID(ctx)
		if got != "user-001" {
			t.Errorf("GetUserID() = %q; 期望 %q", got, "user-001")
		}
	})

	t.Run("空 context", func(t *testing.T) {
		got := GetUserID(context.Background())
		if got != "" {
			t.Errorf("空 context GetUserID() = %q; 期望空字符串", got)
		}
	})

	t.Run("错误类型", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserIDKey, 123)
		got := GetUserID(ctx)
		if got != "" {
			t.Errorf("类型不匹配 GetUserID() = %q; 期望空字符串", got)
		}
	})
}

// TestGetRequestID 测试从 context 获取请求 ID
func TestGetRequestID_Middleware(t *testing.T) {
	t.Run("正常获取", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), RequestIDKey, "req-abc-123")
		got := GetRequestID(ctx)
		if got != "req-abc-123" {
			t.Errorf("GetRequestID() = %q; 期望 %q", got, "req-abc-123")
		}
	})

	t.Run("空 context", func(t *testing.T) {
		got := GetRequestID(context.Background())
		if got != "" {
			t.Errorf("空 context GetRequestID() = %q; 期望空字符串", got)
		}
	})
}

// ============================================================
// 3. writeJSON — JSON 响应写入
// ============================================================

// TestWriteJSON 测试 writeJSON 写入正确的状态码和 JSON 内容
func TestWriteJSON(t *testing.T) {
	recorder := httptest.NewRecorder()

	data := map[string]string{"message": "ok", "status": "success"}
	writeJSON(recorder, http.StatusOK, data)

	resp := recorder.Result()
	defer resp.Body.Close()

	t.Run("状态码正确", func(t *testing.T) {
		if resp.StatusCode != http.StatusOK {
			t.Errorf("状态码 = %d; 期望 %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("Content-Type 正确", func(t *testing.T) {
		ct := resp.Header.Get("Content-Type")
		if ct != "application/json; charset=utf-8" {
			t.Errorf("Content-Type = %q; 期望 %q", ct, "application/json; charset=utf-8")
		}
	})

	t.Run("JSON 内容可解析", func(t *testing.T) {
		var result map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("JSON 解码失败: %v", err)
		}
		if result["message"] != "ok" {
			t.Errorf("message = %q; 期望 %q", result["message"], "ok")
		}
	})
}

// TestWriteJSON_WithDifferentStatus 测试 writeJSON 写入不同状态码
func TestWriteJSON_WithDifferentStatus(t *testing.T) {
	statusCodes := []int{
		http.StatusOK,
		http.StatusUnauthorized,
		http.StatusTooManyRequests,
		http.StatusGatewayTimeout,
		http.StatusInternalServerError,
	}

	for _, code := range statusCodes {
		t.Run(http.StatusText(code), func(t *testing.T) {
			recorder := httptest.NewRecorder()
			writeJSON(recorder, code, map[string]string{"error": "test"})
			if recorder.Code != code {
				t.Errorf("状态码 = %d; 期望 %d", recorder.Code, code)
			}
		})
	}
}

// ============================================================
// 4. Chain — 中间件链
// ============================================================

// TestChain_SingleMiddleware 测试单中间件链
func TestChain_SingleMiddleware(t *testing.T) {
	// 创建一个记录是否被调用的中间件
	var called bool
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			next.ServeHTTP(w, r)
		})
	}

	handler := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		middleware,
	)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if !called {
		t.Error("中间件未被调用")
	}
}

// TestChain_MultipleMiddlewares 测试多个中间件按正确顺序执行
func TestChain_MultipleMiddlewares(t *testing.T) {
	var order []string

	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1_before")
			next.ServeHTTP(w, r)
			order = append(order, "mw1_after")
		})
	}

	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2_before")
			next.ServeHTTP(w, r)
			order = append(order, "mw2_after")
		})
	}

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	// Chain(h, mw1, mw2) 执行顺序: mw1 → mw2 → h → mw2 → mw1
	handler := Chain(finalHandler, mw1, mw2)
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	expected := []string{"mw1_before", "mw2_before", "handler", "mw2_after", "mw1_after"}
	for i, v := range expected {
		if i >= len(order) || order[i] != v {
			t.Errorf("执行顺序[%d] = %q; 期望 %q (全部: %v)", i, safeGet(order, i), v, order)
			break
		}
	}
}

// TestChain_Empty 测试空中间件链
func TestChain_Empty(t *testing.T) {
	var called bool
	handler := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}),
	)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if !called {
		t.Error("空中间件链时最终的 handler 未被调用")
	}
}

// ============================================================
// 5. AuthMiddleware — 认证中间件
// ============================================================

// TestAuthMiddleware_ValidToken 测试有效 Token 通过认证
func TestAuthMiddleware_ValidToken(t *testing.T) {
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		if userID != "user-001" {
			t.Errorf("user_id = %q; 期望 %q", userID, "user-001")
		}
		w.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token-123")
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("状态码 = %d; 期望 %d", recorder.Code, http.StatusOK)
	}
}

// TestAuthMiddleware_NoHeader 测试无 Authorization 头
func TestAuthMiddleware_NoHeader(t *testing.T) {
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("无认证头时不应调用后续 handler")
	}))

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("状态码 = %d; 期望 %d", recorder.Code, http.StatusUnauthorized)
	}
}

// TestAuthMiddleware_InvalidToken 测试无效 Token
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("无效 token 时不应调用后续 handler")
	}))

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("状态码 = %d; 期望 %d", recorder.Code, http.StatusUnauthorized)
	}
}

// TestAuthMiddleware_WrongFormat 测试错误的 Authorization 格式
func TestAuthMiddleware_WrongFormat(t *testing.T) {
	tests := []struct {
		name    string
		header  string
	}{
		{name: "无 Bearer 前缀", header: "Token valid-token-123"},
		{name: "只有 Bearer", header: "Bearer "},
		{name: "空字符串", header: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("格式错误时不应调用后续 handler")
			}))

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", tt.header)
			handler.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusUnauthorized {
				t.Errorf("状态码 = %d; 期望 %d", recorder.Code, http.StatusUnauthorized)
			}
		})
	}
}

// ============================================================
// 6. RateLimitMiddleware — 限流中间件
// ============================================================

// TestRateLimitMiddleware 测试限流中间件的基本行为
func TestRateLimitMiddleware(t *testing.T) {
	// 每秒 1000 个请求，容量 5，用于测试初始满桶
	middleware := RateLimitMiddleware(1000, 5)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("容量内请求通过", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/", nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("第 %d 次请求状态码 = %d; 期望 %d", i+1, recorder.Code, http.StatusOK)
			}
		}
	})

	t.Run("超出容量被限流", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/", nil))
		if recorder.Code != http.StatusTooManyRequests {
			t.Errorf("超出容量后状态码 = %d; 期望 %d", recorder.Code, http.StatusTooManyRequests)
		}
	})
}

// ============================================================
// 7. RecoveryMiddleware — 恢复中间件
// ============================================================

// TestRecoveryMiddleware 测试 panic 恢复中间件
func TestRecoveryMiddleware(t *testing.T) {
	handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("模拟错误")
	}))

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("panic 后状态码 = %d; 期望 %d", recorder.Code, http.StatusInternalServerError)
	}

	// 验证 JSON 响应包含 error 字段
	var body map[string]string
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("JSON 解码失败: %v", err)
	}
	if body["error"] != "internal_server_error" {
		t.Errorf("error = %q; 期望 %q", body["error"], "internal_server_error")
	}
}

// TestRecoveryMiddleware_NormalHandler 测试正常 handler 不受影响
func TestRecoveryMiddleware_NormalHandler(t *testing.T) {
	handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/", nil))

	if recorder.Code != http.StatusOK {
		t.Errorf("正常 handler 状态码 = %d; 期望 %d", recorder.Code, http.StatusOK)
	}
}

// ============================================================
// 8. TimeoutMiddleware — 超时中间件
// ============================================================

// TestTimeoutMiddleware_Normal 测试超时中间件在正常请求下放行
func TestTimeoutMiddleware_Normal(t *testing.T) {
	handler := TimeoutMiddleware(1*time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/", nil))

	if recorder.Code != http.StatusOK {
		t.Errorf("正常请求状态码 = %d; 期望 %d", recorder.Code, http.StatusOK)
	}
}

// TestTimeoutMiddleware_Timeout 测试超时中间件在慢请求时返回超时错误
func TestTimeoutMiddleware_Timeout(t *testing.T) {
	handler := TimeoutMiddleware(50*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // 超过超时时间
		w.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/", nil))

	if recorder.Code != http.StatusGatewayTimeout {
		t.Errorf("超时后状态码 = %d; 期望 %d", recorder.Code, http.StatusGatewayTimeout)
	}

	var body map[string]string
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("JSON 解码失败: %v", err)
	}
	if body["error"] != "request_timeout" {
		t.Errorf("error = %q; 期望 %q", body["error"], "request_timeout")
	}
}

// ============================================================
// 9. LoggingMiddleware — 日志中间件（行为验证）
// ============================================================

// TestLoggingMiddleware 测试日志中间件正确传递请求并记录（通过 response 验证）
func TestLoggingMiddleware(t *testing.T) {
	handler := LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证 context 中设置了 request_id
		reqID := GetRequestID(r.Context())
		if reqID == "" {
			t.Error("日志中间件未设置 request_id")
		}
		if !strings.HasPrefix(reqID, "req-") {
			t.Errorf("request_id 格式异常: %q", reqID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/api/test", nil))

	if recorder.Code != http.StatusOK {
		t.Errorf("日志中间件后状态码 = %d; 期望 %d", recorder.Code, http.StatusOK)
	}
}

// TestLoggingResponseWriter 测试 loggingResponseWriter 捕获状态码
func TestLoggingResponseWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	lrw := &loggingResponseWriter{ResponseWriter: recorder, statusCode: http.StatusOK}

	// 默认状态码应为 200
	if lrw.statusCode != http.StatusOK {
		t.Errorf("初始状态码 = %d; 期望 %d", lrw.statusCode, http.StatusOK)
	}

	// 写入 404 状态码
	lrw.WriteHeader(http.StatusNotFound)
	if lrw.statusCode != http.StatusNotFound {
		t.Errorf("写入后状态码 = %d; 期望 %d", lrw.statusCode, http.StatusNotFound)
	}

	// 使用 lrw 写入后原 recorder 也应正确
	if recorder.Code != http.StatusNotFound {
		t.Errorf("recorder 状态码 = %d; 期望 %d", recorder.Code, http.StatusNotFound)
	}
}

// ============================================================
// 辅助函数
// ============================================================

func safeGet(s []string, i int) string {
	if i < len(s) {
		return s[i]
	}
	return ""
}