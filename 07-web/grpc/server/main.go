// main.go — gRPC 服务端示例
//
// 本文件实现 gRPC HelloService 的服务端，演示：
//   - 实现 proto 定义的服务接口
//   - 启动 gRPC 服务器
//   - 简单 RPC 和服务端流式 RPC
//   - 拦截器（Unary Interceptor）
//   - 优雅关闭
//
// 安装依赖:
//   go get google.golang.org/grpc google.golang.org/protobuf
//
// 运行前先生成 proto 代码:
//   protoc --go_out=. --go_opt=paths=source_relative \
//          --go-grpc_out=. --go-grpc_opt=paths=source_relative \
//          proto/hello.proto
//
// 生成后的目录结构:
//   grpc/
//   ├── proto/
//   │   ├── hello.proto
//   │   └── hello/           # 生成的 Go 代码
//   │       ├── hello.pb.go
//   │       └── hello_grpc.pb.go
//   ├── server/main.go
//   └── client/main.go
//
// 运行:
//   终端 1: go run 07-web/grpc/server/main.go
//   终端 2: go run 07-web/grpc/client/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/pingxin403/go-examples/07-web/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// =========================================================
// 1. 服务端实现
// =========================================================

// helloServer 实现 HelloServiceServer 接口
type helloServer struct {
	pb.UnimplementedHelloServiceServer
}

// SayHello 简单 RPC — 根据名字返回问候语
func (s *helloServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	log.Printf("[SayHello] 收到请求: name=%s", req.GetName())
	message := fmt.Sprintf("你好，%s！欢迎来到 gRPC 世界！", req.GetName())
	return &pb.HelloResponse{Message: message}, nil
}

// SayHelloStream 服务端流式 RPC — 返回多条问候语
func (s *helloServer) SayHelloStream(req *pb.HelloRequest, stream pb.HelloService_SayHelloStreamServer) error {
	name := req.GetName()
	log.Printf("[SayHelloStream] 收到请求: name=%s", name)

	// 发送 3 条问候消息，每次间隔 500ms
	messages := []string{
		fmt.Sprintf("你好，%s！", name),
		fmt.Sprintf("欢迎学习 gRPC，%s！", name),
		fmt.Sprintf("祝你编程愉快，%s！再见！", name),
	}

	for _, msg := range messages {
		resp := &pb.HelloResponse{Message: msg}
		if err := stream.Send(resp); err != nil {
			return fmt.Errorf("发送消息失败: %w", err)
		}
		log.Printf("[SayHelloStream] 已发送: %s", msg)
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

// =========================================================
// 2. 拦截器 (Interceptor)
// =========================================================

// loggingInterceptor 一元 RPC 日志拦截器 — 记录方法名和耗时
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	log.Printf("[Interceptor] 开始处理 RPC: %s", info.FullMethod)

	resp, err := handler(ctx, req)

	duration := time.Since(start)
	if err != nil {
		log.Printf("[Interceptor] RPC %s 失败: %v (耗时: %v)", info.FullMethod, err, duration)
	} else {
		log.Printf("[Interceptor] RPC %s 成功 (耗时: %v)", info.FullMethod, duration)
	}

	return resp, err
}

// =========================================================
// 3. 主函数
// =========================================================

func main() {
	// 监听 TCP 端口
	port := ":50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("监听端口失败: %v", err)
	}
	log.Printf("gRPC 服务器监听于 %s", port)

	// 创建 gRPC 服务器（注册一元拦截器）
	s := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	// 注册 HelloService 的实现
	pb.RegisterHelloServiceServer(s, &helloServer{})

	// 注册反射服务（方便 grpcurl 等工具调试）
	reflection.Register(s)

	// 优雅关闭 — 监听 SIGINT/SIGTERM 信号
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("收到信号 %v，正在优雅关闭...", sig)
		s.GracefulStop()
		log.Println("gRPC 服务器已关闭")
	}()

	log.Println("gRPC HelloService 已就绪")
	log.Printf("服务器监听地址: localhost%s", port)
	log.Println("可用 RPC:")
	log.Println("  SayHello(name)        → 简单问候")
	log.Println("  SayHelloStream(name)  → 流式问候（3 条）")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("gRPC 服务器运行失败: %v", err)
	}
}