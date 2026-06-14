package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// ==============================================
// TokenBucket（令牌桶）测试
// ==============================================

// TestTokenBucketNew 测试令牌桶的创建
func TestTokenBucketNew(t *testing.T) {
	tb := NewTokenBucket(10, 5)
	if tb == nil {
		t.Fatal("NewTokenBucket() 返回 nil")
	}
	// 初始应为满桶
	avail := tb.Available()
	if avail != 5 {
		t.Errorf("初始令牌数 = %.1f, 期望 5", avail)
	}
}

// TestTokenBucketAllow 测试令牌桶的 Allow 方法
func TestTokenBucketAllow(t *testing.T) {
	tb := NewTokenBucket(100, 5) // 高速率确保测试中不会补充太多

	t.Run("初始允许通过", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			if !tb.Allow() {
				t.Errorf("第 %d 个请求应被允许", i+1)
			}
		}
	})

	t.Run("桶空后拒绝", func(t *testing.T) {
		// 第 6 个请求应被拒绝（桶容量=5）
		if tb.Allow() {
			t.Error("桶空后 Allow() 应返回 false")
		}
	})
}

// TestTokenBucketAllowN 测试 AllowN 方法
func TestTokenBucketAllowN(t *testing.T) {
	t.Run("消耗多个令牌后拒绝", func(t *testing.T) {
		tb := NewTokenBucket(1000, 10)
		if !tb.AllowN(8) {
			t.Error("AllowN(8) 应该成功（容量=10）")
		}
		if !tb.AllowN(2) {
			t.Error("AllowN(2) 应该成功（剩余 2）")
		}
		if tb.AllowN(1) {
			t.Error("AllowN(1) 应被拒绝（令牌已耗尽）")
		}
	})

	t.Run("超过容量的请求被拒绝", func(t *testing.T) {
		tb := NewTokenBucket(100, 5)
		if tb.AllowN(6) {
			t.Error("AllowN(6) 应被拒绝（容量=5）")
		}
	})
}

// TestTokenBucketRefill 测试令牌补充
func TestTokenBucketRefill(t *testing.T) {
	rate := 100.0 // 每秒 100 个
	capacity := 10.0
	tb := NewTokenBucket(rate, capacity)

	// 耗光所有令牌
	tb.AllowN(capacity)
	if tb.Allow() {
		t.Error("耗光后 Allow() 应返回 false")
	}

	// 等待 50ms，应补充约 5 个令牌
	time.Sleep(50 * time.Millisecond)
	avail := tb.Available()
	if avail < 4 || avail > 6 {
		t.Errorf("50ms 后令牌数应约为 5，实际 %.2f", avail)
	}

	// 补充不应超过容量
	time.Sleep(200 * time.Millisecond) // 足以补充到满桶
	avail = tb.Available()
	if avail > capacity {
		t.Errorf("令牌不应超过容量 %.0f，实际 %.2f", capacity, avail)
	}
}

// TestTokenBucketConcurrency 测试令牌桶并发安全
func TestTokenBucketConcurrency(t *testing.T) {
	tb := NewTokenBucket(10000, 100)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tb.Allow()
			tb.Available()
		}()
	}
	wg.Wait()
	// 测试不 panic 即通过
}

// ==============================================
// SlidingWindow（滑动窗口）测试
// ==============================================

// TestSlidingWindowNew 测试滑动窗口创建
func TestSlidingWindowNew(t *testing.T) {
	sw := NewSlidingWindow(1*time.Second, 5)
	if sw == nil {
		t.Fatal("NewSlidingWindow() 返回 nil")
	}
	rem := sw.Remaining()
	if rem != 5 {
		t.Errorf("初始剩余配额 = %d, 期望 5", rem)
	}
}

// TestSlidingWindowAllow 测试滑动窗口的 Allow 方法
func TestSlidingWindowAllow(t *testing.T) {
	sw := NewSlidingWindow(1*time.Second, 3)

	t.Run("窗口内允许不超过限制", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			if !sw.Allow() {
				t.Errorf("第 %d 个请求应被允许", i+1)
			}
		}
	})

	t.Run("超过限制后拒绝", func(t *testing.T) {
		if sw.Allow() {
			t.Error("超过限制后 Allow() 应返回 false")
		}
	})

	t.Run("剩余配额计算正确", func(t *testing.T) {
		// 已消耗 3 个，配额 3，剩余应为 0
		rem := sw.Remaining()
		if rem != 0 {
			t.Errorf("剩余配额 = %d, 期望 0", rem)
		}
	})
}

// TestSlidingWindowWindowReset 测试窗口重置
func TestSlidingWindowWindowReset(t *testing.T) {
	sw := NewSlidingWindow(100*time.Millisecond, 2)

	sw.Allow() // 1st
	sw.Allow() // 2nd
	if sw.Allow() {
		t.Error("超过限制应返回 false")
	}

	// 等待窗口过期
	time.Sleep(150 * time.Millisecond)

	// 窗口应已重置
	if !sw.Allow() {
		t.Error("窗口重置后 Allow() 应返回 true")
	}
}

// ==============================================
// IPRateLimiter（按 IP 限流）测试
// ==============================================

// TestIPRateLimiter 测试 IP 级别的限流
func TestIPRateLimiter(t *testing.T) {
	rl := NewIPRateLimiter(100, 3) // 每 IP 容量 3

	t.Run("不同 IP 互不影响", func(t *testing.T) {
		// IP-A 消耗完
		for i := 0; i < 3; i++ {
			if !rl.Allow("192.168.1.1") {
				t.Errorf("IP-A 第 %d 次应被允许", i+1)
			}
		}
		// IP-A 第 4 次应被拒绝
		if rl.Allow("192.168.1.1") {
			t.Error("IP-A 第 4 次应被拒绝")
		}
		// IP-B 应不受影响
		if !rl.Allow("192.168.1.2") {
			t.Error("IP-B 应被允许（不同 IP）")
		}
	})

	t.Run("GetLimiter 返回相同的实例", func(t *testing.T) {
		l1 := rl.GetLimiter("10.0.0.1")
		l2 := rl.GetLimiter("10.0.0.1")
		if l1 != l2 {
			t.Error("同一 IP 的 GetLimiter 应返回相同实例")
		}
	})

	t.Run("新 IP 自动创建限流器", func(t *testing.T) {
		if !rl.Allow("10.0.0.99") {
			t.Error("新 IP 首次应被允许")
		}
	})
}

// ==============================================
// RateLimitMiddleware（HTTP 限流中间件）测试
// ==============================================

// TestRateLimitMiddleware 测试 HTTP 中间件
func TestRateLimitMiddleware(t *testing.T) {
	rm := NewRateLimitMiddleware()

	// 创建一个测试用 handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := rm.Middleware(nextHandler)

	t.Run("正常请求通过", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("状态码 = %d, 期望 %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
	})
}

// TestRateLimitMiddlewareGlobalLimit 测试全局限流
func TestRateLimitMiddlewareGlobalLimit(t *testing.T) {
	// 用较小的容量测试限流效果
	rm := &RateLimitMiddleware{
		global:  NewTokenBucket(1000, 0), // 容量 0，立即拒绝
		perIP:   NewIPRateLimiter(1000, 100),
		sliding: NewSlidingWindow(1*time.Minute, 1000),
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rm.Middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("全局限流应返回 429，实际 %d", rec.Code)
	}
}

// TestRateLimitMiddlewareIPLimit 测试 IP 限流
func TestRateLimitMiddlewareIPLimit(t *testing.T) {
	rm := &RateLimitMiddleware{
		global:  NewTokenBucket(10000, 10000),
		perIP:   NewIPRateLimiter(1000, 3), // 每 IP 最多 3 个
		sliding: NewSlidingWindow(1*time.Minute, 10000),
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rm.Middleware(nextHandler)

	// 前 3 个应通过
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("第 %d 个请求应通过，得到 %d", i+1, rec.Code)
		}
	}

	// 第 4 个应被 IP 限流拒绝
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("第 4 个请求应被限流，得到 %d", rec.Code)
	}
}