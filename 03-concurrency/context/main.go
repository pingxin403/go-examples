package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ============================================================
// context 包 — 上下文管理示例
// 涵盖：Background/TODO、WithCancel、WithTimeout、
// WithDeadline、WithValue、HTTP 请求中的上下文
// ============================================================

func main() {
	// ============================================================
	// 1. context.Background() 和 context.TODO()
	// ============================================================
	fmt.Println("=== 1. Background 和 TODO ===")

	// Background: 根 context，永远不会被取消，用于 main 函数或顶层请求
	ctx := context.Background()
	fmt.Printf("  Background: %v\n", ctx)

	// TODO: 根 context，当不确定用哪个 context 时暂用
	_ = context.TODO()
	fmt.Printf("  TODO: %v\n", ctx)

	// ============================================================
	// 2. WithCancel — 主动取消传播
	// ============================================================
	fmt.Println("\n=== 2. WithCancel — 取消传播 ===")

	ctxCancel, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		fmt.Println("  子 goroutine: 调用 cancel()")
		cancel() // 取消 context
	}()

	// 等待 context 取消
	<-ctxCancel.Done()
	fmt.Printf("  主 goroutine: ctx 已取消, err=%v\n", ctxCancel.Err())

	// ============================================================
	// 3. WithTimeout — 超时取消
	// ============================================================
	fmt.Println("\n=== 3. WithTimeout — 超时取消 ===")

	// 创建 100ms 超时的 context
	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancelTimeout() // 防止资源泄漏

	start := time.Now()

	select {
	case <-time.After(200 * time.Millisecond):
		fmt.Println("  任务完成（但实际上会超时）")
	case <-ctxTimeout.Done():
		fmt.Printf("  ctx 超时取消: err=%v, 耗时=%v\n",
			ctxTimeout.Err(), time.Since(start))
	}

	// ============================================================
	// 4. WithDeadline — 绝对截止时间
	// ============================================================
	fmt.Println("\n=== 4. WithDeadline — 绝对截止时间 ===")

	deadline := time.Now().Add(80 * time.Millisecond)
	ctxDeadline, cancelDeadline := context.WithDeadline(context.Background(), deadline)
	defer cancelDeadline()

	start = time.Now()

	select {
	case <-time.After(200 * time.Millisecond):
		fmt.Println("  任务完成（实际会超时）")
	case <-ctxDeadline.Done():
		fmt.Printf("  ctx 到期: err=%v, 计划截止=%v, 实际耗时=%v\n",
			ctxDeadline.Err(), deadline.Sub(time.Now()).Round(time.Millisecond),
			time.Since(start))
	}

	// ============================================================
	// 5. WithValue — 请求范围的上下文值
	// ============================================================
	fmt.Println("\n=== 5. WithValue — 上下文传值 ===")

	// 定义 context key 类型（避免 key 冲突）
	type contextKey string
	const (
		keyUserID  contextKey = "user_id"
		keyTraceID contextKey = "trace_id"
	)

	// 包装多层 value
	ctxVal := context.WithValue(context.Background(), keyUserID, "u-1001")
	ctxVal = context.WithValue(ctxVal, keyTraceID, "trace-abc-123")

	// 在「子 goroutine」中读取
	go func(ctx context.Context) {
		userID, _ := ctx.Value(keyUserID).(string)
		traceID, _ := ctx.Value(keyTraceID).(string)
		fmt.Printf("  子 goroutine 读取: userID=%s, traceID=%s\n", userID, traceID)
	}(ctxVal)

	time.Sleep(10 * time.Millisecond)

	// ============================================================
	// 6. 综合示例 — 模拟带超时的 HTTP 调用
	// ============================================================
	fmt.Println("\n=== 6. 模拟带超时的 HTTP 调用 ===")

	// 模拟一个外部 HTTP 请求（用定时器代替）
	simulateHTTPCall := func(ctx context.Context, url string) (string, error) {
		// 模拟不确定的响应时间
		wait := 50 * time.Millisecond

		select {
		case <-time.After(wait):
			return fmt.Sprintf("响应来自 %s", url), nil
		case <-ctx.Done():
			return "", ctx.Err() // 超时或取消
		}
	}

	// 正常调用（有足够时间）
	ctxOK, cancelOK := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancelOK()

	result, err := simulateHTTPCall(ctxOK, "/api/ok")
	fmt.Printf("  正常请求: result=%q, err=%v\n", result, err)

	// 超时调用（时间不够）
	ctxTimeout2, cancelTimeout2 := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancelTimeout2()

	result, err = simulateHTTPCall(ctxTimeout2, "/api/slow")
	fmt.Printf("  超时请求: result=%q, err=%v\n", result, err)

	// ============================================================
	// 7. 真实 HTTP 请求中的 context 使用
	// ============================================================
	fmt.Println("\n=== 7. 真实 HTTP 请求中的 Context ===")

	// 用 context 控制 HTTP 请求超时
	ctxHTTP, cancelHTTP := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelHTTP()

	req, err := http.NewRequestWithContext(ctxHTTP, "GET", "https://httpbin.org/delay/1", nil)
	if err != nil {
		fmt.Printf("  创建请求失败: %v\n", err)
	} else {
		// 这个请求会成功（/delay/1 延迟 1 秒，超时 2 秒）
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("  请求失败: %v\n", err)
		} else {
			fmt.Printf("  请求成功: status=%s\n", resp.Status)
			resp.Body.Close()
		}
	}

	// 超时验证：用更短的超时
	ctxHTTP2, cancelHTTP2 := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancelHTTP2()

	req2, _ := http.NewRequestWithContext(ctxHTTP2, "GET", "https://httpbin.org/delay/3", nil)
	_, err = http.DefaultClient.Do(req2)
	if err != nil {
		fmt.Printf("  超时验证: 请求因 context 超时被取消: %v\n", err)
	}

	fmt.Println("\n✅ 所有 context 示例完成")
}