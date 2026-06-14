package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// ==============================================
// 令牌桶算法（Token Bucket）
// ==============================================

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	rate       float64       // 每秒填充速率
	capacity   float64       // 桶容量
	tokens     float64       // 当前令牌数
	lastRefill time.Time     // 上次填充时间
	mu         sync.Mutex    // 并发安全锁
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(rate, capacity float64) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity, // 初始满桶
		lastRefill: time.Now(),
	}
}

// refill 按时间差补充令牌
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now
}

// Allow 是否允许通过（消耗一个令牌）
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

// AllowN 是否允许 n 个请求通过
func (tb *TokenBucket) AllowN(n float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}
	return false
}

// Available 返回当前可用令牌数
func (tb *TokenBucket) Available() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	return tb.tokens
}

// ==============================================
// 滑动窗口计数器（Sliding Window Counter）
// ==============================================

// SlidingWindow 滑动窗口限流器
type SlidingWindow struct {
	windowSize time.Duration // 窗口大小（例如 1秒）
	maxRequests int          // 窗口内最大请求数
	entries     []time.Time  // 窗口内的请求时间戳
	mu          sync.Mutex   // 并发安全锁
}

// NewSlidingWindow 创建滑动窗口限流器
func NewSlidingWindow(windowSize time.Duration, maxRequests int) *SlidingWindow {
	return &SlidingWindow{
		windowSize:  windowSize,
		maxRequests: maxRequests,
		entries:     make([]time.Time, 0, maxRequests),
	}
}

// Allow 是否允许通过
func (sw *SlidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-sw.windowSize)

	// 移除窗口之外的旧时间戳
	cutoff := 0
	for i, t := range sw.entries {
		if t.After(windowStart) {
			cutoff = i
			break
		}
		cutoff = i + 1
	}
	sw.entries = sw.entries[cutoff:]

	// 检查是否超过限制
	if len(sw.entries) >= sw.maxRequests {
		return false
	}

	sw.entries = append(sw.entries, now)
	return true
}

// Remaining 返回窗口内剩余的请求配额
func (sw *SlidingWindow) Remaining() int {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-sw.windowSize)
	count := 0
	for _, t := range sw.entries {
		if t.After(windowStart) {
			count++
		}
	}
	return sw.maxRequests - count
}

// ==============================================
// 每 IP 限流器
// ==============================================

// IPRateLimiter 按 IP 地址隔离的限流器
type IPRateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*TokenBucket
	rate     float64
	capacity float64
}

// NewIPRateLimiter 创建按 IP 限流的限流器
func NewIPRateLimiter(rate, capacity float64) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*TokenBucket),
		rate:     rate,
		capacity: capacity,
	}
}

// GetLimiter 获取（或创建）指定 IP 的限流器
func (iprl *IPRateLimiter) GetLimiter(ip string) *TokenBucket {
	iprl.mu.RLock()
	limiter, exists := iprl.limiters[ip]
	iprl.mu.RUnlock()

	if exists {
		return limiter
	}

	iprl.mu.Lock()
	defer iprl.mu.Unlock()

	// 双重检查
	if limiter, exists = iprl.limiters[ip]; exists {
		return limiter
	}

	limiter = NewTokenBucket(iprl.rate, iprl.capacity)
	iprl.limiters[ip] = limiter
	return limiter
}

// Allow 指定 IP 是否允许通过
func (iprl *IPRateLimiter) Allow(ip string) bool {
	return iprl.GetLimiter(ip).Allow()
}

// ==============================================
// HTTP 中间件
// ==============================================

// RateLimitMiddleware 限流 HTTP 中间件
// 同时使用全局限流和每 IP 限流
type RateLimitMiddleware struct {
	global  *TokenBucket
	perIP   *IPRateLimiter
	sliding *SlidingWindow
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware() *RateLimitMiddleware {
	return &RateLimitMiddleware{
		global:  NewTokenBucket(100, 50),        // 全局：每秒 100 个，桶容量 50
		perIP:   NewIPRateLimiter(10, 5),         // 每 IP：每秒 10 个，桶容量 5
		sliding: NewSlidingWindow(1*time.Second, 20), // 滑动窗口：每秒最多 20 个
	}
}

// Middleware 返回 HTTP 中间件处理器
func (rm *RateLimitMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		// 全局令牌桶检查
		if !rm.global.Allow() {
			http.Error(w, "429 - 全局限流（令牌桶）", http.StatusTooManyRequests)
			return
		}

		// 每 IP 令牌桶检查
		if !rm.perIP.Allow(ip) {
			http.Error(w, fmt.Sprintf("429 - IP 限流（令牌桶），IP: %s", ip), http.StatusTooManyRequests)
			return
		}

		// 滑动窗口检查
		if !rm.sliding.Allow() {
			http.Error(w, "429 - 全局限流（滑动窗口）", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("=== Go 限流算法示例 ===")
	fmt.Println()

	// --- 1. 令牌桶演示 ---
	fmt.Println("--- 1. 令牌桶算法 ---")
	tb := NewTokenBucket(10, 5) // 每秒 10 个，桶容量 5

	fmt.Printf("初始令牌: %.1f\n", tb.Available())
	// 短时间内连续消耗
	for i := range 7 {
		allowed := tb.Allow()
		fmt.Printf("  请求 %d: %v (令牌: %.1f)\n", i+1, allowed, tb.Available())
	}
	time.Sleep(200 * time.Millisecond) // 等待补充
	fmt.Printf("等待 200ms 后令牌: %.1f\n", tb.Available())

	// --- 2. 滑动窗口演示 ---
	fmt.Println("\n--- 2. 滑动窗口计数器 ---")
	sw := NewSlidingWindow(1*time.Second, 5) // 每秒最多 5 个

	for i := range 7 {
		allowed := sw.Allow()
		fmt.Printf("  请求 %d: %v (剩余: %d)\n", i+1, allowed, sw.Remaining())
	}

	// --- 3. 每 IP 限流演示 ---
	fmt.Println("\n--- 3. 每 IP 限流 ---")
	ipLimiter := NewIPRateLimiter(3, 3) // 每 IP 每秒 3 个，容量 3

	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.1"}
	for _, ip := range ips {
		allowed := ipLimiter.Allow(ip)
		fmt.Printf("  IP %s: %v\n", ip, allowed)
	}

	// --- 4. HTTP 中间件演示 ---
	fmt.Println("\n--- 4. HTTP 限流中间件 ---")
	rm := NewRateLimitMiddleware()

	// 创建一个简单的 HTTP 处理器
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "✅ 请求通过限流！")
	})

	// 包装限流中间件
	handler := rm.Middleware(helloHandler)

	// 启动 HTTP 服务
	addr := ":8081"
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// 在后台启动服务器
	go func() {
		log.Printf("🌐 限流测试服务启动在 http://localhost%s", addr)
		log.Printf("   全局令牌桶: 100 req/s, 容量 50")
		log.Printf("   每 IP 令牌桶: 10 req/s, 容量 5")
		log.Printf("   滑动窗口: 20 req/s")
		log.Printf("   测试: curl http://localhost%s", addr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 给服务器一点启动时间
	time.Sleep(100 * time.Millisecond)

	// 发送一些测试请求（通过内建 client）
	fmt.Println("  发送测试请求...")
	client := &http.Client{Timeout: 1 * time.Second}
	for i := range 5 {
		resp, err := client.Get(fmt.Sprintf("http://localhost%s", addr))
		if err != nil {
			fmt.Printf("  请求 %d: 错误 - %v\n", i+1, err)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			fmt.Printf("  请求 %d: ✅ 通过 (200)\n", i+1)
		} else {
			fmt.Printf("  请求 %d: ❌ 限流 (429)\n", i+1)
		}
		resp.Body.Close()
	}

	// 关闭服务器
	server.Close()
	fmt.Println("\n✅ 限流示例运行完成")
}