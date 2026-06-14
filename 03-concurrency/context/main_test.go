// main_test.go — context 包上下文管理示例测试套件
//
// 测试覆盖：
//   - context.Background / TODO
//   - WithCancel 主动取消传播
//   - WithTimeout 超时取消
//   - WithDeadline 绝对截止时间
//   - WithValue 上下文传值
//   - simulateHTTPCall 模拟带超时的 HTTP 调用
package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

// ============================================================
// 1. 测试 context.Background 和 TODO
// ============================================================

// TestBackground 验证 Background context 不会被取消
func TestBackground(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	select {
	case <-ctx.Done():
		t.Error("Background context 不应被取消")
	default:
		// 正常：不会执行到 Done branch
	}

	if ctx.Err() != nil {
		t.Errorf("Background.Err() 应为 nil, got %v", ctx.Err())
	}
}

// TestTODO 验证 TODO context 行为与 Background 一致
func TestTODO(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	select {
	case <-ctx.Done():
		t.Error("TODO context 不应被取消")
	default:
	}

	if ctx.Err() != nil {
		t.Errorf("TODO.Err() 应为 nil, got %v", ctx.Err())
	}
}

// ============================================================
// 2. 测试 WithCancel — 主动取消
// ============================================================

// TestWithCancel 验证 cancel 函数能传播取消信号
func TestWithCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	<-ctx.Done()
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("取消后 context 错误应为 Canceled, got %v", ctx.Err())
	}
}

// TestWithCancelConcurrent 验证多 goroutine 能同时收到取消信号
func TestWithCancelConcurrent(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	n := 5
	done := make(chan struct{}, n)

	for i := 0; i < n; i++ {
		go func() {
			<-ctx.Done()
			done <- struct{}{}
		}()
	}

	// 确保所有 goroutine 都启动
	time.Sleep(10 * time.Millisecond)
	cancel()

	// 等待所有 goroutine 响应取消
	timeout := time.After(1 * time.Second)
	for i := 0; i < n; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatalf("只有 %d/%d goroutine 响应取消", i, n)
		}
	}
}

// TestWithCancelOrder 验证 context 传播链
func TestWithCancelOrder(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithCancel(context.Background())
	child, childCancel := context.WithCancel(parent)
	defer childCancel()

	parentCancel() // 取消 parent
	<-parent.Done()

	// child 也应被取消（继承传播）
	select {
	case <-child.Done():
		if !errors.Is(child.Err(), context.Canceled) {
			t.Errorf("child 错误应为 Canceled, got %v", child.Err())
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("parent 取消后 child 应同步取消")
	}
}

// ============================================================
// 3. 测试 WithTimeout
// ============================================================

// TestWithTimeout 验证超时后 context 自动取消
func TestWithTimeout(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	<-ctx.Done()
	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Errorf("超时后错误应为 DeadlineExceeded, got %v", ctx.Err())
	}
}

// TestWithTimeoutBeforeExpiry 验证在超时前取消
func TestWithTimeoutBeforeExpiry(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	cancel() // 在超时前主动取消

	<-ctx.Done()
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("主动取消后错误应为 Canceled, got %v", ctx.Err())
	}
}

// TestWithTimeoutDuration 验证超时时间大致准确
func TestWithTimeoutDuration(t *testing.T) {
	t.Parallel()

	timeout := 100 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	<-ctx.Done()
	elapsed := time.Since(start)

	// 超时不能在到期前触发
	if elapsed < timeout/2 {
		t.Errorf("超时发生太快: %v, 期望至少 %v", elapsed, timeout/2)
	}
	// 超时不应太迟
	if elapsed > timeout*3 {
		t.Errorf("超时发生太慢: %v, 期望最多 %v", elapsed, timeout*2)
	}
}

// ============================================================
// 4. 测试 WithDeadline
// ============================================================

// TestWithDeadline 验证绝对截止时间
func TestWithDeadline(t *testing.T) {
	t.Parallel()

	deadline := time.Now().Add(50 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	<-ctx.Done()
	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Errorf("到期后错误应为 DeadlineExceeded, got %v", ctx.Err())
	}
}

// TestDeadlineMethod 验证 Deadline 方法返回正确时间
func TestDeadlineMethod(t *testing.T) {
	t.Parallel()

	deadline := time.Now().Add(100 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	d, ok := ctx.Deadline()
	if !ok {
		t.Error("WithDeadline 创建的 context 应返回 ok=true")
	}
	// 精度允许 1ms 误差
	if d.Sub(deadline) > time.Millisecond || deadline.Sub(d) > time.Millisecond {
		t.Errorf("Deadline 不匹配: got %v, want %v", d, deadline)
	}
}

// TestBackgroundDeadline 验证 Background 没有截止时间
func TestBackgroundDeadline(t *testing.T) {
	t.Parallel()

	_, ok := context.Background().Deadline()
	if ok {
		t.Error("Background context 不应有截止时间")
	}
}

// ============================================================
// 5. 测试 WithValue
// ============================================================

// TestWithValue 验证 context 传值
func TestWithValue(t *testing.T) {
	t.Parallel()

	type ctxKey string
	const key ctxKey = "user_id"

	ctx := context.WithValue(context.Background(), key, "u-1001")

	val, ok := ctx.Value(key).(string)
	if !ok {
		t.Fatal("Value 类型断言失败")
	}
	if val != "u-1001" {
		t.Errorf("Value: got %q, want 'u-1001'", val)
	}
}

// TestWithValueMultiple 验证多层 context 传值
func TestWithValueMultiple(t *testing.T) {
	t.Parallel()

	type ctxKey string
	const (
		keyUserID  ctxKey = "user_id"
		keyTraceID ctxKey = "trace_id"
	)

	ctx := context.WithValue(context.Background(), keyUserID, "u-1001")
	ctx = context.WithValue(ctx, keyTraceID, "trace-abc-123")

	userID := ctx.Value(keyUserID).(string)
	traceID := ctx.Value(keyTraceID).(string)

	if userID != "u-1001" {
		t.Errorf("userID: got %q, want 'u-1001'", userID)
	}
	if traceID != "trace-abc-123" {
		t.Errorf("traceID: got %q, want 'trace-abc-123'", traceID)
	}
}

// TestWithValueMissingKey 验证读取不存在的 key 返回 nil
func TestWithValueMissingKey(t *testing.T) {
	t.Parallel()

	type ctxKey string
	ctx := context.Background()

	v := ctx.Value(ctxKey("nonexistent"))
	if v != nil {
		t.Errorf("不存在的 key 应返回 nil, got %v", v)
	}
}

// TestWithValueTypeSafety 验证不同类型 key 不会冲突
func TestWithValueTypeSafety(t *testing.T) {
	t.Parallel()

	type keyA string
	type keyB string

	ctx := context.WithValue(context.Background(), keyA("name"), "value-a")
	ctx = context.WithValue(ctx, keyB("name"), "value-b")

	va := ctx.Value(keyA("name")).(string)
	vb := ctx.Value(keyB("name")).(string)

	if va != "value-a" {
		t.Errorf("keyA: got %q, want 'value-a'", va)
	}
	if vb != "value-b" {
		t.Errorf("keyB: got %q, want 'value-b'", vb)
	}
}

// ============================================================
// 6. 测试 simulateHTTPCall — 带超时的 HTTP 模拟
// ============================================================

// TestSimulateHTTPCallOK 验证正常请求能返回结果
func TestSimulateHTTPCallOK(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result, err := simulateHTTPCall(ctx, "/api/test")
	if err != nil {
		t.Fatalf("正常请求不应返回错误: %v", err)
	}
	if result == "" {
		t.Error("正常请求应返回非空结果")
	}
}

// TestSimulateHTTPCallTimeout 验证超时请求返回 DeadlineExceeded
func TestSimulateHTTPCallTimeout(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := simulateHTTPCall(ctx, "/api/slow")
	if err == nil {
		t.Fatal("超时请求应返回错误")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("错误应为 DeadlineExceeded, got %v", err)
	}
}

// TestSimulateHTTPCallCancel 验证主动取消返回 Canceled
func TestSimulateHTTPCallCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	_, err := simulateHTTPCall(ctx, "/api/cancel")
	if err == nil {
		t.Fatal("取消的请求应返回错误")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("错误应为 Canceled, got %v", err)
	}
}

// simulateHTTPCall — 从 main.go 提取的测试辅助函数
func simulateHTTPCall(ctx context.Context, url string) (string, error) {
	wait := 50 * time.Millisecond

	select {
	case <-time.After(wait):
		return "响应来自 " + url, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}