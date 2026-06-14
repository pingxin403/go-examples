// main_test.go — gRPC 服务端单元测试
//
// 测试内容：
//   - helloServer.SayHello 简单 RPC
//   - helloServer.SayHelloStream 服务端流式 RPC
//   - loggingInterceptor 日志拦截器
//   - 使用 bufconn 实现内存级 gRPC 通信

package main

import (
	"context"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	pb "github.com/pingxin403/go-examples/07-web/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024 // 1MB 缓冲区

// setupTestServer 创建内存中的 gRPC 测试服务器，返回 listener 和 client
func setupTestServer(t *testing.T) (pb.HelloServiceClient, func()) {
	t.Helper()

	lis := bufconn.Listen(bufSize)

	// 创建 gRPC 服务器（注册拦截器）
	s := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)
	pb.RegisterHelloServiceServer(s, &helloServer{})

	// 启动服务器
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("测试服务器停止: %v", err)
		}
	}()

	// 创建客户端连接（通过 bufconn 拨号）
	conn, err := grpc.NewClient(
		"passthrough:bufnet",
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("创建 gRPC 客户端连接失败: %v", err)
	}

	client := pb.NewHelloServiceClient(conn)

	// 返回清理函数
	cleanup := func() {
		conn.Close()
		s.Stop()
	}

	return client, cleanup
}

// ---------- SayHello 测试 ----------

// TestSayHello 测试简单 RPC SayHello
func TestSayHello(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("正常调用 — 中文名", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "张三"})
		if err != nil {
			t.Fatalf("SayHello 调用失败: %v", err)
		}

		expected := "你好，张三！欢迎来到 gRPC 世界！"
		if resp.GetMessage() != expected {
			t.Errorf("预期消息 %q，实际得到 %q", expected, resp.GetMessage())
		}
	})

	t.Run("正常调用 — 英文名", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "Alice"})
		if err != nil {
			t.Fatalf("SayHello 调用失败: %v", err)
		}

		if !strings.Contains(resp.GetMessage(), "Alice") {
			t.Errorf("响应应包含 'Alice'，实际得到 %s", resp.GetMessage())
		}
	})

	t.Run("空名字", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: ""})
		if err != nil {
			t.Fatalf("SayHello 调用失败: %v", err)
		}

		if resp.GetMessage() == "" {
			t.Error("即使名字为空，响应也不应为空")
		}
	})

	t.Run("长名字", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		longName := strings.Repeat("名", 100)
		resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: longName})
		if err != nil {
			t.Fatalf("SayHello 调用失败: %v", err)
		}

		if !strings.Contains(resp.GetMessage(), longName) {
			t.Error("响应应包含长名字")
		}
	})

	t.Run("Context 超时", func(t *testing.T) {
		// 使用极短的超时
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// 给 Context 一点时间让它超时
		time.Sleep(10 * time.Millisecond)

		_, err := client.SayHello(ctx, &pb.HelloRequest{Name: "测试"})
		if err == nil {
			t.Error("超时的 Context 应返回错误")
		}
	})
}

// TestSayHelloStream 测试服务端流式 RPC SayHelloStream
func TestSayHelloStream(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("接收流式消息", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		stream, err := client.SayHelloStream(ctx, &pb.HelloRequest{Name: "李四"})
		if err != nil {
			t.Fatalf("SayHelloStream 启动失败: %v", err)
		}

		var messages []string
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("SayHelloStream 接收失败: %v", err)
			}
			messages = append(messages, resp.GetMessage())
		}

		// 应收到 3 条消息
		if len(messages) != 3 {
			t.Fatalf("预期 3 条消息，实际收到 %d 条", len(messages))
		}

		// 验证消息内容
		expectedPrefixes := []string{
			"你好，李四！",
			"欢迎学习 gRPC，李四！",
			"祝你编程愉快，李四！再见！",
		}
		for i, msg := range messages {
			if msg != expectedPrefixes[i] {
				t.Errorf("消息[%d] 预期 %q，实际得到 %q", i, expectedPrefixes[i], msg)
			}
		}
	})

	t.Run("流式消息顺序正确", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		stream, err := client.SayHelloStream(ctx, &pb.HelloRequest{Name: "test"})
		if err != nil {
			t.Fatalf("SayHelloStream 启动失败: %v", err)
		}

		var messages []string
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("SayHelloStream 接收失败: %v", err)
			}
			messages = append(messages, resp.GetMessage())
		}

		// 验证消息顺序递增
		for i := 1; i < len(messages); i++ {
			if messages[i] <= messages[i-1] {
				t.Errorf("消息[%d] (%q) 应不同于消息[%d] (%q)",
					i, messages[i], i-1, messages[i-1])
			}
		}
	})
}

// TestBatchSayHello 测试批量调用 SayHello
func TestBatchSayHello(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	names := []string{"Alice", "Bob", "Charlie", "张三", "李四"}

	for _, name := range names {
		t.Run("批量-"+name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: name})
			if err != nil {
				t.Fatalf("SayHello('%s') 失败: %v", name, err)
			}

			if !strings.Contains(resp.GetMessage(), name) {
				t.Errorf("响应应包含 %q，实际得到 %q", name, resp.GetMessage())
			}
		})
	}
}

// TestLoggingInterceptor 测试日志拦截器是否正常工作
func TestLoggingInterceptor(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 执行一次 RPC，拦截器应该正常工作
	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "拦截器测试"})
	if err != nil {
		t.Fatalf("SayHello 调用失败: %v", err)
	}

	if resp.GetMessage() == "" {
		t.Error("响应消息不应为空")
	}
}

// TestConcurrentCalls 测试并发 RPC 调用
func TestConcurrentCalls(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	done := make(chan bool, 10)

	// 并发 10 个简单 RPC 调用
	for i := 0; i < 10; i++ {
		go func(idx int) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			name := "User" + string(rune('0'+idx))
			resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: name})
			if err != nil {
				t.Errorf("并发调用 %d 失败: %v", idx, err)
			} else if resp.GetMessage() == "" {
				t.Errorf("并发调用 %d 返回空消息", idx)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestMultipleClients 测试多个客户端同时连接
func TestMultipleClients(t *testing.T) {
	// 每个测试用例独立创建 server/client
	t.Run("客户端 1", func(t *testing.T) {
		client, cleanup := setupTestServer(t)
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "Client1"})
		if err != nil {
			t.Fatalf("客户端 1 调用失败: %v", err)
		}
		if !strings.Contains(resp.GetMessage(), "Client1") {
			t.Errorf("响应应包含 'Client1'")
		}
	})

	t.Run("客户端 2", func(t *testing.T) {
		client, cleanup := setupTestServer(t)
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "Client2"})
		if err != nil {
			t.Fatalf("客户端 2 调用失败: %v", err)
		}
		if !strings.Contains(resp.GetMessage(), "Client2") {
			t.Errorf("响应应包含 'Client2'")
		}
	})
}